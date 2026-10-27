package gallery

import (
	"log"
	"runtime"
	"sync"

	"github.com/egoist/mygo"
)

// Dock menus belong to the application, not to a window. Restore the previous
// menu on cleanup, but do not overwrite a menu another owner installed later.
type dockDemo struct {
	menu, previous *mygo.Menu
}

func (d *dockDemo) close() {
	if d.menu != nil && mygo.App.Dock.Menu() == d.menu {
		mygo.App.Dock.SetMenu(d.previous)
	}
	d.menu, d.previous = nil, nil
}

func (g *Gallery) showMainWindow() {
	if g.Shell == nil {
		return
	}
	win := g.Shell.Window()
	if win.IsMinimized() {
		win.Restore()
	}
	win.Show()
	win.Focus()
}

// dockIcon marks Show Gallery with the application's glyph, in its own
// colors. MyGo decodes it once for every menu made again.
var dockIcon = sync.OnceValue(func() *mygo.MenuIcon {
	icon, err := mygo.NewMenuIcon(trayIcon(false), false)
	if err != nil {
		log.Printf("Dock menu icon: %v", err)
	}
	return icon
})

func (g *Gallery) dockMenu() *mygo.Menu {
	return mygo.NewMenu([]*mygo.MenuItem{
		{ID: "show-gallery", Label: g.text("showGallery"), Icon: dockIcon(), Click: func(*mygo.MenuItem, *mygo.Window) {
			g.showMainWindow()
		}},
		mygo.Separator(),
		{ID: "welcome", Label: g.text("openWelcome"), Click: func(*mygo.MenuItem, *mygo.Window) {
			g.page = welcomePage
			g.settings.open = false
			g.showMainWindow()
		}},
		{ID: "settings", Label: g.text("openSettings"), Click: func(*mygo.MenuItem, *mygo.Window) {
			g.openSettingsFromMenu()
		}},
	})
}

func (g *Gallery) enableDockMenu() {
	if runtime.GOOS != "darwin" || g.Shell == nil || g.dock.menu != nil {
		return
	}
	g.dock.previous = mygo.App.Dock.Menu()
	g.dock.menu = g.dockMenu()
	mygo.App.Dock.SetMenu(g.dock.menu)
}
