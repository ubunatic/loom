// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"sync"

	"codeberg.org/ubunatic/loom/graph"
	"codeberg.org/ubunatic/loom/layout"
)

// ProgressBar is a determinate bracketed progress bar. Producers drive it by
// calling Set (or passing bar.Set as a progress callback), from any goroutine.
// Set requests a host redraw only when the visible bar changes, so high-rate
// callbacks cost one comparison each. Done swaps the bar to a repeating
// pattern (default ":"), the same completion look as SplashView.
type ProgressBar struct {
	// Options configures glyphs, brackets and width. Width zero uses the
	// specced default. Pattern and ANSI are ignored: Done manages the
	// pattern and Style provides colors.
	Options graph.BracketedBarOptions
	// DonePattern replaces the bar after Done. Empty uses the specced default.
	DonePattern string
	// Style is applied to the whole bar, brackets included.
	Style Style
	// Align positions the bar horizontally inside its rect.
	Align layout.Align

	mu         sync.Mutex
	value      float64
	step       int
	done       bool
	invalidate func()
}

// NewProgressBar returns an empty bar with specced defaults and sub-character fill.
func NewProgressBar() *ProgressBar {
	return &ProgressBar{
		Options: graph.BracketedBarOptions{
			Width:   SpeccedDefaults.ProgressBar.Width,
			SubChar: true,
		},
	}
}

// SetInvalidate implements InvalidationAware.
func (p *ProgressBar) SetInvalidate(fn func()) {
	p.mu.Lock()
	p.invalidate = fn
	p.mu.Unlock()
}

// Set updates the value (0 to 100, clamped). It requests a redraw only when
// the number of visible fill subunits changes.
func (p *ProgressBar) Set(value float64) {
	p.mu.Lock()
	p.value = value
	opts := p.Options
	if opts.Width <= 0 {
		opts.Width = SpeccedDefaults.ProgressBar.Width
	}
	opts.Pattern = ""
	step := graph.BracketedBarStep(value, opts)
	changed := step != p.step
	p.step = step
	fn := p.invalidate
	p.mu.Unlock()
	if changed && fn != nil {
		fn()
	}
}

// Value returns the last value passed to Set.
func (p *ProgressBar) Value() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.value
}

// Done marks the bar complete and shows the done pattern.
func (p *ProgressBar) Done() {
	p.mu.Lock()
	changed := !p.done
	p.done = true
	fn := p.invalidate
	p.mu.Unlock()
	if changed && fn != nil {
		fn()
	}
}

// Reset clears the value and the done state.
func (p *ProgressBar) Reset() {
	p.mu.Lock()
	p.value, p.step, p.done = 0, 0, false
	fn := p.invalidate
	p.mu.Unlock()
	if fn != nil {
		fn()
	}
}

// IsDone reports whether Done was called since the last Reset.
func (p *ProgressBar) IsDone() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.done
}

func (p *ProgressBar) renderOptions() graph.BracketedBarOptions {
	opts := p.Options
	if opts.Width <= 0 {
		opts.Width = SpeccedDefaults.ProgressBar.Width
	}
	opts.Pattern = ""
	opts.ANSI = false // colors come from Style; Canvas.Write does not parse SGR
	if p.done {
		opts.Pattern = p.DonePattern
		if opts.Pattern == "" {
			opts.Pattern = SpeccedDefaults.ProgressBar.DonePattern
		}
	}
	return opts
}

// ContentWidth implements ContentWidther.
func (p *ProgressBar) ContentWidth() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return graph.BracketedBarWidth(p.renderOptions())
}

// ContentHeight implements ContentHeighter.
func (p *ProgressBar) ContentHeight() int { return 1 }

// Draw implements Widget.
func (p *ProgressBar) Draw(c *Canvas, r Rect) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	p.mu.Lock()
	opts := p.renderOptions()
	value := p.value
	p.mu.Unlock()
	bar := graph.RenderBracketedBar(value, opts)
	w := graph.BracketedBarWidth(opts)
	if w > r.W {
		bar = TruncateText(bar, r.W, "")
	}
	c.Write(r.X+layout.AlignOffset(r.W, w, p.Align), r.Y, bar, p.Style)
}

// HandleKey implements Widget; the bar takes no input.
func (p *ProgressBar) HandleKey(KeyEvent) bool { return false }

// HandleMouse implements Widget; the bar takes no input.
func (p *ProgressBar) HandleMouse(MouseEvent) bool { return false }
