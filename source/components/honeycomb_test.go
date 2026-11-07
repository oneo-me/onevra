package components

import (
	"fmt"
	"image"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/theme"
)

func honeycombItems(n int) []HoneycombItem {
	bitmap := ui.NewBitmap(image.NewRGBA(image.Rect(0, 0, 8, 8)))
	items := make([]HoneycombItem, n)
	for i := range items {
		id := fmt.Sprintf("p%02d", i)
		items[i] = HoneycombItem{ID: id, Label: id, Image: bitmap}
	}
	return items
}

func honeycombTestMetrics() honeycombMetrics {
	return honeycombMetricsFor(theme.Palette(false), 0)
}

func honeycombTester(items []HoneycombItem, selected *string) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.Palette(false))
		// The column lays out like an item pane, the honeycomb growing to fill it.
		ui.Column(c).Fill().Label("Honeycomb").AlignItems(ui.Center).Children(func() {
			if id := Honeycomb(c, HoneycombOptions{Items: items, Selected: *selected, SelectedLabel: "Item"}); id != "" {
				*selected = id
			}
		})
	}, 600, 300)
}

func settleHoneycomb(tt *ui.Tester) {
	tt.Advance(honeycombDuration + 50*time.Millisecond)
	tt.Frame()
}

func middle(r ui.Rect) (float32, float32) { return r.X + r.W/2, r.Y + r.H/2 }

func TestHoneycombLaysOutRowsAroundTheSelection(t *testing.T) {
	items := honeycombItems(12)
	selected := "p00"
	tt := honeycombTester(items, &selected)
	tt.Frame()
	m := honeycombTestMetrics()
	first, ok := tt.Find("Item")
	if !ok || first.W != m.large {
		t.Fatalf("selected item not large: %+v", first)
	}
	cx, cy := middle(first)
	right, _ := tt.Find("p01")
	if x, y := middle(right); math.Abs(float64(x-cx-m.magnified(m.spacing, 1))) > 1 || y != cy || right.W >= first.W {
		t.Fatalf("neighbor not a smaller item beside the selected one: %+v %+v", right, first)
	}
	// Items fill rows above and below, none smaller than the smallest item.
	above, below := false, false
	for _, item := range items[1:] {
		r, _ := tt.Find(item.ID)
		if r.W < m.tiny-.01 {
			t.Fatalf("%s smaller than the smallest item: %+v", item.ID, r)
		}
		_, y := middle(r)
		above = above || y < cy-m.spacing/2
		below = below || y > cy+m.spacing/2
	}
	if !above || !below {
		t.Fatalf("items not in rows around the selection (above %v, below %v)", above, below)
	}
	// Clear of the edge, the further from the center, the smaller, and gradually
	// so; nearer the edge, items shrink further. An item on the ball is an
	// ellipse, so its larger side is the size it would have facing the viewer.
	view, _ := tt.Find("Honeycomb")
	sizeOf := func(r ui.Rect) float32 { return max(r.W, r.H) }
	clear := func(r ui.Rect) bool {
		x, y := middle(r)
		return r.X >= view.X && r.Y >= view.Y && r.X+r.W <= view.X+view.W && r.Y+r.H <= view.Y+view.H &&
			x-view.X > m.edge && view.X+view.W-x > m.edge && y-view.Y > m.edge && view.Y+view.H-y > m.edge
	}
	edged := false
	for _, a := range items[1:] {
		ra, _ := tt.Find(a.ID)
		if !clear(ra) {
			edged = edged || sizeOf(ra) < m.small*.8
			continue
		}
		for _, b := range items[1:] {
			rb, _ := tt.Find(b.ID)
			ax, ay := middle(ra)
			bx, by := middle(rb)
			if math.Hypot(float64(ax-cx), float64(ay-cy)) < math.Hypot(float64(bx-cx), float64(by-cy))-1 && sizeOf(ra) < sizeOf(rb) {
				t.Fatalf("%s nearer the center but smaller than %s: %+v %+v", a.ID, b.ID, ra, rb)
			}
		}
	}
	if r, _ := tt.Find("p07"); r.W <= m.tiny || r.W >= right.W {
		t.Fatalf("second ring not between the first and the far size: %+v %+v", r, right)
	}
	if !edged {
		t.Fatal("no item shrunk near the edge")
	}
}

func TestHoneycombCentersTheClickedItem(t *testing.T) {
	items := honeycombItems(9)
	selected := "p00"
	tt := honeycombTester(items, &selected)
	m := honeycombTestMetrics()
	tt.Frame()
	center, _ := tt.Find("Item")
	cx, cy := middle(center)
	target, _ := tt.Find("p04")
	if err := tt.Click("p04"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if selected != "p04" {
		t.Fatalf("clicked item not chosen: %q", selected)
	}
	// The honeycomb pans over, growing the chosen item on the way.
	tt.Advance(honeycombDuration / 2)
	moving := false
	if r, ok := tt.Find("Item"); ok {
		x, _ := middle(r)
		tx, _ := middle(target)
		moving = x != tx && x != cx
	}
	tt.Advance(2 * honeycombDuration)
	r, _ := tt.Find("Item")
	if x, y := middle(r); !moving || r.W != m.large || math.Abs(float64(x-cx)) > 1 || math.Abs(float64(y-cy)) > 1 {
		t.Fatalf("clicked item not centered smoothly (moving %v): %+v, center %v,%v", moving, r, cx, cy)
	}
}

func TestHoneycombCanceledPressChoosesNothing(t *testing.T) {
	selected := "p00"
	tt := honeycombTester(honeycombItems(9), &selected)
	tt.Frame()
	x, y := middle(mustFind(t, tt, "p04"))
	tt.Press(x, y)
	tt.CancelPress()
	tt.Frame()
	if selected != "p00" {
		t.Fatalf("canceled press chose %q", selected)
	}
}

func TestHoneycombOffsetsItsMiddle(t *testing.T) {
	items := honeycombItems(19)
	selected := "p00"
	build := func(offset float32) (item, side, view ui.Rect) {
		tt := ui.NewTester(func(c *ui.Context) {
			c.SetTheme(theme.Palette(false))
			ui.Column(c).Fill().Label("Honeycomb").AlignItems(ui.Center).Children(func() {
				Honeycomb(c, HoneycombOptions{Items: items, Selected: selected, SelectedLabel: "Item", OffsetY: offset})
			})
		}, 600, 300)
		tt.Frame()
		item, _ = tt.Find("Item")
		side, _ = tt.Find("p01")
		view, _ = tt.Find("Honeycomb")
		return
	}
	plain, plainSide, plainView := build(0)
	shifted, shiftedSide, shiftedView := build(-60)
	_, cy := middle(plainView)
	// The selected item and its neighbors move with the offset, while the view
	// and the edge fade stay put.
	if _, y := middle(plain); math.Abs(float64(y-cy)) > .01 {
		t.Fatalf("the middle item is not in the middle: %+v", plain)
	}
	if _, y := middle(shifted); math.Abs(float64(y-(cy-60))) > .01 {
		t.Fatalf("the middle item did not move up: %+v, middle %v", shifted, cy)
	}
	if px, _ := middle(plainSide); func() bool {
		sx, sy := middle(shiftedSide)
		return math.Abs(float64(sx-px)) > .01 || math.Abs(float64(sy-(cy-60))) > .01
	}() {
		t.Fatalf("the item beside the middle did not follow: %+v %+v", plainSide, shiftedSide)
	}
	if shifted.W != plain.W || shifted.H != plain.H || shiftedView != plainView {
		t.Fatalf("the offset resized items or the view: %+v %+v, view %+v %+v", plain, shifted, plainView, shiftedView)
	}
}

func TestHoneycombSettlesOnTheItemNearestTheMiddle(t *testing.T) {
	items := honeycombItems(19)
	selected := "p00"
	tt := honeycombTester(items, &selected)
	m := honeycombTestMetrics()
	// Without motion, the ring does not keep the harness repainting, so the
	// scroll below settles only when the clock reaches its settle check.
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	tt.Frame()
	center, _ := tt.Find("Item")
	cx, cy := middle(center)
	start, _ := tt.Find("p01")
	// Dragging moves the magnified middle with the pointer, growing the item it
	// brings there, and chooses nothing yet.
	x, y := middle(start)
	reach := m.spacing * m.large / m.small
	tt.Press(x, y)
	tt.Frame()
	for i := 1; i <= 6; i++ {
		tt.Move(x-float32(i)*reach/6, y)
		tt.Frame()
	}
	if r, _ := tt.Find("p01"); func() bool { mx, my := middle(r); return math.Abs(float64(mx-cx)) > .5 || my != cy }() || r.W != m.large {
		t.Fatalf("items not following the pointer: %+v, center %v,%v", r, cx, cy)
	}
	if r, _ := tt.Find("Item"); r.W >= m.large-10 {
		t.Fatalf("item out of the middle still large: %+v", r)
	}
	// Held still before letting go, the drag does not glide on.
	tt.Advance(100 * time.Millisecond)
	tt.Release(x-reach, y)
	tt.Frame()
	if selected != "p01" {
		t.Fatalf("drag did not choose the item in the middle: %q", selected)
	}
	// Scrolling pans too, and once it stops settles on the item nearest the
	// middle, here the one below and right.
	tt.Scroll(x, y, 0, 40*m.large/m.small)
	tt.Frame()
	// The settle waits honeycombSettle after the scroll is over.
	if selected != "p01" {
		t.Fatalf("scrolling chose an item before it stopped: %q", selected)
	}
	tt.Advance(honeycombSettle + 20*time.Millisecond)
	if selected != "p02" {
		t.Fatalf("scrolling did not choose the item nearest the middle: %q", selected)
	}
	settleHoneycomb(tt)
	if r, _ := tt.Find("Item"); func() bool { mx, my := middle(r); return math.Abs(float64(mx-cx)) > 1 || math.Abs(float64(my-cy)) > 1 }() || r.W != m.large {
		t.Fatalf("chosen item not centered: %+v, center %v,%v", r, cx, cy)
	}
}

func TestHoneycombSettlesInside(t *testing.T) {
	items := honeycombItems(7)
	selected := "p00"
	tt := honeycombTester(items, &selected)
	m := honeycombTestMetrics()
	tt.Frame()
	center, _ := tt.Find("Item")
	cx, cy := middle(center)
	// Past the edge of the honeycomb, items follow the pointer less.
	tt.Press(cx, cy)
	tt.Frame()
	for i := 1; i <= 10; i++ {
		tt.Move(cx+float32(i)*40, cy)
		tt.Frame()
	}
	r, _ := tt.Find("Item")
	if x, _ := middle(r); x-cx >= 400-1 || x-cx < m.spacing {
		t.Fatalf("drag past the edge not resisted: %+v", r)
	}
	tt.Release(cx+400, cy)
	tt.Frame()
	settleHoneycomb(tt)
	// The leftmost cell is the furthest the pan goes.
	cells := honeycombCells(len(items), m.spacing)
	lo, _ := honeycombBounds(cells)
	i := slices.IndexFunc(items, func(item HoneycombItem) bool { return item.ID == selected })
	r, _ = tt.Find("Item")
	if x, _ := middle(r); cells[i].x != lo.x || math.Abs(float64(x-cx)) > 1 {
		t.Fatalf("honeycomb not settled on its leftmost item: %q at %+v, %+v", selected, cells[i], r)
	}
}

func TestHoneycombCellsFillAHoneycomb(t *testing.T) {
	m := honeycombTestMetrics()
	cells := honeycombCells(40, m.spacing)
	if cells[0] != (honeycombPoint{}) {
		t.Fatalf("first cell not in the middle: %+v", cells[0])
	}
	for i, a := range cells {
		for _, b := range cells[i+1:] {
			if d := math.Hypot(float64(a.x-b.x), float64(a.y-b.y)); d < float64(m.spacing)-.01 {
				t.Fatalf("cells overlap: %+v %+v", a, b)
			}
		}
	}
	// Round, filling each ring before the next.
	far := 0.0
	for _, c := range cells {
		far = max(far, math.Hypot(float64(c.x), float64(c.y)))
	}
	for _, c := range honeycombCells(60, m.spacing)[40:] {
		if d := math.Hypot(float64(c.x), float64(c.y)); d < far-.01 {
			t.Fatalf("cell %+v nearer than the farthest of the first, %v", c, far)
		}
	}
	if !slices.Equal(cells[:7], honeycombCells(7, m.spacing)) {
		t.Fatal("cells change with the number of items")
	}
}

func TestHoneycombTiltsItemsOnABall(t *testing.T) {
	items := honeycombItems(19)
	selected := "p00"
	tt := honeycombTester(items, &selected)
	m := honeycombTestMetrics()
	settleHoneycomb(tt)
	center, ok := tt.Find("Item")
	if !ok {
		t.Fatal("item not found")
	}
	cx, cy := middle(center)
	// The item in the middle faces the viewer, so it stays round and full size,
	// unlike the tilted ones around it.
	if center.W != center.H || center.W != m.large {
		t.Fatalf("the middle item is not round and full size: %+v", center)
	}
	// An item beside the middle turns around the vertical axis: it stays as
	// tall, but is narrower than it is tall.
	side, _ := tt.Find("p01")
	if sx, _ := middle(side); sx <= cx || side.W >= side.H {
		t.Fatalf("the item to the right did not turn away: %+v", side)
	}
	// An item directly above turns around the horizontal axis: it stays as wide,
	// but is shorter than it is wide.
	above := ui.Rect{}
	for _, item := range items[1:] {
		r, _ := tt.Find(item.ID)
		if x, y := middle(r); math.Abs(float64(x-cx)) <= 1 && y < cy-1 {
			above = r
			break
		}
	}
	if above.W == 0 {
		t.Fatal("no item directly above the middle")
	}
	if above.H >= above.W {
		t.Fatalf("the item above did not turn away: %+v", above)
	}
}

func TestHoneycombCurvatureScalesTheTiltAndEdge(t *testing.T) {
	items := honeycombItems(19)
	// build returns the item beside the middle and the item furthest above
	// it, which nears the top edge of the view.
	build := func(curvature float32) (side, top ui.Rect) {
		tt := ui.NewTester(func(c *ui.Context) {
			c.SetTheme(theme.Palette(false))
			ui.Column(c).Fill().AlignItems(ui.Center).Children(func() {
				Honeycomb(c, HoneycombOptions{Items: items, Selected: "p00", Curvature: curvature})
			})
		}, 600, 300)
		settleHoneycomb(tt)
		center, _ := tt.Find("p00")
		cx, cy := middle(center)
		side, _ = tt.Find("p01")
		for _, item := range items[1:] {
			r, _ := tt.Find(item.ID)
			if x, y := middle(r); math.Abs(float64(x-cx)) <= 1 && y < cy-1 && (top.W == 0 || r.Y < top.Y) {
				top = r
			}
		}
		if side.W == 0 || top.W == 0 {
			t.Fatal("items not found")
		}
		return side, top
	}
	// More curvature turns the item beside the middle further away, so it is
	// narrower for its height.
	flatSide, flatTop := build(-1)
	side, top := build(0)
	curvedSide, curvedTop := build(2)
	// A flat honeycomb has no lens either: every item is its unmagnified size,
	// unmagnified spacing apart, to show many at once.
	m := honeycombTestMetrics()
	if flatSide.W != flatSide.H || flatSide.W != m.small || flatTop.W != m.small {
		t.Fatalf("a flat honeycomb magnified or tilted items: %+v %+v", flatSide, flatTop)
	}
	if squash := func(r ui.Rect) float32 { return r.W / r.H }; !(squash(curvedSide) < squash(side) && squash(side) < squash(flatSide)) {
		t.Fatalf("curvature did not scale the tilt: flat %+v, default %+v, curved %+v", flatSide, side, curvedSide)
	}
	// Turning around the horizontal axis keeps the top item's width, which the
	// lens and the edge shrink more the more the honeycomb curves.
	if !(curvedTop.W < top.W && top.W < flatTop.W) {
		t.Fatalf("curvature did not scale the edge: flat %+v, default %+v, curved %+v", flatTop, top, curvedTop)
	}
	if oneSide, oneTop := build(1); oneSide != side || oneTop != top {
		t.Fatalf("zero curvature differs from 1")
	}
}

func TestHoneycombGapSpacesItems(t *testing.T) {
	items := honeycombItems(7)
	var gap float32
	// A flat honeycomb draws items unmagnified, so neighbors are an item's
	// size and the gap apart.
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.Palette(false))
		ui.Column(c).Fill().Children(func() {
			Honeycomb(c, HoneycombOptions{Items: items, Selected: "p00", Curvature: -1, Gap: gap})
		})
	}, 600, 300)
	m := honeycombTestMetrics()
	apart := func() (float32, ui.Rect) {
		tt.Frame()
		center, _ := tt.Find("p00")
		side, _ := tt.Find("p01")
		cx, _ := middle(center)
		x, _ := middle(side)
		return x - cx, center
	}
	settleHoneycomb(tt)
	if d, _ := apart(); math.Abs(float64(d-m.small-honeycombGap)) > .01 {
		t.Fatalf("zero gap not the default: %v", d)
	}
	// A new gap rescales the honeycomb at once, keeping the selection centered.
	_, before := apart()
	for _, c := range []struct{ gap, want float32 }{{20, 20}, {-1, 0}, {1000, honeycombGapMax}} {
		gap = c.gap
		d, center := apart()
		if math.Abs(float64(d-m.small-c.want)) > .01 {
			t.Fatalf("gap %v set neighbors %v apart, want %v", c.gap, d, m.small+c.want)
		}
		if center != before {
			t.Fatalf("gap %v moved the selection: %+v, was %+v", c.gap, center, before)
		}
	}
}

func TestHoneycombBuildsContent(t *testing.T) {
	var sizes []float32
	content := func(c *ui.Context, size float32) {
		sizes = append(sizes, size)
		ui.Box(c).Size(size, size)
	}
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.Palette(false))
		sizes = sizes[:0]
		Honeycomb(c, HoneycombOptions{Items: []HoneycombItem{{ID: "a", Label: "A", Content: content},
			{ID: "b", Label: "B", Content: content}}, Selected: "a", SelectedLabel: "Item"})
	}, 600, 300)
	m := honeycombTestMetrics()
	tt.Frame()
	// Content fills the cell's circle itself.
	if len(sizes) != 2 || math.Abs(float64(sizes[0]-m.large)) > .01 || sizes[1] >= sizes[0] {
		t.Fatalf("content not built at each item's size: %v", sizes)
	}
	if err := tt.Click("B"); err != nil {
		t.Fatal(err)
	}
}

func TestHoneycombCentersItsContent(t *testing.T) {
	// A mark in the middle of the content finds whether the item is drawn
	// centered in its round cell.
	content := func(c *ui.Context, size float32) {
		ui.Box(c).Fill().Center().Children(func() {
			ui.Box(c).Size(size/2, size/2).Background(ui.RGB(255, 0, 0))
		})
	}
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.Palette(false))
		Honeycomb(c, HoneycombOptions{Items: []HoneycombItem{{ID: "a", Label: "A", Content: content},
			{ID: "b", Label: "B", Content: content}}, Selected: "a", SelectedLabel: "Item"})
	}, 600, 300)
	tt.Frame()
	itemLogo, ok := tt.Find("Item")
	if !ok {
		t.Fatal("item not found")
	}
	img := tt.Image()
	minX, minY, maxX, maxY, found := 0, 0, 0, 0, false
	for y := int(itemLogo.Y); y < int(itemLogo.Y+itemLogo.H); y++ {
		for x := int(itemLogo.X); x < int(itemLogo.X+itemLogo.W); x++ {
			if p := img.RGBAAt(x, y); p.R > 200 && p.G < 60 && p.B < 60 {
				if !found {
					minX, minY, maxX, maxY, found = x, y, x, y, true
				} else {
					minX, minY = min(minX, x), min(minY, y)
					maxX, maxY = max(maxX, x), max(maxY, y)
				}
			}
		}
	}
	if !found {
		t.Fatal("the content's mark was not drawn")
	}
	mx, my := float32(minX+maxX+1)/2, float32(minY+maxY+1)/2
	cx, cy := middle(itemLogo)
	if math.Abs(float64(mx-cx)) > 1 || math.Abs(float64(my-cy)) > 1 {
		t.Fatalf("content off center: mark %v,%v, item %v,%v", mx, my, cx, cy)
	}
}

func TestHoneycombClicksOnlyInsideTheCircle(t *testing.T) {
	items := honeycombItems(7)
	selected := "p00"
	tt := honeycombTester(items, &selected)
	tt.Frame()
	r, _ := tt.Find("p01")
	tt.ClickAt(r.X+2, r.Y+2)
	tt.Frame()
	if selected != "p00" {
		t.Fatalf("click outside the round item chose %q", selected)
	}
	tt.ClickAt(middle(r))
	tt.Frame()
	if selected != "p01" {
		t.Fatalf("click inside the round item chose %q", selected)
	}
}

func TestHoneycombKeyboardChoosesItems(t *testing.T) {
	for _, dark := range []bool{false, true} {
		items := honeycombItems(5)
		selected := "p00"
		tt := ui.NewTester(func(c *ui.Context) {
			c.SetTheme(theme.Palette(dark))
			ui.Column(c).Fill().Children(func() {
				if id := Honeycomb(c, HoneycombOptions{Items: items, Selected: selected, SelectedLabel: "Item"}); id != "" {
					selected = id
				}
			})
		}, 600, 300)
		tt.Frame()
		// Tab reaches each item in order; Enter chooses the focused one.
		tt.Key(0, ui.KeyTab)
		tt.Key(0, ui.KeyTab)
		tt.Frame()
		tt.Key(0, ui.KeyEnter)
		tt.Frame()
		if selected != "p01" {
			t.Fatalf("dark=%v: keyboard chose %q, want p01", dark, selected)
		}
	}
}

func TestHoneycombItemMenu(t *testing.T) {
	items := honeycombItems(5)
	opened := ""
	for i := range items {
		id := items[i].ID
		items[i].Menu = func(m *ui.Menu) {
			if m.Item("Open " + id).Chosen() {
				opened = id
			}
		}
	}
	selected := "p00"
	tt := honeycombTester(items, &selected)
	tt.Frame()
	// A right-click opens the item's own menu without choosing it.
	if err := tt.RightClick("p02"); err != nil {
		t.Fatal(err)
	}
	if menu := tt.Menu(); !slices.Equal(menu, []string{"Open p02"}) {
		t.Fatalf("menu %v", menu)
	}
	if err := tt.ChooseMenuItem("Open p02"); err != nil {
		t.Fatal(err)
	}
	if opened != "p02" || selected != "p00" {
		t.Fatalf("opened %q, selected %q", opened, selected)
	}
	// Shift+F10 opens the focused item's menu.
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyTab)
	tt.Key(ui.Shift, ui.KeyF10)
	if menu := tt.Menu(); !slices.Equal(menu, []string{"Open p01"}) {
		t.Fatalf("keyboard menu %v", menu)
	}
}

func TestHoneycombRingsInTheItemsColor(t *testing.T) {
	ring := ui.RGB(0, 200, 0)
	rings := func(items []HoneycombItem) int {
		tt := ui.NewTester(func(c *ui.Context) {
			c.SetTheme(theme.Palette(false))
			Honeycomb(c, HoneycombOptions{Items: items, Selected: "a", SelectedLabel: "Item"})
		}, 600, 300)
		settleHoneycomb(tt)
		itemLogo, ok := tt.Find("Item")
		if !ok {
			t.Fatal("item not found")
		}
		img, n := tt.Image(), 0
		for y := int(itemLogo.Y); y < int(itemLogo.Y+itemLogo.H); y++ {
			for x := int(itemLogo.X); x < int(itemLogo.X+itemLogo.W); x++ {
				// The ring is translucent, so look for its tint rather than its
				// solid color.
				if p := img.RGBAAt(x, y); p.G > 40 && int(p.G) > int(p.R)+25 && int(p.G) > int(p.B)+25 {
					n++
				}
			}
		}
		return n
	}
	items := honeycombItems(2)
	items[0].ID, items[0].Label = "a", "A"
	if n := rings(items); n != 0 {
		t.Fatalf("the accent ring drew %d pixels of the item's color", n)
	}
	items[0].Ring = ring
	if n := rings(items); n < 100 {
		t.Fatalf("the ring drew only %d pixels of the item's color", n)
	}
	// An overlay lies over the ring, so one covering the cell hides it; the
	// translucent ring would tint white.
	items[0].Overlay = func(c *ui.Context, size float32) {
		ui.Box(c).Absolute().Left(0).Top(0).Size(size, size).Background(ui.RGB(255, 255, 255))
	}
	if n := rings(items); n != 0 {
		t.Fatalf("the ring drew %d pixels over the overlay", n)
	}
}

func TestHoneycombScalesWithThemeSpacing(t *testing.T) {
	measure := func(spacing float32) ui.Rect {
		tt := ui.NewTester(func(c *ui.Context) {
			th := theme.Palette(false)
			th.Spacing = spacing
			c.SetTheme(th)
			ui.Column(c).Fill().Children(func() {
				Honeycomb(c, HoneycombOptions{Items: honeycombItems(1), Selected: "p00", SelectedLabel: "Item"})
			})
		}, 600, 300)
		tt.Frame()
		r, ok := tt.Find("Item")
		if !ok {
			t.Fatal("item not found")
		}
		return r
	}
	base, wide := measure(4), measure(8)
	if wide.W != 2*base.W || wide.H != 2*base.H {
		t.Fatalf("items did not scale with theme spacing: %+v %+v", base, wide)
	}
}

func TestHoneycombWithoutSelectedItemDrawsNothing(t *testing.T) {
	drawn := false
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.Palette(false))
		Honeycomb(c, HoneycombOptions{Items: honeycombItems(3), Selected: "missing", SelectedLabel: "Item"})
		drawn = true
	}, 600, 300)
	tt.Frame()
	if !drawn {
		t.Fatal("frame not built")
	}
	if _, ok := tt.Find("Item"); ok {
		t.Fatal("an unknown selection still drew items")
	}
}

func TestHoneycombCentersTheSelectionInItsView(t *testing.T) {
	width := float32(600)
	// Every painted frame of the selected item, so a frame placed at a size the
	// view no longer has shows as a jump.
	var painted []ui.Rect
	items := honeycombItems(7)
	items[0].Content = func(c *ui.Context, size float32) {
		ui.Box(c).Fill().Draw(func(_ *ui.Painter, r ui.Rect) { painted = append(painted, r) })
	}
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.Palette(false))
		ui.Column(c).Fill().Children(func() {
			ui.Column(c).Width(width).Grow(1).Children(func() {
				Honeycomb(c, HoneycombOptions{Items: items, Selected: "p00"})
			})
		})
	}, 800, 300)
	check := func(when string) {
		t.Helper()
		if len(painted) == 0 {
			t.Fatalf("selection not painted %s", when)
		}
		for _, r := range painted {
			if x, y := middle(r); math.Abs(float64(x-width/2)) > 1 || math.Abs(float64(y-150)) > 1 {
				t.Fatalf("selection painted off center %s: %+v in %v by 300", when, r, width)
			}
		}
		painted = painted[:0]
	}
	// The first frame has no layout yet to measure.
	tt.Frame()
	check("on entry")
	width = 760
	tt.Frame()
	check("after a resize")
}

func TestHoneycombSlidesAsItemsComeAndGo(t *testing.T) {
	items := honeycombItems(7)
	selected := "p03"
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.Palette(false))
		ui.Column(c).Fill().Children(func() {
			Honeycomb(c, HoneycombOptions{Items: items, Selected: selected, SelectedLabel: "Item"})
		})
	}, 600, 300)
	tt.Frame()
	center, _ := tt.Find("Item")
	cx, cy := middle(center)
	centered := func() bool {
		r, _ := tt.Find("Item")
		x, y := middle(r)
		return math.Abs(float64(x-cx)) <= 1 && math.Abs(float64(y-cy)) <= 1
	}
	// Removing an item before the selection slides the later ones into the
	// cells before them, the selection with them, which stays in the middle.
	before, _ := tt.Find("p05")
	items = slices.Delete(items, 1, 2)
	tt.Frame()
	tt.Advance(honeycombDuration / 2)
	sliding, _ := tt.Find("p05")
	settleHoneycomb(tt)
	after, _ := tt.Find("p05")
	if sliding == before || sliding == after {
		t.Fatalf("item did not slide: before %+v, sliding %+v, after %+v", before, sliding, after)
	}
	if !centered() {
		t.Fatal("selection left the middle as items slid")
	}
	// Adding an item and selecting it pans there smoothly.
	items = append(items, honeycombItems(9)[8])
	selected = "p08"
	tt.Frame()
	tt.Advance(honeycombDuration / 2)
	if centered() {
		t.Fatal("honeycomb jumped to the added item")
	}
	settleHoneycomb(tt)
	if !centered() {
		t.Fatal("added item not centered")
	}
}

func TestHoneycombJumpsWhenAnimationsAreDisabled(t *testing.T) {
	items := honeycombItems(9)
	selected, instant := "p00", false
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.Palette(false))
		if instant {
			c.DisableAnimations()
		}
		ui.Column(c).Fill().Children(func() {
			Honeycomb(c, HoneycombOptions{Items: items, Selected: selected, SelectedLabel: "Item"})
		})
	}, 600, 300)
	settleHoneycomb(tt)
	center, _ := tt.Find("Item")
	selected, instant = "p08", true
	tt.Frame()
	if r, _ := tt.Find("Item"); r != center {
		t.Fatalf("selection glided with animations disabled: %+v, middle %+v", r, center)
	}
}

func TestHoneycombCaptionsTheMiddleItem(t *testing.T) {
	items := honeycombItems(7)
	for i := range items {
		items[i].Caption = "Caption " + items[i].ID
	}
	items[3].Caption = ""
	selected := "p00"
	tt := honeycombTester(items, &selected)
	settleHoneycomb(tt)
	// shows returns the captions shown.
	shows := func() []string {
		return slices.DeleteFunc(tt.Texts(), func(s string) bool { return !strings.HasPrefix(s, "Caption ") })
	}
	captioned := func(id string) ui.Rect {
		t.Helper()
		item, _ := tt.Find("Item")
		caption, _ := tt.Find("Caption " + id)
		if got := shows(); !slices.Equal(got, []string{"Caption " + id}) {
			t.Fatalf("captions %v, want %s's alone", got, id)
		}
		ix, _ := middle(item)
		if x, _ := middle(caption); caption.W == 0 || math.Abs(float64(x-ix)) > 1 || caption.Y < item.Y+item.H {
			t.Fatalf("%s's caption not under it: %+v, item %+v", id, caption, item)
		}
		return caption
	}
	caption := captioned("p00")
	// The caption passes to the chosen item, and none shows under an item
	// without one.
	selected = "p04"
	tt.Frame()
	settleHoneycomb(tt)
	captioned("p04")
	selected = "p03"
	tt.Frame()
	settleHoneycomb(tt)
	if got := shows(); len(got) != 0 {
		t.Fatalf("an item without a caption shows %v", got)
	}
	// A click on the caption over an item below chooses that item.
	selected = "p00"
	tt.Frame()
	settleHoneycomb(tt)
	x, y := caption.X+caption.W-2, caption.Y+caption.H/2
	under := ""
	for _, item := range items[1:] {
		r, _ := tt.Find(item.ID)
		if cx, cy := middle(r); math.Hypot(float64(x-cx), float64(y-cy)) < float64(r.W/2)-2 {
			under = item.ID
		}
	}
	if under == "" {
		t.Fatal("the caption lies over no item")
	}
	tt.ClickAt(x, y)
	tt.Frame()
	if selected != under {
		t.Fatalf("a click on the caption over %s chose %q", under, selected)
	}
}

func TestHoneycombGroupsImages(t *testing.T) {
	var sizes []float32
	image := func(i int) HoneycombImage {
		return HoneycombImage{Content: func(c *ui.Context, size float32) {
			sizes = append(sizes, size)
			ui.Box(c).Size(size, size).Radius(size / 2).Background(ui.RGB(255, 255, 255)).Label(fmt.Sprint("image", i))
		}}
	}
	var images []HoneycombImage
	for i := range 6 {
		images = append(images, image(i))
	}
	build := func(images []HoneycombImage) *ui.Tester {
		tt := ui.NewTester(func(c *ui.Context) {
			c.SetTheme(theme.Palette(false))
			sizes = sizes[:0]
			Honeycomb(c, HoneycombOptions{Items: []HoneycombItem{{ID: "a", Label: "A", Images: images, Ring: ui.RGB(0, 200, 0)},
				{ID: "b", Label: "B"}}, Selected: "a", SelectedLabel: "Item"})
		}, 600, 300)
		settleHoneycomb(tt)
		return tt
	}
	m := honeycombTestMetrics()

	// One image fills the cell alone.
	build(images[:1])
	if len(sizes) != 1 || math.Abs(float64(sizes[0]-m.large)) > .01 {
		t.Fatalf("a single image not built at the cell's size: %v", sizes)
	}

	// More than four show the first three, the first largest and in front,
	// and a badge counting the rest; each wears the ring.
	tt := build(images)
	if len(sizes) != 3 || sizes[2] <= sizes[0] || sizes[2] <= sizes[1] || sizes[2] >= m.large/2 {
		t.Fatalf("group sizes: %v", sizes)
	}
	if !tt.HasText("+3") || tt.HasText("image3") {
		t.Fatalf("group: %v", tt.Texts())
	}
	img := tt.Image()
	for i := range 3 {
		r, ok := tt.Find(fmt.Sprint("image", i))
		if !ok {
			t.Fatalf("image %d missing", i)
		}
		// The ring tints the image's edge just inside the background's.
		n := 0
		for y := int(r.Y) - 3; y < int(r.Y+r.H)+3; y++ {
			for x := int(r.X) - 3; x < int(r.X+r.W)+3; x++ {
				if p := img.RGBAAt(x, y); p.G > 40 && int(p.G) > int(p.R)+25 && int(p.G) > int(p.B)+25 {
					n++
				}
			}
		}
		if n < 20 {
			t.Fatalf("image %d wears %d ring pixels", i, n)
		}
	}

	// Four show every image without a badge.
	tt = build(images[:4])
	if len(sizes) != 4 || tt.HasText("+1") {
		t.Fatalf("four images: %v %v", sizes, tt.Texts())
	}
}

func mustFind(t *testing.T, tt *ui.Tester, label string) ui.Rect {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok {
		t.Fatalf("no %q", label)
	}
	return r
}
