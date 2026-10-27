package gallery

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/egoist/mygo/ui/richtext"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// demoState is owned by one page, survives navigation, and is never persisted.
type demoState struct {
	variant         string
	flyout          flyoutDemo
	glass           glassDemo
	terminal        terminalDemo
	webView         webViewDemo
	audio           audioDemo
	system          systemDemo
	canvas          canvasDemo
	graphics        graphicsDemo
	game            gameDemo
	scene3d         scene3D
	code            *editor.State
	diff            *editor.DiffState
	richText        *richtext.Document
	markdown        markdown.Options
	text            string
	query, queried  string
	buffer          *ui.TextBuffer
	readOnly        bool
	choice          string
	on, other, open bool
	index, count    int
	value, high     float64
	size            float32
	offset          ui.Point
	seek            float32
	tags            []string
	day             time.Time
	color           ui.Color
	list            ui.ListState
	grid            ui.GridState
	outline         ui.OutlineState[string]
}

func newDemoState() *demoState {
	return &demoState{text: "Hello MyGo", choice: "Small", value: 40, high: 80,
		size: 140, day: time.Date(2026, 6, 15, 9, 30, 0, 0, time.Local),
		color: ui.RGB(80, 120, 220), tags: []string{"Go"}}
}

func exampleTextInputBuffer(c *ui.Context, s *demoState) {
	ui.TextInputBuffer(c, s.buffer).Label(tr(c, "name")).Placeholder(tr(c, "enterAName")).ReadOnly(s.readOnly).FillWidth()
}

func exampleTextAreaBuffer(c *ui.Context, s *demoState) {
	ui.TextAreaBuffer(c, s.buffer).Label(tr(c, "notes")).Placeholder(tr(c, "writeSomeNotes")).ReadOnly(s.readOnly).Height(140).FillWidth()
}

func exampleCheckboxGroup(c *ui.Context, s *demoState) {
	ui.CheckboxGroup(c, tr(c, "notifications"), func() {
		components.Checkbox(c, &s.on, tr(c, "email"))
		components.Checkbox(c, &s.other, tr(c, "sound"))
	})
}

func exampleStepSlider(c *ui.Context, s *demoState) {
	ui.StepSlider(c, &s.value, 0, 100, 20).Label(tr(c, "quality"))
	ui.Text(c, msg(c).Quality0f(s.value))
}

func exampleRangeSlider(c *ui.Context, s *demoState) {
	ui.RangeSlider(c, &s.value, &s.high, 0, 100, 5).Label(tr(c, "range"))
	ui.Text(c, msg(c).ValueRange(s.value, s.high))
}

func exampleStepper(c *ui.Context, s *demoState) {
	ui.Stepper(c, &s.value, 0, 100, 1).Label(tr(c, "copies"))
	ui.Text(c, msg(c).Copies0f(s.value))
}

func exampleRating(c *ui.Context, s *demoState) {
	ui.Rating(c, &s.index, 5).Label(tr(c, "rating"))
	ui.Text(c, msg(c).StarsOutOfFive(s.index))
}

func exampleNumberInput(c *ui.Context, s *demoState) {
	ui.NumberInput(c, &s.value, 0, 100, 1).Label(tr(c, "quantity"))
	ui.Text(c, msg(c).Quantity0f(s.value))
}

func exampleTokenField(c *ui.Context, s *demoState) {
	ui.TokenField(c, &s.tags, []string{"Go", "UI", "Native", "Desktop"}).Label(tr(c, "tags")).FillWidth()
	components.Text(c, strings.Join(s.tags, ", ")).Wrap()
}

func exampleEditableText(c *ui.Context, s *demoState) {
	ui.EditableText(c, &s.text).Label(tr(c, "editableName"))
}

func exampleCalendar(c *ui.Context, s *demoState) {
	ui.Calendar(c, &s.day).Label(tr(c, "day"))
	components.Text(c, s.day.Format("2006-01-02"))
}

func exampleDateInput(c *ui.Context, s *demoState) {
	ui.DateInput(c, &s.day).Label(tr(c, "date"))
	components.Text(c, s.day.Format("2006-01-02"))
}

func exampleTimeInput(c *ui.Context, s *demoState) {
	ui.TimeInput(c, &s.day).Label(tr(c, "time"))
	components.Text(c, s.day.Format("15:04"))
}

func exampleColorWell(c *ui.Context, s *demoState) {
	ui.ColorWell(c, &s.color).Label(tr(c, "tint"))
	ui.Box(c).Size(80, 40).Background(s.color)
}

func exampleColorPicker(c *ui.Context, s *demoState) {
	ui.ColorPicker(c, &s.color).Label(tr(c, "color"))
	ui.Box(c).Size(80, 40).Background(s.color)
}

func exampleList(c *ui.Context, s *demoState) {
	s.list.Selected = &s.index
	s.list.Label = func(i int) string { return msg(c).ItemD(i + 1) }
	ui.List(c, &s.list, 1000, func(i int) { ui.Text(c, msg(c).ItemD(i+1)).Padding(8) }).Height(200).FillWidth().Label(tr(c, "items"))
	ui.Text(c, msg(c).SelectedItemD(s.index+1))
}

func exampleTable(c *ui.Context, s *demoState) {
	s.list.Selected = &s.index
	cols := []ui.TableColumn{{Title: tr(c, "name")}, {Title: tr(c, "size"), Width: 90}}
	ui.Table(c, &s.list, cols, 100, func(row, col int) {
		if col == 0 {
			ui.Text(c, msg(c).FileD(row+1))
		} else {
			ui.Text(c, msg(c).SizeKb(row+1))
		}
	}).Height(200).FillWidth().Label(tr(c, "files"))
}

func exampleTree(c *ui.Context, s *demoState) {
	ui.Tree(c, func() {
		ui.TreeItem(c, "source", &s.open, func() {
			if ui.TreeItem(c, "main.go", nil, nil).Selected(s.on).Clicked() {
				s.on = true
			}
		})
	}).Label(tr(c, "project"))
}

func exampleOutline(c *ui.Context, s *demoState) {
	s.outline.List.Selected = &s.index
	ui.Outline(c, &s.outline, []string{"source", "docs"}, func(name string) []string {
		if name == "source" {
			return []string{"main.go", "go.mod"}
		}
		if name == "docs" {
			return []string{"README.md"}
		}
		return nil
	}, func(name string) { components.Text(c, name) }).Height(180).FillWidth().Label(tr(c, "projectOutline"))
}

func exampleOutlineTable(c *ui.Context, s *demoState) {
	s.outline.List.Selected = &s.index
	cols := []ui.TableColumn{{Title: tr(c, "name")}, {Title: tr(c, "kind"), Width: 90}}
	ui.OutlineTable(c, &s.outline, cols, []string{"source"}, func(name string) []string {
		if name == "source" {
			return []string{"main.go", "go.mod"}
		}
		return nil
	}, func(name string, col int) {
		if col == 0 {
			components.Text(c, name)
		} else {
			components.Text(c, tr(c, "file"))
		}
	}).Height(180).FillWidth().Label(tr(c, "projectTable"))
}

func exampleGridView(c *ui.Context, s *demoState) {
	s.grid.Selected = &s.index
	s.grid.Label = func(i int) string { return msg(c).ItemD(i + 1) }
	ui.GridView(c, &s.grid, 1000, 90, 70, func(i int) {
		ui.Text(c, msg(c).ItemD(i+1)).Padding(12)
	}).Height(220).FillWidth().Label(tr(c, "items"))
	ui.Text(c, msg(c).SelectedItemD(s.index+1))
}

func exampleScroll(c *ui.Context, s *demoState) {
	ui.Scroll(c).Height(180).FillWidth().Children(func() {
		for i := range 30 {
			ui.Text(c, msg(c).LineD(i+1)).Padding(4)
		}
	})
}

func exampleScrollHorizontal(c *ui.Context, s *demoState) {
	ui.ScrollHorizontal(c).Height(80).FillWidth().Children(func() {
		ui.Row(c).Gap(12).Children(func() {
			for i := range 20 {
				ui.Text(c, msg(c).CardD(i+1)).Width(120).Padding(16).Background(c.Theme().Surface)
			}
		})
	})
}

func exampleScrollBoth(c *ui.Context, s *demoState) {
	ui.ScrollBoth(c).Height(180).FillWidth().Children(func() {
		ui.Column(c).Width(800).Children(func() {
			for i := range 30 {
				ui.Text(c, msg(c).RowDNameStatusUpdated(i+1)).Padding(8)
			}
		})
	})
}

func exampleGrid(c *ui.Context, s *demoState) {
	ui.Grid(c).Columns(2).Gap(8).FillWidth().Children(func() {
		for i := range 6 {
			ui.Text(c, msg(c).CellD(i+1)).Padding(12).Background(c.Theme().Surface)
		}
	})
}

func exampleSplit(c *ui.Context, s *demoState) {
	ui.Split(c, &s.size, func() { components.Text(c, tr(c, "leftPane")).Wrap() }, func() { components.Text(c, tr(c, "rightPane")).Wrap() }).Height(180).FillWidth()
}

func exampleSplitVertical(c *ui.Context, s *demoState) {
	ui.SplitVertical(c, &s.size, func() { components.Text(c, tr(c, "topPane")) }, func() { components.Text(c, tr(c, "bottomPane")) }).Height(240).FillWidth()
}

func exampleCollapsible(c *ui.Context, s *demoState) {
	ui.Collapsible(c, tr(c, "details"), &s.open, func() { components.Checkbox(c, &s.on, tr(c, "notifications")) })
}

func exampleAccordion(c *ui.Context, s *demoState) {
	ui.Accordion(c, func() {
		ui.AccordionItem(c, tr(c, "general"), &s.open, func() { components.Checkbox(c, &s.on, tr(c, "notifications")) })
		ui.AccordionItem(c, tr(c, "about"), &s.other, func() { components.Text(c, tr(c, "builtWithMygo")) })
	}).FillWidth()
}

func exampleSidebar(c *ui.Context, s *demoState) {
	ui.Sidebar(c, &s.choice, func() {
		ui.SidebarSection(c, tr(c, "locations"), nil, func() {
			ui.SidebarItem(c, tr(c, "home"), lucide.House(), "Home")
			ui.SidebarItem(c, tr(c, "files"), lucide.Folder(), "Files")
		})
	}).Height(180).FillWidth()
	components.Text(c, msg(c).SelectedS(tr(c, s.choice)))
}

func exampleBreadcrumbs(c *ui.Context, s *demoState) {
	path := []string{tr(c, "home"), tr(c, "documents"), tr(c, "project")}
	ui.Breadcrumbs(c, path, &s.index).Label(tr(c, "path"))
	components.Text(c, msg(c).ChosenS(path[s.index]))
}

func exampleRadioGroup(c *ui.Context, s *demoState) {
	ui.RadioGroup(c, func() {
		components.Radio(c, &s.choice, "Small", tr(c, "small"))
		components.Radio(c, &s.choice, "Large", tr(c, "large"))
	}).Row().Gap(16).Label(tr(c, "sizeOptions"))
	components.Text(c, tr(c, s.choice))
}

func exampleModal(c *ui.Context, s *demoState) {
	if components.Button(c, tr(c, "openDialog"), components.ButtonSecondary).Clicked() {
		s.open = true
	}
	ui.Modal(c, &s.open, func() {
		components.Text(c, tr(c, "editName")).Bold()
		components.TextBox(c, &s.text, components.TextBoxOptions{}).Label(tr(c, "name"))
		if components.Button(c, tr(c, "done"), components.ButtonSecondary).Clicked() {
			s.open = false
		}
	})
}

func exampleAlertDialog(c *ui.Context, s *demoState) {
	if components.Button(c, tr(c, "askForConfirmation"), components.ButtonSecondary).Clicked() {
		s.open = true
	}
	if ui.AlertDialog(c, &s.open, tr(c, "applyChange"), tr(c, "applyTheseSettings"), tr(c, "cancel"), tr(c, "apply")) == 1 {
		s.count++
	}
	ui.Text(c, msg(c).ConfirmedD(s.count))
}

func exampleTooltip(c *ui.Context, s *demoState) {
	components.Button(c, tr(c, "hoverOrFocusMe"), components.ButtonSecondary).Tooltip(tr(c, "aHelpfulDescriptionOfThisAction"))
}

func exampleToast(c *ui.Context, s *demoState) {
	if components.Button(c, tr(c, "showToast"), components.ButtonSecondary).Clicked() {
		c.Toast(tr(c, "saved"))
	}
	if components.Button(c, tr(c, "showUndoToast"), components.ButtonSecondary).Clicked() {
		s.count++
		c.ToastAction(tr(c, "countIncreased"), tr(c, "undo"), func() { s.count-- })
	}
	ui.Text(c, msg(c).CountD(s.count))
}

func exampleProgress(c *ui.Context, s *demoState) {
	components.Slider(c, &s.value, 0, 100).Label(tr(c, "progressValue"))
	ui.Progress(c, s.value/100).Label(tr(c, "progress"))
	ui.Progress(c, -1).Label(tr(c, "indeterminateProgress"))
}

func exampleSpinner(c *ui.Context, s *demoState) {
	ui.Row(c).Gap(20).Children(func() {
		for _, size := range []float32{16, 24, 32} {
			ui.Spinner(c).Label(tr(c, "loading")).Size(size, size)
		}
	})
}

func exampleBadge(c *ui.Context, s *demoState) {
	ui.Badge(c, tr(c, "new"))
	ui.Badge(c, tr(c, "twelve")).Background(c.Theme().Accent).TextColor(c.Theme().AccentText)
}

func exampleImage(c *ui.Context, s *demoState) {
	ui.Image(c, demoImage).Size(180, 100).Fit(ui.Contain).Label(tr(c, "contain"))
	ui.Image(c, demoImage).Size(120, 120).Fit(ui.Cover).Radius(12).Label(tr(c, "cover"))
	ui.Image(c, demoImage).Size(180, 100).Grayscale().Label(tr(c, "grayscale"))
}

func exampleRichText(c *ui.Context, s *demoState) {
	ui.RichText(c).Children(func() {
		components.Text(c, tr(c, "combine"))
		components.Text(c, tr(c, "boldText")).Bold()
		components.Text(c, tr(c, "with"))
		if components.Button(c, tr(c, "anInlineAction"), components.ButtonLink).Clicked() {
			s.count++
		}
	}).FillWidth()
	ui.Text(c, msg(c).ClicksD(s.count))
}

func exampleBox(c *ui.Context, s *demoState) {
	ui.Box(c).Size(120, 80).Padding(12).Radius(8).Background(c.Theme().Surface).Center().Children(func() { components.Text(c, tr(c, "box")) })
}

// exampleHidden hides the middle box, which gives its room back while what
// is inside it keeps its state.
func exampleHidden(c *ui.Context, s *demoState) {
	components.Checkbox(c, &s.on, tr(c, "boxHideMiddle"))
	ui.Row(c).Gap(8).Wrap().Children(func() {
		for i, key := range []string{"first", "second", "third"} {
			ui.Box(c).Key(key).Size(120, 80).Padding(12).Radius(8).Background(c.Theme().Surface).Center().
				Hidden(i == 1 && s.on).Children(func() { components.Text(c, tr(c, key)) })
		}
	})
}

func exampleRow(c *ui.Context, s *demoState) {
	ui.Row(c).Gap(8).Children(func() { components.Badge(c, tr(c, "first")); components.Badge(c, tr(c, "second")) })
}

func exampleColumn(c *ui.Context, s *demoState) {
	ui.Column(c).Gap(8).Children(func() { components.Text(c, tr(c, "first")); components.Text(c, tr(c, "second")) })
}

func exampleSpacer(c *ui.Context, s *demoState) {
	ui.Row(c).FillWidth().Children(func() {
		components.Text(c, tr(c, "start"))
		ui.Spacer(c)
		components.Text(c, tr(c, "end"))
	})
}

func exampleDivider(c *ui.Context, s *demoState) {
	components.Text(c, tr(c, "above"))
	ui.Divider(c)
	components.Text(c, tr(c, "below"))
}
