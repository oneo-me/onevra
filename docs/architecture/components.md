# Component design

Each shell style lives in its own package under `shells/`, without a shared facade or style selector. The public [sidebar shell package](../../source/shells/sidebar/doc.go) owns a replaceable sidebar-and-content container; the [plain shell package](../../source/shells/plain/shell.go) owns a single-content window that shows its title in the header. Both own native presentation, not application lifetime.

The [components package](../../source/components/doc.go) provides common controls independently of Shell. Sidebar Shell also provides icon buttons for titlebar and toolbar actions. Both packages compose MyGo directly.

Shell reuses the common Icon and IconButton components; components remain independent of Shell.

Common controls expose at most one visual variant axis, such as button style. Each control has one theme-scaled standard size rather than a separate size enumeration; a few add a compact form for dense hosts. Interaction and value states are not visual variants.

## Composition and ownership

Shell creates and configures the MyGo window because transparency, title-bar style and traffic-light placement must be set at creation time. It exposes MyGo's original handle for application events and lifecycle operations, rather than introducing a second event or shutdown system. Both shell styles can start hidden so applications can prepare content and lifecycle handlers before choosing whether to show or focus the window. Applications should not replace the window's content behind Shell.

Applications own business state, localized text, actions and page selection. The reusable [Sidebar](../../source/components/sidebar.go) composes grouped navigation independently of Shell; it supplies item appearance, while applications handle activation and scrolling and Shell owns the window's sidebar geometry. Content customization uses product-neutral slots rather than domain-specific options: navigation can accept application-built leading content while retaining its own selection and activation.

Content builders and Shell state changes run synchronously on the UI thread; builders must not launch repeated background work during rendering. MyGo's value-based Element handles belong to the current build pass, not application state. Use a bound MyGo Handle for identity or geometry queries between builds.

Shell owns only presentation state, does not persist it, and uses MyGo's rendering, layout and input systems. An application-chosen `StateKey` passes through to MyGo, which owns capturing, restoring and storing window geometry; sidebar width, expansion and pages remain application state.

Components read the frame theme and style their own elements; they must not call `Context.SetTheme`, which also resets the window root's appearance. Only [application-level theme installation](#theme-ownership) owns that operation. Collection controls compose their own rows and component-level selection/navigation on MyGo's input and focus APIs.

Virtualized collections reuse non-selecting MyGo list and grid containers for scrolling and layout, rather than copying its layout engine or overriding its styled row theme. Their application-owned state is defined in the [collection API](../../source/components/list.go). Tables share MyGo's public column, sort and layout types and its sideways row-width scrolling; Onevra owns the header's sorting, moving and resizing, while applications sort rows and may persist the layout.

Double-click column fitting and pinned section headers are not provided. Public role, selection, level and active-descendant metadata describe the rows, but this composition does not claim full parity with MyGo's private collection accessibility actions; native assistive-technology review is still required.

For basic usage, see the [README](../../README.md#getting-started); public contracts and defaults belong to the [API source](../../source/shells/sidebar/shell.go).

## Complex component composition

Complex components each live in an independent package under `source/components/<name>/`, with implementation, tests and package documentation kept together. These subpackages may compose common controls; the root `components` package must not depend on them. MyGo plugins continue to be used directly.

### Charts

The [chart component](../../source/components/chart/chart.go) composes categorical bars, lines, areas, stacked bars, donuts, packed bubbles and a [split bar](../../source/components/chart/splitbar.go) from MyGo's scales, paths, marks and selection. Bubble area, not radius, tracks value; packs fill the plot around the largest bubble. Onevra supplies muted colors, grids, gradients, adaptive labels and legends. Selection details show in a tooltip by the pointer or keyboard selection; the split bar keeps its readout. They own no windows or background work.

Applications retain data and state, provide stable series or slice identities, and supply localized labels and formatting. Cartesian and split bar legends hide data; donut and bubble legends select slices. Selection exposes formatted values to assistive technology. Empty and detail slots take application content instead of business-specific variants. See the [runnable usage example](../../source/components/chart/example_test.go) and API comments for data and state contracts.

Numeric updates transition together through MyGo's animation channel, including Cartesian domains and donut boundaries; bubbles first show packed, then spring, and added ones grow in. Retargeting continues from the current picture; Reduced Motion and `DisableAnimations` skip the transition. Split bar edges follow springs; its figures roll with MyGo's numeric content transitions. Data-shape changes appear immediately, and selection details report the latest application values.

These views target bounded categorical datasets, with work proportional to the data; applications own aggregation or sampling. Continuous/time axes, zooming and specialized charts can compose upstream primitives. The component reuses MyGo's Cartesian selection, with a coordinate adaptation for transformed plots in 0.4.3; donut and bubble selection compose MyGo's shortcuts with shape hit testing.

### Treemaps

The [treemap component](../../source/components/treemap/treemap.go) paints squarified, nested tiles on one canvas. Applications own the tree, labels, formatting, shading measure and opening leaves; state keeps the drilled branch. Clicks drill a level and built-in breadcrumbs lead back; drilling lays both levels out at their own size in one frame, and data changes morph tiles. Nodes style themselves and their hover, and the pointer or keyboard raises tooltips and a per-node menu.

### Relationship graphs

The [graph component](../../source/components/graph/graph.go) draws a force-directed graph. Applications supply stable node IDs, groups, sizes, colors and edge styles, and open nodes. Groups seed placement and even hue spacing; palettes, colors, forces, hover depth and empty or overlay slots are replaceable.

Layout, camera, hover and reveal advance in `Draw` on the UI thread. Initial ticks run in per-frame budgets before showing; Reduced Motion and `DisableAnimations` settle at once.

[Gestures](../../source/components/gesture/gesture.go): wheels and pinches zoom, touchpads pan. MyGo reports pinches as `ui.InputZoom` but does not tell smoothed mice from touchpads; a macOS adapter reads scroll phases.

### Flow diagrams

The [flow component](../../source/components/flow/flow.go) lays out directed diagrams in ranks: it reverses cycle-closing edges, ranks nodes to keep edges short, orders ranks to reduce crossings and routes orthogonal edges between ranks. Edges cross ranks anywhere their neighbours leave free and bend only where blocked. Applications build node content, or use `Card`, whose regions are slots, and own selection and opening.

Content is measured on the frame after it is built; the diagram shows once every node has a size. A change of topology or spacing moves nodes and the camera, as do sizes measured while that motion runs; later size changes place the diagram at once. Layout reruns in full on topology or size changes; applications bound diagram size. The camera uses the graph's [gestures](../../source/components/gesture/gesture.go).

### Commands

The [command component](../../source/components/command/command.go) separates a command view from its dialog. The [view](../../source/components/command/view.go) fills any bounded host, such as a panel window, without painting a surface; the [dialog](../../source/components/command/dialog.go) hosts it in Onevra's modal surface and adds dismissal and focus restoration. Applications retain query and active-command state, supply stable command identities and localized metadata, register window shortcuts, and handle activation. Hosts can show results immediately or start with a compact search field that expands after a committed query and collapses when cleared. Filtering uses local metadata, a synchronous host ranking applied to the same frame's query, or host-supplied results; asynchronous search, stale-result rejection, command execution and history remain application-owned. Loading and errors replace actionable results.

Search keeps keyboard focus while arrows move the active result through Onevra's non-selecting List; MyGo owns virtual layout and text editing. When the query or results change, the first available command becomes active and the list returns to the top, unless the host chose the active command in the same frame. The dialog lets the modal scope set initial focus so the overlay keeps its opener for restoration; other hosts request search focus through the state, as after showing a hidden window. Composition-aware dismissal prevents Escape ending an input-method composition from also closing the dialog. Accessibility metadata exposes the active command; native assistive-technology review remains necessary. See the [compilable usage example](../../source/components/command/example_test.go).

On macOS, palettes can select an enabled keyboard input source through the MyGo responder’s [AppKit input context](https://developer.apple.com/documentation/appkit/nstextinputcontext/selectedkeyboardinputsource). This small adapter fills an upstream gap without intercepting events or implementing text input. Selection waits for search focus and committed composition, runs once per opening, and normally restores the previous source on dismissal if the context and source still match.

Hosts may retain the selected source instead. Unavailable sources leave search usable and report an error through retained state. Embedded hosts release the session on removal, all hosts before window destruction; see the [input-source lifecycle API](../../source/components/command/input_source.go).

### Code and Markdown

Onevra has no code or Markdown component. Applications use MyGo's `ui/markdown` for documents and streaming replies, and `ui/editor` for editable code, diffs and read-only snippets (`State.ReadOnly` with `FitContent`). Markdown options own images, link activation and custom code-block views; links open by default unless `OnLink` is set. See the [Gallery examples](gallery.md#retained-examples).

### Conversations

The [chat component](../../source/components/chat/chat.go) composes conversation scrolling, turns, activity groups and a multiline composer whose proportions follow Waku's transcript. Applications own ordered steps with stable identities, stream updates, localized labels and summaries, Markdown content and request lifecycles. A turn shows the prompt, then intermediate prose and activity groups, then an explicit final response; a completed tool never implicitly completes its turn. Builders run synchronously and perform no network or file work.

An activity group opens while it is the live tail of a running turn and closes once it settles, unless the user chose otherwise. Each activity is a compact row with an optional bounded detail viewport, prose or monospace, that keeps its own disclosure and scroll state. Successful turns with a final response fold their steps behind a centered summary row without building hidden content; running turns end with a pulsing working row.

Manual choices survive updates and outer folding; focus inside the steps or reading above the live edge defers automatic folding. Once folded, scrolling back does not reopen a turn. Footers and response actions remain outside the fold.

MyGo owns keyboard disclosure semantics, input methods and scrolling. Enter sends and Shift+Enter inserts a newline: MyGo's multiline editor consumes Enter before shortcuts, so the composer takes plain Enter through the editor's input handler and leaves active compositions alone. While running, an empty draft shows Stop and a nonempty draft shows only Send, leaving the application to decide what a new message does to the run.

Composition does not disable the Send button; Enter still belongs to the input method. The card accepts dropped file paths or host-selected native data formats without preparing them. A paste callback lets hosts attach clipboard content while preserving normal text editing when declined.

File chips and image tiles have optional removal actions; slots hold attachments, accessory controls, non-shrinking trailing controls beside Send or Stop, and a footer row. Hosts can disable sending while preparing attachments, without blocking editing. Applications accept the draft and perform cancellation themselves.

The package supplies chips, pickers, meters, usage limits and a shimmering label styled for these slots. Applications supply the options, measurements and labels, and apply each choice; the controls hold no model, repository, account or permission policy.

Streaming replies use MyGo's `markdown.Stream`, which parses only appended text and reveals it at an even pace; applications switch to `markdown.View` once the reply completes. Applications may show a running reply as prose after the activity groups and pass it as the final response once complete.

The conversation follows the bottom until the user scrolls away, with a jump control to resume. An optional floating layout lets the transcript scroll behind a glass composer while reserving clearance for the latest turn; messages blur progressively from the bottom edge to the middle of the resting composer card, which stays fixed as the card grows. The composer applies MyGo's glass material only to its card, preserving the host's theme and root transparency.

Conversations do not manage the window title bar or floating headers. Hosts place titles, actions and status overlays themselves, for example with `GlassPanel`, and pad or position the conversation to clear them.

Ordinary scrolling preserves document selection, but this version does not virtualize history or preserve a row anchor when history is prepended or arbitrary content above the viewport changes height. Applications bound transcript retention and tool previews, merge incoming updates before UI publication and cache parsing outside builders. Folded turns cost nothing per hidden step; expanded content and retained data still consume proportional work and memory.

The [gallery example](../../source/examples/gallery/internal/gallery/control_chat.go) demonstrates a conversation, the composer controls and a simulated streamed reply without contacting a model.

## Localization

Component and Shell control names are a MyGo library catalog in namespace `onevra` ([l10n](../../source/internal/l10n/en.json)), in English, Simplified and Traditional Chinese, Japanese and Korean. They follow the window's language (`ui.Context.Locale`, set with `mygo.App.SetLocale`); applications change or add translations with `onevra.<key>` messages in their own catalogs. Calendar dates and the first weekday follow the user's regional formats, as MyGo's date controls do.

The catalog is the only source of these names; there is no per-scope or per-Shell override. Explicit control labels belong to the caller. Calendar presentation never changes stored dates, time zones or selection; numeric editors keep canonical numeric syntax. Keep keys and selection values independent of translated text.

## Page container and layout

Plain Shell hides the system title bar and places content below a draggable strip with native macOS/Linux window controls or Onevra caption buttons on Windows; its opaque theme background covers the whole window, including the strip, with no header inset. `ShellOptions.HeaderLeading` replaces the title and `HeaderActions` builds trailing controls before Windows caption buttons. Transient tray panels are MyGo flyouts shown as system popovers, drawn by the application where the platform lacks one.

The sidebar page container reserves space for system window controls and overlays a transparent draggable header with a Lucide sidebar control. The toggle keeps additional clearance from macOS traffic lights. Its default page contains a collapsible left sidebar and a flexible main region.

Both regions and their application-built content extend to the top of the window behind the header. Applications own the top inset: use Shell's frame-aware header height to keep interactive content clear while painting backgrounds through the titlebar area. The header retains window dragging and control interaction above the body.

Applications can add trailing sidebar-header actions beside the toggle, clipped to the remaining sidebar safe area and omitted while collapsed or animating. Applications can supply the leading main-content titlebar area for a title, breadcrumbs or actions. It aligns with the main surface, clears the sidebar toggle when collapsed, and clips within the remaining native-control safe area.

The optional left bottom toolbar remains beneath the sidebar body and disappears with the sidebar or on replacement pages. It composes MyGo's toolbar, retaining its focus navigation and overflow behavior. Applications supply actions and labels; the reusable [icon button](../../source/shells/sidebar/icon_button.go) uses MyGo button semantics with transparent resting faces and translucent text-color interaction feedback.

The left sidebar keeps its content at the full sidebar width while its visible slice animates, so the opaque main region appears to cover or uncover it without resizing the window. A thin separator overlays the main region's left edge, takes no layout space and disappears when the sidebar is fully closed. Its overlaid drag handle spans the full window height, taking precedence over window dragging at the separator, and supports keyboard arrow adjustment; double-clicking restores the application-configured default width.

Applications can bound resizing with minimum and maximum widths. Shell retains the chosen width across toggles and page changes and removes the handle while collapsed or transitioning. The main region takes the remaining space rather than reserving a fixed share of the window.

A replacement page occupies the full window behind the shared floating header, with the same application-owned inset and hides the default page's titlebar content and sidebar controls. Returning restores the previous sidebar state without changing window geometry. Shell provides this switching mechanism, not a settings implementation or business router.

MyGo's animation values drive the sidebar reveal, without a separate timer or animation loop. Reversing a toggle mid-animation continues from the current visible width; Reduced Motion skips intermediate geometry.

## Settings composition

The public [Settings component](../../source/components/settings.go) renders application-defined categories, groups and setting rows independently of Shell or Dialog. Applications provide stable identities, localized metadata and synchronous control builders, and retain both presentation state and setting values.

Settings owns navigation, searchable presentation and the default row layout, including unsearched group header actions and descriptions composed from inline text and links, whose plain text search still matches; it does not load configuration, persist changes or prescribe when values take effect. Rows are slots: applications may add leading content, stack wide controls below the text or build the whole row, whose metadata keeps it searchable. Hosts choose the narrow presentation from the surface's own space, not the window width.

Small surfaces can omit search or navigation. Without navigation, every category appears on one page, so an application that wants its own navigation passes only the selected category.

Search spans all categories using metadata rather than building hidden controls. Results retain category and group context and allow editing in place; clearing search restores the selected category. Hosts can reuse the same metadata search without rendering the settings surface. Dialog and standalone-window lifecycles remain application-owned; a host may have Settings draw a close button whose action the host still owns. See the [compilable examples](../../source/components/settings_example_test.go) for composition.

## Dialog composition

Dialog composes MyGo overlay, modal focus scope and transition primitives rather than a modal manager. Applications own open state and build layers back to front; nested dialogs retain the underlying page. Each layer controls backdrop and Escape dismissal independently and animates in and out; nested layers close at once, as MyGo paints exits beneath their siblings.

AlertDialog composes the same surface and Onevra buttons, emitting a frame-local result instead of blocking the UI thread. Applications own localized labels, retain results when needed and provide scrolling for long dialog pages. MyGo owns focus trapping, restoration and reduced-motion behavior.

Shell supplies a window-local, synchronous build scope so every Onevra dialog, including nested AlertDialogs and dialogs on replacement pages, inherits its native window-drag strip. Standalone hosts can provide the same scope without depending on Shell; individual dialogs can override or disable dragging. The scope restores its previous value on return and does not introduce a modal manager.

Each modal layer reserves clearance and builds its own strip. This keeps underlying header actions blocked rather than forwarding pointer events through the modal; the strip does not dismiss the backdrop. Dragging uses MyGo's focus behavior, which may reset the focused input within the modal.

## Toast composition

The [toast component](../../source/components/toast.go) styles MyGo's toast queue through its unstyled viewport instead of a separate notification manager. Applications add, update and close toasts with MyGo's context methods and own their text, actions and asynchronous work; MyGo owns the queue, timing, pausing, Escape dismissal and announcements. Onevra positions visible toasts absolutely to overlap and fan them out; the stack takes the pointer so its gaps keep it open.

Both shells build the viewport after their content, inside a localization scope that carries the Shell's close-button label, because application localization scopes end with their builders. Shell also owns placement, adding its header height to top placements to clear the window controls. Status types select an icon; `Data` carries a leading slot, not type variants.

## Native appearance

MyGo and the operating system provide native window interaction semantics. On Windows, Shell replaces the system's window buttons (`HideWindowButtons`) with rounded toolbar-style minimize, maximize/restore and close buttons in the shared header, including replacement pages. Their names come from the catalog; close requests retain MyGo's cancellable application lifecycle. Maximize/restore follows the actual window state and brings up Windows 11's snap layouts.

Sidebar Shell requests a sidebar material on macOS and Mica on Windows. MyGo supports Windows system materials on Windows 11 22H2 and later; it owns compositor setup and native opaque fallback. Shell uses `Context.Vibrancy()` each frame, rather than the operating system or requested material, to decide whether to expose the backdrop.

When effective, the root and sidebar preserve transparency without additional tint in either theme. Otherwise, including Linux or unavailable composition, Shell paints an opaque theme background and contrasting sidebar surface through the header and bottom toolbar. Main content and replacement pages always remain opaque.

Applications can change or disable the material through the exposed MyGo window's `SetVibrancy`; MyGo rebuilds the view and Shell reapplies the same policy. Windows material rendering still requires target-platform native review.

Application builders can obscure the material by painting their own backgrounds. Preserve transparency through the window and sidebar when changing styling; content captures alone do not include the OS-composited blur.

## Theme ownership

Shell does not choose a theme. `mygo.Theme.SetSource` selects light or dark for the whole application, so native chrome, menus, web views and the frame theme agree. MyGo has no window-level theme hook, and Shell paints its own surfaces before Content, so its `Theme` option calls an application function at the start of each frame; nil keeps MyGo's frame theme. Other windows install theirs atop their builders.

The [theme package](../../source/theme/doc.go) is an optional preset following the frame's appearance and, like MyGo's default theme, the desktop's text scale and contrast, but not its accent.
