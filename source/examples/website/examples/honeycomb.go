package examples

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

var _ = registerWindow(&launcher{selected: "documents"})

type launcherApp struct {
	id, name string
	icon     func() *ui.SVG
	color    ui.Color
}

var launcherApps = []launcherApp{
	{"documents", "Documents", lucide.FileText, ui.RGB(79, 124, 255)},
	{"photos", "Photos", lucide.Image, ui.RGB(236, 120, 82)},
	{"notes", "Notes", lucide.StickyNote, ui.RGB(226, 178, 60)},
	{"music", "Music", lucide.Music, ui.RGB(206, 96, 176)},
	{"files", "Files", lucide.Folder, ui.RGB(72, 168, 128)},
	{"mail", "Mail", lucide.Mail, ui.RGB(120, 176, 92)},
	{"camera", "Camera", lucide.Camera, ui.RGB(150, 118, 226)},
	{"cloud", "Cloud", lucide.Cloud, ui.RGB(122, 162, 214)},
	{"settings", "Settings", lucide.Settings, ui.RGB(110, 110, 122)},
	{"videos", "Videos", lucide.Video, ui.RGB(88, 128, 196)},
	{"archive", "Archive", lucide.Archive, ui.RGB(158, 138, 96)},
	{"inbox", "Inbox", lucide.MessageSquare, ui.RGB(66, 176, 176)},
}

type launcher struct {
	selected string
}

func (l *launcher) View(c *ui.Context) {
	items := make([]components.HoneycombItem, len(launcherApps))
	for i, a := range launcherApps {
		items[i] = components.HoneycombItem{
			ID: a.id, Label: a.name, Caption: a.name, Ring: a.color,
			Content: func(c *ui.Context, size float32) { launcherIcon(c, a, size) },
		}
	}
	// A folder shows the apps it holds as a group of images.
	var folder []components.HoneycombImage
	for _, a := range launcherApps[1:4] {
		folder = append(folder, components.HoneycombImage{Content: func(c *ui.Context, size float32) { launcherIcon(c, a, size) }})
	}
	items = append(items, components.HoneycombItem{ID: "media", Label: "Media", Caption: "Media", Images: folder})
	ui.Column(c).Fill().Children(func() {
		if id := components.Honeycomb(c, components.HoneycombOptions{
			Items:    items,
			Selected: l.selected,
		}); id != "" {
			l.selected = id
		}
	})
}

// launcherIcon fills a cell's circle, which the honeycomb does not clip.
func launcherIcon(c *ui.Context, a launcherApp, size float32) {
	ui.Box(c).Size(size, size).Radius(size/2).Background(a.color.Alpha(.16)).
		Border(1, a.color.Alpha(.45)).Center().Children(func() {
		components.Icon(c, a.icon).Size(size*.42, size*.42).TextColor(a.color)
	})
}
