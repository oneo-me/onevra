package gallery

import (
	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// Positions and choices belong to this demo page and survive navigation, not
// app restarts.
type glassDemo struct {
	x, y float32
	// edge indexes glassEdges, what separates the bars from the content.
	edge int
	// clarity clears the draggable pane, a pane large enough to frost.
	clarity float64
}

// glassEdges maps the edge choice to the bars' styles.
var glassEdges = []components.ScrollEdgeStyle{components.ScrollEdgeGlass, components.ScrollEdgeSoft, components.ScrollEdgeHard, components.ScrollEdgeBlur}

// glassTiles color the rows that scroll under the bars.
var glassTiles = []ui.Color{ui.Hex("#ef4444"), ui.Hex("#f59e0b"), ui.Hex("#10b981"), ui.Hex("#06b6d4"), ui.Hex("#6366f1"), ui.Hex("#ec4899")}

func exampleGlass(c *ui.Context, s *demoState) {
	bitmap, err := galleryImage()
	if err != nil {
		components.Text(c, tr(c, "unableToLoadAvatarJpeg")).TextColor(c.Theme().TextMuted)
		return
	}
	glassPreview(c, s, bitmap)
}

func glassPreview(c *ui.Context, s *demoState, bitmap *ui.Bitmap) {
	ui.Row(c).Gap(12).AlignItems(ui.Center).Wrap().Children(func() {
		components.Tabs(c, &s.index, components.TabsSegmented,
			tr(c, "regular"), tr(c, "clearGlassStyle"), tr(c, "tinted")).List.Label(tr(c, "componentGlass"))
		components.Tabs(c, &s.glass.edge, components.TabsSegmented,
			tr(c, "glassBar"), tr(c, "glassSoftEdge"), tr(c, "glassHardEdge"), tr(c, "glassProgressiveBlur")).
			List.Label(tr(c, "glassUnderBars"))
		if components.Button(c, tr(c, "resetPosition"), components.ButtonOutline).Clicked() {
			s.glass.x, s.glass.y = 0, 0
		}
	})
	ui.Row(c).FillWidth().MaxWidth(c.Theme().Space(100)).Gap(c.Theme().Space(3)).AlignItems(ui.Center).Children(func() {
		name := tr(c, "glassClarity")
		components.Text(c, name).Shrink(0).SingleLine()
		components.Slider(c, &s.glass.clarity, 0, 1).Label(name).Grow(1).MinWidth(0)
	})
	t := c.Theme()
	canvas := ui.Box(c).FillWidth().Height(420).Radius(20).Clip().Border(1, t.Border).Label(tr(c, "glassCanvas"))
	previousWidth := canvas.Bounds().W
	// Bounds is the previous layout; settle initial layout and window resizing.
	canvas.Draw(func(_ *ui.Painter, r ui.Rect) {
		if r.W != previousWidth {
			c.Invalidate()
		}
	})
	material := glass.Glass{}
	color := t.Text
	switch s.index {
	case 1:
		material.Style = glass.Clear
		color = ui.RGB(15, 23, 42)
	case 2:
		material.Tint = t.Accent
		color = t.AccentText
	}
	top := components.ScrollEdgeBarOptions{Style: glassEdges[s.glass.edge], Glass: glass.Glass{Style: material.Style}}
	bottom := top
	bottom.Bottom = true
	inset := components.ScrollEdgeBarExtent(c, top)
	canvas.Children(func() {
		// Paint the content first: materials sample what is beneath them. It
		// starts clear of the bars and scrolls under them.
		components.ScrollView(c, components.ScrollViewOptions{}).Fill().Padding(inset+12, 16).
			ScrollbarInsets(inset, 0).Gap(12).Children(func() {
			components.Image(c, bitmap).FillWidth().Height(220).Radius(12).
				Fit(ui.Cover).Label(tr(c, "authorAvatar"))
			for i := range 8 {
				ui.Row(c).Gap(14).AlignItems(ui.Center).Children(func() {
					a, b := glassTiles[i%len(glassTiles)], glassTiles[(i+2)%len(glassTiles)]
					ui.Box(c).Size(120, 72).Radius(12).Shrink(0).Gradient(a, b, 135)
					components.Text(c, tr(c, "glassContentHint")).TextColor(t.TextMuted).Grow(1)
				})
			}
		})
		button := material
		button.Interactive = true
		components.ScrollEdgeBar(c, top, func() {
			glassButtons(c, button, color, []glassAction{{lucide.ChevronLeft, tr(c, "stepperBack")}, {lucide.Star, tr(c, "favorite")}, {lucide.Share, tr(c, "share")}})
			ui.Box(c).Grow(1)
			done := ui.Row(c).Padding(8, 16).Radius(20).
				Material(glass.Glass{Style: material.Style, Tint: t.Accent, Interactive: true}).Children(func() {
				components.Text(c, tr(c, "done")).Bold().TextColor(t.AccentText)
			})
			if done.Clicked() {
				c.AddToast(ui.Toast{ID: "gallery-glass", Title: tr(c, "done")})
			}
		}).Label(tr(c, "toolbar"))
		components.ScrollEdgeBar(c, bottom, func() {
			ui.Box(c).Grow(1)
			glassButtons(c, button, color, []glassAction{{lucide.House, tr(c, "home")}, {lucide.Search, tr(c, "search")}, {lucide.Settings, tr(c, "settings")}})
			ui.Box(c).Grow(1)
		}).Label(tr(c, "glassBottomBar"))
		width := min(float32(220), max(0, canvas.Bounds().W-32))
		maxX := max(0, canvas.Bounds().W-width-32)
		paneMaterial := material
		paneMaterial.Clarity = float32(s.glass.clarity)
		pane := ui.Column(c).Absolute().Width(width).Height(glassPaneHeight).
			Radius(30).Material(paneMaterial).Center().Gap(8).Label(tr(c, "draggableGlass"))
		if dx, dy, ok := pane.Dragged(); ok {
			s.glass.x += dx
			s.glass.y += dy
		}
		// Keep the entire drag target between the bars and reachable after
		// dragging or resizing.
		s.glass.x = min(maxX, max(0, s.glass.x))
		s.glass.y = min(max(0, 420-2*(inset+12)-glassPaneHeight), max(0, s.glass.y))
		pane.Left(16 + s.glass.x).Top(inset + 12 + s.glass.y).Children(func() {
			components.Text(c, tr(c, "liquidGlass")).Bold().TextColor(color)
			components.Text(c, tr(c, "onevraMygo")).TextColor(color)
		})
	})
}

// glassPaneHeight is tall enough for Clarity, which takes full effect on
// panes from 200 DIPs.
const glassPaneHeight = 200

type glassAction struct {
	icon  func() *ui.SVG
	label string
}

// glassButtons builds round glass buttons that toast their labels.
func glassButtons(c *ui.Context, material glass.Glass, color ui.Color, actions []glassAction) {
	for _, action := range actions {
		b := ui.Box(c).Size(40, 40).Radius(20).Center().Material(material).Label(action.label).Children(func() {
			components.Icon(c, action.icon).Size(18, 18).TextColor(color)
		})
		if b.Clicked() {
			c.AddToast(ui.Toast{ID: "gallery-glass", Title: action.label})
		}
	}
}
