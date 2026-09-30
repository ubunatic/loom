// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"testing"
	"time"
)

func TestTimerLifecycleWithInjectedClock(t *testing.T) {
	now := time.Unix(100, 0)
	done := 0
	timer := NewTimer(3 * time.Second)
	timer.Now = func() time.Time { return now }
	timer.OnDone = func() { done++ }
	timer.Start()
	if got := timer.Remaining(); got != 3*time.Second {
		t.Fatalf("remaining after start = %s", got)
	}
	now = now.Add(time.Second)
	timer.Tick(now)
	if got := timer.Remaining(); got != 2*time.Second || done != 0 {
		t.Fatalf("after tick: remaining=%s done=%d", got, done)
	}
	now = now.Add(500 * time.Millisecond)
	timer.Stop()
	if got := timer.Remaining(); got != 1500*time.Millisecond {
		t.Fatalf("remaining after pause = %s", got)
	}
	now = now.Add(5 * time.Second)
	timer.Tick(now)
	// The injected clock advanced by 500 ms before Stop, so that elapsed time
	// must be deducted from the previous two-second tick value.
	if got := timer.Remaining(); got != 1500*time.Millisecond {
		t.Fatalf("stopped remaining = %s", got)
	}
	timer.Start()
	now = now.Add(2 * time.Second)
	timer.Tick(now)
	timer.Tick(now.Add(time.Second))
	if got := timer.Remaining(); got != 0 || done != 1 || timer.TickInterval() != 0 {
		t.Fatalf("completed: remaining=%s done=%d interval=%s", got, done, timer.TickInterval())
	}
	timer.Reset()
	if got := timer.Remaining(); got != 3*time.Second || done != 1 {
		t.Fatalf("after reset: remaining=%s done=%d", got, done)
	}
}

func TestTimerZeroDurationCompletesOnStart(t *testing.T) {
	done := 0
	timer := NewTimer(0)
	timer.OnDone = func() { done++ }
	timer.Start()
	timer.Start()
	if done != 1 || timer.TickInterval() != 0 {
		t.Fatalf("done=%d interval=%s", done, timer.TickInterval())
	}
}

func TestStopwatchLifecycleWithInjectedClock(t *testing.T) {
	now := time.Unix(200, 0)
	watch := NewStopwatch()
	watch.Now = func() time.Time { return now }
	watch.Start()
	now = now.Add(1500 * time.Millisecond)
	watch.Tick(now)
	if got := watch.Elapsed(); got != 1500*time.Millisecond {
		t.Fatalf("elapsed = %s", got)
	}
	watch.Stop()
	now = now.Add(10 * time.Second)
	watch.Tick(now)
	if got := watch.Elapsed(); got != 1500*time.Millisecond || watch.TickInterval() != 0 {
		t.Fatalf("stopped: elapsed=%s interval=%s", got, watch.TickInterval())
	}
	watch.Start()
	now = now.Add(500 * time.Millisecond)
	watch.Tick(now)
	if got := watch.Elapsed(); got != 2*time.Second {
		t.Fatalf("resumed elapsed = %s", got)
	}
	watch.Reset()
	if got := watch.Elapsed(); got != 0 || watch.TickInterval() != 0 {
		t.Fatalf("reset: elapsed=%s interval=%s", got, watch.TickInterval())
	}
}

func TestTimerAndStopwatchFormatAndRender(t *testing.T) {
	timer := NewTimer(65 * time.Second)
	timer.Formatter = func(d time.Duration) string { return "left " + formatMinutesSeconds(d) }
	watch := NewStopwatch()
	watch.Now = func() time.Time { return time.Unix(0, 0) }
	watch.Start()
	watch.Tick(time.Unix(61, 0))
	watch.Formatter = formatMinutesSeconds
	if got := timer.text(); got != "left 01:05" {
		t.Fatalf("timer text = %q", got)
	}
	if got := watch.text(); got != "01:01" {
		t.Fatalf("stopwatch text = %q", got)
	}
	for _, widget := range []Widget{timer, watch} {
		if got := Render(widget, 20, 1)[0]; got == "" {
			t.Fatalf("%T rendered empty text", widget)
		}
	}
}

func formatMinutesSeconds(d time.Duration) string {
	seconds := int(d / time.Second)
	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}
