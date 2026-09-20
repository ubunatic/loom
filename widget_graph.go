// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "codeberg.org/ubunatic/loom/graph"

// Gauge renders the latest value from a metric store as a graph bar. When
// Store and Metric are unset, Value is rendered directly. Width is the number
// of graph cells and does not include RenderBar's optional wrappers.
type Gauge struct {
	Store  *MetricStore
	Metric string
	// Name is accepted as a readable alias for Metric.
	Name  string
	Value float64
	Min   float64
	Max   float64
	Width int
	// Options supplies all graph bar options. Width, Min, and Max fields on
	// Gauge override the corresponding options when set.
	Options graph.BarOptions
}

// NewGauge creates a metric-bound gauge using the supplied store and series
// name. The returned value can be further configured before drawing.
func NewGauge(store *MetricStore, metric string, options graph.BarOptions) *Gauge {
	return &Gauge{Store: store, Metric: metric, Options: options, Width: options.Width}
}

// Draw renders the gauge at the top-left of r and clips it to the allocated
// rectangle. It is safe to draw into a zero-sized rectangle.
func (g *Gauge) Draw(c *Canvas, r Rect) {
	if g == nil || c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	value := g.Value
	if g.Store != nil {
		name := g.Metric
		if name == "" {
			name = g.Name
		}
		if sample, ok := g.Store.Latest(name); ok {
			value = sample.Value
		}
	}
	opts := g.Options
	if g.Width != 0 {
		opts.Width = g.Width
	}
	if g.Min != 0 || g.Max != 0 {
		opts.Min, opts.Max = g.Min, g.Max
	}
	text := graph.RenderBar(value, opts)
	c.Write(r.X, r.Y, TruncateText(text, r.W, ""), Style{})
}

// HandleKey makes Gauge a passive widget.
func (*Gauge) HandleKey(KeyEvent) bool { return false }

// HandleMouse makes Gauge a passive widget.
func (*Gauge) HandleMouse(MouseEvent) bool { return false }

// ContentWidth reports the preferred width including RenderBar wrappers.
func (g *Gauge) ContentWidth() int {
	if g == nil {
		return 0
	}
	opts := g.Options
	if g.Width != 0 {
		opts.Width = g.Width
	}
	if opts.Width < 1 {
		opts.Width = graph.MaxWidth
	}
	if opts.NoWrapper {
		return opts.Width
	}
	return opts.Width + 2
}

// ContentHeight reports the one row used by a gauge.
func (*Gauge) ContentHeight() int { return 1 }

// Sparkline renders a metric history using graph.RenderSparkline. Metric and
// Store bind it to a named series; Values can be used for a direct series.
// Width is the number of graph cells.
type Sparkline struct {
	Store  *MetricStore
	Metric string
	// Series is accepted as a readable alias for Metric.
	Series string
	Values []float64
	Width  int
	Min    float64
	Max    float64
	// Options supplies all graph sparkline options. Width, Min, and Max fields
	// on Sparkline override the corresponding options when set.
	Options graph.SparklineOptions
}

// NewSparkline creates a metric-bound sparkline using the supplied store and
// series name. The returned value can be further configured before drawing.
func NewSparkline(store *MetricStore, metric string, options graph.SparklineOptions) *Sparkline {
	return &Sparkline{Store: store, Metric: metric, Options: options, Width: options.Width}
}

// Draw renders the sparkline at the top-left of r and clips it to the
// allocated rectangle.
func (s *Sparkline) Draw(c *Canvas, r Rect) {
	if s == nil || c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	values := append([]float64(nil), s.Values...)
	if s.Store != nil {
		name := s.Metric
		if name == "" {
			name = s.Series
		}
		values = s.Store.Values(name)
	}
	opts := s.Options
	if s.Width != 0 {
		opts.Width = s.Width
	}
	if s.Min != 0 || s.Max != 0 {
		opts.Min, opts.Max, opts.FixedRange = s.Min, s.Max, true
	}
	text := graph.RenderSparkline(values, opts)
	c.Write(r.X, r.Y, TruncateText(text, r.W, ""), Style{})
}

// HandleKey makes Sparkline a passive widget.
func (*Sparkline) HandleKey(KeyEvent) bool { return false }

// HandleMouse makes Sparkline a passive widget.
func (*Sparkline) HandleMouse(MouseEvent) bool { return false }

// ContentWidth reports the configured sparkline width.
func (s *Sparkline) ContentWidth() int {
	if s == nil {
		return 0
	}
	width := s.Width
	if width == 0 {
		width = s.Options.Width
	}
	if width < 1 {
		width = graph.MaxWidth
	}
	return width
}

// ContentHeight reports the one row used by a sparkline.
func (*Sparkline) ContentHeight() int { return 1 }
