package sidebar

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/internal/l10n"
)

// resizeHandle uses MyGo's pointer capture and keyboard input, overlaid on
// the separator rather than inserting another layout column.
func (s *Shell) resizeHandle(c *ui.Context, width, maxWidth float32) {
	handle := ui.Box(c).Key("left-resize").Absolute().Left(width - 3).Top(0).Bottom(0).Width(6).
		Cursor(ui.CursorResizeEW).Focusable().Role(ui.RoleSplitter).Label(l10n.In(c.Locale()).Text("resizeSidebar"))
	next := width
	if dx, _, held := handle.Dragged(); held {
		handle.Focus()
		next += dx
	}
	if handle.Shortcut(0, ui.KeyLeft) {
		next -= 10
	}
	if handle.Shortcut(0, ui.KeyRight) {
		next += 10
	}
	if handle.DoubleClicked() {
		s.leftWidth = float32(s.options.LeftWidth)
		c.Invalidate()
	} else if next != width {
		s.leftWidth = max(float32(s.options.LeftMinWidth), min(next, maxWidth))
		c.Invalidate()
	}
	handle.FocusRing(handle.FocusVisible())
}

// rightResizeHandle is resizeHandle for the right sidebar, at the window's
// width less the sidebar's.
func (s *Shell) rightResizeHandle(c *ui.Context, windowWidth, width, maxWidth float32) {
	handle := ui.Box(c).Key("right-resize").Absolute().Left(windowWidth - width - 3).Top(0).Bottom(0).Width(6).
		Cursor(ui.CursorResizeEW).Focusable().Role(ui.RoleSplitter).Label(l10n.In(c.Locale()).Text("resizeRightSidebar"))
	next := width
	if dx, _, held := handle.Dragged(); held {
		handle.Focus()
		next -= dx
	}
	if handle.Shortcut(0, ui.KeyLeft) {
		next += 10
	}
	if handle.Shortcut(0, ui.KeyRight) {
		next -= 10
	}
	if handle.DoubleClicked() {
		s.rightWidth = float32(s.options.RightWidth)
		c.Invalidate()
	} else if next != width {
		s.rightWidth = max(float32(s.options.RightMinWidth), min(next, maxWidth))
		c.Invalidate()
	}
	handle.FocusRing(handle.FocusVisible())
}

// Separators overlay the surface; they never reserve layout space.
func shellSeparator(c *ui.Context) {
	ui.Box(c).Key("separator").Absolute().Left(0).Top(0).Bottom(0).Width(1).Background(c.Theme().Border)
}
