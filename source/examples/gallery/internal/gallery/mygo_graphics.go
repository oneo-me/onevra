package gallery

import (
	"math"
	"sync"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/egoist/mygo/ui/shader"
	"github.com/oneo-me/onevra/source/components"
)

// galleryShaders compiles on first use, so pages never shown never start the
// GPU driver's compilation.
type galleryShaders struct {
	plasma, wave, frost *shader.Shader
	plasmaTime          shader.Uniform
	plasmaTint          shader.Uniform
	frostTint           shader.Uniform
}

var loadShaders = sync.OnceValue(func() *galleryShaders {
	s := &galleryShaders{
		plasma: shader.MustCompile(`package main

var Time float
var Tint vec4

func Fragment(dst vec4, pos vec2) vec4 {
	uv := pos / rectSize()
	v := sin(uv.x*9+Time) + sin(uv.y*7-Time*1.3) + sin((uv.x+uv.y)*5+Time*0.7)
	return vec4(mix(vec3(0.12, 0.1, 0.16), Tint.rgb, 0.5+0.5*sin(v)), 1)
}
`),
		wave: shader.MustCompile(`package main

var Time float

func DistortionEffect(pos vec2) vec2 {
	return pos + vec2(0, 4*sin(pos.x/14+Time*3))
}
`),
		frost: shader.MustCompile(`package main

var Tint vec4

func Fragment(dst vec4, pos vec2) vec4 {
	return mix(backdropSample(pos), vec4(Tint.rgb, 1), Tint.a)
}
`),
	}
	s.plasmaTime, s.plasmaTint = s.plasma.Uniform("Time"), s.plasma.Uniform("Tint")
	s.frostTint = s.frost.Uniform("Tint")
	return s
})

// graphicsDemo keeps per-page drawing state: uniform values reused every
// frame, the streaming chart's samples and its reduced points.
type graphicsDemo struct {
	start, last      time.Time
	values, frost    *shader.Values
	accelerated      bool
	shown            float64
	xs               []float32
	ys               [2][]float32
	bucketX, bucketY [2][]float32
	points           [][2]float32
	path             ui.Path
}

func (d *graphicsDemo) elapsed(now time.Time) float32 {
	if d.start.IsZero() {
		d.start = now
	}
	return float32(now.Sub(d.start).Seconds())
}

var spectrum = []ui.GradientStop{
	{Color: ui.RGB(239, 68, 68), At: 0}, {Color: ui.RGB(234, 179, 8), At: .33},
	{Color: ui.RGB(37, 99, 235), At: .66}, {Color: ui.RGB(239, 68, 68), At: 1},
}

func exampleGradients(c *ui.Context, s *demoState) {
	t := c.Theme()
	swatch := func(key, label string) ui.Element {
		return ui.Box(c).Key(key).Size(160, 100).Radius(t.Radius * 2).Label(label)
	}
	switch s.variant {
	case "Conic":
		ui.Row(c).Gap(t.Space(4)).Wrap().Children(func() {
			swatch("conic", tr(c, "gradientConic")).ConicGradient(ui.ConicGradient{Stops: spectrum, Angle: float32(s.value), Oklab: s.on})
			swatch("conic-offset", tr(c, "gradientConicOffset")).
				ConicGradient(ui.ConicGradient{Stops: spectrum, Angle: float32(s.value), Center: ui.Point{X: .25, Y: .25}, Oklab: s.on})
		})
		gradientAngle(c, s)
		components.Checkbox(c, &s.on, tr(c, "gradientOklab"))
	case "Borders":
		ui.Row(c).Gap(t.Space(4)).Wrap().Children(func() {
			swatch("border-conic", tr(c, "gradientBorderConic")).Border(4, ui.Color{}).
				BorderConicGradient(ui.ConicGradient{Stops: spectrum, Angle: float32(s.value), Oklab: true})
			swatch("border-linear", tr(c, "gradientBorderLinear")).Border(4, ui.Color{}).
				BorderLinearGradient(ui.LinearGradient{Stops: spectrum, Angle: float32(s.value), Oklab: true})
		})
		gradientAngle(c, s)
	case "Blur":
		ui.Box(c).Size(220, 120).Padding(20).Children(func() {
			// A glow is a sharp border over blurred copies of itself.
			for i, layer := range []struct{ width, blur float32 }{{12, 10}, {6, 3}, {2, 0}} {
				ui.Box(c).Key(i).Absolute().Top(20).Left(20).Right(20).Bottom(20).PassThrough().
					Radius(16).Border(layer.width, ui.Color{}).
					BorderConicGradient(ui.ConicGradient{Stops: spectrum, Oklab: true}).Blur(layer.blur)
			}
			ui.Box(c).Fill().Center().Children(func() {
				components.Text(c, tr(c, "gradientGlow")).Bold()
			})
		})
		ui.Box(c).Size(220, 60).Padding(12).Radius(t.Radius).Background(t.Surface).Blur(float32(s.value) / 10).Children(func() {
			components.Text(c, tr(c, "gradientBlurred"))
		})
		ui.Row(c).FillWidth().MaxWidth(t.Space(100)).Gap(t.Space(3)).AlignItems(ui.Center).Children(func() {
			name := tr(c, "gradientBlur")
			components.Text(c, name).Shrink(0).SingleLine()
			components.Slider(c, &s.value, 0, 80).Label(name).Grow(1).MinWidth(0)
		})
	case "Radial":
		ui.Row(c).Gap(t.Space(4)).Wrap().Children(func() {
			// A glow fades from a point to nothing; a painted spotlight lights
			// a surface from its own center.
			swatch("radial", tr(c, "gradientRadial")).Background(t.Surface).RadialGradient(ui.RadialGradient{
				Stops:  []ui.GradientStop{{Color: t.Accent.Alpha(.8), At: 0}, {Color: ui.Transparent, At: 1}},
				Radius: float32(s.value) + 20, Oklab: s.on,
			})
			swatch("spotlight", tr(c, "gradientSpotlight")).Draw(func(p *ui.Painter, r ui.Rect) {
				p.FillRadialGradient(r, ui.RadialGradient{
					Stops:  []ui.GradientStop{{Color: ui.RGB(253, 224, 71), At: 0}, {Color: ui.RGB(234, 88, 12), At: .5}, {Color: ui.RGB(30, 27, 75), At: 1}},
					Center: ui.Point{X: .3, Y: .3}, Radius: float32(s.value) + 20, Oklab: s.on,
				}, t.Radius*2)
			})
		})
		gradientRadius(c, s)
		components.Checkbox(c, &s.on, tr(c, "gradientOklab"))
	default:
		ui.Row(c).Gap(t.Space(4)).Wrap().Children(func() {
			swatch("linear", tr(c, "gradientStops")).LinearGradient(ui.LinearGradient{Stops: spectrum, Angle: 90 + float32(s.value), Oklab: s.on})
		})
		gradientAngle(c, s)
		components.Checkbox(c, &s.on, tr(c, "gradientOklab"))
	}
}

func gradientRadius(c *ui.Context, s *demoState) {
	ui.Row(c).FillWidth().MaxWidth(c.Theme().Space(100)).Gap(c.Theme().Space(3)).AlignItems(ui.Center).Children(func() {
		name := tr(c, "gradientRadius")
		components.Text(c, name).Shrink(0).SingleLine()
		components.Slider(c, &s.value, 0, 160).Label(name).Grow(1).MinWidth(0)
	})
}

func gradientAngle(c *ui.Context, s *demoState) {
	ui.Row(c).FillWidth().MaxWidth(c.Theme().Space(100)).Gap(c.Theme().Space(3)).AlignItems(ui.Center).Children(func() {
		name := tr(c, "gradientAngle")
		components.Text(c, name).Shrink(0).SingleLine()
		components.Slider(c, &s.value, 0, 360).Label(name).Grow(1).MinWidth(0)
	})
}

func exampleShaders(c *ui.Context, s *demoState) {
	t := c.Theme()
	sh, d := loadShaders(), &s.graphics
	if d.values == nil {
		d.values, d.frost = sh.plasma.NewValues(), sh.frost.NewValues()
	}
	switch s.variant {
	case "Text and paths":
		ui.Box(c).Size(320, 110).Label(tr(c, "shaderFillLabel")).Draw(func(p *ui.Painter, r ui.Rect) {
			o := shader.Options{Values: d.values.Float(sh.plasmaTime, d.elapsed(p.Now())).Color(sh.plasmaTint, t.Accent)}
			var circle ui.Path
			circle.Circle(r.X+50, r.Y+55, 44)
			sh.plasma.FillPath(p, &circle, o)
			sh.plasma.Text(p, r.X+110, r.Y+30, "MyGo", 48, o)
			p.AnimationFrame()
		})
	case "Element effect":
		box := ui.Box(c).Padding(t.Space(4)).Radius(t.Radius).Background(t.Surface).Draw(func(p *ui.Painter, _ ui.Rect) {
			// The CPU renderer interprets an effect pixel by pixel, a
			// second or so a frame, so only a GPU draws the wave.
			if p.Accelerated() != d.accelerated {
				d.accelerated = p.Accelerated()
				c.Invalidate()
			}
		})
		if d.accelerated {
			box.Effect(sh.wave.Effect(shader.Options{
				Uniforms:        map[string]any{"Time": d.elapsed(c.Now())},
				MaxSampleOffset: ui.Point{Y: 4},
			}))
			c.AnimationFrame()
		}
		box.Children(func() {
			components.Text(c, tr(c, "shaderWave")).FontSize(24).Bold()
			components.Button(c, tr(c, "shaderStillClickable"), components.ButtonOutline)
		})
	case "Backdrop":
		ui.Box(c).Size(320, 160).Radius(t.Radius * 2).Clip().
			ConicGradient(ui.ConicGradient{Stops: spectrum, Oklab: true}).Children(func() {
			ui.Column(c).Padding(t.Space(4)).Gap(t.Space(1)).Children(func() {
				for _, key := range []string{"shaderBackdropLine1", "shaderBackdropLine2", "shaderBackdropLine3"} {
					components.Text(c, tr(c, key)).Bold().TextColor(ui.RGB(255, 255, 255))
				}
			})
			pane := ui.Box(c).Absolute().Top(30).Left(30).Size(140, 100).Cursor(ui.CursorGrab).Label(tr(c, "shaderFrost"))
			pane.Draggable(&s.offset, ui.DraggableOptions{Bounds: &ui.Rect{W: 140, H: 20}}).Draw(func(p *ui.Painter, r ui.Rect) {
				sh.frost.Draw(p, r, shader.Options{Values: d.frost.Color(sh.frostTint, t.Surface.Alpha(.35)), BackdropBlur: float32(s.value) / 4, Radius: 16})
			})
		})
		ui.Row(c).FillWidth().MaxWidth(t.Space(100)).Gap(t.Space(3)).AlignItems(ui.Center).Children(func() {
			name := tr(c, "gradientBlur")
			components.Text(c, name).Shrink(0).SingleLine()
			components.Slider(c, &s.value, 0, 80).Label(name).Grow(1).MinWidth(0)
		})
	default:
		ui.Box(c).Size(320, 140).Label(tr(c, "shaderPlasma")).Draw(func(p *ui.Painter, r ui.Rect) {
			sh.plasma.Draw(p, r, shader.Options{
				Values: d.values.Float(sh.plasmaTime, d.elapsed(p.Now())).Color(sh.plasmaTint, t.Accent),
				Radius: t.Radius * 2,
			})
			p.AnimationFrame()
		})
	}
}

const streamSamples = 4000

// streamValue is a deterministic waveform with isolated spikes, which MinMax
// keeps however narrow the plot is.
func streamValue(i int64, series int) float32 {
	x := float64(i) / 400
	v := .18*math.Sin(x*float64(series+2)*3) + .06*math.Sin(x*41)
	if i%613 == int64(series*97) {
		v += .2
	}
	return float32(.3 + float64(series)*.4 + v)
}

// exampleLiveChart reduces two rolling series to each pixel column's extremes,
// drawing one with StrokePolyline and the other through a reused path.
func exampleLiveChart(c *ui.Context, s *demoState) {
	t, d := c.Theme(), &s.graphics
	colors := [2]ui.Color{t.Accent, t.Success}
	if d.xs == nil {
		d.xs = make([]float32, streamSamples)
		for i := range d.xs {
			d.xs[i] = float32(i) / (streamSamples - 1)
		}
	}
	x, y := chart.NewLinear(-float64(streamSamples)/400, 0), chart.NewLinear(0, 1)
	chart.Plot(c, chart.Axes{Left: y.Ticks(5), Bottom: x.Ticks(5)}, func(area ui.Element) {
		area.Label(tr(c, "liveChartLabel")).Draw(func(p *ui.Painter, r ui.Rect) {
			// The series advances with the frame clock, and stays where it
			// is while paused.
			now := p.Now()
			if s.on {
				d.last = time.Time{}
			} else {
				if !d.last.IsZero() {
					d.shown += now.Sub(d.last).Seconds() * 400
				}
				d.last = now
				p.AnimationFrame()
			}
			next := int64(d.shown)
			for _, tick := range y.Ticks(5) {
				p.Line(r.X, chart.Y(r, tick.At), r.X+r.W, chart.Y(r, tick.At), 1, t.Border)
			}
			p.Clip(r, 0, func() {
				for k := range 2 {
					d.ys[k] = d.ys[k][:0]
					for i := range int64(streamSamples) {
						d.ys[k] = append(d.ys[k], streamValue(next+i, k))
					}
					d.bucketX[k], d.bucketY[k] = chart.MinMax(d.bucketX[k], d.bucketY[k], d.xs, d.ys[k], int(math.Ceil(float64(r.W*p.Scale()))))
				}
				d.points = d.points[:0]
				for i, px := range d.bucketX[0] {
					d.points = append(d.points, [2]float32{chart.X(r, px), chart.Y(r, d.bucketY[0][i])})
				}
				p.StrokePolyline(d.points, 2, colors[0])
				p.StrokePath(chart.LineInto(&d.path, r, d.bucketX[1], d.bucketY[1], chart.CurveLinear), 2, colors[1])
			})
		})
	}).Height(220).FillWidth().Gap(8)
	components.Checkbox(c, &s.on, tr(c, "timelinePause"))
}
