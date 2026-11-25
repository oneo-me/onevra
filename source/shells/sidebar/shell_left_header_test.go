package sidebar

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

func TestShellLeftHeader(t *testing.T) {
	clicks := 0
	s := newShell(ShellOptions{LeftWidth: 252, LeftHeader: func(c *ui.Context) {
		IconButton(c, lucide.ArrowLeft(), "Back").OnClick(func() { clicks++ })
		IconButton(c, lucide.ArrowRight(), "Forward")
	}, ContentTitle: func(c *ui.Context) { ui.Text(c, "Document") }})
	tt := ui.NewTester(s.view, 980, 820)
	tt.SetTitleBar(ui.TitleBar{Height: 48, Left: 80})
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	toggle, _ := tt.Find("Toggle sidebar")
	back, ok := tt.Find("Back")
	if !ok || back.X < toggle.X+toggle.W || back.Y != toggle.Y {
		t.Fatalf("back not beside toggle: %+v %+v", toggle, back)
	}
	forward, ok := tt.Find("Forward")
	if !ok || forward.X < back.X+back.W || forward.X+forward.W != 244 || forward.Y != toggle.Y {
		t.Fatalf("forward not trailing in sidebar header: %+v", forward)
	}
	if err := tt.Click("Back"); err != nil {
		t.Fatal(err)
	}
	if clicks != 1 {
		t.Fatal("header drag strip swallowed action")
	}
	title, _ := tt.Find("Document")
	if title.X != 268 {
		t.Fatalf("actions moved content title: %+v", title)
	}
	if err := tt.Click("Toggle sidebar"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("Back"); ok {
		t.Fatal("collapsed sidebar retained header actions")
	}
	s.SetLeftExpanded(true)
	tt.Frame()
	s.SetPage(func(*ui.Context) {})
	tt.Frame()
	if _, ok := tt.Find("Back"); ok {
		t.Fatal("replacement page retained header actions")
	}
	s.SetPage(nil)
	tt.Frame()
	if _, ok := tt.Find("Back"); !ok {
		t.Fatal("workspace did not restore header actions")
	}
}

func TestShellLeftToggleTrailing(t *testing.T) {
	options := ShellOptions{LeftWidth: 252, LeftHeader: func(c *ui.Context) {
		IconButton(c, lucide.ArrowLeft(), "Back")
	}, ContentTitle: func(c *ui.Context) { ui.Text(c, "Document") }}
	tester := func(s *Shell) *ui.Tester {
		tt := ui.NewTester(s.view, 980, 820)
		tt.SetTitleBar(ui.TitleBar{Height: 48, Left: 80})
		return tt
	}
	find := func(tt *ui.Tester, label string) ui.Rect {
		t.Helper()
		r, ok := tt.Find(label)
		if !ok {
			t.Fatalf("missing %q", label)
		}
		return r
	}
	leading := tester(newShell(options))
	leading.SetPreferences(ui.Preferences{ReduceMotion: true})
	leadingToggle := find(leading, "Toggle sidebar")

	options.LeftToggleTrailing = true
	s := newShell(options)
	tt := tester(s)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	toggle, back := find(tt, "Toggle sidebar"), find(tt, "Back")
	if toggle.X+toggle.W != 244 || toggle.Y != leadingToggle.Y {
		t.Fatalf("toggle not trailing in sidebar header: %+v", toggle)
	}
	if back.X+back.W != toggle.X || back.Y != toggle.Y {
		t.Fatalf("header actions not before toggle: %+v %+v", back, toggle)
	}
	if title := find(tt, "Document"); title.X != find(leading, "Document").X {
		t.Fatalf("trailing toggle moved content title: %+v", title)
	}
	if err := tt.Click("Toggle sidebar"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if s.LeftExpanded() || find(tt, "Toggle sidebar") != leadingToggle {
		t.Fatal("collapsed sidebar did not keep the leading toggle")
	}
	tt.SetPreferences(ui.Preferences{})
	s.SetLeftExpanded(true)
	tt.Frame()
	tt.Advance(55 * time.Millisecond)
	tt.Frame()
	if find(tt, "Toggle sidebar") != leadingToggle {
		t.Fatal("animating sidebar moved the toggle")
	}
	tt.Advance(sidebarDuration + 30*time.Millisecond)
	tt.Frame()
	if r := find(tt, "Toggle sidebar"); r.X+r.W != 244 {
		t.Fatalf("expanded sidebar did not restore trailing toggle: %+v", r)
	}
	s.SetPage(func(*ui.Context) {})
	tt.Frame()
	if _, ok := tt.Find("Toggle sidebar"); ok {
		t.Fatal("replacement page retained the toggle")
	}
}

func TestShellTitleInsetWithTrailingToggle(t *testing.T) {
	options := ShellOptions{LeftWidth: 252, TitleInsetLeft: 12, LeftHeader: func(c *ui.Context) {
		IconButton(c, lucide.ArrowLeft(), "Back")
	}, ContentTitle: func(c *ui.Context) { ui.Text(c, "Document") }}
	title := func(trailing bool) ui.Rect {
		options.LeftToggleTrailing = trailing
		s := newShell(options)
		tt := ui.NewTester(s.view, 980, 820)
		tt.SetTitleBar(ui.TitleBar{Height: 48, Left: 80})
		tt.SetPreferences(ui.Preferences{ReduceMotion: true})
		r, ok := tt.Find("Document")
		if !ok {
			t.Fatal("missing title")
		}
		return r
	}
	leading, trailing := title(false), title(true)
	if trailing.X != leading.X {
		t.Fatalf("trailing toggle moved the inset title: %+v vs %+v", leading, trailing)
	}
	if leading.X != 264 { // 252+12: LeftWidth + TitleInsetLeft
		t.Fatalf("TitleInsetLeft did not shift the aligned title: %+v", leading)
	}
}
