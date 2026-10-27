package gallery

import (
	"sort"
)

//go:generate go run ../cmd/mygocomponents

type mygoDecision struct {
	Replacement     string // Empty means excluded, not awaiting implementation.
	ReplacementName string // Optional display name for a component package.
	Reason          string
}

// Record decisions here, never in the generated inventory. Both replacement and
// exclusion decisions hide upstream examples and suppress pending reminders.
// To exclude a component, add {Reason: "why this component is not needed"}.
var mygoHandled = map[string]mygoDecision{
	"Charts":               {Replacement: "components/chart", ReplacementName: "Charts", Reason: "Onevra categorical bar, line, area, stacked bar and donut views compose ui/chart scales, marks, paths and selection; continuous axes and specialized charts remain upstream primitives"},
	"TimeInput":            {Replacement: "components.TimePicker", Reason: "Clock-icon trigger with hour and minute popup columns"},
	"Calendar":             {Replacement: "components.Calendar", Reason: "Fixed-size calendar view, also used inside CalendarPicker"},
	"DateInput":            {Replacement: "components.CalendarPicker", Reason: "Leading-icon date trigger opening Onevra Calendar"},
	"ColorPicker":          {Replacement: "components.ColorView", Reason: "Onevra RGBA color editing view"},
	"ColorWell":            {Replacement: "components.ColorPicker", Reason: "Customizable swatch opening Onevra ColorView"},
	"Tooltip":              {Replacement: "components.Tooltip", Reason: "Onevra text hints on MyGo tooltip behavior"},
	"Toast":                {Replacement: "components.Toasts", Reason: "Onevra toast look with status icons and a leading slot on MyGo's toast queue; Shell builds it automatically"},
	"Popover":              {Replacement: "components.Popover", Reason: "Onevra anchored non-modal panel on MyGo popover behavior"},
	"Modal":                {Replacement: "components.Dialog", Reason: "Onevra composable modal layers"},
	"AlertDialog":          {Replacement: "components.AlertDialog", Reason: "Onevra result-bearing dialog preset"},
	"Divider":              {Reason: "A plain theme-colored separator; compose a Box or container borders instead of a dedicated Onevra component"},
	"Badge":                {Replacement: "components.Badge", Reason: "Onevra compact status label"},
	"RichText":             {Replacement: "components.RichText", Reason: "Onevra inline text composition on the MyGo text primitive"},
	"Breadcrumbs":          {Replacement: "components.Breadcrumbs", Reason: "Onevra path navigation with application-owned activation"},
	"Rating":               {Replacement: "components.Rating", Reason: "Onevra star rating with matching selected fill and stroke"},
	"Split":                {Replacement: "components.Split", Reason: "Unified split panes with a neutral, hover-free divider"},
	"SplitVertical":        {Replacement: "components.Split", Reason: "Vertical Split orientation"},
	"SplitPercent":         {Replacement: "components.SplitPercent", Reason: "Proportional split panes sharing Split's neutral divider and orientation option"},
	"SplitVerticalPercent": {Replacement: "components.SplitPercent", Reason: "Vertical SplitPercent orientation"},
	"Tree":                 {Replacement: "components.Tree", Reason: "Onevra tree with neutral selection, no hover face and double-click disclosure"},
	"TreeItem":             {Replacement: "components.Tree", Reason: "Onevra TreeItem rows"},
	"Outline":              {Replacement: "components.TreeVirtualized", Reason: "Onevra virtualized tree with neutral selection and double-click disclosure"},
	"Progress":             {Replacement: "components.Progress", Reason: "Onevra determinate and indeterminate progress bar"},
	"Spinner":              {Replacement: "components.Spinner", Reason: "Onevra ring and spoke activity indicators"},
	"Avatar":               {Replacement: "components.Avatar", Reason: "Onevra photo or initials avatar with presence, sizes and an overlapping AvatarGroup"},
	"Meter":                {Reason: "No Onevra Meter component is needed"},
	"BackButton":           {Reason: "No Onevra BackButton component is needed"},
	"ForwardButton":        {Reason: "No Onevra ForwardButton component is needed"},
	"Form":                 {Reason: "No Onevra Form component is needed"},
	"Field":                {Reason: "No Onevra Field component is needed"},
	"Fieldset":             {Reason: "No Onevra Fieldset component is needed"},
	"FindBar":              {Reason: "No Onevra FindBar component is needed"},
	"Toolbar":              {Reason: "No Onevra Toolbar component is needed"},
	"Table":                {Replacement: "components.TableVirtualized", Reason: "Onevra ordinary or virtualized table with neutral selection, no hover face and application-owned sorting"},
	"OutlineTable":         {Reason: "No Onevra OutlineTable component is needed"},
	"TokenField":           {Replacement: "components.TagInput", Reason: "Onevra pill tags with a Backspace pick, gliding motion and duplicate feedback; applications own suggestions"},
	"Scroll":               {Replacement: "components.ScrollView", Reason: "Unified scroll container with vertical scrolling by default"},
	"ScrollHorizontal":     {Replacement: "components.ScrollView", Reason: "Horizontal ScrollView axes"},
	"ScrollBoth":           {Replacement: "components.ScrollView", Reason: "Both ScrollView axes"},
	"List":                 {Replacement: "components.List", Reason: "Onevra virtual list with neutral selection and no hover face"},
	"GridView":             {Replacement: "components.GridView", Reason: "Onevra virtual grid with neutral selection and content-area sizing"},
	"Sidebar":              {Replacement: "components.Sidebar", Reason: "Onevra grouped navigation with application-owned selection"},
	"SidebarSection":       {Replacement: "components.Sidebar", Reason: "Onevra sidebar section headings"},
	"SidebarItem":          {Replacement: "components.Sidebar", Reason: "Onevra sidebar navigation items"},

	"Segmented":       {Replacement: "components.Tabs", Reason: "Onevra Tabs with TabsSegmented, a macOS segmented style"},
	"CheckboxGroup":   {Reason: "No separate Onevra checkbox group component is needed"},
	"RadioGroup":      {Reason: "No separate Onevra radio group component is needed; upstream grouping remains a composition building block"},
	"Toggle":          {Replacement: "components.ToggleButton", Reason: "Onevra toggle button styling"},
	"ToggleGroup":     {Reason: "Grouped toggle styling is not needed; compose independent ToggleButtons with layout primitives"},
	"Checkbox":        {Replacement: "components.Checkbox", Reason: "Onevra checkbox styling"},
	"Switch":          {Replacement: "components.Switch", Reason: "Onevra switch styling"},
	"Radio":           {Replacement: "components.Radio", Reason: "Onevra radio styling"},
	"NumberInput":     {Replacement: "components.NumberInput", Reason: "Onevra numeric text field with integrated vertical step buttons"},
	"Stepper":         {Reason: "Step buttons are integrated into NumberInput; components.Stepper is a flow progress indicator, not a numeric stepper"},
	"TextInputBuffer": {Replacement: "components.TextBoxBuffer", Reason: "Onevra single-line input with application-owned indexed storage"},
	"TextAreaBuffer":  {Replacement: "components.TextBoxBuffer", Reason: "Onevra multiline input with application-owned indexed storage"},
	"TextInput":       {Replacement: "components.TextBox", Reason: "Onevra single-line text input"},
	"TextArea":        {Replacement: "components.TextBox", Reason: "Onevra multiline text input"},
	"SearchField":     {Replacement: "components.TextBox", Reason: "TextBox with a leading search icon; applications own search and clearing behavior"},
	"Select":          {Replacement: "components.Select", Reason: "Onevra select parts with application-owned options"},
	"Combobox":        {Replacement: "components.Select", Reason: "Searchable Select with application-owned filtering"},
	"Autocomplete":    {Replacement: "components.Select", Reason: "Searchable Select with application-owned suggestions"},
	"Slider":          {Replacement: "components.Slider", Reason: "Onevra continuous slider"},
	"StepSlider":      {Replacement: "components.Slider", Reason: "Slider Step property enables snapping and tick marks"},
	"RangeSlider":     {Replacement: "components.Slider", Reason: "Slider Range property enables two independently focusable thumbs"},
	"Accordion":       {Replacement: "components.Accordion", Reason: "Onevra disclosure styling without hover"},
	"AccordionItem":   {Replacement: "components.Accordion", Reason: "AccordionGroup.Item builds Onevra sections"},
	"Collapsible":     {Replacement: "components.Collapsible", Reason: "Onevra disclosure with configurable adjacent arrow"},
	"Tabs":            {Replacement: "components.Tabs", Reason: "Onevra tab styling"},
	"Text":            {Replacement: "components.Text", Reason: "Onevra text defaults"},
	"EditableText":    {Replacement: "components.Text", Reason: "Text with an Editable binding; undecorated inline editing"},
	"Icon":            {Replacement: "components.Icon", Reason: "Onevra icon component with named asset accessors"},
	"Image":           {Replacement: "components.Image", Reason: "Onevra image component with application-owned decoded assets"},

	"Button":        {Replacement: "components.Button", Reason: "Onevra button styles"},
	"MenuButton":    {Replacement: "components.Button / Menu", Reason: "Compose components.Button(...).Menu(...) as demonstrated under Menus; no separate component"},
	"PrimaryButton": {Replacement: "components.Button / ButtonPrimary", Reason: "A button variant, not a separate component"},
	"Link":          {Replacement: "components.Button / ButtonLink", Reason: "An inline button variant"},

	"Installed apps": {Reason: "Demonstrated by the launcher example under source/examples/launcher; the gallery hosts no launcher page"},
}

// These primitives remain composition building blocks, without standalone examples.
// Keep this list explicit so new APIs still prompt review. Base constructors are
// skipped by the scanner itself and are not inventoried as styled components.
var mygoPrimitives = map[string]bool{
	"Box": true, "Column": true, "Row": true, "Spacer": true,
	"Grid": true, "Overlay": true,
	// Formatted display text needs no Onevra behavior or styling.
	"Textf": true,
	// Focus inspection stays with MyGo alongside FocusBind; no styled wrapper.
	"FocusedValue": true,
}

// Native hosts and rendering primitives keep demonstrations of direct use.
// Subpackages and element methods supplement the free-function scanner here.
var mygoDirectExamples = map[string]string{
	"Transforms":         "Use MyGo element transforms directly so drawing, input, accessibility and overlays share coordinates",
	"Motion":             "Use MyGo animation methods directly for scheduling, interruption and reduced motion; no Onevra animation runtime",
	"Flyout":             "Use MyGo parent-owned native popup windows directly; applications own content and lifetime",
	"WebView":            "Use MyGo web views, content rules and Cookies directly; applications own their pages, rules and windows",
	"System information": "Use MyGo Network, Secrets, FontFamilies, desktop preferences and locales directly; applications own their reactions, choices and storage",
	"Recent documents":   "Use MyGo recent documents and the jump list directly; applications own their files",
	"File watching":      "Use MyGo WatchFiles directly; applications own the watched paths and their lifetime",
	"Global shortcuts":   "Use MyGo GlobalShortcut directly; applications own the accelerators, their conflicts and what a press does",
	"Native dialogs":     "Use MyGo Dialog directly for system file and message dialogs; applications own the paths, answers and when to ask",
	"Audio":              "Use MyGo audio directly for decoding, mixing and device output; applications own sounds and player lifetime",
	"Gradients & blur":   "Use MyGo element gradients, gradient stops, conic gradients, gradient borders and blur directly; no Onevra styling layer",
	"Strokes":            "Use MyGo path caps, joins, miter limits and dashes directly; no Onevra drawing layer",
	"Hidden elements":    "Use MyGo's Element.Hidden directly to remove an element from layout while keeping its state",
	"Zoom":               "Use MyGo's Context.SetZoom directly; applications own the zoom level and its shortcuts",
	"Shaders":            "Use MyGo's ui/shader directly for fragment shaders, shader fills, element effects and backdrops; no Onevra drawing layer",
	"Live chart":         "Use ui/chart's MinMax and LineInto with Painter.StrokePolyline directly for streaming continuous series; Onevra charts cover categorical views",
	"Canvas":             "Use MyGo painters, offscreen canvases, bitmap regions and pixelated bitmaps directly; no Onevra drawing layer",
	"Game loop":          "Use the MyGo game plugin directly for fixed-step loops, cameras and gamepads; no Onevra game runtime",
	"3D renderer":        "Use MyGo bitmaps and the game loop directly for a CPU software rasterizer; no Onevra 3D layer",
	"Code editor":        "Use MyGo's ui/editor directly for editable code views; its default look follows the frame theme and every part is an application-replaceable slot",
	"Diff viewer":        "Use MyGo's ui/editor diff viewer directly for code review",
	"Rich text editor":   "Use MyGo's ui/richtext directly for attributed text editing; Onevra RichText composes read-only inline text, not an editor",
	"Markdown":           "Use MyGo's ui/markdown directly for documents; applications own images, links and custom code blocks through its options",
	"Markdown stream":    "Use MyGo's ui/markdown directly for streaming replies",
}

// Element methods are not found by the free-function scanner. These native menu
// APIs are deliberately used directly, including by the Onevra gallery examples.
var mygoNativeMenus = map[string]string{
	"ContextMenu": "ui/menu.go",
	"Menu":        "ui/menu.go",
}

// Reviewed subpackages and application examples outside the top-level
// styled-widget scanner. The value names where the API lives or is demonstrated.
var mygoSubpackages = map[string]string{
	"Charts": "ui/chart", "Code editor": "ui/editor", "Diff viewer": "ui/editor",
	"Rich text editor": "ui/richtext", "Markdown": "ui/markdown", "Markdown stream": "ui/markdown",
	"Shaders": "ui/shader", "Live chart": "ui/chart",
	"Installed apps": "examples/launcher",
}

type mygoComponent struct{ Name, Source string }

type mygoStatus string

const (
	mygoReplaced   mygoStatus = "Replaced by Onevra"
	mygoExcluded   mygoStatus = "Not needed"
	mygoUpstream   mygoStatus = "Still using MyGo"
	mygoUnreviewed mygoStatus = "Awaiting review"
	mygoPrimitive  mygoStatus = "Use MyGo directly"
)

type mygoReviewEntry struct {
	mygoComponent
	Status                               mygoStatus
	Replacement, ReplacementName, Reason string
	HasExample                           bool
}

// The catalog supplements discovery with method-based UI and plugins. An API
// discovered in source is not proof of application usage: only catalog entries
// are labeled as upstream examples; others explicitly await review.
func mygoReviewEntries(discovered []mygoComponent, examples []mygoExample, decisions map[string]mygoDecision) []mygoReviewEntry {
	found := make(map[string]mygoComponent, len(discovered))
	for _, entry := range discovered {
		found[entry.Name] = entry
	}
	for name, source := range mygoNativeMenus {
		found[name] = mygoComponent{name, source}
	}
	demonstrated := map[string]bool{}
	plugins := map[string]bool{}
	for _, entry := range examples {
		demonstrated[entry.name] = true
		plugins[entry.name] = entry.group == "Plugins"
		if _, ok := found[entry.name]; !ok {
			source := mygoSubpackages[entry.name]
			if source == "" {
				source = "gallery catalog"
			}
			found[entry.name] = mygoComponent{entry.name, source}
		}
	}
	for name := range decisions {
		if _, ok := found[name]; !ok {
			source := mygoSubpackages[name]
			if source == "" {
				source = "Not found in current inventory"
			}
			found[name] = mygoComponent{name, source}
		}
	}
	entries := make([]mygoReviewEntry, 0, len(found))
	for name, component := range found {
		entry := mygoReviewEntry{mygoComponent: component, Status: mygoUnreviewed, HasExample: demonstrated[name]}
		if entry.HasExample {
			entry.Status = mygoUpstream
		}
		if mygoPrimitives[name] {
			entry.Status, entry.HasExample = mygoPrimitive, false
		}
		if _, native := mygoNativeMenus[name]; native {
			entry.Status, entry.HasExample = mygoPrimitive, false
			entry.Reason = "Use MyGo native menus directly; no Onevra menu implementation"
		}
		if plugins[name] {
			entry.Status = mygoPrimitive
			entry.Reason = "Use MyGo plugins directly; no Onevra plugin implementation"
		}
		if reason, direct := mygoDirectExamples[name]; direct {
			entry.Status, entry.Reason = mygoPrimitive, reason
		}
		if decision, handled := decisions[name]; handled {
			entry.Status, entry.Replacement, entry.Reason, entry.HasExample = mygoExcluded, decision.Replacement, decision.Reason, false
			entry.ReplacementName = decision.ReplacementName
			if decision.Replacement != "" {
				entry.Status = mygoReplaced
			}
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries
}

func pendingMyGo(discovered []mygoComponent, examples []mygoExample) []mygoComponent {
	var pending []mygoComponent
	for _, entry := range mygoReviewEntries(discovered, examples, mygoHandled) {
		if entry.Status == mygoUpstream || entry.Status == mygoUnreviewed {
			pending = append(pending, entry.mygoComponent)
		}
	}
	return pending
}

// LogMyGoReview reports outstanding decisions once at application startup, never
// during rendering. Pending decisions are reminders, not build or runtime errors.
func LogMyGoReview(printf func(string, ...any)) {
	pending := pendingMyGo(mygoDiscovered, mygoExamples)
	printf("[MyGo review] %s: %d pending; record decisions in mygo_review.go", mygoScanVersion, len(pending))
	for _, entry := range mygoReviewEntries(mygoDiscovered, mygoExamples, mygoHandled) {
		if entry.Status == mygoUpstream || entry.Status == mygoUnreviewed {
			detail := entry.Source
			if !entry.HasExample {
				detail += "; no gallery example"
			}
			printf("[MyGo review] pending: %s (%s)", entry.Name, detail)
		}
		if entry.Source == "Not found in current inventory" {
			printf("[MyGo review] decision no longer matches inventory: %s", entry.Name)
		}
	}
}

func unhandledMyGoExamples(examples []mygoExample) []mygoExample {
	result := make([]mygoExample, 0, len(examples))
	for _, entry := range examples {
		_, native := mygoNativeMenus[entry.name]
		if _, handled := mygoHandled[entry.name]; !handled && !mygoPrimitives[entry.name] && !native {
			result = append(result, entry)
		}
	}
	return result
}
