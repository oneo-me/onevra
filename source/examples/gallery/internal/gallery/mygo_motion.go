package gallery

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

func exampleMotion(c *ui.Context, s *demoState) {
	t := c.Theme()
	switch s.variant {
	case "Timeline":
		exampleTimeline(c, s)
	case "Frame rate":
		exampleFrameRate(c, s)
	case "Keyframes":
		if components.Button(c, tr(c, "motionReplay"), components.ButtonOutline).Clicked() {
			s.count++
		}
		ui.Box(c).FillWidth().Height(80).Padding(12).Children(func() {
			tile := ui.Box(c).Key("sequence").Size(44, 44).Radius(10).Background(t.Accent).Label(tr(c, "motionTile"))
			x := tile.Keyframes("x", []ui.Keyframe{
				{Value: 0}, {At: 200 * time.Millisecond, Value: 110, Ease: ui.EaseOut},
				{At: 500 * time.Millisecond, Value: 55}, {At: 800 * time.Millisecond, Value: 0, Ease: ui.EaseInOut},
			}, ui.KeyframeOptions{Run: uint64(s.count), Repeat: 1, Alternate: true})
			tile.Rotate(x/3).Translate(x, 0)
		})
	case "Animated pose":
		components.Checkbox(c, &s.on, tr(c, "motionLift"))
		ui.Box(c).FillWidth().Height(100).Padding(20).Children(func() {
			card := components.Button(c, tr(c, "motionHover"), components.ButtonOutline)
			pose := ui.Pose{}
			if s.on || card.Hovered() {
				pose = ui.Pose{Y: -8, Scale: 1.04, Rotate: -4}
			}
			card.AnimatePose(pose, ui.Snappy)
		})
	case "Draggable":
		if components.Button(c, tr(c, "motionReset"), components.ButtonOutline).Clicked() {
			s.offset = ui.Point{}
		}
		ui.Box(c).FillWidth().Height(100).Padding(20).Children(func() {
			card := ui.Box(c).Size(90, 60).Radius(10).Background(t.Accent).Label(tr(c, "motionThrow"))
			card.Draggable(&s.offset, ui.DraggableOptions{
				Axis:   ui.Horizontal,
				Bounds: &ui.Rect{W: 120},
				Snap:   ui.Nearest(ui.Point{}, ui.Point{X: 120}),
			})
			lift := ui.Pose{}
			if card.Dragging() {
				lift = ui.Pose{Scale: 1.05}
			}
			card.AnimatePose(lift, ui.Snappy)
		})
	default:
		if components.Button(c, tr(c, "motionReverse"), components.ButtonOutline).Clicked() {
			s.on = !s.on
		}
		for i, animation := range []ui.Animation{ui.Smooth, ui.Snappy, ui.Bouncy} {
			ui.Column(c).Key(i).FillWidth().Gap(4).Children(func() {
				components.Text(c, tr(c, []string{"motionSmooth", "motionSnappy", "motionBouncy"}[i])).TextColor(t.TextMuted)
				ui.Box(c).FillWidth().Height(56).Padding(6).Children(func() {
					tile := ui.Box(c).Size(36, 36).Radius(9).Background(t.Accent)
					target := float32(0)
					if s.on {
						target = 120
					}
					tile.Translate(tile.AnimateTo("x", target, animation), 0)
				})
			})
		}
	}
}

func exampleTransforms(c *ui.Context, s *demoState) {
	t := c.Theme()
	if s.variant == "Perspective" {
		if components.Button(c, tr(c, "transformFlip"), components.ButtonOutline).Clicked() {
			s.on = !s.on
		}
		ui.Box(c).FillWidth().Height(140).Padding(20).Children(func() {
			card := ui.Box(c).Size(150, 96)
			target := float32(0)
			if s.on {
				target = 180
			}
			angle := card.Spring("turn", target, ui.SpringOptions{Duration: 600 * time.Millisecond, Bounce: .1})
			card.Children(func() {
				for i, label := range []string{"transformFront", "transformBack"} {
					ui.Box(c).Absolute().Fill().Radius(12).Background(t.Surface).Border(1, t.Accent).
						RotateY(angle + float32(i)*180).Perspective(800).BackfaceHidden().
						AlignItems(ui.Center).Justify(ui.Center).Children(func() { components.Text(c, tr(c, label)) })
				}
			})
		})
		return
	}
	components.Slider(c, &s.value, -15, 15).Label(tr(c, "transformAngle")).FillWidth()
	components.Checkbox(c, &s.on, tr(c, "transformScale"))
	ui.Box(c).FillWidth().Height(230).Padding(24, 12).Children(func() {
		scale := float32(1)
		if s.on {
			scale = .85
		}
		pose := ui.Scaling(scale, scale).Then(ui.Rotation(float32(s.value))).Then(ui.Translation(8, 8))
		ui.Column(c).Width(160).Padding(12).Gap(10).Radius(12).Background(t.Surface).Border(1, t.Border).
			Transform(pose).TransformOrigin(.5, .5).Children(func() {
			components.TextBox(c, &s.text, components.TextBoxOptions{}).Label(tr(c, "transformNote")).FillWidth()
			anchor := components.Button(c, tr(c, "showPopover"), components.ButtonOutline).FillWidth()
			if anchor.Clicked() {
				s.open = !s.open
			}
			components.Popover(c, anchor, &s.open, func(panel ui.Element) {
				panel.Label(tr(c, "popoverContent"))
				components.Text(c, tr(c, "popoverContent"))
				if components.Button(c, tr(c, "close"), components.ButtonSecondary).Clicked() {
					s.open = false
				}
			})
		})
	})
}

// exampleTimeline drives several tweens from one playhead. Moving the scrubber
// pauses playback there; Play continues from the scrubbed position.
func exampleTimeline(c *ui.Context, s *demoState) {
	t := c.Theme()
	paused := s.on
	ui.Row(c).Gap(8).Children(func() {
		if components.Button(c, tr(c, "timelineRestart"), components.ButtonOutline).Clicked() {
			s.count++
			s.on = false
		}
		toggle := map[bool]string{true: "timelinePlay", false: "timelinePause"}[s.on]
		if components.Button(c, tr(c, toggle), components.ButtonOutline).Clicked() {
			s.on = !s.on
		}
		components.ToggleButton(c, &s.other, tr(c, "timelineReverse"))
	})
	o := ui.TimelineOptions{Run: uint64(s.count), Paused: s.on, Reverse: s.other}
	if paused {
		o.Progress = &s.seek
	}
	var tl *ui.Timeline
	ui.Box(c).FillWidth().Height(150).Padding(12).Children(func() {
		card := ui.Column(c).Key("timeline-card")
		tl = card.Timeline("intro", o)
		slide := tl.Tween(ui.Tween{From: -160, To: 0, Duration: 450 * time.Millisecond})
		fade := tl.Tween(ui.Tween{From: 0, To: 1, Duration: 300 * time.Millisecond, Ease: ui.Linear}, ui.AtTime(150*time.Millisecond))
		tl.Label("rows")
		var rows [3]float32
		for i := range rows {
			rows[i] = tl.Tween(ui.Tween{From: 0, To: 1, Duration: 240 * time.Millisecond, Ease: ui.Linear},
				ui.AtLabel("rows", time.Duration(i)*90*time.Millisecond))
		}
		bounce := tl.Tween(ui.Tween{From: 8, To: 0, Duration: 360 * time.Millisecond, Ease: ui.Bounce(ui.EaseOut)}, ui.AtTime(450*time.Millisecond))
		card.Width(220).Padding(14).Gap(6).Radius(12).Background(t.Surface).Border(1, t.Border).
			Translate(slide, bounce).Label(tr(c, "timelineCard")).Children(func() {
			components.Text(c, tr(c, "timelineCard")).Bold().Opacity(fade)
			for i, label := range []string{"timelinePlan", "timelineBuild", "timelineShip"} {
				components.Text(c, tr(c, label)).TextColor(t.TextMuted).Opacity(rows[i]).Translate(0, (1-rows[i])*10)
			}
		})
	})
	if !paused {
		s.value = float64(tl.Progress())
	}
	if components.Slider(c, &s.value, 0, 1).Label(tr(c, "timelineScrub")).FillWidth().Changed() {
		s.on = true
	}
	s.seek = float32(s.value)
	components.Text(c, msg(c).TimelineTime(int(tl.Time().Milliseconds()), int(tl.Duration().Milliseconds()))).TextColor(t.TextMuted)
}

// motionFrameRates are the record's paint rates; 0 is the display's.
var motionFrameRates = []float32{30, 60, 0}

// exampleFrameRate turns a record at a chosen rate. Its grooves are still
// and costly, so Cache keeps them as one bitmap, and only the turning marker
// above them paints each frame.
func exampleFrameRate(c *ui.Context, s *demoState) {
	t := c.Theme()
	components.Tabs(c, &s.index, components.TabsSegmented,
		msg(c).MotionFps(30), msg(c).MotionFps(60), tr(c, "motionDisplayRate")).List.Label(tr(c, "frameRate"))
	fps, still := motionFrameRates[s.index], c.Preferences().ReduceMotion
	const size = 160
	ui.Box(c).Size(size, size).Label(tr(c, "motionRecord")).Children(func() {
		ui.Box(c).Fill().Cache().Draw(func(p *ui.Painter, r ui.Rect) {
			cx, cy := r.X+r.W/2, r.Y+r.H/2
			var disc ui.Path
			disc.Circle(cx, cy, size/2)
			p.FillPath(&disc, ui.RGB(24, 24, 27))
			for radius := float32(size/2 - 4); radius > 30; radius -= 2.5 {
				var groove ui.Path
				groove.Circle(cx, cy, radius)
				p.StrokePath(&groove, .6, ui.RGB(63, 63, 70))
			}
			var label ui.Path
			label.Circle(cx, cy, 26)
			p.FillPath(&label, t.Accent)
		})
		ui.Box(c).Absolute().Fill().PassThrough().Draw(func(p *ui.Painter, r ui.Rect) {
			// 33⅓ turns a minute.
			angle := 0.
			if !still {
				angle = math.Mod(p.Now().Sub(time.Time{}).Seconds()*2*math.Pi*100/180, 2*math.Pi)
			}
			cx, cy := r.X+r.W/2, r.Y+r.H/2
			sin, cos := math.Sincos(angle)
			p.Line(cx+float32(cos)*30, cy+float32(sin)*30, cx+float32(cos)*(size/2-6), cy+float32(sin)*(size/2-6), 3, ui.RGB(244, 244, 245))
			switch {
			case still:
			case fps > 0:
				p.AnimationFrameRate(fps)
			default:
				p.AnimationFrame()
			}
		})
	})
}
