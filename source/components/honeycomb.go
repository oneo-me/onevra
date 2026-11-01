package components

import (
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

// HoneycombItem describes one round cell of a Honeycomb.
type HoneycombItem struct {
	// ID is returned when the item is chosen and matched against
	// HoneycombOptions.Selected. Keep IDs stable and unique.
	ID, Label string
	// Image draws a decoded bitmap or SVG centered in the cell, replaced by
	// Content when that is set.
	Image ui.ImageSource
	// Content, when set, builds the item at the given size instead of drawing
	// Image, such as a generated monogram. It fills the cell's circle itself:
	// the cell does not clip it, so a rectangular mark would show its corners.
	// A clipped cell would take a mask the GPU makes anew every frame as the
	// honeycomb turns it, which costs the frame.
	Content func(c *ui.Context, size float32)
	// Overlay, when set, builds content over the item and its ring at the
	// given size, such as a badge on its edge. Like Content it is not clipped.
	Overlay func(c *ui.Context, size float32)
	// Ring colors the ring the item nearest the middle of the view wears, such
	// as the main color of its image; transparent uses the theme's accent.
	Ring ui.Color
	// Caption, when set, shows under the item while it is nearest the middle
	// of the view, such as its name, fading in and out with its ring. It lies
	// on a glass label over the items below.
	Caption string
	// Menu, when set, fills the item's context menu, which opens where the
	// item is right-clicked, without choosing it, and, while it has the
	// focus, from the menu key or Shift+F10.
	Menu func(m *ui.Menu)
}

// HoneycombOptions configures a Honeycomb.
type HoneycombOptions struct {
	// Items keep their order from frame to frame: each takes the cell of its
	// position, so a reordered list moves them around. They slide to their new
	// cells, fading in and out, as items are added and removed.
	Items []HoneycombItem
	// Selected is the chosen item's ID, which the honeycomb pans to center.
	Selected string
	// SelectedLabel names the selected item for assistive technology and
	// tests; the others are named by their own labels.
	SelectedLabel string
	// OnPress runs as the honeycomb is built after the primary button goes
	// down on it. The honeycomb takes the press for its pan and choice, so
	// it does not reach elements around it, such as a background that
	// ripples where the pointer presses.
	OnPress func()
	// Curvature scales how sharply the ball the honeycomb lies on curves:
	// how far items further out turn away, how much smaller they are than the
	// one in the middle, and how far in from the edge of the view they shrink
	// and fade out. Zero uses 1. Negative values lay the honeycomb flat, with
	// every item at the same size, to show many at once; values above 1 curve
	// it more, up to 2.
	Curvature float32
	// OffsetY shifts the honeycomb, and the point it settles and chooses
	// around, up or down within its view in DIPs: negative moves it up. Use it
	// when content floating over one edge leaves less room there, such as a
	// composer at the bottom, so the honeycomb centers in the visible space.
	// The view and its edge fade stay where they are.
	OffsetY float32
	// Gap is the space between neighboring items in DIPs, unmagnified, which
	// the lens magnifies with the items, scaled to the theme's spacing like
	// the rest of the honeycomb. Zero uses 10; negative values set the items
	// edge to edge; values above 56 are 56, the size of an item.
	Gap float32
}

const (
	// The honeycomb is magnified as through a lens: an item is honeycombLarge
	// in the middle of the view, shrinking smoothly towards honeycombTiny
	// further out, over about honeycombFall of the honeycomb, and the space
	// between items shrinks with them. honeycombSmall is an item's size
	// unmagnified, which every item has on a flat honeycomb; a curvature
	// below 1 weakens the lens towards it, and one above 1 shortens the fall.
	honeycombLarge = 84
	honeycombSmall = 56
	honeycombTiny  = 14
	honeycombFall  = 150
	// honeycombRing is the width of the accent ring round the item nearest the
	// middle of the view, drawn at honeycombRingOpacity so the item shows
	// through and the ring tints its edge instead of covering it.
	honeycombRing        = 2
	honeycombRingOpacity = .5
	// The caption of the item nearest the middle hangs honeycombCaptionGap
	// below it, at most honeycombCaptionWidth wide.
	honeycombCaptionGap   = 6
	honeycombCaptionWidth = 200
	// honeycombGap is the default space between neighboring items,
	// unmagnified, and honeycombGapMax the largest HoneycombOptions.Gap.
	honeycombGap    = 10
	honeycombGapMax = honeycombSmall
	// The view fills the space it is given; honeycombWidth and honeycombHeight
	// are its size before it is laid out.
	honeycombWidth  = 400
	honeycombHeight = 240
	// honeycombTap is how far the pointer may move while pressed for its
	// release to still click the item under it.
	honeycombTap = 4
	// honeycombInertia is how long a released drag goes on at its speed, as it
	// eases to a stop.
	honeycombInertia = .3
	// honeycombRubber is how much of the pointer's travel moves the items past
	// the edge of the honeycomb.
	honeycombRubber = .35
	// The honeycomb sits on the surface of a ball behind it: an item in the
	// middle faces the viewer, and one honeycombTiltFall or so out turns away
	// in the direction of its distance from the middle, up to honeycombTilt
	// degrees at the far edge. A larger honeycombTiltFall makes a larger,
	// gentler ball; this one covers the whole view, so its surface spans it.
	// HoneycombOptions.Curvature scales the tilt, which stays below
	// honeycombTiltMax so items edge-on to the viewer do not vanish.
	honeycombTilt     = 60
	honeycombTiltFall = 420
	honeycombTiltMax  = 80
	// honeycombCurvatureMax is the largest HoneycombOptions.Curvature.
	honeycombCurvatureMax = 2
)

var (
	honeycombDuration = 420 * time.Millisecond
	// honeycombSettle is how long scrolling must stop before the honeycomb
	// settles on an item.
	honeycombSettle = 150 * time.Millisecond
)

// honeycombMetrics are the lengths of a honeycomb in DIPs, scaled to the frame
// theme. spacing is the distance between the centers of neighboring cells,
// unmagnified. Within edge of the edge of the view, items shrink further, to
// honeycombTiny, and fade out; minHeight keeps the selected item and the rows
// beside it in view when space is short.
type honeycombMetrics struct {
	large, small, tiny, fall, tiltFall, spacing, ring, edge, minHeight, width, height float32
	captionGap, captionWidth                                                          float32
}

// honeycombMetricsFor scales one design DIP to the theme's spacing, so a
// roomier theme makes the whole honeycomb roomier too, with gap between items
// as HoneycombOptions.Gap gives it.
func honeycombMetricsFor(t *ui.Theme, gap float32) honeycombMetrics {
	s := t.Space(.25)
	spacing := honeycombSmall + honeycombGapFor(gap)
	return honeycombMetrics{
		large:        honeycombLarge * s,
		small:        honeycombSmall * s,
		tiny:         honeycombTiny * s,
		fall:         honeycombFall * s,
		tiltFall:     honeycombTiltFall * s,
		spacing:      spacing * s,
		ring:         honeycombRing * s,
		edge:         spacing * s,
		minHeight:    2 * spacing * s,
		width:        honeycombWidth * s,
		height:       honeycombHeight * s,
		captionGap:   honeycombCaptionGap * s,
		captionWidth: honeycombCaptionWidth * s,
	}
}

// Honeycomb draws the items as a honeycomb, as on a watch face, in a view that
// fills the width and grows into the height of its column. The honeycomb is
// magnified most in the middle of the view, shifted by OffsetY, so items and
// the space between them shrink smoothly the further out they are, and shrink
// further and fade out as they near the edge of the view. The honeycomb lies
// on the surface of a ball behind it, so each item turns away from the viewer
// the further out it is, and dragging it pans over that ball. The item nearest
// the middle wears a ring in its own color or the theme's accent, and its
// caption under it.
//
// Dragging the items, or scrolling over them, pans the honeycomb, which then
// settles with the item nearest the middle centered and chosen; choosing an
// item, here or elsewhere, pans it over to center that item. The keyboard
// reaches every item and pans to the focused one. Honeycomb returns the ID of
// an item chosen by input in this frame, or "".
//
// Applications own the items, their decoded images, labels, selected ID and any
// action a choice triggers. Items without a Content or Image builder draw
// nothing. Call Honeycomb within a bounded parent; the view takes its height
// from the space left to it, keeping at least two rows visible.
func Honeycomb(c *ui.Context, options HoneycombOptions) string {
	m := honeycombMetricsFor(c.Theme(), options.Gap)
	curvature := honeycombCurvature(options.Curvature)
	view := ui.Box(c).FillWidth().MinWidth(0).Height(m.height).Grow(1).Shrink(1).MinHeight(m.minHeight).Clip()
	n := len(options.Items)
	selected := slices.IndexFunc(options.Items, func(item HoneycombItem) bool { return item.ID == options.Selected })
	if selected < 0 {
		return ""
	}
	p := ui.Local(view, honeycombPanKey{}, func() honeycombPan { return honeycombPan{} })
	now := c.Now()
	// A new spacing, from the theme or the gap, scales the whole honeycomb at
	// once, the pan and any glide with it, so it stays where it was; items
	// added or removed slide the others to their cells as the pan glides to
	// the selected one.
	rescaled := p.spacing != m.spacing
	if len(p.cells) != n || rescaled {
		k := m.spacing / p.spacing
		p.cells, p.spacing = honeycombCells(n, m.spacing), m.spacing
		if p.ready && rescaled {
			p.at, p.from, p.to = p.at.scaled(k), p.from.scaled(k), p.to.scaled(k)
		} else if p.ready && !p.pressed {
			p.glide(p.cells[selected], now, honeycombDuration, ui.EaseInOut)
		}
	}
	p.lo, p.hi = honeycombBounds(p.cells)
	reduce := c.Preferences().ReduceMotion || c.Instant()
	chosen := ""
	if options.Selected != p.selected {
		p.selected = options.Selected
		if p.ready {
			p.glide(p.cells[selected], now, honeycombDuration, ui.EaseInOut)
		} else {
			p.at = p.cells[selected]
		}
	}
	p.ready = true
	p.curvature = curvature
	if p.pressedNow {
		p.pressedNow = false
		if options.OnPress != nil {
			options.OnPress()
		}
	}
	if p.clicked != "" {
		if i := slices.IndexFunc(options.Items, func(item HoneycombItem) bool { return item.ID == p.clicked }); i >= 0 {
			chosen = p.choose(options.Items[i].ID, p.cells[i], now)
		}
		p.clicked = ""
	}
	if p.released {
		// A drag let go goes on at its speed and eases to a stop on the item
		// nearest where it would come to rest.
		p.released, p.scrolling = false, false
		i := p.nearest(honeycombPoint{p.at.x + p.vx*honeycombInertia, p.at.y + p.vy*honeycombInertia})
		if id := p.choose(options.Items[i].ID, p.cells[i], now); id != "" {
			chosen = id
		}
	}
	if p.scrolling && !p.pressed {
		// Scrolling settles on the item nearest the middle once it has stopped.
		if wait := honeycombSettle - now.Sub(p.scrolled); wait > 0 {
			c.After(wait)
		} else {
			p.scrolling = false
			i := p.nearest(p.at)
			if id := p.choose(options.Items[i].ID, p.cells[i], now); id != "" {
				chosen = id
			}
		}
	}
	if p.gliding {
		t := float32(now.Sub(p.start)) / float32(p.duration)
		if t >= 1 || reduce {
			p.at, p.gliding = p.to, false
		} else {
			e := p.ease(max(t, 0))
			p.at = honeycombPoint{p.from.x + (p.to.x-p.from.x)*e, p.from.y + (p.to.y-p.from.y)*e}
			c.AnimationFrame()
		}
	}
	if n > 1 {
		view.HandleInput(p.input)
		if p.pressed && p.travel >= honeycombTap {
			view.Cursor(ui.CursorGrabbing)
		}
	}
	// The view's size is the last frame's layout, which the first frame lacks:
	// it hides the items rather than show them fading at a guessed edge, and
	// builds again once laid out. The items hang from the middle of the view,
	// so a resize keeps them centered; their fading at its edge follows a frame
	// later.
	width, height := m.width, m.height
	if r := view.LayoutBounds(); r.W > 0 {
		width, height = r.W, r.H
	} else {
		view.Invisible()
		c.AnimationFrame()
	}
	centerX, centerY := width/2, height/2+options.OffsetY
	view.Draw(func(_ *ui.Painter, r ui.Rect) {
		if r.W != width || r.H != height {
			c.Invalidate()
		}
	})
	// Within the edge inset, a cell's spacing, scaled by the curvature as items
	// further round the ball fall away sooner, items shrink and fade out; the
	// inset is less in a view too small for it, and none on a flat honeycomb.
	edge := min(m.edge*curvature, width/2, height/2)
	middle := p.nearest(p.at)
	th := c.Theme()
	accent := th.Accent
	p.hits = p.hits[:0]
	var captions []honeycombCaption
	view.Children(func() {
		// Items hang from the middle of the view, offset by OffsetY.
		ui.Box(c).Absolute().LeftPercent(50).TopPercent(50).Size(0, 0).Children(func() {
			for i, item := range options.Items {
				// Every item keeps its button, faded out or not, so the keyboard
				// reaches each one and pans the honeycomb to it. Items fade in
				// and out as they are added and removed.
				b := ui.ButtonBase(c).Key(item.ID).Background(ui.Color{})
				b.Transition(ui.ElementTransition{Duration: honeycombDuration, Ease: ui.EaseInOut, Colors: true,
					Enter: &ui.Motion{Scale: .5}, Exit: &ui.Motion{Scale: .5}})
				// The item slides to its cell when items before it come or go.
				cell := p.cells[i]
				if rescaled {
					cell.x, cell.y = b.Jump(honeycombCellKey{0}, cell.x), b.Jump(honeycombCellKey{1}, cell.y)
				} else {
					cell.x = b.AnimateWith(honeycombCellKey{0}, cell.x, honeycombDuration, ui.EaseInOut)
					cell.y = b.AnimateWith(honeycombCellKey{1}, cell.y, honeycombDuration, ui.EaseInOut)
				}
				dx, dy := cell.x-p.at.x, cell.y-p.at.y
				d := float32(math.Hypot(float64(dx), float64(dy)))
				cx, cy := centerX, centerY
				if d > 0 {
					f := m.magnified(d, curvature) / d
					cx, cy = cx+dx*f, cy+dy*f
				}
				// How far into the inset along the edge, rounded at the corners as
				// on a watch face, from 0 inside it to 1 at the edge.
				ox, oy := max(edge-cx, cx-width+edge, 0), max(edge-cy, cy-height+edge, 0)
				out := min(float32(math.Hypot(float64(ox), float64(oy)))/max(edge, 1), 1)
				size := m.tiny + (m.size(d, curvature)-m.tiny)*(1-out)
				opacity := 1 - honeycombRamp(out, .35, .85)
				label := item.Label
				if i == selected && options.SelectedLabel != "" {
					label = options.SelectedLabel
				}
				b.Label(label)
				// Round cells, as on a watch face, but not clipped: a rounded clip
				// under the ball's transform would take a mask the GPU makes anew
				// every frame. Items draw their own circle instead.
				b.Absolute().Left(cx-width/2-size/2).Top(cy-height/2-size/2).Size(size, size)
				// Turn the item on the ball's surface, so dragging pans over it.
				b.Transform(honeycombBall(dx, dy, m.tiltAt(d, curvature)))
				// The ring passes from item to item as the honeycomb pans. It is
				// painted over the item rather than set as a border, which would
				// take room within the cell and push the item off center.
				ringTarget := float32(0)
				if i == middle {
					ringTarget = 1
				}
				ring := b.AnimateWith(honeycombRingKey{}, ringTarget, honeycombDuration/2, ui.EaseInOut)
				color := accent
				if item.Ring.A > 0 {
					color = item.Ring
				}
				// The ring is a layer of its own between the item and its
				// overlay, so an overlay can sit on top of it.
				drawRing := func() {
					ui.Box(c).Absolute().Left(0).Top(0).Size(size, size).Draw(func(p *ui.Painter, r ui.Rect) {
						if col := color.Alpha(ring * honeycombRingOpacity); col.A > 0 {
							p.Stroke(r, col, r.W/2, m.ring)
						}
					})
				}
				if i != selected && (b.Hovered() || b.Focused()) {
					opacity = min(1, opacity+.25)
				}
				b.Opacity(opacity)
				if item.Caption != "" && ring > 0 {
					captions = append(captions, honeycombCaption{item.ID, item.Caption, cx - width/2, cy - height/2 + size/2, ring * opacity})
				}
				if item.Menu != nil {
					b.ContextMenu(item.Menu)
				}
				if opacity > .15 {
					p.hits = append(p.hits, honeycombHit{item.ID, ui.Rect{X: cx - size/2, Y: cy - size/2, W: size, H: size}})
				}
				content, image, overlay := item.Content, item.Image, item.Overlay
				b.Children(func() {
					if content != nil {
						content(c, size)
					} else {
						Image(c, image).Absolute().Left(0).Top(0).Size(size, size).Fit(ui.Cover).Radius(size / 2)
					}
					drawRing()
					if overlay != nil {
						overlay(c, size)
					}
				})
				if b.Clicked() {
					// The keyboard or assistive technology; the pointer is the
					// view's.
					chosen = p.choose(item.ID, p.cells[i], now)
				}
				if b.FocusVisible() {
					if p.focus != item.ID {
						p.focus = item.ID
						p.glide(p.cells[i], now, honeycombDuration, ui.EaseInOut)
					}
				} else if p.focus == item.ID {
					// Leaving the items pans back to the chosen one, unless another
					// item has taken focus already.
					p.focus = ""
					p.glide(p.cells[selected], now, honeycombDuration, ui.EaseInOut)
				}
			}
			// Captions follow the items, so they lie over the items below
			// theirs; the pointer passes through them to the items. Each
			// centers its label in a row of the widest a label may be.
			for _, caption := range captions {
				ui.Row(c).Key(honeycombCaptionKey{caption.id}).Absolute().
					Left(caption.x - m.captionWidth/2).Top(caption.y + m.captionGap).Width(m.captionWidth).
					Justify(ui.Center).Opacity(caption.opacity).PassThrough().Children(func() {
					ui.Row(c).MinWidth(0).Padding(th.Space(.5), th.Space(2)).Radius(th.Space(3)).
						Material(glass.Glass{Tint: th.Background.Alpha(.92)}).Border(1, th.Border).
						Shadow(0, th.Space(.5), th.Space(2), 0, ui.RGBA(0, 0, 0, .1)).
						TextColor(th.Text).FontSize(th.FontSize - 1).FontWeight(500).PassThrough().Children(func() {
						Text(c, caption.text).SingleLine().Ellipsis("…").MinWidth(0).Unselectable().PassThrough()
					})
				})
			}
		})
	})
	return chosen
}

// honeycombPoint is a point of the honeycomb, from its first cell.
type honeycombPoint struct{ x, y float32 }

// scaled returns the point k times as far from the first cell.
func (p honeycombPoint) scaled(k float32) honeycombPoint { return honeycombPoint{p.x * k, p.y * k} }

// honeycombPanKey keys the retained pan of a Honeycomb.
type honeycombPanKey struct{}

// honeycombRingKey keys the animation of an item's accent ring.
type honeycombRingKey struct{}

// honeycombCellKey keys the animation of an item's cell, 0 along x and 1
// along y.
type honeycombCellKey struct{ axis int }

// honeycombCaptionKey keys the caption of the item with the ID.
type honeycombCaptionKey struct{ id string }

// honeycombCaption is a caption to show under the item whose bottom middle
// is at x, y from the middle of the view.
type honeycombCaption struct {
	id, text string
	x, y     float32
	opacity  float32
}

// honeycombHit is where an item was drawn in the last frame, for the pointer.
type honeycombHit struct {
	id   string
	rect ui.Rect
}

// honeycombPan is the point of a Honeycomb in the middle of its view, the pan
// it eases along and the pointer pressing it.
type honeycombPan struct {
	cells   []honeycombPoint
	spacing float32
	lo, hi  honeycombPoint
	at      honeycombPoint
	ready   bool
	// selected is the selection last panned to; focus the item the keyboard
	// last panned to.
	selected, focus string

	gliding        bool
	from, to       honeycombPoint
	start          time.Time
	duration       time.Duration
	ease           ui.Easing
	pressed        bool
	pressedNow     bool
	released       bool
	lastX, lastY   float32
	travel, vx, vy float32
	moved          time.Time
	scrolling      bool
	scrolled       time.Time
	hits           []honeycombHit
	// curvature is the last frame's, which sets how far a drag pans.
	curvature float32
	clicked   string
}

// glide eases the pan to the point to, unless it is going there already.
func (p *honeycombPan) glide(to honeycombPoint, now time.Time, d time.Duration, ease ui.Easing) {
	if p.gliding && p.to == to {
		return
	}
	p.gliding, p.from, p.to, p.start, p.duration, p.ease = true, p.at, to, now, d, ease
}

// choose pans to the item with id at cell and returns id, unless it is selected
// already; it pans back to it then.
func (p *honeycombPan) choose(id string, cell honeycombPoint, now time.Time) string {
	p.glide(cell, now, honeycombDuration, ui.EaseInOut)
	if id == p.selected {
		return ""
	}
	// The pan goes on as the selection follows.
	p.selected = id
	return id
}

// input takes the pointer pressing, dragging and scrolling over the view.
func (p *honeycombPan) input(ev ui.InputEvent) bool {
	switch ev.Kind {
	case ui.InputPointerDown:
		if ev.Button != 0 {
			return false
		}
		// A press catches a moving honeycomb.
		p.pressed, p.gliding = true, false
		p.pressedNow = true
		p.lastX, p.lastY, p.travel, p.vx, p.vy, p.moved = ev.X, ev.Y, 0, 0, 0, ev.Time
	case ui.InputPointerMove:
		if !p.pressed {
			return false
		}
		dx, dy := ev.X-p.lastX, ev.Y-p.lastY
		p.lastX, p.lastY = ev.X, ev.Y
		p.travel += float32(math.Hypot(float64(dx), float64(dy)))
		// The magnified middle follows the pointer.
		k := p.drag()
		p.pan(-dx*k, -dy*k)
		if dt := float32(ev.Time.Sub(p.moved).Seconds()); dt > 0 {
			// The speed of the last moves, smoothed.
			p.vx = .7*(-dx/dt) + .3*p.vx
			p.vy = .7*(-dy/dt) + .3*p.vy
		}
		p.moved = ev.Time
	case ui.InputPointerUp:
		if !p.pressed || ev.Button != 0 {
			return false
		}
		p.pressed = false
		if p.travel < honeycombTap {
			p.clicked = p.hit(ev.X, ev.Y)
			return true
		}
		if ev.Time.Sub(p.moved) > 80*time.Millisecond {
			// Held still before letting go.
			p.vx, p.vy = 0, 0
		}
		p.released = true
	case ui.InputScroll:
		p.gliding, p.scrolling = false, true
		k := p.drag()
		p.pan(ev.DX*k, ev.DY*k)
		p.scrolled = ev.Time
	default:
		return false
	}
	return true
}

// drag is how far the pan moves for each DIP the pointer travels, so the
// middle of the view, magnified by the curvature, follows the pointer.
func (p *honeycombPan) drag() float32 {
	return 1 / (1 + (honeycombLarge/float32(honeycombSmall)-1)*min(p.curvature, 1))
}

// pan moves the pan by dx, dy, less past the edge of the honeycomb.
func (p *honeycombPan) pan(dx, dy float32) {
	past := func(v, lo, hi, d float32) float32 {
		if (v < lo && d < 0) || (v > hi && d > 0) {
			return d * honeycombRubber
		}
		return d
	}
	p.at.x += past(p.at.x, p.lo.x, p.hi.x, dx)
	p.at.y += past(p.at.y, p.lo.y, p.hi.y, dy)
}

// nearest returns the index of the cell nearest to the point at.
func (p *honeycombPan) nearest(at honeycombPoint) int {
	best, far := 0, math.Inf(1)
	for i, c := range p.cells {
		if d := math.Hypot(float64(c.x-at.x), float64(c.y-at.y)); d < far {
			best, far = i, d
		}
	}
	return best
}

// hit returns the ID of the round item drawn at x, y in the last frame, the
// nearest if several are.
func (p *honeycombPan) hit(x, y float32) string {
	id, best := "", float32(math.Inf(1))
	for _, h := range p.hits {
		d := float32(math.Hypot(float64(x-h.rect.X-h.rect.W/2), float64(y-h.rect.Y-h.rect.H/2)))
		if d <= h.rect.W/2 && d < best {
			id, best = h.id, d
		}
	}
	return id
}

// honeycombCells returns the centers of n cells of a honeycomb with rows of
// cells, filling outwards from the first in a circle.
func honeycombCells(n int, spacing float32) []honeycombPoint {
	type cell struct {
		p           honeycombPoint
		dist, angle float64
	}
	var cells []cell
	for rings := 1; len(cells) < n; rings *= 2 {
		cells = cells[:0]
		for row := -rings; row <= rings; row++ {
			for col := -rings - 1; col <= rings+1; col++ {
				// Every other row is offset by half a cell.
				x := (float64(col) + float64(row&1)/2) * float64(spacing)
				y := float64(row) * float64(spacing) * math.Sqrt(3) / 2
				angle := math.Atan2(y, x)
				if angle < 0 {
					angle += 2 * math.Pi
				}
				cells = append(cells, cell{honeycombPoint{float32(x), float32(y)}, math.Hypot(x, y), angle})
			}
		}
		// Only cells nearer than the nearest beyond the rings are sure to be
		// among the nearest.
		reach := float64(rings) * float64(spacing) * math.Sqrt(3) / 2
		cells = slices.DeleteFunc(cells, func(c cell) bool { return c.dist > reach })
	}
	slices.SortStableFunc(cells, func(a, b cell) int {
		if d := a.dist - b.dist; math.Abs(d) > 1e-6 {
			return int(math.Copysign(1, d))
		}
		if a.angle < b.angle {
			return -1
		}
		if a.angle > b.angle {
			return 1
		}
		return 0
	})
	points := make([]honeycombPoint, n)
	for i := range points {
		points[i] = cells[i].p
	}
	return points
}

// honeycombBounds returns the corners of the rectangle around the cells'
// centers, within which the pan stays.
func honeycombBounds(cells []honeycombPoint) (lo, hi honeycombPoint) {
	for i, p := range cells {
		if i == 0 {
			lo, hi = p, p
		}
		lo = honeycombPoint{min(lo.x, p.x), min(lo.y, p.y)}
		hi = honeycombPoint{max(hi.x, p.x), max(hi.y, p.y)}
	}
	return lo, hi
}

// size is the size of an item d from the middle of the view, unmagnified, as
// the lens of a honeycomb of the given curvature magnifies it.
func (m honeycombMetrics) size(d, curvature float32) float32 {
	return m.lens(curvature, m.small, func(fall float32) float32 {
		return m.tiny + (m.large-m.tiny)*float32(math.Exp(float64(-d/fall)))
	})
}

// tiltAt is how far, in degrees, an item d from the middle of the view turns
// away from the viewer: it faces the viewer in the middle and turns up to
// honeycombTilt, scaled by curvature, as it nears the edge, as the surface of
// a ball does.
func (m honeycombMetrics) tiltAt(d, curvature float32) float32 {
	return min(honeycombTilt*curvature*(1-float32(math.Exp(float64(-d/m.tiltFall)))), honeycombTiltMax)
}

// honeycombCurvature resolves HoneycombOptions.Curvature: zero is 1, negative
// is flat and the rest is at most honeycombCurvatureMax.
func honeycombCurvature(v float32) float32 {
	switch {
	case v == 0:
		return 1
	case v < 0 || math.IsNaN(float64(v)):
		return 0
	}
	return min(v, honeycombCurvatureMax)
}

// honeycombGapFor resolves HoneycombOptions.Gap: zero is honeycombGap,
// negative is none and the rest is at most honeycombGapMax.
func honeycombGapFor(v float32) float32 {
	switch {
	case v == 0:
		return honeycombGap
	case v < 0 || math.IsNaN(float64(v)):
		return 0
	}
	return min(v, honeycombGapMax)
}

// magnified is how far from the middle of the view a point d from it,
// unmagnified, is drawn: the space up to it, magnified as the items there.
func (m honeycombMetrics) magnified(d, curvature float32) float32 {
	return m.lens(curvature, d, func(fall float32) float32 {
		e := float32(math.Exp(float64(-d / fall)))
		return (m.tiny*d + (m.large-m.tiny)*fall*(1-e)) / m.small
	})
}

// lens applies the curvature to a measure of the lens, given as a function of
// its fall: below 1 the lens weakens towards flat, the measure unmagnified,
// and above 1 it falls off sooner.
func (m honeycombMetrics) lens(curvature, flat float32, measure func(fall float32) float32) float32 {
	if curvature >= 1 {
		return measure(m.fall / curvature)
	}
	return flat + (measure(m.fall)-flat)*curvature
}

// honeycombBall turns an item at dx, dy from the middle of the view to lie on
// the surface of a ball there: it faces the viewer in the middle and is squashed
// along the direction of its distance as it turns away, by tilt degrees at its
// edge. The squash is affine, not a perspective, so the renderer antialiases it
// from screen derivatives instead of integrating coverage at sample points,
// which a screenful of items cannot afford.
func honeycombBall(dx, dy, tilt float32) ui.Transform {
	if tilt <= 0 || dx == 0 && dy == 0 {
		return ui.Transform{}
	}
	// Squash along the radial direction: turn it to the horizontal axis, scale
	// it there, and turn it back.
	angle := float32(math.Atan2(float64(dy), float64(dx)) * 180 / math.Pi)
	cos := float32(math.Cos(float64(tilt) * math.Pi / 180))
	return ui.Rotation(-angle).Then(ui.Scaling(cos, 1)).Then(ui.Rotation(angle))
}

// honeycombRamp goes from 0 at from to 1 at to, smoothly.
func honeycombRamp(v, from, to float32) float32 {
	t := min(max((v-from)/(to-from), 0), 1)
	return t * t * (3 - 2*t)
}
