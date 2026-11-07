package gallery

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// baseItem paints an unstyled item the way every variant of this page does.
func baseItem(c *ui.Context, item ui.Element, label string) ui.Element {
	t := c.Theme()
	item.Padding(t.Space(1.5), t.Space(2.5)).Radius(t.Radius).Gap(t.Space(2))
	if item.Highlighted() {
		item.Background(t.Text.Alpha(.1))
	}
	return item.Children(func() { components.Text(c, label) })
}

// basePanel paints an unstyled popup.
func basePanel(c *ui.Context, panel ui.Element) {
	t := c.Theme()
	panel.Padding(t.Space(1)).Radius(t.Radius*1.5).Background(t.Background).Border(1, t.Border).MinWidth(t.Space(40))
}

// exampleWidgetBases builds MyGo's look-free menus, accordions, trees, mixed
// check boxes and caller-opened selects in the page's own layout.
func exampleWidgetBases(c *ui.Context, s *demoState) {
	t := c.Theme()
	switch s.variant {
	case "ContextMenuBase":
		target := ui.Box(c).Size(t.Space(70), t.Space(24)).Radius(t.Radius).Border(1, t.Border).Center().Focusable().
			Label(tr(c, "rightClickHere")).Children(func() { components.Text(c, tr(c, "rightClickHere")).TextColor(t.TextMuted) })
		m := ui.ContextMenuBase(c, target, &s.open)
		m.Popup(func(panel ui.Element) {
			basePanel(c, panel)
			for _, action := range []string{"open", "duplicate"} {
				if baseItem(c, m.Item(), tr(c, action)).Clicked() {
					s.choice = action
				}
			}
		})
	case "AccordionBase":
		a := ui.AccordionBase(c)
		a.Root.Gap(t.Space(1)).FillWidth().MaxWidth(t.Space(80)).Children(func() {
			item := a.Item(&s.open)
			item.Trigger.Padding(t.Space(2)).Radius(t.Radius).Background(t.Surface).Children(func() { components.Text(c, tr(c, "general")) })
			item.Panel(func() { components.Checkbox(c, &s.on, tr(c, "notifications")) })
			other := a.Item(&s.other)
			other.Trigger.Padding(t.Space(2)).Radius(t.Radius).Background(t.Surface).Children(func() { components.Text(c, tr(c, "details")) })
			other.Panel(func() { components.Text(c, tr(c, "stickyEvent")) })
		})
	case "TreeBase":
		tree := ui.TreeBase(c)
		tree.Root.Gap(t.Space(.5)).Children(func() {
			folder := tree.Item(&s.open)
			folder.Row.Padding(t.Space(1), t.Space(2)).Radius(t.Radius).Children(func() { components.Text(c, tr(c, "general")) })
			folder.Group(func() {
				for _, name := range []string{"open", "duplicate"} {
					c.Key(name)
					leaf := tree.Item(nil)
					leaf.Row.Padding(t.Space(1), t.Space(2)).PaddingStart(t.Space(2) + t.Space(4)*float32(leaf.Depth())).Radius(t.Radius).
						Children(func() { components.Text(c, tr(c, name)) })
					if leaf.Row.FocusVisible() {
						leaf.Row.Background(t.Text.Alpha(.1))
					}
				}
			})
			if folder.Row.FocusVisible() {
				folder.Row.Background(t.Text.Alpha(.1))
			}
		})
	case "CheckboxState":
		ui.CheckboxState(c, &s.check, tr(c, "notifications"))
		base := ui.CheckboxStateBase(c, &s.check).Label("CheckboxStateBase").Padding(t.Space(1), t.Space(2)).Radius(t.Radius).Background(t.Surface)
		base.Children(func() { components.Text(c, []string{"–", "✓", "•"}[s.check]) })
	case "SelectBaseOpen":
		if components.Button(c, "SetOpen", components.ButtonOutline).Clicked() {
			s.open = true
		}
		sel := ui.SelectBaseOpen(c, &s.choice, &s.open)
		sel.Trigger.Padding(t.Space(1.5), t.Space(2.5)).Radius(t.Radius).Border(1, t.Border).Children(func() { components.Text(c, tr(c, strings.ToLower(s.choice))) })
		sel.Popup(func(panel ui.Element) {
			basePanel(c, panel)
			for _, option := range []string{"Small", "Large"} {
				baseItem(c, sel.Item(option), tr(c, strings.ToLower(option)))
			}
		})
	case "ComboboxBase":
		combo := ui.ComboboxBase(c, &s.query)
		combo.Input.Label(tr(c, "searchSizes")).Padding(t.Space(1.5), t.Space(2.5)).Radius(t.Radius).Border(1, t.Border).Width(t.Space(60))
		combo.Popup(func(panel ui.Element) {
			basePanel(c, panel)
			shown := 0
			for _, option := range []string{"Small", "Large"} {
				label := tr(c, strings.ToLower(option))
				if strings.Contains(strings.ToLower(label), strings.ToLower(s.query)) {
					baseItem(c, combo.Item(label), label)
					shown++
				}
			}
			if shown == 0 {
				components.Text(c, tr(c, "noMatchingSizes")).TextColor(t.TextMuted).Padding(t.Space(1.5), t.Space(2.5))
			}
		})
	case "SplitReset":
		ui.Split(c, &s.size, func() { components.Text(c, tr(c, "details")) }, func() { components.Text(c, tr(c, "general")) }).
			SplitReset(140).FillWidth().Height(t.Space(24)).Border(1, t.Border)
	case "AttachEdgeMargin":
		anchor := components.Button(c, tr(c, "open"), components.ButtonOutline)
		if anchor.Clicked() {
			s.open = !s.open
		}
		ui.PopoverBase(c, anchor, &s.open, func(panel ui.Element) {
			basePanel(c, panel)
			panel.AttachTo(anchor, ui.AnchorEnd, ui.AnchorStart).AttachEdgeMargin(t.Space(6)).AttachedOrigin().MarginStart(t.Space(2)).
				Transition(ui.ElementTransition{Enter: &ui.Motion{Scale: .9, Flip: true, Opacity: 0}})
			components.Text(c, "AttachEdgeMargin")
		})
	default:
		m := ui.MenuBase(c, &s.open)
		m.Trigger.Padding(t.Space(1.5), t.Space(2.5)).Radius(t.Radius).Border(1, t.Border).Children(func() { components.Text(c, tr(c, "popoverMenuSortBy")) })
		m.Popup(func(panel ui.Element) {
			basePanel(c, panel)
			m.Group(func(group ui.Element) {
				group.Row().Gap(t.Space(1))
				for i, key := range sortKeys {
					item := baseItem(c, m.RadioItem(s.index == i), tr(c, key)).KeepOpen()
					if item.Clicked() {
						s.index = i
					}
				}
			})
			if baseItem(c, m.CheckItem(s.on), tr(c, "popoverMenuDescending")).Clicked() {
				s.on = !s.on
			}
		})
	}
}
