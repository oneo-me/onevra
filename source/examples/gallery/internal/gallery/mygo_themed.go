package gallery

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/theme"
)

// exampleThemed previews Onevra controls in the light and the dark palette
// side by side, whatever the window's appearance.
func exampleThemed(c *ui.Context, s *demoState) {
	t := c.Theme()
	components.Switch(c, &s.other, "InactiveAccent")
	ui.Row(c).FillWidth().Gap(t.Space(4)).Wrap().Children(func() {
		for i, dark := range []bool{false, true} {
			preview := theme.Palette(dark)
			// Keep the desktop's text scale, which Palette ignores.
			preview.FontSize = t.FontSize
			preview.TooltipDelay = 100 * time.Millisecond
			if s.other {
				preview.InactiveAccent = preview.Accent
			}
			name := tr(c, "light")
			if dark {
				name = tr(c, "dark")
			}
			ui.Themed(c, preview).Key(i).Width(t.Space(60)).Padding(t.Space(4)).Gap(t.Space(3)).
				Radius(t.Radius*2).Border(1, preview.Border).Label(name).Children(func() {
				components.Text(c, name).Bold()
				components.Switch(c, &s.on, tr(c, "notifications"))
				ui.Row(c).Gap(t.Space(2)).Children(func() {
					components.Tooltip(c, components.Button(c, tr(c, "ok"), components.ButtonPrimary), "TooltipDelay")
					components.Button(c, tr(c, "cancel"), components.ButtonOutline)
				})
			})
		}
	})
}
