package gallery

import (
	"testing"
	"time"
)

func TestAudioControlsWithoutOutput(t *testing.T) {
	for _, lang := range []string{"en", "zh-Hans"} {
		for _, width := range []int{280, 540} {
			g, s, tt := featureTester("Audio", "Default", lang, width)
			if s.audio.output || s.audio.volume != .8 {
				t.Fatalf("headless audio state: %+v", &s.audio)
			}
			for _, note := range audioNotes {
				if err := tt.Click(note.name); err != nil {
					t.Fatal(err)
				}
			}
			if s.audio.played != len(audioNotes) || s.audio.pluck != nil {
				t.Fatalf("headless notes: played %d, synthesized %v", s.audio.played, s.audio.pluck != nil)
			}
			if err := tt.Click(g.text("audioPlayLoop")); err != nil {
				t.Fatal(err)
			}
			if !s.audio.looping || s.audio.loop != nil {
				t.Fatalf("loop toggled without a player: %+v", &s.audio)
			}
			if _, ok := tt.Find(g.text("audioPauseLoop")); !ok {
				t.Fatal("missing pause action")
			}
			// Headless views keep the options without reaching the system.
			for _, label := range []string{"audioPlaybackLatency", "audioMediaControls"} {
				if err := tt.Click(g.text(label)); err != nil {
					t.Fatal(err)
				}
			}
			if !s.audio.playback || !s.audio.mediaControls || s.audio.release != nil {
				t.Fatalf("headless options: playback %v, media controls %v, controlled %v",
					s.audio.playback, s.audio.mediaControls, s.audio.release != nil)
			}
			pan, ok := tt.Find(g.text("audioPan"))
			if !ok {
				t.Fatal("missing pan label")
			}
			// Press near the end of the slider that follows the label.
			x, y := pan.X+min(float32(width)-32, 400)*.95, pan.Y+pan.H/2
			tt.Press(x, y)
			tt.Release(x, y)
			if s.audio.pan <= 0 {
				t.Fatalf("pan did not move right: %v", s.audio.pan)
			}
			s.audio.close()
			if s.audio.looping {
				t.Fatal("close kept the loop playing")
			}
		}
	}
}

func TestAudioSamplesLoopWithoutSeam(t *testing.T) {
	pad := padSamples()
	if d := pad[len(pad)-1] - pad[0]; d > .01 || d < -.01 {
		t.Fatalf("loop seam jumps by %v", d)
	}
	loop := loopSamples()
	intro := len(loop) - len(pad)
	if loop[0] != 0 || loop[intro] != pad[0] || abs32(loop[intro-1]-pad[len(pad)-1]) > .01 {
		t.Fatalf("intro does not join the loop: starts %v, joins %v -> %v", loop[0], loop[intro-1], loop[intro])
	}
	if time.Duration(intro)*time.Second/audioRate != padDuration {
		t.Fatalf("loop range starts at %d samples", intro)
	}
	pluck := pluckSamples()
	if pluck[0] != 0 || abs32(pluck[len(pluck)-1]) > .01 {
		t.Fatalf("pluck clicks: starts %v, ends %v", pluck[0], pluck[len(pluck)-1])
	}
}

func abs32(v float32) float32 { return max(v, -v) }
