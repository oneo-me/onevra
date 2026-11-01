package gallery

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"math"
	"sync"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// videoDemo holds the page's player, made once, never in the view.
type videoDemo struct {
	player *ui.VideoPlayer
	err    error
}

func (d *videoDemo) close() {
	if d.player != nil {
		d.player.Close()
	}
}

// wavesGIF is an animated GIF drawn in memory on first use: a band of color
// rolling across, frame by frame.
var wavesGIF = sync.OnceValue(func() []byte {
	const w, h, frames = 160, 90, 24
	palette := color.Palette{}
	for i := range 64 {
		v := float64(i) / 63
		palette = append(palette, color.RGBA{uint8(37 + v*200), uint8(99 + v*80), uint8(235 - v*150), 255})
	}
	anim := &gif.GIF{}
	for f := range frames {
		img := image.NewPaletted(image.Rect(0, 0, w, h), palette)
		phase := float64(f) / frames * 2 * math.Pi
		for y := range h {
			for x := range w {
				v := (math.Sin(float64(x)/18+phase) + math.Sin(float64(y)/12-phase)) / 4
				img.SetColorIndex(x, y, uint8((v+.5)*63))
			}
		}
		anim.Image = append(anim.Image, img)
		anim.Delay = append(anim.Delay, 6)
	}
	var b bytes.Buffer
	if err := gif.EncodeAll(&b, anim); err != nil {
		panic(err)
	}
	return b.Bytes()
})

// exampleVideo plays an animated GIF in native UI, looping; it shows its
// first frame and stays still while the desktop asks for less motion.
func exampleVideo(c *ui.Context, s *demoState) {
	t := c.Theme()
	d := &s.video
	if d.player == nil && d.err == nil {
		var dec ui.VideoDecoder
		if dec, d.err = ui.NewGIFDecoder(wavesGIF()); d.err == nil {
			if d.player, d.err = ui.NewVideoPlayer(dec); d.err == nil {
				d.player.SetLoop(true)
			}
		}
	}
	if d.err != nil {
		components.Text(c, msg(c).SystemError(d.err.Error())).TextColor(t.TextMuted)
		return
	}
	ui.Video(c, d.player).Size(320, 180).Radius(t.Radius * 2).Fit(ui.Cover).Label(tr(c, "videoWaves"))
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
