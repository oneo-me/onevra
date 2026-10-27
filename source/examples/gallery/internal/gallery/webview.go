package gallery

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"image/color"
	"log"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// webViewDemo owns one web view of the gallery window. It is created once on
// first entry and kept while other pages show; the view only places it.
// loaded is the theme and language its page was last rendered for. covered
// tells that a modal layer of the gallery is open over the window.
// safariTab leaves Tab to fields, as Safari does, in the next view made.
// contentRules hides the demonstration page's sponsored banner; rules is
// what it compiled on first use. themeColor and audible are what the page
// last reported, and opened what became of its last window.open.
type webViewDemo struct {
	view         *mygo.WebView
	err          error
	loaded       string
	menu         bool
	covered      bool
	safariTab    bool
	contentRules bool
	rules        *mygo.ContentRules
	rulesErr     error
	rulesPending bool
	window       func() *mygo.Window
	openURL      func(string) error
	themeColor   ui.Color
	hasTheme     bool
	audible      bool
	muted        bool
	opened       string
	offs         []func()
}

func (d *webViewDemo) start(create func() (*mygo.WebView, error)) {
	if d.view != nil || d.err != nil {
		return
	}
	d.view, d.err = create()
	if d.view != nil {
		d.watch(d.view.Page())
	}
}

// watch follows the page's theme color and sound, and decides its window
// opens: one the user asked for opens in the browser, and a pop-up the page
// opened on its own is refused. Events arrive on the main thread.
func (d *webViewDemo) watch(page *mygo.Page) {
	refresh := func() {
		if w := d.window; w != nil && w() != nil {
			w().Invalidate()
		}
	}
	d.muted = false
	tint, ok := page.ThemeColor()
	d.themeColor, d.hasTheme = ui.RGB(tint.R, tint.G, tint.B), ok
	d.offs = append(d.offs,
		page.OnThemeColorChanged(func(tint color.NRGBA, ok bool) {
			d.themeColor, d.hasTheme = ui.RGB(tint.R, tint.G, tint.B), ok
			refresh()
		}),
		page.OnAudibleChanged(func(audible bool) {
			d.audible = audible
			refresh()
		}))
	page.SetWindowOpenHandler(func(req mygo.WindowOpenRequest) *mygo.WindowOptions {
		d.opened = "blocked"
		if req.UserInitiated {
			d.opened = "opened"
			if d.openURL != nil {
				if err := d.openURL(req.URL); err != nil {
					log.Printf("Open %s: %v", req.URL, err)
				}
			}
		}
		refresh()
		return nil
	})
}

func (d *webViewDemo) close() {
	for _, off := range d.offs {
		off()
	}
	if d.view != nil {
		d.view.Destroy()
	}
	d.view, d.err, d.loaded, d.offs = nil, nil, "", nil
	d.audible, d.hasTheme, d.opened = false, false, ""
}

// newWebView names Gallery in the page's user agent, after the system web
// engine's own. A view made again keeps rules already compiled.
func (g *Gallery) newWebView() (*mygo.WebView, error) {
	d := &g.mygoState("WebView", "Default", true).webView
	page := mygo.PageOptions{
		UserAgentSuffix: "OnevraGallery/" + galleryVersion,
		SafariTabFocus:  d.safariTab,
	}
	if d.contentRules && d.rules != nil {
		page.ContentRules = []*mygo.ContentRules{d.rules}
	}
	return g.Shell.Window().NewWebView(mygo.WebViewOptions{Page: page})
}

// setContentRules compiles the rules off the UI thread on first use, then
// applies them to the view and reloads it: rules reach the loads from then
// on, and leave what the page loaded until it reloads.
func (d *webViewDemo) setContentRules() {
	if !d.contentRules || d.rules != nil {
		d.applyContentRules()
		return
	}
	if d.rulesPending {
		return
	}
	w := d.window()
	if w == nil {
		return
	}
	d.rulesPending = true
	go func() {
		rules, err := mygo.CompileContentRules(webViewContentRulesID, webViewContentRules)
		w.Update(func() {
			d.rulesPending = false
			d.rules, d.rulesErr = rules, err
			if err == nil {
				d.applyContentRules()
			}
			w.Invalidate()
		})
	}()
}

func (d *webViewDemo) applyContentRules() {
	if d.view == nil {
		return
	}
	if d.contentRules {
		d.view.Page().SetContentRules(d.rules)
	} else {
		d.view.Page().SetContentRules()
	}
	d.reload()
}

// reload shows the page again under the rules just set.
func (d *webViewDemo) reload() {
	if d.view != nil && d.loaded != "" {
		d.view.Page().LoadHTML(d.loaded, webViewBaseURL)
	}
}

// webViewBaseURL gives the demonstration page a real origin, so the page's
// document.cookie and the native Cookies module share one store.
const webViewBaseURL = "https://example.com/"

// webViewContentRules hide the demonstration page's sponsored banner, as an
// ad blocker would, and leave the rest of the page alone.
const webViewContentRules = `[
	{"trigger": {"url-filter": ".*"}, "action": {"type": "css-display-none", "selector": ".sponsored"}}]`

const webViewContentRulesID = "gallery-web-view"

var webViewPage = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="{{.Lang}}"><head><meta charset="utf-8"><meta name="theme-color" content="{{.Accent}}"><style>
body { margin: 0; padding: 24px; font: 14px/1.5 -apple-system, "Segoe UI", system-ui, sans-serif;
	background: {{.Background}}; color: {{.Text}}; }
h1 { margin: 0 0 8px; font-size: 20px; }
h2 { margin: 24px 0 8px; font-size: 16px; }
p { margin: 0 0 16px; color: {{.Muted}}; }
code { font: 12px/1.5 ui-monospace, Menlo, monospace; color: {{.Muted}}; }
.sponsored { margin: 0 0 16px; padding: 10px 12px; border-radius: 8px;
	background: {{.Accent}}; color: {{.AccentText}}; }
.row { display: flex; gap: 8px; align-items: center; }
input, button { font: inherit; padding: 6px 10px; border-radius: 6px; border: 1px solid {{.Border}};
	background: transparent; color: inherit; }
button { background: {{.Accent}}; color: {{.AccentText}}; border-color: transparent; }
</style></head><body>
<div class="sponsored"><strong>{{.Sponsored}}</strong> {{.SponsoredBody}}</div>
<h1>{{.Title}}</h1>
<p>{{.Body}}</p>
<div class="row">
	<input placeholder="{{.Placeholder}}" aria-label="{{.Placeholder}}">
	<button onclick="count.value = Number(count.value) + 1">{{.Count}}</button>
	<output id="count">0</output>
</div>
<h2>{{.UserAgent}}</h2>
<p><code id="agent"></code></p>
<h2>{{.PageEvents}}</h2>
<div class="row">
	<button onclick="tint()">{{.ThemeColor}}</button>
	<button onclick="tone()">{{.Tone}}</button>
	<a href="https://example.com/" target="_blank">{{.OpenLink}}</a>
	<button onclick="setTimeout(() => window.open('https://example.com/'), 500)">{{.PopUp}}</button>
</div>
<h2>{{.Cookies}}</h2>
<p><code id="jar"></code></p>
<div class="row">
	<button onclick="setCookie()">{{.CookieSet}}</button>
</div>
<script>
function showJar() {
	document.getElementById("jar").textContent = document.cookie || "(empty)";
}
function setCookie() {
	document.cookie = "from_page=" + Math.random().toString(36).slice(2, 6) + "; path=/; max-age=3600";
	showJar();
}
function tint() {
	document.querySelector('meta[name="theme-color"]').content = "hsl(" + Math.floor(Math.random() * 360) + " 70% 50%)";
}
function tone() {
	const audio = new AudioContext(), osc = audio.createOscillator(), gain = audio.createGain();
	gain.gain.value = 0.1;
	osc.connect(gain).connect(audio.destination);
	osc.start();
	osc.stop(audio.currentTime + 2);
	osc.onended = () => audio.close();
}
showJar();
document.getElementById("agent").textContent = navigator.userAgent;
</script>
</body></html>`))

func cssColor(c ui.Color) template.CSS {
	return template.CSS(fmt.Sprintf("rgba(%d, %d, %d, %.3f)", c.R, c.G, c.B, float32(c.A)/255))
}

// webViewHTML renders the demonstration page in the gallery's current theme
// and language, so the page reads as part of the native content around it.
func webViewHTML(c *ui.Context) string {
	t := c.Theme()
	lang := "en"
	if l := *languageSlot(c); l != nil {
		lang = l.code
	}
	var b bytes.Buffer
	_ = webViewPage.Execute(&b, map[string]any{
		"Lang":       lang,
		"Background": cssColor(t.Background), "Text": cssColor(t.Text), "Muted": cssColor(t.TextMuted),
		"Border": cssColor(t.Border), "Accent": cssColor(t.Accent), "AccentText": cssColor(t.AccentText),
		"Title":         tr(c, "aWebPageInsideNativeUi"),
		"Body":          tr(c, "theSystemWebEngineRendersThisPageNativePopoversDialogsAnd"),
		"Placeholder":   tr(c, "typeIntoThePage"),
		"Count":         tr(c, "count"),
		"UserAgent":     tr(c, "webViewUserAgent"),
		"Cookies":       tr(c, "webViewCookies"),
		"CookieSet":     tr(c, "webViewCookieSet"),
		"PageEvents":    tr(c, "webViewPageEvents"),
		"ThemeColor":    tr(c, "webViewChangeThemeColor"),
		"Tone":          tr(c, "webViewPlayTone"),
		"OpenLink":      tr(c, "webViewOpenLink"),
		"PopUp":         tr(c, "webViewPopUp"),
		"Sponsored":     tr(c, "webViewSponsored"),
		"SponsoredBody": tr(c, "webViewSponsoredBody"),
	})
	return b.String()
}

func exampleWebView(c *ui.Context, s *demoState) {
	t := c.Theme()
	d := &s.webView
	switch {
	case d.err != nil:
		components.Text(c, tr(c, "webViewUnavailable")).Bold()
		components.Text(c, d.err.Error()).TextColor(t.Danger).Wrap()
	case d.view == nil:
		components.Text(c, tr(c, "webViewsNeedANativeWindow")).TextColor(t.TextMuted).Wrap()
	default:
		d.window = s.system.window
		if html := webViewHTML(c); d.loaded != html {
			d.loaded = html
			// Fill with the new theme until the reloaded page paints, not white.
			if err := d.view.SetBackgroundColor(string(cssColor(t.Background))); err != nil {
				log.Printf("Set web view background: %v", err)
			}
			d.view.Page().LoadHTML(html, webViewBaseURL)
		}
		ui.Row(c).Gap(t.Space(2)).AlignItems(ui.Center).Children(func() {
			if components.Button(c, tr(c, "reloadPage"), components.ButtonSecondary).Clicked() {
				d.view.Page().LoadHTML(d.loaded, webViewBaseURL)
			}
			anchor := components.Button(c, tr(c, "overThePage"), components.ButtonOutline)
			if anchor.Clicked() {
				d.menu = !d.menu
			}
			components.Popover(c, anchor, &d.menu, func(panel ui.Element) {
				components.Text(c, tr(c, "nativeUiPaintedAfterTheWebViewShowsOverItsPage")).Wrap().MaxWidth(t.Space(60))
			})
			// Tab focus is the page's from its creation: a change makes the
			// web view again, which the page then loads into.
			if runtime.GOOS == "darwin" && components.Checkbox(c, &d.safariTab, tr(c, "webViewSafariTab")).Changed() {
				d.close()
				c.Invalidate()
			}
		})
		if components.Checkbox(c, &d.contentRules, tr(c, "webViewContentRules")).Changed() {
			d.setContentRules()
		}
		if d.rulesErr != nil {
			components.Text(c, tr(c, "webViewContentRulesFailed")+": "+d.rulesErr.Error()).TextColor(t.Danger).Wrap()
		}
		webViewPageState(c, d)
		var frame ui.Element
		if d.covered {
			// Glass materials cannot see a web page under them, so a modal layer
			// over the window would show a hole. Hide the web view, keeping its
			// page, behind a box of the page's background meanwhile.
			frame = ui.Box(c).Background(t.Background)
		} else {
			frame = ui.WebView(c, d.view)
		}
		frame.FillWidth().Height(320).Radius(t.Radius*1.5).Border(1, t.Border).Children(func() {
			components.Badge(c, "HTML").Absolute().Top(t.Space(3)).Right(t.Space(3))
		})
	}
	webDataExample(c, s, t)
}

// webViewPageState shows what the page tells the app: its theme color, its
// sound, which the app can mute as a browser mutes a tab, the process that
// runs it and what became of its last window.open. Nothing shows in the
// frame that closed the view to make it again.
func webViewPageState(c *ui.Context, d *webViewDemo) {
	if d.view == nil {
		return
	}
	t, m := c.Theme(), msg(c)
	ui.Row(c).Gap(t.Space(3)).AlignItems(ui.Center).Wrap().Children(func() {
		components.Text(c, tr(c, "webViewThemeColor")).TextColor(t.TextMuted)
		if d.hasTheme {
			ui.Box(c).Size(16, 16).Radius(4).Background(d.themeColor).Border(1, t.Border).Label(tr(c, "webViewThemeColor"))
		} else {
			components.Text(c, yesNo(m, false))
		}
		components.Text(c, tr(c, "webViewAudible")).TextColor(t.TextMuted)
		components.Text(c, yesNo(m, d.audible))
		if components.Checkbox(c, &d.muted, tr(c, "webViewMute")).Changed() {
			d.view.Page().SetAudioMuted(d.muted)
		}
	})
	if pid := d.view.Page().ProcessID(); pid > 0 {
		systemRow(c, tr(c, "webViewProcess"), fmt.Sprint(pid))
	}
	switch d.opened {
	case "opened":
		components.Text(c, tr(c, "webViewOpened")).TextColor(t.TextMuted).Wrap()
	case "blocked":
		components.Text(c, tr(c, "webViewBlocked")).TextColor(t.TextMuted).Wrap()
	}
}

// webDataExample reads and changes the cookies of the app's pages: the data
// the system web engine keeps for every window's page. Setting and clearing
// from Go refresh the page's own view of document.cookie afterwards.
func webDataExample(c *ui.Context, s *demoState, t *ui.Theme) {
	d := &s.system
	m := msg(c)
	components.Text(c, m.ComponentWebData()).Bold().TextColor(t.TextMuted)
	available := false
	if d.native {
		available = mygo.WebViewAvailable()
	}
	systemRow(c, m.WebDataAvailable(), yesNo(m, available))
	ui.Row(c).Gap(8).Children(func() {
		if components.Button(c, m.WebDataList(), components.ButtonPrimary).Clicked() {
			d.background(func() func() {
				cookies, err := mygo.Cookies.List(mygo.CookieFilter{})
				return func() { d.cookies, d.cookieStatus = cookies, cookieResult(m, err) }
			})
		}
		if components.Button(c, m.WebDataSet(), components.ButtonOutline).Clicked() {
			d.background(func() func() {
				err := mygo.Cookies.Set(mygo.Cookie{Name: "onevra_gallery", Value: "demo", Domain: "example.com", Expires: time.Now().Add(time.Hour)})
				cookies, listErr := mygo.Cookies.List(mygo.CookieFilter{})
				return func() {
					d.cookies, d.cookieStatus = cookies, cookieResult(m, errors.Join(err, listErr))
					showPageJar(s)
				}
			})
		}
		if components.Button(c, m.WebDataClear(), components.ButtonOutline).Clicked() {
			d.background(func() func() {
				err := mygo.Cookies.Clear()
				return func() {
					d.cookies, d.cookieStatus = nil, cookieResult(m, err)
					showPageJar(s)
				}
			})
		}
	})
	if d.cookieStatus != "" {
		components.Text(c, d.cookieStatus).Wrap().TextColor(t.TextMuted)
	}
	cookies := slices.Clone(d.cookies)
	slices.SortFunc(cookies, func(a, b mygo.Cookie) int { return strings.Compare(a.Domain+a.Name, b.Domain+b.Name) })
	for i, cookie := range cookies {
		if i == 30 {
			break
		}
		systemRow(c, cookie.Domain, cookie.Name+"="+cookie.Value)
	}
}

// showPageJar asks the web view's page to show document.cookie again, so a
// cookie set or cleared by Go shows in the page's own view of the store.
func showPageJar(s *demoState) {
	if v := s.webView.view; v != nil {
		v.Page().Eval("showJar()")
	}
}
