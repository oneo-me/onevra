# Changelog

User-facing updates for Onevra library consumers.

## [Unreleased]

- Breaking: Require MyGo 0.7.1.

## [0.2026.15] - 2026-10-23

- Require MyGo 0.6.3.
- Change `gesture.Classify` and `gesture.Navigation` to zoom on `ui.InputZoom` pinches.

## [0.2026.14] - 2026-10-21

- Breaking: Require MyGo 0.6.0.
- Breaking: Remove `chat.SwitchChip` in favor of `components.SwitchOptions.Compact`.
- Breaking: Remove `components.IconButtonToolbarExpanding` in favor of `components.IconButtonToolbar`.
- Add `chat.ComposerOptions.StopDisabled`.
- Add `components.HoneycombItem.Caption` and `components.HoneycombOptions.Gap`.
- Add `components.ScrollEdgeOptions.Uncolored`.
- Add `components.SwitchOptions.Compact`.
- Add `components.TabsOptions.Trailing`.
- Change `components.ScrollEdge` to tint its progressive blur with the background by default.
- Change `components.ScrollEdgeBar` to make its default surface and edge more opaque.
- Change `components.Switch` to widen its knob and lift it as draggable glass while held.
- Change `plain.Shell` and `sidebar.Shell` to bring up Windows 11 snap layouts from their maximize buttons.
- Change `shaders.SignalField` to repaint its drift 20 times a second.
- Fix `chat.EffortPicker` repainting the window while its icon stands still.

## [0.2026.13] - 2026-10-19

- Breaking: Make `components.Breadcrumbs` fill its parent's width.
- Breaking: Show `chart.Options.Detail` in a selection tooltip instead of a panel below the plot.
- Breaking: Show `treemap.Options.Tip` inside a tooltip like charts' instead of as the whole bubble.
- Breaking: Stack `components.Toasts` at the top center by default.
- Add `chat.EffortPicker` and `chat.EffortState`.
- Add `chat.FloatingConversationOptions.TopInset`.
- Add `components.AvatarGroupOptions.Overlap` and `components.AvatarGroupOptions.Reverse`.
- Add `components.HoneycombItem.Menu`, `components.HoneycombOptions.Curvature` and `components.HoneycombOptions.OffsetY`.
- Add `components.IconButtonCompact`.
- Add `components.IconButtonToolbarExpanding`.
- Add `components.ScrollEdgeBar` and `components.ScrollEdge`.
- Add `components.SettingsCategory.Placeholder`.
- Add `components.SettingsOptions.SectionActions`.
- Add `components.SidebarSectionWithActions`.
- Add `components.TabsWith`.
- Add `components.TabsWithLeading`.
- Add `plain.ShellOptions.ContentBehindHeader`.
- Add `flow.CardOptions.Body` and `flow.CardOptions.Trailing`.
- Add `treemap.Options.Menu`.
- Add `treemap.Style` with `treemap.Node.Style` and `treemap.Node.HoverStyle`.
- Change `components.Avatar` to drop its outline.
- Change `components.AvatarGroup` to turn the hovered member's separator into the accent.
- Change `components.Breadcrumbs` to collapse overflowing ancestors into an ellipsis menu.
- Change `chart.Bubble` to show its first pack without growing in.
- Change `chat.FloatingConversation` to blur the transcript progressively behind the lower half of the composer.
- Change `flow.View` edges to bend less, with rounded corners.
- Animate `treemap.View` drilling without stretching tiles, corners or labels.
- Localize component and Shell labels in Traditional Chinese, Japanese and Korean.
- Fix `chart.SplitBar` flashing segment labels in the theme text color as they fade.
- Fix `chart.SplitBar` showing a focus ring after a click.
- Fix `components.Breadcrumbs` layout shifting as items become current.
- Fix `flow.View` animating its layout as it appears.

## [0.2026.12] - 2026-10-18

- Add `components.Avatar` and `components.AvatarGroup`.
- Add `components.BadgeOptions` with tones and a leading icon.
- Add `components.CardStack`.
- Add `components.Carousel`.
- Add `chart.Bubble`.
- Add `chart.SplitBar`.
- Add `components/flow`.
- Add `components.CollapsibleOptions.Leading`.
- Add `components.HoldToConfirm`.
- Add `components.FloatingButtonGroup`.
- Add `components.HoneycombItem.Overlay` and `components.HoneycombOptions.OnPress`.
- Add `components.ImageCompare`.
- Add `components.SidebarItemWithSlots` and `components.SidebarItemWithContent`.
- Add `components.Skeleton` and `components.SkeletonContent`.
- Add `components.SplitOptions.Default`.
- Add `components.Stepper`.
- Add `components.TagInput`.
- Add `shaders.Ripple`.
- Add `components.TreeItemWithLeading`.
- Add `treemap.View`.
- Let `chat.Shimmer` shrink and end in an ellipsis.
- Animate `components.Badge` text and icon changes.
- Stack `components.Toasts` behind the newest toast and fan them out on hover or focus.
- Fix development-build panics when a closing `components.Dialog` contains `components.Checkbox`, `components.Radio`, `components.Accordion`, `components.Slider` or color controls.

## [0.2026.11] - 2026-10-16

- Breaking: remove the hybrid application template.
- Breaking: remove `components.Localization` and `components.WithLocalization`.
- Breaking: remove Shell label overrides `sidebar.Labels`, `plain.Labels` and their `SetLabels` and label options.
- Breaking: remove `Header`, `HeaderOverlay`, `HeaderTopInset` and `TopInset` from `chat.FloatingConversationOptions`.
- Breaking: rename `command.CommandPalette` to `command.CommandView`.
- Add `command.Options.Filter` and `command.State.Focus`.
- Add `HideSearch`, `HideSidebar` and `Close` to `components.SettingsOptions`.
- Add `Leading`, `Stacked` and `Content` to `components.SettingsItem`.
- Add `components.SplitPercent`.
- Add `components.TabsHighlight`.
- Add `shaders.SignalField`.
- Add `components.ListState.KeepOffset`.
- Add `chat.Chip`, `chat.SwitchChip`, `chat.Picker`, `chat.Attachment`, `chat.ContextMeter`, `chat.Quota`, `chat.QuotaSummary` and `chat.Shimmer`.
- Require MyGo 0.5.5.
- Change the default `sidebar.ShellOptions` window height and minimum height to 740.
- Return command results to the top when they change.
- Animate `components.Dialog` and `components.AlertDialog` as they open and close.
- Animate `components.Honeycomb` items as they are added or removed.
- Fix `components.Honeycomb` centering on first display and after its view resizes.

## [0.2026.10] - 2026-10-15

- Breaking: require MyGo 0.5.1.
- Breaking: replace `components/markdown` and `components/code` with MyGo's `ui/markdown` and `ui/editor`.
- Breaking: add a `TabsVariant` parameter to `components.Tabs`.
- Add `components.IconButton`.
- Add `components.ButtonToolbar`.
- Add `components.Honeycomb`.
- Add `HeaderLeading` and `HeaderActions` to `plain.ShellOptions`.
- Add the Lawn defense and Launcher examples.
- Add the official website.
- Show the window title in the plain Shell's header.
- Localize component and Shell labels in English and Simplified Chinese.
- Follow regional date formats in calendars.

## [0.2026.9] - 2026-10-14

- Breaking: replace Shell appearance options with `mygo.Theme.SetSource`.
- Breaking: replace Packer with MyGo's CLI.
- Add `components/graph`.
- Add `components/gesture`.
- Follow the desktop's text scale and contrast settings.

## [0.2026.8] - 2026-10-13

- Breaking: require MyGo 0.4.8.
- Add toasts.
- Add 13 Lucide icons.

## [0.2026.7] - 2026-10-12

- Breaking: require MyGo 0.4.6.
- Add `Leading` and `Trailing` slots to command items.
- Add `TopInset` to floating conversations.
- Add `StateKey` to remember window bounds.

## [0.2026.6] - 2026-10-12

- Breaking: require MyGo 0.4.3.
- Breaking: move command, code, markdown and chat from `plugins/` to `components/`.
- Add charts.
- Add macOS Info.plist entries to Packer.
- Fix the scroll-to-latest button in dark mode.

## [0.2026.5] - 2026-10-11

- Breaking: require MyGo 0.4.2.
- Add tables.
- Add a trailing slot to the composer.
- Add leading slots to sidebar items and settings categories.
- Bundle native libraries with Packer.
- Let conversation headers overlay messages.
- Keep Send available during input-method composition.

## [0.2026.4] - 2026-10-11

- Add `ShellOptions.Hidden`.
- Add `SettingsItem.DescriptionContent`.

## [0.2026.3] - 2026-10-11

- Breaking: require MyGo 0.4.1.
- Fix the composer's Send button not drawing in some themes.
- Fix a composer crash on paste.

## [0.2026.2] - 2026-10-11

- Breaking: require MyGo 0.4.0.
- Add an action slot to settings group headers.

## [0.2026.1] - 2026-10-10

Initial experimental release.

- Add themed controls, popovers, dialogs and settings.
- Add sidebar and plain window Shells.
- Add command palettes, Markdown and syntax highlighting.
- Add conversation views and a message composer.
- Add Lucide icons.
- Add desktop, menu bar and hybrid application templates.
