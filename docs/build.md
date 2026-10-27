# Development

Use Go 1.27.3 or later. Run Go commands from `source/`; dependencies are pinned in [go.mod](../source/go.mod). No sibling project or workspace file is required.

Native windows require a desktop session. Linux also needs GTK 3, Pango and OpenGL system libraries. See the [MyGo native guides](https://github.com/egoist/mygo/tree/0f79256101ac/docs/ui) for platform setup.

## App bundles

Applications are packaged with MyGo's own CLI, pinned as a Go tool in [go.mod](../source/go.mod). Run it from `source/`:

```sh
go tool mygo dev
go tool mygo build -platform darwin/arm64,windows/amd64,linux/amd64,linux/arm64
```

[mygo.json](../source/mygo.json) owns the Gallery's packaging configuration. Its fields, platform outputs, resources, plugin native libraries, signing, notarization and signed updates follow MyGo's [configuration](https://mygo.egoist.dev/docs/configuration) and [distribution](https://mygo.egoist.dev/docs/distribution) guides; Onevra adds no packaging behavior of its own. VS Code's [tasks](../.vscode/tasks.json) provide `run` (`mygo dev`) and a default `build` task for the Gallery's release targets.

Packaging does not run code generators. The gallery uses its checked-in [MyGo component review inventory](architecture/gallery.md#component-review).

Everything in `source/resources/` ships as application payload, so icon sources live outside it. The gallery uses a 1024 × 1024 [PNG](../source/assets/icon.png), from which MyGo derives the Windows, Linux and fallback macOS icons, and the layered [Icon Composer project](../source/assets/icon.icon) on macOS, which requires Xcode 26 or later. The PNG is a static export of the layered icon's default appearance; refresh it when changing the design. The gallery's Image example loads the author's avatar from `resources/images/avatar.jpeg` through MyGo's resource path; direct `go run` uses the module's resource directory.

On macOS, the gallery's Terminal page requires macOS 13 or later. MyGo bundles the terminal plugin's library among the app's resources, where the plugin finds it. The first build may need network access to download it.

For an unpackaged run on any supported desktop, use `go run ./examples/gallery -theme dark`.

## Starter applications

Five independent applications provide small starting points alongside Gallery. From `source/`, run one of:

```sh
GOWORK=off go run ./examples/desktop
GOWORK=off go run ./examples/menubar
GOWORK=off go run ./examples/launcher
GOWORK=off go run ./examples/documents
GOWORK=off go run ./examples/games/lawn
```

| Example | Startup and close behavior | Settings host |
| --- | --- | --- |
| [Desktop](../source/examples/desktop/main.go) | Shows the main window. On macOS, closing it keeps the app in the Dock by default; Settings can change this to quit. Other platforms quit on close. | Dialog inside the main window |
| [Menu bar](../source/examples/menubar/main.go) | Shows a first-launch introduction, then runs from its tray icon without a main window. The quick-note panel is optional. | Independent window |
| [Launcher](../source/examples/launcher/main.go) | Runs without a Dock icon or a main window, with a tray icon and a global shortcut. The panel floats over other apps and hides on Escape or loss of focus; the app keeps running in the background. | Independent window |
| [Documents](../source/examples/documents/main.go) | A notes editor of Markdown and text files on `mygo.NewDocuments`: MyGo owns the File menu, a window per note, autosave, recovery and changes by other apps. | None; follows the system language |
| [Lawn defense](../source/examples/games/lawn/main.go) | A pixel-art tower-defense game on MyGo's game plugin, offscreen canvas and audio plugin: the round, its short effects and a looping tune are synthesized in memory, so the example ships no assets. Closing the window quits. | None; follows the system language |

The main, settings and introduction windows use `shells/plain`, with native window controls and the window title over an integrated draggable header instead of a default title bar. Launcher's panel is a MyGo panel window, so it floats over other apps and takes the keyboard without activating the launcher; its search and results are a `command.CommandView`.

Opening an already running application activates its existing instance. Desktop restores and focuses its main window, including after minimization. Menu bar shows its introduction until acknowledged, then its panel or settings. Explicit Quit always exits, regardless of window-close preferences. The examples use MyGo's single-instance lock and native activation events; they do not implement their own IPC or native runtime.

Trays follow the [tray convention](#tray-convention). In Desktop and Menu bar, left-click toggles the quick-note panel; Settings stays in the tray menu. Disabling the panel makes left-click open settings, an application-owned action that can be replaced when adapting the template. Launcher's left-click toggles its panel window instead of a popover. If icon creation fails, an accessible window displays the error; closing it exits rather than leaving an inaccessible background process.

English, Simplified Chinese, Traditional Chinese, Japanese, Korean and Follow system apply to application content, settings, window titles and menus; each example's messages are a MyGo catalog under `locales/`, and the choice also calls `mygo.App.SetLocale`, so MyGo's and Onevra's own strings and role menu items follow it. Unsupported system languages fall back to English.

Preferences and the menu bar introduction acknowledgement are saved as JSON in `config.json` under `os.UserConfigDir()/onevra-examples/<example>/`; each example has its own directory. Writes replace the file atomically. Read failures use defaults and save failures leave changes active for the current session, with feedback in the interface. Note text is intentionally session-only.

Each example has a local `mygo.json`. From an example's directory, such as `source/examples/desktop`, run `go tool mygo dev` or `go tool mygo build`. Builds write to separate `dist/<example>/` directories and need no Gallery resources or custom app icon. None shows a web page, so `linux.webview` is false and their Linux packages need no WebKitGTK. See [App bundles](#app-bundles).

To start a project, copy one complete example directory, including `locales/` and `mygo.json`, into the application's Go module. Change the application name, package identifier, preference directory and output path; provide module dependencies using this repository's [module guidance](../ARCHITECTURE.md#repository-landmarks). The examples have no imports from Gallery or sibling examples. Message keys and stored option values remain stable across language changes; extend every catalog together and run `go tool mygo generate` from the example directory, which `mygo dev` and `mygo build` also do.

For native lifecycle checks and settings captures, run from `source/` in a desktop session:

```sh
ONEVRA_NATIVE_SMOKE=1 ONEVRA_NATIVE_FOREGROUND=1 ONEVRA_SCREENSHOTS="$PWD/../dist/template-preview" \
  GOWORK=off go test -p 1 ./examples/desktop ./examples/menubar ./examples/launcher \
  -run '^TestNative' -count=1 -timeout 45s
```

The tests exercise startup, close-to-hide, restoration, settings hosting, language updates, panel-free startup, and quit cleanup. The launcher test additionally toggles its panel, filters a sample app and captures both appearances.

In a packaged app, also review actual Dock and tray clicks, opening the application again, right-click dismissal followed by left-click, keyboard shortcuts, focus, accessibility, both languages and themes, and panel placement across displays. Programmatic callbacks and content captures do not establish those native interactions. Windows and Linux require their own desktop review.

## Tray convention

Gallery and the starter applications share one tray pattern. Change it here and in every tray together:

- `mygo.NewTray` uses a 32 px PNG rendered in memory from the application's own glyph, black as a macOS template and in colors visible on light and dark taskbars elsewhere, with the application name as its tooltip. No packaged resource is needed.
- A stable `ID` lets macOS keep the icon where the user put it and Windows keep it shown or in the overflow.
- On macOS and Windows, left-click toggles the primary panel. Right-click closes it, then sets the tray menu only for the duration of `PopUpMenu`, because a permanent menu would consume primary clicks on macOS.
- Linux AppIndicator reports neither clicks nor icon bounds, so the menu stays installed and starts with an action opening the panel, which MyGo centers in the primary work area. It needs `libayatana-appindicator3` and a desktop supporting AppIndicator.
- The panel is a `mygo.NewFlyout` with `Tray`, `PlacementBottom`, `Focusable`, `Shadow` and `Popover`, so it is a system popover: the platform owns placement, light dismissal and, on macOS, the material. `VibrancyPopover` adds acrylic on Windows 11.
- Escape and Command+W also close the panel. Its root is transparent; where `c.Vibrancy()` reports no material, the panel draws its own themed surface.
- Each show creates a new flyout, so the application keeps panel state such as notes. The panel holds content only; the menu holds the panel action, application actions such as Settings or showing the main window, a separator and Quit.
- Language changes relabel the menu in place and reinstall it on Linux, where AppIndicator keeps a snapshot. Disabling the tray or quitting destroys the panel and icon.

## Gallery tray and Dock menu

Gallery has no pages for these integrations; it installs both at startup. Its tray icon is the Onevra artwork from [`tray.svg`](../source/examples/gallery/internal/gallery/tray.svg) and follows the [tray convention](#tray-convention): left-click toggles a brief popover with a Show Gallery action, and the menu offers Open tray panel, Show Gallery and Quit. On macOS, right-click the Dock icon to find Show Gallery, marked with the Onevra glyph, Open Welcome and Open Settings above the standard items; closing Gallery releases the tray and restores the previous Dock menu.

From `source/`, run `ONEVRA_NATIVE_SMOKE=1 ONEVRA_NATIVE_FOREGROUND=1 GOWORK=off go test ./examples/gallery/internal/gallery -run '^TestNative(Tray|DockMenu)$' -count=1 -timeout 30s` in a desktop session. The checks cover installation, panel toggling, menu callbacks, language updates and cleanup, not pointer interaction or appearance.

Manually verify the icon in both menu bar appearances, repeated left-click toggling, right-click menu dismissal followed by another left-click, Escape and click-away dismissal, the Dock shortcuts with Gallery visible, hidden or minimized, display edges and multiple displays. A packaged run is recommended to review the Dock icon's application identity. Windows and Linux require separate native review.

## Notification example

Select **Desktop / System notifications** and click **Send notification**. On macOS, use `go tool mygo dev` so Gallery has an application bundle, and allow notifications when prompted. An unpackaged executable reports that notifications are unavailable. Linux requires a desktop notification service; Windows uses notification-area balloons.

Sending as a message from the author adds a sender avatar on macOS only in a build signed with the Communication Notifications capability and its provisioning profile; Gallery's ordinary builds show it as an ordinary notification. Review delivery with sound enabled and disabled, denied permission, repeated sends, withdrawal, and clicking a notification while Gallery is hidden or minimized or showing another page.

Clicking restores the notification page; closing Gallery withdraws the demo notification. System notification settings and Do Not Disturb can suppress presentation even after successful submission. Check both interface languages and themes, narrow layouts and keyboard navigation. Windows and Linux need their own native review.

## Development checks

```sh
gofmt -l .
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go build -o ../dist/onevra-gallery ./examples/gallery
(cd .. && git diff --check)
```

Run checks locally before submitting changes. While iterating, test only the affected packages, for example `GOWORK=off go test ./components -run '^TestButton'`; the full race check takes several minutes, mostly in the gallery package. The only [GitHub Actions workflow](../.github/workflows/release.yml) creates a Release when a version tag is pushed; it runs no builds or tests. A successful cross-build is not native platform acceptance.

The formatting check should produce no output. Use a `.exe` output suffix on Windows. `GOWORK=off` checks that the module works independently of external workspaces.

## Native review

From `source/`, in a macOS desktop session:

```sh
ONEVRA_NATIVE_SMOKE=1 ONEVRA_SCREENSHOTS="$PWD/../dist/preview" \
  GOWORK=off go test ./examples/gallery/internal/gallery -run '^TestNative' -count=1 -timeout 5m -v
```

On macOS, these tests create hidden windows without activating the test application or adding a Dock icon. They capture appearance variants (including every Onevra control and MyGo example in light, dark and narrow layouts), exercise terminal input and cleanup, check that the settings dialog preserves window bounds, and destroy the windows. Other platforms still use visible windows.

Tests for tray, Dock, flyouts, input sources, application-menu activation and visible state restoration are skipped unless `ONEVRA_NATIVE_FOREGROUND=1` is also set; that switch also makes the capture windows visible for manual review. Starter lifecycle tests require the same foreground opt-in. Window captures leave a hole where a MyGo web view shows, so the WebView example also writes the page's own capture. Omit `ONEVRA_SCREENSHOTS` to use temporary output.

For Flyout captures and parent-window cleanup checks, add `ONEVRA_NATIVE_FOREGROUND=1` and use `-run '^TestNativeFlyout$'`. Review the editor, system-popover and material entry points, Escape and outside-press dismissal, placement near display edges, and note retention after reopening.

For command-palette captures, use the same environment with `-run '^TestNativeCommand(Expansion|InputSource)?$'`. The macOS input-source test additionally requires `ONEVRA_NATIVE_FOREGROUND=1`, ABC and another enabled input source, temporarily switches between them, and restores the original source; it checks automatic selection, restoration, manual overrides and retaining ABC on close.

In Gallery, review Command+F (Ctrl+F on other platforms), typing and Chinese input-method composition, arrow navigation, single-click and Enter activation, Escape dismissal, restored focus, and opening Settings from a command. Check both interface languages and themes, narrow windows and screen-reader announcements; captures alone do not verify these interactions.

For packaging changes, build the affected targets with `go tool mygo build` and run the packaged gallery in a desktop session, including the Terminal page; build output does not establish Finder or Dock appearance.

For UI changes, manually review both themes, narrow layouts, text scaling, window controls, keyboard navigation and accessibility. Verify that sidebar dragging respects configured width limits and double-clicking restores the default width without resizing the window. Inspect background blending over a contrasting backdrop. Captures do not replace interaction review; Windows and Linux need separate desktop checks. See the [gallery verification approach](architecture/gallery.md#verification-approach).

## Updating Lucide icons

Ordinary builds use checked-in generated Go assets and do not need Node or a Lucide checkout. To update, obtain a clean checkout of [Lucide](https://github.com/lucide-icons/lucide) and explicitly select its full commit hash. From `source/`:

```sh
go run ./tools/lucidegen -upstream /path/to/lucide -commit FULL_COMMIT_HASH
# Repeat without writing to check reproducibility:
go run ./tools/lucidegen -upstream /path/to/lucide -commit FULL_COMMIT_HASH -check
```

The generator validates every regular SVG with MyGo and records the commit and input digest in [upstream.json](../source/icons/lucide/upstream.json). It copies the full upstream license and reports accessor additions and removals. Review the generated diff for renames or removals before accepting an update: upstream changes can affect this library's public API. The test-only collection exercises every accessor without introducing a runtime registry.

The same invocation generates the gallery-only catalog directly from the upstream icon list, in filename order; `-check` verifies it alongside the library assets. The catalog is checked in and is not regenerated during ordinary gallery runs or packaging. Run the development checks and review the gallery's icon grid and icon samples in both appearances, including small sizes and complex strokes. Parsing tests alone do not prove visual fidelity. When changing the cache or generated representation, compare minimal named-icon and full-collection builds to check unused-resource elimination.

## Upgrading MyGo

On every MyGo upgrade:

1. Read the release notes and compare upstream changes against Onevra's components and Shell, especially adapted collection behavior, input, accessibility, animation and native appearance.
2. Adopt applicable fixes through upstream primitives instead of duplicating runtime internals. Preserve intentional Onevra interaction contracts.
3. From `source/`, run `GOWORK=off go generate ./examples/gallery/internal/gallery` and review the inventory diff. Packaging does not refresh it.
4. Add runnable gallery examples for every new component, window and application API, including methods and plugins the generator cannot discover. Give them localized labels, lifecycle cleanup and interaction coverage.
5. Never omit an example on your own judgment. The user decides exclusions; record each reason in the [decision map](architecture/gallery.md#component-review).
6. Add regression tests where Onevra's own behavior changes. Report review conclusions and native-verification gaps.

## Publication

The library module is `github.com/oneo-me/onevra/source`; the repository retains `source/` as the module root. Consumers can also use a local Go workspace containing their application and this repository's `source/` directory. Keep machine-specific workspace and replacement paths out of commits.

### Versioning

The first release is **0.2026.1**. Experimental releases use `0.YEAR.SEQUENCE`, with a sequence starting at 1 each year and increasing for every release, including fixes. For example, `0.2026.2` follows `0.2026.1`, and the first release of the next year is `0.2027.1`. Go treats these as semantic versions with major version zero; this calendar convention does not promise API stability or patch-level compatibility. Document consumer-impacting changes in the changelog. Reconsider the convention before declaring a stable major version.

| Purpose | First release |
| --- | --- |
| Release title / application version | `0.2026.1` |
| Go module version | `v0.2026.1` |
| Git tag | `source/v0.2026.1` |
| Consumer dependency | `github.com/oneo-me/onevra/source@v0.2026.1` |

The `source/` tag prefix identifies the module subdirectory; it is not part of the version passed to `go get`. See the [Go module source rules](https://go.dev/doc/modules/managing-source). Do not move or reuse a published version tag; ship corrections as a new version.

### Release preparation

Before an authorized release:

1. Run the development checks above with `GOWORK=off`; confirm `go mod tidy -diff` and `go mod verify` pass and dependencies can be downloaded without local replacements or private credentials.
2. Review native behavior on the platforms being claimed, including keyboard navigation, input methods and accessibility. macOS native captures are available; Windows and Linux still require their own desktop review. Do not imply native acceptance from compilation.
3. Follow the [update-summary rules](../AGENTS.md#working-with-the-user) when writing release notes.
4. Review tracked files and the history being made public for credentials, private configuration and unintended assets. Review dependency and bundled native library licenses when distributing application archives.
5. Retain the root [MIT license](../LICENSE), [adapted MyGo notices](../source/components/LICENSE.mygo) and [Lucide notices](../source/icons/lucide/LICENSE).
6. Review the shipping scope in [CHANGELOG.md](../CHANGELOG.md). At release time, move only shipping entries from `Unreleased` into the confirmed version and retain `Unreleased` for pending changes.
7. Leave Gallery's New components and Updated components groups unchanged so the release still marks the pages it ships. The first change after the release prunes them, as described beside the [lists](../source/examples/gallery/internal/gallery/navigation.go).
8. Keep both READMEs accurate about release availability and any repository access requirements.
9. Align Gallery's application version with `mygo.json` when shipping Gallery alongside the module. Gallery archives are optional release attachments; the Go library is distributed as source and needs no application archive.

### Authorized publication

A request such as “publish 0.2026.1” authorizes the complete module release workflow below, including the necessary release commit, tag and pushes. Do not ask separately for each authorized step. Preparing a release or recording these rules is not a request to publish. The [release workflow](../.github/workflows/release.yml) creates the GitHub Release page after the version tag is pushed.

1. Complete release preparation, update the changelog and both READMEs to reflect the requested release, and verify the intended branch, commit scope and GitHub remote.
2. Confirm the repository is publicly readable. Changing repository visibility is a separate action requiring explicit authorization.
3. Commit the reviewed release changes if needed, preserving unrelated work, and push the release commit to the intended branch. If the intended destination is ambiguous, resolve it before pushing.
4. Create an annotated tag on that exact commit using `source/v<VERSION>` and push that specific tag. For `0.2026.1`, use `source/v0.2026.1`, never a root-level `v0.2026.1` or `source/0.2026.1`.
5. Do not force-push or replace an existing version tag. If it already identifies the intended published commit, proceed to verification.
6. Check that the tag-triggered GitHub Release workflow succeeds, verify public module retrieval and build an independent consumer as described below. Report the commit, tag, Release URL, download result and any remaining indexing or native-verification gaps.

Use the requested version without `v` in human-facing titles and application metadata, and with `v` in Go module commands. The public source commit and tag are the module release; no package upload is needed. The tag push automatically creates the GitHub Release page. Distribute Gallery archives only when requested; they are not required for the module release.

The workflow runs one Ubuntu job only for tags matching `source/v*`, validates `0.YEAR.SEQUENCE`, and uses the version without `v` as the Release title. It extracts notes from exactly one nonempty `## [VERSION]` section in the tagged commit's `CHANGELOG.md` (an optional date suffix is supported). Move shipping notes out of `Unreleased` before tagging; missing or empty version notes fail the job.

Existing Releases are left unchanged on reruns. The workflow uses GitHub's built-in token with `contents: write`, needs no additional secret, and does not build or upload application archives. Ensure GitHub Actions is enabled and this workflow is included in the tagged commit. If creation fails after the tag is public, fix the cause and rerun the failed job without moving the tag; the Go module is already published independently of the Release page.

Once the tag is public, run the following from a fresh consumer module without workspace overrides or local replacements:

```sh
GOWORK=off GOPROXY=https://proxy.golang.org go list -m github.com/oneo-me/onevra/source@v0.2026.1
GOWORK=off go get github.com/oneo-me/onevra/source@v0.2026.1
```

Build the README example and check [pkg.go.dev](https://pkg.go.dev/github.com/oneo-me/onevra/source) for package documentation and license detection. Go includes the repository-root `LICENSE` in a subdirectory module archive when that directory has no license of its own. Verify the downloaded archive contains the license and required embedded assets. These public download and indexing checks cannot be completed before publication. See the [official publishing workflow](https://go.dev/doc/modules/publishing).
