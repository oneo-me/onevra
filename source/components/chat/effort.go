package chat

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// EffortPickerOptions describes a stepped choice of reasoning effort.
// Levels run from weakest to strongest; only their ID and Label are used.
// Off makes the first level turn reasoning off: it leaves the bar empty
// and dims the icon. Label is the trigger's and the bar's accessible name
// and heads the panel; Tooltip is the trigger's hover text, Label by
// default. Icon draws the trigger. Width is the panel's width, 260 by
// default.
type EffortPickerOptions struct {
	Label, Tooltip string
	Icon           func() *ui.SVG
	Levels         []Option
	Selected       string
	Off            bool
	Width          float32
}

// EffortState retains whether an EffortPicker's panel is open, and its
// animation's own clock.
type EffortState struct {
	Open bool
	time effortTime
}

// effortTransition is how long the motion takes to settle on a new level.
const effortTransition = 900 * time.Millisecond

// EffortPicker shows a Chip with the icon, the selected level's name and
// a caret, suited to the composer's Accessory slot, that opens a panel
// with the selected level's name and a progress bar above the level names,
// spread from edge to edge; each level fills the bar to the end of its
// name. Clicking or dragging across the bar or the names, or the arrows,
// Home and End, choose a level and keep the panel open; the application
// applies it. The panel widens to fit the names.
//
// The stronger the level, the livelier all three: the icon turns violet
// and gains a breathing halo, then streams blocks of data outward; the
// bar's violet fill gains a glint, then streaming blocks of data, a glow,
// flowing colors, sparkles and rising sparks; the panel gains a violet
// wash, then drifting aurora, floating motes and a turning border. Motion
// stays slow, quickening a little at the top, and each effect fades in or
// out as the level changes. Reduced Motion keeps them still.
func EffortPicker(c *ui.Context, state *EffortState, options EffortPickerOptions) PickerResult {
	n := len(options.Levels)
	if n == 0 {
		return PickerResult{}
	}
	t := c.Theme()
	selected := 0
	for i, o := range options.Levels {
		if o.ID == options.Selected {
			selected = i
		}
	}
	off := options.Off && selected == 0
	level := options.Levels[selected].Label

	var icon *ui.SVG
	if options.Icon != nil {
		icon = options.Icon()
	}
	// The icon paints after the build, so it reads the motion animated on
	// the finished trigger.
	var m effortMotion
	trigger := Chip(c, level, ChipOptions{Caret: true, Leading: func(c *ui.Context, size float32) {
		ui.Box(c).Size(size, size).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) { paintEffortIcon(p, r, icon, m) })
	}})
	trigger.Label(options.Label).Value(level)
	if state.Open {
		trigger.Background(overlayStrong(t))
	}
	// Effects follow these values, not the level, so changing the level
	// blends them rather than switching.
	m = effortMotion{
		strength: trigger.Animate("strength", effortStrength(selected, n, options.Off), effortTransition),
		off:      trigger.Animate("off", effortFlag(off), effortTransition/2),
		still:    c.Preferences().ReduceMotion,
		time:     &state.time,
	}
	tooltip := options.Tooltip
	if tooltip == "" {
		tooltip = options.Label
	}
	components.Tooltip(c, trigger, tooltip)
	if trigger.Clicked() {
		state.Open = !state.Open
	}

	var chosen string
	width := options.Width
	if width <= 0 {
		width = 260
	}
	components.Popover(c, trigger, &state.Open, func(panel ui.Element) {
		const padding, minGap = 14, 12
		// Names are laid out at their selected weight, so selecting one
		// does not move the others.
		names := make([][]ui.Glyph, n)
		slots := make([]float32, n)
		total := float32(0)
		for i, o := range options.Levels {
			names[i] = ui.Shape(o.Label, effortNameFont(i == selected))
			slots[i] = advance(ui.Shape(o.Label, effortNameFont(true)))
			total += slots[i]
		}
		width := max(width, total+float32(n-1)*minGap+2*padding)
		inner := width - 2*padding
		// The names spread from edge to edge with equal gaps, and each
		// level's fill ends where its name does; Off fills nothing.
		starts, ends := make([]float32, n), make([]float32, n)
		gap := float32(0)
		if n > 1 {
			gap = (inner - total) / float32(n-1)
		}
		for i, u := 0, float32(0); i < n; i++ {
			starts[i], ends[i] = u, (u+slots[i])/inner
			u += slots[i] + gap
		}
		ends[n-1] = 1
		if options.Off {
			ends[0] = 0
		}
		// The popover's radius, to keep the backdrop inside its corners.
		radius := t.Radius * 1.5
		panel.Width(width).Padding(padding).Gap(10).Draw(func(p *ui.Painter, r ui.Rect) { paintEffortBackdrop(p, r, radius, m) })
		panel.Children(func() {
			ui.Row(c).FillWidth().MinWidth(0).AlignItems(ui.Center).Gap(8).Children(func() {
				components.Text(c, options.Label).Grow(1).MinWidth(0).SingleLine().Ellipsis("…").FontSize(12.5)
				components.Text(c, level).Shrink(0).FontSize(12.5).FontWeight(600).TextColor(m.ink(t))
			})
			value := float64(selected)
			// The bar's width divides evenly between the levels around the
			// spread names, so pressing a name or the bar above it selects
			// that level.
			bar := ui.SliderBase(c, &value, 0, float64(n-1)).Step(1).Label(options.Label).Value(level).
				Size(inner, effortBarHeight+effortNameGap+effortNameHeight).
				FocusRing(false).Cursor(ui.CursorPointer).AutoFocus()
			if bar.Changed() {
				if i := int(math.Round(value)); i >= 0 && i < n && i != selected {
					chosen = options.Levels[i].ID
				}
			}
			e := effortPaint{effortMotion: m, fraction: effortEnd(ends, bar.Animate("fill", float32(selected), effortTransition*2/3)),
				hovered: bar.Hovered(), focus: bar.FocusVisible(), rtl: c.Direction() == ui.RightToLeft,
				names: names, starts: starts, slots: slots, selected: selected}
			bar.Draw(func(p *ui.Painter, r ui.Rect) { paintEffort(p, r, e) })
		})
	})
	return PickerResult{Trigger: trigger, Chosen: chosen}
}

// effortStrength places a level from 0, the weakest, or Off and the
// weakest level after it, to 1, the strongest.
func effortStrength(selected, n int, off bool) float32 {
	if off {
		selected, n = max(selected-1, 0), n-1
	}
	if n <= 1 {
		return 1
	}
	return float32(selected) / float32(n-1)
}

func effortFlag(on bool) float32 {
	if on {
		return 1
	}
	return 0
}

// effortTime is an animation clock whose pace follows the strength. It
// advances by each frame's elapsed time, so changing the pace never moves
// anything abruptly.
type effortTime struct {
	at      time.Time
	seconds float64
}

// advance moves the clock to now at pace, once per frame however many
// painters ask, and returns it. Gaps between frames, as while nothing
// animates, count as at most a frame.
func (c *effortTime) advance(now time.Time, pace float64) float64 {
	if !c.at.IsZero() && now.After(c.at) {
		c.seconds += min(now.Sub(c.at).Seconds(), .05) * pace
	}
	if now.After(c.at) {
		c.at = now
	}
	return c.seconds
}

// effortMotion is the animated level the icon, bar and panel draw.
type effortMotion struct {
	strength, off float32
	still         bool
	time          *effortTime
}

// effortFrame is how often the panel's effects move. The trigger's icon,
// shown all the time, moves every effortIconFrame, as its halo breathes
// and its blocks drift slowly.
const (
	effortFrame     = 16 * time.Millisecond
	effortIconFrame = 33 * time.Millisecond
)

// clock advances the animation clock and reports whether the effects
// move, asking for the next frame after every when they do.
func (m effortMotion) clock(p *ui.Painter, every time.Duration) (float64, bool) {
	if m.still || m.strength < .001 {
		return m.time.seconds, false
	}
	p.After(every)
	return m.time.advance(p.Now(), 1+.8*float64(smoothstep(.6, 1, m.strength))), true
}

// ink is readable violet text, warming toward fuchsia with strength and
// muted while off.
func (m effortMotion) ink(t *ui.Theme) ui.Color {
	return effortInk(t, m.strength).Mix(t.TextMuted, m.off)
}

// effortEnd is the fill's end for a level index between whole levels.
func effortEnd(ends []float32, index float32) float32 {
	i := int(math.Floor(float64(index)))
	if i < 0 {
		return ends[0] * max(0, 1+index)
	}
	if i >= len(ends)-1 {
		return ends[len(ends)-1]
	}
	f := index - float32(i)
	return ends[i] + (ends[i+1]-ends[i])*f
}

const (
	effortBarHeight  = 18
	effortNameGap    = 8
	effortNameHeight = 16
)

func effortNameFont(selected bool) ui.Font {
	if selected {
		return ui.Font{Size: 11.5, Weight: 600}
	}
	return ui.Font{Size: 11.5}
}

// The bar's violets, from indigo to fuchsia.
var (
	effortIndigo  = ui.Hex("#6366f1")
	effortViolet  = ui.Hex("#8b5cf6")
	effortPurple  = ui.Hex("#a855f7")
	effortFuchsia = ui.Hex("#d946ef")
	effortCycle   = []ui.Color{effortIndigo, effortViolet, effortFuchsia, effortPurple}
	effortWhite   = ui.RGB(255, 255, 255)
)

func effortInk(t *ui.Theme, strength float32) ui.Color {
	if t.Dark {
		return ui.Hex("#a78bfa").Mix(ui.Hex("#e879f9"), .5*strength)
	}
	return ui.Hex("#7c3aed").Mix(ui.Hex("#c026d3"), .5*strength)
}

type effortPaint struct {
	effortMotion
	fraction            float32
	hovered, focus, rtl bool
	names               [][]ui.Glyph
	starts, slots       []float32
	selected            int
}

// paintEffort draws the bar with its fill and effects in a start-relative
// coordinate u, mirrored right to left as the slider's arrows go, and the
// level names beneath it: the first at the start, the last at the end and
// the others centered in their slots.
func paintEffort(p *ui.Painter, r ui.Rect, e effortPaint) {
	t := p.Theme()
	track := ui.Rect{X: r.X, Y: r.Y, W: r.W, H: effortBarHeight}
	radius := track.H / 2
	at := func(u float32) float32 {
		if e.rtl {
			return track.X + track.W - u
		}
		return track.X + u
	}
	span := func(u0, u1 float32) ui.Rect {
		x0, x1 := at(u0), at(u1)
		return ui.Rect{X: min(x0, x1), Y: track.Y, W: abs32(x1 - x0), H: track.H}
	}
	end := track.W * clamp01(e.fraction)
	clock, lively := e.clock(p, effortFrame)
	s := e.strength
	top := smoothstep(.75, 1, s)
	// The fill runs from indigo across to a violet warming with strength,
	// and blends into a flow through all the violets at the top.
	flow := float32(0)
	if lively {
		flow = effortPhase(clock, 16)
	}
	color := func(u float32) ui.Color {
		f := u / max(track.W, 1)
		across := effortIndigo.Mix(effortViolet, .25).Mix(effortViolet.Mix(effortFuchsia, .6*s), clamp01(f))
		if top <= 0 {
			return across
		}
		return across.Mix(effortColor(effortCycle, f-flow), top)
	}

	// The fill is one pill from the start. Shorter than its own cap, it
	// keeps the cap's shape and fades instead, so it never squeezes. It
	// clips its effects alone: the GPU renderer rounds only the innermost
	// of nested clips, so a second clip around it would leak at the
	// corners.
	vis := clamp01(end / (2 * radius))
	body := span(0, max(end, 2*radius))
	// The glow stays within the panel's padding, so it fades out before
	// the border.
	if glow := smoothstep(.45, 1, s) * vis; glow > 0 {
		pulse := float32(.8)
		if lively {
			pulse = .7 + .3*float32(math.Sin(clock*1.4))
		}
		p.Shadow(body, radius, 0, 0, 8, 0, effortViolet.Alpha(.55*glow*pulse))
		if top > 0 {
			p.Shadow(body, radius, 0, 0, 12, 0, color(end*.5).Alpha(.45*top*pulse*vis))
		}
	}
	rest := t.Text.Alpha(.08)
	if e.hovered {
		rest = t.Text.Alpha(.12)
	}
	p.Fill(track, rest, radius)
	if vis > 0 {
		const stops = 9
		g := ui.LinearGradient{Angle: 90, Stops: make([]ui.GradientStop, stops)}
		if e.rtl {
			g.Angle = 270
		}
		for i := range stops {
			f := float32(i) / (stops - 1)
			g.Stops[i] = ui.GradientStop{Color: color(f * body.W).Alpha(vis), At: f}
		}
		p.FillGradient(body, g, radius)
		p.Clip(body, radius, func() {
			// A gloss across the top gives the fill depth.
			p.FillGradient(ui.Rect{X: body.X, Y: body.Y, W: body.W, H: body.H / 2}, ui.LinearGradient{
				From: effortWhite.Alpha(.22 * vis), To: effortWhite.Alpha(0), Angle: 180}, 0)
			if lively {
				effortBlocks(p, track, at, s, clock, end, vis)
				effortGlint(p, track, at, s, clock, vis)
				effortSparkles(p, track, at, s, clock, end, vis)
			}
		})
	}
	if lively {
		effortSparks(p, track, at, s, clock, end)
	}
	if e.focus {
		p.FocusRing(track, [4]float32{radius, radius, radius, radius})
	}

	n := len(e.names)
	m := effortNameFont(false).Metrics()
	baseline := track.Y + track.H + effortNameGap + (effortNameHeight-m.Ascent-m.Descent)/2 + m.Ascent
	for i, glyphs := range e.names {
		ink := t.TextMuted
		if i == e.selected {
			ink = t.Text
		}
		w := advance(glyphs)
		u := e.starts[i] + (e.slots[i]-w)/2
		switch {
		case i == 0:
			u = e.starts[i]
		case i == n-1:
			u = e.starts[i] + e.slots[i] - w
		}
		x := track.X + u
		if e.rtl {
			x = track.X + track.W - u - w
		}
		p.Glyphs(glyphs, x, baseline, ink)
	}
}

// paintEffortBackdrop tints the panel behind its content: a violet wash
// rising with strength, then drifting aurora, and at the top floating
// motes and a border of turning violets.
func paintEffortBackdrop(p *ui.Painter, r ui.Rect, radius float32, m effortMotion) {
	t := p.Theme()
	s := m.strength
	wash := smoothstep(.2, 1, s)
	if wash <= 0 {
		return
	}
	clock, lively := m.clock(p, effortFrame)
	// Dark panels need more color for the same presence.
	boost := float32(1)
	if t.Dark {
		boost = 1.5
	}
	p.Clip(r, radius, func() {
		p.FillGradient(r, ui.LinearGradient{Angle: 180, Stops: []ui.GradientStop{
			{Color: effortViolet.Alpha(0), At: .25}, {Color: effortViolet.Alpha(.1 * wash * boost), At: 1},
		}}, 0)
		// Aurora: soft glows wandering slowly, more of them as the level
		// rises.
		for j, c := range []ui.Color{effortViolet, effortFuchsia, effortIndigo} {
			shown := smoothstep(.45+.15*float32(j), 1, s)
			if shown <= 0 {
				continue
			}
			k := float64(j)
			x := .5 + .4*math.Sin(clock*.23+k*2.1)
			y := .7 + .25*math.Sin(clock*.31+k*1.3)
			if !lively {
				x, y = .3+.2*k, .8
			}
			blob := ui.Rect{X: r.X + r.W*float32(x), Y: r.Y + r.H*float32(y)}
			p.Shadow(blob, 0, 0, 0, 40, 16, c.Alpha(.16*shown*boost))
		}
		top := smoothstep(.75, 1, s)
		if top <= 0 || !lively {
			return
		}
		// Motes float up through the panel and fade.
		for j := range 12 {
			seed := effortHash(j + 101)
			period := 5 + 4*float64(seed)
			tm := clock + float64(seed)*11
			f := effortPhase(tm, period)
			cycle := int(math.Floor(tm/period)) % 1_000_003
			x := r.X + 12 + (r.W-24)*effortHash(j*7+cycle) + 6*float32(math.Sin(tm*1.1))
			y := r.Y + r.H*(1-f)
			a := top * float32(math.Sin(float64(f)*math.Pi)) * (.35 + .4*seed)
			var dot ui.Path
			dot.Circle(x, y, .8+1.2*seed)
			p.FillPath(&dot, effortCycle[j%len(effortCycle)].Mix(effortWhite, .3).Alpha(a))
		}
		// The border turns through the violets.
		stops := make([]ui.GradientStop, 0, len(effortCycle)+1)
		for i, c := range effortCycle {
			stops = append(stops, ui.GradientStop{Color: c.Alpha(.75 * top), At: float32(i) / float32(len(effortCycle))})
		}
		stops = append(stops, ui.GradientStop{Color: stops[0].Color, At: 1})
		var edge ui.Path
		roundedRect(&edge, ui.Rect{X: r.X + .75, Y: r.Y + .75, W: r.W - 1.5, H: r.H - 1.5}, radius-.75)
		p.StrokePathConicGradient(&edge, 1.5, ui.ConicGradient{Stops: stops, Angle: 360 * effortPhase(clock, 9)})
	})
}

// paintEffortIcon draws the trigger's icon: dimmed while off, muted at the
// weakest level and turning violet with strength, breathing a halo in the
// upper levels, and at the top shifting color while blocks of data stream
// outward from it.
func paintEffortIcon(p *ui.Painter, r ui.Rect, icon *ui.SVG, m effortMotion) {
	t := p.Theme()
	cx, cy := r.X+r.W/2, r.Y+r.H/2
	s := m.strength
	// Below the halo nothing of the icon moves, so it asks for no frames.
	halo := smoothstep(.4, 1, s)
	clock, lively := m.time.seconds, false
	if halo > 0 {
		clock, lively = m.clock(p, effortIconFrame)
	}
	ink := tertiary(t).Mix(effortInk(t, s), min(1, s*1.5))
	if halo > 0 {
		pulse := float32(.6)
		if lively {
			pulse = .5 + .5*float32(math.Sin(clock*1.3))
		}
		var circle ui.Path
		circle.Circle(cx, cy, 7+1.5*pulse)
		p.FillPath(&circle, effortViolet.Alpha(halo*(.12+.18*pulse)))
	}
	if top := smoothstep(.75, 1, s); top > 0 && lively {
		ink = ink.Mix(effortColor([]ui.Color{effortInk(t, 0), effortInk(t, 1), effortInk(t, .5)}, effortPhase(clock, 12)), top)
		effortBurst(p, cx, cy, top, clock)
	}
	ink = ink.Alpha(1 - .5*m.off)
	if icon != nil {
		const size = 12
		p.Icon(icon, ui.Rect{X: cx - size/2, Y: cy - size/2, W: size, H: size}, ink)
	}
}

// effortBurst sends small squares out from the icon's edge like packets
// of data. Each leaves in a new direction every trip, shrinking and fading
// as it goes; the burst reaches less far vertically, to stay in the chip.
func effortBurst(p *ui.Painter, cx, cy, shown float32, clock float64) {
	for j := range 10 {
		seed := effortHash(j + 211)
		period := 1.4 + 1.2*float64(seed)
		tm := clock + float64(seed)*5
		f := effortPhase(tm, period)
		cycle := int(math.Floor(tm/period)) % 1_000_003
		angle := 2 * math.Pi * float64(effortHash(j*19+cycle))
		d := 5 + f*(7+3*seed)
		a := shown * smoothstep(0, .15, f) * (1 - f) * (.6 + .4*seed)
		if a <= .01 {
			continue
		}
		size := (1.6 + 1.2*seed) * (1 - .4*f)
		x := cx + d*float32(math.Cos(angle))
		y := cy + .8*d*float32(math.Sin(angle))
		p.Fill(ui.Rect{X: x - size/2, Y: y - size/2, W: size, H: size}, effortCycle[j%len(effortCycle)].Mix(effortWhite, .2).Alpha(a), .4)
	}
}

// effortGlint sweeps a soft highlight along the bar, brighter with
// strength, and at the top a second one the other way; the fill's clip
// shows only the part over it.
func effortGlint(p *ui.Painter, track ui.Rect, at func(float32) float32, strength float32, clock float64, vis float32) {
	sweep := func(alpha float32, period float64, back bool) {
		if alpha <= 0 {
			return
		}
		band := max(track.W*.25, 32)
		phase := effortPhase(clock, period)
		if back {
			phase = 1 - phase
		}
		center := -band + phase*(track.W+2*band)
		x0, x1 := at(center-band), at(center+band)
		p.FillGradient(ui.Rect{X: min(x0, x1), Y: track.Y, W: abs32(x1 - x0), H: track.H}, ui.LinearGradient{Angle: 90, Stops: []ui.GradientStop{
			{Color: effortWhite.Alpha(0), At: 0}, {Color: effortWhite.Alpha(alpha * .5), At: .3}, {Color: effortWhite.Alpha(alpha), At: .5},
			{Color: effortWhite.Alpha(alpha * .5), At: .7}, {Color: effortWhite.Alpha(0), At: 1},
		}}, 0)
	}
	sweep(vis*smoothstep(.1, .4, strength)*(.12+.28*strength), 4.5, false)
	sweep(vis*smoothstep(.75, 1, strength)*.3, 6.5, true)
}

// effortBlocks drifts small squares along the bar like packets of data.
// Each fades in at its own strength, so more appear as the level rises,
// with trails at the top; they travel the whole bar at a fixed pace and
// fade where the fill ends, so neither a new level nor the fill's motion
// moves them.
func effortBlocks(p *ui.Painter, track ui.Rect, at func(float32) float32, strength float32, clock float64, end, vis float32) {
	const count, size = 20, 3
	if strength <= .4 {
		return
	}
	lanes := []float32{.28, .5, .72}
	trail := smoothstep(.8, 1, strength)
	loop := track.W + 8*size
	for j := range count {
		seed := effortHash(j)
		shown := clamp01((strength - .4 - .5*float32(j)/count) / .1)
		if shown <= 0 {
			continue
		}
		v := float64(18 + 22*seed)
		u := float32(math.Mod(clock*v+float64(seed*loop*7), float64(loop))) - 4*size
		y := track.Y + track.H*lanes[j%len(lanes)] - size/2
		for k := range 3 {
			alpha := shown * .65 * vis
			if k > 0 {
				alpha *= trail * (.5 - .18*float32(k))
			}
			uk := u - float32(k)*size*1.6
			fade := clamp01(min((uk+size)/(3*size), (end-uk)/(3*size)))
			if a := alpha * fade; a > .005 {
				p.Fill(ui.Rect{X: at(uk) - size/2, Y: y, W: size, H: size}, effortWhite.Alpha(a), .75)
			}
		}
	}
}

// effortSparkles twinkles four-pointed stars slowly at the top levels.
func effortSparkles(p *ui.Painter, track ui.Rect, at func(float32) float32, strength float32, clock float64, end, vis float32) {
	shown := smoothstep(.75, 1, strength) * vis
	if shown <= 0 {
		return
	}
	for j := range 6 {
		seed := effortHash(j + 17)
		period := 2.4 + 1.6*float64(seed)
		tm := clock + float64(seed)*7
		glow := float32(math.Sin(float64(effortPhase(tm, period)) * math.Pi))
		glow *= glow
		// Each twinkle lands somewhere new, while it is invisible.
		cycle := int(math.Floor(tm/period)) % 1_000_003
		u := track.W * effortHash(j*31+cycle)
		fade := clamp01((end - u) / 12)
		if a := shown * glow * fade; a > .01 {
			cy := track.Y + 4 + (track.H-8)*effortHash(j*47+cycle)
			effortStar(p, at(u), cy, (2+2.5*seed)*glow, effortWhite.Alpha(.95*a))
		}
	}
}

// effortSparks rise from the strongest fill's top edge and fade.
func effortSparks(p *ui.Painter, track ui.Rect, at func(float32) float32, strength float32, clock float64, end float32) {
	shown := smoothstep(.75, 1, strength)
	if shown <= 0 {
		return
	}
	for j := range 14 {
		seed := effortHash(j + 53)
		period := 2.2 + 1.6*float64(seed)
		tm := clock + float64(seed)*9
		f := effortPhase(tm, period)
		cycle := int(math.Floor(tm/period)) % 1_000_003
		u := track.W * effortHash(j*13+cycle)
		fade := clamp01((end - u) / 12)
		a := shown * fade * float32(math.Sin(float64(f)*math.Pi))
		if a <= .01 {
			continue
		}
		x := at(u) + 3*float32(math.Sin(tm*1.3))
		y := track.Y + 2 - f*(10+8*seed)
		var dot ui.Path
		dot.Circle(x, y, (.7+.8*seed)*(1-.4*f))
		p.FillPath(&dot, effortCycle[j%len(effortCycle)].Mix(effortWhite, .45).Alpha(a))
	}
}

// effortStar fills a four-pointed star of radius size.
func effortStar(p *ui.Painter, cx, cy, size float32, c ui.Color) {
	var star ui.Path
	star.MoveTo(cx, cy-size)
	star.QuadTo(cx, cy, cx+size, cy)
	star.QuadTo(cx, cy, cx, cy+size)
	star.QuadTo(cx, cy, cx-size, cy)
	star.QuadTo(cx, cy, cx, cy-size)
	p.FillPath(&star, c)
}

// roundedRect adds a closed rounded rectangle with circular corners.
func roundedRect(path *ui.Path, r ui.Rect, radius float32) {
	k := radius * (1 - .5523)
	x0, y0, x1, y1 := r.X, r.Y, r.X+r.W, r.Y+r.H
	path.MoveTo(x0+radius, y0)
	path.LineTo(x1-radius, y0)
	path.CubeTo(x1-k, y0, x1, y0+k, x1, y0+radius)
	path.LineTo(x1, y1-radius)
	path.CubeTo(x1, y1-k, x1-k, y1, x1-radius, y1)
	path.LineTo(x0+radius, y1)
	path.CubeTo(x0+k, y1, x0, y1-k, x0, y1-radius)
	path.LineTo(x0, y0+radius)
	path.CubeTo(x0, y0+k, x0+k, y0, x0+radius, y0)
	path.Close()
}

// effortColor samples a cyclic palette at f, so a flow repeats seamlessly.
func effortColor(palette []ui.Color, f float32) ui.Color {
	n := len(palette)
	f -= float32(math.Floor(float64(f)))
	x := f * float32(n)
	i := int(x) % n
	return palette[i].Mix(palette[(i+1)%n], x-float32(int(x)))
}

// effortPhase is how far clock is through a period, from 0 to 1.
func effortPhase(clock, period float64) float32 {
	return float32(math.Mod(clock, period) / period)
}

// effortHash is a fixed pseudo-random fraction for an index.
func effortHash(i int) float32 {
	s := math.Sin(float64(i)*12.9898+78.233) * 43758.5453
	return float32(s - math.Floor(s))
}

func smoothstep(lo, hi, v float32) float32 {
	x := clamp01((v - lo) / (hi - lo))
	return x * x * (3 - 2*x)
}

func clamp01(v float32) float32 { return min(max(v, 0), 1) }

func abs32(v float32) float32 { return max(v, -v) }
