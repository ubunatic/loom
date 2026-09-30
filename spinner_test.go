// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
	"time"
)

func TestSpinnerAdvancesFramesAndStopsTicking(t *testing.T) {
	spinner := NewSpinner("Working")
	if spinner.TickInterval() != 0 {
		t.Fatalf("new spinner interval = %s, want stopped", spinner.TickInterval())
	}
	spinner.Start()
	interval := SpeccedDefaults.Spinner.TickInterval
	if spinner.TickInterval() != interval {
		t.Fatalf("started interval = %s, want %s", spinner.TickInterval(), interval)
	}
	spinner.Tick(time.Now())
	if got := spinner.Frame(); got != 1 {
		t.Fatalf("frame after tick = %d, want 1", got)
	}
	spinner.Stop()
	spinner.Tick(time.Now())
	if spinner.TickInterval() != 0 || spinner.Frame() != 1 {
		t.Fatalf("stopped spinner interval/frame = %s/%d, want 0/1", spinner.TickInterval(), spinner.Frame())
	}
}

func TestSpinnerStartStopResetsPaneTickAndInvalidates(t *testing.T) {
	spinner := NewSpinner()
	reset, invalidated := 0, 0
	spinner.SetResetTick(func() { reset++ })
	spinner.SetInvalidate(func() { invalidated++ })
	spinner.Start()
	spinner.Start()
	spinner.Stop()
	spinner.Stop()
	if reset != 2 || invalidated != 2 {
		t.Fatalf("start/stop callbacks = reset %d invalidate %d, want 2/2", reset, invalidated)
	}
}

func TestSpinnerDrawIncludesOptionalLabelAndClips(t *testing.T) {
	spinner := NewSpinner("Loading")
	canvas := NewCanvas(5, 1)
	spinner.Draw(canvas, Rect{W: 5, H: 1})
	var got strings.Builder
	for x := 0; x < 5; x++ {
		got.WriteString(canvas.Get(x, 0).Text)
	}
	if got.String() != "⠋ Loa" {
		t.Fatalf("spinner drawing = %q, want %q", got.String(), "⠋ Loa")
	}
	if spinner.ContentWidth() != StringWidth("⠋ Loading") || spinner.ContentHeight() != 1 {
		t.Fatalf("content size = %dx%d", spinner.ContentWidth(), spinner.ContentHeight())
	}
}
