package gallery

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/audio"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

const audioRate = 48000

// audioNotes are the pluck's playback speeds relative to its A4 recording.
var audioNotes = []struct {
	name  string
	speed float64
}{{"C4", 261.63 / 440}, {"E4", 329.63 / 440}, {"G4", 392.00 / 440}, {"C5", 523.25 / 440}}

// audioDemo synthesizes its sounds in memory on first use, so the page ships
// no resources. Output is false in headless views, which keep the controls'
// state without opening an audio device. ended counts the notes heard to
// their end. mediaControls shows the loop where the system shows media, and
// release stops the system driving it.
type audioDemo struct {
	output        bool
	pluck, pad    *audio.Sound
	loop          *audio.Player
	looping       bool
	played        int
	ended         atomic.Int32
	volume, pan   float64
	playback      bool
	mediaControls bool
	release       func()
	stopCommands  func()
	window        func() *mygo.Window
}

func (d *audioDemo) sounds() {
	if d.pluck == nil {
		d.pluck, d.pad = audio.NewSound(audioRate, 1, pluckSamples()), audio.NewSound(audioRate, 1, loopSamples())
	}
}

func (d *audioDemo) play(speed float64) {
	d.played++
	if !d.output {
		return
	}
	d.sounds()
	player := d.pluck.NewPlayer().SetSpeed(speed).SetVolume(d.volume).SetPan(d.pan)
	player.OnEnd(func() {
		d.ended.Add(1)
		d.refresh()
	})
	player.Play()
}

// refresh shows changes made off the page's frames, as a note's end or a
// command of the system's media controls.
func (d *audioDemo) refresh() {
	if d.window == nil {
		return
	}
	if w := d.window(); w != nil {
		w.Update(func() {
			if d.loop != nil {
				d.looping = d.loop.Playing()
			}
			w.Invalidate()
		})
	}
}

// setLatency buffers further ahead for music, which plays on through the
// moments the system keeps the app waiting.
func (d *audioDemo) setLatency() {
	if !d.output {
		return
	}
	if d.playback {
		audio.SetLatency(audio.LatencyPlayback)
	} else {
		audio.SetLatency(audio.LatencyInteractive)
	}
}

// showMedia lets the system's media controls and the media keys play and
// pause the loop, while the option is on and the loop exists.
func (d *audioDemo) showMedia() {
	if !d.output || d.loop == nil {
		return
	}
	if !d.mediaControls {
		d.hideMedia()
		return
	}
	if d.release != nil {
		return
	}
	d.release = mygo.NowPlaying.Control(d.loop)
	d.stopCommands = mygo.NowPlaying.OnCommand(func(mygo.MediaCommand, time.Duration) { d.refresh() })
	mygo.NowPlaying.Set(mygo.NowPlayingInfo{Title: "A major pad", Artist: "Onevra Gallery"})
}

func (d *audioDemo) hideMedia() {
	if d.release == nil {
		return
	}
	d.release()
	d.stopCommands()
	d.release, d.stopCommands = nil, nil
	mygo.NowPlaying.Clear()
}

func (d *audioDemo) toggleLoop() {
	d.looping = !d.looping
	if !d.output {
		return
	}
	if d.loop == nil {
		d.sounds()
		// The swell-in plays once; the steady pad after it repeats.
		d.loop = d.pad.NewPlayer().SetLoopRange(padDuration, 0)
	}
	if d.looping {
		d.loop.SetVolume(d.volume).SetPan(d.pan).Play()
	} else {
		d.loop.Pause()
	}
	d.showMedia()
	mygo.NowPlaying.Sync()
}

// mix applies the shared volume and pan to the loop as the sliders move.
func (d *audioDemo) mix() {
	if d.loop != nil {
		d.loop.SetVolume(d.volume).SetPan(d.pan)
	}
}

func (d *audioDemo) close() {
	d.hideMedia()
	if d.loop != nil {
		d.loop.Stop()
	}
	d.loop, d.looping = nil, false
}

// pluckSamples is a decaying A4 with two harmonics, about 0.8 seconds long.
func pluckSamples() []float32 {
	samples := make([]float32, audioRate*8/10)
	for i := range samples {
		t := float64(i) / audioRate
		v := math.Sin(2*math.Pi*440*t) + .4*math.Sin(2*math.Pi*880*t) + .15*math.Sin(2*math.Pi*1320*t)
		attack := min(1, t/.005)
		samples[i] = float32(.35 * v * attack * math.Exp(-5*t))
	}
	return samples
}

const padDuration = 2 * time.Second

// loopSamples is a fading-in intro followed by the pad. The intro is one
// cycle of the pad itself, so it joins the loop without a seam.
func loopSamples() []float32 {
	pad := padSamples()
	intro := make([]float32, len(pad))
	for i, v := range pad {
		intro[i] = v * float32(i) / float32(len(pad))
	}
	return append(intro, pad...)
}

// padSamples is two seconds of a soft A-major chord. Every partial and the
// swell complete whole cycles, so the loop has no seam.
func padSamples() []float32 {
	samples := make([]float32, int(audioRate*padDuration.Seconds()))
	for i := range samples {
		t := float64(i) / audioRate
		v := math.Sin(2*math.Pi*220*t) + .8*math.Sin(2*math.Pi*277*t) + .7*math.Sin(2*math.Pi*330*t)
		swell := .75 + .25*math.Sin(2*math.Pi*.5*t)
		samples[i] = float32(.12 * v * swell)
	}
	return samples
}

func exampleAudio(c *ui.Context, s *demoState) {
	t := c.Theme()
	d := &s.audio
	components.Text(c, tr(c, "audioNotes")).Bold().TextColor(t.TextMuted)
	ui.Row(c).Gap(t.Space(2)).Wrap().Children(func() {
		for _, note := range audioNotes {
			if components.Button(c, note.name, components.ButtonSecondary).Key(note.name).Clicked() {
				d.play(note.speed)
			}
		}
	})
	label := tr(c, "audioPlayLoop")
	if d.looping {
		label = tr(c, "audioPauseLoop")
	}
	ui.Row(c).Children(func() {
		if components.Button(c, label, components.ButtonPrimary).Key("loop").Clicked() {
			d.toggleLoop()
		}
	})
	components.Text(c, msg(c).AudioNotesEnded(int(d.ended.Load()))).TextColor(t.TextMuted)
	if components.Checkbox(c, &d.playback, tr(c, "audioPlaybackLatency")).Changed() {
		d.setLatency()
	}
	if components.Checkbox(c, &d.mediaControls, tr(c, "audioMediaControls")).Changed() {
		d.showMedia()
	}
	changed := false
	for _, slider := range []struct {
		id     string
		value  *float64
		lo, hi float64
	}{{"volume", &d.volume, 0, 1}, {"audioPan", &d.pan, -1, 1}} {
		ui.Row(c).Key(slider.id).FillWidth().MaxWidth(t.Space(100)).Gap(t.Space(3)).AlignItems(ui.Center).Children(func() {
			name := tr(c, slider.id)
			components.Text(c, name).Width(t.Space(14)).Shrink(0).SingleLine()
			control := components.Slider(c, slider.value, slider.lo, slider.hi).Label(name)
			control.Grow(1).MinWidth(0)
			changed = changed || control.Changed()
		})
	}
	if changed {
		d.mix()
	}
	if err := audio.Err(); d.output && err != nil {
		components.Text(c, msg(c).AudioUnavailable(err.Error())).TextColor(t.Danger).Wrap()
	}
}
