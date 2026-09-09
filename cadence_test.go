// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCadenceIndependentTicks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	samples, frames := make(chan time.Time), make(chan time.Time)
	type snapshot struct {
		at      time.Time
		samples int
	}
	collected, drawn := make(chan snapshot), make(chan snapshot)
	done := make(chan error, 1)
	initial := time.Unix(0, 0)
	go func() {
		state := snapshot{}
		done <- runCadence(ctx, initial, samples, frames, func(at time.Time) error {
			state = snapshot{at, state.samples + 1}
			collected <- state
			return nil
		}, func() error { drawn <- state; return nil })
	}()
	first := <-collected
	if got := <-drawn; got != first {
		t.Fatal(got)
	}
	for i := 1; i <= 20; i++ {
		frames <- initial.Add(time.Duration(i) * 50 * time.Millisecond)
		if got := <-drawn; got != first {
			t.Fatalf("redraw mutated snapshot: %+v", got)
		}
	}
	samples <- initial.Add(time.Second)
	next := <-collected
	if next.samples != 2 || !next.at.Equal(initial.Add(time.Second)) {
		t.Fatal(next)
	}
	frames <- initial.Add(time.Second)
	if got := <-drawn; got != next {
		t.Fatal(got)
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation blocked")
	}
}

func TestCadenceFailures(t *testing.T) {
	want := errors.New("callback failed")
	for _, tc := range []struct {
		name                string
		collectErr, drawErr error
	}{
		{"collect", want, nil}, {"draw", nil, want},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := (Cadence{time.Second, time.Millisecond}).Run(context.Background(), func(time.Time) error { return tc.collectErr }, func() error { return tc.drawErr })
			if !errors.Is(err, want) {
				t.Fatal(err)
			}
		})
	}
	for _, c := range []Cadence{{}, {time.Second, 0}, {-time.Second, time.Second}} {
		if err := c.Validate(); err == nil {
			t.Fatalf("accepted %+v", c)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := (Cadence{time.Second, time.Second}).Run(ctx, func(time.Time) error { t.Fatal("collected after cancellation"); return nil }, func() error { t.Fatal("drew after cancellation"); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
