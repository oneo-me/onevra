package components

import (
	"math"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/oneo-me/onevra/source/internal/l10n"
)

// CardDecision is the side a card of a CardStack is sent to.
type CardDecision int

const (
	// CardLeft passes on a card, by its left action.
	CardLeft CardDecision = iota + 1
	// CardRight keeps it, by its right, primary action.
	CardRight
)

// CardStackItem is one card of a CardStack.
type CardStackItem struct {
	// ID identifies the card in CardStackState.Decisions. Keep IDs stable and
	// unique.
	ID string
	// Label is a short name for the card, such as "Lisbon", in its accessible
	// name and announcements.
	Label string
	// Content builds the card's face, which fills it. The stack takes the
	// pointer over the top card for its drag, so the content is presentation.
	// The card does not clip it: a clip that moves and turns with a drag
	// would take a mask rasterized anew every frame. Content reaching the
	// card's edges rounds its own corners to radius, the card's.
	Content func(c *ui.Context, radius float32)
}

// CardStackDecision is a card decided, by its ID.
type CardStackDecision struct {
	ID       string
	Decision CardDecision
}

// CardStackState is the application-owned progress of a CardStack.
type CardStackState struct {
	// Decisions are the cards decided so far, in order; the next card on top
	// is Items[len(Decisions)]. Appending, removing or clearing them, here or
	// by the stack's own input, throws cards off or brings them back.
	Decisions []CardStackDecision
}

// CardStackOptions configure a CardStack.
type CardStackOptions struct {
	Items []CardStackItem
	// LeftLabel and RightLabel name the actions on the buttons and on the
	// stamps a dragged card shows: "Pass" and "Keep" from Onevra's catalog
	// when empty.
	LeftLabel, RightLabel string
	// LeftOutcome and RightOutcome, the past tense of each action, such as
	// "skipped" and "shortlisted", phrase announcements; without them they
	// name the action.
	LeftOutcome, RightOutcome string
	// Label is the accessible name of the stack: "Cards" when empty.
	Label string
	// Empty, if set, builds what shows once every card is decided, given a
	// function that brings all cards back, in place of the default title and
	// Start over button.
	Empty func(c *ui.Context, reset func())
}

// CardStackAction is what the stack's own input did in a frame.
type CardStackAction int

const (
	CardStackNone CardStackAction = iota
	// CardStackDecided decided the card ID.
	CardStackDecided
	// CardStackUndone brought the card ID back with its Decision undone.
	CardStackUndone
	// CardStackReset brought every card back.
	CardStackReset
)

// CardStackEvent reports a change the stack's input made to its state.
type CardStackEvent struct {
	Action   CardStackAction
	ID       string
	Decision CardDecision
}

const (
	// cardStackVisible cards are built in the stack; the last waits invisibly
	// one step behind and fades in as it rises.
	cardStackVisible = 4
	// Each card behind peeks cardStackPeek out below the one in front and is
	// cardStackShrink smaller.
	cardStackPeek   = 12
	cardStackShrink = .05
	// The cards behind rise cardStackLift stack steps while the top card is
	// dragged away, fully once it travelled cardStackLiftDistance.
	cardStackLift         = .4
	cardStackLiftDistance = 160
	// cardStackRotate is the degrees of turn per DIP of horizontal travel.
	cardStackRotate = .065
	// A release faster than cardStackFlick DIP/s throws the card even when it
	// moved only a little.
	cardStackFlick = 500
	// Vertical travel is no decision: past a small free zone it resists as a
	// rubber band.
	cardStackFreeY    = 48
	cardStackStretchY = 140
	// Stamps fade in between these distances of horizontal travel.
	cardStackStampFrom = 16
	cardStackStampTo   = 96
	// The stage is cardStackWidth wide at most and holds a card
	// cardStackHeight tall plus the edges peeking below it.
	cardStackWidth     = 320
	cardStackHeight    = 304
	cardStackPeekSpace = 24
	// cardStackStagger staggers cards dealt back by a reset.
	cardStackStagger = 60 * time.Millisecond
)

var (
	cardStackSmooth = ui.Animation{Spring: ui.SpringOptions{Duration: 480 * time.Millisecond}}
	cardStackMorph  = ui.Animation{Spring: ui.SpringOptions{Duration: 504 * time.Millisecond, Bounce: .16}}
	cardStackThrow  = ui.Animation{Spring: ui.SpringOptions{Duration: 600 * time.Millisecond}}
	cardStackFade   = ui.Animation{Duration: 220 * time.Millisecond, Delay: 120 * time.Millisecond, Ease: ui.CubicBezier(.22, 1, .36, 1)}
	cardStackEnter  = ui.CubicBezier(.16, 1, .3, 1)
)

type cardPoint struct{ x, y float32 }

// cardExit is where a thrown card left the screen, and its tilt.
type cardExit struct {
	x, y, tilt float32
}

// cardReturn brings a card back from its exit after delay.
type cardReturn struct {
	exit  cardExit
	delay time.Duration
}

type cardSample struct {
	t    time.Time
	x, y float32
}

// cardDrag is the pointer dragging the top card.
type cardDrag struct {
	id      string
	ox, oy  float32
	x, y    float32
	tilt    float32
	samples []cardSample
}

// cardRelease is a drag let go, for the next build to decide or return.
type cardRelease struct {
	id       string
	x, y     float32
	v        cardPoint
	canceled bool
}

// cardStackMotion is the retained presentation of a CardStack.
type cardStackMotion struct {
	known     []CardStackDecision
	built     map[string]bool
	flying    []string
	returning map[string]cardReturn
	exits     map[string]cardExit
	flings    map[string]cardPoint
	backs     map[string]cardPoint
	lift      float32
	drag      *cardDrag
	release   *cardRelease
	// top is the card on top in the last frame and where it showed, in the
	// stage's coordinates.
	top     string
	topRect ui.Rect
	// scale is the theme's DIP scale the pointer's distances follow.
	scale float32
}

// cardMotion is the retained presentation of one card.
type cardMotion struct {
	ready    bool
	decision CardDecision
	x, y     float32
	tilt     float32
	homing   bool
	tx, ty   float32
	ax, ay   ui.Animation
	opacity  float32
	fade     ui.Animation
}

type cardStackMotionKey struct{}
type cardMotionKey struct{}

// CardStack shows a deck reviewed one card at a time: drag or flick the top
// card left or right, or use the buttons or the arrow keys, to decide it. The
// card follows the pointer with a tilt and a stamp naming the action, and a
// throw carries it off with the pointer's velocity while the cards behind
// rise. Undo, by its button, Backspace or Command/Ctrl+Z, flies the last card
// back from where it left; once all are decided, Start over deals them back.
//
// state holds the decisions, which applications own and may change; the stack
// returns the change its own input made this frame. Applications supply the
// cards' stable IDs, localized labels and content, and act on decisions.
// Reduced Motion moves cards at once.
func CardStack(c *ui.Context, state *CardStackState, options CardStackOptions) CardStackEvent {
	t := c.Theme()
	s := t.Space(.25)
	m := l10n.In(c.Locale())
	left, right, label := options.LeftLabel, options.RightLabel, options.Label
	if left == "" {
		left = m.Text("cardStackPass")
	}
	if right == "" {
		right = m.Text("cardStackKeep")
	}
	if label == "" {
		label = m.Text("cardStack")
	}
	items := options.Items
	var event CardStackEvent
	position := func() int { return min(len(state.Decisions), len(items)) }
	find := func(id string) int {
		return slices.IndexFunc(items, func(it CardStackItem) bool { return it.ID == id })
	}
	decide := func(d CardDecision) {
		p := position()
		if p >= len(items) {
			return
		}
		top := items[p]
		state.Decisions = append(state.Decisions, CardStackDecision{top.ID, d})
		event = CardStackEvent{CardStackDecided, top.ID, d}
		action, outcome := left, options.LeftOutcome
		if d == CardRight {
			action, outcome = right, options.RightOutcome
		}
		text := m.Format("cardStackDecided", map[string]any{"label": top.Label, "action": action})
		if outcome != "" {
			text = m.Format("cardStackOutcome", map[string]any{"label": top.Label, "outcome": outcome})
		}
		c.Announce(text + " " + m.Format("cardStackRemaining", map[string]any{"count": len(items) - p - 1}))
		c.Invalidate()
	}
	undo := func() {
		n := len(state.Decisions)
		if n == 0 {
			return
		}
		last := state.Decisions[n-1]
		state.Decisions = state.Decisions[:n-1]
		event = CardStackEvent{CardStackUndone, last.ID, last.Decision}
		if i := find(last.ID); i >= 0 {
			c.Announce(m.Format("cardStackRestored", map[string]any{"label": items[i].Label}))
		}
		c.Invalidate()
	}
	reset := func() {
		if len(state.Decisions) == 0 {
			return
		}
		state.Decisions = nil
		event = CardStackEvent{Action: CardStackReset}
		c.Announce(m.Text("cardStackRestoredAll"))
		c.Invalidate()
	}

	root := ui.Column(c).FillWidth().MinWidth(0).AlignItems(ui.Center).Gap(t.Space(5))
	root.Children(func() {
		stage := ui.Box(c).FillWidth().MaxWidth(cardStackWidth * s).Height((cardStackHeight + cardStackPeekSpace) * s).
			Focusable().FocusRing(false).Role(ui.RoleGroup).Label(label).
			Description(m.Format("cardStackHint", map[string]any{"left": strings.ToLower(left), "right": strings.ToLower(right)}))
		sm := ui.Local(stage, cardStackMotionKey{}, func() cardStackMotion {
			return cardStackMotion{built: map[string]bool{}, returning: map[string]cardReturn{}, exits: map[string]cardExit{},
				flings: map[string]cardPoint{}, backs: map[string]cardPoint{}}
		})
		sm.scale = s
		width := cardStackWidth * s
		if r := stage.LayoutBounds(); r.W > 0 {
			width = r.W
		}

		// Keys decide the top card or undo while the stage has the focus.
		switch {
		case stage.Shortcut(0, ui.KeyLeft):
			decide(CardLeft)
		case stage.Shortcut(0, ui.KeyRight):
			decide(CardRight)
		case stage.Shortcut(0, ui.KeyBackspace), stage.Shortcut(ui.Cmd, ui.KeyZ):
			undo()
		}

		// A drag let go throws the card when it went far or was flicked, and
		// otherwise, as when the system took the press, springs it home with
		// its velocity.
		if r := sm.release; r != nil {
			sm.release = nil
			if p := position(); p < len(items) && items[p].ID == r.id {
				dir := cardSign(r.x)
				flick := cardAbs(r.v.x) > cardStackFlick*s && cardSign(r.v.x) == dir && cardAbs(r.x) > 12*s
				far := cardAbs(r.x) > min(width*.35, 140*s) && !(cardSign(r.v.x) == -dir && cardAbs(r.v.x) > 300*s)
				if dir != 0 && !r.canceled && (flick || far) {
					sm.flings[r.id] = r.v
					if dir > 0 {
						decide(CardRight)
					} else {
						decide(CardLeft)
					}
				} else {
					sm.backs[r.id] = r.v
				}
			}
		}

		sm.reconcile(state.Decisions, items, position())

		pos := position()
		end := min(pos+cardStackVisible, len(items))
		decided := make(map[string]CardDecision, len(state.Decisions))
		for _, d := range state.Decisions {
			decided[d.ID] = d.Decision
		}
		// A drag puts the top card under the pointer, so the cards behind,
		// built before it, rise with it in the same frame.
		if d := sm.drag; d != nil && pos < len(items) && items[pos].ID == d.id {
			sm.lift = cardLift(d.x, d.y, s)
		}
		built := map[string]bool{}
		sm.top = ""
		stage.HandleInput(sm.input)
		stage.Children(func() {
			card := func(i int) {
				item := items[i]
				built[item.ID] = true
				depth := i - pos
				d := decided[item.ID]
				if d != 0 {
					depth = 0
				}
				sm.card(c, item, cardFace{
					depth: depth, decision: d, interactive: d == 0 && depth == 0, focus: stage.FocusVisible(),
					label: m.Format("cardStackCard", map[string]any{"label": item.Label, "index": i + 1, "count": len(items)}),
					left:  left, right: right, width: width, scale: s,
				})
			}
			// The deepest card first, the top card over it and cards in flight
			// over everything, in the order they were thrown.
			for i := end - 1; i >= pos; i-- {
				card(i)
			}
			for _, id := range slices.Clone(sm.flying) {
				if i := find(id); i >= 0 && !built[id] {
					card(i)
				}
			}
			if pos >= len(items) && len(sm.flying) == 0 {
				empty := ui.Column(c).Key("empty").Absolute().Left(0).Right(0).Top(0).Height(cardStackHeight*s).
					Center().Gap(t.Space(3)).Padding(t.Space(6)).Radius(t.Radius*2).
					Background(t.Background).Border(1, t.Border).TextColor(t.Text).
					Transition(ui.ElementTransition{Duration: 240 * time.Millisecond, Ease: cardStackEnter,
						Enter: &ui.Motion{Y: 8 * s}, Exit: &ui.Motion{}})
				if stage.FocusVisible() {
					empty.Border(2, t.Focus)
				}
				empty.Children(func() {
					if options.Empty != nil {
						options.Empty(c, reset)
						return
					}
					Text(c, m.Text("cardStackReviewed")).FontWeight(500).TextAlign(ui.Center)
					if cardStackButton(c, lucide.RotateCcw, m.Text("cardStackStartOver"), ButtonOutline, false).Clicked() {
						reset()
					}
				})
			}
		})
		sm.built = built
		if sm.drag != nil && sm.drag.id == sm.top {
			stage.Cursor(ui.CursorGrabbing)
		}

		ui.Row(c).Wrap().Justify(ui.Center).Gap(t.Space(2)).Children(func() {
			none := position() >= len(items)
			if cardStackButton(c, lucide.X, left, ButtonOutline, none).Clicked() {
				decide(CardLeft)
			}
			if cardStackButton(c, lucide.RotateCcw, m.Text("cardStackUndo"), ButtonGhost, len(state.Decisions) == 0).Clicked() {
				undo()
			}
			if cardStackButton(c, lucide.Check, right, ButtonPrimary, none).Clicked() {
				decide(CardRight)
			}
		})
	})
	return event
}

// cardStackButton is a labeled action with a leading icon.
func cardStackButton(c *ui.Context, icon func() *ui.SVG, label string, variant ButtonVariant, disabled bool) ui.Element {
	t := c.Theme()
	var b ui.Element
	ui.Box(c).Key(label).Children(func() {
		b = Button(c, "", variant).Label(label).Gap(t.Space(1.5)).Disabled(disabled)
		dimmed(b)
		b.Children(func() {
			Icon(c, icon).Size(t.Rem(1), t.Rem(1))
			Text(c, label).SingleLine()
		})
	})
	return b
}

// reconcile follows changes to the decisions, by the stack or the
// application: cards newly decided are thrown, and cards whose decisions went
// away come back from where they left.
func (sm *cardStackMotion) reconcile(decisions []CardStackDecision, items []CardStackItem, position int) {
	p := 0
	for p < len(sm.known) && p < len(decisions) && sm.known[p] == decisions[p] {
		p++
	}
	removed, added := sm.known[p:], decisions[p:]
	if len(removed) == 0 && len(added) == 0 {
		return
	}
	sm.lift = 0
	clear(sm.returning)
	for _, d := range added {
		if sm.drag != nil && sm.drag.id == d.ID {
			// A key or button decided the card mid-drag; the throw takes over.
			sm.drag = nil
		}
		if sm.built[d.ID] && !slices.Contains(sm.flying, d.ID) {
			sm.flying = append(sm.flying, d.ID)
		}
	}
	// The deck deals itself back: the card that will sit deepest returns
	// first and the top card last.
	count := min(cardStackVisible, len(removed), len(items)-position)
	for _, d := range removed {
		sm.flying = slices.DeleteFunc(sm.flying, func(id string) bool { return id == d.ID })
		depth := slices.IndexFunc(items, func(it CardStackItem) bool { return it.ID == d.ID }) - position
		exit, ok := sm.exits[d.ID]
		if !ok || depth < 0 || depth >= cardStackVisible {
			continue
		}
		var delay time.Duration
		if len(removed) > 1 {
			delay = time.Duration(max(count-1-depth, 0)) * cardStackStagger
		}
		sm.returning[d.ID] = cardReturn{exit, delay}
	}
	sm.known = slices.Clone(decisions)
}

// input takes the pointer pressing, dragging and letting go of the top card.
func (sm *cardStackMotion) input(ev ui.InputEvent) bool {
	switch ev.Kind {
	case ui.InputPointerDown:
		if ev.Button != 0 || sm.top == "" || !sm.topRect.Contains(ev.X, ev.Y) {
			return false
		}
		x, y := sm.topRect.X, sm.topRect.Y
		d := &cardDrag{id: sm.top, ox: ev.X - x, oy: ev.Y - cardUnresist(y, sm.scale), x: x, y: y,
			samples: []cardSample{{ev.Time, ev.X, ev.Y}}}
		// Grabbing the top half tips the card the way the hand moves; the
		// bottom half tips it the other way.
		if cardAbs(x) < 2 {
			d.tilt = 1
			if ev.Y >= sm.topRect.Y+sm.topRect.H/2 {
				d.tilt = -1
			}
		}
		sm.drag = d
	case ui.InputPointerMove:
		d := sm.drag
		if d == nil {
			return false
		}
		d.x, d.y = ev.X-d.ox, cardResist(ev.Y-d.oy, sm.scale)
		d.samples = append(d.samples, cardSample{ev.Time, ev.X, ev.Y})
		if len(d.samples) > 12 {
			d.samples = d.samples[1:]
		}
	case ui.InputPointerUp:
		d := sm.drag
		if d == nil || ev.Button != 0 {
			return false
		}
		sm.drag = nil
		sm.release = &cardRelease{d.id, d.x, d.y, cardVelocity(d.samples, ev.Time), ev.Canceled}
	default:
		return false
	}
	return true
}

// cardFace is how one card shows in a frame.
type cardFace struct {
	depth       int
	decision    CardDecision
	interactive bool
	focus       bool
	label       string
	left, right string
	width       float32
	scale       float32
}

// card builds one card: a slot that sinks and shrinks with its depth, and the
// card in it, which moves and turns as it is dragged and thrown.
func (sm *cardStackMotion) card(c *ui.Context, item CardStackItem, f cardFace) {
	t := c.Theme()
	s := f.scale
	slot := ui.Box(c).Key(item.ID).Absolute().Left(0).Right(0).Top(0).Height(cardStackHeight * s)
	cm := ui.Local(slot, cardMotionKey{}, func() cardMotion {
		return cardMotion{tilt: 1, opacity: 1, ax: cardStackSmooth, ay: cardStackSmooth, fade: ui.Animation{Duration: 160 * time.Millisecond, Ease: cardStackEnter}}
	})
	id := item.ID
	ret, returning := sm.returning[id]
	switch {
	case !cm.ready && returning:
		// Brought back by undo or a reset after it was gone: it comes back
		// from where it left.
		slot.Jump("x", ret.exit.x)
		slot.Jump("y", ret.exit.y)
		slot.Jump("opacity", 0)
		cm.x, cm.y, cm.tilt = ret.exit.x, ret.exit.y, ret.exit.tilt
		cm.home(ret.delay)
		delete(sm.returning, id)
	case cm.ready && f.decision != 0 && cm.decision == 0:
		sm.throw(cm, id, f.decision, f.width, s)
	case cm.ready && f.decision == 0 && cm.decision != 0:
		// Undone in flight: it turns around from where it is.
		cm.home(ret.delay)
		delete(sm.returning, id)
	}
	cm.ready, cm.decision = true, f.decision
	if v, ok := sm.backs[id]; ok {
		delete(sm.backs, id)
		cm.tx, cm.ty, cm.ax, cm.ay = 0, 0, cardStackMorph, cardStackMorph
		cm.ax.Spring.Velocity, cm.ay.Spring.Velocity = v.x, v.y
	}
	dragging := f.interactive && sm.drag != nil && sm.drag.id == id
	if dragging {
		if sm.drag.tilt != 0 {
			cm.tilt, sm.drag.tilt = sm.drag.tilt, 0
		}
		cm.homing = false
		cm.tx, cm.ty = sm.drag.x, sm.drag.y
		cm.x, cm.y = slot.Jump("x", sm.drag.x), slot.Jump("y", sm.drag.y)
	} else {
		cm.x, cm.y = slot.AnimateTo("x", cm.tx, cm.ax), slot.AnimateTo("y", cm.ty, cm.ay)
	}
	opacity := slot.AnimateTo("opacity", cm.opacity, cm.fade)
	if cm.homing && slot.Settled("x") && slot.Settled("y") {
		cm.homing = false
	}
	if f.decision != 0 && slot.Settled("opacity") && opacity <= 0 {
		// Gone: the flight ends and the card leaves the stack.
		sm.flying = slices.DeleteFunc(sm.flying, func(other string) bool { return other == id })
		c.Invalidate()
	}
	if f.interactive {
		sm.top, sm.topRect = id, ui.Rect{X: cm.x, Y: cm.y, W: f.width, H: cardStackHeight * s}
		if !cm.homing && !dragging {
			// The cards behind respond to the card springing home.
			if lift := cardLift(cm.x, cm.y, s); lift != sm.lift {
				// The cards behind, built before this one, follow in the next
				// frame the display shows rather than in one built at once.
				sm.lift = lift
				c.AnimationFrame()
			}
		}
	}

	// The slot springs toward its depth; while the top card is dragged away
	// the cards behind already start to rise.
	target := float32(f.depth)
	if f.depth > 0 {
		target -= sm.lift * cardStackLift
	}
	depth := slot.AnimateTo("depth", target, cardStackSmooth)
	scale := 1 - depth*cardStackShrink
	slot.TransformOrigin(.5, 1).Transform(ui.Scaling(scale, scale).Then(ui.Translation(0, depth*cardStackPeek*s))).
		Opacity(1 - cardClamp(depth-(cardStackVisible-2)))
	if !f.interactive {
		slot.PassThrough()
	}

	// Resting cards sit flat on their borders; a card in the hand or in flight
	// floats and casts a shadow.
	lifted := float32(0)
	if dragging || f.decision != 0 {
		lifted = 1
	}
	shadow := slot.AnimateTo("shadow", lifted, ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.CubicBezier(.22, 1, .36, 1)})
	radius := t.Radius * 2
	slot.Children(func() {
		card := ui.Box(c).Absolute().Left(0).Right(0).Top(0).Bottom(0).Radius(radius).
			Transform(ui.Rotation(cm.x / s * cardStackRotate * cm.tilt).Then(ui.Translation(cm.x, cm.y))).Opacity(opacity)
		if f.interactive {
			card.Role(ui.RoleGroup).Label(f.label)
			if dragging {
				card.Cursor(ui.CursorGrabbing)
			} else {
				card.Cursor(ui.CursorGrab)
			}
		}
		card.Children(func() {
			face := ui.Box(c).Absolute().Left(0).Right(0).Top(0).Bottom(0).Radius(radius).
				Background(t.Background).TextColor(t.Text)
			if shadow > 0 {
				face.Shadow(0, 12*s, 32*s, 0, ui.RGBA(0, 0, 0, .16*shadow))
			}
			face.Children(func() {
				if item.Content != nil {
					item.Content(c, radius)
				}
				// The cards behind darken a little with their depth.
				if shade := cardClamp(depth / 2); shade > 0 {
					a := float32(.035)
					if t.Dark {
						a = .22
					}
					ui.Box(c).Absolute().Left(0).Right(0).Top(0).Bottom(0).Radius(radius).PassThrough().Background(ui.RGBA(0, 0, 0, a*shade))
				}
				// The border goes over content reaching the edges.
				ui.Box(c).Absolute().Left(0).Right(0).Top(0).Bottom(0).Radius(radius).PassThrough().Border(1, t.Border)
			})
			// Cards returning to the back of the stack carry no verdict.
			if f.interactive || f.decision != 0 {
				keep := cardClamp((cm.x - cardStackStampFrom*s) / ((cardStackStampTo - cardStackStampFrom) * s))
				pass := cardClamp((-cardStackStampFrom*s - cm.x) / ((cardStackStampTo - cardStackStampFrom) * s))
				if keep > 0 {
					cardStamp(c, lucide.Check, f.right, t.Success, true, s).Opacity(keep)
				}
				if pass > 0 {
					cardStamp(c, lucide.X, f.left, t.TextMuted, false, s).Opacity(pass)
				}
			}
			// Keyboard focus on the stack outlines its top card.
			if f.interactive && f.focus {
				o := 3 * s
				ui.Box(c).Absolute().Left(-o-2).Right(-o-2).Top(-o-2).Bottom(-o-2).PassThrough().
					Radius(radius+o+2).Border(2, t.Focus)
			}
		})
	})
}

// cardStamp is the pill naming the action a card dragged that way takes.
func cardStamp(c *ui.Context, icon func() *ui.SVG, label string, tint ui.Color, keep bool, s float32) ui.Element {
	t := c.Theme()
	border, bg := t.Border, t.Surface
	if keep {
		border, bg = t.Border.Mix(tint, .32), t.Surface.Mix(tint, .09)
	}
	stamp := ui.Row(c).Absolute().Top(16*s).Height(32*s).AlignItems(ui.Center).Gap(6*s).
		Padding(0, 12*s, 0, 10*s).Radius(16*s).Border(1, border).Background(bg).TextColor(t.Text).PassThrough()
	if keep {
		stamp.Left(16 * s)
	} else {
		stamp.Right(16 * s)
	}
	stamp.Children(func() {
		Icon(c, icon).Size(16*s, 16*s).TextColor(tint)
		Text(c, label).SingleLine().FontWeight(500).FontSize(t.Rem(.875))
	})
	return stamp
}

// throw sends a card off toward the side d: along the flick when it heads
// that way, otherwise along the drag, otherwise straight out.
func (sm *cardStackMotion) throw(cm *cardMotion, id string, d CardDecision, width, s float32) {
	dir := float32(-1)
	if d == CardRight {
		dir = 1
	}
	// Buttons and keys throw with a default flick; a drag keeps the pointer's
	// velocity.
	v, ok := sm.flings[id]
	delete(sm.flings, id)
	if !ok {
		v = cardPoint{dir * 900 * s, -120 * s}
	}
	cx, cy := cm.x, cm.y
	ux, uy := v.x, v.y
	if float32(math.Hypot(float64(ux), float64(uy))) < 200*s || cardSign(ux) != dir {
		if cardAbs(cx) > 8*s {
			ux, uy = cx, cy*.5
		} else {
			ux, uy = dir, -.12
		}
	}
	length := float32(math.Hypot(float64(ux), float64(uy)))
	if length == 0 {
		length = 1
	}
	ux, uy = ux/length, uy/length
	if cardAbs(ux) < .45 {
		ux = dir * .45
		uy = cardSign(uy) * float32(math.Sqrt(1-.45*.45))
		if uy == 0 {
			uy = -float32(math.Sqrt(1 - .45*.45))
		}
	}
	// Far enough to clear a typical column; the fade ends before the spring.
	reach := width * 1.6
	exit := cardExit{cx + ux*reach, cy + uy*reach, cm.tilt}
	sm.exits[id] = exit
	cm.tx, cm.ty, cm.ax, cm.ay = exit.x, exit.y, cardStackThrow, cardStackThrow
	cm.ax.Spring.Velocity, cm.ay.Spring.Velocity = v.x, v.y
	cm.opacity, cm.fade = 0, cardStackFade
	cm.homing = false
}

// home brings a card back to its place after delay.
func (cm *cardMotion) home(delay time.Duration) {
	cm.tx, cm.ty, cm.ax, cm.ay = 0, 0, cardStackSmooth, cardStackSmooth
	cm.ax.Delay, cm.ay.Delay = delay, delay
	cm.opacity = 1
	cm.fade = ui.Animation{Duration: 160 * time.Millisecond, Ease: cardStackEnter, Delay: delay}
	cm.homing = true
}

// cardLift is how far the cards behind have risen, from 0 to 1, for the top
// card at x, y.
func cardLift(x, y, s float32) float32 {
	return min(float32(math.Hypot(float64(x), float64(y*.5)))/(cardStackLiftDistance*s), 1)
}

// cardResist holds vertical travel past a small free zone as a rubber band.
func cardResist(raw, s float32) float32 {
	free, stretch := cardStackFreeY*s, cardStackStretchY*s
	d := cardAbs(raw)
	if d <= free {
		return raw
	}
	return cardSign(raw) * (free + (1-1/((d-free)*.55/stretch+1))*stretch)
}

// cardUnresist is cardResist's inverse, for a drag catching a card in place.
func cardUnresist(shown, s float32) float32 {
	free, stretch := cardStackFreeY*s, cardStackStretchY*s
	d := cardAbs(shown)
	if d <= free {
		return shown
	}
	over := min(d-free, stretch-1)
	return cardSign(shown) * (free + (1/(1-over/stretch)-1)*stretch/.55)
}

// cardVelocity is the pointer's velocity over its last 80 ms, zero when it
// held still for 60 ms before letting go.
func cardVelocity(samples []cardSample, now time.Time) cardPoint {
	var recent []cardSample
	for _, sample := range samples {
		if now.Sub(sample.t) <= 80*time.Millisecond {
			recent = append(recent, sample)
		}
	}
	if len(recent) < 2 || now.Sub(recent[len(recent)-1].t) > 60*time.Millisecond {
		return cardPoint{}
	}
	first, last := recent[0], recent[len(recent)-1]
	seconds := float32(last.t.Sub(first.t).Seconds())
	if seconds <= 0 {
		return cardPoint{}
	}
	return cardPoint{(last.x - first.x) / seconds, (last.y - first.y) / seconds}
}

func cardSign(v float32) float32 {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

func cardAbs(v float32) float32 { return float32(math.Abs(float64(v))) }

func cardClamp(v float32) float32 { return min(max(v, 0), 1) }
