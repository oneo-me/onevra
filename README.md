# Onevra

**English** | [简体中文](README.zh-CN.md)

Onevra is a Go component library for native desktop applications, built on [MyGo](https://github.com/egoist/mygo). It provides themed controls, window shells, Lucide icons and complex components. The name combines `One` from **ONEO** with the coined ending `vra`.

![Onevra Gallery main interface with sidebar navigation and the Welcome page](docs/images/gallery.png)

## Design principles

Onevra builds on MyGo rather than replacing it:

- **Refine built-in components**: adjust the styling of MyGo's built-in components, and merge and simplify them into a smaller, consistent API.
- **Fill the gaps**: provide components that MyGo does not officially support.
- **Defer to upstream**: when MyGo officially provides a feature or component that Onevra already offers, Onevra may remove its own version and switch to MyGo's.

## Additional components

Building on MyGo's native UI capabilities, Onevra adds composed components for common application scenarios:

- **Conversation interfaces**: bring message turns, tool activity, folded summaries and attachment input together into a complete chat interface.

- **Lucide icons**: provide SVG icons for MyGo through named accessors, loaded on demand with configurable sizing and inherited text color.

## Examples

Follow [MyGo’s official requirements](https://mygo.egoist.dev/docs) for the development environment and system dependencies.

From the repository root, start the interactive Gallery:

```sh
cd source
go run ./examples/gallery
```

Append `-theme light` or `-theme dark` to preview an appearance. Five smaller applications are also available; run these commands from `source/`:

| Example | Purpose | Command |
| --- | --- | --- |
| Desktop | A main window with settings | `go run ./examples/desktop` |
| Menu bar | A tray application with a quick-note panel | `go run ./examples/menubar` |
| Launcher | A Spotlight-style panel that searches and opens installed apps | `go run ./examples/launcher` |
| Documents | A document-based notes editor | `go run ./examples/documents` |
| Lawn defense | A pixel-art tower-defense game | `go run ./examples/games/lawn` |

Each template includes localization and its own packaging configuration; all but the document editor and the game keep preferences. See [starter applications](docs/build.md#starter-applications) for copying and adapting one. Windows and Linux still require their own [native review](docs/build.md#native-review).

## Getting started

Install **0.2026.16** in your application module with:

```sh
go get github.com/oneo-me/onevra/source@v0.2026.16
```

Save this minimal application as `main.go` and run `go run .` from your application module:

```go
package main

import (
    "log"

    "github.com/egoist/mygo"
    "github.com/egoist/mygo/ui"
    "github.com/oneo-me/onevra/source/components"
    "github.com/oneo-me/onevra/source/shells/plain"
    "github.com/oneo-me/onevra/source/theme"
)

func main() {
    mygo.App.WhenReady(func() {
        shell := plain.NewShell(plain.ShellOptions{
            Title: "Hello Onevra",
            Theme: theme.Onevra,
            Content: func(c *ui.Context) {
                ui.Column(c).Fill().Padding(24).Children(func() {
                    components.Text(c, "Hello, Onevra!")
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

Create windows after MyGo is ready. Build content with Onevra controls and MyGo layout primitives; your application owns state, persistence and asynchronous work.

## Packaging

Package applications with MyGo's own CLI; Onevra adds no packager. To run or package Gallery, execute these commands from `source/`:

```sh
go tool mygo dev
go tool mygo build
```

`dev` builds and launches the host application with live reload; `build` packages it for the host, or for the targets given with `-platform`. Configuration lives in `mygo.json`. See [packaging the Gallery](docs/build.md#app-bundles) and MyGo's [distribution guide](https://mygo.egoist.dev/docs/distribution).

## Shells

Shells create native windows while leaving application lifecycle, business state and the theme to the caller.

| Package | Layout |
| --- | --- |
| [`shells/plain`](source/shells/plain/shell.go) | A single content area with automatic clearance for window controls and the window title shown in its header; used in the example above |
| [`shells/sidebar`](source/shells/sidebar/shell.go) | A collapsible, resizable sidebar with a content area, header actions and replaceable pages |

Import either package under `github.com/oneo-me/onevra/source/`. Their `Theme` option installs the application's theme before each frame: pass `theme.Onevra` for Onevra's palettes, your own function returning a `*ui.Theme`, or nothing to keep MyGo's theme. Choose light, dark or system appearance with `mygo.Theme.SetSource`; see the [theme package](source/theme/doc.go). For Sidebar, use `sidebar.HeaderHeight(c)` when content needs to clear the floating header; see [layout guidance](docs/architecture/components.md#page-container-and-layout).

## Icons

The [Lucide package](source/icons/lucide/doc.go) provides named SVG accessors. Import `github.com/oneo-me/onevra/source/icons/lucide` alongside `components`, then use them in a content builder:

```go
components.Icon(c, lucide.Search).Size(20, 20)
```

Icons are parsed on first use and reused. They inherit text color, and the Icon component also accepts a parsed MyGo SVG.

## Components

Import `github.com/oneo-me/onevra/source/components` for themed inputs, selection, navigation, collections and overlays. Inside a content builder, use retained application state such as `agreed` and `wifi`:

```go
if components.Button(c, "Save", components.ButtonPrimary).Clicked() {
    // Save application state.
}
components.Checkbox(c, &agreed, "I agree")
components.Switch(c, &wifi, "Wi-Fi")
```

Components read the frame theme and work independently of Shell; a standalone window installs its theme with `c.SetTheme` or `theme.Apply` at the top of its builder. See the [component contract](source/components/doc.go) and [theme usage](source/theme/doc.go). Complex components each have a separate package under `components/`: `chart`, `graph`, `treemap`, `flow`, `gesture`, `command` and `chat`. See [complex component composition](docs/architecture/components.md#complex-component-composition).

## Contributing

Read [AGENTS.md](AGENTS.md) and the [architecture overview](ARCHITECTURE.md) before making changes. From `source/`, run:

```sh
gofmt -l .
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go build ./...
```

Formatting should produce no output. See [development checks](docs/build.md#development-checks) and [native review](docs/build.md#native-review) for verification. Keep both READMEs equivalent and record consumer-visible changes in [CHANGELOG.md](CHANGELOG.md). Versioning and release procedures are in [Publication](docs/build.md#publication).

## Website

The documentation site lives in [website/](website): guides and a page for every component, whose live examples are Go compiled to WebAssembly. The [Website workflow](.github/workflows/website.yml) deploys it to GitHub Pages. It needs Bun and Go 1.27.3 or later:

```sh
cd website
bun install
bun run dev
```

See [website/README.md](website/README.md) for adding pages and examples.

## Author

[Blog](https://oneo.me) · [X @oneo666](https://x.com/oneo666)

## License

Project code: [MIT](LICENSE). Components adapted from MyGo retain the [upstream MIT notice](source/components/LICENSE.mygo). Lucide assets retain their [upstream ISC and Feather MIT notices](source/icons/lucide/LICENSE).
