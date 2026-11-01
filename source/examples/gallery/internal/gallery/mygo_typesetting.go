package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// typesettingProse is fixture text in English, as other sample documents
// are. Soft hyphens (U+00AD) let long words break with a hyphen.
const typesettingProse = "Typesetting turns a manu­script into pages: it justi­fies every line but the last of a paragraph, indents the first, and hyphen­ates long words where they break. A footnote mark rides above the baseline"

// exampleTypesetting lays out a paragraph as a book sets it, and marks where
// each of its lines sits.
func exampleTypesetting(c *ui.Context, s *demoState) {
	t := c.Theme()
	width := float32(240 + s.value*2.4)
	font := ui.Font{Size: t.FontSize}
	layout := ui.ShapeRichTextWith([]ui.Span{
		{Text: typesettingProse},
		{Text: "1", Size: t.FontSize * .7, BaselineShift: t.FontSize * .4, Color: t.Accent},
		{Text: "."},
	}, font, ui.TextShape{Width: width, Justify: s.on, Indent: t.FontSize * 2})
	lines := layout.Lines()
	_, h := layout.Size()
	ui.Box(c).Size(width, h).Label(typesettingProse).Draw(func(p *ui.Painter, r ui.Rect) {
		for _, line := range lines {
			y := r.Y + line.Baseline
			p.Line(r.X+line.Rect.X, y, r.X+line.Rect.X+line.Rect.W, y, 1, t.Accent.Alpha(.25))
		}
		p.TextLayout(layout, r.X, r.Y, t.Text)
	})
	components.Text(c, msg(c).TypesettingLines(len(lines))).TextColor(t.TextMuted)
	components.Checkbox(c, &s.on, tr(c, "typesettingJustify"))
	ui.Row(c).FillWidth().MaxWidth(t.Space(100)).Gap(t.Space(3)).AlignItems(ui.Center).Children(func() {
		name := tr(c, "typesettingWidth")
		components.Text(c, name).Shrink(0).SingleLine()
		components.Slider(c, &s.value, 0, 100).Label(name).Grow(1).MinWidth(0)
	})
}
