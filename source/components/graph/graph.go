// Package graph draws an interactive force-directed relationship graph.
// Applications own nodes, edges, localized labels and what opening a node
// means; the view owns layout, camera and hover presentation.
package graph

import (
	"fmt"
	"hash/fnv"
	"math"
	"slices"
	"sort"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// Node is one vertex. ID must be nonempty, unique and stable across frames;
// later duplicates are ignored. Group places related nodes near each other
// initially and colors them by group; an empty Group uses the palette's
// neutral node color, or its isolated color for a node without edges.
// Weight sizes the node (zero uses its edge count) and Size, when positive,
// sets its diameter instead. Nonzero Color and LabelColor override the
// palette. Openable nodes take double-clicks, reported to Options.Open.
type Node struct {
	ID, Label         string
	Group             string
	Weight            float64
	Size              float32
	Color, LabelColor ui.Color
	Openable          bool
}

// EdgeStyle describes an edge's layout and appearance. Zero values use the
// defaults: Forces.LinkDistance, a strength weakened by the ends' degrees,
// width 1, the palette's edge opacity, a hover peak opacity of 0.92 and the
// source node's color.
type EdgeStyle struct {
	Distance, Strength    float64
	Width                 float32
	Opacity, HoverOpacity float32
	Color                 ui.Color
}

// Edge joins two node IDs. Edges to unknown nodes and self-loops are ignored.
type Edge struct {
	Source, Target string
	Style          EdgeStyle
}

// Data is borrowed for the frame and never modified by the view.
type Data struct {
	Nodes []Node
	Edges []Edge
}

// Palette colors nodes and edges that do not set their own. Muted nodes and
// edges fade toward Node while another node is hovered. A zero Background
// leaves the host's surface visible.
type Palette struct {
	Node, NodeLabel         ui.Color
	Isolated, IsolatedLabel ui.Color
	EdgeOpacity             float32
	Background              ui.Color
}

// DefaultPalette returns the built-in light or dark palette.
func DefaultPalette(dark bool) Palette {
	if dark {
		return Palette{
			Node: ui.Hex("#9aa7b3"), NodeLabel: ui.Hex("#d4dbe2"),
			Isolated: ui.Hex("#7d8894"), IsolatedLabel: ui.Hex("#b8c0c8"), EdgeOpacity: .25,
		}
	}
	return Palette{
		Node: ui.Hex("#626a70"), NodeLabel: ui.Hex("#3f474d"),
		Isolated: ui.Hex("#7f878d"), IsolatedLabel: ui.Hex("#596168"), EdgeOpacity: .2,
	}
}

// GroupColor colors group index of count, with groups sorted by name.
type GroupColor func(group string, index, count int, dark bool) (fill, label ui.Color)

// DefaultGroupColor spaces group hues evenly around the color wheel, so any
// two groups stay distinguishable, and varies saturation and lightness by
// name.
func DefaultGroupColor(group string, index, count int, dark bool) (fill, label ui.Color) {
	hash := hashString(group)
	hue := math.Round(math.Mod(208+float64(index)*360/float64(max(1, count)), 360)*10) / 10
	saturation := float64(48 + (hash>>8)%28)
	lightness := float64(36 + (hash>>16)%18)
	labelSaturation, labelLightness := 38.0, 28.0
	if dark {
		saturation = float64(50 + (hash>>8)%28)
		lightness = float64(56 + (hash>>16)%18)
		labelSaturation, labelLightness = 28, 82
	}
	return hsl(hue, saturation, lightness), hsl(hue, labelSaturation, labelLightness)
}

// Options supplies presentation and application-owned content.
//
// Label is the accessible description (a node and edge count by default).
// Height sets a fixed height; zero fills the parent. Nil Palette uses
// DefaultPalette and nil GroupColor uses DefaultGroupColor; both are
// re-evaluated only when the data, theme or palette change. HoverDepth is how
// many edges away from the hovered node stay emphasized (default 1).
//
// Open is called on the UI thread, during input handling, when an Openable
// node is double-clicked; a frame follows. Empty replaces the view when there
// are no nodes (EmptyText, default "—", otherwise). Overlay builds content in
// the top trailing corner above the graph, such as view settings.
type Options struct {
	Label, EmptyText string
	Height           float32
	Palette          *Palette
	GroupColor       GroupColor
	Forces           Forces
	HoverDepth       int
	Open             func(id string)
	Empty            func(*ui.Context)
	Overlay          func(*ui.Context)
}

// State retains layout, camera and hover between frames. Keep one per graph;
// its zero value is ready to use. Nil state uses element-local storage.
type State struct {
	graphView
}

// Hovered returns the ID of the node under the pointer, or "".
func (s *State) Hovered() string {
	if s.hovered < 0 || s.hovered >= len(s.nodes) {
		return ""
	}
	return s.nodes[s.hovered].id
}

// Settled reports whether the layout has come to rest.
func (s *State) Settled() bool { return s.sim == nil || !s.sim.running }

// Zoom returns the camera's scale.
func (s *State) Zoom() float32 { return float32(s.camera.zoom) }

// Fit moves the camera to show the whole graph.
func (s *State) Fit() { s.fit(false) }

// View builds the graph. Drag empty space or two fingers on a touchpad to
// pan, drag nodes to move them, and use the wheel or a pinch to zoom around
// the pointer; double-clicking empty space zooms in, or out with Shift.
func View(c *ui.Context, state *State, data Data, opts Options) ui.Element {
	t := c.Theme()
	root := ui.Box(c).Clip().MinWidth(0)
	if opts.Height > 0 {
		root.FillWidth().Height(opts.Height)
	} else {
		root.Fill()
	}
	if state == nil {
		state = ui.Local(root, "onevra.graph", func() State { return State{} })
	}
	palette := DefaultPalette(t.Dark)
	if opts.Palette != nil {
		palette = *opts.Palette
	}
	if palette.Background != (ui.Color{}) {
		root.Background(palette.Background)
	}
	reduce := c.Preferences().ReduceMotion || c.Instant()
	g := &state.graphView
	g.update(data, opts, palette, t.Dark, reduce)
	g.measureLabels(c)
	root.Children(func() {
		if len(g.nodes) == 0 {
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
			label = fmt.Sprintf("%d nodes, %d edges", len(g.nodes), len(g.edges))
		}
		canvas := ui.Box(c).Key("onevra.graph.canvas").Fill().Role(ui.RoleImage).Label(label).Cursor(g.cursor())
		if !canvas.Hovered() && g.gesture == gestureNone {
			g.setHovered(-1)
		}
		open := opts.Open
		g.navigation.Handle(c, canvas, g.navigate, func(ev ui.InputEvent) bool { return g.input(ev, open) })
		canvas.Draw(func(p *ui.Painter, r ui.Rect) { g.draw(p, r) })
		if opts.Overlay != nil {
			inset := t.Space(2)
			ui.Box(c).Absolute().Top(inset).Right(inset).Children(func() { opts.Overlay(c) })
		}
	})
	return root
}

type renderNode struct {
	id, label        string
	fill, labelColor ui.Color
	diameter         float32
	openable         bool
	fitted           string
	fittedWidth      float32
	measuredFor      string
}

type renderEdge struct {
	source, target int
	color          ui.Color
	opacity, hover float32
	width          float32
}

// update reconciles application data with the retained layout. Topology and
// layout options restart the simulation from the current positions; colors
// alone repaint in place, keeping the layout, camera and hover.
func (g *graphView) update(data Data, opts Options, palette Palette, dark, reduce bool) {
	g.init()
	g.reduceMotion = reduce
	g.hoverDepth = max(1, opts.HoverDepth)
	index := make(map[string]int, len(data.Nodes))
	nodes := make([]Node, 0, len(data.Nodes))
	for _, n := range data.Nodes {
		if _, ok := index[n.ID]; ok || n.ID == "" {
			continue
		}
		index[n.ID] = len(nodes)
		nodes = append(nodes, n)
	}
	type link struct {
		edge           Edge
		source, target int
	}
	links := make([]link, 0, len(data.Edges))
	degree := make([]int, len(nodes))
	for _, e := range data.Edges {
		s, ok1 := index[e.Source]
		d, ok2 := index[e.Target]
		if !ok1 || !ok2 || s == d {
			continue
		}
		links = append(links, link{e, s, d})
		degree[s]++
		degree[d]++
	}

	var layout hasher
	layout.reset()
	for _, n := range nodes {
		layout.string(n.ID)
		layout.string(n.Group)
	}
	for _, l := range links {
		layout.int(l.source)
		layout.int(l.target)
		layout.float(l.edge.Style.Distance)
		layout.float(l.edge.Style.Strength)
	}
	f := opts.Forces
	for _, v := range [...]float64{f.Center, f.Repel, f.Collision, f.CollisionStrength, f.LinkDistance} {
		layout.float(v)
	}
	layoutKey := layout.sum

	var look hasher
	look.reset()
	for _, n := range nodes {
		look.string(n.Label)
		look.float(n.Weight)
		look.float(float64(n.Size))
		look.color(n.Color)
		look.color(n.LabelColor)
		look.bool(n.Openable)
	}
	for _, l := range links {
		st := l.edge.Style
		for _, v := range [...]float32{st.Width, st.Opacity, st.HoverOpacity} {
			look.float(float64(v))
		}
		look.color(st.Color)
	}
	for _, c := range [...]ui.Color{palette.Node, palette.NodeLabel, palette.Isolated, palette.IsolatedLabel, palette.Background} {
		look.color(c)
	}
	look.float(float64(palette.EdgeOpacity))
	look.bool(dark)
	look.bool(opts.GroupColor != nil)
	lookKey := look.sum ^ layoutKey

	if layoutKey == g.layoutKey && lookKey == g.lookKey && (g.sim != nil) == (len(nodes) > 0) {
		return
	}

	groupColor := opts.GroupColor
	if groupColor == nil {
		groupColor = DefaultGroupColor
	}
	var groups []string
	for _, n := range nodes {
		if n.Group != "" {
			groups = append(groups, n.Group)
		}
	}
	sort.Strings(groups)
	groups = slices.Compact(groups)
	groupIndex := make(map[string]int, len(groups))
	for i, name := range groups {
		groupIndex[name] = i
	}
	type colors struct{ fill, label ui.Color }
	groupColors := make([]colors, len(groups))
	for i, name := range groups {
		groupColors[i].fill, groupColors[i].label = groupColor(name, i, len(groups), dark)
	}

	previous := make(map[string]renderNode, len(g.nodes))
	for _, n := range g.nodes {
		previous[n.id] = n
	}
	hoveredID, previousHoveredID, gestureID := g.nodeID(g.hovered), g.nodeID(g.previousHovered), ""
	if g.gesture == gestureNode || g.gesture == gestureNodePending {
		gestureID = g.nodeID(g.dragIndex)
	}
	render := make([]renderNode, len(nodes))
	for i, n := range nodes {
		fill, labelColor := palette.Node, palette.NodeLabel
		if gi, ok := groupIndex[n.Group]; ok {
			fill, labelColor = groupColors[gi].fill, groupColors[gi].label
		} else if degree[i] == 0 {
			fill, labelColor = palette.Isolated, palette.IsolatedLabel
		}
		if n.Color != (ui.Color{}) {
			fill = n.Color
		}
		if n.LabelColor != (ui.Color{}) {
			labelColor = n.LabelColor
		}
		weight := n.Weight
		if weight == 0 {
			weight = float64(degree[i])
		}
		diameter := n.Size
		if diameter <= 0 {
			diameter = nodeDiameter(weight)
		}
		rn := renderNode{id: n.ID, label: n.Label, fill: fill, labelColor: labelColor, diameter: diameter, openable: n.Openable}
		if old, ok := previous[n.ID]; ok && old.measuredFor == n.Label {
			rn.fitted, rn.fittedWidth, rn.measuredFor = old.fitted, old.fittedWidth, old.measuredFor
		}
		render[i] = rn
	}
	edges := make([]renderEdge, len(links))
	for i, l := range links {
		st := l.edge.Style
		e := renderEdge{source: l.source, target: l.target, color: st.Color, opacity: st.Opacity, hover: st.HoverOpacity, width: st.Width}
		if e.color == (ui.Color{}) {
			if gi, ok := groupIndex[nodes[l.source].Group]; ok {
				e.color = groupColors[gi].fill
			} else {
				e.color = palette.Node
			}
		}
		if e.opacity <= 0 {
			e.opacity = palette.EdgeOpacity
		}
		if e.hover <= 0 {
			e.hover = hoverEdgePeakOpacity
		}
		if e.width <= 0 {
			e.width = 1
		}
		edges[i] = e
	}
	g.nodes, g.edges, g.muted, g.lookKey = render, edges, palette.Node, lookKey
	g.adjacency = adjacency(len(render), edges)
	g.hovered, g.previousHovered = g.indexOf(hoveredID), g.indexOf(previousHoveredID)
	g.refreshHover()
	if gestureID != "" {
		if i := g.indexOf(gestureID); i >= 0 {
			g.dragIndex = i
		} else {
			g.gesture = gestureNone
		}
	}
	if layoutKey == g.layoutKey && g.sim != nil {
		return
	}
	g.layoutKey = layoutKey
	if len(nodes) == 0 {
		g.sim = nil
		return
	}

	springs := make([]springInput, len(links))
	for i, l := range links {
		springs[i] = springInput{source: l.source, target: l.target, distance: l.edge.Style.Distance, strength: l.edge.Style.Strength}
	}
	bodies := make([]body, len(nodes))
	for i, n := range nodes {
		cluster := -1
		if gi, ok := groupIndex[n.Group]; ok {
			cluster = gi
		}
		bodies[i].x, bodies[i].y = initialPosition(n.ID, len(nodes), cluster, len(groups))
	}
	defer func() {
		g.ids = g.ids[:0]
		for _, n := range nodes {
			g.ids = append(g.ids, n.ID)
		}
	}()
	if g.sim == nil {
		g.sim = newSimulation(bodies, springs, opts.Forces)
		g.ready, g.ticked, g.autoFit = false, 0, false
		return
	}
	// Existing nodes keep their place and motion; new ones start beside
	// their neighbours.
	known := make(map[string]body, len(g.ids))
	for i, id := range g.ids {
		known[id] = g.sim.bodies[i]
	}
	for i, n := range nodes {
		if b, ok := known[n.ID]; ok {
			bodies[i] = b
		}
	}
	for i, n := range nodes {
		if _, ok := known[n.ID]; ok {
			continue
		}
		var sx, sy float64
		count := 0
		for _, other := range g.adjacency[i] {
			if _, ok := known[nodes[other].ID]; ok {
				sx += bodies[other].x
				sy += bodies[other].y
				count++
			}
		}
		if count > 0 {
			angle := math.Atan2(bodies[i].y, bodies[i].x)
			bodies[i].x = sx/float64(count) + math.Cos(angle)*newNodeOffset
			bodies[i].y = sy/float64(count) + math.Sin(angle)*newNodeOffset
		}
	}
	g.sim.bodies = bodies
	g.sim.forces = opts.Forces.resolved()
	g.sim.setSprings(springs)
	g.sim.reheat()
	if g.ready {
		if reduce {
			g.sim.settle()
		} else {
			g.sim.step(1)
		}
	}
}

func (g *graphView) nodeID(i int) string {
	if i < 0 || i >= len(g.nodes) {
		return ""
	}
	return g.nodes[i].id
}

func (g *graphView) indexOf(id string) int {
	if id == "" {
		return -1
	}
	for i, n := range g.nodes {
		if n.id == id {
			return i
		}
	}
	return -1
}

// nodeDiameter grows with weight along a square-root curve between 12 and
// 44 DIPs.
func nodeDiameter(weight float64) float32 {
	radius := max(6, min(2.5*math.Sqrt(max(0, weight)+1), 22))
	return float32(radius * 2)
}

// initialPosition seeds a node deterministically from its ID in a disk whose
// area grows with the node count; grouped nodes start in their group's
// sector so related nodes need not cross the whole graph to meet.
func initialPosition(id string, count, cluster, clusters int) (float64, float64) {
	hash := hashString(id)
	angle := float64(hash&0xffff) / 0x10000 * 2 * math.Pi
	sample := (float64((hash>>16)&0xffff) + .5) / 0x10000
	outer := math.Sqrt(float64(max(1, count)) * initialAreaPerNode / math.Pi)
	if cluster < 0 || clusters <= 1 {
		r := math.Sqrt(sample) * outer
		return math.Cos(angle) * r, math.Sin(angle) * r
	}
	sector := float64(cluster) / float64(clusters) * 2 * math.Pi
	ring := outer * .72
	perCluster := max(1, float64(count)/float64(clusters))
	radius := max(64, math.Sqrt(perCluster*initialAreaPerNode/math.Pi)*.95)
	r := math.Sqrt(sample) * radius
	return math.Cos(sector)*ring + math.Cos(angle)*r, math.Sin(sector)*ring + math.Sin(angle)*r
}

// hasher is FNV-1a over the fields that decide whether data changed,
// written without formatting so large graphs compare cheaply each build.
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

func (h *hasher) bool(v bool) {
	if v {
		h.byte(1)
	} else {
		h.byte(0)
	}
}

func (h *hasher) color(c ui.Color) {
	h.uint(uint64(c.R) | uint64(c.G)<<8 | uint64(c.B)<<16 | uint64(c.A)<<24)
	if c != c.SRGB() {
		// Wide-gamut components are unexported; formatting reaches them.
		h.string(fmt.Sprint(c))
	}
}

func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func hsl(hue, saturation, lightness float64) ui.Color {
	h := math.Mod(math.Mod(hue, 360)+360, 360) / 60
	s := max(0, min(saturation/100, 1))
	l := max(0, min(lightness/100, 1))
	chroma := (1 - math.Abs(2*l-1)) * s
	second := chroma * (1 - math.Abs(math.Mod(h, 2)-1))
	var r, g, b float64
	switch {
	case h < 1:
		r, g = chroma, second
	case h < 2:
		r, g = second, chroma
	case h < 3:
		g, b = chroma, second
	case h < 4:
		g, b = second, chroma
	case h < 5:
		r, b = second, chroma
	default:
		r, b = chroma, second
	}
	m := l - chroma/2
	channel := func(v float64) uint8 { return uint8(math.Round((v + m) * 255)) }
	return ui.RGB(channel(r), channel(g), channel(b))
}
