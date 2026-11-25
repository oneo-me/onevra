package sidebar

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/oneo-me/onevra/source/internal/l10n"
)

type shellCaptionWindow interface {
	Minimize()
	ToggleMaximize()
	IsMaximized() bool
	Close()
}

func (s *Shell) captionButtons(c *ui.Context) {
	w, _ := s.host.(shellCaptionWindow)
	m := l10n.In(c.Locale())
	// Use a consistent compact glyph size without shrinking the hit targets.
	iconSize := c.Theme().Rem(0.875)
	ui.Row(c).Key("window-controls").Shrink(0).Gap(4).AlignItems(ui.Center).Children(func() {
		// Asymmetric padding lowers the centered glyph by 2 DIPs at normal scale.
		minimize := iconButton(c, lucide.Minus(), m.Text("minimizeWindow"), iconSize).
			Padding(c.Theme().Rem(0.25), 0, 0, 0)
		if minimize.Clicked() && w != nil {
			w.Minimize()
		}
		icon, label := lucide.Square(), m.Text("maximizeWindow")
		if w != nil && w.IsMaximized() {
			icon, label = lucide.Copy(), m.Text("restoreWindow")
		}
		if iconButton(c, icon, label, iconSize).MaximizeButton().Clicked() && w != nil {
			w.ToggleMaximize()
		}
		closeButton := iconButton(c, lucide.X(), m.Text("closeWindow"), iconSize)
		if closeButton.Hovered() || closeButton.Pressed() {
			closeButton.TextColor(ui.RGB(255, 255, 255))
		}
		closeButton.Draw(func(p *ui.Painter, r ui.Rect) {
			if closeButton.IsDisabled() {
				return
			}
			if closeButton.Pressed() {
				p.Fill(r, ui.RGB(196, 43, 28), c.Theme().Radius)
			} else if closeButton.Hovered() {
				p.Fill(r, ui.RGB(232, 17, 35), c.Theme().Radius)
			}
		})
		if closeButton.Clicked() && w != nil {
			// Close (not Destroy) preserves application close cancellation.
			w.Close()
		}
	})
}

// captionWidth is the width of the caption buttons' row.
func captionWidth(c *ui.Context) float32 {
	return 3*max(32, c.Theme().Rem(2)) + 2*4
}
