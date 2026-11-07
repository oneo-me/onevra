package treemap

import (
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/theme"
)

var sample = []Node{
	{ID: "work", Label: "work", Children: []Node{
		{ID: "work/app", Label: "app", Children: []Node{
			{ID: "work/app/target", Label: "draft", Caption: "Reclaimable", Value: 960, Measure: 2},
			{ID: "work/app/web", Label: "web", Value: 80, Measure: 10},
		}},
		{ID: "work/kit", Label: "kit", Value: 750, Measure: -4},
	}},
	{ID: "cache", Label: "cache", Value: 1100, Children: []Node{
		{ID: "cache/mise", Label: "mise", Value: 400, Measure: 1},
	}},
	{ID: "local", Label: "local", Value: 900, Measure: 6},
	{ID: "empty", Label: "empty"},
}

func TestSquarify(t *testing.T) {
	values := []float64{6, 6, 4, 3, 2, 2, 1}
	r := ui.Rect{X: 10, Y: 20, W: 600, H: 400}
	rects := squarify(values, r)
	total := 24.0
	for i, got := range rects {
		want := values[i] / total * float64(r.W*r.H)
		if area := float64(got.W * got.H); math.Abs(area-want) > want*1e-3 {
			t.Fatalf("rect %d area %g, want %g", i, area, want)
		}
		if got.X < r.X-1e-3 || got.Y < r.Y-1e-3 || got.X+got.W > r.X+r.W+1e-3 || got.Y+got.H > r.Y+r.H+1e-3 {
			t.Fatalf("rect %d %+v outside %+v", i, got, r)
		}
		if ratio := max(got.W/got.H, got.H/got.W); ratio > 3 {
			t.Fatalf("rect %d aspect ratio %g", i, ratio)
		}
	}
}

func TestSizesMeasuresAndPath(t *testing.T) {
	var v view
	v.sizes, v.measures = make(map[*Node]float64), make(map[*Node]float64)
	if got := v.size(&sample[0]); got != 1790 {
		t.Fatalf("branch size %g, want the children's total", got)
	}
	if got, free := v.size(&sample[1]), v.free(&sample[1]); got != 1100 || free != 700 {
		t.Fatalf("larger branch value: size %g, free %g", got, free)
	}
	if order := v.order(sample, 0); len(order) != 3 || order[0] != 0 || order[2] != 2 {
		t.Fatalf("order %v", order)
	}
	if got, want := v.measure(&sample[0].Children[0]), (2*960+10*80)/1040.0; math.Abs(got-want) > 1e-9 {
		t.Fatalf("branch measure %g, want the size-weighted mean %g", got, want)
	}
	if got := v.leafRange(sample); got != [2]float64{-4, 10} {
		t.Fatalf("leaf range %v", got)
	}
	path := Path(sample, "work/app/web")
	if len(path) != 3 || path[0].ID != "work" || path[2].Label != "web" {
		t.Fatalf("path %+v", path)
	}
	if Path(sample, "missing") != nil || Path(sample, "") != nil {
		t.Fatal("missing nodes have no path")
	}
}

// tester shows sample in a 600×400 view at scale, without its header.
func tester(state *State, data *[]Node, opts Options, scale float32) (*ui.Tester, func(id string) (float32, float32)) {
	tt := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Size(600, 400).Scale(scale, scale).TransformOrigin(0, 0).Children(func() {
			View(c, state, *data, opts)
		})
	}, int(600*scale), int(400*scale))
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	tt.Frame()
	at := func(id string) (float32, float32) {
		for _, tl := range state.tiles {
			if tl.id == id {
				return (tl.rect.X + tl.rect.W/2) * scale, (tl.rect.Y + tl.rect.H/2) * scale
			}
		}
		return -1, -1
	}
	return tt, at
}

func TestPointerDrilling(t *testing.T) {
	for _, scale := range []float32{1, 1.5} {
		state, data := State{}, sample
		var opened []string
		tt, at := tester(&state, &data, Options{Label: "Usage", HideHeader: true, Open: func(id string) { opened = append(opened, id) }}, scale)
		x, y := at("work/app/target")
		if x < 0 {
			t.Fatalf("scale %g: no nested tile for target", scale)
		}
		// Clicks drill one level at a time into the shown level's tile.
		for _, want := range []string{"work", "work/app"} {
			tt.ClickAt(x, y)
			if state.Root != want {
				t.Fatalf("scale %g: click drilled to %q, want %q", scale, state.Root, want)
			}
			tt.Frame()
			x, y = at("work/app/target")
		}
		for _, tl := range state.tiles {
			if tl.id == "work" || tl.id == "cache" {
				t.Fatalf("scale %g: drilled view still tiles %q", scale, tl.id)
			}
		}
		tt.Move(x, y)
		if !tt.HasText("work / app") || !tt.HasText("960 · 92%") {
			t.Fatalf("scale %g: tip lacks the path or share: %q", scale, tt.Texts())
		}
		tt.ClickAt(x, y)
		if len(opened) != 1 || opened[0] != "work/app/target" || state.Active != "work/app/target" {
			t.Fatalf("scale %g: clicking a leaf opened %v, active %q", scale, opened, state.Active)
		}
		tt.Key(0, ui.KeyEscape)
		if state.Root != "work" || state.Active != "work/app" {
			t.Fatalf("scale %g: Escape left root %q, active %q", scale, state.Root, state.Active)
		}
		tt.Key(0, ui.KeyBackspace)
		if state.Root != "" || state.Active != "work" {
			t.Fatalf("scale %g: Backspace left root %q, active %q", scale, state.Root, state.Active)
		}
	}
}

func TestCanceledPressDrillsNothing(t *testing.T) {
	state, data := State{}, sample
	tt, at := tester(&state, &data, Options{Label: "Usage", HideHeader: true}, 1)
	x, y := at("work/app/target")
	tt.Press(x, y)
	tt.CancelPress()
	tt.Frame()
	if state.Root != "" {
		t.Fatalf("canceled press drilled to %q", state.Root)
	}
}

func TestKeyboard(t *testing.T) {
	state, data := State{}, sample
	tt, _ := tester(&state, &data, Options{Label: "Usage", HideHeader: true, MeasureLabel: "Growth"}, 1)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyHome)
	if state.Active != "work" {
		t.Fatalf("Home chose %q", state.Active)
	}
	tt.Key(0, ui.KeyRight)
	right := state.Active
	if right == "work" || right == "" {
		t.Fatal("Right stayed on the largest tile")
	}
	tt.Key(0, ui.KeyLeft)
	if state.Active != "work" {
		t.Fatalf("Left went to %q", state.Active)
	}
	tt.Key(0, ui.KeyLeft)
	if state.Active != "work" {
		t.Fatal("Left passed the edge")
	}
	tt.Key(0, ui.KeyEnd)
	if state.Active != "local" {
		t.Fatalf("End chose %q", state.Active)
	}
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyEnter)
	if state.Root != "work" || state.Active != "work/app" {
		t.Fatalf("Enter drilled to %q, active %q", state.Root, state.Active)
	}
	announced := tt.Announcements()
	if len(announced) == 0 || !strings.Contains(announced[len(announced)-1], "Growth: ") {
		t.Fatalf("announcements lack the measure: %q", announced)
	}
	tt.Key(0, ui.KeySpace)
	if state.Root != "work/app" || state.Active != "work/app/target" {
		t.Fatalf("Space drilled to %q, active %q", state.Root, state.Active)
	}
	tt.Key(0, ui.KeyEscape)
	tt.Key(0, ui.KeyEscape)
	tt.Key(0, ui.KeyEscape)
	if state.Root != "" || state.Active != "" {
		t.Fatalf("Escape left root %q, active %q", state.Root, state.Active)
	}
}

func TestTooltip(t *testing.T) {
	state, data := State{}, sample
	tt, at := tester(&state, &data, Options{Label: "Usage", HideHeader: true}, 1)
	x, y := at("local")
	tt.Move(x, y)
	r, ok := tt.Find("local")
	if !ok || r.X < x || !tt.HasText("900 · 24%") {
		t.Fatalf("tooltip is not beside the pointer: %+v %q", r, tt.Texts())
	}
	tt.Move(x, 900)
	if tt.HasText("900 · 24%") {
		t.Fatal("tooltip stayed after the pointer left")
	}
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnd)
	if !tt.HasText("900 · 24%") {
		t.Fatal("keyboard choice has no tooltip")
	}
	tt.Key(0, ui.KeyEscape)
	if tt.HasText("900 · 24%") {
		t.Fatal("tooltip stayed without an active tile")
	}
}

func TestContextMenu(t *testing.T) {
	state, data := State{}, sample
	var chosen []string
	opts := Options{Label: "Usage", HideHeader: true, Menu: func(id string) func(m *ui.Menu) {
		if id == "local" {
			return nil
		}
		return func(m *ui.Menu) {
			if m.Item("Reveal " + id).Chosen() {
				chosen = append(chosen, id)
			}
		}
	}}
	tt, at := tester(&state, &data, opts, 1)
	x, y := at("work/app/target")
	tt.RightClickAt(x, y)
	if menu := tt.Menu(); len(menu) != 1 || menu[0] != "Reveal work/app/target" {
		t.Fatalf("right-click opened %q", menu)
	}
	tt.Move(x+400, y)
	if err := tt.ChooseMenuItem("Reveal work/app/target"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(chosen) != 1 || chosen[0] != "work/app/target" || state.Root != "" || state.Active != "" {
		t.Fatalf("chose %v, root %q, active %q", chosen, state.Root, state.Active)
	}
	x, y = at("local")
	tt.RightClickAt(x, y)
	if menu := tt.Menu(); menu != nil {
		t.Fatalf("a node without a menu opened %q", menu)
	}
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyHome)
	tt.Key(ui.Shift, ui.KeyF10)
	if menu := tt.Menu(); len(menu) != 1 || menu[0] != "Reveal work" {
		t.Fatalf("the menu key opened %q", menu)
	}
}

func TestBreadcrumbs(t *testing.T) {
	state, data := State{Root: "work/app"}, sample
	tt, _ := tester(&state, &data, Options{Label: "Usage"}, 1)
	if err := tt.Click("work"); err != nil {
		t.Fatal(err)
	}
	if state.Root != "work" || state.Active != "work/app" {
		t.Fatalf("crumb left root %q, active %q", state.Root, state.Active)
	}
	if !tt.HasText("1790") {
		t.Fatal("header lacks the shown level's total")
	}
}

func TestTransitions(t *testing.T) {
	state, data := State{}, sample
	tt, at := tester(&state, &data, Options{Label: "Usage", HideHeader: true, AnimationDuration: 40 * time.Millisecond}, 1)
	tt.SetPreferences(ui.Preferences{})
	tt.Frame()
	x, y := at("work")
	tt.ClickAt(x, y)
	tt.Frame()
	if !state.zooming || !state.zoomIn || state.zoomRoot != "work" {
		t.Fatal("drilling did not zoom")
	}
	// The branch grows from its tile to the whole view, carrying its old
	// contents as the new level fades in; corners follow each tile's size.
	frame := func(e float32, id string, old bool) zoomed {
		for _, z := range state.zoomFrame(e) {
			if z.tl.id == id && slices.ContainsFunc(state.old, func(tl tile) bool { return tl == z.tl }) == old {
				return z
			}
		}
		t.Fatalf("no zoomed tile %q", id)
		return zoomed{}
	}
	branch := state.old[slices.IndexFunc(state.old, func(tl tile) bool { return tl.id == "work" })].rect
	if z := frame(0, "work", true); !near(z.b, branch) || z.opacity != 1 {
		t.Fatalf("the branch starts at %+v, opacity %g", z.b, z.opacity)
	}
	if z := frame(1, "work", true); !near(z.b, ui.Rect{W: state.width, H: state.height}) || z.opacity != 0 {
		t.Fatalf("the branch ends at %+v, opacity %g", z.b, z.opacity)
	}
	if z := frame(0, "local", true); z.opacity != 1 {
		t.Fatal("a tile beside the branch fades instead of flying out")
	}
	if z := frame(0, "work/kit", false); z.opacity != 0 || z.b.X < branch.X-.5 || z.b.X+z.b.W > branch.X+branch.W+.5 {
		t.Fatalf("the new level does not open inside the branch: %+v, opacity %g", z.b, z.opacity)
	}
	if z := frame(.5, "work/kit", false); z.radius <= 0 || z.radius > min(z.b.W, z.b.H)/3 {
		t.Fatalf("corners do not follow the tile's size: radius %g for %+v", z.radius, z.b)
	}
	tt.Advance(100 * time.Millisecond)
	if state.zooming || tt.FramePending() {
		t.Fatal("zoom did not settle")
	}

	tt.Key(0, ui.KeyEscape)
	if !state.zooming || state.zoomIn {
		t.Fatal("going up did not zoom out")
	}
	if z := frame(0, "work", false); !near(z.b, ui.Rect{W: state.width, H: state.height}) || z.opacity != 0 {
		t.Fatalf("zooming out does not shrink the branch from the whole view: %+v, opacity %g", z.b, z.opacity)
	}
	tt.Advance(100 * time.Millisecond)
	data = []Node{sample[0], sample[1], {ID: "local", Label: "local", Value: 2000}}
	tt.Frame()
	if !state.morphing {
		t.Fatal("changed sizes did not morph")
	}
	tt.Advance(100 * time.Millisecond)
	if state.morphing {
		t.Fatal("morph did not settle")
	}
}

func TestEmptyAndInvalidRoot(t *testing.T) {
	state := State{Root: "local"}
	data := sample
	tt := ui.NewTester(func(c *ui.Context) {
		View(c, &state, data, Options{Height: 200, EmptyText: "Nothing"})
	}, 400, 300)
	tt.Frame()
	if state.Root != "" {
		t.Fatalf("a leaf root was kept: %q", state.Root)
	}
	data = []Node{{ID: "zero", Label: "Zero"}, {ID: "nan", Label: "NaN", Value: math.NaN()}}
	tt.Frame()
	if !tt.HasText("Nothing") {
		t.Fatal("empty text not shown")
	}
}
