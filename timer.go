// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const clockWidgetInterval = time.Second

// DurationFormatter formats the remaining or elapsed duration shown by a
// Timer or Stopwatch. A nil formatter uses mm:ss, growing to h:mm:ss as needed.
type DurationFormatter func(time.Duration) string

// Timer displays a countdown. Start, Stop and Reset control its Pane-driven
// updates; OnDone is called once when a running timer reaches zero.
type Timer struct {
	// Controls enables visible start/stop/reset buttons, Space to toggle, and R to reset.
	Controls  bool
	Duration  time.Duration
	Formatter DurationFormatter
	OnDone    func()
	Now       func() time.Time
	Style     Style

	mu         sync.Mutex
	remaining  time.Duration
	last       time.Time
	started    bool
	done       bool
	invalidate func()
	resetTick  func()
}

// NewTimer constructs a stopped countdown with duration d.
func NewTimer(d time.Duration) *Timer {
	return &Timer{Duration: nonnegativeDuration(d), remaining: nonnegativeDuration(d), Now: time.Now}
}

// Start starts or resumes the countdown and requests a redraw.
func (t *Timer) Start() {
	t.mu.Lock()
	if t.started || t.done {
		t.mu.Unlock()
		return
	}
	if t.remaining <= 0 {
		t.done = true
		callback, invalidate := t.OnDone, t.invalidate
		t.mu.Unlock()
		if invalidate != nil {
			invalidate()
		}
		if callback != nil {
			callback()
		}
		return
	}
	t.started = true
	t.last = t.now()
	reset, invalidate := t.resetTick, t.invalidate
	t.mu.Unlock()
	if reset != nil {
		reset()
	}
	if invalidate != nil {
		invalidate()
	}
}

// Stop pauses the countdown.
func (t *Timer) Stop() {
	t.mu.Lock()
	if !t.started {
		t.mu.Unlock()
		return
	}
	t.remaining = nonnegativeDuration(t.remaining - t.now().Sub(t.last))
	t.started = false
	reset, invalidate := t.resetTick, t.invalidate
	t.mu.Unlock()
	if reset != nil {
		reset()
	}
	if invalidate != nil {
		invalidate()
	}
}

// Reset stops the timer and restores its original duration.
func (t *Timer) Reset() {
	t.mu.Lock()
	t.started, t.done = false, false
	t.remaining = nonnegativeDuration(t.Duration)
	reset, invalidate := t.resetTick, t.invalidate
	t.mu.Unlock()
	if reset != nil {
		reset()
	}
	if invalidate != nil {
		invalidate()
	}
}

// Remaining reports the time remaining, clamped at zero.
func (t *Timer) Remaining() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.started {
		return t.remaining
	}
	return nonnegativeDuration(t.remaining - t.now().Sub(t.last))
}

func (t *Timer) now() time.Time {
	if t.Now != nil {
		return t.Now()
	}
	return time.Now()
}

// TickInterval implements Ticker. The countdown updates once per displayed second.
func (t *Timer) TickInterval() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.started || t.done {
		return 0
	}
	return clockWidgetInterval
}

// Tick updates the countdown and invokes OnDone once on completion.
func (t *Timer) Tick(now time.Time) {
	t.mu.Lock()
	if !t.started || t.done {
		t.mu.Unlock()
		return
	}
	elapsed := now.Sub(t.last)
	if elapsed < 0 {
		elapsed = 0
	}
	t.remaining = nonnegativeDuration(t.remaining - elapsed)
	t.last = now
	finished := t.remaining == 0
	if finished {
		t.started, t.done = false, true
	}
	callback, invalidate, reset := t.OnDone, t.invalidate, t.resetTick
	t.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
	if finished && reset != nil {
		reset()
	}
	if finished && callback != nil {
		callback()
	}
}

// SetInvalidate implements InvalidationAware.
func (t *Timer) SetInvalidate(fn func()) { t.mu.Lock(); t.invalidate = fn; t.mu.Unlock() }

// SetResetTick implements TickerControlAware.
func (t *Timer) SetResetTick(fn func()) { t.mu.Lock(); t.resetTick = fn; t.mu.Unlock() }

func (t *Timer) text() string {
	d := t.Remaining()
	if t.Formatter != nil {
		return t.Formatter(d)
	}
	return formatDuration(d)
}

func (t *Timer) Draw(c *Canvas, r Rect) {
	if r.W > 0 && r.H > 0 {
		c.Write(r.X, r.Y, TruncateText(t.text(), r.W, ""), t.Style)
		if t.Controls && r.H > 1 {
			c.Write(r.X, r.Y+1, TruncateText(clockControlText(), r.W, ""), t.Style)
		}
	}
}
func (t *Timer) ConsumeKey(e KeyEvent) EventResult {
	if !t.Controls {
		return Ignored()
	}
	if e.Is("space") || e.Text == " " {
		if t.TickInterval() > 0 {
			t.Stop()
		} else {
			t.Start()
		}
		return Handled()
	}
	if e.Is("r") || e.Text == "R" {
		t.Reset()
		return Handled()
	}
	return Ignored()
}
func (t *Timer) ConsumeMouse(e MouseEvent) EventResult {
	if !t.Controls {
		return Ignored()
	}
	return consumeClockMouse(e, t.Start, t.Stop, t.Reset)
}
func (t *Timer) ContentWidth() int {
	if t.Controls {
		return max(StringWidth(t.text()), StringWidth(clockControlText()))
	}
	return StringWidth(t.text())
}
func (t *Timer) ContentHeight() int {
	if t.Controls {
		return 2
	}
	return 1
}

func clockControlLabels() []string {
	d := SpeccedDefaults.Clock
	return []string{"[" + d.Start + "]", "[" + d.Stop + "]", "[" + d.Reset + "]"}
}
func clockControlText() string { return strings.Join(clockControlLabels(), " ") }
func consumeClockMouse(e MouseEvent, start, stop, reset func()) EventResult {
	if e.Action != MousePress || e.Button != MouseLeft || e.Y != 1 {
		return Ignored()
	}
	x := 0
	for i, label := range clockControlLabels() {
		width := StringWidth(label)
		if e.X >= x && e.X < x+width {
			[]func(){start, stop, reset}[i]()
			return Handled()
		}
		x += width + 1
	}
	return Ignored()
}

// Stopwatch displays elapsed time and can be paused and resumed.
type Stopwatch struct {
	Formatter DurationFormatter
	// Controls enables Space to start/stop and R to reset.
	Controls bool
	Now      func() time.Time
	Style    Style

	mu         sync.Mutex
	elapsed    time.Duration
	last       time.Time
	started    bool
	invalidate func()
	resetTick  func()
}

// NewStopwatch constructs a stopped stopwatch.
func NewStopwatch() *Stopwatch { return &Stopwatch{Now: time.Now} }

// Start starts or resumes the stopwatch.
func (s *Stopwatch) Start() {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started, s.last = true, s.now()
	reset, invalidate := s.resetTick, s.invalidate
	s.mu.Unlock()
	if reset != nil {
		reset()
	}
	if invalidate != nil {
		invalidate()
	}
}

// Stop pauses the stopwatch.
func (s *Stopwatch) Stop() {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return
	}
	s.elapsed += nonnegativeDuration(s.now().Sub(s.last))
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

// Reset stops the stopwatch and clears its elapsed duration.
func (s *Stopwatch) Reset() {
	s.mu.Lock()
	s.elapsed, s.started = 0, false
	reset, invalidate := s.resetTick, s.invalidate
	s.mu.Unlock()
	if reset != nil {
		reset()
	}
	if invalidate != nil {
		invalidate()
	}
}

// Elapsed reports the accumulated running time.
func (s *Stopwatch) Elapsed() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return s.elapsed
	}
	return s.elapsed + nonnegativeDuration(s.now().Sub(s.last))
}

func (s *Stopwatch) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}
func (s *Stopwatch) TickInterval() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return 0
	}
	return clockWidgetInterval
}
func (s *Stopwatch) Tick(now time.Time) {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return
	}
	s.elapsed += nonnegativeDuration(now.Sub(s.last))
	s.last = now
	invalidate := s.invalidate
	s.mu.Unlock()
	if invalidate != nil {
		invalidate()
	}
}
func (s *Stopwatch) SetInvalidate(fn func()) { s.mu.Lock(); s.invalidate = fn; s.mu.Unlock() }
func (s *Stopwatch) SetResetTick(fn func())  { s.mu.Lock(); s.resetTick = fn; s.mu.Unlock() }
func (s *Stopwatch) text() string {
	d := s.Elapsed()
	if s.Formatter != nil {
		return s.Formatter(d)
	}
	return formatDuration(d)
}
func (s *Stopwatch) Draw(c *Canvas, r Rect) {
	if r.W > 0 && r.H > 0 {
		c.Write(r.X, r.Y, TruncateText(s.text(), r.W, ""), s.Style)
	}
}
func (s *Stopwatch) ConsumeKey(e KeyEvent) EventResult {
	if !s.Controls {
		return Ignored()
	}
	switch e.Key {
	case "space":
		if s.TickInterval() > 0 {
			s.Stop()
		} else {
			s.Start()
		}
		return Handled()
	case "r":
		s.Reset()
		return Handled()
	}
	if e.Text == " " {
		if s.TickInterval() > 0 {
			s.Stop()
		} else {
			s.Start()
		}
		return Handled()
	}
	if e.Text == "r" || e.Text == "R" {
		s.Reset()
		return Handled()
	}
	return Ignored()
}
func (*Stopwatch) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
func (s *Stopwatch) ContentWidth() int                 { return StringWidth(s.text()) }
func (*Stopwatch) ContentHeight() int                  { return 1 }

func formatDuration(d time.Duration) string {
	seconds := int64(nonnegativeDuration(d) / time.Second)
	if hours := seconds / 3600; hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, seconds/60%60, seconds%60)
	}
	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}

func nonnegativeDuration(d time.Duration) time.Duration {
	if d < 0 {
		return 0
	}
	return d
}
