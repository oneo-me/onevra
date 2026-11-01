// Package chart composes Onevra charts from MyGo's scales, paths and selection.
// Applications own data, localized labels and formatting. Cartesian charts use
// equally spaced categories; use MyGo's primitives for continuous or time axes.
package chart

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// Series holds values aligned with Data.Labels. Missing and non-finite values
// are gaps; excess values are ignored. ID must be nonempty, unique and stable.
// A zero Color chooses a muted theme-aware color by series order.
type Series struct {
	ID, Label string
	Values    []float64
	Color     ui.Color
}

// Data is borrowed for the frame and never modified by a chart.
type Data struct {
	Labels []string
	Series []Series
}

// Slice is a donut or bubble category. IDs have the same requirements as
// Series.ID. Only finite positive values take space in the chart and legend.
type Slice struct {
	ID, Label string
	Value     float64
	Color     ui.Color
}

// State retains interaction between frames. Index is a category (or original
// slice) index, with -1 meaning no selection; its zero value selects the first.
// Hidden records series hidden by Cartesian legends, keyed by Series.ID.
// Keep separate state per chart. Nil state uses element-local storage instead.
type State struct {
	Index  int
	Hidden map[string]bool
}

// Value describes one visible value in a selection.
type Value struct {
	ID, Label string
	Value     float64
	Color     ui.Color
}

// Selection is passed to Detail. Index is -1 when nothing is selected.
type Selection struct {
	Index  int
	Label  string
	Values []Value
}

// Options supplies presentation and application-owned content. PlotHeight
// defaults to 220 DIPs; legends add their own height below the plot.
// FormatValue formats axes, tooltips and accessible values (nil uses numbers).
// Description describes the plot to assistive technology.
// Detail replaces the content of the selection tooltip, which follows the
// pointer over the plot or sits beside a selection the keyboard made; it is
// called only while the tooltip shows. Empty replaces the empty plot,
// including when all series are hidden; legends remain available. With no
// Empty slot, EmptyText is shown (default "—").
type Options struct {
	Label, Description, EmptyText string
	PlotHeight                    float32
	HideLegend                    bool
	FormatValue                   func(float64) string
	Detail                        func(*ui.Context, Selection)
	Empty                         func(*ui.Context)
	// AnimationDuration controls numeric data transitions: zero uses 300 ms,
	// negative disables them. Reduced Motion and
	// ui.Context.DisableAnimations always skip intermediate frames.
	// Initial data, category/series changes and missing-value changes appear
	// immediately. Tooltips and accessible values always report the latest data.
	AnimationDuration time.Duration
}

// Bar draws grouped bars with a zero baseline, including negative values.
func Bar(c *ui.Context, state *State, data Data, opts Options) ui.Element {
	return cartesian(c, state, data, opts, bars)
}

// Line draws monotone lines, breaking paths at missing values.
func Line(c *ui.Context, state *State, data Data, opts Options) ui.Element {
	return cartesian(c, state, data, opts, lines)
}

// Area draws overlapping, translucent areas from zero to each series.
// Use StackedBar for additive totals; areas are not stacked.
func Area(c *ui.Context, state *State, data Data, opts Options) ui.Element {
	return cartesian(c, state, data, opts, areas)
}

// StackedBar stacks positive and negative values separately around zero.
func StackedBar(c *ui.Context, state *State, data Data, opts Options) ui.Element {
	return cartesian(c, state, data, opts, stacks)
}

func frame(c *ui.Context, state *State) (ui.Element, *State) {
	e := ui.Column(c).FillWidth().MinWidth(0).Gap(c.Theme().Space(3))
	if state == nil {
		state = ui.Local(e, "onevra.chart", func() State { return State{Index: -1} })
	}
	return e, state
}

func plotHeight(opts Options) float32 {
	if opts.PlotHeight > 0 {
		return opts.PlotHeight
	}
	return 220
}

func empty(c *ui.Context, opts Options) {
	ui.Column(c).FillWidth().Height(plotHeight(opts)).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
		if opts.Empty != nil {
			opts.Empty(c)
			return
		}
		text := opts.EmptyText
		if text == "" {
			text = "—"
		}
		components.Text(c, text).TextColor(c.Theme().TextMuted)
	})
}

func format(opts Options, v float64) string {
	if opts.FormatValue != nil {
		return opts.FormatValue(v)
	}
	return strconv.FormatFloat(v, 'g', 5, 64)
}

func describe(opts Options, selection Selection) string {
	parts := []string{selection.Label}
	for _, v := range selection.Values {
		parts = append(parts, v.Label+": "+format(opts, v.Value))
	}
	return strings.Join(parts, "; ")
}

func swatch(c *ui.Context, color ui.Color) {
	ui.Box(c).Size(8, 8).Shrink(0).Radius(3).Background(color)
}

func color(c *ui.Context, index int, override ui.Color) ui.Color {
	if override != (ui.Color{}) {
		return override
	}
	t := c.Theme()
	palette := [...]ui.Color{t.Accent, ui.Hex("#6b9483"), ui.Hex("#718eac"), ui.Hex("#a286a5"), ui.Hex("#b09b67"), ui.Hex("#819b9e")}
	value := palette[index%len(palette)]
	if t.Dark {
		return value.Mix(t.Text, .23)
	}
	return value.Mix(t.Background, .24)
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func valueAt(series Series, index int) (float64, bool) {
	if index < 0 || index >= len(series.Values) {
		return 0, false
	}
	v := series.Values[index]
	return v, finite(v)
}

func seriesLegend(c *ui.Context, state *State, data Data) {
	t := c.Theme()
	ui.Row(c).FillWidth().MinWidth(0).Wrap().Gap(t.Space(1)).Children(func() {
		for j, s := range data.Series {
			ui.Box(c).Key(s.ID).MaxWidthPercent(100).Children(func() {
				visible := !state.Hidden[s.ID]
				b := legendButton(c, s.Label, visible)
				b.Children(func() {
					shade := color(c, j, s.Color)
					if !visible {
						shade = t.Border
					}
					swatch(c, shade)
					components.Text(c, s.Label).FontSize(t.FontSize - 1)
				})
				if b.Clicked() {
					if state.Hidden == nil {
						state.Hidden = make(map[string]bool)
					}
					state.Hidden[s.ID] = visible
					c.Invalidate()
				}
			})
		}
	})
}

func legendButton(c *ui.Context, label string, selected bool) ui.Element {
	t := c.Theme()
	b := components.Button(c, "", components.ButtonGhost).Label(label).Selected(selected).
		Padding(t.Space(1), t.Space(2)).Gap(t.Space(1.5)).MaxWidthPercent(100).
		TextColor(t.TextMuted).Background(ui.Color{})
	if selected {
		b.Background(t.Text.Alpha(.045)).TextColor(t.Text)
	}
	return b
}
