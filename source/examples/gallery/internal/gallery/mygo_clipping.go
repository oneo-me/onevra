package gallery

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// starPath is the SVG path data of a star in a box of 100 by 100.
const starPath = "M50 4 L62 38 L98 38 L69 60 L79 95 L50 74 L21 95 L31 60 L2 38 L38 38 Z"

// exampleClipping clips and masks elements, gives them a shape for the pointer
// and orders them among their siblings and in depth.
func exampleClipping(c *ui.Context, s *demoState) {
	t := c.Theme()
	size := t.Space(28)
	tile := func(name string) ui.Element {
		return ui.Box(c).Size(size, size).Label(name).Center().
			LinearGradient(ui.LinearGradient{From: t.Accent, To: t.Success, Angle: 135, Oklab: true})
	}
	ui.Row(c).Gap(t.Space(4)).Wrap().AlignItems(ui.Center).Children(func() {
		switch s.variant {
		case "Masks":
			fade := []ui.GradientStop{{Color: ui.RGB(0, 0, 0), At: 0}, {Color: ui.RGBA(0, 0, 0, 0), At: 1}}
			tile("MaskLinearGradient").MaskLinearGradient(ui.LinearGradient{From: ui.RGB(0, 0, 0), To: ui.RGBA(0, 0, 0, 0), Angle: 180})
			tile("MaskRadialGradient").MaskRadialGradient(ui.RadialGradient{Stops: fade})
			tile("MaskConicGradient").MaskConicGradient(ui.ConicGradient{Stops: fade})
			tile("Masks").Masks(
				ui.Mask{Image: ui.LinearGradient{From: ui.RGB(0, 0, 0), To: ui.RGBA(0, 0, 0, 0), Angle: 90}},
				ui.Mask{Image: ui.RadialGradient{Stops: fade}, Composite: ui.MaskIntersect})
		case "Hit testing":
			hitTesting(c, s, size)
		case "Depth":
			depth(c, s, size)
		default:
			tile("ClipInset").ClipInset(t.Space(2), t.Space(2), t.Space(2), t.Space(2), t.Radius*2)
			tile("ClipInsetPercent").ClipInsetPercent(10, 25, 10, 25, t.Radius)
			star, err := ui.ParsePath(starPath)
			if err == nil {
				ui.Box(c).Size(size, size).Label("ClipPath").Draw(func(p *ui.Painter, r ui.Rect) {
					p.Transform(ui.Translation(r.X, r.Y).Then(ui.Scaling(r.W/100, r.H/100)), 0, 0, func() {
						p.ClipPath(star, func() { p.Fill(ui.Rect{W: 100, H: 100}, t.Accent, 0) })
					})
				})
			}
			if bitmap, err := galleryImage(); err == nil {
				for _, x := range []float32{0, 1} {
					components.Image(c, bitmap).Key(x).Size(size, size/2).Fit(ui.Cover).ObjectPosition(x, .5).Label("ObjectPosition").Radius(t.Radius)
				}
			}
		}
	})
}

// hitTesting makes a ring take the pointer only on its band, and counts the
// presses that reach it.
func hitTesting(c *ui.Context, s *demoState, size float32) {
	t := c.Theme()
	ring := func(r ui.Rect) *ui.Path {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		return new(ui.Path).Circle(cx, cy, r.W/2).Circle(cx, cy, r.W/4).FillRule(ui.EvenOdd)
	}
	target := ui.Box(c).Size(size, size).Label("HitTest").Cursor(ui.CursorPointer).
		HitTest(func(x, y float32, r ui.Rect) bool { return ring(r).Contains(x, y) })
	color := t.Accent
	if target.Hovered() {
		color = t.AccentHover
	}
	target.Draw(func(p *ui.Painter, r ui.Rect) { p.FillPath(ring(r), color) })
	if target.Clicked() {
		s.count++
	}
	components.Text(c, msg(c).CountD(s.count))
}

// depth stacks cards by ZIndex, a click raising one, and moves one toward the
// viewer with TranslateZ.
func depth(c *ui.Context, s *demoState, size float32) {
	t := c.Theme()
	ui.Box(c).Size(size*2, size).Children(func() {
		for i, color := range []ui.Color{t.Accent, t.Success, t.Warning} {
			z := 0
			if s.index == i {
				z = 1
			}
			card := ui.Box(c).Key(i).Absolute().Left(float32(i)*size/2).Top(0).Size(size, size).Radius(t.Radius * 2).
				Background(color).ZIndex(z).Label("ZIndex").Focusable()
			if card.Clicked() {
				s.index = i
			}
		}
	})
	card := ui.Box(c).Size(size, size).Radius(t.Radius*2).Background(t.Surface).Border(1, t.Border).Center().Label("TranslateZ")
	lift := float32(0)
	if card.Hovered() {
		lift = 80
	}
	z := card.Spring("lift", lift, ui.SpringOptions{Duration: 300 * time.Millisecond})
	card.RotateY(20).Perspective(600).TranslateZ(z).Children(func() { components.Text(c, "TranslateZ") })
	ui.Box(c).Size(size, size).Radius(t.Radius*2).Background(t.Surface).Border(1, t.Border).Center().Label("Translation3D").
		Perspective(600).Transform(ui.Translation3D(0, 0, -60)).Children(func() { components.Text(c, "Translation3D") })
	ui.Box(c).Size(size, size).Radius(t.Radius*2).Background(t.Surface).Border(1, t.Border).Center().Label("TranslationZ").
		Perspective(600).Transform(ui.TranslationZ(60)).Children(func() { components.Text(c, "TranslationZ") })
}
