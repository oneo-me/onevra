package chart

import (
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	base "github.com/egoist/mygo/ui/chart"
)

const dataProgress = "onevra.chart.data.progress"

// Snapshots belong to the chart element; MyGo owns the clock, easing and frame
// scheduling. One shared progress keeps stacked edges and pie boundaries joined.
type dataMotion struct {
	kind         kind
	columns      int
	keys         []string
	from, target []float64
}

func animateValues(c *ui.Context, e ui.Element, kind kind, columns int, keys []string, target []float64, opts Options) []float64 {
	st := ui.Local(e, "onevra.chart.data", func() dataMotion { return dataMotion{} })
	duration := opts.AnimationDuration
	if duration == 0 {
		duration = 300 * time.Millisecond
	}
	compatible := st.kind == kind && st.columns == columns && slices.Equal(st.keys, keys) && len(st.target) == len(target)
	if compatible {
		for i, value := range target {
			if finite(value) != finite(st.target[i]) {
				compatible = false
				break
			}
		}
	}
	if !compatible || duration < 0 || c.Preferences().ReduceMotion || c.Instant() {
		st.kind, st.columns, st.keys = kind, columns, slices.Clone(keys)
		st.from, st.target = slices.Clone(target), slices.Clone(target)
		e.Jump(dataProgress, 1)
		return target
	}
	changed := !slices.EqualFunc(target, st.target, func(a, b float64) bool { return a == b || (math.IsNaN(a) && math.IsNaN(b)) })
	if !changed && e.Settled(dataProgress) {
		return target
	}
	progress := e.Animate(dataProgress, 1, duration)
	current := interpolateValues(st.from, st.target, progress)
	if changed {
		// Sample the ongoing transition before adopting another update, including
		// caller-owned slices mutated in place. Retain no application data aliases.
		st.from, st.target = current, slices.Clone(target)
		e.Jump(dataProgress, 0)
		progress = e.Animate(dataProgress, 1, duration)
		current = interpolateValues(st.from, st.target, progress)
	}
	return current
}

func interpolateValues(from, to []float64, progress float32) []float64 {
	values := make([]float64, len(to))
	for i, value := range to {
		switch {
		case progress >= 1 || !finite(value):
			values[i] = value
		case progress <= 0:
			values[i] = from[i]
		default:
			// A convex combination avoids overflowing (to - from) for large data.
			values[i] = from[i]*(1-float64(progress)) + value*float64(progress)
		}
	}
	return values
}

func animatePrepared(c *ui.Context, e ui.Element, data Data, opts Options, kind kind, p prepared) prepared {
	if !p.valid {
		animateValues(c, e, kind, 0, nil, nil, opts)
		return p
	}
	keys := slices.Clone(data.Labels)
	values := make([]float64, 0, 2+len(data.Labels)*len(p.series))
	values = append(values, p.y.Lo, p.y.Hi)
	for k, j := range p.series {
		keys = append(keys, data.Series[j].ID)
		values = append(values, p.values[k]...)
	}
	values = animateValues(c, e, kind, len(data.Labels), keys, values, opts)
	if e.Settled(dataProgress) {
		return p
	}
	p.y.Lo, p.y.Hi = values[0], values[1]
	for k := range p.values {
		p.values[k] = values[2+k*len(data.Labels) : 2+(k+1)*len(data.Labels)]
	}
	if kind == stacks {
		p.restack()
	}
	return p
}

func animateSlices(c *ui.Context, e ui.Element, data []Slice, indices []int, arcs []base.Slice, opts Options) []base.Slice {
	keys := make([]string, len(indices))
	boundaries := make([]float64, len(arcs))
	for j, i := range indices {
		keys[j], boundaries[j] = data[i].ID, arcs[j].End
	}
	boundaries = animateValues(c, e, donuts, 0, keys, boundaries, opts)
	start := 0.0
	for i := range arcs {
		arcs[i].Start, arcs[i].End = start, boundaries[i]
		start = boundaries[i]
	}
	return arcs
}
