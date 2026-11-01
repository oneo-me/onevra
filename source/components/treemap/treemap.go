// Package treemap draws hierarchical values as nested, squarified rectangles
// whose areas follow their values, as disk-usage and budget views do.
// Applications own the tree, localized labels, formatting and what opening a
// leaf means; the view owns layout, drilling, hover and transitions.
package treemap

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// Node is one rectangle. ID must be nonempty, unique and stable across
// frames. A leaf's Value sizes it; a branch takes the total of its children,
// or its Value when larger, leaving the difference as unlabeled space. Nodes
// without a finite positive size are omitted. Caption is secondary text a
// large leaf shows below its size.
//
// Measure is a second quantity, such as growth, that shades leaves when
// Options.MeasureLabel is set; branches report their leaves' size-weighted
// mean. A nonzero Color tints the node and its descendants; other nodes
// inherit their parent's, and top-level nodes take a palette color by order.
//
// Style overrides how the node's own tile paints, and HoverStyle how it
// paints while the pointer is on it rather than on a descendant.
type Node struct {
	ID, Label, Caption string
	Value, Measure     float64
	Color              ui.Color
	Style, HoverStyle  Style
	Children           []Node
}

// Style overrides a tile's appearance; zero fields keep the tile's own.
// Background replaces the tinted gradient with a flat fill. Border, BorderWidth
// and BorderStyle replace the faint edge, such as with a dashed outline. A
// HoverStyle's fields blend from the tile's as the pointer arrives, and a
// border in it replaces the hover ring.
type Style struct {
	Background, Border ui.Color
	BorderWidth        float32
	BorderStyle        ui.BorderStyle
}

// State retains navigation between frames. Root is the branch drilled into,
// "" for the whole tree; an unknown Root, or one without children, resets to
// "". Active is the tile of the shown level the keyboard is on, "" for none.
// Keep one per view; its zero value is ready. Nil state uses element-local
// storage.
type State struct {
	Root, Active string
	view
}

// Detail describes a node for tips and assistive technology. Share is its
// part of its parent, or of the whole tree at the top level. Measure is NaN
// without Options.MeasureLabel.
type Detail struct {
	Path                 []Node
	Size, Share, Measure float64
}

// Options supplies presentation and application-owned content.
//
// Label names the chart and the first breadcrumb; Description guides
// assistive technology, such as through the keys. Height sets a fixed height
// for the breadcrumbs and tiles; zero fills the parent. HideHeader omits the
// breadcrumbs and total, so a host can drive State.Root itself. FormatValue formats sizes (nil uses numbers).
//
// MeasureLabel, when set, names Node.Measure and shades leaves by it within
// MeasureDomain, or the leaves' range when that is empty. FormatMeasure
// formats it (nil uses numbers).
//
// Clicking a tile of the shown level drills into it, and breadcrumbs lead
// back. Arrow keys move to the nearest tile in their direction, Home and End
// to the largest and smallest; Enter or Space drills in, and Escape or
// Backspace goes up a level. Opening a leaf calls Open on the UI thread.
//
// Menu, when set, returns the builder of a node's context menu, or nil for a
// node without one. The menu opens where the deepest tile is right-clicked,
// without drilling or choosing it, and for the active tile from the menu key
// or Shift+F10.
//
// As in charts, a tooltip follows the pointer for the deepest hovered node, or
// sits beside the active tile the keyboard chose while the view has the
// focus. Tip replaces its content, which by default shows the node's path,
// size, share, measure and caption. Empty replaces the tiles when nothing has a positive size
// (EmptyText, default "—", otherwise). AnimationDuration controls drilling and
// data transitions: zero uses 400 ms, negative disables them, and Reduced
// Motion and ui.Context.DisableAnimations always skip them.
type Options struct {
	Label, Description, EmptyText string
	Height                        float32
	HideHeader                    bool
	FormatValue                   func(float64) string
	MeasureLabel                  string
	MeasureDomain                 [2]float64
	FormatMeasure                 func(float64) string
	Open                          func(id string)
	Menu                          func(id string) func(m *ui.Menu)
	Tip                           func(c *ui.Context, d Detail)
	Empty                         func(*ui.Context)
	AnimationDuration             time.Duration
}

// Path returns the nodes from the top level down to the node with id, or nil
// when it is absent.
func Path(nodes []Node, id string) []Node {
	return copyPath(find(nodes, id))
}

// level is the shown level: the branch drilled into, its path and children,
// with the free space it leaves and its size.
type level struct {
	path   []*Node
	nodes  []Node
	color  ui.Color
	extra  float64
	parent float64
}

// View builds the treemap.
func View(c *ui.Context, state *State, nodes []Node, opts Options) ui.Element {
	t := c.Theme()
	root := ui.Column(c).MinWidth(0).Gap(t.Space(2))
	if opts.Height > 0 {
		root.FillWidth().Height(opts.Height)
	} else {
		root.Fill()
	}
	if state == nil {
		state = ui.Local(root, "onevra.treemap", func() State { return State{} })
	}
	v := &state.view
	v.begin(c, nodes, opts)
	v.pointerClicks(c, state, nodes, opts)
	shown := v.resolve(state, nodes)
	root.Children(func() {
		if !opts.HideHeader {
			if v.header(c, state, shown, opts) {
				shown = v.resolve(state, nodes)
			}
		}
		if v.total(shown.nodes) <= 0 {
			v.tiles, v.hovered = v.tiles[:0], -1
			ui.Column(c).FillWidth().Grow(1).MinHeight(0).AlignItems(ui.Center).Justify(ui.Center).Padding(t.Space(6)).Children(func() {
				if opts.Empty != nil {
					opts.Empty(c)
					return
				}
				text := opts.EmptyText
				if text == "" {
					text = "—"
				}
				components.Text(c, text).TextColor(t.TextMuted)
			})
			return
		}
		canvas := ui.Box(c).Key("onevra.treemap.canvas").FillWidth().Grow(1).MinHeight(0).Clip().
			Focusable().Label(opts.Label).Description(opts.Description)
		canvas.HandleInput(v.input)
		if opts.Menu != nil {
			// The input handler chose the node as the menu was asked for; it
			// stays the same when the menu builds again for the item chosen.
			canvas.ContextMenu(func(m *ui.Menu) {
				if v.menuID == "" {
					return
				}
				if build := opts.Menu(v.menuID); build != nil {
					build(m)
				}
			})
		}
		if v.keyboard(c, canvas, state, nodes, shown, opts) {
			shown = v.resolve(state, nodes)
		}
		r := canvas.LayoutBounds()
		if r.W <= 0 || r.H <= 0 {
			// The first frame has no size to lay out in: build again once laid out.
			canvas.Invisible()
			c.AnimationFrame()
			return
		}
		v.width, v.height = r.W, r.H
		v.place(c, state.Root, shown)
		x, y, over := canvas.PointerPosition()
		v.hovered = -1
		if over && !v.zooming {
			if v.hovered = v.hit(x, y); v.hovered >= 0 {
				canvas.Cursor(ui.CursorPointer)
			}
		}
		hoverID := ""
		if v.hovered >= 0 {
			hoverID = v.tiles[v.hovered].id
		}
		if hoverID != v.hoverID {
			v.hoverOut, v.hoverID, v.hoverStart = v.hoverID, hoverID, c.Now()
		}
		active := -1
		for i, tl := range v.tiles {
			if tl.parent < 0 && tl.id == state.Active {
				active = i
			}
		}
		if active < 0 {
			state.Active = ""
		} else if d, ok := v.detail(nodes, state.Active); ok {
			canvas.Value(v.describe(d))
		}
		focused := canvas.FocusVisible()
		canvas.Draw(func(p *ui.Painter, r ui.Rect) {
			if r.W != v.width || r.H != v.height {
				c.Invalidate()
			}
			v.draw(p, r, t, active, focused)
		})
		// As in charts, the tooltip follows the pointer while it moved last,
		// and sits beside the active tile while the keyboard did.
		if over && (!v.tipOver || x != v.tipX || y != v.tipY) {
			v.tipPointer = true
		} else if state.Active != v.activeID {
			v.tipPointer = false
		}
		v.tipX, v.tipY, v.tipOver, v.activeID = x, y, over, state.Active
		shownTip := -1
		switch {
		case v.tipPointer && v.hovered >= 0 && !v.pressed:
			shownTip = v.hovered
		case !v.tipPointer && active >= 0 && canvas.Focused() && !v.zooming:
			shownTip = active
		}
		if shownTip >= 0 {
			tl := v.tiles[shownTip]
			if d, ok := v.detail(nodes, tl.id); ok {
				canvas.Children(func() {
					var anchor ui.Element
					if v.tipPointer {
						anchor = ui.Box(c).Absolute().Left(x).Top(y).Size(0, 0)
					} else {
						anchor = ui.Box(c).Absolute().Left(tl.rect.X).Top(tl.rect.Y).Size(tl.rect.W, tl.rect.H)
					}
					v.tip(c, anchor.PassThrough(), d, tl.color, opts)
				})
			}
		}
	})
	return root
}

// resolve finds the shown level, resetting a Root that cannot be drilled into.
func (v *view) resolve(state *State, nodes []Node) level {
	path := find(nodes, state.Root)
	if state.Root != "" && (path == nil || !v.branch(path[len(path)-1])) {
		state.Root, path = "", nil
	}
	l := level{path: path, nodes: nodes, parent: v.total(nodes)}
	if len(path) > 0 {
		n := path[len(path)-1]
		l.nodes, l.extra, l.parent = n.Children, v.free(n), v.size(n)
		l.color = paletteColor(topIndex(nodes, path))
		for _, n := range path {
			if n.Color != (ui.Color{}) {
				l.color = n.Color
			}
		}
	}
	return l
}

// header shows the breadcrumbs to the shown level and its rolling total,
// reporting whether a crumb changed the root.
func (v *view) header(c *ui.Context, state *State, shown level, opts Options) bool {
	t := c.Theme()
	changed := false
	ui.Row(c).FillWidth().MinWidth(0).AlignItems(ui.Center).Gap(t.Space(3)).Children(func() {
		items := []string{opts.Label}
		for _, n := range shown.path {
			items = append(items, n.Label)
		}
		chosen := len(items) - 1
		crumbs := components.Breadcrumbs(c, items, &chosen)
		crumbs.Grow(1).Shrink(1)
		if crumbs.Changed() && chosen < len(shown.path) {
			state.Root = ""
			if chosen > 0 {
				state.Root = shown.path[chosen-1].ID
			}
			// Keep the keyboard on the branch the view came up from.
			state.Active, changed = shown.path[chosen].ID, true
			c.Invalidate()
		}
		total := ui.Box(c).Shrink(0)
		// The total rolls to its new value as the view drills or the data
		// changes.
		value := float64(total.AnimateWith("onevra.treemap.total", float32(shown.parent), v.duration, ui.EaseInOut))
		if v.duration == 0 || !finite(value) {
			value = shown.parent
		}
		total.Children(func() {
			components.Text(c, v.format(value)).FontSize(t.FontSize - 1).TextColor(t.TextMuted).NoWrap()
		})
	})
	return changed
}

// tip shows a node's detail in a tooltip beside anchor, styled as charts'.
func (v *view) tip(c *ui.Context, anchor ui.Element, d Detail, color ui.Color, opts Options) {
	t := c.Theme()
	ui.Overlay(c, func() {
		ui.Column(c).Absolute().PassThrough().Role(ui.RoleTooltip).
			AttachTo(anchor, ui.AnchorEnd, ui.AnchorStart).MarginStart(t.Space(3)).
			MinWidth(t.Space(30)).MaxWidth(t.Space(80)).Padding(t.Space(2), t.Space(2.5)).Gap(t.Space(1)).
			Radius(t.Radius).Material(glass.Glass{Tint: t.Background.Alpha(.92)}).Border(1, t.Border).
			TextColor(t.Text).FontSize(t.FontSize-1).
			Shadow(0, t.Space(1), t.Space(4), 0, ui.RGBA(0, 0, 0, .14)).Children(func() {
			if opts.Tip != nil {
				opts.Tip(c, d)
				return
			}
			n := d.Path[len(d.Path)-1]
			if len(d.Path) > 1 {
				components.Text(c, pathLabel(d.Path[:len(d.Path)-1])).TextColor(t.TextMuted).NoWrap().Ellipsis("…")
			}
			ui.Row(c).FillWidth().MinWidth(0).AlignItems(ui.Center).Gap(t.Space(1.5)).Children(func() {
				ui.Box(c).Size(8, 8).Shrink(0).Radius(3).Background(color)
				components.Text(c, n.Label).MinWidth(0).Shrink(1).NoWrap().Ellipsis("…")
				ui.Box(c).Grow(1)
				components.Text(c, v.format(d.Size)+" · "+percent(d.Share)).Bold().NoWrap()
			})
			if !math.IsNaN(d.Measure) {
				ui.Row(c).FillWidth().MinWidth(0).AlignItems(ui.Center).Gap(t.Space(1.5)).Children(func() {
					components.Text(c, opts.MeasureLabel).TextColor(t.TextMuted).MinWidth(0).Shrink(1).NoWrap().Ellipsis("…")
					ui.Box(c).Grow(1)
					components.Text(c, v.formatMeasure(d.Measure)).Bold().NoWrap()
				})
			}
			if n.Caption != "" {
				components.Text(c, n.Caption).TextColor(t.TextMuted)
			}
		})
	})
}

// detail describes the node with id.
func (v *view) detail(nodes []Node, id string) (Detail, bool) {
	path := find(nodes, id)
	if path == nil {
		return Detail{}, false
	}
	n := path[len(path)-1]
	d := Detail{Path: copyPath(path), Size: v.size(n), Measure: v.measureOrNaN(n)}
	parent := v.total(nodes)
	if len(path) > 1 {
		parent = v.size(path[len(path)-2])
	}
	if parent > 0 {
		d.Share = d.Size / parent
	}
	return d, true
}

func (v *view) describe(d Detail) string {
	parts := []string{pathLabel(d.Path), v.format(d.Size), percent(d.Share)}
	if !math.IsNaN(d.Measure) {
		parts = append(parts, v.measureLabel+": "+v.formatMeasure(d.Measure))
	}
	return strings.Join(parts, ", ")
}

// pointerClicks applies the clicks taken since the last frame to the tiles
// as drawn then: the shown level's tile under the pointer drills in, or opens
// as a leaf.
func (v *view) pointerClicks(c *ui.Context, state *State, nodes []Node, opts Options) {
	for _, k := range v.clicks {
		if i := v.hit(k.x, k.y); i >= 0 && !v.zooming {
			v.activate(c, state, nodes, v.tiles[v.topOf(i)].id, opts, false)
		}
	}
	v.clicks = v.clicks[:0]
}

// activate drills into the branch with id, or opens it as a leaf. Drilling by
// keyboard puts the keyboard on the branch's largest child.
func (v *view) activate(c *ui.Context, state *State, nodes []Node, id string, opts Options, keyboard bool) {
	d, ok := v.detail(nodes, id)
	if !ok {
		return
	}
	n := find(nodes, id)
	branch := n[len(n)-1]
	if !v.branch(branch) {
		state.Active = id
		if opts.Open != nil {
			opts.Open(id)
		}
		return
	}
	state.Root, state.Active = id, ""
	if o := v.order(branch.Children, 0); keyboard && len(o) > 0 {
		state.Active = branch.Children[o[0]].ID
	}
	c.Announce(v.describe(d))
	c.Invalidate()
}

// keyboard moves the active tile and drills with the keys, reporting whether
// the root changed.
func (v *view) keyboard(c *ui.Context, canvas ui.Element, state *State, nodes []Node, shown level, opts Options) bool {
	root := state.Root
	var tops []int
	current := -1
	for i, tl := range v.tiles {
		if tl.parent < 0 {
			tops = append(tops, i)
			if tl.id == state.Active {
				current = i
			}
		}
	}
	to := -2
	switch {
	case canvas.Shortcut(0, ui.KeyRight):
		to = v.nearest(tops, current, 1, 0)
	case canvas.Shortcut(0, ui.KeyLeft):
		to = v.nearest(tops, current, -1, 0)
	case canvas.Shortcut(0, ui.KeyDown):
		to = v.nearest(tops, current, 0, 1)
	case canvas.Shortcut(0, ui.KeyUp):
		to = v.nearest(tops, current, 0, -1)
	case canvas.Shortcut(0, ui.KeyHome):
		to = -1
		if len(tops) > 0 {
			to = tops[0]
		}
	case canvas.Shortcut(0, ui.KeyEnd):
		to = -1
		if len(tops) > 0 {
			to = tops[len(tops)-1]
		}
	case canvas.Shortcut(0, ui.KeyEnter), canvas.Shortcut(0, ui.KeySpace):
		if current >= 0 {
			v.activate(c, state, nodes, state.Active, opts, true)
		}
		return state.Root != root
	case canvas.Shortcut(0, ui.KeyEscape), canvas.Shortcut(0, ui.KeyBackspace):
		if len(shown.path) == 0 {
			state.Active = ""
		} else {
			state.Active, state.Root = state.Root, ""
			if len(shown.path) > 1 {
				state.Root = shown.path[len(shown.path)-2].ID
			}
		}
	default:
		return false
	}
	if to >= 0 {
		state.Active = v.tiles[to].id
	}
	if d, ok := v.detail(nodes, state.Active); ok {
		c.Announce(v.describe(d))
	}
	c.Invalidate()
	return state.Root != root
}

// nearest returns the top-level tile nearest current in the direction dx, dy,
// or current when none lies that way; without current, the largest tile.
func (v *view) nearest(tops []int, current int, dx, dy float32) int {
	if current < 0 {
		if len(tops) == 0 {
			return -1
		}
		return tops[0]
	}
	from := v.tiles[current].rect
	fx, fy := from.X+from.W/2, from.Y+from.H/2
	best, score := current, float32(math.Inf(1))
	for _, i := range tops {
		r := v.tiles[i].rect
		ox, oy := r.X+r.W/2-fx, r.Y+r.H/2-fy
		along, across := ox*dx+oy*dy, ox*dy-oy*dx
		if i == current || along <= 0 {
			continue
		}
		// Prefer tiles straight ahead over nearer ones off to the side.
		if s := along + 2*float32(math.Abs(float64(across))); s < score {
			best, score = i, s
		}
	}
	return best
}

var palette = [...]ui.Color{
	ui.Hex("#5b8ff0"), ui.Hex("#e8a33d"), ui.Hex("#3fb68b"), ui.Hex("#a07ce0"),
	ui.Hex("#7d93ab"), ui.Hex("#e0707f"), ui.Hex("#3aa9bd"), ui.Hex("#b58f5e"),
}

func paletteColor(index int) ui.Color { return palette[index%len(palette)] }

// find returns the path from the top level to the node with id.
func find(nodes []Node, id string) []*Node {
	if id == "" {
		return nil
	}
	var path []*Node
	var walk func([]Node) bool
	walk = func(list []Node) bool {
		for i := range list {
			path = append(path, &list[i])
			if list[i].ID == id || walk(list[i].Children) {
				return true
			}
			path = path[:len(path)-1]
		}
		return false
	}
	if walk(nodes) {
		return path
	}
	return nil
}

// topIndex returns the index of the top-level node a path starts from.
func topIndex(nodes []Node, path []*Node) int {
	for i := range nodes {
		if &nodes[i] == path[0] {
			return i
		}
	}
	return 0
}

func copyPath(path []*Node) []Node {
	if len(path) == 0 {
		return nil
	}
	out := make([]Node, len(path))
	for i, n := range path {
		out[i] = *n
	}
	return out
}

func pathLabel(path []Node) string {
	labels := make([]string, len(path))
	for i, n := range path {
		labels[i] = n.Label
	}
	return strings.Join(labels, " / ")
}

func percent(share float64) string {
	p := share * 100
	if p >= 10 || p == 0 {
		return strconv.FormatFloat(p, 'f', 0, 64) + "%"
	}
	return strconv.FormatFloat(p, 'f', 1, 64) + "%"
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func (v *view) format(value float64) string {
	if v.formatValue != nil {
		return v.formatValue(value)
	}
	return strconv.FormatFloat(value, 'g', 5, 64)
}

func (v *view) formatMeasure(value float64) string {
	if v.formatMeasureFn != nil {
		return v.formatMeasureFn(value)
	}
	return strconv.FormatFloat(value, 'g', 4, 64)
}
