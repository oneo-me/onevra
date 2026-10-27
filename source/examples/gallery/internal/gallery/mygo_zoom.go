package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// zoomLevels are the steps the page zooms the window through, as browsers'.
var zoomLevels = []float32{0.5, 0.67, 0.75, 0.8, 0.9, 1, 1.1, 1.25, 1.5, 1.75, 2}

// zoomOne is the index of 1 in zoomLevels.
const zoomOne = 5

// exampleZoom zooms the whole Gallery window, sharp at every size, as a
// browser zooms a page. The zoom stays when another page shows, until reset.
func exampleZoom(c *ui.Context, s *demoState) {
	t := c.Theme()
	level := zoomOne + s.count
	ui.Row(c).Gap(t.Space(2)).AlignItems(ui.Center).Children(func() {
		if components.Button(c, tr(c, "zoomOut"), components.ButtonOutline).Disabled(level == 0).Clicked() {
			s.count--
		}
		if components.Button(c, tr(c, "zoomActualSize"), components.ButtonOutline).Disabled(level == zoomOne).Clicked() {
			s.count = 0
		}
		if components.Button(c, tr(c, "zoomIn"), components.ButtonOutline).Disabled(level == len(zoomLevels)-1).Clicked() {
			s.count++
		}
	})
	z := zoomLevels[zoomOne+s.count]
	c.SetZoom(z)
	components.Text(c, msg(c).Percent(float64(z)*100))
}
