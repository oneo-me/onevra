package components

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// TimeControl exposes layout and a frame-local time change. Keep this handle
// when applying Element layout methods, which return ui.Element.
type TimeControl struct {
	ui.Element
	changed bool
}

// Changed reports a confirmed time change, not opening or draft navigation.
func (t *TimeControl) Changed() bool { return t.changed }

type timePickerState struct {
	hour, minute int
	last         time.Time
	reveal       bool
}

func (s *timePickerState) reset(value time.Time) {
	s.hour, s.minute, s.last, s.reveal = value.Hour(), value.Minute(), value, true
}

func timePickerSet(value *time.Time, hour, minute int) bool {
	next := time.Date(value.Year(), value.Month(), value.Day(), hour, minute, value.Second(), value.Nanosecond(), value.Location())
	if next.Equal(*value) {
		return false
	}
	*value = next
	return true
}

// TimePicker displays a leading clock icon and a 24-hour HH:mm value. Its fixed
// popup contains independently scrolling hour (left) and minute (right) columns
// without scrollbars. Choosing an hour moves focus to minutes; choosing a minute
// confirms without closing the popup.
// Up/Down and Home/End move the focused column's draft cursor; Left/Right switch
// columns; Enter/Space activates it. Escape/outside press discards the draft.
// Opening reveals the current hour and minute. External changes reset the draft.
// Confirmation preserves the date, seconds, nanoseconds and location, subject to
// time.Date's normalization of nonexistent/ambiguous local times at DST changes.
// Key or disable a parent before construction; Label names the trigger.
func TimePicker(c *ui.Context, value *time.Time) *TimeControl {
	words := localized(c)
	if value == nil {
		panic("components.TimePicker: nil time")
	}
	t := c.Theme()
	out := &TimeControl{Element: pickerTrigger(c, lucide.Clock())}
	open := ui.Local(out.Element, "time-open", func() bool { return false })
	st := ui.Local(out.Element, "time-draft", func() timePickerState {
		s := timePickerState{}
		s.reset(*value)
		return s
	})
	if out.IsDisabled() {
		*open = false
	} else if out.Clicked() {
		*open = !*open
		if *open {
			st.reset(*value)
		}
	}
	if st.last != *value {
		st.reset(*value)
	}
	Popover(c, out.Element, open, func(panel ui.Element) {
		panel.Row().Width(t.Space(48)).Height(t.Space(64)).Padding(t.Space(2)).Gap(t.Space(2)).Label(words.ChooseTime)
		var columns [2]ui.Element
		focus := -1
		confirm := false
		for index, label := range []string{words.Hours, words.Minutes} {
			ui.Column(c).Key(index).Width(0).Grow(1).FillHeight().Gap(t.Space(2)).Children(func() {
				Text(c, label).SingleLine().TextColor(t.TextMuted).FillWidth().TextAlign(ui.Center)
				current, count := &st.hour, 24
				if index == 1 {
					current, count = &st.minute, 60
				}
				column := ScrollView(c, ScrollViewOptions{}).Scrollbars(false).
					FillWidth().Grow(1).Focusable().Role(ui.RoleList).Label(label)
				columns[index] = column
				if index == 0 {
					column.AutoFocus()
				}
				before := *current
				activate := func() {
					if index == 0 {
						focus = 1
					} else {
						confirm = true
					}
				}
				if !column.IsDisabled() {
					switch {
					case column.Shortcut(0, ui.KeyUp):
						*current = max(0, *current-1)
					case column.Shortcut(0, ui.KeyDown):
						*current = min(count-1, *current+1)
					case column.Shortcut(0, ui.KeyHome):
						*current = 0
					case column.Shortcut(0, ui.KeyEnd):
						*current = count - 1
					case column.Shortcut(0, ui.KeyLeft):
						focus = 0
					case column.Shortcut(0, ui.KeyRight):
						focus = 1
					case column.Shortcut(0, ui.KeyEnter), column.Shortcut(0, ui.KeySpace):
						activate()
					}
				}
				column.Value(fmt.Sprintf("%02d", *current))
				column.Children(func() {
					for n := range count {
						selected := n == *current
						row := ui.Box(c).FillWidth().Height(t.Space(8)).Shrink(0).Center().Radius(t.Radius).
							Role(ui.RoleListItem).Label(fmt.Sprintf("%s %02d", label, n)).Selected(selected)
						if row.Clicked() {
							*current = n
							column.Focus()
							activate()
						}
						if selected {
							row.Background(t.Accent).TextColor(t.AccentText)
							column.ActiveDescendant(row)
							if st.reveal || before != *current {
								row.ScrollIntoView()
							}
						} else if !row.IsDisabled() && row.Hovered() {
							row.Background(t.SurfaceHover)
						}
						row.Children(func() { Text(c, fmt.Sprintf("%02d", n)).FontFeatures("tnum") })
					}
				})
			})
		}
		st.reveal = false
		if focus >= 0 {
			columns[focus].Focus()
		}
		if confirm {
			out.changed = timePickerSet(value, st.hour, st.minute)
			st.last = *value
		}
	})
	out.Children(func() { Text(c, value.Format("15:04")).SingleLine().FontFeatures("tnum") })
	out.Expanded(*open).Value(value.Format("15:04"))
	return out
}
