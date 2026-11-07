package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// exampleTextEffects casts and outlines text, wraps it by shape and pads the
// spans of rich text.
func exampleTextEffects(c *ui.Context, s *demoState) {
	t := c.Theme()
	sample := tr(c, "textEffectsSample")
	column := ui.Column(c).Gap(t.Space(3)).FillWidth().MaxWidth(t.Space(90))
	column.Children(func() {
		switch s.variant {
		case "Wrapping":
			components.Text(c, sample).Balance().Label("Balance").FontSize(t.FontSize * 1.5).Bold()
			components.Text(c, sample).Pretty().Label("Pretty")
			components.Text(c, tr(c, "textEffectsSpaced")).PreserveSpaces().Label("PreserveSpaces")
			components.Text(c, "a\tb\tc").TabSize(8).Font("monospace").Label("TabSize")
		case "Spans":
			components.RichText(c, ui.Span{Text: sample, Shadow: ui.NewTextShadow(0, 1, 3, t.Accent)},
				ui.Span{Text: " "},
				ui.Span{Text: "mygo", Font: "monospace", Background: t.Surface, Radius: t.Radius / 2, PaddingX: t.Space(1), PaddingY: t.Space(.5), NoLetterSpacing: true}).
				LetterSpacing(.5).Label("Span")
		default:
			components.Text(c, sample).FontSize(t.FontSize*2).Bold().TextShadow(0, 2, 6, t.Accent).Label("TextShadow")
			components.Text(c, sample).FontSize(t.FontSize*2).Bold().TextStroke(1, t.Accent).TextColor(t.Background).Label("TextStroke")
			card := ui.Box(c).Padding(t.Space(3)).Radius(t.Radius).Background(t.Surface).Shadow(0, 2, 8, 0, ui.RGBA(0, 0, 0, .3)).Label("NoShadow")
			components.Switch(c, &s.on, "NoShadow")
			if s.on {
				card.NoShadow()
			}
			card.Children(func() { components.Text(c, sample) })
			components.Switch(c, &s.other, "DimDisabled").FillWidth()
			components.Button(c, sample, components.ButtonOutline).Disabled(true).DimDisabled(s.other)
			ui.Box(c).MaxWidth(t.Space(60)).MaxWidthPercent(50).Padding(t.Space(2)).Radius(t.Radius).Background(t.Surface).Label("MaxWidthPercent").
				Children(func() { components.Text(c, sample) })
			ui.Grid(c).ColumnTracks(ui.Fr(1).Min(t.Space(30)), ui.Fr(1).MinContent()).FillWidth().Gap(t.Space(2)).Label("Track.Min").Children(func() {
				components.Text(c, "Track.Min")
				components.Text(c, "Track.MinContent")
			})
		}
	})
}
