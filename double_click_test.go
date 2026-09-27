// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"testing"
	"time"
)

func TestDoubleClickRecognizerMatchesSameButtonAndTarget(t *testing.T) {
	now := time.Unix(0, 0)
	recognizer := NewDoubleClickRecognizer(func() time.Time { return now })
	click := MouseEvent{Action: MousePress, Button: MouseLeft, X: 10, Y: 4}

	if recognizer.Handle(click, "item-a") {
		t.Fatal("first click was recognized as a double-click")
	}
	now = now.Add(SpeccedDefaults.Mouse.DoubleClickInterval)
	if !recognizer.Handle(click, "item-a") {
		t.Fatal("matching click at the timing boundary was not recognized")
	}
	if recognizer.Handle(click, "item-a") {
		t.Fatal("second click pair was not reset")
	}
}

func TestDoubleClickRecognizerRejectsMovementAndAllowsTolerance(t *testing.T) {
	now := time.Unix(0, 0)
	recognizer := NewDoubleClickRecognizer(func() time.Time { return now })
	click := MouseEvent{Action: MousePress, Button: MouseLeft, X: 10, Y: 4}
	if recognizer.Handle(click, "item") {
		t.Fatal("first click was recognized")
	}

	tolerance := SpeccedDefaults.Mouse.MovementTolerance
	near := click
	near.X += tolerance
	now = now.Add(10 * time.Millisecond)
	if !recognizer.Handle(near, "item") {
		t.Fatal("click within the movement tolerance was not recognized")
	}

	if recognizer.Handle(click, "item") {
		t.Fatal("first click after a completed pair was recognized")
	}
	far := click
	far.X += tolerance + 1
	now = now.Add(10 * time.Millisecond)
	if recognizer.Handle(far, "item") {
		t.Fatal("click outside the movement tolerance was recognized")
	}
	now = now.Add(10 * time.Millisecond)
	if recognizer.Handle(click, "item") {
		t.Fatal("click at the old position completed the moved click pair")
	}
	now = now.Add(10 * time.Millisecond)
	if !recognizer.Handle(click, "item") {
		t.Fatal("matching click after movement reset did not complete the new pair")
	}
}

func TestDoubleClickRecognizerRequiresSameTargetAndButton(t *testing.T) {
	now := time.Unix(0, 0)
	recognizer := NewDoubleClickRecognizer(func() time.Time { return now })
	left := MouseEvent{Action: MousePress, Button: MouseLeft, X: 3, Y: 2}

	if recognizer.Handle(left, "item-a") {
		t.Fatal("first click was recognized")
	}
	now = now.Add(10 * time.Millisecond)
	if recognizer.Handle(left, "item-b") {
		t.Fatal("click on a different target was recognized")
	}
	now = now.Add(10 * time.Millisecond)
	if !recognizer.Handle(left, "item-b") {
		t.Fatal("same-target click after mismatch did not begin a new pair")
	}

	right := left
	right.Button = MouseRight
	if recognizer.Handle(left, "item-b") {
		t.Fatal("first click after a completed pair was recognized")
	}
	now = now.Add(10 * time.Millisecond)
	if recognizer.Handle(right, "item-b") {
		t.Fatal("click with a different button was recognized")
	}
	now = now.Add(10 * time.Millisecond)
	if !recognizer.Handle(right, "item-b") {
		t.Fatal("same-button click after mismatch did not begin a new pair")
	}
}

func TestDoubleClickRecognizerExpiresAndIgnoresNonClicks(t *testing.T) {
	now := time.Unix(0, 0)
	recognizer := NewDoubleClickRecognizer(func() time.Time { return now })
	click := MouseEvent{Action: MousePress, Button: MouseLeft, X: 3, Y: 2}
	if recognizer.Handle(click, "item") {
		t.Fatal("first click was recognized")
	}
	now = now.Add(SpeccedDefaults.Mouse.DoubleClickInterval + time.Nanosecond)
	if recognizer.Handle(click, "item") {
		t.Fatal("click after the timing threshold was recognized")
	}

	motion := click
	motion.Action = MouseDrag
	motion.X++
	if recognizer.Handle(motion, "item") {
		t.Fatal("drag was recognized as a click")
	}
	now = now.Add(10 * time.Millisecond)
	if !recognizer.Handle(click, "item") {
		t.Fatal("click after drag reset did not begin a new pair")
	}
	if recognizer.Handle(click, "") {
		t.Fatal("click without a target was recognized")
	}
	if recognizer.Handle(click, "item") {
		t.Fatal("click after an empty-target event was recognized")
	}
}

func TestDoubleClickDefaultsAreLoadedFromSpec(t *testing.T) {
	if got := SpeccedDefaults.Mouse.DoubleClickInterval; got != 400*time.Millisecond {
		t.Fatalf("double-click interval = %v, want 400ms", got)
	}
	if got := SpeccedDefaults.Mouse.MovementTolerance; got != 1 {
		t.Fatalf("mouse movement tolerance = %d, want 1 cell", got)
	}
}
