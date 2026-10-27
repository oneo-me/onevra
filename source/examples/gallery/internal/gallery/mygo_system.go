package gallery

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/examples/gallery/internal/gallery/locales"
)

// systemDemo holds the system-service examples. Native is false in headless
// views, which show the controls without calling the running app; results of
// background work return to the UI thread through the gallery window.
type systemDemo struct {
	native bool
	window func() *mygo.Window

	network    mygo.NetworkStatus
	networkOff func()

	secretKey, secretValue, secretStatus string

	recentStatus string

	watchPath   string
	watchHints  bool
	watchEvents []mygo.FileEvent
	watchErr    string
	watchStop   context.CancelFunc

	shortcut, shortcutDraft, shortcutErr string
	shortcutFront                        bool
	shortcutPresses                      int
	shortcutLast                         time.Time

	fontQuery  string
	monospace  bool
	fontFamily string

	systemTab int

	cookies      []mygo.Cookie
	cookieStatus string

	dialogDetached bool
	dialogType     int
	dialogRows     [][2]string
	dialogNote     string
}

// update applies fn on the UI thread and redraws the gallery.
func (d *systemDemo) update(fn func()) {
	if w := d.window(); w != nil {
		w.Update(fn)
	}
}

// background runs work off the UI thread, as dialogs and system stores block.
func (d *systemDemo) background(work func() func()) {
	if !d.native {
		return
	}
	go func() {
		apply := work()
		d.update(apply)
	}()
}

func (d *systemDemo) close() {
	if d.networkOff != nil {
		d.networkOff()
		d.networkOff = nil
	}
	d.stopWatching()
	d.unregisterShortcut()
}

func systemRow(c *ui.Context, label, value string) {
	if value == "" {
		value = "—"
	}
	ui.Row(c).FillWidth().Gap(12).AlignItems(ui.Center).Children(func() {
		components.Text(c, label).TextColor(c.Theme().TextMuted).Width(160)
		components.Text(c, value).Grow(1).Wrap()
	})
}

func yesNo(m locales.Messages, v bool) string {
	if v {
		return m.SystemYes()
	}
	return m.SystemNo()
}

func exampleNetwork(c *ui.Context, s *demoState) {
	d, m := &s.system, msg(c)
	if d.native && d.networkOff == nil {
		d.network = mygo.Network.Status()
		d.networkOff = mygo.Network.OnChanged(func(status mygo.NetworkStatus) {
			d.update(func() { d.network = status })
		})
	}
	status := m.NetworkOffline()
	if d.network.Online {
		status = m.NetworkOnline()
	}
	components.Badge(c, status)
	systemRow(c, m.NetworkMetered(), yesNo(m, d.network.Metered))
}

func exampleSecrets(c *ui.Context, s *demoState) {
	d, m := &s.system, msg(c)
	if d.secretKey == "" {
		d.secretKey = "gallery-demo"
	}
	components.TextBox(c, &d.secretKey, components.TextBoxOptions{}).Label(m.SecretsKey()).FillWidth()
	components.TextBox(c, &d.secretValue, components.TextBoxOptions{}).Label(m.SecretsValue()).FillWidth()
	key, value := d.secretKey, d.secretValue
	ui.Row(c).Gap(8).Children(func() {
		if components.Button(c, m.SecretsSave(), components.ButtonPrimary).Disabled(key == "").Clicked() {
			d.background(func() func() {
				err := mygo.Secrets.Set(key, []byte(value))
				return func() { d.secretStatus = secretResult(m, err, m.SecretsSaved()) }
			})
		}
		if components.Button(c, m.SecretsRead(), components.ButtonOutline).Disabled(key == "").Clicked() {
			d.background(func() func() {
				got, err := mygo.Secrets.Get(key)
				return func() {
					if err == nil {
						d.secretValue = string(got)
					}
					d.secretStatus = secretResult(m, err, m.SecretsLoaded())
				}
			})
		}
		if components.Button(c, m.Delete(), components.ButtonOutline).Disabled(key == "").Clicked() {
			d.background(func() func() {
				err := mygo.Secrets.Delete(key)
				return func() { d.secretStatus = secretResult(m, err, m.SecretsDeleted()) }
			})
		}
	})
	if d.secretStatus != "" {
		components.Text(c, d.secretStatus).Wrap().TextColor(c.Theme().TextMuted)
	}
}

func secretResult(m locales.Messages, err error, ok string) string {
	switch {
	case err == nil:
		return ok
	case errors.Is(err, mygo.ErrSecretNotFound):
		return m.SecretsNotFound()
	default:
		return m.SystemError(err.Error())
	}
}

func exampleRecentDocuments(c *ui.Context, s *demoState) {
	d, m := &s.system, msg(c)
	ui.Row(c).Gap(8).Children(func() {
		if components.Button(c, m.RecentChoose(), components.ButtonPrimary).Clicked() {
			d.background(func() func() {
				paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Title: m.RecentChoose()})
				if err == nil && len(paths) > 0 {
					err = mygo.App.AddRecentDocument(paths[0])
				}
				return func() {
					if err != nil {
						d.recentStatus = m.SystemError(err.Error())
					}
				}
			})
		}
		if components.Button(c, m.Clear(), components.ButtonOutline).Clicked() && d.native {
			if err := mygo.App.ClearRecentDocuments(); err != nil {
				d.recentStatus = m.SystemError(err.Error())
			}
		}
		if runtime.GOOS == "windows" && components.Button(c, m.RecentJumpList(), components.ButtonOutline).Clicked() && d.native {
			err := mygo.App.JumpList.SetTasks([]mygo.JumpListTask{{Title: m.RecentJumpListTask(), Description: m.RecentJumpListHint()}})
			if err != nil {
				d.recentStatus = m.SystemError(err.Error())
			}
		}
	})
	var recent []string
	if d.native {
		recent = mygo.App.RecentDocuments()
	}
	if len(recent) == 0 {
		components.Text(c, m.RecentEmpty()).TextColor(c.Theme().TextMuted)
	}
	for i, path := range recent {
		ui.Column(c).Key(i).FillWidth().Children(func() {
			components.Text(c, filepath.Base(path)).Bold()
			components.Text(c, path).TextColor(c.Theme().TextMuted).FontSize(12).Wrap()
		})
	}
	if d.recentStatus != "" {
		components.Text(c, d.recentStatus).Wrap().TextColor(c.Theme().Danger)
	}
}

func (d *systemDemo) stopWatching() {
	if d.watchStop != nil {
		d.watchStop()
		d.watchStop = nil
	}
}

// watch follows a directory until stopped, keeping the latest events.
func (d *systemDemo) watch(path string) {
	d.stopWatching()
	ctx, cancel := context.WithCancel(context.Background())
	d.watchStop, d.watchPath, d.watchEvents, d.watchErr = cancel, path, nil, ""
	hints := d.watchHints
	go func() {
		w, err := mygo.WatchFiles(ctx, path, mygo.FileWatchOptions{Recursive: true, HintsOnly: hints,
			Ignore: func(p string, dir bool) bool { return dir && (p == ".git" || p == "node_modules") }})
		if err != nil {
			d.update(func() { d.watchErr = err.Error() })
			return
		}
		defer w.Close()
		for {
			events, err := w.Next(ctx)
			if err != nil {
				if ctx.Err() == nil {
					d.update(func() { d.watchErr = err.Error() })
				}
				return
			}
			d.update(func() {
				d.watchEvents = append(d.watchEvents, events...)
				if extra := len(d.watchEvents) - 50; extra > 0 {
					d.watchEvents = d.watchEvents[extra:]
				}
			})
		}
	}()
}

func exampleFileWatching(c *ui.Context, s *demoState) {
	d, m := &s.system, msg(c)
	ui.Row(c).Gap(8).Children(func() {
		if components.Button(c, m.WatchChoose(), components.ButtonPrimary).Clicked() {
			d.background(func() func() {
				paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Title: m.WatchChoose(), Directory: true})
				return func() {
					if err != nil {
						d.watchErr = err.Error()
					} else if len(paths) > 0 {
						d.watch(paths[0])
					}
				}
			})
		}
		if components.Button(c, m.WatchStop(), components.ButtonOutline).Disabled(d.watchStop == nil).Clicked() {
			d.stopWatching()
		}
	})
	// The choice applies to the next watch.
	components.Checkbox(c, &d.watchHints, m.WatchHintsOnly()).Disabled(d.watchStop != nil)
	if d.watchPath != "" {
		systemRow(c, m.WatchFolder(), d.watchPath)
	}
	if d.watchErr != "" {
		components.Text(c, m.SystemError(d.watchErr)).Wrap().TextColor(c.Theme().Danger)
	}
	if len(d.watchEvents) == 0 {
		components.Text(c, m.WatchEmpty()).TextColor(c.Theme().TextMuted)
	}
	for i := len(d.watchEvents) - 1; i >= 0; i-- {
		e := d.watchEvents[i]
		ui.Row(c).Key(i).Gap(8).Children(func() {
			components.Badge(c, string(e.Op))
			components.Text(c, e.Path).Grow(1).Wrap()
		})
	}
}

// exampleSystemInformation gathers what an application reads from the system
// it runs on, one tab each: the desktop's appearance and settings, its fonts,
// the user's languages and regional formats, the network state and the
// credential store.
func exampleSystemInformation(c *ui.Context, s *demoState) {
	d, m, t := &s.system, msg(c), c.Theme()
	components.Tabs(c, &d.systemTab, components.TabsSegmented,
		m.AppearanceTab(), m.ComponentFonts(), m.LanguagesTab(), m.ComponentNetwork(), m.ComponentSecrets())
	ui.Column(c).Key(d.systemTab).FillWidth().Gap(t.Space(3)).Children(func() {
		switch d.systemTab {
		case 1:
			exampleFonts(c, s)
		case 2:
			exampleLanguages(c, s)
		case 3:
			exampleNetwork(c, s)
		case 4:
			exampleSecrets(c, s)
		default:
			exampleSystemAppearance(c, s)
		}
	})
}

// exampleFonts chooses a family from a drop-down and previews the sample text
// in it. The list is long, so it is searched and scrolled inside a bounded
// popup; the monospaced filter sits beside the trigger, where it applies on
// reopening.
func exampleFonts(c *ui.Context, s *demoState) {
	d, m, t := &s.system, msg(c), c.Theme()
	families := ui.FontFamilies()
	if d.fontFamily == "" {
		switch {
		case t.Font != "":
			d.fontFamily = t.Font
		case len(families) > 0:
			d.fontFamily = families[0].Name
		}
	}
	ui.Row(c).FillWidth().Gap(t.Space(2)).AlignItems(ui.Center).Children(func() {
		sel := components.Select(c, &d.fontFamily)
		sel.Trigger.Label(m.FontsFamily()).Grow(1).Children(func() {
			components.Text(c, d.fontFamily).SingleLine()
		})
		sel.Popup(func(panel ui.Element) {
			panel.MinWidth(t.Rem(16))
			sel.SearchField(&d.fontQuery).Label(m.FontsSearch()).Placeholder(m.FontsSearch())
			// Bound the list and scroll it: a font list is long, and the popup
			// must not grow past the window.
			list := components.ScrollView(c, components.ScrollViewOptions{}).FillWidth().MaxHeight(t.Space(70))
			list.Children(func() {
				shown := 0
				for _, family := range families {
					if d.monospace && !family.Monospaced {
						continue
					}
					if d.fontQuery != "" && !strings.Contains(strings.ToLower(family.Name), strings.ToLower(d.fontQuery)) {
						continue
					}
					if shown++; shown > 40 {
						break
					}
					item := sel.Item(family.Name).Label(family.Name)
					item.Children(func() { components.Text(c, family.Name).Font(family.Name).SingleLine() })
				}
				if shown == 0 {
					components.Text(c, m.FontsNoMatch()).Padding(t.Space(2)).TextColor(t.TextMuted)
				}
			})
		})
		components.Checkbox(c, &d.monospace, m.FontsMonospaced())
	})
	components.Text(c, m.FontsSample()).Font(d.fontFamily).FontSize(t.Rem(1.4)).SingleLine()
	components.Text(c, m.FontsCount(len(families))).TextColor(t.TextMuted)
}

func exampleSystemAppearance(c *ui.Context, s *demoState) {
	m, p, t := msg(c), c.Preferences(), c.Theme()
	ui.Row(c).FillWidth().Gap(12).AlignItems(ui.Center).Children(func() {
		components.Text(c, m.AppearanceAccent()).TextColor(t.TextMuted).Width(160)
		if p.Accent.A == 0 {
			components.Text(c, m.AppearanceNoAccent())
		} else {
			ui.Box(c).Size(20, 20).Radius(10).Background(p.Accent).Label(m.AppearanceAccent())
		}
	})
	systemRow(c, m.AppearanceReduceMotion(), yesNo(m, p.ReduceMotion))
	systemRow(c, m.AppearanceHighContrast(), yesNo(m, p.HighContrast))
	scale := p.TextScale
	if scale == 0 {
		scale = 1
	}
	systemRow(c, m.AppearanceTextScale(), i18n.FormatNumber(c.Locale(), scale, i18n.NumberOptions{Percent: true}))
	if s.system.native {
		systemRow(c, m.AppearanceAppActive(), yesNo(m, mygo.App.IsActive()))
	}
}

func exampleLanguages(c *ui.Context, s *demoState) {
	m, r := msg(c), c.Region()
	_, format := formats(c)
	locale := c.Locale()
	if locale == "" {
		locale = "en"
	}
	systemRow(c, m.LanguagesShown(), i18n.Autonym(locale))
	var preferred []string
	if s.system.native {
		preferred = mygo.App.PreferredLanguages()
	}
	names := make([]string, 0, len(preferred))
	for _, tag := range preferred {
		names = append(names, i18n.Autonym(tag))
	}
	systemRow(c, m.LanguagesPreferred(), strings.Join(names, ", "))
	systemRow(c, m.LanguagesRegion(), r.Locale)
	systemRow(c, m.LanguagesFirstWeekday(), i18n.Weekdays(format, i18n.Wide, true)[r.FirstWeekday])
	systemRow(c, m.LanguagesNumber(), i18n.FormatNumber(format, 1234567.891, i18n.NumberOptions{}))
	systemRow(c, m.LanguagesDate(), i18n.FormatDate(format, time.Date(2026, 10, 15, 13, 7, 0, 0, time.Local), i18n.DateOptions{Date: i18n.FullStyle, Time: i18n.ShortStyle}))
}

// formats mirrors the regional formats MyGo's date controls use, with CLDR's
// defaults for the language where the system reports none.
func formats(c *ui.Context) (i18n.Region, string) {
	lang := c.Locale()
	if lang == "" {
		lang = "en"
	}
	r := c.Region()
	if r.HourCycle != 12 && r.HourCycle != 24 {
		r = i18n.RegionFor(lang)
		r.Locale = ""
	}
	return r, r.FormatLocale(lang)
}

func cookieResult(m locales.Messages, err error) string {
	if err != nil {
		return m.SystemError(err.Error())
	}
	return ""
}
