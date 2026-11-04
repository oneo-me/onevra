package examples

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

var _ = register(&fileIcons{})

type fileIcons struct{}

func (fileIcons) View(c *ui.Context) {
	ui.Grid(c).Columns(3).Gap(12).Children(func() {
		for _, name := range []string{"main.go", "go.mod", "App.tsx", "index.ts", "package.json", "README.md", "logo.svg", "build.sh", ".gitignore"} {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				components.FileIcon(c, name).Size(16, 16)
				components.Text(c, name).FontSize(13)
			})
		}
	})
}
