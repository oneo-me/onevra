package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// sortKeys are the sort orders the popover menus choose among, as radio
// items.
var sortKeys = []string{"name", "date", "size"}

// examplePopoverMenuButton draws a menu button's menu and a context menu in
// the window with the theme's look, rather than the system's menus.
func examplePopoverMenuButton(c *ui.Context, s *demoState) {
	t := c.Theme()
	sortBy := func(m *ui.Menu) {
		for i, key := range sortKeys {
			if m.Item(tr(c, key)).Radio(s.index == i).Chosen() {
				s.index = i
			}
		}
	}
	ui.PopoverMenuButton(c, tr(c, "popoverMenuSortBy"), func(m *ui.Menu) {
		sortBy(m)
		m.Separator()
		if m.Item(tr(c, "popoverMenuDescending")).Checked(s.on).Chosen() {
			s.on = !s.on
		}
	})
	ui.Box(c).Size(t.Space(70), t.Space(24)).Radius(t.Radius).Border(1, t.Border).Center().Focusable().
		Label(tr(c, "rightClickHere")).PopoverMenu(func(m *ui.Menu) {
		for _, action := range []string{"open", "duplicate"} {
			if m.Item(tr(c, action)).Chosen() {
				s.choice, s.count = action, s.count+1
			}
		}
		m.Submenu(tr(c, "popoverMenuSortBy"), sortBy)
		m.Separator()
		if m.Item(tr(c, "delete")).Chosen() {
			s.choice, s.count = "delete", s.count+1
		}
	}).Children(func() { components.Text(c, tr(c, "rightClickHere")).TextColor(t.TextMuted) })
	order := tr(c, sortKeys[s.index])
	if s.on {
		order = msg(c).PopoverMenuDescendingOrder(order)
	}
	components.Text(c, order)
	if s.count > 0 {
		ui.Text(c, msg(c).LabelCount(tr(c, s.choice), s.count))
	} else {
		components.Text(c, tr(c, "noActionYet"))
	}
}
