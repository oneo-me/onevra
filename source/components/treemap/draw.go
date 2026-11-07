package treemap

import (
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
)

// input records primary clicks for the next build, and the node a context
// menu is asked for. It takes nothing, so pressing still focuses the view
// and the menu still opens.
func (v *view) input(ev ui.InputEvent) bool {
	switch ev.Kind {
	case ui.InputPointerDown:
		switch ev.Button {
		case 0:
			v.pressed, v.pressX, v.pressY = true, ev.X, ev.Y
		case 1:
			v.menuID = ""
			if i := v.hit(ev.X, ev.Y); i >= 0 && !v.zooming {
				v.menuID = v.tiles[i].id
			}
		}
	case ui.InputKeyDown:
		if ev.Key == ui.KeyContextMenu && ev.Mods == 0 || ev.Key == ui.KeyF10 && ev.Mods == ui.Shift {
			v.menuID = v.activeID
		}
	case ui.InputPointerUp:
		if v.pressed && ev.Button == 0 {
			v.pressed = false
			if !ev.Canceled && math.Hypot(float64(ev.X-v.pressX), float64(ev.Y-v.pressY)) < tapSlop {
				v.clicks = append(v.clicks, click{v.pressX, v.pressY})
			}
		}
	}
	return false
}

// hoverFade is how long tile highlights take to cross-fade, at most the
// view's transition duration.
const hoverFade = 160 * time.Millisecond

// draw paints the tiles at the canvas r, through a drill or data transition
// while one runs.
func (v *view) draw(p *ui.Painter, r ui.Rect, t *ui.Theme, active int, focused bool) {
	if v.zooming {
		e := float32(p.Now().Sub(v.zoomStart)) / float32(v.duration)
		if e < 1 {
			p.AnimationFrame()
			for _, z := range v.zoomFrame(ui.EaseInOut(max(e, 0))) {
				b := offset(z.b, r)
				v.fill(p, t, z.tl, b, z.tl.amount, z.radius, z.opacity, 0)
				v.labels(p, t, z.tl, b, z.radius, z.text, true)
			}
			return
		}
		v.zooming, v.old = false, nil
	}
	e := v.morphProgress(p.Now())
	if v.morphing {
		if e < 1 {
			p.AnimationFrame()
		} else {
			v.morphing = false
		}
	}
	// The highlight moves from the tile the pointer left to the one it is on.
	h := float32(1)
	if v.duration > 0 {
		if h = float32(p.Now().Sub(v.hoverStart)) / float32(min(hoverFade, v.duration)); h < 1 {
			h = ui.EaseInOut(max(h, 0))
			p.AnimationFrame()
		} else {
			h = 1
		}
	}
	lift := func(id string) float32 {
		switch id {
		case "":
			return 0
		case v.hoverID:
			return h
		case v.hoverOut:
			return 1 - h
		}
		return 0
	}
	for _, tl := range v.tiles {
		b := offset(lerp(tl.from, tl.rect, e), r)
		if b.W < .5 || b.H < .5 {
			continue
		}
		l := lift(tl.id)
		radius := v.radius(b, tl.depth)
		v.fill(p, t, tl, b, tl.fromAmount+(tl.amount-tl.fromAmount)*e, radius, 1, l)
		// Labels are fitted to where the tile settles; keep them inside it on
		// the way.
		v.labels(p, t, tl, b, radius, 1, e < 1 && tl.from != tl.rect)
	}
	outline := func(i int, width float32, c ui.Color) {
		tl := v.tiles[i]
		b := offset(inset(lerp(tl.from, tl.rect, e), -.5, -.5), r)
		radius := v.radius(b, tl.depth)
		if c == (ui.Color{}) {
			p.FocusRing(b, [4]float32{radius, radius, radius, radius})
			return
		}
		p.Stroke(b, c, radius, width)
	}
	// The shown level's tile under the pointer is what a click drills into;
	// as in charts, a ring in its own color marks it.
	for _, id := range [...]string{v.hoverOut, v.hoverID} {
		if i := v.index(id); i >= 0 {
			if top := v.topOf(i); top != active && !(top == i && v.tiles[i].hover.bordered()) {
				outline(top, 2, v.tiles[top].color.Alpha(lift(id)))
			}
		}
	}
	if active >= 0 && active < len(v.tiles) {
		if focused {
			outline(active, 0, ui.Color{})
		} else {
			outline(active, 2, v.tiles[active].color.Mix(t.Text, .2))
		}
	}
}

// fill paints a tile's background and border at b, in window coordinates,
// tinted by amount and deepened by h of the hover highlight.
func (v *view) fill(p *ui.Painter, t *ui.Theme, tl tile, b ui.Rect, amount, radius, opacity, h float32) {
	if b.W < .5 || b.H < .5 || opacity <= 0 {
		return
	}
	amount += .08 * h
	// A soft diagonal gradient, lighter at the top left, with a faint edge in
	// the same hue, unless the tile's style replaces them.
	from, to := t.Background.Mix(tl.color, amount*.6), t.Background.Mix(tl.color, min(amount*1.3, .7))
	if bg := tl.style.Background; bg != (ui.Color{}) {
		from, to = bg.Mix(tl.color, .08*h), bg.Mix(tl.color, .08*h)
	}
	if bg := tl.hover.Background; bg != (ui.Color{}) {
		from, to = from.Mix(bg, h), to.Mix(bg, h)
	}
	p.FillGradient(b, ui.LinearGradient{From: from.Alpha(opacity), To: to.Alpha(opacity), Angle: 135}, radius)
	edge, width, dashed := t.Background.Mix(tl.color, min(amount+.2, .8)).Alpha(.5), float32(1), false
	for k, s := range [...]Style{tl.style, tl.hover} {
		w := float32(1)
		if k == 1 {
			w = h
		}
		if s.Border != (ui.Color{}) {
			edge = edge.Mix(s.Border, w)
		}
		if s.BorderWidth > 0 {
			width += (s.BorderWidth - width) * w
		}
		if s.BorderStyle == ui.BorderDashed && w >= .5 {
			dashed = true
		}
	}
	if dashed {
		p.StrokeDashed(b, edge.Alpha(opacity), radius, width)
	} else {
		p.Stroke(b, edge.Alpha(opacity), radius, width)
	}
}

// labels paints a tile's texts at b at opacity, clipped to the tile when
// clip is set.
func (v *view) labels(p *ui.Painter, t *ui.Theme, tl tile, b ui.Rect, radius, opacity float32, clip bool) {
	if (tl.label == "" && tl.big == "") || b.W < .5 || b.H < .5 || opacity <= 0 {
		return
	}
	m := v.metrics
	text, muted := t.Text.Alpha(opacity), t.TextMuted.Alpha(opacity)
	draw := func() {
		x, y := b.X+m.padX, b.Y+m.padY
		if tl.label != "" {
			p.RichText(x, y, 0, ui.Span{Text: tl.label, Size: m.small, Weight: labelWeight, Color: text})
		}
		if tl.value != "" {
			p.RichText(b.X+b.W-m.padX-tl.valueWidth, y, 0, ui.Span{Text: tl.value, Size: m.small, Weight: labelWeight, Color: muted})
		}
		if tl.big != "" {
			y += m.line + m.padY
			p.RichText(x, y, 0, ui.Span{Text: tl.big, Size: m.big, Weight: bigWeight, Color: text})
			if tl.caption != "" {
				p.RichText(x, y+m.bigLine, 0, ui.Span{Text: tl.caption, Size: m.small, Weight: labelWeight, Color: muted})
			}
		}
	}
	if clip {
		p.Clip(b, radius, draw)
	} else {
		draw()
	}
}

// zoomed is a tile e of the way through a drill, at b in the canvas.
type zoomed struct {
	tl                    tile
	b                     ui.Rect
	radius, opacity, text float32
}

// zoomFrame places both levels e of the way through a drill, in paint order.
// Both ride one frame that carries the drilled branch from its tile to the
// whole view, or back, so no tile crosses another; each is laid out at its
// own size, so corners and labels keep their proportions. Within the branch
// the old level fades out as the new one fades in, the labels of one gone
// before the other's appear; tiles beside it fly outward or back in whole.
func (v *view) zoomFrame(e float32) []zoomed {
	full := ui.Rect{W: v.width, H: v.height}
	// The branch's place in each level: its tile, or the whole view in the
	// level it is the root of.
	rootOld, rootNew := full, full
	if i := slices.IndexFunc(v.old, func(tl tile) bool { return tl.id == v.zoomRoot }); i >= 0 {
		rootOld = v.old[i].rect
	}
	if i := v.index(v.zoomRoot); i >= 0 {
		rootNew = v.tiles[i].rect
	}
	root := lerp(rootOld, rootNew, e)
	ramp := func(from, to float32) float32 { return max(0, min(1, (e-from)/(to-from))) }
	place := func(tiles []tile, home ui.Rect, fade, text float32) []zoomed {
		// A level that tiles the branch fades only the branch's part.
		inside := make([]bool, len(tiles))
		tiled := slices.ContainsFunc(tiles, func(tl tile) bool { return tl.id == v.zoomRoot })
		list := make([]zoomed, len(tiles))
		for i, tl := range tiles {
			inside[i] = !tiled || tl.id == v.zoomRoot || (tl.parent >= 0 && inside[tl.parent])
			z := zoomed{tl: tl, b: relate(tl.rect, home, root), opacity: 1, text: 1}
			if inside[i] {
				z.opacity, z.text = fade, text
			}
			z.radius = v.radius(z.b, tl.depth)
			list[i] = z
		}
		return list
	}
	olds := place(v.old, rootOld, 1-ramp(0, .8), 1-ramp(0, .35))
	news := place(v.tiles, rootNew, ramp(.2, 1), ramp(.65, 1))
	// The level opening inside the branch lies above the old one, and the
	// level closing into it above the new one.
	if v.zoomIn {
		return append(olds, news...)
	}
	return append(news, olds...)
}

// bordered reports whether s replaces a border.
func (s Style) bordered() bool {
	return s.Border != (ui.Color{}) || s.BorderWidth > 0 || s.BorderStyle != ui.BorderSolid
}

// radius rounds a tile, less with depth so nested corners stay concentric
// with their parents', and never more than a small tile can take.
func (v *view) radius(b ui.Rect, depth int) float32 {
	m := v.metrics
	r := max(m.radius-m.radiusStep*float32(depth), m.radiusStep)
	return max(0, min(r, b.W/3, b.H/3))
}

// index returns the tile with id, or -1.
func (v *view) index(id string) int {
	if id == "" {
		return -1
	}
	for i := range v.tiles {
		if v.tiles[i].id == id {
			return i
		}
	}
	return -1
}

// relate places r, a rectangle given within from, at the same place within
// to.
func relate(r, from, to ui.Rect) ui.Rect {
	if from.W <= 0 || from.H <= 0 {
		return to
	}
	sx, sy := to.W/from.W, to.H/from.H
	return ui.Rect{X: to.X + (r.X-from.X)*sx, Y: to.Y + (r.Y-from.Y)*sy, W: r.W * sx, H: r.H * sy}
}

// offset moves a canvas rectangle into window coordinates.
func offset(b, r ui.Rect) ui.Rect { b.X, b.Y = b.X+r.X, b.Y+r.Y; return b }
