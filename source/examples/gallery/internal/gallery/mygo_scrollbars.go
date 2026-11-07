package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// scrollRows is how many rows the self-scrolling view holds.
const scrollRows = 60

// exampleScrollbars hides a scroll container's bars, which still scrolls by
// wheel, keyboard and assistive technology, and gives a view that scrolls
// itself the same bars as a scroll container.
func exampleScrollbars(c *ui.Context, s *demoState) {
	t := c.Theme()
	switch s.variant {
	case "Kept":
		components.ScrollView(c, components.ScrollViewOptions{}).KeepScrollbars(true).
			FillWidth().MaxWidth(t.Space(80)).Height(t.Space(36)).Gap(t.Space(1)).Label("KeepScrollbars").Focusable().Children(func() {
			for i := range scrollRows {
				components.Text(c, msg(c).ScrollbarsCard(i+1)).Key(i)
			}
		})
	case "Custom view":
		row, height := t.Space(6), t.Space(48)
		view := ui.Box(c).Size(t.Space(80), height).Radius(t.Radius).Border(1, t.Border).Clip().
			Label(tr(c, "scrollbarsCustomView"))
		hovered := view.Hovered()
		view.HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind != ui.InputScroll {
				return false
			}
			s.value = max(0, min(float64(scrollRows*row-height), s.value+float64(ev.DY)))
			return true
		}).Draw(func(p *ui.Painter, r ui.Rect) {
			// Rows drawn by the view itself, scrolled by its own offset.
			for i := range scrollRows {
				y := r.Y + float32(i)*row - float32(s.value)
				if y+row < r.Y || y > r.Y+r.H {
					continue
				}
				w := (r.W - t.Space(10)) * float32(3+(i*7)%5) / 7
				p.Fill(ui.Rect{X: r.X + t.Space(3), Y: y + row/4, W: w, H: row / 2}, t.Surface, row/4)
			}
			if hovered {
				p.Scrollbar(ui.Scrollbar{Box: r, View: r.H, Content: scrollRows * row, Offset: float32(s.value)}, ui.ScrollbarState{})
			}
		})
	default:
		components.ScrollView(c, components.ScrollViewOptions{Axes: components.ScrollHorizontal}).Scrollbars(false).
			FillWidth().MaxWidth(t.Space(120)).Gap(t.Space(3)).Label(tr(c, "scrollbarsHidden")).Focusable().Children(func() {
			for i := range 12 {
				ui.Box(c).Key(i).Size(t.Space(36), t.Space(24)).Shrink(0).Radius(t.Radius * 2).Background(t.Surface).Center().
					Children(func() { components.Text(c, msg(c).ScrollbarsCard(i+1)) })
			}
		})
	}
}
