package gallery

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
)

func TestCodeEditorExample(t *testing.T) {
	for _, lang := range []string{"en", "zh-Hans"} {
		g, s, tt := featureTester("Code editor", "Editable", lang, 540)
		if s.code == nil || s.code.DiffBase == nil || s.code.Label != g.text("codeEditorLabel") {
			t.Fatalf("editor state: %+v", s.code)
		}
		if err := tt.Click(g.text("codeEditorLabel")); err != nil {
			t.Fatal(err)
		}
		tt.Type("// edited\n")
		tt.Frame()
		if !s.code.Buffer.Modified() || s.code.Buffer.String() == s.code.DiffBase.String() {
			t.Fatal("typing did not edit the buffer")
		}
		if s.code.DiffBase.String() != editorSample {
			t.Fatal("editing changed the diff base")
		}
		pos := s.code.CursorPos()
		if !tt.HasText(g.messages().CodeEditorPosition(pos.Line+1, pos.Col+1)) {
			t.Fatalf("missing caret position %v: %q", pos, tt.Texts())
		}
	}
}

func TestDiffViewerExample(t *testing.T) {
	for _, variant := range []string{"Side by side", "Unified"} {
		g, s, tt := featureTester("Diff viewer", variant, "zh-Hans", 540)
		if want := map[string]editor.DiffLayout{"Side by side": editor.DiffSplit, "Unified": editor.DiffUnified}[variant]; s.diff.Layout != want {
			t.Fatalf("%s layout = %v", variant, s.diff.Layout)
		}
		added, deleted := s.diff.Stats()
		if added == 0 || deleted == 0 || !tt.HasText(g.messages().DiffViewerStats(added, deleted)) {
			t.Fatalf("%s stats +%d -%d: %q", variant, added, deleted, tt.Texts())
		}
		// Reverting every change restores the old text in the new buffer.
		for len(s.diff.Hunks()) > 0 {
			s.diff.Revert(0)
		}
		tt.Frame()
		if s.diff.New.String() != editorSample || !tt.HasText(g.messages().DiffViewerStats(0, 0)) {
			t.Fatalf("%s revert left %q", variant, tt.Texts())
		}
	}
}

func TestCanvasExample(t *testing.T) {
	g, s, tt := featureTester("Canvas", "Default", "en", 540)
	if s.value != 0 || s.canvas.smooth == nil || s.canvas.drawn != 0 {
		t.Fatalf("initial canvas state: value %v, drawn %d", s.value, s.canvas.drawn)
	}
	for _, label := range []string{"canvasSmooth", "canvasPixelated", "canvasLayer"} {
		if _, ok := tt.Find(g.text(label)); !ok {
			t.Fatalf("missing %s", label)
		}
	}
	if err := tt.Click(g.text("canvasNextFrame")); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click(g.text("canvasMirror")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	// The offscreen layer is redrawn for the new frame, mirrored.
	if s.index != 1 || !s.on || s.canvas.drawn != 3 {
		t.Fatalf("frame %d, mirrored %v, layer drawn for %d", s.index, s.on, s.canvas.drawn)
	}
	for range spriteFrames - 1 {
		if err := tt.Click(g.text("canvasNextFrame")); err != nil {
			t.Fatal(err)
		}
	}
	if s.index != 0 {
		t.Fatalf("frames did not wrap: %d", s.index)
	}
}

func TestSpriteSheet(t *testing.T) {
	img := spriteSheet()
	if b := img.Bounds(); b.Dx() != spriteSize*spriteFrames || b.Dy() != spriteSize {
		t.Fatalf("sheet bounds %v", b)
	}
	for f, rows := range spriteRows {
		for y, row := range rows {
			if len(row) != spriteSize {
				t.Fatalf("frame %d row %d has %d pixels", f, y, len(row))
			}
			for _, p := range []byte(row) {
				if _, ok := spritePalette[p]; !ok && p != '.' {
					t.Fatalf("frame %d uses unknown pixel %q", f, p)
				}
			}
		}
	}
}

func TestGameExampleMovesWithHeldKeys(t *testing.T) {
	_, s, tt := featureTester("Game loop", "Default", "en", 540)
	d := &s.game
	if d.gamepads {
		t.Fatal("headless views must not read gamepads")
	}
	x, y := d.x, d.y
	tt.KeyDown(ui.KeyD)
	tt.KeyDown(ui.KeyS)
	tt.Advance(120 * time.Millisecond)
	tt.KeyUp(ui.KeyD)
	tt.KeyUp(ui.KeyS)
	if d.steps == 0 || d.x <= x || d.y <= y || d.left {
		t.Fatalf("after %d steps: (%v, %v) from (%v, %v), left %v", d.steps, d.x, d.y, x, y, d.left)
	}
	// The player stays inside the world.
	d.x, d.y = gameWorld, -gameWorld
	tt.KeyDown(ui.KeyA)
	tt.Advance(40 * time.Millisecond)
	tt.KeyUp(ui.KeyA)
	if d.x > gameWorld-spriteSize/2 || d.y < spriteSize/2 || !d.left {
		t.Fatalf("player left the world: (%v, %v), left %v", d.x, d.y, d.left)
	}
}

func TestNewMyGoReviewEntries(t *testing.T) {
	entries := map[string]mygoReviewEntry{}
	for _, entry := range mygoReviewEntries(mygoDiscovered, mygoExamples, mygoHandled) {
		entries[entry.Name] = entry
	}
	for _, name := range []string{"Code editor", "Diff viewer"} {
		entry := entries[name]
		if entry.Status != mygoPrimitive || entry.Source != "ui/editor" || !entry.HasExample || entry.Reason == "" {
			t.Fatalf("%s: %+v", name, entry)
		}
	}
	for _, name := range []string{"Canvas", "Game loop"} {
		if entry := entries[name]; entry.Status != mygoPrimitive || entry.Reason == "" {
			t.Fatalf("%s: %+v", name, entry)
		}
	}
}

func TestCodeAndGameGroups(t *testing.T) {
	want := map[string][]string{"Code & diffs": {"Code editor", "Diff viewer"}, "Games": {"Canvas", "Shaders", "Game loop", "3D renderer"}}
	for _, group := range mygoNavigationGroups() {
		if entries, ok := want[group.label]; ok {
			var got []string
			for _, entry := range group.entries {
				got = append(got, entry.label)
			}
			// Pages in New or Updated components leave their purpose group.
			entries = slices.DeleteFunc(slices.Clone(entries), func(name string) bool {
				return newPages["mygo/"+name] || updatedPages["mygo/"+name]
			})
			if fmt.Sprint(got) != fmt.Sprint(entries) {
				t.Fatalf("%s = %v, want %v", group.label, got, entries)
			}
			delete(want, group.label)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing groups: %v", want)
	}
}

// catalogWithIndexes returns the navigation groups with New components and
// Updated components at 1 and 2 after Overview, empty where the sidebar
// leaves them out for having no pages.
func catalogWithIndexes(t *testing.T) []navigationGroup {
	t.Helper()
	groups := navigationCatalog()
	if len(groups) == 0 || groups[0].label != "Overview" {
		t.Fatalf("navigation must start with Overview: %v", groups)
	}
	for i, label := range []string{newGroupLabel, updatedGroupLabel} {
		entries := newNavigationEntries(false)
		if label == updatedGroupLabel {
			entries = updatedNavigationEntries(false)
		}
		at := i + 1
		shown := len(groups) > at && groups[at].label == label
		switch {
		case len(entries) == 0 && shown:
			t.Fatalf("empty %s group is shown", label)
		case len(entries) == 0:
			groups = slices.Insert(groups, at, navigationGroup{label: label})
		case !shown:
			t.Fatalf("%s group must follow %s: %v", label, groups[at-1].label, groups)
		}
	}
	return groups
}

func TestNewComponentsGroup(t *testing.T) {
	groups := catalogWithIndexes(t)
	mygoGroups := mygoNavigationGroups()
	// The MyGo list has a New components group only while MyGo pages are new.
	if len(newNavigationEntries(true)) > 0 {
		if mygoGroups[0].label != newGroupLabel {
			t.Fatalf("MyGo list must start with its New components group: %v", mygoGroups[0])
		}
	} else {
		mygoGroups = append([]navigationGroup{{}}, mygoGroups...)
	}
	want := map[galleryPage]bool{}
	for key := range newPages {
		if pageKey(pageFromKey(key)) != key {
			t.Fatalf("new page %q has no route", key)
		}
		want[pageFromKey(key)] = true
	}
	// Onevra pages are new in the sidebar, MyGo pages in the MyGo entry's
	// in-page list; each appears only there, not again in a purpose group.
	for _, entry := range groups[1].entries {
		if isMyGoPage(entry.page) {
			t.Fatalf("MyGo page %s is listed outside the MyGo entry", entry.label)
		}
	}
	for _, entry := range append(slices.Clone(groups[1].entries), mygoGroups[0].entries...) {
		if !want[entry.page] {
			t.Fatalf("unexpected New components entry: %s", entry.label)
		}
		delete(want, entry.page)
	}
	if len(want) != 0 {
		t.Fatalf("missing New components entries: %v", want)
	}
	for _, group := range append(groups[2:], mygoGroups[1:]...) {
		for _, entry := range group.entries {
			if newPages[pageKey(entry.page)] {
				t.Fatalf("%s repeats in %s", entry.label, group.label)
			}
		}
	}
}

func TestUpdatedComponentsGroup(t *testing.T) {
	groups := catalogWithIndexes(t)
	want := map[galleryPage]bool{}
	for key := range updatedPages {
		if newPages[key] {
			t.Fatalf("updated page %q is also new", key)
		}
		if pageKey(pageFromKey(key)) != key {
			t.Fatalf("updated page %q has no route", key)
		}
		want[pageFromKey(key)] = true
	}
	mygoGroups := mygoNavigationGroups()
	// The MyGo list omits an empty New or Updated components group; insert
	// placeholders so both keep their fixed positions here.
	if len(newNavigationEntries(true)) == 0 {
		mygoGroups = append([]navigationGroup{{}}, mygoGroups...)
	} else if mygoGroups[0].label != newGroupLabel {
		t.Fatalf("MyGo list must start with its New components group: %v", mygoGroups[0])
	}
	if len(updatedNavigationEntries(true)) == 0 {
		mygoGroups = slices.Insert(mygoGroups, 1, navigationGroup{})
	} else if len(mygoGroups) < 2 || mygoGroups[1].label != updatedGroupLabel {
		t.Fatalf("MyGo list must follow its New components group with Updated: %v", mygoGroups)
	}
	mygoUpdated := mygoGroups[1].entries
	for _, entry := range groups[2].entries {
		if isMyGoPage(entry.page) {
			t.Fatalf("MyGo page %s is listed outside the MyGo entry", entry.label)
		}
		if !want[entry.page] {
			t.Fatalf("unexpected Updated components entry: %s", entry.label)
		}
		delete(want, entry.page)
	}
	for _, entry := range mygoUpdated {
		if !want[entry.page] {
			t.Fatalf("unexpected MyGo Updated components entry: %s", entry.label)
		}
		delete(want, entry.page)
	}
	if len(want) != 0 {
		t.Fatalf("missing Updated components entries: %v", want)
	}
	// Updated pages leave their purpose group, as new pages do. MyGo pages
	// belong to the MyGo entry's own Updated group, when it has any.
	for _, group := range append(groups[3:], mygoGroups[2:]...) {
		for _, entry := range group.entries {
			if updatedPages[pageKey(entry.page)] {
				t.Fatalf("%s repeats in %s", entry.label, group.label)
			}
		}
	}
}

func TestUpdatedComponentsGroupRenders(t *testing.T) {
	for _, lang := range []string{"en", "zh-Hans"} {
		g := &Gallery{language: newTranslator(lang, "en")}
		tt := ui.NewTester(g.Left, 220, 6000)
		index, ok := tt.Find(g.text(updatedGroupLabel))
		if len(updatedNavigationEntries(false)) == 0 {
			if ok {
				t.Fatalf("%s: empty Updated components group is shown", lang)
			}
			continue
		}
		if !ok {
			t.Fatalf("%s: missing Updated components group", lang)
		}
		lucide, ok := tt.Find(g.text("Lucide"))
		if !ok || index.Y >= lucide.Y {
			t.Fatalf("%s: Updated components group is not above the icon library", lang)
		}
		for _, entry := range updatedNavigationEntries(false) {
			if !tt.HasText(g.text(entry.label)) {
				t.Fatalf("%s: missing %s", lang, entry.label)
			}
		}
	}
}

func TestNewComponentsGroupRenders(t *testing.T) {
	for _, lang := range []string{"en", "zh-Hans"} {
		g := &Gallery{language: newTranslator(lang, "en")}
		tt := ui.NewTester(g.Left, 220, 6000)
		index, ok := tt.Find(g.text(newGroupLabel))
		if len(newNavigationEntries(false)) == 0 {
			if ok {
				t.Fatalf("%s: empty New components group is shown", lang)
			}
			continue
		}
		if !ok {
			t.Fatalf("%s: missing New components group", lang)
		}
		lucide, ok := tt.Find(g.text("Lucide"))
		if !ok || index.Y >= lucide.Y {
			t.Fatalf("%s: New components group is not above the icon library", lang)
		}
		for _, entry := range newNavigationEntries(false) {
			if !tt.HasText(g.text(entry.label)) {
				t.Fatalf("%s: missing %s", lang, entry.label)
			}
		}
	}
}

func TestHiddenElementGivesBackItsRoom(t *testing.T) {
	g, s, tt := featureTester("Hidden elements", "Default", "zh-Hans", 540)
	if !tt.HasText(g.text("second")) {
		t.Fatalf("missing middle box: %q", tt.Texts())
	}
	if err := tt.Click(g.text("boxHideMiddle")); err != nil {
		t.Fatal(err)
	}
	first, _ := tt.Find(g.text("first"))
	third, _ := tt.Find(g.text("third"))
	if !s.on || tt.HasText(g.text("second")) || third.X-first.X > 200 {
		t.Fatalf("hidden box kept its room: first %v, third %v", first, third)
	}
}

func TestZoomExampleZoomsTheWindow(t *testing.T) {
	g, s, tt := featureTester("Zoom", "Default", "en", 540)
	if err := tt.Click(g.text("zoomIn")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if s.count != 1 || !tt.HasText(g.messages().Percent(110)) {
		t.Fatalf("zoom step %d: %q", s.count, tt.Texts())
	}
	if err := tt.Click(g.text("zoomActualSize")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if s.count != 0 || !tt.HasText(g.messages().Percent(100)) {
		t.Fatalf("reset zoom step %d: %q", s.count, tt.Texts())
	}
}

func TestCodeEditorHighlightsMatches(t *testing.T) {
	g, s, tt := featureTester("Code editor", "Editable", "en", 540)
	if err := tt.Click(g.text("codeEditorHighlight")); err != nil {
		t.Fatal(err)
	}
	tt.Type("name")
	tt.Frame()
	want := strings.Count(editorSample, "name")
	if len(s.code.TextBackgrounds) != want || want == 0 {
		t.Fatalf("highlighted %d of %d matches", len(s.code.TextBackgrounds), want)
	}
	for _, bg := range s.code.TextBackgrounds {
		if editorSample[bg.Start:bg.End] != "name" {
			t.Fatalf("highlight covers %q", editorSample[bg.Start:bg.End])
		}
	}
	s.query = ""
	tt.Frame()
	if len(s.code.TextBackgrounds) != 0 {
		t.Fatal("clearing the query kept highlights")
	}
}

func TestInlineMathKeepsItsFormulas(t *testing.T) {
	render := inlineMath()
	a, ok := render(`2\pi \cdot r`, 14)
	if !ok || a.Width <= 0 || a.Height <= 0 || a.Depth < 0 || a.Draw == nil {
		t.Fatalf("formula %+v, %v", a, ok)
	}
	b, _ := render(`2\pi \cdot r`, 14)
	larger, _ := render(`2\pi \cdot r`, 28)
	if b.Width != a.Width || larger.Width <= a.Width {
		t.Fatalf("widths %v, %v, %v", a.Width, b.Width, larger.Width)
	}
	if got := texSymbols.Replace(`a^2 + 2\pi \cdot r`); got != "a² + 2π · r" {
		t.Fatalf("symbols %q", got)
	}
}

func TestFittedFlyoutPage(t *testing.T) {
	g := &Gallery{language: newTranslator("ja", "en")}
	var page string
	ui.NewTester(func(c *ui.Context) {
		g.withLanguage(c, func() { page = flyoutHTML(c) })
	}, 200, 200)
	for _, key := range []string{"flyoutFittedBody", "flyoutFittedAdd", "flyoutFittedItem"} {
		if !strings.Contains(page, g.text(key)) {
			t.Fatalf("page lacks %s: %s", key, page)
		}
	}
}
