package examples

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

var _ = register(&menuDemo{})

type menuDemo struct {
	last   string
	pinned bool
}

func (d *menuDemo) View(c *ui.Context) {
	actions := func(m *components.Menu) {
		if m.Item("Open").Shortcut("⌘O").Chosen() {
			d.last = "Open"
		}
		if m.Item("Pinned").Checked(d.pinned).Chosen() {
			d.pinned = !d.pinned
		}
		m.Group("Move to", func(m *components.Menu) {
			for _, folder := range []string{"Desktop", "Documents"} {
				if m.Item(folder).Chosen() {
					d.last = "Move to " + folder
				}
			}
		})
		if m.Item("Delete").Disabled(d.pinned).Chosen() {
			d.last = "Delete"
		}
	}
	ui.Column(c).Gap(16).AlignItems(ui.Center).Children(func() {
		ui.Row(c).Gap(8).Children(func() {
			components.MenuButton(c, components.Button(c, "Actions", components.ButtonSecondary), actions)
			components.ContextMenu(c, components.Button(c, "Right-click me", components.ButtonOutline), actions)
		})
		text := "Choose an action"
		if d.last != "" {
			text = "Chose " + d.last
		}
		components.Text(c, text).TextColor(c.Theme().TextMuted)
	})
}
