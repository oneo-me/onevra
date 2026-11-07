package components

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/theme"
)

func TestScaledControlCoordinates(t *testing.T) {
	for _, scale := range []float32{.75, 1.5} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			low, high := 20.0, 80.0
			color := ui.RGB(255, 0, 0)
			size := float32(100)
			tt := ui.NewTester(func(c *ui.Context) {
				ui.Column(c).Width(300).Gap(12).Scale(scale, scale).TransformOrigin(0, 0).Children(func() {
					Slider(c, &low, 0, 100, SliderOptions{Range: &high, Step: 10}).Label("Price").Width(240)
					Split(c, &size, nil, nil).Size(300, 100)
					ColorView(c, &color)
				})
			}, 500, 950)
			r, _ := tt.Find("Price")
			tt.ClickAt(r.X+r.W*.35, r.Y+r.H/2)
			if low != 30 || high != 80 {
				t.Fatalf("scaled range = %v, %v", low, high)
			}
			// The handle over the line takes the press and focuses it.
			r, _ = tt.Find("Divider")
			tt.ClickAt(r.X+r.W/2, r.Y+r.H/2)
			tt.Key(0, ui.KeyEnd)
			if size != 259 {
				t.Fatalf("scaled split maximum = %v", size)
			}
			r, _ = tt.Find("Saturation and brightness")
			tt.ClickAt(r.X+r.W/2, r.Y+r.H/2)
			want := ui.RGB(128, 64, 64)
			near := func(a, b uint8) bool { return max(int(a)-int(b), int(b)-int(a)) <= 1 }
			if !near(color.R, want.R) || !near(color.G, want.G) || !near(color.B, want.B) {
				t.Fatalf("scaled color midpoint = %v, want %v", color, want)
			}
		})
	}
}

func TestScaledTabsUnderline(t *testing.T) {
	for _, dark := range []bool{false, true} {
		selected := 0
		var list ui.Rect
		tt := ui.NewTester(func(c *ui.Context) {
			c.SetTheme(theme.Palette(dark))
			ui.Column(c).Scale(1.5, 1.5).TransformOrigin(0, 0).Children(func() {
				parts := Tabs(c, &selected, TabsUnderline, "General", "Appearance")
				list = parts.List.Bounds()
			})
		}, 500, 130)
		tt.SetPreferences(ui.Preferences{ReduceMotion: true})
		if err := tt.Click("Appearance"); err != nil {
			t.Fatal(err)
		}
		tt.Move(490, 120)
		tt.Frame()
		r, _ := tt.Find("Appearance")
		p := tt.Image().RGBAAt(int(r.X+r.W/2), int(list.Y+list.H-2))
		accent := theme.LightPalette().Accent
		if dark {
			accent = theme.DarkPalette().Accent
		}
		if p.R != accent.R || p.G != accent.G || p.B != accent.B {
			t.Fatalf("underline does not follow transformed tab: %v, want %v", p, accent)
		}
	}
}
