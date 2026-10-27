package chat

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

func TestChip(t *testing.T) {
	clicks := 0
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Row(c).Children(func() {
			if Chip(c, "Branch", ChipOptions{Icon: lucide.GitBranch, Caret: true}).Clicked() {
				clicks++
			}
			Chip(c, "Project", ChipOptions{ReadOnly: true})
		})
	}, 500, 100)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyTab)
	if tt.Focused("Project") {
		t.Fatal("read-only chip took focus")
	}
	if err := tt.Click("Branch"); err != nil || clicks != 1 {
		t.Fatalf("click: %v clicks=%d", err, clicks)
	}
}

func TestPickerGroupsSearchAndChoice(t *testing.T) {
	var state PickerState
	selected, chosen := "b", ""
	options := []Option{{ID: "a", Label: "Alpha", Group: "One"}, {ID: "b", Label: "Beta", Group: "One", Detail: "200k"}}
	for i := range 8 {
		options = append(options, Option{ID: fmt.Sprint("x", i), Label: fmt.Sprint("Extra ", i), Group: "Two"})
	}
	tt := ui.NewTester(func(c *ui.Context) {
		r := Picker(c, &state, PickerOptions{Label: selected, SearchLabel: "Search", EmptyText: "Nothing", Options: options, Selected: selected})
		if r.Chosen != "" {
			chosen, selected = r.Chosen, r.Chosen
		}
	}, 500, 600)
	if err := tt.Click("b"); err != nil || !state.Open {
		t.Fatalf("open: %v %v", err, state.Open)
	}
	if !tt.HasText("One") || !tt.HasText("Two") || !tt.HasText("200k") {
		t.Fatal("groups or detail missing")
	}
	if !tt.Focused("Search") {
		t.Fatal("search not focused")
	}
	tt.Type("zzz")
	if !tt.HasText("Nothing") || tt.HasText("Alpha") {
		t.Fatal("empty search result missing")
	}
	tt.Command("selectAll")
	tt.Type("alp")
	if tt.HasText("Beta") {
		t.Fatal("search did not filter")
	}
	if err := tt.Click("Alpha"); err != nil || chosen != "a" || state.Open {
		t.Fatalf("choice: %v chosen=%q open=%v", err, chosen, state.Open)
	}
}

func TestContextMeterPanel(t *testing.T) {
	var state MeterState
	tt := ui.NewTester(func(c *ui.Context) {
		ContextMeter(c, &state, ContextMeterOptions{Label: "Context", Title: "Context window", Value: "61k / 272k (22%)", Fraction: .22})
	}, 500, 300)
	if tt.HasText("61k") {
		t.Fatal("panel shown before opening")
	}
	if err := tt.Click("Context"); err != nil || !state.Open || !tt.HasText("61k / 272k (22%)") {
		t.Fatalf("panel: %v open=%v", err, state.Open)
	}
}

func TestQuotaSummaryPanel(t *testing.T) {
	var state MeterState
	var refresh, open string
	quota := QuotaOptions{ID: "acct", Title: "Account", Countdown: "04:59", RefreshLabel: "Refresh", OpenLabel: "Open page", Rows: []QuotaRow{
		{Label: "5-hour limit", Detail: "2h 14m", Value: "63% left", Fraction: .63},
		{Label: "Credits", Value: "$12.40", Fraction: -1},
	}}
	row := quota.Rows[0]
	tt := ui.NewTester(func(c *ui.Context) {
		r := QuotaSummary(c, &state, QuotaSummaryOptions{Label: "Usage limits", Row: row, Quotas: []QuotaOptions{quota}})
		if r.Refresh != "" {
			refresh = r.Refresh
		}
		if r.Open != "" {
			open = r.Open
		}
	}, 500, 400)
	if !tt.HasText("2h 14m") || tt.HasText("63% left") {
		t.Fatal("summary should show only the reset time and bar")
	}
	if err := tt.Click("Usage limits"); err != nil || !state.Open {
		t.Fatalf("open: %v", err)
	}
	for _, text := range []string{"Account", "04:59", "5-hour limit", "63% left", "$12.40"} {
		if !tt.HasText(text) {
			t.Fatalf("panel missing %q", text)
		}
	}
	if err := tt.Click("Refresh"); err != nil || refresh != "acct" {
		t.Fatalf("refresh: %v %q", err, refresh)
	}
	if err := tt.Click("Open page"); err != nil || open != "acct" {
		t.Fatalf("open: %v %q", err, open)
	}
	quota.Refreshing, refresh = true, ""
	tt.Frame()
	_ = tt.Click("Refresh")
	if refresh != "" {
		t.Fatal("refresh while refreshing")
	}
	state.Open, row = false, quota.Rows[1]
	tt.Frame()
	if !tt.HasText("$12.40") {
		t.Fatal("balance summary should show its value")
	}
}

func TestEffortPickerChoosesZones(t *testing.T) {
	levels := []Option{{ID: "low", Label: "Low"}, {ID: "medium", Label: "Medium"}, {ID: "high", Label: "High"}, {ID: "max", Label: "Max"}}
	var state EffortState
	selected := "low"
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Padding(10).Children(func() {
			if r := EffortPicker(c, &state, EffortPickerOptions{Label: "Effort", Tooltip: "Effort tip", Levels: levels, Selected: selected, Width: 228}); r.Chosen != "" {
				selected = r.Chosen
			}
		})
	}, 400, 300)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	// The chip names the selected level.
	texts := len(tt.Texts())
	if !tt.HasText("Low") {
		t.Fatal("chip does not name the level")
	}
	chip, _ := tt.Find("Effort")
	if err := tt.Click("Effort"); err != nil || !state.Open || len(tt.Texts()) <= texts {
		t.Fatalf("open: %v open=%v", err, state.Open)
	}
	// The bar is the first pointer target below the chip, inside the
	// panel's start; it fills the panel's 200 DIP content width.
	pointer := func(x, y float32) bool { tt.Move(x, y); return tt.Cursor() == ui.CursorPointer }
	top := chip.Y + chip.H
	for !pointer(chip.X+100, top) && top < chip.Y+120 {
		top++
	}
	left := chip.X + 100
	for pointer(left-1, top) {
		left--
	}
	top += 6
	for i, k := range []int{2, 3, 0, 1} {
		// Alternate between the bar and the name beneath it.
		x, y := left+min(max(float32(k)*200/3, 4), 196), top
		if i%2 == 1 {
			y += 9 + 8 + 8
		}
		tt.ClickAt(x, y)
		if selected != levels[k].ID || !state.Open {
			t.Fatalf("zone %d chose %q open=%v", k, selected, state.Open)
		}
	}
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyLeft)
	if selected != "high" {
		t.Fatalf("keys chose %q", selected)
	}
}

func TestEffortStrengthWithOff(t *testing.T) {
	for _, c := range []struct {
		selected, n int
		off         bool
		want        float32
	}{{0, 5, true, 0}, {1, 5, true, 0}, {3, 5, true, 2. / 3}, {4, 5, true, 1}, {0, 4, false, 0}, {2, 4, false, 2. / 3}, {0, 1, false, 1}} {
		if got := effortStrength(c.selected, c.n, c.off); math.Abs(float64(got-c.want)) > 1e-6 {
			t.Errorf("effortStrength(%d, %d, %v) = %v, want %v", c.selected, c.n, c.off, got, c.want)
		}
	}
	// Leaving Off grows the fill from the bar's start.
	if got := effortEnd([]float32{0, .3, 1}, .5); got != .15 {
		t.Errorf("fill between Off and the first level = %v", got)
	}
}

// The trigger's icon shows all the time: it asks for frames only while
// something of it moves, not at the weaker levels, and at most every
// effortIconFrame. Each frame it paints advances its clock.
func TestEffortIconPaintsOnlyWhileMoving(t *testing.T) {
	levels := []Option{{ID: "low", Label: "Low"}, {ID: "medium", Label: "Medium"}, {ID: "high", Label: "High"}, {ID: "max", Label: "Max"}}
	frames := func(selected string) int {
		var state EffortState
		tt := ui.NewTester(func(c *ui.Context) {
			EffortPicker(c, &state, EffortPickerOptions{Label: "Effort", Icon: lucide.Brain, Levels: levels, Selected: selected})
		}, 300, 100)
		tt.Advance(2 * effortTransition)
		paints, last := 0, state.time.at
		for range 60 {
			tt.Advance(time.Second / 60)
			if state.time.at != last {
				paints, last = paints+1, state.time.at
			}
		}
		return paints
	}
	if n := frames("medium"); n != 0 {
		t.Fatalf("a still icon painted %d frames in a second", n)
	}
	if n := frames("max"); n == 0 || n > int(time.Second/effortIconFrame)+1 {
		t.Fatalf("a moving icon painted %d frames in a second", n)
	}
}

func TestEffortTimeAdvancesOncePerFrame(t *testing.T) {
	var clock effortTime
	start := time.Unix(100, 0)
	clock.advance(start, 1)
	now := start.Add(20 * time.Millisecond)
	if got := clock.advance(now, 2); math.Abs(got-.04) > 1e-9 {
		t.Fatalf("paced advance = %v", got)
	}
	// Another painter in the same frame, or a stale time, adds nothing.
	if got := clock.advance(now, 2); math.Abs(got-.04) > 1e-9 {
		t.Fatalf("repeated frame = %v", got)
	}
	// A long pause counts as one short step, so nothing leaps.
	if got := clock.advance(now.Add(time.Minute), 1); math.Abs(got-.09) > 1e-9 {
		t.Fatalf("after a pause = %v", got)
	}
}
