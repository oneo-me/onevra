package gallery

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/icons/lucide"
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
	richMenu(c, s)
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

// richMenu places a drawn menu at its button's end, opens it from the app and
// shows items with icons, subtitles, trailing values and content of their own.
func richMenu(c *ui.Context, s *demoState) {
	t := c.Theme()
	ui.Row(c).Gap(t.Space(2)).Wrap().AlignItems(ui.Center).Children(func() {
		if components.Button(c, "Open", components.ButtonOutline).Clicked() {
			s.open = true
		}
		star, err := mygo.NewMenuIconSVG(lucide.Star(), true)
		ui.PopoverMenuButton(c, tr(c, "open"), func(m *ui.Menu) {
			m.MaxWidth(t.Space(70))
			item := m.Item(tr(c, "duplicate")).ID("duplicate").Subtitle(tr(c, "details")).Trailing("12 KB").Tooltip(tr(c, "general"))
			if err == nil {
				item.Icon(star).IconColor(t.Accent)
			}
			s.other = item.Highlighted()
			m.Item(tr(c, "general")).ID("progress").Content(func(content ui.MenuItemContent) {
				components.Progress(content.Context, .6).FillWidth()
			})
			m.Item(tr(c, "delete")).ID("delete").Danger()
		}, ui.PopoverMenuButtonOptions{At: ui.AnchorBottomEnd, Self: ui.AnchorTopEnd, Open: &s.open})
		presentation := c.MenuPresentation() == ui.PopoverMenus
		if components.Switch(c, &presentation, "SetMenuPresentation").Changed() {
			if presentation {
				c.SetMenuPresentation(ui.PopoverMenus)
			} else {
				c.SetMenuPresentation(ui.SystemMenus)
			}
		}
	})
}
