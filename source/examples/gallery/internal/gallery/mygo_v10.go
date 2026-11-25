package gallery

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net"
	"net/http"
	"sync"

	"github.com/egoist/mygo/plugins/render"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/netimage"
	"github.com/oneo-me/onevra/source/components"
)

// exampleSplitButton shows MyGo's SplitButton with both system and drawn menus
// as well as disabling the primary segment independently of its menu arrow.
func exampleSplitButton(c *ui.Context, s *demoState) {
	t := c.Theme()
	ui.Row(c).Gap(t.Space(3)).Wrap().AlignItems(ui.Center).Children(func() {
		save := ui.SplitButton(c, tr(c, "secretsSave"))
		save.Arrow.Menu(func(m *ui.Menu) {
			for _, action := range []string{"open", "duplicate"} {
				if m.Item(tr(c, action)).Chosen() {
					s.choice, s.count = action, s.count+1
				}
			}
		})
		if save.Button.Clicked() {
			s.choice, s.count = "secretsSave", s.count+1
		}

		drawn := ui.SplitButton(c, tr(c, "open"))
		drawn.Button.Disabled(s.on)
		drawn.Arrow.PopoverMenuButton(func(m *ui.Menu) {
			for _, action := range []string{"duplicate", "delete"} {
				if m.Item(tr(c, action)).Chosen() {
					s.choice, s.count = action, s.count+1
				}
			}
		}, ui.PopoverMenuButtonOptions{At: ui.AnchorBottomEnd, Self: ui.AnchorTopEnd})
		if drawn.Button.Clicked() {
			s.choice, s.count = "open", s.count+1
		}
	})
	components.Checkbox(c, &s.on, tr(c, "disabled"))
	if s.count > 0 {
		ui.Text(c, msg(c).LabelCount(tr(c, s.choice), s.count))
	} else {
		components.Text(c, tr(c, "noActionYet")).TextColor(t.TextMuted)
	}
}

// examplePrimarySplitButton shows MyGo's accented PrimarySplitButton with a
// drawn popover menu aligned to its trailing edge.
func examplePrimarySplitButton(c *ui.Context, s *demoState) {
	t := c.Theme()
	run := ui.PrimarySplitButton(c, tr(c, "apply"))
	run.Arrow.PopoverMenuButton(func(m *ui.Menu) {
		for _, action := range []string{"open", "duplicate"} {
			if m.Item(tr(c, action)).Chosen() {
				s.choice, s.count = action, s.count+1
			}
		}
	}, ui.PopoverMenuButtonOptions{At: ui.AnchorBottomEnd, Self: ui.AnchorTopEnd})
	if run.Button.Clicked() {
		s.choice, s.count = "apply", s.count+1
	}
	if s.count > 0 {
		ui.Text(c, msg(c).LabelCount(tr(c, s.choice), s.count))
	} else {
		components.Text(c, tr(c, "noActionYet")).TextColor(t.TextMuted)
	}
}

// exampleNativeView demonstrates ui.NativeView with a nil PlatformView, which
// renders as a plain box so fallback content and overlays stay visible when a
// native module has no platform view.
func exampleNativeView(c *ui.Context, _ *demoState) {
	t := c.Theme()
	ui.NativeView(c, nil).Size(t.Space(75), t.Space(30)).Radius(t.Radius*2).Background(t.Surface).Border(1, t.Border).Center().
		Label("NativeView").Children(func() {
		components.Text(c, "NativeView (nil fallback)").TextColor(t.TextMuted)
	})
}

// renderDemo owns the preview player created for the Render plugin example.
type renderDemo struct {
	player *ui.VideoPlayer
	err    error
}

func (d *renderDemo) close() {
	if d.player != nil {
		d.player.Close()
		d.player = nil
	}
}

func renderComposition() render.Composition {
	return render.Composition{
		Width: 320, Height: 180, FPS: 30, Frames: 90, Scale: 2,
		View: func(c *ui.Context, f render.Frame) {
			t := c.Theme()
			f.Sequence(c, 0, 90, func(f render.Frame) {
				s := f.Spring(ui.Snappy.Spring)
				fade := float32(f.Interpolate([]float64{0, 15}, []float64{0, 1}))
				ui.Column(c).Fill().AlignItems(ui.Center).Justify(ui.Center).Gap(6).Children(func() {
					ui.Text(c, "MyGo Render").FontSize(20).Bold().TextColor(t.Accent).
						Opacity(fade).Translate(0, float32(20*(1-s)))
					ui.Box(c).Size(float32(60+40*s), 6).Radius(3).Background(t.Accent.Alpha(.6))
				})
			})
		},
	}
}

// exampleRender previews a plugins/render Composition in a ui.Video element.
func exampleRender(c *ui.Context, s *demoState) {
	t := c.Theme()
	d := &s.render
	if d.player == nil && d.err == nil {
		d.player, d.err = render.NewPlayer(renderComposition())
		if d.err == nil {
			d.player.SetLoop(true)
		}
	}
	if d.err != nil {
		components.Text(c, msg(c).SystemError(d.err.Error())).TextColor(t.TextMuted)
		return
	}
	ui.Video(c, d.player).Size(320, 180).Radius(t.Radius * 2).Fit(ui.Cover).Label("Render")
	label := tr(c, "timelinePause")
	if !d.player.Playing() {
		label = tr(c, "timelinePlay")
	}
	if components.Button(c, label, components.ButtonOutline).Clicked() {
		if d.player.Playing() {
			d.player.Pause()
		} else {
			d.player.Play()
		}
	}
}

// netImageDemo serves an in-memory PNG over loopback only after the user asks
// to load it, so headless views and offline runs never contact the network.
type netImageDemo struct {
	server *http.Server
	url    string
	err    string
}

var sampleNetPNG = sync.OnceValue(func() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 160, 100))
	for y := range 100 {
		for x := range 160 {
			img.SetRGBA(x, y, color.RGBA{uint8(40 + x), uint8(100 + y), 220, 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		panic(err)
	}
	return b.Bytes()
})

func (d *netImageDemo) start() {
	if d.server != nil {
		return
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		d.err = err.Error()
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/sample.png", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=60")
		_, _ = w.Write(sampleNetPNG())
	})
	d.server = &http.Server{Handler: mux}
	d.url = "http://" + ln.Addr().String() + "/sample.png"
	d.err = ""
	go func() { _ = d.server.Serve(ln) }()
}

func (d *netImageDemo) close() {
	if d.server != nil {
		_ = d.server.Shutdown(context.Background())
		d.server = nil
	}
	if d.url != "" {
		netimage.Forget(d.url)
		d.url = ""
	}
}

// exampleNetImage demonstrates ui/netimage with a placeholder, local in-memory
// HTTP source and Forget cache reset.
func exampleNetImage(c *ui.Context, s *demoState) {
	t := c.Theme()
	d := &s.netImage
	ui.Row(c).Gap(t.Space(2)).Wrap().Children(func() {
		if components.Button(c, tr(c, "open"), components.ButtonPrimary).Clicked() {
			d.start()
		}
		if components.Button(c, tr(c, "clear"), components.ButtonOutline).Disabled(d.url == "").Clicked() {
			netimage.Forget(d.url)
		}
	})
	if d.url == "" {
		ui.Box(c).Size(160, 100).Radius(t.Radius*2).Background(t.Surface).Border(1, t.Border).Center().Label("Network image").
			Children(func() { components.Text(c, tr(c, "open")).TextColor(t.TextMuted) })
	} else {
		netimage.Image(c, d.url, netimage.Options{Placeholder: t.Surface, Fit: ui.Cover}).
			Size(160, 100).Radius(t.Radius * 2).Label("Network image")
	}
	if d.err != "" {
		components.Text(c, msg(c).SystemError(d.err)).TextColor(t.Danger)
	}
}
