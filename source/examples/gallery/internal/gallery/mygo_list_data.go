package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// exampleListData marks a table's header, paints its chosen cells, drags rows
// out to a target and keeps a chat-like list in place as a fold opens.
func exampleListData(c *ui.Context, s *demoState) {
	t := c.Theme()
	rows := []string{"open", "duplicate", "delete", "name", "date", "size"}
	switch s.variant {
	case "DragRows":
		s.list.DragRows = func(chosen []int) any { return chosen }
		ui.Row(c).Gap(t.Space(4)).FillWidth().Children(func() {
			ui.List(c, &s.list, len(rows), func(i int) {
				components.Text(c, tr(c, rows[i])).Padding(t.Space(1), t.Space(2))
			}).Width(t.Space(50)).Height(t.Space(48)).Border(1, t.Border).Radius(t.Radius)
			folder := ui.Box(c).Size(t.Space(40), t.Space(48)).Radius(t.Radius).Border(1, t.Border).Center().Label("DragRows")
			if dropped, ok := ui.Drop[[]int](folder); ok {
				s.count += len(dropped)
			}
			if _, over := ui.DragOver[[]int](folder); over {
				folder.Background(t.SurfaceHover)
			}
			folder.Children(func() { components.Text(c, msg(c).CountD(s.count)) })
		})
	case "HoldPlace":
		s.list.FollowEnd = true
		ui.List(c, &s.list, len(rows), func(i int) {
			ui.Column(c).Padding(t.Space(1), t.Space(2)).FillWidth().Gap(t.Space(1)).Children(func() {
				if components.Text(c, tr(c, rows[i])).Clicked() {
					s.open = !s.open
					s.list.HoldPlace()
				}
				if s.open && i == len(rows)-1 {
					ui.Box(c).FillWidth().Height(t.Space(32)).Radius(t.Radius).Background(t.Surface)
				}
			})
		}).FillWidth().MaxWidth(t.Space(70)).Height(t.Space(40)).Border(1, t.Border).Radius(t.Radius).KeepScrollbars(true)
	default:
		s.list.CellSelect = true
		s.list.CellsStyle = ui.CellsStyle{Cell: t.Accent.Alpha(.3), CellText: t.Text, Row: t.Accent.Alpha(.08), Stripe: t.Surface}
		s.list.HeaderCell = func(cell ui.Element, col int) {
			if col == 1 {
				components.Text(c, "•").TextColor(t.Danger)
			}
		}
		ui.Table(c, &s.list, []ui.TableColumn{{Title: tr(c, "name"), Width: t.Space(40)}, {Title: tr(c, "size"), Width: t.Space(30)}}, len(rows),
			func(row, col int) {
				if col == 0 {
					components.Text(c, tr(c, rows[row]))
				} else {
					components.Text(c, msg(c).SizeKb(row*128))
				}
			}).FillWidth().MaxWidth(t.Space(90)).Height(t.Space(48))
	}
}
