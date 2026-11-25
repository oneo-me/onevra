package sidebar

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestShellContentTitle(t *testing.T) {
	s, _ := fixture()
	var title ui.Rect
	s.options.ContentTitle = func(c *ui.Context) {
		ui.Text(c, "Current document").SingleLine().Grow(1).MinWidth(0).DrawOver(func(_ *ui.Painter, r ui.Rect) { title = r })
	}
	tt := ui.NewTester(s.view, 820, 600)
	tt.SetTitleBar(ui.TitleBar{Height: 48, Left: 80})
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	if r := title; r.X != 236 || r.Y < 0 || r.Y+r.H > 48 {
		t.Fatal("title is not on the leading main-content titlebar", r)
	}
	for _, width := range []int{360, 820} {
		for _, scale := range []float32{1, 1.5, 2} {
			for _, leftOpen := range []bool{true, false} {
				s.SetLeftExpanded(leftOpen)
				tt.SetSize(width, 600)
				tt.SetPreferences(ui.Preferences{TextScale: scale, ReduceMotion: true})
				left, _ := tt.Find("Toggle sidebar")
				r := title
				if r.X < left.X+left.W || r.X+r.W > float32(width)-8 || r.Y < 0 || r.Y+r.H > max(48, 48*scale) {
					t.Fatalf("width=%d scale=%v left=%v: title overlaps controls: %+v, left %+v", width, scale, leftOpen, r, left)
				}
			}
		}
	}
	s.SetPage(func(*ui.Context) {})
	tt.Frame()
	if tt.HasText("Current document") {
		t.Fatal("replacement page retained content title")
	}
	s.SetPage(nil)
	tt.Frame()
	if !tt.HasText("Current document") {
		t.Fatal("returning did not restore content title")
	}
}

func TestShellTitleInsetOptions(t *testing.T) {
	// Relative assertions on one geometry: base and shifted differ only in
	// the insets, so the numbers hold on every platform the suite runs on.
	rect := func(insetLeft, insetRight float32) ui.Rect {
		s, _ := fixture()
		s.options.TitleInsetLeft, s.options.TitleInsetRight = insetLeft, insetRight
		var title ui.Rect
		s.options.ContentTitle = func(c *ui.Context) {
			ui.Text(c, "Current document").SingleLine().Grow(1).MinWidth(0).DrawOver(func(_ *ui.Painter, r ui.Rect) { title = r })
		}
		tt := ui.NewTester(s.view, 820, 600)
		tt.SetTitleBar(ui.TitleBar{Height: 48, Left: 80})
		tt.SetPreferences(ui.Preferences{ReduceMotion: true})
		return title
	}
	base := rect(0, 0) // non-positive values use the defaults 24/8
	shifted := rect(12, 20)
	if shifted.X != base.X-12 {
		t.Fatal("TitleInsetLeft did not move the title by 24-12", base, shifted)
	}
	if shifted.X+shifted.W != base.X+base.W-12 {
		t.Fatal("TitleInsetRight did not narrow the trailing room", base, shifted)
	}
	collapsed := func(insetLeft float32) ui.Rect {
		s, _ := fixture()
		s.options.TitleInsetLeft = insetLeft
		var title ui.Rect
		s.options.ContentTitle = func(c *ui.Context) {
			ui.Text(c, "Current document").SingleLine().Grow(1).MinWidth(0).DrawOver(func(_ *ui.Painter, r ui.Rect) { title = r })
		}
		tt := ui.NewTester(s.view, 820, 600)
		tt.SetTitleBar(ui.TitleBar{Height: 48, Left: 80})
		tt.SetPreferences(ui.Preferences{ReduceMotion: true})
		s.SetLeftExpanded(false)
		tt.Frame()
		return title
	}
	baseC, shiftedC := collapsed(0), collapsed(12)
	if shiftedC.X != baseC.X-12 {
		t.Fatal("collapsed title did not shift by 24-12", baseC, shiftedC)
	}
}
