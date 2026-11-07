package components

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/theme"
)

// cardStackTester builds a stack of three cards. Reduced Motion settles
// cards at once, keeping the tests to a few frames.
func cardStackTester(state *CardStackState, events *[]CardStackEvent) *ui.Tester {
	items := []CardStackItem{{ID: "a", Label: "Lisbon"}, {ID: "b", Label: "Porto"}, {ID: "c", Label: "Faro"}}
	for i := range items {
		label := items[i].Label
		items[i].Content = func(c *ui.Context, _ float32) { Text(c, label) }
	}
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.Palette(false))
		ui.Column(c).Fill().Padding(40).Children(func() {
			if e := CardStack(c, state, CardStackOptions{Items: items, Label: "Trips"}); e.Action != CardStackNone {
				*events = append(*events, e)
			}
		})
	}, 400, 500)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	return tt
}

func settleCardStack(tt *ui.Tester) {
	tt.Advance(time.Second)
	tt.Frame()
}

func TestCardStackButtonsDecideUndoAndReset(t *testing.T) {
	var state CardStackState
	var events []CardStackEvent
	tt := cardStackTester(&state, &events)
	tt.Frame()
	if _, ok := tt.Find("Lisbon, 1 of 3"); !ok {
		t.Fatal("top card not named with its position")
	}
	if err := tt.Click("Keep"); err != nil {
		t.Fatal(err)
	}
	settleCardStack(tt)
	if want := []CardStackDecision{{"a", CardRight}}; !slices.Equal(state.Decisions, want) {
		t.Fatalf("decisions %v, want %v", state.Decisions, want)
	}
	if len(events) != 1 || events[0] != (CardStackEvent{CardStackDecided, "a", CardRight}) {
		t.Fatalf("events %v", events)
	}
	if _, ok := tt.Find("Porto, 2 of 3"); !ok {
		t.Fatal("next card not on top")
	}
	if tt.HasText("Lisbon") {
		t.Fatal("thrown card still built after its flight")
	}
	if a := tt.Announcements(); len(a) == 0 || a[len(a)-1] != "Lisbon: Keep. 2 left." {
		t.Fatalf("announcements %q", a)
	}

	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	settleCardStack(tt)
	if len(state.Decisions) != 0 || events[len(events)-1] != (CardStackEvent{CardStackUndone, "a", CardRight}) {
		t.Fatalf("undo left %v, events %v", state.Decisions, events)
	}
	if _, ok := tt.Find("Lisbon, 1 of 3"); !ok {
		t.Fatal("undone card not back on top")
	}

	for _, action := range []string{"Pass", "Keep", "Pass"} {
		if err := tt.Click(action); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
	}
	settleCardStack(tt)
	if len(state.Decisions) != 3 || !tt.HasText("All cards reviewed") {
		t.Fatalf("decisions %v; empty state shown %v", state.Decisions, tt.HasText("All cards reviewed"))
	}
	if err := tt.Click("Start over"); err != nil {
		t.Fatal(err)
	}
	settleCardStack(tt)
	if len(state.Decisions) != 0 || events[len(events)-1].Action != CardStackReset {
		t.Fatalf("reset left %v, events %v", state.Decisions, events)
	}
	if _, ok := tt.Find("Lisbon, 1 of 3"); !ok || tt.HasText("All cards reviewed") {
		t.Fatal("reset did not deal the cards back")
	}
}

func TestCardStackDragThrowsOrReturns(t *testing.T) {
	var state CardStackState
	var events []CardStackEvent
	tt := cardStackTester(&state, &events)
	tt.Frame()
	r, ok := tt.Find("Lisbon, 1 of 3")
	if !ok {
		t.Fatal("no top card")
	}
	x, y := r.X+r.W/2, r.Y+r.H/4
	drag := func(dx float32) {
		tt.Press(x, y)
		for i := 1; i <= 3; i++ {
			tt.Advance(16 * time.Millisecond)
			tt.Move(x+dx*float32(i)/3, y)
		}
		tt.Advance(16 * time.Millisecond)
		tt.Release(x+dx, y)
		settleCardStack(tt)
	}

	// A short, slow drag springs back home.
	tt.Press(x, y)
	tt.Advance(16 * time.Millisecond)
	tt.Move(x-30, y)
	tt.Advance(200 * time.Millisecond)
	tt.Move(x-30, y)
	tt.Advance(200 * time.Millisecond)
	tt.Release(x-30, y)
	settleCardStack(tt)
	if len(state.Decisions) != 0 {
		t.Fatalf("short drag decided %v", state.Decisions)
	}
	if back, _ := tt.Find("Lisbon, 1 of 3"); back.X != r.X || back.Y != r.Y {
		t.Fatalf("card not home: %+v, was %+v", back, r)
	}

	drag(-200)
	if want := []CardStackDecision{{"a", CardLeft}}; !slices.Equal(state.Decisions, want) {
		t.Fatalf("far drag left decided %v", state.Decisions)
	}
	drag(200)
	if want := []CardStackDecision{{"a", CardLeft}, {"b", CardRight}}; !slices.Equal(state.Decisions, want) {
		t.Fatalf("far drag right decided %v", state.Decisions)
	}
}

func TestCardStackKeys(t *testing.T) {
	var state CardStackState
	var events []CardStackEvent
	tt := cardStackTester(&state, &events)
	tt.Frame()
	if err := tt.Click("Trips"); err != nil {
		t.Fatal(err)
	}
	settleCardStack(tt)
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	tt.Key(0, ui.KeyLeft)
	settleCardStack(tt)
	if want := []CardStackDecision{{"a", CardRight}, {"b", CardLeft}}; !slices.Equal(state.Decisions, want) {
		t.Fatalf("arrow keys decided %v", state.Decisions)
	}
	tt.Key(0, ui.KeyBackspace)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyZ)
	settleCardStack(tt)
	if len(state.Decisions) != 0 {
		t.Fatalf("undo keys left %v", state.Decisions)
	}
}

func TestCardStackCanceledDragReturns(t *testing.T) {
	var state CardStackState
	var events []CardStackEvent
	tt := cardStackTester(&state, &events)
	tt.Frame()
	r, ok := tt.Find("Lisbon, 1 of 3")
	if !ok {
		t.Fatal("no top card")
	}
	x, y := r.X+r.W/2, r.Y+r.H/4
	tt.Press(x, y)
	tt.Advance(16 * time.Millisecond)
	tt.Move(x+r.W, y)
	tt.Advance(16 * time.Millisecond)
	tt.CancelPress()
	settleCardStack(tt)
	if len(state.Decisions) != 0 || len(events) != 0 {
		t.Fatalf("canceled drag decided %v, events %v", state.Decisions, events)
	}
}

func TestCardStackFollowsApplicationDecisions(t *testing.T) {
	state := CardStackState{Decisions: []CardStackDecision{{"a", CardLeft}}}
	var events []CardStackEvent
	tt := cardStackTester(&state, &events)
	tt.Frame()
	if _, ok := tt.Find("Porto, 2 of 3"); !ok || tt.HasText("Lisbon") {
		t.Fatal("initial decisions not applied")
	}
	state.Decisions = nil
	tt.Frame()
	settleCardStack(tt)
	if _, ok := tt.Find("Lisbon, 1 of 3"); !ok || len(events) != 0 {
		t.Fatalf("application undo not followed; events %v", events)
	}
	// A thrown card stays in flight until its fade ends.
	tt.SetPreferences(ui.Preferences{})
	state.Decisions = []CardStackDecision{{"a", CardRight}}
	tt.Frame()
	tt.Advance(100 * time.Millisecond)
	if !tt.HasText("Lisbon") {
		t.Fatal("thrown card gone before its flight")
	}
	settleCardStack(tt)
	if tt.HasText("Lisbon") {
		t.Fatal("thrown card still built after its flight")
	}
}
