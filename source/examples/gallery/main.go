package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/oneo-me/onevra/source/examples/gallery/internal/gallery"
)

func main() {
	theme := flag.String("theme", "system", "Appearance: system, light, dark")
	flag.Parse()
	if *theme != "system" && *theme != "light" && *theme != "dark" {
		log.Fatal("theme must be system, light, or dark")
	}
	gallery.LogMyGoReview(log.Printf)
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	configDir := filepath.Join(home, ".onevra")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		log.Fatal(err)
	}
	mygo.App.SetPath(mygo.PathUserData, configDir)
	g := &gallery.Gallery{}
	mygo.App.SetName("Onevra Gallery")
	var override mygo.ThemeSource
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "theme" {
			override = mygo.ThemeSource(*theme)
		}
	})
	g.Configure(filepath.Join(configDir, "config.json"), override)
	mygo.App.WhenReady(func() {
		win := g.NewPersistent(filepath.Join(configDir, "state.json")).Window()
		if menu := g.ApplicationMenu(); menu != nil {
			mygo.App.SetMenu(menu)
		}
		g.EnableDesktopIntegrations()
		win.OnClosed(func() { mygo.App.Quit() })
	})
	mygo.App.SetDMABUFRenderer(mygo.DMABUFRendererAuto)
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
