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

func TestSplitButtonsClick(t *testing.T) {
	g, s, tt := featureTester("SplitButton", "Default", "en", 640)
	if err := tt.Click(g.text("secretsSave")); err != nil {
		t.Fatal(err)
	}
	if s.choice != "secretsSave" || s.count != 1 {
		t.Fatalf("choice=%q count=%d", s.choice, s.count)
	}
	g2, s2, tt2 := featureTester("PrimarySplitButton", "Default", "en", 640)
	if err := tt2.Click(g2.text("apply")); err != nil {
		t.Fatal(err)
	}
	if s2.choice != "apply" || s2.count != 1 {
		t.Fatalf("primary choice=%q count=%d", s2.choice, s2.count)
	}
}

func TestRenderPreviewsAndCloses(t *testing.T) {
	g, s, tt := featureTester("Render", "Default", "en", 640)
	if s.render.player == nil {
		t.Fatal("render player not created")
	}
	tt.Frame()
	g.Close()
	if s.render.player != nil {
		t.Fatal("render player not closed")
	}
}

func TestNetworkImageStartsOnActionAndCloses(t *testing.T) {
	g, s, tt := featureTester("Network image", "Default", "en", 640)
	if s.netImage.server != nil {
		t.Fatal("server started before user action")
	}
	if err := tt.Click(g.text("open")); err != nil {
		t.Fatal(err)
	}
	if s.netImage.server == nil || s.netImage.url == "" {
		t.Fatal("server did not start on action")
	}
	tt.Frame()
	g.Close()
	if s.netImage.server != nil {
		t.Fatal("server not closed")
	}
}

func TestListDataMoveAnimationAndChoose(t *testing.T) {
	_, s, tt := featureTester("List data", "MoveAnimation", "en", 640)
	if err := tt.Click("MoveAnimation"); err != nil {
		t.Fatal(err)
	}
	if !s.on {
		t.Fatal("MoveAnimation toggle did not flip")
	}
	if err := tt.Click("Choose(2)"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if s.index != 2 {
		t.Fatalf("selected index = %d", s.index)
	}
}

func TestShadersInstancesDraws(t *testing.T) {
	_, _, tt := featureTester("Shaders", "Instances", "en", 640)
	tt.Frame()
	if _, ok := tt.Find("DrawInstances"); !ok {
		t.Fatal("DrawInstances box not found")
	}
}
