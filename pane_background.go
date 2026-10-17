// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"time"
)

// AnimatedBackground is a background that needs periodic redraws.
type AnimatedBackground interface {
	Background
	DrawBackgroundAt(*Canvas, Rect, time.Time)
}

// backgroundTicker owns the optional timer that drives animated backgrounds.
// Reconcile avoids timer work when the requested interval is unchanged.
type backgroundTicker struct {
	ticker   *time.Ticker
	frames   <-chan time.Time
	interval time.Duration
}

func (b *backgroundTicker) reconcile(background Background, reduceMotion, onRedraw bool) bool {
	interval := time.Duration(0)
	if !reduceMotion && !onRedraw {
		if _, ok := background.(AnimatedBackground); ok {
			interval = SpeccedBackground.RedrawInterval
			if cadence, ok := background.(BackgroundCadence); ok {
				interval = cadence.BackgroundInterval()
			}
			if interval < 0 {
				interval = 0
			}
		}
	}
	if interval == b.interval {
		return false
	}
	if b.ticker != nil {
		b.ticker.Stop()
		b.ticker = nil
		b.frames = nil
	}
	b.interval = interval
	if interval > 0 {
		b.ticker = time.NewTicker(interval)
		b.frames = b.ticker.C
	}
	return true
}

func (b *backgroundTicker) stop() {
	if b.ticker != nil {
		b.ticker.Stop()
		b.ticker = nil
		b.frames = nil
	}
}

// BackgroundCadence optionally lets an animated background choose its redraw
// interval. Implementations that do not provide it use the spec default.
type BackgroundCadence interface {
	BackgroundInterval() time.Duration
}
