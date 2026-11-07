package gallery

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// sharedCover is what the tile and the page of the Fly variant take over from
// each other.
type sharedCover struct{}

// exampleSharedMotion flies shared elements between layouts, morphs text and
// times what enters and leaves.
func exampleSharedMotion(c *ui.Context, s *demoState) {
	t := c.Theme()
	toggle := components.Button(c, tr(c, "transformFlip"), components.ButtonOutline)
	if toggle.Clicked() {
		s.on = !s.on
	}
	spring := ui.Smooth
	switch s.variant {
	case "Morph":
		text := tr(c, "morphShort")
		if s.on {
			text = tr(c, "morphLong")
		}
		components.Text(c, text).FontSize(t.FontSize * 2).Bold().ContentTransition(ui.ContentTransition{Morph: true})
	case "Exits":
		ui.Row(c).Gap(t.Space(2)).Height(t.Space(16)).Children(func() {
			fade := &ui.Motion{Blur: 6, YPercent: 50, Opacity: 0}
			if !s.on {
				ui.Box(c).Key("first").Size(t.Space(16), t.Space(16)).Radius(t.Radius).Background(t.Accent).Transition(ui.ElementTransition{
					Enter: fade, Exit: fade, ExitDuration: 400 * time.Millisecond, ExitEase: ui.EaseInOut, Shadows: true})
			} else {
				ui.Box(c).Key("second").Size(t.Space(16), t.Space(16)).Radius(t.Radius).Background(t.Success).Transition(ui.ElementTransition{
					Enter: fade, Exit: &ui.Motion{Flip: true, Scale: .8, Opacity: 0}, AfterExits: true})
			}
		}).ChildExit(ui.Motion{Blur: 2, Opacity: 0})
		box := ui.Box(c).Size(t.Space(40), t.Space(16)).Radius(t.Radius).Border(1, t.Border).Center()
		if s.on {
			box.Width(t.Space(60)).Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .3))
		}
		box.Transition(ui.ElementTransition{Width: true, Shadows: true, Animation: &spring}).Children(func() {
			components.Text(c, "Width, Shadows")
		})
	case "Color":
		for _, oklab := range []bool{false, true} {
			target := ui.Hex("#2563eb")
			if s.on {
				target = ui.Hex("#facc15")
			}
			ui.Box(c).Key(oklab).Size(t.Space(40), t.Space(12)).Radius(t.Radius).Background(target).
				Transition(ui.ElementTransition{Colors: true, Animation: &ui.Animation{Duration: time.Second, Oklab: oklab}})
		}
		ui.Box(c).Size(t.Space(40), t.Space(12)).Radius(t.Radius).Background(ui.Hex("#2563eb").MixOklab(ui.Hex("#facc15"), .5)).Label("MixOklab")
		box := ui.Box(c).Size(t.Space(40), t.Space(12)).Radius(t.Radius).Background(t.Accent).Label("NonMotion")
		target := float32(1)
		if s.on {
			target = .2
		}
		box.Opacity(box.AnimateTo("fade", target, ui.Animation{Duration: 400 * time.Millisecond, NonMotion: true}))
		box2 := ui.Box(c).Size(t.Space(20), t.Space(12)).Radius(t.Radius).Background(t.Success).Label("TranslatePercent")
		if s.on {
			box2.TranslatePercent(50, 0)
		}
	default:
		fly := ui.ElementTransition{Match: sharedCover{}, Fly: true, Animation: &spring}
		if s.on {
			ui.Box(c).Key("page").FillWidth().MaxWidth(t.Space(100)).Height(t.Space(48)).Radius(t.Radius * 3).Background(t.Accent).Transition(fly)
			components.Text(c, tr(c, "morphLong")).Transition(ui.ElementTransition{Match: "title", Fly: true}).ContentTransition(ui.ContentTransition{Morph: true})
		} else {
			ui.Box(c).Key("tile").Size(t.Space(24), t.Space(24)).Radius(t.Radius).Background(t.Accent).Transition(fly)
			components.Text(c, tr(c, "morphShort")).Transition(ui.ElementTransition{Match: "title", Fly: true}).ContentTransition(ui.ContentTransition{Morph: true})
		}
	}
}

// exampleRouterMatch fades a router's pages into each other while their
// shared elements fly.
func exampleRouterMatch(c *ui.Context, s *demoState) {
	t := c.Theme()
	if s.router == nil {
		s.router = ui.NewRouter("/")
		s.router.Transition = ui.TransitionMatch
	}
	fly := ui.ElementTransition{Match: sharedCover{}, Fly: true}
	s.router.View(c, func(r *ui.Route) {
		size, label, target := t.Space(24), "morphShort", "/page"
		if r.Match("/page") {
			size, label, target = t.Space(48), "morphLong", "/"
		}
		if components.Button(c, tr(c, label), components.ButtonOutline).Clicked() {
			s.router.Push(target)
		}
		ui.Box(c).Size(size, size).Radius(t.Radius * 2).Background(t.Accent).Transition(fly)
	}).FillWidth().Height(t.Space(64)).Gap(t.Space(2))
}
