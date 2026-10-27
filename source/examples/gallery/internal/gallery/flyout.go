package gallery

import (
	"bytes"
	"html/template"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/theme"
)

// flyoutDemo retains application state separately from the transient window.
// page is the fitted flyout's page, rendered in the theme and language it
// opens in.
type flyoutDemo struct {
	gallery *Gallery
	window  *mygo.Window
	note    string
	page    string
}

func (d *flyoutDemo) close() {
	if d.window != nil {
		d.window.Destroy()
	}
	d.window = nil
}

// flyoutStyle is how the page's flyout looks: the editor draws its own
// surface, the system popover is the platform's, the material flyout shows a
// blurred material cut to rounded corners where the platform has one, and the
// fitted flyout is a web page that sizes it.
type flyoutStyle int

const (
	flyoutEditor flyoutStyle = iota
	flyoutPopover
	flyoutMaterial
	flyoutFitted
)

func (d *flyoutDemo) show(bounds ui.Rect, style flyoutStyle) {
	if d.gallery == nil || d.gallery.Shell == nil {
		return
	}
	d.close()
	parent := d.gallery.Shell.Window()
	o := mygo.FlyoutOptions{
		Parent:    parent,
		Anchor:    mygo.Rectangle{X: int(bounds.X), Y: int(bounds.Y), Width: int(bounds.W), Height: int(bounds.H)},
		Placement: mygo.PlacementBottomStart, Gap: 6,
		Width: 300, Height: 180, Focusable: true, Shadow: true, Popover: style == flyoutPopover,
		Content: ui.View(d.view),
	}
	if style == flyoutMaterial {
		// Windows 11 rounds flyouts itself; CornerRadius rounds them on macOS.
		o.Vibrancy, o.CornerRadius = mygo.VibrancyPopover, 10
	}
	if style == flyoutFitted {
		// The page sizes the flyout, and resizes it as items are added,
		// up to the bounds beyond which it scrolls.
		o.Content, o.FitContent = nil, true
		o.MinWidth, o.MaxWidth, o.MaxHeight = 200, 360, 320
	}
	w := mygo.NewFlyout(o)
	if style == flyoutFitted {
		w.Page().LoadHTML(d.page, "")
	}
	d.window = w
	w.OnClosed(func() {
		if d.window == w {
			d.window = nil
		}
		parent.Invalidate()
	})
}

func (d *flyoutDemo) view(c *ui.Context) {
	theme.Apply(c)
	c.Root().Background(ui.Transparent)
	d.gallery.withLanguage(c, func() {
		t := c.Theme()
		panel := ui.Column(c).Fill().Padding(16).Gap(12)
		if !c.Vibrancy() {
			panel.Background(t.Surface)
			switch runtime.GOOS {
			case "darwin":
				panel.Radius(10)
			case "windows":
			default:
				panel.Radius(10).Border(1, t.Border)
			}
		}
		panel.Children(func() {
			components.TextBox(c, &d.note, components.TextBoxOptions{}).Label(tr(c, "notes")).FillWidth().AutoFocus()
			components.Button(c, tr(c, "close"), components.ButtonOutline).OnClick(func() {
				if d.window != nil {
					d.window.Close()
				}
			})
		})
	})
}

var flyoutPage = template.Must(template.New("flyout").Parse(`<!doctype html>
<html lang="{{.Lang}}"><head><meta charset="utf-8"><style>
body { margin: 0; padding: 14px; font: 13px/1.5 -apple-system, "Segoe UI", system-ui, sans-serif;
	background: {{.Background}}; color: {{.Text}}; }
p { margin: 0 0 8px; }
li { margin: 2px 0; }
button { font: inherit; padding: 4px 10px; border: 0; border-radius: 6px; background: {{.Accent}}; color: {{.AccentText}}; }
</style></head><body>
<p>{{.Body}}</p>
<ul id="items"></ul>
<button onclick="add()">{{.Add}}</button>
<script>
let n = 0;
function add() {
	const li = document.createElement("li");
	li.textContent = "{{.Item}} " + ++n;
	document.getElementById("items").append(li);
}
</script>
</body></html>`))

// flyoutHTML renders the fitted flyout's page in the gallery's theme and
// language.
func flyoutHTML(c *ui.Context) string {
	t := c.Theme()
	lang := "en"
	if l := *languageSlot(c); l != nil {
		lang = l.code
	}
	var b bytes.Buffer
	_ = flyoutPage.Execute(&b, map[string]any{
		"Lang": lang, "Background": cssColor(t.Surface), "Text": cssColor(t.Text),
		"Accent": cssColor(t.Accent), "AccentText": cssColor(t.AccentText),
		"Body": tr(c, "flyoutFittedBody"), "Add": tr(c, "flyoutFittedAdd"), "Item": tr(c, "flyoutFittedItem"),
	})
	return b.String()
}

func exampleFlyout(c *ui.Context, s *demoState) {
	d := &s.flyout
	components.Text(c, tr(c, "flyoutDescription")).Wrap()
	for i, label := range []string{"flyoutEditor", "flyoutPopover", "flyoutMaterial", "flyoutFitted"} {
		b := components.Button(c, tr(c, label), components.ButtonOutline).Disabled(d.gallery == nil || d.gallery.Shell == nil)
		if b.Clicked() {
			d.page = flyoutHTML(c)
			d.show(b.Bounds(), flyoutStyle(i))
		}
	}
	components.Text(c, d.note).Wrap()
}
