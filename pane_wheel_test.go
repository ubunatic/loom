// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"testing"
	"time"
)

// wheelReportsPerNotch is what foot sends for one wheel notch (issue 451,
// measured with loom-probe mouse). Tests read the curve from the spec, but the
// notch size is a terminal fact.
const wheelReportsPerNotch = 3

const (
	wheelUpReport   = "\x1b[<64;10;5M"
	wheelDownReport = "\x1b[<65;10;5M"
	leftPressReport = "\x1b[<0;10;5M"
)

// wheelCounter counts every mouse event delivered to it.
type wheelCounter struct {
	events int
}

func (*wheelCounter) Draw(*Canvas, Rect) {}

func (w *wheelCounter) ConsumeKey(KeyEvent) EventResult { return Ignored() }

func (w *wheelCounter) ConsumeMouse(MouseEvent) EventResult {
	w.events++
	return Handled()
}

// wheelHarness drives Pane.dispatchMouseRead with a scripted clock.
type wheelHarness struct {
	t     *testing.T
	pane  *Pane
	clock time.Time
	root  *wheelCounter
}

func newWheelHarness(t *testing.T) *wheelHarness {
	t.Helper()
	h := &wheelHarness{t: t, clock: time.Unix(1_000_000, 0), root: &wheelCounter{}}
	h.pane = &Pane{now: func() time.Time { return h.clock }}
	return h
}

// send delivers one SGR report after waiting gap on the scripted clock. It
// returns how many times the widget received the report.
func (h *wheelHarness) send(report string, gap time.Duration) int {
	h.t.Helper()
	h.clock = h.clock.Add(gap)
	before := h.root.events
	h.pane.dispatchMouseRead(h.root, nil, []byte(report), func(bool) {})
	return h.root.events - before
}

// wheelBurstGap returns the spec's burst gap as a duration.
func wheelBurstGap() time.Duration {
	return time.Duration(SpeccedDefaults.Pane.Wheel.BurstGapMS) * time.Millisecond
}

// fastWheelGap is well inside a burst; slowWheelGap is well outside one.
func fastWheelGap() time.Duration { return wheelBurstGap() / 10 }
func slowWheelGap() time.Duration { return wheelBurstGap() * 2 }

// sendScroll delivers n wheel reports of one direction, each after gap, and
// returns the total number of deliveries.
func (h *wheelHarness) sendScroll(report string, n int, gap time.Duration) int {
	h.t.Helper()
	delivered := 0
	for range n {
		delivered += h.send(report, gap)
	}
	return delivered
}

func TestWheelSlowNotchScrollsOneLinePerReport(t *testing.T) {
	h := newWheelHarness(t)
	// One notch: wheelReportsPerNotch reports, the first after a pause.
	if got := h.sendScroll(wheelDownReport, wheelReportsPerNotch, slowWheelGap()); got != wheelReportsPerNotch {
		t.Fatalf("one slow notch delivered %d reports, want %d", got, wheelReportsPerNotch)
	}
}

func TestWheelOneOrTwoNotchesStayUnchanged(t *testing.T) {
	// Two notches are 2*wheelReportsPerNotch reports in one burst. The ramp
	// starts after ramp_after reports, so both notches must still deliver one
	// report per report. This fails if the spec ramp_after drops below six.
	const notches = 2
	h := newWheelHarness(t)
	got := 0
	for range notches {
		got += h.sendScroll(wheelDownReport, wheelReportsPerNotch, fastWheelGap())
		h.clock = h.clock.Add(slowWheelGap())
	}
	if want := notches * wheelReportsPerNotch; got != want {
		t.Fatalf("two notches delivered %d reports, want %d (gain must stay 1)", got, want)
	}
}

func TestWheelSlowScrollDeliversEachReportOnce(t *testing.T) {
	h := newWheelHarness(t)
	const reports = 30
	if got := h.sendScroll(wheelDownReport, reports, slowWheelGap()); got != reports {
		t.Fatalf("slow scroll delivered %d reports, want %d", got, reports)
	}
}

func TestWheelFastBurstScrollsMoreThanSlowNotches(t *testing.T) {
	const reports = 30
	slow := newWheelHarness(t)
	slowGot := slow.sendScroll(wheelDownReport, reports, slowWheelGap())

	fast := newWheelHarness(t)
	fastGot := fast.sendScroll(wheelDownReport, reports, fastWheelGap())

	if slowGot != reports {
		t.Fatalf("slow control delivered %d, want %d", slowGot, reports)
	}
	if fastGot <= slowGot {
		t.Fatalf("fast burst delivered %d, slow notches delivered %d: want more", fastGot, slowGot)
	}
}

func TestWheelGainRampFollowsSpec(t *testing.T) {
	cfg := SpeccedDefaults.Pane.Wheel
	cases := []struct {
		count, want int
	}{
		{1, 1},
		{cfg.RampAfter, 1},
		{cfg.RampAfter + cfg.RampStep - 1, 1},
		{cfg.RampAfter + cfg.RampStep, 2},
		{cfg.RampAfter + 2*cfg.RampStep, 3},
	}
	for _, tc := range cases {
		if got := wheelGain(tc.count, cfg); got != tc.want {
			t.Errorf("wheelGain(%d) = %d, want %d", tc.count, got, tc.want)
		}
	}
}

func TestWheelGainCapHolds(t *testing.T) {
	cfg := SpeccedDefaults.Pane.Wheel
	h := newWheelHarness(t)
	var last int
	for range 4 * wheelGainCapReports(cfg) {
		last = h.send(wheelUpReport, fastWheelGap())
		if last > cfg.MaxGain {
			t.Fatalf("one report delivered %d times, cap is %d", last, cfg.MaxGain)
		}
	}
	if last != cfg.MaxGain {
		t.Fatalf("long burst settled at gain %d, want the cap %d", last, cfg.MaxGain)
	}
}

// wheelGainCapReports is the burst length at which the spec reaches max_gain.
func wheelGainCapReports(cfg PaneWheelDefaults) int {
	return cfg.RampAfter + (cfg.MaxGain-1)*cfg.RampStep
}

// wheelFreshBurstTotal is what a burst of n fast reports delivers when it is
// the first thing the pane sees.
func wheelFreshBurstTotal(t *testing.T, report string, n int) int {
	t.Helper()
	return newWheelHarness(t).sendScroll(report, n, fastWheelGap())
}

func TestWheelPauseResetsBurst(t *testing.T) {
	n := wheelGainCapReports(SpeccedDefaults.Pane.Wheel) + 5
	fresh := wheelFreshBurstTotal(t, wheelDownReport, n)

	h := newWheelHarness(t)
	h.sendScroll(wheelDownReport, n, fastWheelGap())
	h.clock = h.clock.Add(slowWheelGap())
	// Without a reset the second burst would start at the cap and deliver more.
	if got := h.sendScroll(wheelDownReport, n, fastWheelGap()); got != fresh {
		t.Fatalf("burst after a pause delivered %d, a fresh burst delivers %d", got, fresh)
	}
}

func TestWheelGapAtBurstLimitResets(t *testing.T) {
	cfg := SpeccedDefaults.Pane.Wheel
	h := newWheelHarness(t)
	h.sendScroll(wheelDownReport, wheelGainCapReports(cfg)+5, fastWheelGap())
	// The burst gap is exclusive: a gap equal to burst_gap_ms ends the burst.
	if got := h.send(wheelDownReport, wheelBurstGap()); got != 1 {
		t.Fatalf("report at exactly burst_gap_ms delivered %d times, want 1", got)
	}
}

func TestWheelDirectionChangeResetsBurst(t *testing.T) {
	n := wheelGainCapReports(SpeccedDefaults.Pane.Wheel) + 5
	fresh := wheelFreshBurstTotal(t, wheelUpReport, n)

	h := newWheelHarness(t)
	h.sendScroll(wheelDownReport, n, fastWheelGap())
	// Reversing at full speed starts a new burst: the up burst ramps from gain 1.
	if got := h.sendScroll(wheelUpReport, n, fastWheelGap()); got != fresh {
		t.Fatalf("up burst after a down burst delivered %d, a fresh up burst delivers %d", got, fresh)
	}
}

func TestWheelRepeatsOnlyWheelReports(t *testing.T) {
	h := newWheelHarness(t)
	h.sendScroll(wheelDownReport, wheelGainCapReports(SpeccedDefaults.Pane.Wheel)+5, fastWheelGap())
	if got := h.send(leftPressReport, fastWheelGap()); got != 1 {
		t.Fatalf("left press during a wheel burst delivered %d times, want 1", got)
	}
}

func TestPaneWheelSpecValidates(t *testing.T) {
	if err := SpeccedDefaults.Pane.Validate(); err != nil {
		t.Fatalf("embedded pane defaults invalid: %v", err)
	}
	bad := SpeccedDefaults.Pane
	bad.Wheel.MaxGain = 0
	if err := bad.Validate(); err == nil {
		t.Fatal("pane.wheel.max_gain 0 passed validation")
	}
}
