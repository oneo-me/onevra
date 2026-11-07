package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// scrollAligns are where the Into view variant lines its target up.
var scrollAligns = []struct {
	name  string
	align ui.ScrollAlign
}{{"ScrollStart", ui.ScrollStart}, {"ScrollCenter", ui.ScrollCenter}, {"ScrollEnd", ui.ScrollEnd}}

// exampleScrollPositions keeps headings in view with Sticky, rests a carousel
// on the places SnapAlign marks and lines a row up with ScrollIntoView.
func exampleScrollPositions(c *ui.Context, s *demoState) {
	t := c.Theme()
	card := func(i int) ui.Element {
		return ui.Box(c).Key(i).Padding(t.Space(2), t.Space(3)).Radius(t.Radius).Background(t.Surface)
	}
	switch s.variant {
	case "Snap":
		components.ScrollView(c, components.ScrollViewOptions{Axes: components.ScrollHorizontal}).ScrollSnap(ui.SnapMandatory).
			FillWidth().MaxWidth(t.Space(100)).Gap(t.Space(3)).Label("ScrollSnap").Focusable().Children(func() {
			for i := range 8 {
				card(i).Size(t.Space(70), t.Space(36)).Shrink(0).SnapAlign(ui.Center).Center().
					Children(func() { components.Text(c, msg(c).ScrollbarsCard(i+1)) })
			}
		})
	case "Into view":
		ui.Row(c).Gap(t.Space(2)).Wrap().Children(func() {
			for i, align := range scrollAligns {
				if components.Button(c, align.name, components.ButtonOutline).Clicked() {
					s.index, s.other, s.count = i, true, (s.count+7)%24
				}
			}
		})
		components.ScrollView(c, components.ScrollViewOptions{}).FillWidth().MaxWidth(t.Space(80)).Height(t.Space(56)).Gap(t.Space(1)).Children(func() {
			for i := range 24 {
				row := card(i).FillWidth().Children(func() { components.Text(c, msg(c).ScrollbarsCard(i+1)) })
				if i == s.count {
					row.Background(t.SurfacePressed)
					if s.other {
						s.other = false
						row.ScrollIntoView(ui.ScrollIntoViewOptions{Block: scrollAligns[s.index].align, Animation: &ui.Smooth})
					}
				}
			}
		})
	default:
		components.ScrollView(c, components.ScrollViewOptions{}).FillWidth().MaxWidth(t.Space(80)).Height(t.Space(64)).Children(func() {
			for day := range 4 {
				ui.Column(c).Key(day).Gap(t.Space(1)).Children(func() {
					components.Text(c, msg(c).ScrollbarsCard(day+1)).Bold().FillWidth().Padding(t.Space(2), t.Space(3)).
						Background(t.Background).Sticky().Top(0)
					for i := range 6 {
						card(i).FillWidth().Children(func() { components.Text(c, tr(c, "stickyEvent")) })
					}
				})
			}
		})
	}
}
