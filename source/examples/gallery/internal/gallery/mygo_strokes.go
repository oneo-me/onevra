package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// strokeWidth is wide enough to tell the caps and joins apart.
const strokeWidth = 12

// exampleStrokes strokes paths with each cap and join, as SVG strokes do, and
// dashes one whose offset the slider moves. The paths are rebuilt only when
// their box changes.
func exampleStrokes(c *ui.Context, s *demoState) {
	t := c.Theme()
	tile := func(key, label string, draw func(p *ui.Painter, r ui.Rect)) {
		ui.Column(c).Key(key).Gap(t.Space(2)).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(120, 90).Radius(8).Background(t.Surface).Label(label).Draw(draw)
			components.Text(c, label).TextColor(t.TextMuted)
		})
	}
	caps := []struct {
		key string
		cap ui.LineCap
	}{{"strokeButtCap", ui.ButtCap}, {"strokeRoundCap", ui.RoundCap}, {"strokeSquareCap", ui.SquareCap}}
	joins := []struct {
		key  string
		join ui.LineJoin
	}{{"strokeMiterJoin", ui.MiterJoin}, {"strokeRoundJoin", ui.RoundJoin}, {"strokeBevelJoin", ui.BevelJoin}}
	ui.Row(c).Gap(t.Space(4)).Wrap().Children(func() {
		for _, v := range caps {
			tile(v.key, tr(c, v.key), func(p *ui.Painter, r ui.Rect) {
				var line ui.Path
				line.MoveTo(r.X+30, r.Y+r.H/2).LineTo(r.X+r.W-30, r.Y+r.H/2).Cap(v.cap)
				p.StrokePath(&line, strokeWidth, t.Accent)
				// The line's own ends, which square caps reach past.
				p.Line(r.X+30, r.Y+20, r.X+30, r.Y+r.H-20, 1, t.TextMuted)
				p.Line(r.X+r.W-30, r.Y+20, r.X+r.W-30, r.Y+r.H-20, 1, t.TextMuted)
			})
		}
	})
	ui.Row(c).Gap(t.Space(4)).Wrap().Children(func() {
		for _, v := range joins {
			tile(v.key, tr(c, v.key), func(p *ui.Painter, r ui.Rect) {
				var zigzag ui.Path
				zigzag.MoveTo(r.X+24, r.Y+r.H-22).LineTo(r.X+r.W/2, r.Y+22).LineTo(r.X+r.W-24, r.Y+r.H-22).
					Join(v.join).Cap(ui.ButtCap)
				p.StrokePath(&zigzag, strokeWidth, t.Accent)
			})
		}
	})
	ui.Box(c).FillWidth().MaxWidth(t.Space(100)).Height(60).Radius(8).Background(t.Surface).
		Label(tr(c, "strokeDashes")).Draw(func(p *ui.Painter, r ui.Rect) {
		var dashed ui.Path
		dashed.MoveTo(r.X+20, r.Y+r.H/2).LineTo(r.X+r.W-20, r.Y+r.H/2).Dash(float32(s.value), 24, 12, 4, 12)
		p.StrokePath(&dashed, 6, t.Accent)
	})
	ui.Row(c).FillWidth().MaxWidth(t.Space(100)).Gap(t.Space(3)).AlignItems(ui.Center).Children(func() {
		name := tr(c, "strokeDashOffset")
		components.Text(c, name).Shrink(0).SingleLine()
		components.Slider(c, &s.value, 0, 52).Label(name).Grow(1).MinWidth(0)
	})
}
