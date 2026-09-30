// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"sync"
	"time"

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
	// Indeterminate animates a one-third-width pulse when enabled.
	Indeterminate bool
	// ShowPercent appends a percentage label.
	ShowPercent bool
	// ShowCount appends the current value and total, followed by Unit.
	ShowCount bool
	// Unit labels the values in the count label.
	Unit string
	// StyleFill and StyleEmpty style filled and empty cells independently.
	StyleFill  Style
	StyleEmpty Style
	// Interval controls indeterminate animation cadence. Zero uses the default.
	Interval time.Duration
	// Total is the determinate maximum. Zero uses 100 for compatibility.
	Total float64

	mu         sync.Mutex
	value      float64
	step       int
	done       bool
	invalidate func()
	resetTick  func()
	frame      int
}

// NewProgressBar returns an empty bar with specced defaults and sub-character fill.
func NewProgressBar() *ProgressBar {
	return &ProgressBar{
		Options: graph.BracketedBarOptions{
			Width:   SpeccedDefaults.ProgressBar.Width,
			SubChar: true,
		},
		Total: 100,
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
	if value < 0 || value != value {
		value = 0
	}
	total := p.Total
	if total <= 0 {
		total = 100
	}
	if value > total {
		value = total
	}
	p.value = value
	opts := p.Options
	if opts.Width <= 0 {
		opts.Width = SpeccedDefaults.ProgressBar.Width
	}
	opts.Pattern = ""
	step := graph.BracketedBarStep(value*100/total, opts)
	changed := step != p.step
	p.step = step
	fn := p.invalidate
	p.mu.Unlock()
	if changed && fn != nil {
		fn()
	}
}

// TickInterval implements Ticker. Determinate bars do not request ticks.
func (p *ProgressBar) TickInterval() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.Indeterminate {
		return 0
	}
	if p.Interval > 0 {
		return p.Interval
	}
	return 100 * time.Millisecond
}

// Tick advances the marquee pulse and requests a redraw.
func (p *ProgressBar) Tick(time.Time) {
	p.mu.Lock()
	if !p.Indeterminate {
		p.mu.Unlock()
		return
	}
	width := p.Options.Width
	if width <= 0 {
		width = SpeccedDefaults.ProgressBar.Width
	}
	p.frame = (p.frame + 1) % max(1, width*2)
	fn := p.invalidate
	p.mu.Unlock()
	if fn != nil {
		fn()
	}
}

// SetResetTick implements TickerControlAware.
func (p *ProgressBar) SetResetTick(fn func()) { p.mu.Lock(); p.resetTick = fn; p.mu.Unlock() }

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
	total := p.Total
	if total <= 0 {
		total = 100
	}
	p.mu.Unlock()
	indeterminate, frame := p.Indeterminate, p.frame
	showPercent, showCount, unit := p.ShowPercent, p.ShowCount, p.Unit
	fillStyle, emptyStyle, wholeStyle := p.StyleFill, p.StyleEmpty, p.Style
	bar := graph.RenderBracketedBar(value*100/total, opts)
	w := graph.BracketedBarWidth(opts)
	left, right := opts.Left, opts.Right
	if left == "" {
		left = "["
	}
	if right == "" {
		right = "]"
	}
	pulseStart, pulse := 0, 0
	if indeterminate {
		body := make([]rune, opts.Width)
		for i := range body {
			body[i] = ' '
		}
		pulse := max(1, opts.Width/3)
		fill := opts.Fill
		if fill == 0 {
			fill = '⣿'
		}
		position := frame
		if cycle := opts.Width*2 - pulse; cycle > 0 {
			position %= cycle
			if position >= cycle/2 {
				position = cycle - position
			}
		}
		pulseStart = position
		for i := position; i < position+pulse && i < len(body); i++ {
			body[i] = fill
		}
		bar = left + string(body) + right
	}
	label := ""
	if showPercent {
		label = fmt.Sprintf(" %3.0f%%", value*100/total)
	}
	if showCount {
		label += fmt.Sprintf(" %.0f/%.0f%s", value, total, unit)
	}
	bar += label
	if w > r.W {
		bar = TruncateText(bar, r.W, "")
	}
	x := r.X + layout.AlignOffset(r.W, min(w+len([]rune(label)), r.W), p.Align)
	if fillStyle == (Style{}) && emptyStyle == (Style{}) {
		c.Write(x, r.Y, bar, wholeStyle)
		return
	}
	for i, ch := range []rune(bar) {
		style := wholeStyle
		if i >= len([]rune(left)) && i < len([]rune(left))+opts.Width {
			if indeterminate {
				bodyIndex := i - len([]rune(left))
				if bodyIndex >= pulseStart && bodyIndex < pulseStart+pulse {
					style = fillStyle
				} else {
					style = emptyStyle
				}
			} else if ch == opts.Empty || ch == ' ' {
				style = emptyStyle
			} else {
				style = fillStyle
			}
		}
		c.Set(x+i, r.Y, Cell{Text: string(ch), Style: style})
	}
}

// ApplyTheme implements Themeable.
func (p *ProgressBar) ApplyTheme(theme ThemeColors) {
	p.Style = Style{FG: theme.StatusFG.Color(), BG: theme.StatusBG.Color(), Bold: theme.StatusBold, Dim: theme.StatusDim}
	p.StyleFill = Style{FG: theme.SelectedFG.Color(), BG: theme.SelectedBG.Color(), Bold: theme.SelectedBold}
	p.StyleEmpty = Style{FG: theme.ScrollbarTrackFG.Color(), BG: theme.ScrollbarTrackBG.Color(), Dim: theme.ScrollbarTrackDim}
}

// HandleKey implements Widget; the bar takes no input.
func (p *ProgressBar) HandleKey(KeyEvent) bool { return false }

// HandleMouse implements Widget; the bar takes no input.
func (p *ProgressBar) HandleMouse(MouseEvent) bool { return false }
