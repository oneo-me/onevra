package sidebar

import (
	"runtime"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/theme"
)

type fakeWindow struct {
	bounds        mygo.Rectangle
	invalidations int
}

func (w *fakeWindow) Invalidate() { w.invalidations++ }

func fixture() (*Shell, *fakeWindow) {
	s := newShell(ShellOptions{
		Left:    func(c *ui.Context) { ui.Text(c, "Navigation") },
		Content: func(c *ui.Context) { ui.Text(c, "Document") },
		Theme:   theme.Onevra,
	})
	w := &fakeWindow{bounds: mygo.Rectangle{X: 50, Y: 60, Width: 820, Height: 600}}
	s.host = w
	return s, w
}

func TestShellNativeOptions(t *testing.T) {
	s := newShell(ShellOptions{Title: "Example", Width: 820, Height: 600, MinWidth: 360, MinHeight: 320})
	o := s.nativeOptions()
	if o.Title != "Example" || o.Width != 820 || o.Height != 600 || o.MinWidth != 360 || o.MinHeight != 320 {
		t.Fatal(o)
	}
	if o.Content == nil || o.TitleBarStyle != mygo.TitleBarHidden || o.TrafficLightPosition == nil || *o.TrafficLightPosition != (mygo.Point{X: 16, Y: 16}) {
		t.Fatal(o)
	}
	material := mygo.VibrancyNone
	if runtime.GOOS == "darwin" {
		material = mygo.VibrancySidebar
	} else if runtime.GOOS == "windows" {
		material = mygo.VibrancyMica
	}
	if o.Transparent != (runtime.GOOS == "darwin") || o.Vibrancy != material || o.Frameless || o.HideWindowButtons != (runtime.GOOS == "windows") || o.DisableResize || o.StateKey != "" {
		t.Fatal(o)
	}
	if o := newShell(ShellOptions{StateKey: "main"}).nativeOptions(); o.StateKey != "main" {
		t.Fatalf("state key: %q", o.StateKey)
	}
}

func TestShellWidthAndPages(t *testing.T) {
	s, w := fixture()
	original := w.bounds
	for _, open := range []bool{false, true} {
		s.SetLeftExpanded(open)
		s.SetPage(func(c *ui.Context) { ui.Text(c, "Settings") })
		ui.NewTester(s.view, 820, 600)
		s.SetPage(nil)
		ui.NewTester(s.view, 820, 600)
		if w.bounds != original || s.LeftExpanded() != open {
			t.Fatal("page replacement changed window bounds or sidebar state")
		}
	}
	if w.invalidations != 6 {
		t.Fatal("state changes did not invalidate the window", w.invalidations)
	}
}

func TestShellLayoutAndReplacement(t *testing.T) {
	s, _ := fixture()
	tt := ui.NewTester(s.view, 820, 600)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	if !tt.HasText("Navigation") || !tt.HasText("Document") || tt.HasText("Inspector") {
		t.Fatal(tt.Texts())
	}
	if err := tt.Click("Toggle sidebar"); err != nil {
		t.Fatal(err)
	}
	if tt.HasText("Navigation") {
		t.Fatal("left sidebar did not collapse internally")
	}
	s.SetPage(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(HeaderHeight(c), 0, 0, 0).Children(func() {
			ui.Text(c, "Settings")
			if ui.Button(c, "Back").Clicked() {
				s.SetPage(nil)
			}
		})
	})
	tt = ui.NewTester(s.view, 820, 600)
	_, hasLeftToggle := tt.Find("Toggle sidebar")
	_, hasRightToggle := tt.Find("Toggle inspector")
	if !tt.HasText("Settings") || tt.HasText("Document") || tt.HasText("Inspector") || hasLeftToggle || hasRightToggle {
		t.Fatal(tt.Texts())
	}
	if err := tt.Click("Back"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(sidebarDuration + 30*time.Millisecond)
	tt.Frame()
	if !tt.HasText("Document") || tt.HasText("Navigation") {
		t.Fatal(tt.Texts())
	}
}

func TestShellControlClearance(t *testing.T) {
	s, _ := fixture()
	for _, scale := range []float32{1, 1.5, 2} {
		tt := ui.NewTester(s.view, 360, 480)
		tt.SetTitleBar(ui.TitleBar{Height: 48, Left: 80, Right: 100})
		tt.SetPreferences(ui.Preferences{TextScale: scale})
		r, ok := tt.Find("Toggle sidebar")
		if !ok || r.X < 80 || r.X+r.W > 260 || r.Y < 0 || r.Y+r.H > max(48, 48*scale) {
			t.Fatalf("scale %v: sidebar control overlaps native chrome: %+v", scale, r)
		}
	}
}

func TestShellSidebarAnimations(t *testing.T) {
	s, w := fixture()
	tt := ui.NewTester(s.view, 820, 600)
	contentX := func() float32 {
		t.Helper()
		r, ok := tt.Find("Document")
		if !ok {
			t.Fatal("missing main content")
		}
		return r.X
	}
	step := func() {
		tt.Advance(55 * time.Millisecond)
		tt.Frame()
	}
	finish := func() {
		tt.Advance(sidebarDuration + 30*time.Millisecond)
		tt.Frame()
	}
	startX := contentX()
	s.SetLeftExpanded(false)
	tt.Frame()
	step()
	if x := contentX(); x <= 0 || x >= startX || w.bounds.Width != 820 {
		t.Fatalf("left collapse should be intermediate without window resizing: x=%v, bounds=%+v", x, w.bounds)
	}
	// Reverse before completion, without losing the current layout.
	s.SetLeftExpanded(true)
	tt.Frame()
	finish()
	if x := contentX(); x != startX {
		t.Fatalf("left reversal ended at %v", x)
	}
	s.SetLeftExpanded(false)
	tt.Frame()
	finish()
	if contentX() != 0 || tt.HasText("Navigation") {
		t.Fatal("left collapse did not finish")
	}

	// Enabling reduced motion midway must snap the layout.
	s.SetLeftExpanded(true)
	tt.Frame()
	step()
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	if contentX() != startX || w.bounds.Width != 820 || tt.HasText("Inspector") {
		t.Fatal("Reduce Motion did not finish active animations")
	}
	s.SetLeftExpanded(false)
	tt.Frame()
	if contentX() != 0 || w.bounds.Width != 820 {
		t.Fatal("Reduce Motion animated a new toggle")
	}
}

func TestShellSidebarRevealsWithoutReflow(t *testing.T) {
	var left, main ui.Rect
	s, _ := fixture()
	s.options.Left = func(c *ui.Context) {
		ui.Box(c).FillWidth().Height(32).DrawOver(func(_ *ui.Painter, r ui.Rect) { left = r })
	}
	s.options.Content = func(c *ui.Context) { ui.Box(c).Fill().DrawOver(func(_ *ui.Painter, r ui.Rect) { main = r }) }
	tt := ui.NewTester(s.view, 820, 600)
	leftBefore := left
	s.SetLeftExpanded(false)
	tt.Frame()
	tt.Advance(55 * time.Millisecond)
	tt.Frame()
	if got := left; got != leftBefore {
		t.Fatalf("left content reflowed instead of being covered: %+v -> %+v", leftBefore, got)
	}
	if x := main.X; x <= 0 || x >= leftBefore.W {
		t.Fatalf("main surface did not move over sidebar: %v", x)
	}
	tt.Advance(sidebarDuration + 30*time.Millisecond)
	tt.Frame()
	if r := main; r.X != 0 || r.W != 820 {
		t.Fatal("collapsed sidebar reserved space", r)
	}
}

func TestShellSidebarResize(t *testing.T) {
	s, w := fixture()
	tt := ui.NewTester(s.view, 820, 600)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	separator := func() ui.Rect {
		t.Helper()
		r, ok := tt.Find("Resize sidebar")
		if !ok {
			t.Fatal("missing resize handle")
		}
		return r
	}
	drag := func(dx float32) {
		r := separator()
		x, y := r.X+r.W/2, r.Y+30
		tt.Move(x, y)
		if tt.Cursor() != ui.CursorResizeEW {
			t.Fatal("missing resize cursor")
		}
		tt.Press(x, y)
		tt.Move(x+dx, y)
		tt.Release(x+dx, y)
	}
	drag(50)
	if s.leftWidth != 270 || w.bounds.Width != 820 {
		t.Fatal("drag did not resize interior sidebar", s.leftWidth, w.bounds)
	}
	tt.Key(0, ui.KeyRight)
	if s.leftWidth != 280 {
		t.Fatal("keyboard resize did not move separator", s.leftWidth)
	}
	drag(1000)
	if s.leftWidth != 817 {
		t.Fatal("drag exceeded available width", s.leftWidth)
	}
	drag(-1000)
	if s.leftWidth != 120 {
		t.Fatal("drag exceeded minimum width", s.leftWidth)
	}
	drag(80)
	s.SetLeftExpanded(false)
	tt.Frame()
	if _, ok := tt.Find("Resize sidebar"); ok {
		t.Fatal("collapsed sidebar kept resize hit area")
	}
	s.SetLeftExpanded(true)
	tt.Frame()
	if s.leftWidth != 200 {
		t.Fatal("toggle lost resized width")
	}
	if r := separator(); r.X+r.W/2 != 200 || r.Y != 0 || r.H != 600 {
		t.Fatal("misplaced resize hit area", r)
	}
	s.SetPage(func(*ui.Context) {})
	tt.Frame()
	if _, ok := tt.Find("Resize sidebar"); ok {
		t.Fatal("replacement page kept resize hit area")
	}
}

func TestShellSidebarResizeFullHeight(t *testing.T) {
	for _, y := range []float32{0, 2, 24, 47, 300, 599} {
		s, w := fixture()
		tt := ui.NewTester(s.view, 820, 600)
		tt.SetTitleBar(ui.TitleBar{Height: 48, Left: 80})
		tt.Move(220, y)
		if tt.Cursor() != ui.CursorResizeEW {
			t.Fatalf("y=%v: missing resize cursor", y)
		}
		tt.Press(220, y)
		tt.Move(250, y)
		tt.Release(250, y)
		if s.leftWidth != 250 || w.bounds.Width != 820 {
			t.Fatalf("y=%v: drag did not resize sidebar: width=%v bounds=%+v", y, s.leftWidth, w.bounds)
		}
	}
}

func TestShellSeparatorsDoNotReserveSpace(t *testing.T) {
	var main ui.Rect
	s, _ := fixture()
	// Shell installs the frame theme; distinguishable pixels come from the
	// palette it installs.
	s.options.Theme = func(*ui.Context) *ui.Theme { return theme.DarkPalette() }
	s.options.Content = func(c *ui.Context) { ui.Box(c).Fill().DrawOver(func(_ *ui.Painter, r ui.Rect) { main = r }) }
	tt := ui.NewTester(s.view, 820, 600)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	palette := theme.DarkPalette()
	assertPixel := func(x, y int, separator bool) {
		t.Helper()
		p := tt.Image().RGBAAt(x, y)
		want := palette.Background
		if separator {
			want = palette.Border
		}
		if p.R != want.R || p.G != want.G || p.B != want.B || p.A != want.A {
			t.Fatalf("pixel (%d,%d) = %+v, separator=%v", x, y, p, separator)
		}
	}
	if r := main; r.X != 220 || r.W != 600 {
		t.Fatal("separator reserved main layout space", r)
	}
	for _, y := range []int{2, 400, 599} {
		assertPixel(220, y, true)
		assertPixel(221, y, false)
	}
	s.SetLeftExpanded(false)
	tt.Frame()
	if r := main; r.X != 0 || r.W != 820 {
		t.Fatal("closed left separator reserved space", r)
	}
	assertPixel(0, 400, false)
	s.SetPage(func(*ui.Context) {})
	tt.Frame()
	assertPixel(220, 400, false)
}

func TestShellHasOnlyLeftSidebarControl(t *testing.T) {
	s, _ := fixture()
	tt := ui.NewTester(s.view, 820, 600)
	if _, ok := tt.Find("Toggle sidebar"); !ok {
		t.Fatal("missing sidebar control")
	}
	if _, ok := tt.Find("Toggle inspector"); ok {
		t.Fatal("unexpected inspector control")
	}
}

func TestShellSidebarBackground(t *testing.T) {
	for _, dark := range []bool{false, true} {
		s, _ := fixture()
		s.options.Left = nil
		s.options.Content = nil
		s.options.LeftToolbar = func(c *ui.Context) { ui.Spacer(c) }
		tt := ui.NewTester(s.view, 820, 600)
		tt.SetPreferences(ui.Preferences{ReduceMotion: true})
		tt.SetDark(dark)
		img := tt.Image()
		sidebar, main := img.RGBAAt(100, 400), img.RGBAAt(400, 400)
		wantAlpha := uint8(255)
		if sidebar == main || sidebar.A != wantAlpha || (dark && !(sidebar.R > sidebar.G && sidebar.G > sidebar.B)) {
			t.Fatalf("dark=%v: sidebar=%v main=%v", dark, sidebar, main)
		}
		for _, y := range []int{2, 599} {
			if got := img.RGBAAt(100, y); got != sidebar {
				t.Fatalf("dark=%v y=%d: sidebar=%v want=%v", dark, y, got, sidebar)
			}
		}
		s.SetLeftExpanded(false)
		tt.Frame()
		if got := tt.Image().RGBAAt(100, 400); got != main {
			t.Fatalf("collapsed sidebar left a background: %v", got)
		}
		s.SetLeftExpanded(true)
		s.SetPage(func(*ui.Context) {})
		tt.Frame()
		if got := tt.Image().RGBAAt(100, 400); got != main {
			t.Fatalf("replacement page retained sidebar background: %v", got)
		}
	}
}

func TestShellBackgroundsAndScaling(t *testing.T) {
	for _, dark := range []bool{false, true} {
		s, _ := fixture()
		tt := ui.NewTester(s.view, 1100, 600)
		tt.SetDark(dark)
		// Headless windows have no effective material by default.
		for _, x := range []int{100, 400, 1000} {
			for _, y := range []int{2, 400, 599} {
				_, _, _, a := tt.Image().At(x, y).RGBA()
				if a != 65535 {
					t.Fatalf("dark=%v x=%d y=%d alpha=%d want=65535", dark, x, y, a)
				}
			}
		}
		for _, label := range []string{"Navigation", "Document"} {
			r, ok := tt.Find(label)
			if !ok || r.Y != 0 {
				t.Fatalf("%q does not extend behind header: %+v", label, r)
			}
		}
		s.SetLeftExpanded(false)
		tt = ui.NewTester(s.view, 1100, 600)
		_, _, _, a := tt.Image().At(100, 2).RGBA()
		if a != 65535 {
			t.Fatal("collapsed sidebar left a transparent header gap")
		}
		s.SetPage(func(*ui.Context) {})
		tt = ui.NewTester(s.view, 820, 600)
		for _, y := range []int{2, 400} {
			_, _, _, a := tt.Image().At(100, y).RGBA()
			if a != 65535 {
				t.Fatal("replacement page is not opaque")
			}
		}
		s.SetPage(nil)
		tt = ui.NewTester(s.view, 360, 480)
		tt.SetPreferences(ui.Preferences{TextScale: 1.5})
		if !tt.HasText("Document") {
			t.Fatal(tt.Texts())
		}
	}
}

func TestShellBuildsLocalizedToasts(t *testing.T) {
	closed := 0
	s := newShell(ShellOptions{Content: func(c *ui.Context) {
		ui.Column(c).Padding(HeaderHeight(c), 0, 0, 0).Children(func() {
			if ui.Button(c, "Save").Clicked() {
				c.AddToast(ui.Toast{Title: "Saved", Timeout: -1, OnClose: func() { closed++ }})
			}
		})
	}})
	s.host = &fakeWindow{}
	tt := ui.NewTester(s.view, 820, 600)
	tt.SetLocale("zh-Hans")
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("关闭"); err != nil {
		t.Fatal(err)
	}
	if closed != 1 {
		t.Fatalf("toast close button closed %d toasts", closed)
	}
}

func TestShellToastsClearHeaderByDefault(t *testing.T) {
	s := newShell(ShellOptions{Content: func(c *ui.Context) {
		ui.Column(c).Padding(HeaderHeight(c), 0, 0, 0).Children(func() {
			if ui.Button(c, "Save").Clicked() {
				c.AddToast(ui.Toast{Title: "Saved", Timeout: -1})
			}
		})
	}})
	s.host = &fakeWindow{}
	tt := ui.NewTester(s.view, 820, 600)
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	var header float32
	ui.NewTester(func(c *ui.Context) { header = HeaderHeight(c) }, 820, 600)
	tt.Advance(time.Second) // The toast enters from beyond the window's edge.
	if r, ok := tt.Find("Saved"); !ok || r.Y < header || r.Y > header+80 {
		t.Fatalf("default toast does not sit below the header %v: %+v", header, r)
	}
	s.SetToasts(components.ToastOptions{Placement: components.ToastBottom})
	tt.Frame()
	if r, ok := tt.Find("Saved"); !ok || r.Y < 500 {
		t.Fatalf("bottom toast is not at the bottom: %+v", r)
	}
}
