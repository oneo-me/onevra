package gallery

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/shells/sidebar"
	"github.com/oneo-me/onevra/source/theme"
)

// sectionExample returns a component shown in a section of a gathered page.
func sectionExample(name string) controlExample {
	for _, sections := range controlSections {
		for _, section := range sections {
			if section.name == name {
				return section
			}
		}
	}
	return controlExample{}
}

// controlComponents lists every component example, with gathered pages
// replaced by their sections.
func controlComponents() []controlExample {
	var examples []controlExample
	for _, example := range controlCatalog {
		if sections := controlSections[example.name]; sections != nil {
			examples = append(examples, sections...)
		} else {
			examples = append(examples, example)
		}
	}
	return examples
}

func TestControlPages(t *testing.T) {
	if len(controlCatalog) != 46 {
		t.Fatal("missing control examples")
	}
	for i, example := range controlCatalog {
		t.Run(example.name, func(t *testing.T) {
			page := controlPageStart() + galleryPage(i)
			if got, ok := controlExampleFor(page); !ok || got.name != example.name {
				t.Fatal("incorrect route")
			}
			if _, ok := mygoExampleFor(page); ok {
				t.Fatal("route overlaps MyGo")
			}
			for _, width := range []int{540, 280} {
				for _, dark := range []bool{false, true} {
					g := &Gallery{page: page}
					tt := ui.NewTester(func(c *ui.Context) { c.SetTheme(theme.Palette(dark)); g.View(c) }, width, 1800)
					title := ui.NewTester(func(c *ui.Context) { c.SetTheme(theme.Palette(dark)); g.Titlebar(c) }, width, 40)
					wantStates, labels := len(example.variants), example.variants
					if example.name == "Data charts" {
						wantStates = 1 // The charts share their data actions.
					}
					if sections := controlSections[example.name]; sections != nil {
						wantStates, labels = 0, nil
						for _, section := range sections {
							wantStates += len(section.variants)
							labels = append(labels, section.name)
							if len(section.variants) > 1 {
								labels = append(labels, section.variants...)
							}
						}
					}
					if tt.Image() == nil || len(g.controlStates) != wantStates {
						t.Fatal("missing page or variant")
					}
					if tt.HasText("Disabled") {
						t.Fatal("unexpected disabled example")
					}
					if !title.HasText(example.name) {
						t.Fatalf("missing page title %s", example.name)
					}
					// The titlebar names the page; content repeats the name
					// only as a variant or section label, or where the
					// component itself shows or announces it.
					ownName := slices.Contains([]string{"Breadcrumbs", "Chat", "Rating", "Select"}, example.name)
					if tt.HasText(example.name) && !ownName && !slices.Contains(labels, example.name) {
						t.Fatalf("content repeats the page title %s", example.name)
					}
					if len(labels) > 1 {
						for _, label := range labels {
							if !tt.HasText(label) {
								t.Fatalf("missing variant or section %s", label)
							}
						}
					}
				}
			}
		})
	}
	for _, page := range []galleryPage{shellPage, lucidePage, mygoPageStart, controlPageStart() - 1, controlPageStart() + galleryPage(len(controlCatalog))} {
		if _, ok := controlExampleFor(page); ok {
			t.Fatalf("unexpected route %d", page)
		}
	}
}

// shellFrame builds a page below its titlebar, as Shell overlays it.
func (g *Gallery) shellFrame(c *ui.Context) {
	ui.Box(c).Fill().Children(func() {
		g.View(c)
		ui.Row(c).Absolute().Top(0).Left(0).Right(0).Height(sidebar.HeaderHeight(c)).AlignItems(ui.Center).Children(func() { g.Titlebar(c) })
	})
}

// The SplitPercent cell needs a bounded width: sized to its content, the
// split grew with the panes and the divider outran the pointer.
func TestSplitPercentDragFollowsPointer(t *testing.T) {
	g := &Gallery{}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(24).Children(func() { g.controlVariants(c, sectionExample("SplitPercent"), true) })
	}, 900, 1200)
	r, ok := tt.Find("Horizontal divider")
	if !ok {
		t.Fatal("missing divider")
	}
	x, y := r.X+r.W/2, r.Y+r.H/2
	tt.Press(x, y)
	for i := 1; i <= 20; i++ {
		px := x + float32(i*5)
		tt.Move(px, y)
		if d, _ := tt.Find("Horizontal divider"); d.X+d.W/2 < px-1 || d.X+d.W/2 > px+1 {
			t.Fatalf("step %d: divider at %v, pointer at %v", i, d.X+d.W/2, px)
		}
	}
}

func TestControlNavigationAndState(t *testing.T) {
	g := &Gallery{}
	tt := ui.NewTester(g.Left, 220, 6000)
	for i, example := range controlCatalog {
		if err := tt.Click(example.label); err != nil {
			t.Fatal(err)
		}
		if g.page != controlPageStart()+galleryPage(i) {
			t.Fatal("incorrect navigation")
		}
		title := ui.NewTester(g.Titlebar, 540, 60)
		if !title.HasText(example.group + " / " + example.name) {
			t.Fatal("incorrect title")
		}
	}
	off, on := g.controlState("Checkbox", "Off"), g.controlState("Checkbox", "On")
	if off.on || !on.on {
		t.Fatal("incorrect initial states")
	}
	off.on = true
	if g.controlState("Checkbox", "Off") != off || g.controlState("Switch", "Off").on {
		t.Fatal("state lost or shared between pages")
	}
	on.on = false
	if !off.on {
		t.Fatal("state shared between variants")
	}
	if g.mygoState("Checkbox", "Off", true).on {
		t.Fatal("state shared with MyGo")
	}
}

func TestEditableTextExample(t *testing.T) {
	g := &Gallery{}
	s := g.controlState("Text", "Editable")
	tt := ui.NewTester(func(c *ui.Context) { controlText(c, s, "Editable") }, 540, 100)
	tt.Key(0, ui.KeyTab)
	tt.Type("Edited name")
	if s.text != "Edited name" {
		t.Fatalf("text binding not updated: %q", s.text)
	}
	if g.controlState("Text", "Default").text != "Hello Onevra" {
		t.Fatal("text variants share state")
	}
	if mygoHandled["EditableText"].Replacement != "components.Text" {
		t.Fatal("EditableText is not merged into Text")
	}
}

func TestControlExampleInteractions(t *testing.T) {
	for _, example := range controlComponents() {
		if example.name == "Selection" {
			continue // Grouped selection interactions are covered separately.
		}
		for _, disabled := range []bool{false, true} {
			t.Run(example.name+map[bool]string{false: "/enabled", true: "/disabled"}[disabled], func(t *testing.T) {
				s := (&Gallery{}).controlState(example.name, "Empty")
				tt := ui.NewTester(func(c *ui.Context) {
					c.SetTheme(theme.LightPalette())
					ui.Column(c).Fill().Gap(12).Disabled(disabled).Children(func() {
						if toolbar := controlToolbars[example.name]; toolbar != nil {
							ui.Row(c).Children(func() { toolbar(c, s) })
						}
						variant := "Default"
						switch example.name {
						case "Data charts":
							variant = "Bar chart"
						case "AvatarGroup":
							variant = "Changing members"
						case "Badge":
							variant = "Live updates"
						case "Skeleton":
							variant = "Content"
						}
						example.build(c, s, variant)
					})
				}, 540, 900)
				click := func(label string) {
					t.Helper()
					if err := tt.Click(label); err != nil {
						t.Fatal(err)
					}
				}
				changed := false
				switch example.name {
				case "Data charts":
					click("Update data")
					legend := tt.HasText("Current")
					click("Hide legend")
					changed = chartDemo(s).tile("Bar chart").round == 1 && legend && !tt.HasText("Current")
				case "Relationship graph":
					click("Add note")
					changed = s.graph.added == 1
				case "Treemap":
					click("Update data")
					changed = s.treemap.updated
				case "Flow diagram":
					click("Add a step")
					changed = len(s.flow.added) == 1
				case "Calendar":
					before := s.day
					tt.Key(0, ui.KeyTab)
					tt.Key(0, ui.KeyRight)
					tt.Key(0, ui.KeyEnter)
					changed = s.day != before
				case "TimePicker":
					click("Time")
					_, changed = tt.Find("Hours")
				case "CalendarPicker":
					click("Date")
					_, changed = tt.Find("Next month")
				case "ColorView":
					before := s.color
					click("Saturation and brightness")
					tt.Key(0, ui.KeyRight)
					changed = s.color != before
				case "ColorPicker":
					click("Default color")
					_, changed = tt.Find("Hue")
				case "Tooltip":
					tt.Key(0, ui.KeyTab)
					changed = tt.HasText("A helpful description of this action")
				case "Toast":
					click("Info")
					changed = tt.HasText("A new version is available")
				case "Popover":
					click("Show popover")
					changed = s.dialogOpen
				case "Dialog":
					click("Open dialog")
					changed = s.dialogOpen
				case "AlertDialog":
					click("Show Default")
					changed = s.dialogOpen
				case "Chat":
					click("Attach a sample file")
					changed = len(s.chat.files) == 1
				case "RichText":
					click("inline action")
					changed = s.index == 1
				case "Breadcrumbs":
					click("Components")
					changed = s.index == 1
				case "Rating":
					tt.Key(0, ui.KeyTab)
					tt.Key(0, ui.KeyRight)
					changed = s.index == 1
				case "AvatarGroup":
					click("Remove member")
					changed = s.index == 4
				case "Badge":
					click("Next status")
					changed = s.badge.status == 1 && tt.HasText("Syncing")
				case "Image", "Spinner", "Avatar":
					return // Display-only components have no interaction state.
				case "ScrollView":
					return // ScrollView interactions are covered in the component tests.
				case "Text":
					if !tt.HasText("Count: 42") || !tt.HasText("Inline bold") {
						t.Fatal("missing display text")
					}
					return // Display text has no interaction state.
				case "Menu":
					click("Actions")
					if !disabled {
						if err := tt.ChooseMenuItem("Open"); err != nil {
							t.Fatal(err)
						}
					}
					changed = s.index == 1
				case "Button":
					click("Primary")
					changed = s.button.clicks[0] == 1
				case "HoldToConfirm":
					r, ok := tt.Find("Hold to archive")
					if !ok {
						t.Fatal("missing hold button")
					}
					tt.Press(r.X+r.W/2, r.Y+r.H/2)
					tt.Advance(1300 * time.Millisecond)
					tt.Frame()
					tt.Release(r.X+r.W/2, r.Y+r.H/2)
					changed = s.on
				case "TextBox", "TextBox (multiline)":
					label := "Name"
					if example.name == "TextBox (multiline)" {
						label = "Notes"
					}
					click(label)
					tt.Type("Hello")
					changed = s.text == "Hello"
				case "TextBoxBuffer", "TextBoxBuffer (multiline)":
					label := "Name"
					if example.name == "TextBoxBuffer (multiline)" {
						label = "Notes"
					}
					click(label)
					tt.Type("Hello")
					changed = s.buffer.String() == "Hello"
				case "Select":
					click("Size")
					if !disabled {
						click("Large")
					}
					changed = s.choice == "Large"
				case "Progress", "ProgressRing":
					click("Default value")
					tt.Key(0, ui.KeyRight)
					changed = s.value != 40
				case "Split":
					clickDivider(tt)
					tt.Key(0, ui.KeyRight)
					changed = s.split != 100
				case "SplitPercent":
					clickDivider(tt)
					tt.Key(0, ui.KeyRight)
					changed = s.splitPercent != 30
				case "TagInput":
					click("Topics")
					tt.Type("native")
					tt.Key(0, ui.KeyEnter)
					changed = slices.Contains(s.tags, "native")
				case "ImageCompare":
					click("Default comparison")
					tt.Key(0, ui.KeyRight)
					changed = s.compare != 50
				case "NumberInput":
					click("Quantity")
					tt.Key(0, ui.KeyUp)
					changed = s.value == 41
				case "Slider":
					click("Volume")
					tt.Key(0, ui.KeyRight)
					changed = s.value != 40
				case "Tree":
					click("source")
					click("source")
					changed = s.on
				case "Table":
					click("go.mod")
					changed = s.choice == "go.mod"
				case "List", "GridView":
					click("Item 2")
					changed = s.index == 1
				case "Sidebar":
					click("Files")
					changed = s.index == 1
				case "Floating button group":
					click("Comment")
					changed = s.floating.edits == 1
				case "Card stack":
					click("Shortlist")
					changed = len(s.cards.Decisions) == 1
				case "Honeycomb":
					click("Images")
					tt.Frame()
					changed = s.honeycomb == "images"
				case "Accordion":
					click("General")
					changed = s.toggles[0]
				case "Collapsible":
					click("Details")
					changed = s.on
				case "Skeleton":
					click("Loading")
					changed = s.on
				case "Carousel":
					click("Next slide")
					changed = s.carousel.Index == 1
				case "Tabs":
					click("Style")
					changed = s.index == 1
				case "Stepper":
					click("Next")
					changed = s.index == 1
				case "ScrollEdgeBar":
					click("Search")
					changed = s.index == 1
				case "Shader backgrounds":
					click("Pause drift")
					changed = s.shader.still
				}
				if changed == disabled {
					t.Fatalf("changed=%t, disabled=%t", changed, disabled)
				}
			})
		}
	}
}

func TestControlVariantLayout(t *testing.T) {
	for _, name := range []string{"Text", "TextBox", "TextBox (multiline)", "Select", "Tabs", "Collapsible", "Image", "Slider"} {
		for _, dark := range []bool{false, true} {
			for _, width := range []int{900, 280} {
				g := &Gallery{}
				example := sectionExample(name)
				if example.build == nil {
					example, _ = controlExampleFor(pageFromKey("controls/" + name))
				}
				tt := ui.NewTester(func(c *ui.Context) {
					c.SetTheme(theme.Palette(dark))
					ui.Column(c).Fill().Padding(24).Children(func() { g.controlVariants(c, example, false) })
				}, width, 1200)
				var previous ui.Rect
				for i, variant := range example.variants {
					label := variant
					r, ok := tt.Find(label)
					if !ok || r.X < 20 || r.X+r.W > float32(width)-20 {
						t.Fatalf("%s/%s outside preview: %+v", name, variant, r)
					}
					if i > 0 {
						vertical := name == "Text" || name == "Collapsible" || name == "Tabs"
						if vertical && r.Y < previous.Y+previous.H {
							t.Fatalf("%s variants are not vertical", name)
						}
						if !vertical && width == 900 && i == 1 && r.Y != previous.Y {
							t.Fatalf("%s variants are not horizontal", name)
						}
						if r.Y < previous.Y+previous.H && r.X < previous.X+previous.W {
							t.Fatalf("%s variants overlap", name)
						}
					}
					previous = r
				}
				for _, text := range []string{"Selected:", "characters", "General settings", "Style settings", "Grid spacing:"} {
					if tt.HasText(text) {
						t.Fatalf("%s retains extra status text %q", name, text)
					}
				}
				if name == "TextBox" || name == "TextBox (multiline)" {
					label := "Name"
					if name == "TextBox (multiline)" {
						label = "Notes"
					}
					if err := tt.Click(label); err != nil {
						t.Fatal(err)
					}
					tt.Type("Edited")
					if g.controlState(name, "Empty").text != "Edited" || g.controlState(name, "Filled").text == "Edited" || g.controlState(name, "Read only").text == "Edited" {
						t.Fatalf("%s variant values are not independent", name)
					}
				}
			}
		}
	}
}

func TestHoneycombExampleAddRemove(t *testing.T) {
	s := (&Gallery{}).controlState("Honeycomb", "Empty")
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.LightPalette())
		ui.Column(c).Fill().Gap(12).Children(func() {
			ui.Row(c).Children(func() { honeycombToolbar(c, s) })
			controlHoneycomb(c, s, "Default")
		})
	}, 540, 900)
	tt.Frame()
	before := len(s.honeycombItems)
	if before == 0 {
		t.Fatal("no default items")
	}
	if err := tt.Click("Add item"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(s.honeycombItems) != before+1 {
		t.Fatalf("Add item left %d items", len(s.honeycombItems))
	}
	if added := s.honeycombItems[len(s.honeycombItems)-1]; s.honeycomb != added.id {
		t.Fatalf("added item not selected: %q", s.honeycomb)
	}
	selected := func() bool {
		for _, item := range s.honeycombItems {
			if item.id == s.honeycomb {
				return true
			}
		}
		return false
	}
	// A cell's context menu removes that cell.
	target := s.honeycombItems[1] // Images
	if err := tt.RightClick("Images"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Remove item"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if slices.ContainsFunc(s.honeycombItems, func(item honeycombDemoItem) bool { return item.id == target.id }) {
		t.Fatalf("context menu kept %q", target.id)
	}
	// Remove until empty; the selection stays valid as items disappear.
	for len(s.honeycombItems) > 0 {
		count := len(s.honeycombItems)
		if err := tt.Click("Remove item"); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		if len(s.honeycombItems) != count-1 {
			t.Fatalf("Remove item left %d items", len(s.honeycombItems))
		}
		if len(s.honeycombItems) > 0 && !selected() {
			t.Fatalf("selection %q lost after removal", s.honeycomb)
		}
	}
	if s.honeycomb != "" {
		t.Fatalf("empty honeycomb kept selection %q", s.honeycomb)
	}
}

func TestControlReadOnlyExamples(t *testing.T) {
	for _, example := range controlSections["TextBox"][:2] {
		s := (&Gallery{}).controlState(example.name, "Read only")
		before := s.text
		tt := ui.NewTester(func(c *ui.Context) { example.build(c, s, "Read only") }, 540, 300)
		label := "Name"
		if example.name == "TextBox (multiline)" {
			label = "Notes"
		}
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
		tt.Type("Changed")
		if s.text != before {
			t.Fatal("read-only example accepted input")
		}
	}
}

// clickDivider presses a split's handle, which lies over the named divider and
// focuses it.
func clickDivider(tt *ui.Tester) {
	if r, ok := tt.Find("Default divider"); ok {
		tt.ClickAt(r.X+r.W/2, r.Y+r.H/2)
	}
}
