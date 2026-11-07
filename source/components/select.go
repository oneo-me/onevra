package components

import (
	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// SelectParts composes MyGo's select behavior with a themed popup and an
// optional search field. Applications own the query, filtering and option text.
// Embedded MyGo parts retain trigger, item and open-state behavior.
type SelectParts[T comparable] struct {
	ui.SelectParts[T]
	c         *ui.Context
	selected  *T
	searching bool
	options   []selectOption[T]
}

// selectOption is an option built in the current Popup pass.
type selectOption[T comparable] struct {
	value T
	item  ui.Element
}

// Popup shows a themed, shadowed panel while the select is open. The callback
// can override its appearance and build a SearchField followed by filtered Items.
// It returns an absent Element when closed.
func (s *SelectParts[T]) Popup(build func(panel ui.Element)) ui.Element {
	t := s.c.Theme()
	s.options = s.options[:0]
	popup := s.SelectParts.Popup(func(panel ui.Element) {
		panel.Padding(t.Space(1)).Radius(t.Radius*1.5).Material(glass.Glass{Tint: t.Background.Alpha(.92)}).Border(1, t.Border).
			Shadow(0, t.Space(1), t.Space(4), 0, ui.RGBA(0, 0, 0, 0.14))
		build(panel)
	})
	if s.searching {
		s.highlightResult()
	}
	return popup
}

// highlightResult keeps the highlight on a filtered option: MyGo's arrows move
// from the highlighted option and do not bring back one the query filtered out.
func (s *SelectParts[T]) highlightResult() {
	first := -1
	for i, option := range s.options {
		if option.item.Highlighted() {
			return
		}
		if first < 0 && !option.item.IsDisabled() {
			first = i
		}
	}
	if first >= 0 {
		s.Highlight(s.options[first].value)
	}
}

// Item creates a themed option with the same neutral selected face as SidebarItem.
// Unselected options use a weaker neutral face while hovered. The caller
// supplies its label and children.
func (s *SelectParts[T]) Item(value T) ui.Element {
	t := s.c.Theme()
	item := s.SelectParts.Item(value).Padding(t.Space(1.5), t.Space(2.5)).Radius(t.Radius).
		Background(ui.Color{}).TextColor(t.Text)
	s.options = append(s.options, selectOption[T]{value, item})
	if value == *s.selected {
		item.Background(t.Text.Alpha(.1))
	} else if item.Hovered() {
		item.Background(t.Text.Alpha(.04))
	}
	return item
}

// SearchField adds an optional search input inside Popup. Give it an accessible
// Label and localized Placeholder. Build only matching Items after it; an empty
// result message and query reset policy belong to the caller. The highlight
// moves to the first enabled result when filtering removes its option. Down or
// Up returns focus to the trigger, where MyGo handles option navigation and
// selection.
func (s *SelectParts[T]) SearchField(query *string) ui.Element {
	t := s.c.Theme()
	s.searching = true
	input := TextBox(s.c, query, TextBoxOptions{LeadingIcon: lucide.Search()}).
		FillWidth().AutoFocus().Border(1, t.Border).Radius(t.Radius).Margin(0, 0, t.Space(1.5), 0)
	if input.Shortcut(0, ui.KeyDown) || input.Shortcut(0, ui.KeyUp) {
		s.Trigger.Focus()
	}
	return input
}

// Select creates a drop-down choosing *selected from options the caller
// builds in Popup with Item. The trigger, popup and items have Onevra styling:
//
//	sel := components.Select(c, &choice)
//	sel.Trigger.Children(func() { components.Text(c, string(choice)).SingleLine() })
//	sel.Popup(func(panel ui.Element) {
//		for _, opt := range options {
//			item := sel.Item(opt)
//			item.Children(func() { components.Text(c, string(opt)).SingleLine() })
//		}
//	})
//
// The trigger's Changed reports a new choice.
func Select[T comparable](c *ui.Context, selected *T) *SelectParts[T] {
	sel := ui.SelectBase(c, selected)
	styleSelectTrigger(c, sel.Trigger)
	return &SelectParts[T]{SelectParts: sel, c: c, selected: selected}
}

func styleSelectTrigger(c *ui.Context, b ui.Element) {
	t := c.Theme()
	b.Padding(t.Space(1.5), t.Space(6), t.Space(1.5), t.Space(2.5)).Radius(t.Radius)
	b.Justify(ui.Start)
	bg := t.Surface
	if !b.IsDisabled() {
		switch {
		case b.Pressed():
			bg = t.SurfacePressed
		case b.Hovered():
			bg = t.SurfaceHover
		}
	}
	b.Background(bg).TextColor(t.Text).Border(1, t.Border)
	b.DrawOver(func(p *ui.Painter, r ui.Rect) {
		s := t.Space(2)
		cx := r.X + r.W - t.Space(3)
		cy := r.Y + r.H/2
		var path ui.Path
		path.MoveTo(cx-s*0.35, cy-s*0.15).LineTo(cx, cy+s*0.2).LineTo(cx+s*0.35, cy-s*0.15)
		p.StrokePath(&path, 1.5, t.TextMuted)
	})
	dimmed(b)
}
