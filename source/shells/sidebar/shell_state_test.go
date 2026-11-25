package sidebar

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestLeftWidthPersistence(t *testing.T) {
	s := newShell(ShellOptions{LeftWidth: 220, LeftMinWidth: 120, LeftMaxWidth: 520})
	for _, tc := range []struct{ input, want float32 }{{315.5, 315.5}, {0, 120}, {900, 520}, {float32(math.NaN()), 520}, {float32(math.Inf(1)), 520}} {
		s.SetLeftWidth(tc.input)
		if got := s.LeftWidth(); got != tc.want {
			t.Fatalf("width=%v want=%v", got, tc.want)
		}
	}
	s.SetLeftWidth(300)
	s.SetLeftExpanded(false)
	s.SetPage(func(*ui.Context) {})
	s.SetPage(nil)
	if s.LeftWidth() != 300 || s.options.LeftWidth != 220 {
		t.Fatal("lost chosen width or changed reset default")
	}
}

func TestLanguageChangePreservesShell(t *testing.T) {
	s := newShell(ShellOptions{LeftWidth: 220})
	s.SetLeftWidth(300)
	tt := ui.NewTester(s.view, 820, 600)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	if _, ok := tt.Find("Toggle sidebar"); !ok {
		t.Fatal("English label missing")
	}
	tt.SetLocale("zh-Hans")
	if _, ok := tt.Find("展开或收起侧边栏"); !ok {
		t.Fatal("localized label missing")
	}
	if _, ok := tt.Find("调整侧边栏宽度"); !ok {
		t.Fatal("localized resize label missing")
	}
	if s.LeftWidth() != 300 || !s.LeftExpanded() {
		t.Fatal("language change reset sidebar")
	}
	tt.SetLocale("en")
	if _, ok := tt.Find("Toggle sidebar"); !ok {
		t.Fatal("English defaults not restored")
	}
}

func TestShellLeftExpandedCallback(t *testing.T) {
	s, _ := fixture()
	var calls []bool
	s.options.OnLeftExpanded = func(expanded bool) { calls = append(calls, expanded) }
	tt := ui.NewTester(s.view, 820, 600)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	if len(calls) != 0 {
		t.Fatal("construction notified the callback", calls)
	}
	s.SetLeftExpanded(true) // same as the initial state: no call
	s.SetLeftExpanded(false)
	s.SetLeftExpanded(true)
	if len(calls) != 2 || calls[0] != false || calls[1] != true {
		t.Fatal("callback sequence wrong", calls)
	}
	if err := tt.Click("Toggle sidebar"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(calls) != 3 || calls[2] != false {
		t.Fatal("toggle click did not notify", calls)
	}
}

func TestShellLeftExpandedCallbackNil(t *testing.T) {
	s, _ := fixture() // OnLeftExpanded nil: toggling must not panic
	tt := ui.NewTester(s.view, 820, 600)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	s.SetLeftExpanded(false)
	s.SetLeftExpanded(true)
	tt.Frame()
	if !s.LeftExpanded() {
		t.Fatal("nil callback broke the toggle")
	}
}
