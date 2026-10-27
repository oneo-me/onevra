package gallery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/oneo-me/onevra/source/components"
)

// Read the resource once, independently of frames, including its error.
var galleryMarkdown = sync.OnceValues(func() (string, error) {
	dir, err := mygo.App.Path(mygo.PathResources)
	if err != nil {
		return "", err
	}
	return loadGalleryMarkdown(dir)
})

func loadGalleryMarkdown(dir string) (string, error) {
	source, err := os.ReadFile(filepath.Join(dir, "markdown.md"))
	return string(source), err
}

// exampleMarkdown shows a finished document with resource images, gallery-owned
// link activation and an application-built code block.
func exampleMarkdown(c *ui.Context, s *demoState) {
	source, err := galleryMarkdown()
	if err != nil {
		components.Text(c, tr(c, "markdownLoadError")).TextColor(c.Theme().TextMuted)
		return
	}
	if s.markdown.LoadImage == nil {
		// Missing images keep their alt text.
		if dir, err := mygo.App.Path(mygo.PathResources); err == nil {
			s.markdown.LoadImage = markdown.FileImages(dir)
		}
		s.markdown.OnLink = func(string) { s.index++ }
		s.markdown.InlineMath = inlineMath()
	}
	s.markdown.Code = map[string]func(*ui.Context, *markdown.CodeBlockView){
		"counter": func(c *ui.Context, block *markdown.CodeBlockView) {
			ui.Column(c).Gap(c.Theme().Space(2)).Children(func() {
				components.Text(c, strings.TrimSuffix(block.Code(), "\n")).Bold()
				if components.Button(c, tr(c, "increment"), components.ButtonSecondary).Clicked() {
					s.index++
				}
				components.Text(c, msg(c).CountD(s.index))
			})
		},
	}
	markdown.View(c, source, &s.markdown).FillWidth()
}

// inlineMath draws math within text as the application's own renderer would:
// a few TeX commands as their symbols, in italics. Formulas are laid out once
// for each TeX and size, since a paragraph asks again whenever it changes.
func inlineMath() func(tex string, size float32) (markdown.InlineObject, bool) {
	formulas := map[string]markdown.InlineObject{}
	return func(tex string, size float32) (markdown.InlineObject, bool) {
		key := fmt.Sprint(size, tex)
		if f, ok := formulas[key]; ok {
			return f, true
		}
		font := ui.Font{Size: size, Italic: true}
		layout := ui.ShapeText(texSymbols.Replace(tex), font, 0)
		w, _ := layout.Size()
		m := font.Metrics()
		f := markdown.InlineObject{Width: w + size/4, Height: m.Ascent, Depth: m.Descent, Draw: func(p *ui.Painter, r ui.Rect) {
			p.TextLayout(layout, r.X+size/8, r.Y, p.Theme().Accent)
		}}
		formulas[key] = f
		return f, true
	}
}

// texSymbols are the commands inlineMath shows as symbols; other TeX stays
// as written.
var texSymbols = strings.NewReplacer(
	`\pi`, "π", `\alpha`, "α", `\beta`, "β", `\theta`, "θ", `\infty`, "∞", `\sum`, "∑",
	`\cdot`, "·", `\times`, "×", `\leq`, "≤", `\geq`, "≥", `^2`, "²", "{", "", "}", "",
)
