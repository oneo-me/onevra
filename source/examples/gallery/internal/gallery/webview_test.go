package gallery

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/theme"
)

func webViewGallery() *Gallery {
	g := &Gallery{}
	for i, entry := range mygoCatalog {
		if entry.name == "WebView" {
			g.page = mygoPageStart + galleryPage(i)
		}
	}
	return g
}

func TestWebViewHeadlessShowsPlaceholder(t *testing.T) {
	g := webViewGallery()
	defer g.Close()
	if g.page == 0 {
		t.Fatal("WebView missing from the catalog")
	}
	tt := ui.NewTester(g.View, 540, 850)
	if !tt.HasText("Web views need a native window.") {
		t.Fatalf("missing placeholder: %q", tt.Texts())
	}
	// The app's cookies belong to the same demonstration: the data the
	// system web engine keeps for its pages.
	for _, key := range []string{"componentWebData", "webDataAvailable", "webDataList", "webDataSet", "webDataClear"} {
		if !tt.HasText(g.text(key)) {
			t.Errorf("missing %s in %q", key, tt.Texts())
		}
	}
}

func TestWebViewCreatedOnce(t *testing.T) {
	g := webViewGallery()
	defer g.Close()
	calls := 0
	g.createWebView = func() (*mygo.WebView, error) {
		calls++
		return nil, errors.New("no web engine")
	}
	tt := ui.NewTester(g.View, 540, 850)
	tt.Frame()
	tt.Frame()
	if calls != 1 {
		t.Fatalf("created %d web views, want 1", calls)
	}
	if !tt.HasText("Web view unavailable") || !tt.HasText("no web engine") {
		t.Fatalf("missing error: %q", tt.Texts())
	}
}

func TestWebViewHTMLFollowsThemeAndLanguage(t *testing.T) {
	pages := map[string]string{}
	for _, dark := range []bool{false, true} {
		for _, code := range []string{"en", "zh-Hans"} {
			g := &Gallery{language: newTranslator(code, "")}
			var page string
			ui.NewTester(func(c *ui.Context) {
				c.SetTheme(theme.Palette(dark))
				g.withLanguage(c, func() { page = webViewHTML(c) })
			}, 200, 200)
			if !strings.Contains(page, `lang="`+code+`"`) {
				t.Errorf("%s page lacks its language: %s", code, page)
			}
			pages[fmt.Sprint(dark, code)] = page
		}
	}
	seen := map[string]bool{}
	for _, page := range pages {
		if seen[page] {
			t.Fatal("theme or language did not change the page")
		}
		seen[page] = true
	}
	if !strings.Contains(pages[fmt.Sprint(false, "zh-Hans")], "原生界面中的网页") {
		t.Error("Chinese page is not translated")
	}
}

func TestWebViewContentRules(t *testing.T) {
	var rules []struct {
		Trigger struct {
			URLFilter string `json:"url-filter"`
		} `json:"trigger"`
		Action struct {
			Type     string `json:"type"`
			Selector string `json:"selector"`
		} `json:"action"`
	}
	if err := json.Unmarshal([]byte(webViewContentRules), &rules); err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].Trigger.URLFilter != ".*" ||
		rules[0].Action.Type != "css-display-none" || rules[0].Action.Selector != ".sponsored" {
		t.Fatalf("rules = %+v", rules)
	}
	g := &Gallery{language: newTranslator("en", "")}
	var page string
	ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.Palette(false))
		g.withLanguage(c, func() { page = webViewHTML(c) })
	}, 200, 200)
	if !strings.Contains(page, `class="sponsored"`) || !strings.Contains(page, g.text("webViewSponsoredBody")) {
		t.Fatalf("page lacks its sponsored banner: %s", page)
	}
}

// Toggling Safari Tab closes the view in the frame that builds its page
// state, to make it again in the next one.
func TestWebViewPageStateAfterClose(t *testing.T) {
	g := &Gallery{language: newTranslator("en", "en")}
	d := &webViewDemo{opened: "blocked"}
	tt := ui.NewTester(func(c *ui.Context) {
		g.withLanguage(c, func() { webViewPageState(c, d) })
	}, 540, 200)
	if tt.HasText(g.text("webViewBlocked")) {
		t.Fatalf("closed view showed page state: %q", tt.Texts())
	}
}
