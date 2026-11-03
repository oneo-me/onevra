package sidebar

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// The right sidebar spans the window's height beside the content; the
// header's title ends before it, and its own header shows, ending with the
// sidebar's toggle, which moves to the main header once it closes.
func TestShellRightSidebar(t *testing.T) {
	var content, right ui.Rect
	s := newShell(ShellOptions{
		RightWidth: 300,
		Content:    func(c *ui.Context) { ui.Box(c).Fill().DrawOver(func(_ *ui.Painter, r ui.Rect) { content = r }) },
		Right:      func(c *ui.Context) { ui.Box(c).Fill().DrawOver(func(_ *ui.Painter, r ui.Rect) { right = r }) },
		RightHeader: func(c *ui.Context) {
			ui.Text(c, "Right title")
		},
		ContentTitle: func(c *ui.Context) { ui.Text(c, "Content title") },
	})
	tt := ui.NewTester(s.view, 1000, 600)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	tt.Frame()
	if right.X != 700 || right.Y != 0 || right.W != 300 || right.H != 600 || content.X+content.W != 700 {
		t.Fatalf("layout: content %+v, right %+v", content, right)
	}
	title, ok := tt.Find("Right title")
	if !ok || title.X < 700 {
		t.Fatalf("right header: %+v %v", title, ok)
	}
	toggle, ok := tt.Find("Toggle right sidebar")
	if !ok || toggle.X < title.X+title.W || toggle.X+toggle.W > 1000-8 {
		t.Fatalf("toggle in the right header: %+v %v", toggle, ok)
	}

	handle, ok := tt.Find("Resize right sidebar")
	if !ok {
		t.Fatal("missing resize handle")
	}
	x, y := handle.X+handle.W/2, handle.Y+30
	tt.Press(x, y)
	tt.Move(x-100, y)
	tt.Release(x-100, y)
	if s.RightWidth() != 400 || right.W != 400 {
		t.Fatalf("drag: width %v, right %+v", s.RightWidth(), right)
	}

	if err := tt.Click("Toggle right sidebar"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if s.RightExpanded() || content.X+content.W != 1000 {
		t.Fatalf("collapsed: %v, content %+v", s.RightExpanded(), content)
	}
	if _, ok := tt.Find("Right title"); ok {
		t.Fatal("collapsed sidebar kept its header")
	}
	contentTitle, ok := tt.Find("Content title")
	if toggle, found := tt.Find("Toggle right sidebar"); !found || !ok || toggle.X < contentTitle.X+contentTitle.W {
		t.Fatalf("toggle in the main header: %+v %v", toggle, found)
	}
}

// Without Right, the shell has no right sidebar or toggle.
func TestShellWithoutRightSidebar(t *testing.T) {
	s := newShell(ShellOptions{})
	tt := ui.NewTester(s.view, 1000, 600)
	if _, ok := tt.Find("Toggle right sidebar"); ok || s.RightExpanded() {
		t.Fatal("right sidebar without Right")
	}
}
