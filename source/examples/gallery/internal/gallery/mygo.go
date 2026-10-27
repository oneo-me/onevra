package gallery

import (
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

const mygoPageStart galleryPage = notificationsPage + 1

type mygoExample struct {
	name, group string
	build       func(*ui.Context, *demoState)
}

var mygoGroups = []string{"Text input", "Content display", "Code & diffs", "Charts", "Feedback & overlays", "Layout & containers", "Games"}

var demoImage = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 180 100"><rect width="180" height="100" rx="12" fill="#dbeafe"/><circle cx="140" cy="28" r="15" fill="#fbbf24"/><path d="M0 100L55 25L110 100Z" fill="#2563eb"/><path d="M65 100L120 45L180 100Z" fill="#059669"/></svg>`))

func mygoExampleFor(page galleryPage) (mygoExample, bool) {
	i := int(page - mygoPageStart)
	if i < 0 || i >= len(mygoCatalog) {
		return mygoExample{}, false
	}
	return mygoCatalog[i], true
}

// Variants are starting states, not frozen mockups: enabled controls remain interactive.
func mygoVariants(name string) []string {
	switch name {
	case "Motion":
		return []string{"Springs", "Keyframes", "Timeline", "Animated pose", "Draggable", "Frame rate"}
	case "Code editor":
		return []string{"Editable", "Read only"}
	case "Diff viewer":
		return []string{"Side by side", "Unified"}
	case "Gradients & blur":
		return []string{"Gradient stops", "Conic", "Borders", "Blur"}
	case "Shaders":
		return []string{"Fragment", "Text and paths", "Element effect", "Backdrop"}
	case "Transforms":
		return []string{"Editable controls", "Perspective"}
	case "Toggle", "ToggleGroup", "Checkbox", "Switch":
		return []string{"Off", "On", "Disabled"}
	case "CheckboxGroup":
		return []string{"Unchecked", "Mixed", "Checked", "Disabled"}
	case "TextInput", "TextArea", "TextInputBuffer", "TextAreaBuffer":
		return []string{"Empty", "Filled", "Read only", "Disabled"}
	case "SearchField", "TokenField":
		return []string{"Empty", "Filled", "Disabled"}
	case "Tree", "Collapsible", "Accordion":
		return []string{"Collapsed", "Expanded"}
	case "Rating":
		return []string{"Unrated", "Rated", "Disabled"}
	case "Radio", "RadioGroup", "Select", "Tabs", "Sidebar":
		return []string{"Default", "Selected", "Disabled"}
	case "Button", "PrimaryButton", "Link",
		"Slider", "StepSlider", "RangeSlider", "Stepper", "NumberInput", "EditableText",
		"Calendar", "DateInput", "TimeInput", "ColorWell", "ColorPicker":
		return []string{"Default", "Disabled"}
	default:
		return []string{"Default"}
	}
}

func (g *Gallery) mygoState(name, variant string, first bool) *demoState {
	if g.mygoStates == nil {
		g.mygoStates = make(map[string]*demoState)
	}
	key := name + "/" + variant
	if first {
		key = name
	}
	if s := g.mygoStates[key]; s != nil {
		return s
	}
	s := newDemoState()
	s.variant = variant
	if name == "Transforms" {
		s.value = 8
	}
	if name == "Audio" {
		s.audio.output, s.audio.volume, s.audio.window = g.audioOutput, .8, g.window
	}
	if name == "Canvas" {
		s.value = 0
	}
	if name == "Gradients & blur" || name == "Shaders" {
		s.value, s.on = 30, true
	}
	s.system.native, s.system.window = g.systemServices, g.window
	s.game.gamepads = g.gamepads
	s.flyout.gallery = g
	switch variant {
	case "Empty":
		s.text = ""
		s.tags = nil
	case "On", "Checked":
		s.on, s.other = true, true
	case "Mixed":
		s.on = true
	case "Selected":
		s.index = 1
		s.choice = "Large"
	case "Rated":
		s.index = 4
	case "Expanded":
		s.open, s.other = true, true
	}
	if name == "Sidebar" {
		s.choice = "Home"
		if variant == "Selected" {
			s.choice = "Files"
		}
	}
	if name == "TextInputBuffer" || name == "TextAreaBuffer" {
		text := s.text
		if name == "TextAreaBuffer" && variant != "Empty" {
			text = "Hello MyGo\nEdit this document with indexed text storage.\n你好，世界 🌍"
		}
		s.buffer = ui.NewTextBuffer(text)
		s.readOnly = variant == "Read only"
	}
	g.mygoStates[key] = s
	return s
}

func (g *Gallery) MyGo(c *ui.Context, example mygoExample) {
	g.mygoHost(c, func() { g.myGoLocalized(c, example) })
}

// mygoHost places a page behind the MyGo entry beside the list that switches
// between them.
func (g *Gallery) mygoHost(c *ui.Context, page func()) {
	g.withLanguage(c, func() {
		g.mygoPage = g.page
		t := c.Theme()
		width, _ := c.Size()
		if g.Shell != nil && g.Shell.LeftExpanded() {
			width -= g.Shell.LeftWidth()
		}
		pane := func() ui.Element {
			if g.exampleGrid() {
				return components.PixelGrid(c, components.PixelGridOptions{Spacing: 22, MajorEvery: -1}).Column()
			}
			return ui.Column(c)
		}
		ui.Column(c).Fill().Children(func() {
			// Separate the list and example from the shell's floating header.
			ui.Box(c).Height(1).FillWidth().Shrink(0).Background(t.Border)
			if width < t.Space(130) {
				// Narrow content keeps the example's width and switches pages
				// with a selector, as Settings does.
				g.mygoSelect(c)
				g.mygoTitlebar(c)
				pane().Fill().MinHeight(0).Children(page)
				return
			}
			ui.Row(c).Fill().MinHeight(0).Children(func() {
				g.mygoList(c)
				ui.Box(c).Width(1).FillHeight().Shrink(0).Background(t.Border)
				ui.Column(c).Fill().MinWidth(0).Children(func() {
					g.mygoTitlebar(c)
					pane().Fill().MinHeight(0).Children(page)
				})
			})
		})
	})
}

// mygoToolbars hold page-level sample actions, which the example's titlebar
// shows after its title so the content stays with the example itself.
var mygoToolbars = map[string]func(*ui.Context, *demoState){
	"Rich text editor": richTextToolbar, "Markdown stream": markdownStreamToolbar, "Terminal": terminalToolbar,
	"Loading indicators": loadingToolbar,
}

// mygoTitlebar names the selected page above it, followed by its actions.
func (g *Gallery) mygoTitlebar(c *ui.Context) {
	t := c.Theme()
	example, _ := mygoExampleFor(g.page)
	toolbar := mygoToolbars[example.name]
	ui.Row(c).Key("mygo-titlebar").FillWidth().Shrink(0).Height(t.Space(11)).Padding(0, t.Space(3)).
		AlignItems(ui.Center).Gap(t.Space(1)).Background(t.Background).Children(func() {
		components.Text(c, g.pageTitle(c)).Bold().SingleLine().MinWidth(0).Shrink(1)
		if toolbar != nil {
			ui.Spacer(c)
			toolbar(c, g.mygoState(example.name, mygoVariants(example.name)[0], true))
		}
	})
	ui.Box(c).Height(1).FillWidth().Shrink(0).Background(t.Border)
}

func (g *Gallery) mygoSelect(c *ui.Context) {
	t := c.Theme()
	groups := mygoNavigationGroups()
	label := ""
	for _, group := range groups {
		for _, entry := range group.entries {
			if entry.page == g.page {
				label = entry.label
			}
		}
	}
	ui.Column(c).Key("mygo-select").FillWidth().Shrink(0).Padding(t.Space(3), t.Space(6), 0, t.Space(6)).Children(func() {
		sel := components.Select(c, &g.page)
		sel.Trigger.Label(tr(c, "officialComponents")).FillWidth().Children(func() { components.Text(c, tr(c, label)).SingleLine() })
		sel.Popup(func(panel ui.Element) {
			for _, group := range groups {
				components.Text(c, tr(c, group.label)).FontSize(t.Rem(.75)).TextColor(t.TextMuted).Padding(t.Space(1.5), t.Space(2.5), t.Space(.5), t.Space(2.5))
				for _, entry := range group.entries {
					sel.Item(entry.page).Label(tr(c, entry.label)).Children(func() { components.Text(c, tr(c, entry.label)) })
				}
			}
		})
	})
}

// lastMyGoPage is the page the single MyGo sidebar entry opens.
func (g *Gallery) lastMyGoPage() galleryPage {
	if isMyGoPage(g.mygoPage) {
		return g.mygoPage
	}
	return mygoNavigationGroups()[0].entries[0].page
}

// mygoList switches between the pages behind the MyGo sidebar entry.
func (g *Gallery) mygoList(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Key("mygo-list").Width(t.Space(52)).FillHeight().Shrink(0).Background(t.Background).Label(tr(c, "officialComponents")).Children(func() {
		components.ScrollView(c, components.ScrollViewOptions{}).Fill().Children(func() {
			components.Sidebar(c).Padding(t.Space(3)).Children(func() {
				for _, group := range mygoNavigationGroups() {
					components.SidebarSection(c, tr(c, group.label), func() {
						for _, entry := range group.entries {
							if components.SidebarItem(c, tr(c, entry.label), entry.icon(), g.page == entry.page).Key(entry.page).Clicked() {
								g.page = entry.page
							}
						}
					})
				}
			})
		})
	})
}

// mygoDocuments fill the example pane as a single document, without variant
// sections; those that scroll themselves own the whole pane.
var mygoDocuments = map[string]bool{
	"Terminal": true, "Rich text editor": true, "Markdown": false, "Markdown stream": false,
}

func (g *Gallery) myGoLocalized(c *ui.Context, example mygoExample) {
	t := c.Theme()
	if selfScrolling, ok := mygoDocuments[example.name]; ok {
		s := g.mygoState(example.name, "Default", true)
		if example.name == "Terminal" && g.createTerminal != nil {
			d := &s.terminal
			d.start(func() (*terminal.Terminal, error) { return g.createTerminal(d.options(g.refresh)) })
		}
		if selfScrolling {
			ui.Column(c).Key("mygo-" + example.name).Fill().Children(func() { example.build(c, s) })
			return
		}
		components.ScrollView(c, components.ScrollViewOptions{}).Key("mygo-" + example.name).Fill().Children(func() {
			ui.Column(c).FillWidth().Padding(t.Space(6)).Children(func() { example.build(c, s) })
		})
		return
	}
	if example.name == "WebView" {
		d := &g.mygoState("WebView", "Default", true).webView
		d.covered = g.settings.open || g.commands.open
		d.openURL = g.openURL
		if g.createWebView != nil {
			d.start(g.createWebView)
		}
	}
	variants := mygoVariants(example.name)
	components.ScrollView(c, components.ScrollViewOptions{}).Key("mygo-" + example.name).Fill().Children(func() {
		ui.Column(c).FillWidth().Padding(t.Space(6)).Gap(t.Space(6)).Children(func() {
			for i, variant := range variants {
				s := g.mygoState(example.name, variant, i == 0)
				ui.Column(c).Key(variant).FillWidth().MinWidth(0).Gap(12).Children(func() {
					if len(variants) > 1 {
						components.Text(c, tr(c, variant)).Bold().TextColor(t.TextMuted)
					}
					ui.Column(c).FillWidth().MinWidth(0).Gap(12).Disabled(variant == "Disabled").Children(func() {
						if variant == "Read only" && (example.name == "TextInput" || example.name == "TextArea") {
							if example.name == "TextInput" {
								components.TextBox(c, &s.text, components.TextBoxOptions{ReadOnly: true}).Label(tr(c, "name")).FillWidth()
							} else {
								components.TextBox(c, &s.text, components.TextBoxOptions{Multiline: true, ReadOnly: true}).Label(tr(c, "notes")).Height(140).FillWidth()
							}
						} else {
							example.build(c, s)
						}
					})
				})
			}
		})
	})
}
