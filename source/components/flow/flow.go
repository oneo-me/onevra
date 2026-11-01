// Package flow draws a directed diagram, such as a workflow, a dependency
// graph or a plan of tasks, and lays it out automatically: nodes fall into
// ranks along the edges' direction, ranks are ordered to avoid crossings
// and edges run between them as orthogonal lines. Applications build each
// node's content and own selection and what opening a node means; the view
// owns layout, routing and the camera.
package flow

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// Direction is the way edges point.
type Direction uint8

const (
	// TopToBottom places sources at the top and edges pointing down.
	TopToBottom Direction = iota
	// LeftToRight places sources at the left and edges pointing right.
	LeftToRight
)

// Node is one card of the diagram. ID must be nonempty, unique and stable
// across frames; later duplicates are ignored. Label is its accessible
// name. Content builds the card; Card provides a styled default.
//
// Positive Width and Height fix the node's size; otherwise the layout
// measures what Content builds, laying the diagram out again when that
// size changes. Muted fades the node and its edges, as for skipped or
// inactive steps.
type Node struct {
	ID, Label     string
	Width, Height float32
	Muted         bool
	Content       func(c *ui.Context)
}

// Edge points from Source to Target. Edges to unknown nodes and self-loops
// are ignored; edges closing a cycle are drawn against the layout's
// direction. Label, when set, sits on the line. Zero Color and Width use
// the theme's soft line, 1.5 DIPs wide.
type Edge struct {
	Source, Target string
	Label          string
	Dashed         bool
	Color          ui.Color
	Width          float32
}

// Data is borrowed for the frame and never modified by the view.
type Data struct {
	Nodes []Node
	Edges []Edge
}

// Options supplies presentation and application-owned behavior.
//
// Label is the accessible description (a node and edge count by default).
// Height sets a fixed height; zero fills the parent. NodeGap separates
// nodes within a rank (default 40) and RankGap the ranks (default 72).
//
// Selected is the application's selected node, which the view rings and
// whose edges it emphasizes. Select is called on the UI thread when a node
// is clicked or activated from the keyboard, and with "" when empty space
// is clicked; Open when a node is double-clicked. Empty replaces the view
// when there are no nodes (EmptyText, default "—", otherwise). Overlay
// builds content in the top trailing corner above the diagram, such as
// zoom controls.
type Options struct {
	Label, EmptyText string
	Height           float32
	Direction        Direction
	NodeGap, RankGap float32
	Selected         string
	Select, Open     func(id string)
	Empty, Overlay   func(*ui.Context)
}

// State retains measured sizes, layout and camera between frames. Keep one
// per diagram; its zero value is ready to use. Nil state uses element-local
// storage.
type State struct {
	view
}

// Ready reports whether every node has been measured, laid out and shown.
func (s *State) Ready() bool { return s.ready }

// Zoom returns the camera's scale.
func (s *State) Zoom() float32 { return float32(s.camera.zoom) }

// Fit moves the camera to show the whole diagram.
func (s *State) Fit() { s.fit(false) }

// ZoomIn and ZoomOut scale the view around its center.
func (s *State) ZoomIn()  { s.zoomAt(s.width/2, s.height/2, zoomStep, true) }
func (s *State) ZoomOut() { s.zoomAt(s.width/2, s.height/2, 1/zoomStep, true) }

// Reveal moves the camera to center the node with id, if it is shown.
func (s *State) Reveal(id string) { s.reveal(id) }

// Position returns the top left corner of the node with id in diagram
// coordinates, and false if it is not laid out.
func (s *State) Position(id string) (x, y float32, ok bool) {
	i, ok := s.index[id]
	if !ok || i >= len(s.result.x) {
		return 0, 0, false
	}
	return float32(s.result.x[i]), float32(s.result.y[i]), true
}

// View builds the diagram. Drag empty space or two fingers on a touchpad to
// pan, use the wheel or a pinch to zoom around the pointer, and
// double-click empty space to zoom in, or out with Shift. Nodes take the
// keyboard focus in data order when the application can select or open
// them, and the camera follows the focused node.
func View(c *ui.Context, state *State, data Data, opts Options) ui.Element {
	t := c.Theme()
	root := ui.Box(c).Clip().MinWidth(0)
	if opts.Height > 0 {
		root.FillWidth().Height(opts.Height)
	} else {
		root.Fill()
	}
	if state == nil {
		state = ui.Local(root, "onevra.flow", func() State { return State{} })
	}
	v := &state.view
	v.update(data, opts, c.Preferences().ReduceMotion || c.Instant())
	root.Children(func() {
		if len(v.nodes) == 0 {
			ui.Column(c).Fill().AlignItems(ui.Center).Justify(ui.Center).Padding(t.Space(6)).Children(func() {
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
		label := opts.Label
		if label == "" {
			label = fmt.Sprintf("%d nodes, %d edges", len(v.nodes), len(v.edges))
		}
		canvas := ui.Box(c).Key("onevra.flow.canvas").Fill().Role(ui.RoleGroup).Label(label).Cursor(v.cursor())
		bounds := canvas.LayoutBounds()
		v.resize(float64(bounds.W), float64(bounds.H))
		v.navigation.Handle(c, canvas, v.navigate, v.input)
		v.build(c, canvas, data, opts)
		if opts.Overlay != nil {
			inset := t.Space(2)
			ui.Box(c).Absolute().Top(inset).Right(inset).Children(func() { opts.Overlay(c) })
		}
	})
	return root
}

// build places the nodes on the camera's plane and measures them for the
// next layout; edges paint beneath them.
func (v *view) build(c *ui.Context, canvas ui.Element, data Data, opts Options) {
	t := c.Theme()
	// A diagram shown again builds its nodes anew; they keep their last
	// size, rather than the empty one of their first frame, until measured.
	if mounted := ui.Local(canvas, "onevra.flow.mounted", func() bool { return false }); !*mounted {
		*mounted = true
		for i := range v.nodes {
			v.nodes[i].built = false
		}
	}
	v.prepare(c.Now())
	moving := v.advanceCamera(c.Now())
	resized := false
	services := c.Services()
	interactive := opts.Select != nil || opts.Open != nil
	canvas.Children(func() {
		// Absolute nodes size to their content within their parent, so the
		// plane leaves them room to the right and below.
		plane := ui.Box(c).Key("onevra.flow.plane").Absolute().Left(0).Top(0).
			Size(float32(v.result.width)+planeRoom, float32(v.result.height)+planeRoom).
			TransformOrigin(0, 0).
			Transform(ui.Scaling(float32(v.shown.zoom), float32(v.shown.zoom)).Then(ui.Translation(float32(v.shown.x), float32(v.shown.y))))
		if !v.ready {
			plane.Opacity(0)
		}
		colors := v.palette(t)
		plane.Draw(func(p *ui.Painter, r ui.Rect) { v.paintEdges(p, r, colors) })
		plane.DrawOver(func(p *ui.Painter, r ui.Rect) { v.paintSelection(p, r, colors) })
		plane.Children(func() {
			hovered := ""
			focused := ""
			for i := range v.nodes {
				n := &v.nodes[i]
				src := data.Nodes[n.source]
				el := ui.Box(c).Key(n.id).Absolute().Left(0).Top(0).Label(n.label)
				if src.Width > 0 {
					el.Width(src.Width)
				}
				if src.Height > 0 {
					el.Height(src.Height)
				}
				x, y := float32(v.result.x[i]), float32(v.result.y[i])
				switch {
				case v.ready && n.measured && v.snap:
					x, y = el.Jump("x", x), el.Jump("y", y)
				case v.ready && n.measured:
					x, y = el.AnimateTo("x", x, ui.Smooth), el.AnimateTo("y", y, ui.Smooth)
				}
				n.x, n.y = float64(x), float64(y)
				el.Translate(x, y)
				switch {
				case !n.measured:
					// Unmeasured nodes take their size before showing.
					el.Opacity(0)
				case n.muted:
					el.Opacity(mutedNodeOpacity)
				}
				if interactive {
					el.Focusable().Role(ui.RoleButton).Cursor(ui.CursorPointer)
					if el.Clicked() && opts.Select != nil {
						v.selected = n.id
						opts.Select(n.id)
					}
					if el.DoubleClicked() && opts.Open != nil {
						opts.Open(n.id)
					}
					if el.Focused() && el.FocusVisible() {
						focused = n.id
					}
				}
				if el.Hovered() {
					hovered = n.id
				}
				if src.Content != nil {
					el.Children(func() { src.Content(c) })
				}
				if n.built {
					b := el.LayoutBounds()
					resized = n.measure(float64(b.W), float64(b.H)) || resized
				}
				n.built = true
				if n.measured {
					// Content may change size in a frame nothing else
					// follows; the next one measures it.
					w, h := float32(n.width), float32(n.height)
					el.Draw(func(_ *ui.Painter, r ui.Rect) {
						if math.Abs(float64(r.W-w)) > .25 || math.Abs(float64(r.H-h)) > .25 {
							services.Invalidate()
						}
					})
				}
			}
			v.hovered = hovered
			if focused != "" && focused != v.focused {
				v.revealIfHidden(focused)
			}
			v.focused = focused
		})
	})
	// Nodes built for the first time are measured on the next frame, and
	// resized ones laid out again.
	if !v.ready || moving || resized || slices.ContainsFunc(v.nodes, func(n renderNode) bool { return !n.measured }) {
		c.AnimationFrame()
	}
}

type renderNode struct {
	id, label string
	source    int
	muted     bool
	fixedW    float64
	fixedH    float64
	// width and height are the measured size, valid once measured; built
	// nodes have a layout box from the previous frame to measure.
	width, height   float64
	measured, built bool
	// x and y are where the node shows this frame, which leads its laid
	// out place while it moves there.
	x, y float64
}

// measure records the node's size and reports whether it changed.
func (n *renderNode) measure(w, h float64) bool {
	if n.fixedW > 0 {
		w = n.fixedW
	}
	if n.fixedH > 0 {
		h = n.fixedH
	}
	if n.measured && math.Abs(n.width-w) <= .25 && math.Abs(n.height-h) <= .25 {
		return false
	}
	n.width, n.height, n.measured = w, h, true
	return true
}

func (n *renderNode) size() (float64, float64) {
	if n.measured {
		return n.width, n.height
	}
	w, h := n.fixedW, n.fixedH
	if w <= 0 {
		w = fallbackWidth
	}
	if h <= 0 {
		h = fallbackHeight
	}
	return w, h
}

type renderEdge struct {
	source, target int
	label          string
	dashed         bool
	color          ui.Color
	width          float32
}

// update reconciles application data with the retained nodes, keeping
// measured sizes by ID; layout follows in prepare, once sizes are known.
func (v *view) update(data Data, opts Options, reduce bool) {
	v.init()
	v.reduceMotion = reduce
	v.selected, v.selectFn, v.direction = opts.Selected, opts.Select, opts.Direction
	v.gaps = spacing{node: defaultNodeGap, rank: defaultRankGap, edge: edgeGap}
	if opts.NodeGap > 0 {
		v.gaps.node = float64(opts.NodeGap)
	}
	if opts.RankGap > 0 {
		v.gaps.rank = float64(opts.RankGap)
	}
	previous := make(map[string]renderNode, len(v.nodes))
	for _, n := range v.nodes {
		previous[n.id] = n
	}
	index := make(map[string]int, len(data.Nodes))
	nodes := v.nodes[:0:0]
	for i, n := range data.Nodes {
		if _, ok := index[n.ID]; ok || n.ID == "" {
			continue
		}
		index[n.ID] = len(nodes)
		rn := renderNode{id: n.ID, label: n.Label, source: i, muted: n.Muted, fixedW: float64(max(0, n.Width)), fixedH: float64(max(0, n.Height))}
		if old, ok := previous[n.ID]; ok {
			rn.width, rn.height, rn.measured, rn.built, rn.x, rn.y = old.width, old.height, old.measured, old.built, old.x, old.y
			if old.fixedW != rn.fixedW || old.fixedH != rn.fixedH {
				rn.measured = false
			}
		}
		if rn.label == "" {
			rn.label = n.ID
		}
		nodes = append(nodes, rn)
	}
	edges := make([]renderEdge, 0, len(data.Edges))
	for _, e := range data.Edges {
		s, ok1 := index[e.Source]
		d, ok2 := index[e.Target]
		if !ok1 || !ok2 || s == d {
			continue
		}
		edges = append(edges, renderEdge{source: s, target: d, label: e.Label, dashed: e.Dashed, color: e.Color, width: e.Width})
	}
	v.nodes, v.edges, v.index = nodes, edges, index
}

// prepare lays the diagram out again when its topology, node sizes or
// spacing change, and shows it once every node is measured and the view
// has a size. Nodes and the camera move to a new topology or spacing, and
// to sizes measured while they do; other new sizes, such as those of
// content settling after the diagram appears, take effect at once.
func (v *view) prepare(now time.Time) {
	var h, structure hasher
	structure.reset()
	structure.int(int(v.direction))
	structure.float(v.gaps.node)
	structure.float(v.gaps.rank)
	for _, n := range v.nodes {
		structure.string(n.id)
	}
	for _, e := range v.edges {
		structure.int(e.source)
		structure.int(e.target)
		dashed := byte(0)
		if e.dashed {
			dashed = 1
		}
		structure.byte(dashed)
	}
	h.sum = structure.sum
	all := true
	for i := range v.nodes {
		n := &v.nodes[i]
		w, ht := n.size()
		h.float(math.Round(w * 2))
		h.float(math.Round(ht * 2))
		all = all && n.measured
	}
	if structure.sum != v.structureKey {
		v.structureKey = structure.sum
		if v.ready {
			v.structureAt = now
		}
	}
	v.snap = false
	if h.sum != v.layoutKey || len(v.result.x) != len(v.nodes) {
		v.layoutKey = h.sum
		v.snap = now.Sub(v.structureAt) > relayoutMotion
		nodes := make([]layoutNode, len(v.nodes))
		for i := range v.nodes {
			w, ht := v.nodes[i].size()
			nodes[i] = layoutNode{w, ht}
		}
		edges := make([]layoutEdge, len(v.edges))
		for i, e := range v.edges {
			edges[i] = layoutEdge{e.source, e.target, e.dashed}
		}
		v.result = layout(nodes, edges, v.direction, v.gaps)
		if v.ready && !v.userMoved {
			v.fit(v.snap)
		}
	}
	if !v.ready && all && v.width > 0 {
		v.ready = true
		v.fit(true)
	}
}

// hasher is FNV-1a over the fields that decide whether the layout changed.
type hasher struct{ sum uint64 }

func (h *hasher) reset() { h.sum = 14695981039346656037 }

func (h *hasher) byte(b byte) {
	h.sum ^= uint64(b)
	h.sum *= 1099511628211
}

func (h *hasher) uint(v uint64) {
	for range 8 {
		h.byte(byte(v))
		v >>= 8
	}
}

func (h *hasher) string(s string) {
	for i := 0; i < len(s); i++ {
		h.byte(s[i])
	}
	h.byte(0)
}

func (h *hasher) int(v int)       { h.uint(uint64(v)) }
func (h *hasher) float(v float64) { h.uint(math.Float64bits(v)) }
