package gallery

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// Every new page builds each variant in both widths, languages
// and appearances without opening native resources.
func TestNewPagesBuild(t *testing.T) {
	for route := range newPages {
		name := route[len("mygo/"):]
		for _, variant := range mygoVariants(name) {
			for _, lang := range []string{"en", "zh-Hans"} {
				for _, width := range []int{280, 640} {
					_, _, tt := featureTester(name, variant, lang, width)
					tt.SetDark(width == 280)
					tt.Frame()
					if len(tt.Texts()) == 0 && name != "Clip & mask" {
						t.Fatalf("%s/%s shows nothing", name, variant)
					}
				}
			}
		}
	}
}

func TestScrollPositionsBringTheRowIntoView(t *testing.T) {
	g, _, tt := featureTester("Scroll positions", "Into view", "en", 640)
	if err := tt.Click("ScrollEnd"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(time.Second)
	if !tt.HasText(g.messages().ScrollbarsCard(8)) {
		t.Fatalf("target row not shown: %q", tt.Texts())
	}
}

func TestHitTestTakesTheBandOnly(t *testing.T) {
	_, s, tt := featureTester("Clip & mask", "Hit testing", "en", 640)
	r, ok := tt.Find("HitTest")
	if !ok {
		t.Fatal("no ring")
	}
	tt.ClickAt(r.X+r.W/2, r.Y+r.H/2)
	if s.count != 0 {
		t.Fatal("the hole took the press")
	}
	tt.ClickAt(r.X+r.W/2, r.Y+r.H/8)
	if s.count != 1 {
		t.Fatal("the band did not take the press")
	}
}

func TestInputEventsObservesKeysByPosition(t *testing.T) {
	_, s, tt := featureTester("Input events", "Keys", "en", 640)
	if err := tt.Click("ObserveInput"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyA)
	if s.input.observed != 1 {
		t.Fatalf("observed %d keys", s.input.observed)
	}
}

func TestWidgetBasesOpenFromTheApplication(t *testing.T) {
	g, s, tt := featureTester("Widget bases", "SelectBaseOpen", "en", 640)
	if err := tt.Click("SetOpen"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click(g.text("large")); err != nil {
		t.Fatal(err)
	}
	if s.choice != "Large" {
		t.Fatalf("choice = %q", s.choice)
	}
}

func TestRouterMatchFliesBetweenPages(t *testing.T) {
	g, s, tt := featureTester("Router match", "Default", "en", 640)
	if err := tt.Click(g.text("morphShort")); err != nil {
		t.Fatal(err)
	}
	if s.router.Path() != "/page" {
		t.Fatalf("path = %q", s.router.Path())
	}
}

func TestListDataFollowsItsEnd(t *testing.T) {
	_, s, tt := featureTester("List data", "HoldPlace", "en", 640)
	if !s.list.FollowEnd {
		t.Fatal("the list does not follow its end")
	}
	tt.Frame()
}
