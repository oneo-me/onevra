package gallery

import (
	"log"
	"math"
	"sync"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// Keep the application version aligned with mygo.json; a test checks both.
const galleryVersion = "0.2026.16"

const (
	repositoryURL   = "https://github.com/oneo-me/onevra"
	repositoryLabel = "github.com/oneo-me/onevra"
)

func (g *Gallery) Welcome(c *ui.Context) {
	g.withLanguage(c, func() { g.welcomeLocalized(c) })
}

func (g *Gallery) welcomeLocalized(c *ui.Context) {
	t := c.Theme()
	_, height := c.Size()
	pad := t.Space(6)
	if g.welcomeContentWidth(c) < 400 {
		pad = t.Space(3)
	}
	components.ScrollView(c, components.ScrollViewOptions{}).Key("welcome").Fill().Children(func() {
		ui.Column(c).FillWidth().MinWidth(0).MinHeight(height * .8).MaxWidth(t.Space(170)).AlignSelf(ui.Center).
			Justify(ui.Center).AlignItems(ui.Center).Padding(pad).Gap(t.Space(8)).Children(func() {
			g.welcomeTitle(c)
			welcomeEnter(c, components.Text(c, tr(c, "builtWithOnevraPoweredByMygoWrittenInGo")).
				FontSize(t.Rem(.8)).TextColor(t.TextMuted).TextAlign(ui.Center).Wrap(), 4)
		})
	})
}

// welcomeBackground tints the page behind the titlebar inset as well as the
// scrolling content.
func welcomeBackground(c *ui.Context, content ui.Element) {
	t := c.Theme()
	welcomeBackdrop(c, content,
		welcomeGlow{x: 1, y: 0, radius: .35, color: t.Accent, alpha: .14},
		welcomeGlow{x: 0, y: 1, radius: .35, color: brandSky, alpha: .16},
	)
}

// welcomeIconField fills the page with a faint, tilted lattice of icons whose
// tint runs diagonally from the accent to the brand color. It paints icons
// directly rather than building an element for each one, and fades them
// around the center so the title stays legible.
func welcomeIconField(c *ui.Context) {
	t := c.Theme()
	width, height := c.Size()
	icons := welcomeFieldIcons()
	step, size := t.Space(13), t.Space(5)
	alpha := float32(.13)
	if t.Dark {
		alpha = .16
	}
	// Clip to the content so the oversized, rotated field never reaches
	// under the translucent sidebar.
	ui.Box(c).Key("icon-field").Absolute().Left(0).Top(0).Right(0).Bottom(0).Clip().PassThrough().Children(func() {
		ui.Box(c).Absolute().Left(-width*.25).Top(-height*.25).
			Size(width*1.5, height*1.5).Rotate(-14).PassThrough().
			Draw(func(p *ui.Painter, r ui.Rect) {
				cx, cy := r.X+r.W/2, r.Y+r.H/2
				reach := float32(math.Hypot(float64(width), float64(height))) / 2
				for row := 0; float32(row)*step < r.H; row++ {
					offset := float32(row%2) * step / 2
					for col := 0; float32(col)*step < r.W+step; col++ {
						x, y := r.X+float32(col)*step-offset, r.Y+float32(row)*step
						diagonal := min(1, max(0, (x-r.X)/r.W*.5+(1-(y-r.Y)/r.H)*.5))
						dx, dy := (x-cx)/reach, (y-cy)/reach
						center := min(1, max(.25, float32(math.Sqrt(float64(dx*dx+dy*dy)))*1.4))
						color := brandSky.Mix(t.Accent, diagonal)
						icon := icons[(row*7+col*3)%len(icons)]
						p.Icon(icon, ui.Rect{X: x, Y: y, W: size, H: size}, color.Alpha(alpha*center*(.45+diagonal*.55)))
					}
				}
			})
	})
}

// welcomeFieldIcons samples the catalog sparsely, so the field parses a few
// dozen icons once rather than the whole library.
var welcomeFieldIcons = sync.OnceValue(func() []*ui.SVG {
	icons := make([]*ui.SVG, 48)
	for i := range icons {
		icons[i] = iconCatalog[(i*97+13)%len(iconCatalog)].get()
	}
	return icons
})

// welcomeTitle presents the product name, its repository and version.
func (g *Gallery) welcomeTitle(c *ui.Context) {
	t := c.Theme()
	welcomeEnter(c, welcomeBrandMark(c, t.Space(18)), 0)
	welcomeEnter(c, ui.Column(c).FillWidth().AlignItems(ui.Center).Gap(t.Space(3)), 1).Children(func() {
		// The name is one word, so it shrinks to fit rather than wrapping.
		components.Text(c, "Onevra").FontSize(min(t.Rem(4.2), (g.welcomeContentWidth(c)-t.Space(8))/3.2)).SingleLine().FontWeight(800).LetterSpacing(-2).LineHeight(1)
		components.Text(c, tr(c, "smallComponentsRealPossibilities")).
			FontSize(t.Rem(1.3)).FontWeight(600).TextAlign(ui.Center).Wrap()
		components.Text(c, tr(c, "welcomeEntranceSummary")).
			MaxWidth(t.Space(120)).TextColor(t.TextMuted).LineHeight(1.45).TextAlign(ui.Center).Wrap()
	})
	compact := g.welcomeContentWidth(c) < 400
	welcomeEnter(c, ui.Row(c).FillWidth().MinWidth(0).Wrap().Justify(ui.Center).AlignItems(ui.Center).Gap(t.Space(3)), 2).Children(func() {
		repo := components.Button(c, "", components.ButtonOutline).Label(repositoryLabel).
			Gap(t.Space(2)).Radius(t.Space(5)).MinWidth(0).MaxWidthPercent(100).Background(t.Surface.Alpha(.8))
		repo.Children(func() {
			components.Icon(c, lucide.GitBranch).Size(t.Space(4), t.Space(4)).Shrink(0)
			components.Text(c, repositoryLabel).Wrap()
			if !compact {
				components.Icon(c, lucide.ExternalLink).Size(t.Space(3.5), t.Space(3.5)).Shrink(0).TextColor(t.TextMuted)
			}
		})
		if repo.Clicked() {
			g.openRepository()
		}
		_, shortcut := commandShortcut()
		search := components.Button(c, "", components.ButtonPrimary).Label(tr(c, "welcomeSearchEverything")).
			Gap(t.Space(2)).Radius(t.Space(5)).MinWidth(0).MaxWidthPercent(100)
		search.Children(func() {
			components.Icon(c, lucide.Search).Size(t.Space(4), t.Space(4)).Shrink(0)
			components.Text(c, tr(c, "welcomeSearchEverything")).Wrap()
			if !compact {
				components.Text(c, shortcut).FontSize(t.Rem(.75)).Opacity(.7).Shrink(0).SingleLine()
			}
		})
		if search.Clicked() {
			g.openCommands()
		}
	})
	welcomeEnter(c, ui.Row(c).Wrap().Justify(ui.Center).Gap(t.Space(2)), 3).Children(func() {
		for _, badge := range []string{msg(c).VersionS(galleryVersion), "MIT", "Go", "MyGo"} {
			components.Badge(c, badge).Background(t.Surface.Alpha(.8)).TextColor(t.TextMuted)
		}
	})
}

// welcomeContentWidth is the window width beside the expanded sidebar.
func (g *Gallery) welcomeContentWidth(c *ui.Context) float32 {
	width, _ := c.Size()
	if g.Shell != nil && g.Shell.LeftExpanded() {
		width -= g.Shell.LeftWidth()
	}
	return width
}

func (g *Gallery) openRepository() {
	if g.openURL == nil {
		return
	}
	if err := g.openURL(repositoryURL); err != nil {
		log.Printf("open repository: %v", err)
	}
}
