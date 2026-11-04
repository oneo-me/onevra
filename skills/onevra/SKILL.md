---
name: onevra
description: Build, test and package native Go desktop applications with Onevra components on MyGo. Use when writing or reviewing UI that imports github.com/oneo-me/onevra/source, choosing between Onevra and MyGo controls, writing headless or hidden-window UI tests, or setting up mygo.json packaging and the source/ project layout. Excludes changes to MyGo itself.
---

# Onevra

Onevra (`github.com/oneo-me/onevra/source`) is a product-neutral component library built on MyGo. MyGo owns rendering, layout, input, windows and plugins; Onevra supplies themed controls, window shells, Lucide icons and complex components; applications package with MyGo's CLI. The application owns business state, localized text, persistence and asynchronous work.

Before using an API, read its public signature in the version-matched module source rather than guessing:

```sh
go list -m -f '{{.Dir}}' github.com/oneo-me/onevra/source   # Onevra source and examples
go list -m -f '{{.Dir}}' github.com/egoist/mygo              # MyGo docs/ and ui/
```

Run these inside the application module (`source/`); add `GOWORK=off` there when checking that the module builds independently. The Onevra module directory also contains runnable starter applications under `examples/desktop`, `examples/menubar`, `examples/launcher`, `examples/documents` and `examples/games/lawn`; copy their structure rather than inventing a lifecycle.

## Basic usage

```sh
go get github.com/oneo-me/onevra/source@latest
```

```go
func main() {
	mygo.App.WhenReady(func() {
		shell := plain.NewShell(plain.ShellOptions{
			Title: "Hello",
			Theme: theme.Onevra,
			Content: func(c *ui.Context) {
				ui.Column(c).Fill().Padding(24).Gap(12).Children(func() {
					components.Text(c, "Hello, Onevra!")
					if components.Button(c, "Save", components.ButtonPrimary).Clicked() {
						// Act on application state.
					}
					components.Icon(c, lucide.Search).Size(20, 20)
				})
			},
		})
		shell.Window().OnClosed(func() { mygo.App.Quit() })
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```

- Create windows only after `mygo.App.WhenReady`. Choose `shells/plain` for a single content area, which shows `Title` in its header, or `shells/sidebar` for sidebar navigation; use `shell.Window()` for native events instead of adding a second lifecycle. Pass `Hidden: true` when lifecycle handlers or content must be prepared before showing. Put a view switcher, such as `components.Tabs`, in `plain.ShellOptions.HeaderLeading` and window-level actions, such as a Settings button, in `plain.ShellOptions.HeaderActions`; they sit at the left and right of the header.
- Give a new project's main window an 820 × 740 initial and minimum size. `sidebar.NewShell` defaults to it; on a `plain` main window set `Width: 820, Height: 740, MinWidth: 820, MinHeight: 740`. Size secondary windows, such as settings or welcome, to their content.
- Builders run synchronously on the UI thread every frame. Keep state in retained application values passed by pointer (`components.Checkbox(c, &agreed, "I agree")`); never start background work, I/O or parsing inside a builder. Do background work elsewhere and apply results on the main thread, then request a frame.
- Element handles are valid only for the current build pass. Bind a `ui.Handle` for identity or geometry across frames.
- The application owns the theme. Choose light/dark/system with `mygo.Theme.SetSource` (before `App.Run` for a saved preference). Pass `Theme: theme.Onevra` (or a function returning your own `*ui.Theme`) in Shell options; without a Shell, call `theme.Apply(c)` or `c.SetTheme` at the top of the root builder. Components never call `Context.SetTheme`.
- Keep application messages in MyGo catalogs (`i18n` in `mygo.json`, typed functions from `mygo generate`) and pass translated labels explicitly. Component-owned labels follow the window's language from Onevra's catalog; change them with `onevra.<key>` messages. Keep keys and stored values independent of translated text.
- Complex views each have their own package in `components/chart`, `components/graph`, `components/treemap`, `components/flow`, `components/command` and `components/chat`. Keep their documents and results in retained state, prepared outside builders.

## Choosing controls

Use Onevra's control whenever one exists; it replaces MyGo's styled widget of the same purpose. Check `components/` (including its subpackages) and `shells/` in the Onevra module before choosing a control.

| Need | Use |
| --- | --- |
| Button, icon button, link, primary button, menu button | `components.Button` with a variant; `components.IconButton` for icon-only actions with a tooltip, plain, compact or raised; `ButtonToolbar` matches `TabsSegmented`'s chosen segment; `components.MenuButton` and `components.ContextMenu` for menus in the window's look, with `Group` headings, or `.Menu(...)` and `.ContextMenu(...)` for the system's; `components.HoldToConfirm` for destructive actions confirmed by holding |
| Action toolbar, floating editor or selection toolbar, tools rail | `components.FloatingButtonGroup`, muted or raised, in a row or column |
| Title bar, toolbar or bottom bar floating over scrolling content | `components.ScrollEdgeBar` with a glass bar, soft or hard scroll edge, or progressive blur; pad the content by `ScrollEdgeBarExtent`; `ScrollEdge` for the effect alone |
| Text, editable text, rich inline text | `components.Text`, `components.RichText` |
| Text input, text area, search field | `components.TextBox`, `components.TextBoxBuffer` |
| Tags or other short free-form values | `components.TagInput` instead of `ui.TokenField` |
| Checkbox, switch, radio, toggle | `components.Checkbox`, `Switch`, `Radio`, `ToggleButton` |
| Select, combobox, autocomplete | `components.Select` (searchable), `components.NativeSelect` |
| Number, slider, date, time, color | `NumberInput`, `Slider`, `Calendar`/`CalendarPicker`, `TimePicker`, `ColorPicker`/`ColorView` |
| Scroll, list, grid, tree, outline, table | `ScrollView`, `List`, `GridView`, `Tree`, `TreeVirtualized`, `Table`, `TableVirtualized` |
| Tabs, accordion, collapsible, split, breadcrumbs, sidebar | The Onevra component of the same name; `components.Tabs` takes `TabsUnderline`, `TabsSegmented` or `TabsHighlight`, `TabsWithLeading` adds icons before labels and `TabsWith` also adds content after them and context menus |
| Circular item picker in a magnified honeycomb | `components.Honeycomb` |
| Slides browsed by dragging, swiping or arrows, with optional rotation | `components.Carousel` with an application-owned `CarouselState` |
| One-by-one triage of a deck, swiped left or right with undo | `components.CardStack` with an application-owned `CardStackState` |
| Modal, alert, popover, tooltip | `Dialog`, `AlertDialog`, `Popover`, `Tooltip` |
| Toast | MyGo's `c.Toast`, `ToastAction`, `AddToast`; Shell builds them with `components.Toasts`, which standalone hosts call themselves; place them with `ShellOptions.Toasts` (`components.ToastOptions`). `ui.Toast.Type` takes `components.ToastInfo`/`Success`/`Warning`/`Error` |
| Progress, spinner, badge, rating, icon, image | The Onevra component of the same name; icons from `icons/lucide`; `components.FileIcon` for a file's type in file trees |
| Progress through a multi-step flow | `components.Stepper`; MyGo's `ui.Stepper` is a numeric stepper, which `NumberInput` replaces |
| Avatar or avatar group | `components.Avatar`, `components.AvatarGroup` |
| Categorical charts, donuts, bubbles and part-to-whole bars | `components/chart`: `Bar`, `Line`, `Area`, `StackedBar`, `Donut`, `Bubble`, `SplitBar`; see the [usage example](../../source/components/chart/example_test.go) |
| Relationship graph | `components/graph`: `View`; see the [usage example](../../source/components/graph/example_test.go) |
| Treemap of nested sizes, such as disk usage | `components/treemap`: `View`; see the [usage example](../../source/components/treemap/example_test.go) |
| Flow diagram with automatic layout | `components/flow`: `View` and `Card`; see the [usage example](../../source/components/flow/example_test.go) |
| Pan and zoom on a custom canvas | `components/gesture`: `Navigation.Handle` for wheel and pinch zoom and touchpad panning, or `Classify` inside an existing `HandleInput` |
| Settings surface | `components.Settings` |
| Conversation transcript and composer | `components/chat`: `Turn`, `FloatingConversation`, `GlassComposer`; fill composer slots with `Chip`, `Picker`, `EffortPicker`, `Attachment` (also for images in mixed rows), `ContextMeter`, `QuotaSummary` and a compact `components.Switch`; stream replies with MyGo's `markdown.Stream` |

Let a scroll view reach the edges of the surface it scrolls within and balance the visuals with content padding instead of an outer margin: an inset view clips scrollbars, scroll gestures and edge hit targets before they reach the edge. `ScrollbarInsets` moves the track independently of content padding.

Use MyGo directly for:

- Layout primitives: `ui.Box`, `Column`, `Row`, `Spacer`, `Grid`, `Overlay`, and `ui.Textf` for formatted display text.
- Specialized charts and continuous/time axes from `ui/chart`, and element transforms, motion methods and timelines. Use `LayoutBounds` for layout calculations and local pointer coordinates; `Element.Bounds` includes visual transforms.
- Native menus, such as the menu bar's, tray and Dock menus, `Flyout`, `WebView`, `ui/editor` code, snippet and diff views, the `ui/richtext` editor, `ui/markdown` documents and streaming replies, and every MyGo plugin, including Glass. Do not write Onevra-style replacements.
- MyGo's unstyled bases when building a new reusable control.
- A MyGo widget with no Onevra equivalent. Onevra intentionally omits Divider, Meter, Form, Field, Toolbar and OutlineTable: compose dividers from Box or borders before reaching for the upstream widget. For a segmented control, use `components.Tabs` with `TabsSegmented` instead of `ui.Segmented`.

Charts accept application-owned data and formatting, with generic empty and detail slots. Keep series/slice IDs stable and state separate per chart; `chart.State{Index: -1}` starts unselected. Prepare large datasets outside builders; these views do not downsample or virtualize data. Graphs and flow diagrams keep layout and camera in one retained `graph.State` or `flow.State` per view and need stable node IDs.

The authoritative replacement and exclusion map, with reasons, is `examples/gallery/internal/gallery/mygo_review.go` in the Onevra module.

## Tests

Write tests that never show a window by default. Read [references/testing.md](references/testing.md) before adding UI or native tests.

## Project layout and packaging

Package applications with MyGo's CLI and `mygo.json`. Read [references/packaging.md](references/packaging.md) before creating a project, changing its layout or packaging configuration.

## Checks

From the Go module directory:

```sh
gofmt -l .
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go build ./...
```

`gofmt -l .` must print nothing. While iterating, run only the affected packages with `-run`; run the full list before handing off, and leave slow runs such as a full race check to the user. Headless tests and successful cross-builds do not prove native behavior: review UI changes in a desktop session on each claimed platform, covering both themes, narrow windows, keyboard navigation, input methods and accessibility, and report what remains unreviewed.
