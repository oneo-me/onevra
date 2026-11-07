package gallery

import (
	"context"
	"fmt"

	"github.com/egoist/mygo/plugins/permissionflow"
	"github.com/egoist/mygo/plugins/sfsymbols"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// exampleSFSymbols draws system symbols, which fall back to a Lucide icon on
// systems that have none.
func exampleSFSymbols(c *ui.Context, _ *demoState) {
	t := c.Theme()
	ui.Row(c).Gap(t.Space(4)).Wrap().AlignItems(ui.Center).Children(func() {
		for i, name := range []string{"gearshape", "sparkles", "terminal", "doc.text"} {
			sfsymbols.Symbol(c, name).Size(t.Space(8)).Weight(sfsymbols.Weight(i)).Color(t.Accent).Fallback(lucide.Sparkles())
		}
		sfsymbols.Symbol(c, "heart").Size(t.Space(8)).Fill().Color(t.Danger).Fallback(lucide.Heart())
	})
	components.Text(c, fmt.Sprint("Available: ", sfsymbols.Available())).TextColor(t.TextMuted)
}

// permissionDemo keeps the status each check or guide last reported.
type permissionDemo struct {
	status string
}

// examplePermissionFlow checks a permission and opens the guide to it. Headless
// views never ask the system.
func examplePermissionFlow(c *ui.Context, s *demoState) {
	t := c.Theme()
	d := &s.permission
	ui.Row(c).Gap(t.Space(2)).Wrap().Children(func() {
		for _, name := range []struct {
			label string
			run   func() (permissionflow.Status, error)
		}{
			{"Check", func() (permissionflow.Status, error) { return permissionflow.Check(permissionflow.Accessibility) }},
			{"Guide", func() (permissionflow.Status, error) {
				return permissionflow.Guide(context.Background(), permissionflow.Accessibility, permissionflow.GuideOptions{})
			}},
		} {
			if components.Button(c, name.label, components.ButtonOutline).Clicked() && s.system.native {
				window := s.system.window()
				go func() {
					status, err := name.run()
					window.Update(func() {
						d.status = status.String()
						if err != nil {
							d.status = err.Error()
						}
					})
				}()
			}
		}
	})
	components.Text(c, d.status).TextColor(t.TextMuted)
}
