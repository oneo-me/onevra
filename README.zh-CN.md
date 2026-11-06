<p align="center">
  <img src="source/assets/icon.png" alt="Onevra" width="128">
</p>

<h1 align="center">Onevra</h1>

<p align="center">
  <strong>简体中文</strong> |
  <a href="README.md">English</a>
</p>

<p align="center">
  <a href="https://github.com/oneo-me"><img alt="GitHub" src="https://img.shields.io/badge/GitHub-oneo--me-181717?logo=github&logoColor=white"></a>
  <a href="https://oneo.me"><img alt="博客" src="https://img.shields.io/badge/%E5%8D%9A%E5%AE%A2-oneo.me-0969da"></a>
  <a href="https://x.com/oneo666"><img alt="X" src="https://img.shields.io/badge/X-%40oneo666-000000?logo=x&logoColor=white"></a>
  <a href="https://github.com/sponsors/oneo-me"><img alt="GitHub 赞助" src="https://img.shields.io/badge/GitHub%20%E8%B5%9E%E5%8A%A9-ea4aaa?logo=githubsponsors&logoColor=white"></a>
</p>

<p align="center">
  <a href="https://awesomemygo.dev/ui"><img alt="已收录于 Awesome MyGo" src="https://awesomemygo.dev/badge.svg"></a>
</p>

<p align="center">
  Onevra 是基于 <a href="https://github.com/egoist/mygo">MyGo</a> 的原生桌面 Go 组件库，提供主题控件、窗口 Shell、Lucide 图标和复杂组件。
</p>

<p align="center">
  <img src="docs/images/gallery.png" alt="Onevra Gallery 主界面，包含侧边导航和欢迎页">
</p>

## 安装

Onevra 需要 Go 1.27.3 或更高版本。在应用模块中安装 **0.2026.18**：

```sh
go get github.com/oneo-me/onevra/source@v0.2026.18
```

## 用法

将下面这个最小应用保存为 `main.go`，然后运行 `go run .`：

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

在 MyGo 就绪后创建窗口。使用 Onevra 控件和 MyGo 布局原语构建内容；应用负责状态、持久化和异步工作。参见[组件契约](source/components/doc.go)和此处用到的 [Shell](source/shells/plain/shell.go)。

## 文档

| 主题 | 文档 |
| --- | --- |
| 指南与全部组件 | [oneo-me.github.io/onevra](https://oneo-me.github.io/onevra/) |
| 组件契约 | [source/components/doc.go](source/components/doc.go) |
| 窗口 Shell | [plain](source/shells/plain/shell.go)、[sidebar](source/shells/sidebar/shell.go) |
| 主题 | [source/theme/doc.go](source/theme/doc.go) |
| Lucide 图标 | [source/icons/lucide/doc.go](source/icons/lucide/doc.go) |
| 组件设计 | [docs/architecture/components.md](docs/architecture/components.md) |
| 架构 | [ARCHITECTURE.md](ARCHITECTURE.md) |
| 开发指南 | [docs/build.md](docs/build.md) |

## 开发

在 `source/` 中运行以下命令检查改动：

```sh
gofmt -l .
GOWORK=off go test ./...
GOWORK=off go vet ./...
GOWORK=off go build ./...
```

- 用 Gallery 预览全部组件：`go run ./examples/gallery`；追加 `-theme light` 或 `-theme dark` 选择外观。
- 用 MyGo CLI 打包应用：`go tool mygo dev` 或 `go tool mygo build`。参见[应用包](docs/build.md#app-bundles)。
- 运行 `source/examples/` 下的五个起始应用；参见[起始应用](docs/build.md#starter-applications)。
- 开发文档站点：`cd website && bun install && bun run dev`；参见[website/README.md](website/README.md)。
- 修改代码前先阅读 [AGENTS.md](AGENTS.md)。

影响使用者的变更记录在 [CHANGELOG.md](CHANGELOG.md)。

## 许可证

项目代码：[MIT](LICENSE)。改编自 MyGo 的组件保留[上游 MIT 声明](source/components/LICENSE.mygo)。Lucide 资源保留[上游 ISC 与 Feather MIT 声明](source/icons/lucide/LICENSE)。
