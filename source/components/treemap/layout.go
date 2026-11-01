package treemap

import (
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
)

// tile is a node drawn in the last layout, in the canvas's coordinates.
// Tiles are in preorder, so a node's descendants follow it.
type tile struct {
	id     string
	parent int
	depth  int
	// rect and amount are where the tile settles and how strongly its color
	// tints it; from and fromAmount are where a data transition starts.
	rect, from         ui.Rect
	amount, fromAmount float32
	color              ui.Color
	style, hover       Style
	branch             bool
	// label and value share the top row; big and caption stack below it in
	// large leaves. Each is fitted to the tile, or empty where it has no room.
	label, value, big, caption string
	valueWidth                 float32
}

// view is the retained presentation of a treemap.
type view struct {
	tiles         []tile
	width, height float32
	hovered       int
	duration      time.Duration

	// The root last laid out and the size it was laid out at.
	root         string
	placed       bool
	placedW      float32
	placedH      float32
	previousByID map[string]int
	morphing     bool
	morphStart   time.Time
	zooming      bool
	zoomIn       bool
	// zoomRoot is the branch drilled into or out of.
	zoomRoot  string
	zoomStart time.Time
	old       []tile
	// hoverID is the tile the pointer is on and hoverOut the one it left,
	// whose highlights cross-fade from hoverStart.
	hoverID, hoverOut string
	hoverStart        time.Time
	// The tooltip follows the pointer, at tipX, tipY, while it moved since
	// the keyboard last changed activeID, the active tile as last built.
	tipPointer, tipOver bool
	tipX, tipY          float32
	activeID            string
	// menuID is the node the context menu was last asked for.
	menuID         string
	pressed        bool
	pressX, pressY float32
	clicks         []click

	sizes           map[*Node]float64
	measures        map[*Node]float64
	measured        bool
	domain          [2]float64
	texts           map[textKey]textSize
	shape           func(s string, size float32, weight int) (float32, float32)
	formatValue     func(float64) string
	formatMeasureFn func(float64) string
	measureLabel    string
	metrics         metrics
}

type click struct{ x, y float32 }

type textKey struct {
	s      string
	size   float32
	weight int
}

type textSize struct{ w, h float32 }

// metrics are the view's lengths in DIPs, from the frame theme, and its tints.
type metrics struct {
	gap, pad, padX, padY, row, rowGap float32
	small, big, line, bigLine         float32
	radius, radiusStep                float32
	tint, step, high                  float32
}

const (
	labelWeight = 500
	bigWeight   = 600
	// tapSlop is how far the pointer may move while pressed for its release
	// to still click.
	tapSlop = 4
	// maxCachedTexts bounds the measured-text cache.
	maxCachedTexts = 4096
)

// begin prepares the frame's measurements; sizes are recomputed each frame
// because the data is borrowed.
func (v *view) begin(c *ui.Context, nodes []Node, opts Options) {
	t := c.Theme()
	if v.sizes == nil {
		v.sizes, v.measures = make(map[*Node]float64), make(map[*Node]float64)
	}
	clear(v.sizes)
	clear(v.measures)
	if v.texts == nil || len(v.texts) > maxCachedTexts {
		v.texts = make(map[textKey]textSize)
	}
	v.formatValue, v.formatMeasureFn, v.measureLabel = opts.FormatValue, opts.FormatMeasure, opts.MeasureLabel
	v.duration = opts.AnimationDuration
	if v.duration == 0 {
		v.duration = 400 * time.Millisecond
	}
	if v.duration < 0 || c.Preferences().ReduceMotion || c.Instant() {
		v.duration = 0
	}
	v.shape = func(s string, size float32, weight int) (float32, float32) {
		k := textKey{s, size, weight}
		if m, ok := v.texts[k]; ok {
			return m.w, m.h
		}
		w, h := c.MeasureText(0, ui.Span{Text: s, Size: size, Weight: weight})
		v.texts[k] = textSize{w, h}
		return w, h
	}
	m := metrics{
		gap: t.Space(1.5), pad: t.Space(1.5), padX: t.Space(2.5), padY: t.Space(1.5), rowGap: t.Space(2),
		small: t.FontSize - 1, big: t.FontSize * 1.5,
		radius: t.Space(3), radiusStep: t.Space(1),
		tint: .13, step: .07, high: .5,
	}
	if t.Dark {
		m.tint, m.step, m.high = .2, .08, .55
	}
	_, m.line = v.shape("Ag", m.small, labelWeight)
	_, m.bigLine = v.shape("Ag", m.big, bigWeight)
	m.row = m.line + 2*m.padY
	v.metrics = m
	v.measured = opts.MeasureLabel != ""
	if v.measured {
		v.domain = opts.MeasureDomain
		if !(v.domain[0] < v.domain[1]) {
			v.domain = v.leafRange(nodes)
		}
	}
	if v.hovered >= len(v.tiles) {
		v.hovered = -1
	}
}

// size is a node's area: a leaf's value, or its children's total, or the
// branch's value when larger.
func (v *view) size(n *Node) float64 {
	if s, ok := v.sizes[n]; ok {
		return s
	}
	s := v.total(n.Children)
	if finite(n.Value) && n.Value > s {
		s = n.Value
	}
	v.sizes[n] = s
	return s
}

func (v *view) total(list []Node) float64 {
	s := 0.0
	for i := range list {
		s += v.size(&list[i])
	}
	return s
}

// free is the room a branch leaves beyond its children.
func (v *view) free(n *Node) float64 { return max(0, v.size(n)-v.total(n.Children)) }

func (v *view) branch(n *Node) bool { return v.total(n.Children) > 0 }

// measure is a leaf's Measure, or a branch's size-weighted mean of its
// children's; NaN when none is known.
func (v *view) measure(n *Node) float64 {
	if m, ok := v.measures[n]; ok {
		return m
	}
	m := n.Measure
	if v.branch(n) {
		sum, weight := 0.0, 0.0
		for i := range n.Children {
			child := &n.Children[i]
			if s, cm := v.size(child), v.measure(child); s > 0 && finite(cm) {
				sum, weight = sum+cm*s, weight+s
			}
		}
		m = math.NaN()
		if weight > 0 {
			m = sum / weight
		}
	} else if !finite(m) {
		m = math.NaN()
	}
	v.measures[n] = m
	return m
}

func (v *view) measureOrNaN(n *Node) float64 {
	if !v.measured {
		return math.NaN()
	}
	return v.measure(n)
}

// leafRange is the range of the measures of leaves with sizes.
func (v *view) leafRange(nodes []Node) [2]float64 {
	lo, hi := math.Inf(1), math.Inf(-1)
	var walk func([]Node)
	walk = func(list []Node) {
		for i := range list {
			n := &list[i]
			if v.size(n) <= 0 {
				continue
			}
			if v.branch(n) {
				walk(n.Children)
			} else if finite(n.Measure) {
				lo, hi = min(lo, n.Measure), max(hi, n.Measure)
			}
		}
	}
	walk(nodes)
	if lo > hi {
		return [2]float64{}
	}
	return [2]float64{lo, hi}
}

// order returns the indices of the nodes with positive sizes, largest first,
// and -1 for extra free space.
func (v *view) order(list []Node, extra float64) []int {
	var out []int
	for i := range list {
		if v.size(&list[i]) > 0 {
			out = append(out, i)
		}
	}
	if extra > 0 {
		out = append(out, -1)
	}
	sizeOf := func(i int) float64 {
		if i < 0 {
			return extra
		}
		return v.size(&list[i])
	}
	slices.SortStableFunc(out, func(a, b int) int {
		switch sa, sb := sizeOf(a), sizeOf(b); {
		case sa > sb:
			return -1
		case sa < sb:
			return 1
		}
		return 0
	})
	return out
}

// place lays out the shown level and prepares the transition from the last
// layout: a zoom when the root changed, or a morph when sizes or shades did.
func (v *view) place(c *ui.Context, root string, shown level) {
	previous := v.tiles
	v.tiles = make([]tile, 0, len(previous))
	v.layout(shown.nodes, shown.extra, ui.Rect{W: v.width, H: v.height}, -1, shown.color)
	now := c.Now()
	resized := v.placedW != v.width || v.placedH != v.height
	switch {
	case v.placed && v.root != root:
		v.startZoom(root, previous, now)
	case v.placed && !v.zooming && !resized:
		v.morph(previous, now)
	}
	v.root, v.placed, v.placedW, v.placedH = root, true, v.width, v.height
}

// startZoom drills from the previous layout, into a branch tiled there or
// out of the old root, tiled in the new layout.
func (v *view) startZoom(root string, previous []tile, now time.Time) {
	old := v.root
	v.zooming, v.morphing, v.old = false, false, nil
	if v.duration == 0 {
		return
	}
	in := root != "" && slices.ContainsFunc(previous, func(p tile) bool { return p.id == root })
	switch {
	case in:
		v.zoomRoot = root
	case old != "" && v.index(old) >= 0:
		v.zoomRoot = old
	default:
		return
	}
	v.zooming, v.zoomIn, v.zoomStart, v.old = true, in, now, previous
}

// morph carries tiles from where they were shown to their new places and
// shades when the data changes, continuing a transition already running.
func (v *view) morph(previous []tile, now time.Time) {
	if v.duration == 0 {
		v.morphing = false
		return
	}
	if v.previousByID == nil {
		v.previousByID = make(map[string]int)
	}
	clear(v.previousByID)
	for i, p := range previous {
		v.previousByID[p.id] = i
	}
	changed := len(previous) != len(v.tiles)
	for _, tl := range v.tiles {
		j, ok := v.previousByID[tl.id]
		if !ok || !near(previous[j].rect, tl.rect) || math.Abs(float64(previous[j].amount-tl.amount)) > .005 {
			changed = true
			break
		}
	}
	e := v.morphProgress(now)
	for i := range v.tiles {
		tl := &v.tiles[i]
		j, ok := v.previousByID[tl.id]
		if !ok {
			continue
		}
		p := previous[j]
		if changed {
			tl.from, tl.fromAmount = lerp(p.from, p.rect, e), p.fromAmount+(p.amount-p.fromAmount)*e
		} else {
			tl.from, tl.fromAmount = p.from, p.fromAmount
		}
	}
	if changed {
		v.morphing, v.morphStart = true, now
	}
}

// morphProgress is how far the data transition has eased, 1 when none runs.
func (v *view) morphProgress(now time.Time) float32 {
	if !v.morphing || v.duration == 0 {
		return 1
	}
	e := float32(now.Sub(v.morphStart)) / float32(v.duration)
	if e >= 1 {
		return 1
	}
	return ui.EaseInOut(max(e, 0))
}

// layout tiles the list within r, below the tile parent. A zero color gives
// top-level nodes palette colors by their order.
func (v *view) layout(list []Node, extra float64, r ui.Rect, parent int, color ui.Color) {
	m := v.metrics
	order := v.order(list, extra)
	values := make([]float64, len(order))
	for k, i := range order {
		values[k] = extra
		if i >= 0 {
			values[k] = v.size(&list[i])
		}
	}
	depth := 0
	if parent >= 0 {
		depth = v.tiles[parent].depth + 1
	}
	for k, cell := range squarify(values, r) {
		i := order[k]
		if i < 0 {
			continue
		}
		n := &list[i]
		cell = inset(cell, m.gap/2, m.gap/2)
		if cell.W < 1 || cell.H < 1 {
			continue
		}
		shade := color
		if shade == (ui.Color{}) {
			shade = paletteColor(i)
		}
		if n.Color != (ui.Color{}) {
			shade = n.Color
		}
		tl := tile{id: n.ID, parent: parent, depth: depth, rect: cell, color: shade, style: n.Style, hover: n.HoverStyle, branch: v.branch(n)}
		tl.amount = min(m.tint+m.step*float32(depth), .45)
		if v.measured && !tl.branch {
			if mv := v.measure(n); finite(mv) && v.domain[1] > v.domain[0] {
				norm := float32(max(0, min(1, (mv-v.domain[0])/(v.domain[1]-v.domain[0]))))
				// The lowest measure keeps the plain tint; higher ones deepen it.
				tl.amount += (max(m.high, tl.amount) - tl.amount) * norm
			}
		}
		tl.from, tl.fromAmount = tl.rect, tl.amount
		var content ui.Rect
		header := false
		if tl.branch {
			switch {
			case cell.W >= 2*m.padX+2*m.line && cell.H >= m.row+m.line+m.pad:
				header = true
				content = ui.Rect{X: cell.X + m.pad, Y: cell.Y + m.row, W: cell.W - 2*m.pad, H: cell.H - m.row - m.pad}
			case cell.W >= 4*m.pad && cell.H >= 4*m.pad:
				content = inset(cell, m.pad, m.pad)
			}
		}
		if !tl.branch || header {
			v.fitTexts(&tl, n)
		}
		v.tiles = append(v.tiles, tl)
		if content.W > 0 && content.H > 0 {
			v.layout(n.Children, v.free(n), content, len(v.tiles)-1, shade)
		}
	}
}

// fitTexts fits a tile's labels: a row with the label and the size where
// there is room for it, and in large leaves the size below it in large type,
// then the caption.
func (v *view) fitTexts(tl *tile, n *Node) {
	m := v.metrics
	r := tl.rect
	room := r.W - 2*m.padX
	if room < m.line || r.H < m.line+2*m.padY {
		return
	}
	value := v.format(v.size(n))
	if !tl.branch {
		bigWidth, _ := v.shape(value, m.big, bigWeight)
		if bigWidth <= room && r.H >= 2*m.padY+m.line+m.padY+m.bigLine {
			tl.label = v.fit(n.Label, m.small, labelWeight, room)
			tl.big = value
			if n.Caption != "" && r.H >= 2*m.padY+2*m.line+m.padY+m.bigLine {
				tl.caption = v.fit(n.Caption, m.small, labelWeight, room)
			}
			return
		}
	}
	valueWidth, _ := v.shape(value, m.small, labelWeight)
	if room >= valueWidth+m.rowGap+2*m.line {
		tl.label = v.fit(n.Label, m.small, labelWeight, room-valueWidth-m.rowGap)
		tl.value, tl.valueWidth = value, valueWidth
		return
	}
	tl.label = v.fit(n.Label, m.small, labelWeight, room)
}

// fit shortens s with an ellipsis to fit width, or returns "" when fewer
// than two of its characters would remain.
func (v *view) fit(s string, size float32, weight int, width float32) string {
	if w, _ := v.shape(s, size, weight); w <= width {
		return s
	}
	const ellipsis = "…"
	runes := []rune(s)
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if w, _ := v.shape(string(runes[:mid])+ellipsis, size, weight); w <= width {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	if lo < 2 {
		return ""
	}
	return string(runes[:lo]) + ellipsis
}

// squarify divides r into rectangles with the areas of values, sorted from
// largest, keeping them as square as it can (Bruls, Huizing and van Wijk).
func squarify(values []float64, r ui.Rect) []ui.Rect {
	out := make([]ui.Rect, len(values))
	total := 0.0
	for _, v := range values {
		total += v
	}
	if total <= 0 || r.W <= 0 || r.H <= 0 {
		return out
	}
	scale := float64(r.W) * float64(r.H) / total
	x, y, w, h := float64(r.X), float64(r.Y), float64(r.W), float64(r.H)
	worst := func(sum, largest, smallest, side float64) float64 {
		return max(side*side*largest/(sum*sum), sum*sum/(side*side*smallest))
	}
	for i := 0; i < len(values); {
		side := min(w, h)
		largest := values[i] * scale
		sum, j := largest, i+1
		ratio := worst(sum, largest, largest, side)
		for ; j < len(values); j++ {
			a := values[j] * scale
			next := worst(sum+a, largest, a, side)
			if next > ratio {
				break
			}
			sum, ratio = sum+a, next
		}
		// Items share the row by proportion, and the last row takes what is
		// left, so rounding never leaves a gap.
		row := values[i:j]
		rowTotal := 0.0
		for _, v := range row {
			rowTotal += v
		}
		if w >= h {
			// A column along the left edge.
			cw := sum / h
			if j == len(out) {
				cw = w
			}
			cy := y
			for k, v := range row {
				ch := v / rowTotal * h
				out[i+k] = ui.Rect{X: float32(x), Y: float32(cy), W: float32(cw), H: float32(ch)}
				cy += ch
			}
			x, w = x+cw, w-cw
		} else {
			// A row along the top edge.
			rh := sum / w
			if j == len(out) {
				rh = h
			}
			cx := x
			for k, v := range row {
				cw := v / rowTotal * w
				out[i+k] = ui.Rect{X: float32(cx), Y: float32(y), W: float32(cw), H: float32(rh)}
				cx += cw
			}
			y, h = y+rh, h-rh
		}
		i = j
	}
	return out
}

func inset(r ui.Rect, dx, dy float32) ui.Rect {
	return ui.Rect{X: r.X + dx, Y: r.Y + dy, W: max(0, r.W-2*dx), H: max(0, r.H-2*dy)}
}

func near(a, b ui.Rect) bool {
	d := func(x, y float32) bool { return math.Abs(float64(x-y)) < .5 }
	return d(a.X, b.X) && d(a.Y, b.Y) && d(a.W, b.W) && d(a.H, b.H)
}

func lerp(a, b ui.Rect, e float32) ui.Rect {
	return ui.Rect{X: a.X + (b.X-a.X)*e, Y: a.Y + (b.Y-a.Y)*e, W: a.W + (b.W-a.W)*e, H: a.H + (b.H-a.H)*e}
}

func contains(r ui.Rect, x, y float32) bool {
	return x >= r.X && y >= r.Y && x < r.X+r.W && y < r.Y+r.H
}

// hit returns the deepest tile at x, y, or -1.
func (v *view) hit(x, y float32) int {
	for i := len(v.tiles) - 1; i >= 0; i-- {
		if contains(v.tiles[i].rect, x, y) {
			return i
		}
	}
	return -1
}

// topOf returns the shown level's tile containing tile i.
func (v *view) topOf(i int) int {
	for v.tiles[i].parent >= 0 {
		i = v.tiles[i].parent
	}
	return i
}
