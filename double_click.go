// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "time"

// DoubleClickRecognizer recognizes two presses on the same target and button
// within the specced time and movement thresholds. Call Handle for each mouse
// event, including releases and motion, with the identity of the item under the
// pointer. An empty target clears a pending click.
type DoubleClickRecognizer struct {
	interval  time.Duration
	tolerance int
	now       func() time.Time
	pending   bool
	button    MouseButton
	target    string
	x, y      int
	at        time.Time
}

// NewDoubleClickRecognizer creates a recognizer. Pass a clock for deterministic
// tests; nil uses the current wall clock.
func NewDoubleClickRecognizer(now func() time.Time) *DoubleClickRecognizer {
	if now == nil {
		now = time.Now
	}
	return &DoubleClickRecognizer{
		interval:  SpeccedDefaults.Mouse.DoubleClickInterval,
		tolerance: SpeccedDefaults.Mouse.MovementTolerance,
		now:       now,
	}
}

// Handle consumes an input event for target and reports whether it completed a
// double-click. Only presses with a real button and non-empty target can start
// or complete a click pair.
func (r *DoubleClickRecognizer) Handle(e MouseEvent, target string) bool {
	if r == nil {
		return false
	}
	now := r.now()
	if !r.pending {
		return r.start(e, target, now)
	}

	if now.Before(r.at) || now.Sub(r.at) > r.interval || target == "" || target != r.target ||
		!withinMovementTolerance(e.X, e.Y, r.x, r.y, r.tolerance) {
		r.pending = false
		return r.start(e, target, now)
	}
	if e.Action != MousePress {
		return false
	}
	if e.Button != r.button || !validClickButton(e.Button) {
		return r.start(e, target, now)
	}
	r.pending = false
	return true
}

func (r *DoubleClickRecognizer) start(e MouseEvent, target string, now time.Time) bool {
	if e.Action != MousePress || !validClickButton(e.Button) || target == "" {
		return false
	}
	r.pending = true
	r.button = e.Button
	r.target = target
	r.x, r.y = e.X, e.Y
	r.at = now
	return false
}

func validClickButton(button MouseButton) bool {
	return button == MouseLeft || button == MouseMiddle || button == MouseRight
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func withinMovementTolerance(x, y, originX, originY, tolerance int) bool {
	return absInt(x-originX) <= tolerance && absInt(y-originY) <= tolerance
}
