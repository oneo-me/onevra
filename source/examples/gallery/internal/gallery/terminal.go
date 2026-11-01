package gallery

import (
	"io"
	"sync"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/oneo-me/onevra/source/shells/sidebar"
)

// terminalDemo loads the optional native library once when its page is entered.
// Loading never blocks a frame. A stopped load cannot resurrect its terminal.
type terminalDemo struct {
	mu                        sync.Mutex
	term                      *terminal.Terminal
	started, loading, stopped bool
	err                       error

	// The page's own state, on the UI thread: the find bar and its query, a
	// user's copy not yet told, the contrast choice and the mirror. refocus
	// returns the keyboard to the session as the find bar closes.
	finding, refocus, copied, contrast bool
	query                              string
	split                              float32
	mirror                             *terminalMirror
}

func (d *terminalDemo) start(create func() (*terminal.Terminal, error)) {
	d.mu.Lock()
	if d.started {
		d.mu.Unlock()
		return
	}
	d.started, d.loading, d.stopped, d.err = true, true, false, nil
	d.mu.Unlock()
	go func() {
		term, err := create()
		d.mu.Lock()
		defer d.mu.Unlock()
		d.loading = false
		if d.stopped {
			if term != nil {
				_ = term.Close()
			}
			return
		}
		d.term, d.err = term, err
	}()
}

func (d *terminalDemo) snapshot() (*terminal.Terminal, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.term, d.loading, d.err
}

func (d *terminalDemo) close() {
	d.closeMirror()
	d.finding, d.query = false, ""
	d.mu.Lock()
	d.started, d.stopped = false, true
	term := d.term
	d.term, d.err = nil, nil
	d.mu.Unlock()
	if term != nil {
		_ = term.Close()
	}
}

// options are the session's: its padding takes the colors of the cells at
// the edges, and matches and the user's copies reach the page through
// refresh, on the main thread.
func (d *terminalDemo) options(refresh func()) terminal.Options {
	return terminal.Options{
		PaddingColor: terminal.PaddingExtend,
		OnSearch:     func(terminal.SearchStatus) { refresh() },
		OnCopy: func(_ string, by terminal.CopySource) {
			// Programs copy through OSC 52 as they please; only the user's
			// copies are told.
			if by != terminal.CopyProgram {
				d.copied = true
				refresh()
			}
		},
	}
}

func newShellTerminal(o terminal.Options) (*terminal.Terminal, error) {
	return terminal.New(o)
}

func exampleTerminal(c *ui.Context, s *demoState) {
	d := &s.terminal
	term, loading, err := d.snapshot()
	if loading {
		components.Spinner(c, components.SpinnerSpokes).Label(tr(c, "loadingTerminal"))
		c.After(50 * time.Millisecond)
		return
	}
	if err != nil {
		components.Text(c, tr(c, "terminalUnavailable")).Bold()
		components.Text(c, err.Error()).TextColor(c.Theme().Danger).Wrap()
		return
	}
	if term == nil {
		return
	}
	if d.copied {
		d.copied = false
		c.AddToast(ui.Toast{ID: "gallery-terminal-copy", Title: tr(c, "terminalCopied")})
	}
	if d.finding {
		terminalFindBar(c, d, term)
	}
	session := func() {
		view := terminalPane(c, term, "terminal-session", "")
		switch {
		case d.refocus:
			view.Focus()
			d.refocus = false
		case !d.finding:
			view.AutoFocus()
		}
	}
	if d.mirror == nil {
		session()
		return
	}
	components.SplitPercent(c, &d.split, session, func() {
		terminalPane(c, d.mirror.term, "terminal-mirror", tr(c, "terminalMirrorLabel"))
	}).Fill()
}

// terminalPane shows a terminal and, over it, the link Command (Control
// elsewhere) is held over, as a preview of where a click goes.
func terminalPane(c *ui.Context, term *terminal.Terminal, key, label string) ui.Element {
	t := c.Theme()
	var view ui.Element
	ui.Box(c).Key(key).Fill().MinWidth(0).Children(func() {
		view = terminal.View(c, term).Fill()
		if label != "" {
			view.Label(label)
		}
		link, path, at := term.HoveredLink()
		if link == "" {
			return
		}
		icon := lucide.Link()
		if path {
			icon = lucide.File()
		}
		ui.Row(c).Absolute().Left(at.X).Top(at.Y+at.H+t.Space(1)).MaxWidth(t.Space(90)).PassThrough().
			Padding(t.Space(1), t.Space(2)).Gap(t.Space(1.5)).AlignItems(ui.Center).
			Radius(t.Radius).Background(t.Surface).Border(1, t.Border).Children(func() {
			components.Icon(c, icon).Size(t.Rem(.875), t.Rem(.875)).TextColor(t.TextMuted)
			components.Text(c, link).FontSize(t.Rem(.75)).SingleLine().Ellipsis("…").MinWidth(0).Shrink(1)
		})
	})
	return view
}

// terminalFindBar searches the session's screen and scrollback: the
// terminal highlights the matches as it finds them, and the bar counts them.
// Older matches are up the scrollback.
func terminalFindBar(c *ui.Context, d *terminalDemo, term *terminal.Terminal) {
	t := c.Theme()
	status := term.SearchStatus()
	bar := ui.Row(c).Key("terminal-find").FillWidth().Shrink(0).Padding(t.Space(2), t.Space(3)).
		Gap(t.Space(1)).AlignItems(ui.Center).Background(t.Background)
	bar.Children(func() {
		field := components.TextBox(c, &d.query, components.TextBoxOptions{LeadingIcon: lucide.Search()}).
			Label(tr(c, "terminalFind")).Placeholder(tr(c, "terminalFind")).Width(t.Space(60)).Shrink(1).AutoFocus()
		if field.Changed() {
			term.Search(d.query)
		}
		if field.Submitted() {
			term.SearchNext()
		}
		if status.Text != "" {
			components.Text(c, msg(c).TerminalMatches(status.Selected, status.Total)).
				FontSize(t.Rem(.75)).TextColor(t.TextMuted).SingleLine().Padding(0, t.Space(1))
		}
		ui.Spacer(c)
		if sidebar.IconButton(c, lucide.ChevronUp(), tr(c, "terminalOlderMatch")).Disabled(status.Total == 0).Clicked() {
			term.SearchNext()
		}
		if sidebar.IconButton(c, lucide.ChevronDown(), tr(c, "terminalNewerMatch")).Disabled(status.Total == 0).Clicked() {
			term.SearchPrevious()
		}
		if sidebar.IconButton(c, lucide.X(), tr(c, "terminalEndFind")).Clicked() {
			d.endFind(term)
		}
	})
	// Escape belongs to programs in the terminal; it closes the bar only
	// while the focus is in it.
	if bar.Shortcut(0, ui.KeyEscape) {
		d.endFind(term)
	}
	ui.Box(c).Height(1).FillWidth().Shrink(0).Background(t.Border)
}

func (d *terminalDemo) endFind(term *terminal.Terminal) {
	d.finding, d.refocus, d.query = false, true, ""
	term.EndSearch()
}

// terminalToolbar finds and clears from the titlebar once the session is
// running; its menu switches copy on select, the minimum contrast and the
// mirror.
func terminalToolbar(c *ui.Context, s *demoState) {
	d := &s.terminal
	term, _, _ := d.snapshot()
	if term == nil {
		return
	}
	if sidebar.IconButton(c, lucide.Search(), tr(c, "terminalFind")).Clicked() {
		if d.finding {
			d.endFind(term)
		} else {
			d.finding = true
		}
	}
	if sidebar.IconButton(c, lucide.Eraser(), tr(c, "clear")).Clicked() {
		term.ClearScreen()
	}
	sidebar.IconButton(c, lucide.SlidersHorizontal(), tr(c, "terminalOptions")).Menu(func(m *ui.Menu) {
		if m.Item(tr(c, "terminalCopyOnSelect")).Checked(s.on).Chosen() {
			s.on = !s.on
			mode := terminal.CopyOnSelectNone
			if s.on {
				mode = terminal.CopyOnSelectClipboard
			}
			term.SetCopyOnSelect(mode)
		}
		if m.Item(tr(c, "terminalMinimumContrast")).Checked(d.contrast).Chosen() {
			d.contrast = !d.contrast
			// WCAG's AA ratio for body text; 1 leaves the colors as programs set them.
			ratio := float32(1)
			if d.contrast {
				ratio = 4.5
			}
			term.SetMinimumContrast(ratio)
		}
		if m.Item(tr(c, "terminalMirror")).Checked(d.mirror != nil).Chosen() {
			if d.mirror != nil {
				d.closeMirror()
			} else if err := d.openMirror(term); err != nil {
				c.AddToast(ui.Toast{ID: "gallery-terminal-mirror", Title: tr(c, "terminalMirrorFailed"),
					Description: err.Error(), Type: components.ToastError})
			}
		}
	})
}

func (d *terminalDemo) openMirror(source *terminal.Terminal) error {
	m, err := newTerminalMirror(source)
	if err != nil {
		return err
	}
	d.mirror, d.split = m, 50
	return nil
}

func (d *terminalDemo) closeMirror() {
	if d.mirror != nil {
		d.mirror.close()
		d.mirror = nil
	}
}

// terminalMirror shows the session again from its Mirror's frames, as the
// window of a multiplexer shows a session its server keeps. The screen
// keeps the session's size (ResizeByConn), and what is typed in it goes to
// the session's shell.
type terminalMirror struct {
	source *terminal.Terminal
	frames *terminal.Mirror
	term   *terminal.Terminal
	stop   chan struct{}
	done   chan struct{}
}

func newTerminalMirror(source *terminal.Terminal) (*terminalMirror, error) {
	frames, err := source.NewMirror()
	if err != nil {
		return nil, err
	}
	conn := &mirrorConn{source: source, closed: make(chan struct{})}
	term, err := terminal.New(terminal.Options{Conn: conn, ResizeByConn: true, PaddingColor: terminal.PaddingExtend})
	if err != nil {
		frames.Close()
		_ = conn.Close()
		return nil, err
	}
	m := &terminalMirror{source: source, frames: frames, term: term, stop: make(chan struct{}), done: make(chan struct{})}
	go m.pump()
	return m, nil
}

// pump brings the mirror to what the session shows, at the session's size,
// 30 times a second, as a server's stream would.
func (m *terminalMirror) pump() {
	defer close(m.done)
	tick := time.NewTicker(33 * time.Millisecond)
	defer tick.Stop()
	cols, rows := 0, 0
	for {
		select {
		case <-m.stop:
			return
		case <-tick.C:
		}
		if c, r := m.source.Size(); c != cols || r != rows {
			cols, rows = c, r
			m.term.Resize(cols, rows)
		}
		if frame, _ := m.frames.Frame(); len(frame) > 0 {
			m.term.Feed(frame)
		}
	}
}

func (m *terminalMirror) close() {
	close(m.stop)
	<-m.done
	m.frames.Close()
	_ = m.term.Close()
}

// mirrorConn connects the mirror: its frames come through Feed, and what
// is typed goes to the session's shell.
type mirrorConn struct {
	source *terminal.Terminal
	once   sync.Once
	closed chan struct{}
}

func (c *mirrorConn) Read([]byte) (int, error) {
	<-c.closed
	return 0, io.EOF
}

func (c *mirrorConn) Write(p []byte) (int, error) {
	c.source.Send(p)
	return len(p), nil
}

func (c *mirrorConn) Close() error {
	c.once.Do(func() { close(c.closed) })
	return nil
}

// Close releases gallery-owned native resources. Call on the UI thread when
// its window closes, or after a headless view is no longer used.
func (g *Gallery) Close() {
	g.commands.palette.ReleaseInputSource()
	g.notification.close()
	g.dock.close()
	g.tray.close()
	for _, s := range g.mygoStates {
		s.terminal.close()
		s.webView.close()
		s.audio.close()
		s.system.close()
		s.flyout.close()
		s.video.close()
	}
}
