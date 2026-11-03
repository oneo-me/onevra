package sidebar

import (
	"github.com/oneo-me/onevra/source/internal/l10n"
	"math"
	"runtime"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// ShellOptions describes a window and its default sidebar-and-content page.
// Content builders run synchronously on the UI thread.
type ShellOptions struct {
	Title string
	// Hidden creates the window without showing or focusing it. Applications
	// can later call Window().Show or Window().ShowInactive.
	Hidden bool
	// Window dimensions and constraints are outer DIPs. Non-positive width and
	// height default to 820 and 740, respectively, for both initial and minimum
	// sizes. Explicit positive values override each default independently.
	// Initial dimensions are raised to the corresponding minimum if needed.
	Width, Height, MinWidth, MinHeight int
	// StateKey, when non-empty, has MyGo remember the window's bounds and
	// maximized or full-screen state under this key and restore them at
	// creation. MyGo stores them in its user-data directory, or in the store
	// set with mygo.App.SetWindowStateStore. Use a distinct key per window kind.
	StateKey string
	// Left and Content build full-height regions behind the floating header.
	// Use HeaderHeight(c) as top padding to keep controls below it while
	// allowing a region's background to extend to the top. Nil leaves it empty.
	Left, Content func(*ui.Context)
	// ContentTitle builds the leading main-content titlebar area. It may contain
	// a title, breadcrumbs or actions; nil leaves it blank. Shell clips it to
	// the room after the sidebar control and hides it on replacement pages.
	ContentTitle func(*ui.Context)
	// LeftHeader builds trailing actions beside the sidebar toggle, within the
	// expanded sidebar's header. It is clipped to the available safe area and
	// omitted while collapsed, animating, or showing a replacement page.
	LeftHeader func(*ui.Context)
	// LeftToolbar builds a fixed bottom toolbar beneath Left; nil omits it.
	// Use IconButton for actions and ui.Spacer for trailing alignment.
	LeftToolbar func(*ui.Context)
	// Default left sidebar width in DIPs; non-positive values use 220.
	// Clamped to LeftMinWidth/LeftMaxWidth at creation. Dragging or focused
	// arrow keys adjust the width; double-clicking the separator restores it.
	// Shell retains the chosen width across toggles and page changes.
	LeftWidth int
	// LeftMinWidth defaults to 120 DIPs when non-positive. A non-positive
	// LeftMaxWidth adds no fixed limit; a positive maximum below the minimum
	// is raised to the minimum. Visible width is capped only to keep the resize
	// handle inside the viewport, even below the minimum, without discarding
	// the chosen width.
	LeftMinWidth, LeftMaxWidth int
	// The left sidebar starts open.
	LeftCollapsed bool
	// Right, when set, adds a full-height sidebar at the window's trailing
	// edge, outside the floating header, which then spans only Left and
	// Content and ends with the right sidebar's toggle while it is closed. Like Left, it shows
	// the window's material and owns its top inset: use HeaderHeight(c) to
	// keep controls below RightHeader.
	Right func(*ui.Context)
	// RightHeader builds the right sidebar's own draggable header, clipped to
	// its width, before the sidebar's toggle, which ends that header while
	// the sidebar is open. It is omitted while the right sidebar is
	// collapsed, animating, or a replacement page shows. On Windows the
	// caption buttons follow the toggle.
	RightHeader func(*ui.Context)
	// RightWidth, RightMinWidth and RightMaxWidth are LeftWidth's and its
	// limits for the right sidebar; non-positive values use 320 and 200.
	RightWidth                   int
	RightMinWidth, RightMaxWidth int
	// RightCollapsed starts the right sidebar closed.
	RightCollapsed bool
	// Toasts places the window's toasts; top placements also clear the header.
	Toasts components.ToastOptions
	// Theme, when set, returns the frame's theme; Shell installs it before
	// building anything, so its own surfaces and all content use it. It runs
	// on every frame, so it can follow application state; pass theme.Onevra
	// for Onevra's palettes. Nil keeps MyGo's frame theme. Light or dark
	// belongs to mygo.Theme.SetSource, which also covers native chrome.
	Theme func(*ui.Context) *ui.Theme
}

// Shell owns window presentation, sidebar state and the active page.
// Applications retain ownership of business state and application lifetime.
// Call its methods on the UI thread, after MyGo is ready.
type Shell struct {
	window    *mygo.Window
	host      shellWindow
	options   ShellOptions
	page      func(*ui.Context)
	leftOpen  bool
	leftWidth float32
	// rightOpen and rightWidth are the right sidebar's, as leftOpen and
	// leftWidth.
	rightOpen  bool
	rightWidth float32
}

type shellWindow interface {
	Invalidate()
}

// NewShell creates a native window and installs the page container.
// It requests a titlebar material on macOS and Mica with custom caption buttons
// on Windows, whose maximize button brings up Windows 11's snap layouts. If no
// material is effective, it draws an opaque theme background. Other platforms
// retain their native controls. It persists no state itself; StateKey
// delegates window geometry to MyGo.
func NewShell(options ShellOptions) *Shell {
	s := newShell(options)
	s.window = mygo.NewWindow(s.nativeOptions())
	s.host = s.window
	if runtime.GOOS == "windows" {
		s.window.OnMaximize(s.refresh)
		s.window.OnUnmaximize(s.refresh)
	}
	return s
}

const (
	defaultWindowWidth  = 820
	defaultWindowHeight = 740
)

func newShell(options ShellOptions) *Shell {
	if options.Width <= 0 {
		options.Width = defaultWindowWidth
	}
	if options.Height <= 0 {
		options.Height = defaultWindowHeight
	}
	if options.MinWidth <= 0 {
		options.MinWidth = defaultWindowWidth
	}
	if options.MinHeight <= 0 {
		options.MinHeight = defaultWindowHeight
	}
	options.Width = max(options.Width, options.MinWidth)
	options.Height = max(options.Height, options.MinHeight)
	if options.LeftWidth <= 0 {
		options.LeftWidth = 220
	}
	if options.LeftMinWidth <= 0 {
		options.LeftMinWidth = 120
	}
	options.LeftWidth = max(options.LeftWidth, options.LeftMinWidth)
	if options.LeftMaxWidth > 0 {
		options.LeftMaxWidth = max(options.LeftMaxWidth, options.LeftMinWidth)
		options.LeftWidth = min(options.LeftWidth, options.LeftMaxWidth)
	}
	if options.RightWidth <= 0 {
		options.RightWidth = 320
	}
	if options.RightMinWidth <= 0 {
		options.RightMinWidth = 200
	}
	options.RightWidth = max(options.RightWidth, options.RightMinWidth)
	if options.RightMaxWidth > 0 {
		options.RightMaxWidth = max(options.RightMaxWidth, options.RightMinWidth)
		options.RightWidth = min(options.RightWidth, options.RightMaxWidth)
	}
	return &Shell{
		options: options, leftOpen: !options.LeftCollapsed, leftWidth: float32(options.LeftWidth),
		rightOpen: !options.RightCollapsed, rightWidth: float32(options.RightWidth),
	}
}

func (s *Shell) nativeOptions() mygo.WindowOptions {
	o := s.options
	return mygo.WindowOptions{
		Hidden: o.Hidden, Title: o.Title, Width: o.Width, Height: o.Height,
		MinWidth: o.MinWidth, MinHeight: o.MinHeight, StateKey: o.StateKey,
		TitleBarStyle:        mygo.TitleBarHidden,
		HideWindowButtons:    runtime.GOOS == "windows",
		TrafficLightPosition: &mygo.Point{X: 16, Y: 16},
		Transparent:          runtime.GOOS == "darwin",
		Vibrancy:             shellVibrancy(),
		Content:              ui.View(s.view),
	}
}

func shellVibrancy() mygo.Vibrancy {
	switch runtime.GOOS {
	case "darwin":
		return mygo.VibrancySidebar
	case "windows":
		return mygo.VibrancyMica
	default:
		return mygo.VibrancyNone
	}
}

// Window exposes the original MyGo handle for events and lifecycle operations.
// Shell owns its content, title-bar configuration and sidebar layout.
func (s *Shell) Window() *mygo.Window { return s.window }

// SetPage replaces the entire default page, including sidebar controls.
// Nil restores the default page and its previous sidebar expansion state.
// Replacement pages receive an opaque background and the same floating header
// as the default page. The builder owns its top inset; use
// HeaderHeight(c) to keep interactive content below the header.
func (s *Shell) SetPage(page func(*ui.Context)) {
	s.page = page
	s.refresh()
}

// LeftWidth reports the chosen sidebar width in DIPs, independent of viewport
// clipping, collapse or page replacement.
func (s *Shell) LeftWidth() float32 { return s.leftWidth }

// SetLeftWidth restores or changes the chosen width within the configured limits.
// Non-finite values are ignored. This does not change the double-click default.
func (s *Shell) SetLeftWidth(width float32) {
	if math.IsNaN(float64(width)) || math.IsInf(float64(width), 0) {
		return
	}
	s.leftWidth = max(float32(s.options.LeftMinWidth), width)
	if s.options.LeftMaxWidth > 0 {
		s.leftWidth = min(s.leftWidth, float32(s.options.LeftMaxWidth))
	}
	s.refresh()
}

// LeftExpanded reports the requested state, including during an animation.
func (s *Shell) LeftExpanded() bool { return s.leftOpen }

// SetLeftExpanded animates the interior layout without resizing the window.
func (s *Shell) SetLeftExpanded(open bool) {
	s.leftOpen = open
	s.refresh()
}

// RightWidth reports the chosen right sidebar width in DIPs, as LeftWidth.
func (s *Shell) RightWidth() float32 { return s.rightWidth }

// SetRightWidth restores or changes the chosen right sidebar width, as
// SetLeftWidth.
func (s *Shell) SetRightWidth(width float32) {
	if math.IsNaN(float64(width)) || math.IsInf(float64(width), 0) {
		return
	}
	s.rightWidth = max(float32(s.options.RightMinWidth), width)
	if s.options.RightMaxWidth > 0 {
		s.rightWidth = min(s.rightWidth, float32(s.options.RightMaxWidth))
	}
	s.refresh()
}

// RightExpanded reports the right sidebar's requested state, including
// during an animation. It is false without ShellOptions.Right.
func (s *Shell) RightExpanded() bool { return s.rightOpen && s.options.Right != nil }

// SetRightExpanded animates the right sidebar without resizing the window.
func (s *Shell) SetRightExpanded(open bool) {
	s.rightOpen = open
	s.refresh()
}

func (s *Shell) refresh() {
	if s.host != nil {
		s.host.Invalidate()
	}
}

// HeaderHeight returns the floating Shell header's height in DIPs for this
// frame, accounting for native window controls and the active theme's font size.
// Call it from a content builder; Shell does not automatically inset Left,
// Content or pages.
// The header remains above content for window dragging and header controls;
// keep interactive body content below this height.
func HeaderHeight(c *ui.Context) float32 {
	return max(48, c.TitleBar().Height, c.Theme().Rem(3))
}

const sidebarDuration = 220 * time.Millisecond

func openFraction(open bool) float32 {
	if open {
		return 1
	}
	return 0
}

func (s *Shell) view(c *ui.Context) {
	// The theme must be installed before any widget is created.
	if s.options.Theme != nil {
		c.SetTheme(s.options.Theme(c))
	}
	components.WithDialogWindowDrag(c, HeaderHeight(c), func() { s.build(c) })
	toasts := s.options.Toasts
	if toasts.Placement.Top() {
		toasts.Inset += HeaderHeight(c)
	}
	components.Toasts(c, toasts)
}

// SetToasts changes toast placement without recreating the window or toasts.
// Call on the UI thread.
func (s *Shell) SetToasts(options components.ToastOptions) {
	s.options.Toasts = options
	s.refresh()
}

func (s *Shell) build(c *ui.Context) {
	vibrant := c.Vibrancy()
	if vibrant {
		c.Root().Background(ui.Transparent)
	}
	background := c.Theme().Background.Alpha(1)
	root := ui.Column(c).Fill().Clip()
	if !vibrant {
		root.Background(background)
	}
	// Snapshot the page for this frame; callbacks can change it for the next.
	page := s.page
	leftTarget := openFraction(s.leftOpen)
	leftProgress := root.Animate("left", leftTarget, sidebarDuration)
	// Also snap an animation already running when Reduce Motion is enabled.
	if c.Preferences().ReduceMotion {
		leftProgress = leftTarget
	}
	bar := c.TitleBar()
	headerHeight := HeaderHeight(c)
	width, _ := c.Size()
	// Keep the full resize handle reachable at the window's right edge.
	maxLeftWidth := max(float32(0), width-3)
	if s.options.LeftMaxWidth > 0 {
		maxLeftWidth = min(maxLeftWidth, float32(s.options.LeftMaxWidth))
	}
	leftWidth := min(s.leftWidth, maxLeftWidth)
	hasRight := s.options.Right != nil
	rightTarget := openFraction(s.rightOpen && hasRight)
	rightProgress := root.Animate("right", rightTarget, sidebarDuration)
	if c.Preferences().ReduceMotion {
		rightProgress = rightTarget
	}
	// The right sidebar yields to the left one, keeping its resize handle
	// inside the room the left sidebar leaves.
	maxRightWidth := max(float32(0), width-leftWidth*leftProgress-3)
	if s.options.RightMaxWidth > 0 {
		maxRightWidth = min(maxRightWidth, float32(s.options.RightMaxWidth))
	}
	rightWidth := min(s.rightWidth, maxRightWidth)
	rightShown := rightWidth * rightProgress
	if page != nil {
		rightShown = 0
	}
	rightReady := page == nil && s.rightOpen && hasRight && rightProgress == 1
	root.Children(func() {
		ui.Column(c).Key(page != nil).Fill().Clip().Children(func() {
			if page != nil {
				ui.Column(c).Fill().Background(background).Children(func() { page(c) })
				return
			}
			ui.Row(c).Fill().Children(func() {
				if leftProgress > 0 {
					// Only the visible slice changes width. Sidebar content stays put
					// while the opaque main region covers or uncovers it.
					sidebar := ui.Column(c).Key("left").Width(leftWidth * leftProgress).Shrink(0).FillHeight().Clip()
					if !vibrant {
						sidebar.Background(c.Theme().Surface.Alpha(1))
					}
					sidebar.Children(func() {
						ui.Column(c).Width(leftWidth).Shrink(0).FillHeight().
							Disabled(!s.leftOpen).Children(func() {
							ui.Column(c).FillWidth().Grow(1).MinHeight(0).Clip().Children(func() {
								if s.options.Left != nil {
									s.options.Left(c)
								}
							})
							if s.options.LeftToolbar != nil {
								ui.Toolbar(c, func() { s.options.LeftToolbar(c) }).
									FillWidth().Shrink(0).Padding(8)
							}
						})
					})
				}
				ui.Column(c).Key("content").Grow(1).MinWidth(0).FillHeight().Background(background).Clip().Children(func() {
					ui.Column(c).Fill().Children(func() {
						if s.options.Content != nil {
							s.options.Content(c)
						}
					})
					if leftProgress > 0 {
						shellSeparator(c)
					}
					if rightProgress > 0 {
						ui.Box(c).Key("right-separator").Absolute().Right(0).Top(0).Bottom(0).Width(1).Background(c.Theme().Border)
					}
				})
				if rightProgress > 0 {
					// As the left sidebar, only the visible slice changes width.
					right := ui.Column(c).Key("right").Width(rightWidth * rightProgress).Shrink(0).FillHeight().Clip()
					if !vibrant {
						right.Background(c.Theme().Surface.Alpha(1))
					}
					right.Children(func() {
						ui.Column(c).Width(rightWidth).Shrink(0).FillHeight().
							Disabled(!s.rightOpen).Children(func() { s.options.Right(c) })
					})
				}
			})
		})
		// An unpainted header overlays the full-height columns. Only its controls
		// occupy the safe area; builders decide whether to inset their content.
		leftInset := max(8, bar.Left)
		if runtime.GOOS == "darwin" && bar.Left > 0 {
			leftInset += 16
		}
		// On Windows the caption buttons keep the window's top-right corner in
		// their own layer; the header ending there reserves their room.
		rightInset := max(8, bar.Right)
		captions := runtime.GOOS == "windows"
		if captions {
			rightInset += captionWidth(c)
		}
		mainInset := rightInset
		if rightShown > 0 {
			mainInset = max(8, rightInset-rightShown)
		}
		// The right sidebar's toggle ends its header while it is open, and the
		// main header otherwise.
		toggleInRight := page == nil && hasRight && s.rightOpen
		toggleRight := func() {
			if IconButton(c, lucide.PanelRight(), l10n.In(c.Locale()).Text("toggleRightSidebar")).Clicked() {
				s.SetRightExpanded(!s.rightOpen)
			}
		}
		ui.Row(c).Key("header").Absolute().Top(0).Left(0).Right(rightShown).Height(headerHeight).
			AlignItems(ui.Center).Padding(6, mainInset, 6, leftInset).
			DragWindow().Children(func() {
			if page == nil {
				if IconButton(c, lucide.PanelLeft(), l10n.In(c.Locale()).Text("toggleSidebar")).Clicked() {
					s.SetLeftExpanded(!s.leftOpen)
				}
				// Align the title with the main surface, but keep it beyond the
				// leading toggle when the sidebar is narrow or collapsed.
				titleInset := max(8, leftWidth*leftProgress+16-leftInset-max(32, c.Theme().Rem(2)))
				if s.options.LeftHeader != nil && s.leftOpen && leftProgress == 1 && titleInset > 24 {
					ui.Row(c).Key("left-header").Width(titleInset - 24).Shrink(0).FillHeight().
						AlignItems(ui.Center).Clip().Children(func() {
						ui.Spacer(c)
						s.options.LeftHeader(c)
					})
					titleInset = 24
				}
				ui.Row(c).Grow(1).MinWidth(0).FillHeight().Clip().Padding(0, 8, 0, titleInset).Children(func() {
					if s.options.ContentTitle != nil {
						s.options.ContentTitle(c)
					}
				})
				if hasRight && !toggleInRight {
					toggleRight()
				}
			}
		})
		if rightShown > 0 {
			// The header keeps the sidebar's full width at the window's edge
			// while the sidebar slides, so its toggle stays in place.
			ui.Box(c).Key("right-header").Absolute().Top(0).Right(0).Width(rightShown).Height(headerHeight).
				Clip().DragWindow().Children(func() {
				ui.Row(c).Absolute().Top(0).Right(0).Width(rightWidth).Height(headerHeight).
					AlignItems(ui.Center).Gap(4).Padding(6, rightInset, 6, 8).Children(func() {
					if rightReady && s.options.RightHeader != nil {
						ui.Row(c).Key("right-header-content").Grow(1).MinWidth(0).FillHeight().AlignItems(ui.Center).Clip().
							Children(func() { s.options.RightHeader(c) })
					} else {
						ui.Spacer(c)
					}
					if toggleInRight {
						toggleRight()
					}
				})
			})
		}
		if captions {
			ui.Row(c).Key("window-controls-layer").Absolute().Top(0).Right(0).Height(headerHeight).
				AlignItems(ui.Center).Padding(6, max(8, bar.Right), 6, 0).Children(func() { s.captionButtons(c) })
		}
		// The full-height handles win hit testing over the draggable headers.
		if page == nil && s.leftOpen && leftProgress == 1 {
			s.resizeHandle(c, leftWidth, maxLeftWidth)
		}
		if rightReady {
			s.rightResizeHandle(c, width, rightWidth, maxRightWidth)
		}
	})
}

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

// captionWidth is the width of the caption buttons' row.
func captionWidth(c *ui.Context) float32 {
	return 3*max(32, c.Theme().Rem(2)) + 2*4
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
