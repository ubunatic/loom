// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"sync"
	"time"

	"codeberg.org/ubunatic/loom/graph"
)

// Spinner shows a braille activity indicator and an optional label. It starts
// stopped; Start and Stop control its Pane-driven animation.
type Spinner struct {
	Label    string
	Interval time.Duration
	Style    Style

	mu         sync.Mutex
	frame      int
	started    bool
	invalidate func()
	resetTick  func()
}

// NewSpinner creates a stopped spinner. The optional label appears after the
// animated glyph; the default frame cadence comes from spec/defaults.yaml.
func NewSpinner(label ...string) *Spinner {
	s := &Spinner{Interval: SpeccedDefaults.Spinner.TickInterval}
	if len(label) > 0 {
		s.Label = label[0]
	}
	return s
}

// Start enables animation and requests an immediate redraw.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started = true
	reset, invalidate := s.resetTick, s.invalidate
	s.mu.Unlock()
	if reset != nil {
		reset()
	}
	if invalidate != nil {
		invalidate()
	}
}

// Stop pauses animation and requests an immediate redraw.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return
	}
	s.started = false
	reset, invalidate := s.resetTick, s.invalidate
	s.mu.Unlock()
	if reset != nil {
		reset()
	}
	if invalidate != nil {
		invalidate()
	}
}

// TickInterval implements Ticker. A stopped spinner does not keep the Pane
// ticking; a nonpositive configured interval falls back to the spec default.
func (s *Spinner) TickInterval() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return 0
	}
	if s.Interval > 0 {
		return s.Interval
	}
	return SpeccedDefaults.Spinner.TickInterval
}

// Tick advances one animation frame when the spinner is running.
func (s *Spinner) Tick(time.Time) {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return
	}
	s.frame = (s.frame + 1) % len(graph.DefaultSpinnerFrames)
	invalidate := s.invalidate
	s.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}

// Frame reports the current spinner frame index.
func (s *Spinner) Frame() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.frame
}

// SetInvalidate implements InvalidationAware.
func (s *Spinner) SetInvalidate(fn func()) {
	s.mu.Lock()
	s.invalidate = fn
	s.mu.Unlock()
}

// SetResetTick implements TickerControlAware.
func (s *Spinner) SetResetTick(fn func()) {
	s.mu.Lock()
	s.resetTick = fn
	s.mu.Unlock()
}

func (s *Spinner) text() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	glyph := graph.SpinnerGlyph(s.frame)
	if s.Label == "" {
		return string(glyph)
	}
	return string(glyph) + " " + s.Label
}

// ContentWidth implements ContentWidther.
func (s *Spinner) ContentWidth() int { return StringWidth(s.text()) }

// ContentHeight implements ContentHeighter.
func (s *Spinner) ContentHeight() int { return 1 }

// Draw implements Widget.
func (s *Spinner) Draw(c *Canvas, r Rect) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	text := s.text()
	c.Write(r.X, r.Y, TruncateText(text, r.W, ""), s.Style)
}

// HandleKey implements Widget; the spinner takes no input.
func (*Spinner) HandleKey(KeyEvent) bool { return false }

// HandleMouse implements Widget; the spinner takes no input.
func (*Spinner) HandleMouse(MouseEvent) bool { return false }
