# Onevra

[English](README.md) | **简体中文**

Onevra 是基于 [MyGo](https://github.com/egoist/mygo) 的原生桌面 Go 组件库，提供主题控件、窗口 Shell、Lucide 图标和复杂组件。名字由 **ONEO** 中的 `One` 与自造词尾 `vra` 组合而成。

![Onevra Gallery 主界面，包含侧边导航和欢迎页](docs/images/gallery.png)

## 设计原则

Onevra 构建在 MyGo 之上，而不是取代它：

- **微调内置组件**：微调 MyGo 内置组件的风格，并将其合并、简化为更精简一致的 API。
- **补齐缺失组件**：提供 MyGo 官方不支持的组件。
- **优先采用官方实现**：当 MyGo 官方提供了 Onevra 已有的功能或组件时，Onevra 可能会移除自己的版本，转而使用 MyGo 官方版本。

## 补充组件

在 MyGo 官方提供的原生 UI 能力之上，Onevra 重点补充面向应用场景的组合组件：

- **对话界面**：将消息轮次、工具执行过程、折叠摘要和附件输入组合成完整的聊天界面。

- **Lucide 图标**：提供可直接用于 MyGo 的 SVG 图标，通过具名函数按需加载，支持尺寸设置与文字颜色继承。

## 示例程序

开发环境与系统依赖遵循 [MyGo 官方要求](https://mygo.egoist.dev/docs)。

从仓库根目录启动可交互的 Gallery：

```sh
cd source
go run ./examples/gallery
```

追加 `-theme light` 或 `-theme dark` 可预览指定外观。另外提供五个小型应用，在 `source/` 中运行：

| 示例 | 用途 | 命令 |
| --- | --- | --- |
| Desktop | 带设置界面的主窗口应用 | `go run ./examples/desktop` |
| Menu bar | 带便笺面板的托盘应用 | `go run ./examples/menubar` |
| Launcher | Spotlight 风格的浮层，搜索并打开已安装应用 | `go run ./examples/launcher` |
| Documents | 基于文档的笔记编辑器 | `go run ./examples/documents` |
| Lawn defense | 像素风塔防游戏 | `go run ./examples/games/lawn` |

每个模板都包含多语言及独立打包配置，除文档编辑器和游戏外均包含偏好设置；复制与修改方式见[应用模板](docs/build.md#starter-applications)。Windows 和 Linux 仍需各自的[原生界面验证](docs/build.md#native-review)。

## 开发者入门

在应用模块中安装 **0.2026.16**：

```sh
go get github.com/oneo-me/onevra/source@v0.2026.16
```

将以下最小应用保存为 `main.go`，在应用模块中执行 `go run .`：

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

在 MyGo 就绪后创建窗口，用 Onevra 控件和 MyGo 布局原语构建内容；状态、持久化和异步任务由应用管理。

## 打包

应用使用 MyGo 官方 CLI 打包，Onevra 不提供单独的打包工具。在 `source/` 中运行或打包 Gallery：

```sh
go tool mygo dev
go tool mygo build
```

`dev` 编译并启动当前平台的应用，支持热重载；`build` 为当前平台或 `-platform` 指定的目标打包。配置位于 `mygo.json`。详见 [Gallery 打包](docs/build.md#app-bundles)和 MyGo 的[分发指南](https://mygo.egoist.dev/docs/distribution)。

## Shells

Shell 负责创建原生窗口，应用生命周期、业务状态和主题由调用方管理。

| 包 | 布局 |
| --- | --- |
| [`shells/plain`](source/shells/plain/shell.go) | 单内容区，自动避让窗口控制区，并在头部显示窗口标题；上方最小示例使用此布局 |
| [`shells/sidebar`](source/shells/sidebar/shell.go) | 可折叠、可调整宽度的侧栏与内容区，支持标题栏操作及页面替换 |

通过 `github.com/oneo-me/onevra/source/` 下的对应路径导入。两者的 `Theme` 选项会在每帧构建前安装应用的主题：传入 `theme.Onevra` 使用 Onevra 配色，传入返回 `*ui.Theme` 的自定义函数，或留空保留 MyGo 主题。使用 `mygo.Theme.SetSource` 选择浅色、深色或跟随系统，详见 [theme 包](source/theme/doc.go)。Sidebar 内容需要避让悬浮标题栏时，使用 `sidebar.HeaderHeight(c)`，详见[布局说明](docs/architecture/components.md#page-container-and-layout)。

## 图标

[Lucide 包](source/icons/lucide/doc.go) 提供具名 SVG 访问函数。导入 `github.com/oneo-me/onevra/source/icons/lucide` 和 `components`，在内容构建函数中使用：

```go
components.Icon(c, lucide.Search).Size(20, 20)
```

图标首次使用时解析，之后复用，并继承文字颜色。Icon 组件也接受已解析的 MyGo SVG。

## 组件

导入 `github.com/oneo-me/onevra/source/components` 使用带主题的输入、选择、导航、集合与浮层组件。在内容构建函数中使用由应用持有的 `agreed`、`wifi` 等状态：

```go
if components.Button(c, "Save", components.ButtonPrimary).Clicked() {
    // Save application state.
}
components.Checkbox(c, &agreed, "I agree")
components.Switch(c, &wifi, "Wi-Fi")
```

组件读取当前帧主题，可脱离 Shell 独立使用；独立窗口在构建函数开头通过 `c.SetTheme` 或 `theme.Apply` 安装主题。参见[组件契约](source/components/doc.go)和[主题用法](source/theme/doc.go)。复杂组件各自在 `components/` 下保留独立包：`chart`、`graph`、`treemap`、`flow`、`gesture`、`command` 和 `chat`。详见[复杂组件组合](docs/architecture/components.md#complex-component-composition)。

## 参与开发

修改前阅读 [AGENTS.md](AGENTS.md) 和[架构概览](ARCHITECTURE.md)。在 `source/` 中运行：

```sh
gofmt -l .
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go build ./...
```

格式检查应无输出，完整验证要求见[开发检查](docs/build.md#development-checks)和[原生界面检查](docs/build.md#native-review)。保持中英文 README 内容一致，并将面向使用者的变化记录到 [CHANGELOG.md](CHANGELOG.md)。版本规则与发布流程见[发布说明](docs/build.md#publication)。

## 官网

文档站点位于 [website/](website)：包含指南和每个组件的页面，页面中的实时示例由 Go 编译为 WebAssembly 运行。[Website workflow](.github/workflows/website.yml) 会将其部署到 GitHub Pages。运行它需要 Bun 和 Go 1.27.3 或更高版本：

```sh
cd website
bun install
bun run dev
```

添加页面和示例的方法见 [website/README.md](website/README.md)。

## 作者

[博客](https://oneo.me) · [X @oneo666](https://x.com/oneo666)

## 协议

项目代码采用 [MIT](LICENSE) 协议。改编自 MyGo 的控件保留[上游 MIT 声明](source/components/LICENSE.mygo)。Lucide 资源保留[上游 ISC 和 Feather MIT 声明](source/icons/lucide/LICENSE)。
