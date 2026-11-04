package components

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// menuState is what a menu's target keeps between frames: whether the menu
// shows, the point it opened at, from the target's top left corner, unless
// it shows below the target, and how it takes the focus.
type menuState struct {
	open   bool
	below  bool
	dx, dy float32
	// focus is set from the frame the menu opens until it took the focus:
	// its first item, when the keyboard opened it, else the panel.
	focus, keyboard bool
	// pointer is where the pointer last was over the panel, so an item
	// takes the focus as the pointer moves onto it, not while it rests.
	pointer [2]float32
}

// ContextMenu gives target a menu, drawn in a Popover with the look of the
// other popovers, that opens where target is right-clicked (Control-clicked
// on macOS), or for the menu key or Shift+F10 while it or an element inside
// it has the focus, below it unless the pointer rests on it. It leaves
// target's clicks to it and its parents. Call it every frame after building
// target, which must be keyed where it repeats. build fills the menu while
// it shows, and the item chosen reports Chosen, which closes it:
//
//	components.ContextMenu(c, row, func(m *components.Menu) {
//		if m.Item("Rename").Chosen() {
//			app.renaming = i
//		}
//		m.Group("Move to", func(m *components.Menu) {
//			for _, f := range app.folders {
//				if m.Item(f.Name).Chosen() {
//					app.move(i, f)
//				}
//			}
//		})
//		m.Separator()
//		if m.Item("Delete").Disabled(note.Locked).Chosen() {
//			app.delete(i)
//		}
//	})
//
// A disabled target opens none. Up and Down move among the items, past the
// disabled ones, Enter and Space choose one, and Escape or a press outside
// closes the menu.
func ContextMenu(c *ui.Context, target ui.Element, build func(m *Menu)) {
	st := ui.Local(target, "onevra-context-menu", func() menuState { return menuState{} })
	if st == nil {
		return
	}
	// MyGo asks for the menu as the system's would open, which it does not
	// with no items, leaving target's own clicks to it and its parents.
	asked := false
	target.ContextMenu(func(*ui.Menu) { asked = true })
	if asked {
		if x, y, over := target.PointerPosition(); over {
			*st = menuState{open: true, dx: x, dy: y, focus: true}
		} else {
			*st = menuState{open: true, below: true, focus: true, keyboard: true}
		}
	}
	showMenu(c, target, st, build)
}

// MenuButton makes button open a menu below it, as ContextMenu's, when it
// is clicked or activated with Enter or Space, and close it when clicked
// again. Call it every frame after building button, instead of reading
// button's Clicked:
//
//	more := components.IconButton(c, lucide.Ellipsis(), "More")
//	components.MenuButton(c, more, func(m *components.Menu) {
//		if m.Item("Duplicate").Chosen() {
//			app.duplicate()
//		}
//	})
func MenuButton(c *ui.Context, button ui.Element, build func(m *Menu)) {
	st := ui.Local(button, "onevra-menu-button", func() menuState { return menuState{} })
	if st == nil {
		return
	}
	if button.Clicked() {
		open := !st.open
		*st = menuState{open: open, below: true, focus: open, keyboard: button.FocusVisible()}
	}
	button.Role(ui.RoleMenuButton)
	showMenu(c, button, st, build)
}

// showMenu builds the menu of target while st shows it.
func showMenu(c *ui.Context, target ui.Element, st *menuState, build func(m *Menu)) {
	if !st.open {
		return
	}
	if target.IsDisabled() {
		st.open = false
		return
	}
	Popover(c, target, &st.open, func(panel ui.Element) {
		panel.Role(ui.RoleMenu).Focusable().MinWidth(180).MaxWidth(360).Padding(4).Gap(1)
		if !st.below {
			panel.AttachTo(target, ui.AnchorTopLeft, ui.AnchorTopLeft).Margin(st.dy, 0, 0, st.dx)
		}
		x, y, _ := panel.PointerPosition()
		moved := [2]float32{x, y} != st.pointer
		st.pointer = [2]float32{x, y}
		m := &Menu{c: c, st: st, moved: moved}
		panel.Children(func() {
			build(m)
			m.flush()
		})
		switch {
		case !st.open:
		case panel.Shortcut(0, ui.KeyDown):
			m.move(1)
		case panel.Shortcut(0, ui.KeyUp):
			m.move(-1)
		case panel.Shortcut(0, ui.KeyHome):
			m.focusAt(0, 1)
		case panel.Shortcut(0, ui.KeyEnd):
			m.focusAt(len(m.items)-1, -1)
		case st.focus && st.keyboard:
			m.focusAt(0, 1)
		case st.focus:
			panel.Focus()
		}
		st.focus = false
	})
}

// Menu is a menu being built, by the function given to ContextMenu or
// MenuButton. Its items show in the order they are added.
type Menu struct {
	c     *ui.Context
	st    *menuState
	moved bool
	// pending is the item added last, whose row builds once its options are
	// set: as the next is added, Chosen is called or the menu ends.
	pending *MenuItem
	// divide is set when a separator goes before the next item; shown, once
	// any row showed.
	divide, shown bool
	items         []menuRow
}

// menuRow is a row the keys may move the focus to.
type menuRow struct {
	e       ui.Element
	enabled bool
}

// MenuItem is an item of a Menu. Its options are set before Chosen, which
// reports whether it was chosen.
type MenuItem struct {
	m                 *Menu
	label, shortcut   string
	disabled, checked bool
	check             bool
	built, chosen     bool
}

// Item adds an item showing label.
func (m *Menu) Item(label string) *MenuItem {
	m.flush()
	it := &MenuItem{m: m, label: label}
	m.pending = it
	return it
}

// Separator adds a line between groups of items. It shows only between
// items, once.
func (m *Menu) Separator() {
	m.flush()
	m.divide = true
}

// Group adds the items build adds under a heading of label, which sets
// them apart without a separator.
func (m *Menu) Group(label string, build func(m *Menu)) {
	m.flush()
	m.dividing()
	t := m.c.Theme()
	Text(m.c, label).FontSize(11.5).FontWeight(500).TextColor(t.TextMuted).Padding(8, 8, 3, 8).SingleLine().Ellipsis("…")
	m.shown = true
	build(m)
	m.flush()
}

// dividing draws the separator asked for before a row, unless it would
// lead the menu.
func (m *Menu) dividing() {
	if m.divide && m.shown {
		t := m.c.Theme()
		ui.Box(m.c).FillWidth().Height(1).Margin(4, 8, 4, 8).Background(t.Border).Shrink(0)
	}
	m.divide = false
}

// Disabled shows the item grayed out when d is true: it cannot be chosen.
func (it *MenuItem) Disabled(d bool) *MenuItem {
	it.disabled = d
	return it
}

// Checked shows a check mark by the item when on is true, as for a setting
// that choosing the item toggles. The menu leaves room for it either way.
func (it *MenuItem) Checked(on bool) *MenuItem {
	it.check, it.checked = true, on
	return it
}

// Shortcut shows a key hint, such as "⌘B", after the item's label. It only
// shows the key; the application handles it.
func (it *MenuItem) Shortcut(hint string) *MenuItem {
	it.shortcut = hint
	return it
}

// Chosen reports whether the item was chosen in this frame, which closes
// the menu.
func (it *MenuItem) Chosen() bool {
	it.m.flush()
	return it.chosen
}

// flush builds the row of the pending item.
func (m *Menu) flush() {
	it := m.pending
	if it == nil || it.built {
		return
	}
	m.pending, it.built = nil, true
	m.dividing()
	m.shown = true
	c, t := m.c, m.c.Theme()
	b := ui.ButtonBase(c).Key(len(m.items)).Label(it.label).Disabled(it.disabled).FillWidth().MinWidth(0).
		Padding(6, 8).Radius(6).Justify(ui.Start).AlignItems(ui.Center).Gap(8).Background(ui.Color{}).TextColor(t.Text)
	if it.check {
		b.Role(ui.RoleMenuItemCheckBox).Checked(it.checked)
	} else {
		b.Role(ui.RoleMenuItem)
	}
	m.items = append(m.items, menuRow{e: b, enabled: !it.disabled})
	switch {
	case it.disabled:
		b.TextColor(t.TextMuted)
	case b.Hovered() && m.moved && !b.Focused():
		// The pointer moved onto it: the keys go on from it.
		b.Focus()
		fallthrough
	case b.Hovered(), b.Focused():
		b.Background(t.Text.Alpha(.05))
	}
	b.Children(func() {
		if it.check {
			mark := ui.Box(c).Size(12, 12).Shrink(0)
			if it.checked {
				mark.Children(func() { Icon(c, lucide.Check).Size(12, 12).TextColor(t.Accent) })
			}
		}
		Text(c, it.label).Grow(1).MinWidth(0).SingleLine().Ellipsis("…").FontSize(13)
		if it.shortcut != "" {
			Text(c, it.shortcut).Shrink(0).FontSize(11.5).TextColor(t.TextMuted).MarginStart(16)
		}
	})
	if b.Clicked() {
		it.chosen, m.st.open = true, false
	}
}

// move moves the focus by step among the items that can be chosen, round,
// from the one that has it, or from the start or the end.
func (m *Menu) move(step int) {
	at := -1
	for i, r := range m.items {
		if r.e.Focused() {
			at = i
		}
	}
	if at < 0 {
		if step > 0 {
			m.focusAt(0, 1)
		} else {
			m.focusAt(len(m.items)-1, -1)
		}
		return
	}
	n := len(m.items)
	for i := 1; i <= n; i++ {
		if r := m.items[((at+step*i)%n+n)%n]; r.enabled {
			r.e.Focus()
			return
		}
	}
}

// focusAt gives the focus to the first item that can be chosen from i on,
// going by step.
func (m *Menu) focusAt(i, step int) {
	for ; i >= 0 && i < len(m.items); i += step {
		if m.items[i].enabled {
			m.items[i].e.Focus()
			return
		}
	}
}
