package gallery

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// inputDemo holds what the Keys & focus page last saw. It stays in memory.
type inputDemo struct {
	code      ui.KeyCode
	observed  int
	repeated  bool
	windowKey int
	left      int
	missed    ui.Point
	missedSet bool
	inside    int
	pointer   string
	draft     string
	written   bool
	writeErr  error
}

// exampleInputEvents reads keys by position, follows focus, tells the pointer
// and reports what shows of the page and the window.
func exampleInputEvents(c *ui.Context, s *demoState) {
	t := c.Theme()
	d := &s.input
	line := func(name, value string) {
		ui.Row(c).Key(name).Gap(t.Space(3)).Children(func() {
			components.Text(c, name).TextColor(t.TextMuted).Font("monospace")
			components.Text(c, value)
		})
	}
	column := ui.Column(c).Gap(t.Space(2)).FillWidth().MaxWidth(t.Space(100))
	column.Children(func() {
		switch s.variant {
		case "Focus":
			field := components.TextBox(c, &d.draft, components.TextBoxOptions{}).Label("FocusLeft").MaxLength(12).FillWidth()
			if field.FocusLeft() {
				d.left++
			}
			line("FocusLeft", fmt.Sprint(d.left))
			line("FocusTakesText", fmt.Sprint(c.FocusTakesText()))
			ui.Row(c).Gap(t.Space(2)).Children(func() {
				components.Button(c, "TabStop(true)", components.ButtonOutline)
				components.Button(c, "TabStop(false)", components.ButtonOutline).TabStop(false)
			})
			ui.Row(c).Gap(t.Space(2)).AlignItems(ui.Center).Children(func() {
				caption := components.Text(c, tr(c, "notifications"))
				components.Switch(c, &s.on, "").LabelledBy(caption.Element)
			})
		case "Pointer":
			pointerEvents(c, s, line)
		case "Page":
			p := c.Preferences()
			line("Preferences.Pointer", fmt.Sprint(p.Pointer))
			line("Preferences.AnyPointer", fmt.Sprint(p.AnyPointer))
			line("Preferences.NoHover", fmt.Sprint(p.NoHover))
			line("Preferences.NoAnyHover", fmt.Sprint(p.NoAnyHover))
			line("WindowHidden", fmt.Sprint(c.WindowHidden()))
			var soon, half ui.Element
			components.ScrollView(c, components.ScrollViewOptions{}).FillWidth().Height(t.Space(30)).Children(func() {
				ui.Box(c).Height(t.Space(40))
				soon = ui.Box(c).Height(t.Space(10)).Background(t.Surface)
				ui.Box(c).Height(t.Space(30))
				half = ui.Box(c).Height(t.Space(30)).Background(t.SurfaceHover)
			})
			line("InViewMargin", fmt.Sprint(soon.InViewMargin(t.Space(24))))
			line("InViewAmount", fmt.Sprint(half.InViewAmount(.5)))
		default:
			keyEvents(c, s, line)
		}
	})
}

func keyEvents(c *ui.Context, s *demoState, line func(name, value string)) {
	t := c.Theme()
	d := &s.input
	ui.Box(c).FillWidth().Height(t.Space(24)).Radius(t.Radius).Border(1, t.Border).Center().Focusable().
		Label("ObserveInput").ObserveInput(func(ev ui.InputEvent) {
		if ev.Kind == ui.InputKeyDown {
			d.code, d.observed = ev.Code, d.observed+1
		}
	}).Children(func() {
		components.Text(c, "ObserveInput").TextColor(t.TextMuted)
	})
	c.OnWindowShortcut(ui.Shift, ui.KeyF8, func() { d.windowKey++ })
	if c.Shortcut(ui.Shift, ui.KeyF7) {
		d.repeated = c.ShortcutRepeated()
	}
	var held []string
	for _, key := range c.KeysDown() {
		held = append(held, fmt.Sprint(key))
	}
	line("InputEvent.Code", fmt.Sprint(d.code))
	line("ObserveInput", fmt.Sprint(d.observed))
	line("KeysDown", strings.Join(held, " + "))
	line("KeyDown(Shift+F7)", fmt.Sprint(c.KeyDown(ui.KeyF7)))
	line("ShortcutRepeated", fmt.Sprint(d.repeated))
	line("OnWindowShortcut(Shift+F8)", fmt.Sprint(d.windowKey))
}

func pointerEvents(c *ui.Context, s *demoState, line func(name, value string)) {
	t := c.Theme()
	d := &s.input
	area := ui.Box(c).FillWidth().Height(t.Space(20)).Radius(t.Radius).Border(1, t.Border).Center().Label("PressedPointer")
	if pointer, pressure, ok := area.PressedPointer(); ok {
		d.pointer = fmt.Sprintf("%v %.2f", pointer, pressure)
	}
	if button, ok := area.PressedButton(); ok {
		line("PressedButton", fmt.Sprint(button))
	}
	if area.PressedInside() {
		d.inside++
	}
	if area.PressCanceled() {
		d.pointer = "PressCanceled"
	}
	area.Children(func() {
		components.Button(c, "PressedInside", components.ButtonOutline)
	})
	line("PressedPointer", d.pointer)
	line("PressedInside", fmt.Sprint(d.inside))
	chip := ui.Box(c).Padding(t.Space(2), t.Space(3)).Radius(t.Radius).Background(t.Surface).Label("DragMissed").Drag("chip")
	chip.Children(func() { components.Text(c, "DragMissed") })
	if at, ok := chip.DragMissed(); ok {
		d.missed, d.missedSet = at, true
	}
	if d.missedSet {
		line("DragMissed", fmt.Sprintf("%.0f, %.0f", d.missed.X, d.missed.Y))
	}
	ui.Row(c).Gap(t.Space(1)).Wrap().Children(func() {
		for _, cursor := range []struct {
			name   string
			cursor ui.Cursor
		}{{"Progress", ui.CursorProgress}, {"Wait", ui.CursorWait}, {"Help", ui.CursorHelp}, {"Cell", ui.CursorCell}, {"ZoomIn", ui.CursorZoomIn}} {
			ui.Box(c).Key(cursor.name).Padding(t.Space(1), t.Space(2)).Radius(t.Radius).Background(t.Surface).Cursor(cursor.cursor).
				Children(func() { components.Text(c, cursor.name) })
		}
	})
	if components.Button(c, "WriteClipboardThen", components.ButtonOutline).Clicked() {
		c.WriteClipboardThen("MyGo", func(err error) { d.written, d.writeErr = true, err })
	}
	if d.written {
		line("WriteClipboardThen", fmt.Sprint(d.writeErr))
	}
}
