package components

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

// menuTexts returns the texts tt shows from first on, without the
// accessible labels that repeat them.
func menuTexts(tt *ui.Tester, first string) []string {
	texts := tt.Texts()
	i := slices.Index(texts, first)
	if i < 0 {
		return nil
	}
	var out []string
	for _, s := range texts[i:] {
		if len(out) == 0 || out[len(out)-1] != s {
			out = append(out, s)
		}
	}
	return out
}

func TestContextMenu(t *testing.T) {
	var chosen []string
	pinned := false
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(20).Children(func() {
			row := ui.ButtonBase(c).Key("row").Label("Report").Size(200, 40)
			ContextMenu(c, row, func(m *Menu) {
				// A separator leading the menu does not show.
				m.Separator()
				if m.Item("Open").Shortcut("⌘O").Chosen() {
					chosen = append(chosen, "Open")
				}
				if m.Item("Pinned").Checked(pinned).Chosen() {
					pinned = !pinned
				}
				m.Group("Move to", func(m *Menu) {
					for _, f := range []string{"Desktop", "Documents"} {
						if m.Item(f).Chosen() {
							chosen = append(chosen, f)
						}
					}
				})
				if m.Item("Delete").Disabled(true).Chosen() {
					chosen = append(chosen, "Delete")
				}
			})
		})
	}, 500, 400)
	if tt.HasText("Open") {
		t.Fatal("menu shows before it opens")
	}
	if err := tt.RightClick("Report"); err != nil {
		t.Fatal(err)
	}
	if got := menuTexts(tt, "Open"); !slices.Equal(got, []string{"Open", "⌘O", "Pinned", "Move to", "Desktop", "Documents", "Delete"}) {
		t.Fatalf("menu: %q", got)
	}
	// It opens where it was clicked, the center of the row.
	if r, _ := tt.Find("Open"); r.Y < 40 || r.X < 100 {
		t.Fatalf("menu at %+v", r)
	}
	// The heading and a disabled item cannot be chosen.
	for _, label := range []string{"Move to", "Delete"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
		if !tt.HasText("Open") || len(chosen) != 0 {
			t.Fatalf("%s closed the menu or was chosen: %v", label, chosen)
		}
	}
	if err := tt.Click("Documents"); err != nil {
		t.Fatal(err)
	}
	if tt.HasText("Open") || !slices.Equal(chosen, []string{"Documents"}) {
		t.Fatalf("choosing: %v", chosen)
	}
	// The keys move past the disabled item, round, and choose.
	if err := tt.RightClick("Report"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyUp)
	tt.Key(0, ui.KeyUp)
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(chosen, []string{"Documents", "Desktop"}) {
		t.Fatalf("keys chose %v", chosen)
	}
	if err := tt.RightClick("Report"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if !pinned || tt.HasText("Open") {
		t.Fatal("Pinned was not chosen")
	}
	// Escape closes it.
	if err := tt.RightClick("Report"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEscape)
	if tt.HasText("Open") {
		t.Fatal("Escape left the menu open")
	}
	// Shift+F10 opens it below the focus, away from the pointer, on its
	// first item.
	if err := tt.Click("Report"); err != nil {
		t.Fatal(err)
	}
	tt.Move(490, 390)
	tt.Key(ui.Shift, ui.KeyF10)
	if r, _ := tt.Find("Open"); r.Y < 60 || r.X > 40 {
		t.Fatalf("menu at %+v", r)
	}
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(chosen, []string{"Documents", "Desktop", "Open"}) {
		t.Fatalf("keys chose %v", chosen)
	}
}

func TestMenuButton(t *testing.T) {
	chosen := ""
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(20).Children(func() {
			b := Button(c, "Actions", ButtonSecondary)
			MenuButton(c, b, func(m *Menu) {
				for _, action := range []string{"Duplicate", "Rename"} {
					if m.Item(action).Chosen() {
						chosen = action
					}
				}
			})
		})
	}, 400, 300)
	click := func(label string) {
		t.Helper()
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	click("Actions")
	if !tt.HasText("Rename") {
		t.Fatal("menu did not open")
	}
	click("Actions")
	if tt.HasText("Rename") {
		t.Fatal("a second click left the menu open")
	}
	click("Actions")
	click("Rename")
	if chosen != "Rename" || tt.HasText("Rename") {
		t.Fatalf("chose %q", chosen)
	}
}

func TestContextMenuLeavesClicksToParents(t *testing.T) {
	clicks, chosen := 0, false
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(20).Children(func() {
			row := ui.ButtonBase(c).Key("row").Label("Row").Padding(8)
			row.Children(func() {
				label := ui.Box(c).Key("inside").Children(func() { Text(c, "Inside") })
				ContextMenu(c, label, func(m *Menu) {
					if m.Item("Copy").Chosen() {
						chosen = true
					}
				})
			})
			if row.Clicked() {
				clicks++
			}
		})
	}, 400, 300)
	if err := tt.Click("Inside"); err != nil {
		t.Fatal(err)
	}
	if clicks != 1 || tt.HasText("Copy") {
		t.Fatalf("clicks %d", clicks)
	}
	if err := tt.RightClick("Inside"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Copy"); err != nil {
		t.Fatal(err)
	}
	if !chosen || clicks != 1 {
		t.Fatalf("chosen %v, clicks %d", chosen, clicks)
	}
}
