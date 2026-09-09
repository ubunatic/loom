// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"context"
	"fmt"
	"time"
)

// Cadence configures independent collection and rendering intervals.
// Callbacks execute serially: drawing must only read the latest collected state.
// Collection should be bounded; slow external I/O belongs outside this loop.
type Cadence struct {
	Collect time.Duration
	Redraw  time.Duration
}

// Validate rejects intervals that cannot drive a ticker.
func (c Cadence) Validate() error {
	if c.Collect <= 0 || c.Redraw <= 0 {
		return fmt.Errorf("cadence: collection and redraw intervals must be positive")
	}
	return nil
}

// Run collects and draws initially, then runs each callback at its own cadence.
// Cancellation and callback errors stop both timers before returning. Missed
// ticks coalesce; redraws never synthesize samples or advance collected history.
func (c Cadence) Run(ctx context.Context, collect func(time.Time) error, draw func() error) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if collect == nil || draw == nil {
		return fmt.Errorf("cadence: collection and draw callbacks required")
	}
	samples := time.NewTicker(c.Collect)
	defer samples.Stop()
	frames := time.NewTicker(c.Redraw)
	defer frames.Stop()
	return runCadence(ctx, time.Now(), samples.C, frames.C, collect, draw)
}

func runCadence(ctx context.Context, initial time.Time, samples, frames <-chan time.Time, collect func(time.Time) error, draw func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := collect(initial); err != nil {
		return err
	}
	if err := draw(); err != nil {
		return err
	}
	for {
		// Give cancellation priority over another ready tick.
		if err := ctx.Err(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now, ok := <-samples:
			if !ok {
				return fmt.Errorf("cadence: sample clock closed")
			}
			if err := collect(now); err != nil {
				return err
			}
		case _, ok := <-frames:
			if !ok {
				return fmt.Errorf("cadence: redraw clock closed")
			}
			if err := draw(); err != nil {
				return err
			}
		}
	}
}
