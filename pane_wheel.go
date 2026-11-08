// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "time"

// wheelBurst tracks the current burst of same-direction wheel reports (issue
// 451). The curve comes from spec pane.wheel; Go holds no copy of it.
type wheelBurst struct {
	action MouseAction
	last   time.Time
	count  int // reports in the current burst; 0 before the first report
}

// record adds one wheel report at now and returns how many times the report is
// delivered to the widget. A gap of burst_gap_ms or more, or a change of
// direction, starts a new burst.
func (b *wheelBurst) record(now time.Time, action MouseAction, cfg PaneWheelDefaults) int {
	gap := time.Duration(cfg.BurstGapMS) * time.Millisecond
	if b.count == 0 || action != b.action || now.Sub(b.last) >= gap {
		b.count = 0
	}
	b.count++
	b.action = action
	b.last = now
	return wheelGain(b.count, cfg)
}

// wheelGain returns the delivery count for the count-th report of a burst. It
// is 1 up to ramp_after reports, then rises by 1 for every ramp_step further
// reports, capped at max_gain.
func wheelGain(count int, cfg PaneWheelDefaults) int {
	if count <= cfg.RampAfter {
		return 1
	}
	return min(1+(count-cfg.RampAfter)/cfg.RampStep, cfg.MaxGain)
}

// wheelRepeat returns how many times to deliver the wheel report of action
// that arrives now. It reads p.now, or time.Now when that is nil.
func (p *Pane) wheelRepeat(action MouseAction) int {
	now := time.Now()
	if p.now != nil {
		now = p.now()
	}
	return p.wheel.record(now, action, SpeccedDefaults.Pane.Wheel)
}
