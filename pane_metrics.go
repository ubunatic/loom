// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"time"
)

// RenderMetrics reports recent completed redraw performance. RedrawTime and
// the phase times are the last frame; the Avg and Max fields and BytesPerFrame
// cover the last completed one-second window, which makes them steadier for
// comparing features.
type RenderMetrics struct {
	LoomFPS        float64
	AstraFPS       float64
	AstraTargetFPS float64
	RedrawTime     time.Duration
	DrawTime       time.Duration // clear + widget draw (last frame)
	BackgroundTime time.Duration // background composition (last frame)
	FlushTime      time.Duration // building and writing the frame to the terminal (last frame)
	RedrawAvg      time.Duration
	RedrawMax      time.Duration
	BytesPerFrame  int
	windowStart    time.Time
	loomFrames     int
	astraFrames    int
	windowTotal    time.Duration
	windowMax      time.Duration
	windowBytes    int
}

// framePhases is what one redraw measured.
type framePhases struct {
	total, draw, background, flush time.Duration
	bytes                          int
}

func (m *RenderMetrics) record(now time.Time, astra bool, f framePhases) {
	if m.windowStart.IsZero() {
		m.windowStart = now
	}
	m.loomFrames++
	if astra {
		m.astraFrames++
	}
	if f.total > 0 {
		m.RedrawTime = f.total
	}
	m.DrawTime, m.BackgroundTime, m.FlushTime = f.draw, f.background, f.flush
	m.windowTotal += f.total
	if f.total > m.windowMax {
		m.windowMax = f.total
	}
	m.windowBytes += f.bytes
	if d := now.Sub(m.windowStart); d >= time.Second {
		m.LoomFPS = float64(m.loomFrames) / d.Seconds()
		m.AstraFPS = float64(m.astraFrames) / d.Seconds()
		m.RedrawAvg = m.windowTotal / time.Duration(m.loomFrames)
		m.RedrawMax = m.windowMax
		m.BytesPerFrame = m.windowBytes / m.loomFrames
		m.loomFrames, m.astraFrames, m.windowStart = 0, 0, now
		m.windowTotal, m.windowMax, m.windowBytes = 0, 0, 0
	}
}

// countingWriter counts the bytes written through it.
type countingWriter struct {
	w interface{ WriteString(string) (int, error) }
	n int
}

func (c *countingWriter) WriteString(s string) (int, error) {
	n, err := c.w.WriteString(s)
	c.n += n
	return n, err
}
