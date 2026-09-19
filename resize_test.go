// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"testing"
	"time"
)

func TestSpeccedResizeModesIntegrity(t *testing.T) {
	if len(SpeccedResizeModes.Modes) != len(SpeccedResizeModeIDs) {
		t.Fatalf("modes count = %d, want %d", len(SpeccedResizeModes.Modes), len(SpeccedResizeModeIDs))
	}

	for _, id := range SpeccedResizeModeIDs {
		mode, ok := SpeccedResizeModes.Modes[id]
		if !ok {
			t.Fatalf("mode %q missing from SpeccedResizeModes", id)
		}
		if mode.ID != id {
			t.Errorf("mode.ID = %q, want %q", mode.ID, id)
		}
		if mode.Title == "" {
			t.Errorf("mode %q has empty Title", id)
		}
		if mode.Description == "" {
			t.Errorf("mode %q has empty Description", id)
		}
		if mode.Key == "" {
			t.Errorf("mode %q has empty Key", id)
		}
	}

	// Verify candidate defaults match the specced contract
	cfg := DefaultResizeConfig()
	if !cfg.Coalesce {
		t.Error("coalesce default should be true")
	}
	if !cfg.AtomicFlush {
		t.Error("atomic_flush default should be true")
	}
	if !cfg.RowClear {
		t.Error("row_clear default should be true")
	}
	if !cfg.SynchronizedOutput {
		t.Error("synchronized_output default should be true")
	}
	if !cfg.AutoWrap {
		t.Error("auto_wrap default should be true")
	}
	if cfg.OutOfBandClear {
		t.Error("out_of_band_clear default should be false (diagnostic only)")
	}
	if !cfg.FullScreenBuffer {
		t.Error("full_screen_buffer default should be true")
	}
	if !cfg.ResizeHandling {
		t.Error("resize_handling default should be true")
	}
	if !cfg.WidthGuard {
		t.Error("width_guard default should be true")
	}
	if cfg.WidthGuardN != 1 {
		t.Errorf("width_guard_n default = %d, want 1", cfg.WidthGuardN)
	}
}

func TestResizeConfigGetSetToggleReset(t *testing.T) {
	cfg := DefaultResizeConfig()

	for _, id := range SpeccedResizeModeIDs {
		val, ok := cfg.Get(id)
		if !ok {
			t.Fatalf("Get(%q) returned false", id)
		}
		if !cfg.Toggle(id) {
			t.Fatalf("Toggle(%q) returned false", id)
		}
		newVal, _ := cfg.Get(id)
		if newVal == val {
			t.Fatalf("Toggle(%q) did not invert state (%v -> %v)", id, val, newVal)
		}
		if !cfg.Set(id, val) {
			t.Fatalf("Set(%q, %v) returned false", id, val)
		}
		restoredVal, _ := cfg.Get(id)
		if restoredVal != val {
			t.Fatalf("Set(%q, %v) did not restore state", id, val)
		}
	}

	// Unknown mode
	if _, ok := cfg.Get("unknown_mode"); ok {
		t.Fatal("Get(unknown_mode) returned true")
	}
	if cfg.Set("unknown_mode", true) {
		t.Fatal("Set(unknown_mode) returned true")
	}
	if cfg.Toggle("unknown_mode") {
		t.Fatal("Toggle(unknown_mode) returned true")
	}

	// Reset
	cfg.Set("coalesce", false)
	cfg.Set("out_of_band_clear", true)
	cfg.Reset()
	if !cfg.Coalesce || cfg.OutOfBandClear {
		t.Fatal("Reset did not restore default configuration")
	}
}

func TestPaneResizeModeHelpers(t *testing.T) {
	p := &Pane{ResizeConfig: DefaultResizeConfig()}
	if !p.ToggleResizeMode("coalesce") {
		t.Fatal("ToggleResizeMode failed")
	}
	if p.ResizeConfig.Coalesce {
		t.Fatal("ToggleResizeMode did not invert Coalesce")
	}
	if !p.SetResizeMode("coalesce", true) {
		t.Fatal("SetResizeMode failed")
	}
	if !p.ResizeConfig.Coalesce {
		t.Fatal("SetResizeMode did not set Coalesce")
	}
	p.SetResizeMode("out_of_band_clear", true)
	p.SetWidthGuardN(3)
	if p.ResizeConfig.WidthGuardN != 3 {
		t.Fatalf("SetWidthGuardN(3) = %d, want 3", p.ResizeConfig.WidthGuardN)
	}
	p.SetWidthGuardN(0) // should clamp to 1
	if p.ResizeConfig.WidthGuardN != 1 {
		t.Fatalf("SetWidthGuardN(0) = %d, want 1", p.ResizeConfig.WidthGuardN)
	}
	if p.WidthGuardActive() {
		t.Fatal("WidthGuardActive() should be false initially")
	}
	p.ResetResizeModes()
	if p.ResizeConfig.OutOfBandClear {
		t.Fatal("ResetResizeModes did not reset OutOfBandClear")
	}
	if !p.ResizeConfig.WidthGuard || p.ResizeConfig.WidthGuardN != 1 {
		t.Fatal("ResetResizeModes did not restore WidthGuard defaults")
	}
}

func TestWinchMeterAndAdaptiveGuardN(t *testing.T) {
	a := SpeccedResizeModes.Modes["adaptive_guard"]
	window := time.Duration(a.WindowMS) * time.Millisecond
	t0 := time.Unix(1000, 0)
	const manual = 3

	var m WinchMeter
	if got := AdaptiveGuardN(m.Rate(t0), m.Events(t0), manual); got != manual {
		t.Fatalf("no events: n=%d, want manual %d", got, manual)
	}
	m.Record(t0)
	if got := AdaptiveGuardN(m.Rate(t0), m.Events(t0), manual); got != manual {
		t.Fatalf("one event: n=%d, want manual %d", got, manual)
	}

	// Two events far apart (a slow rate) give min_n; two events 10ms apart
	// (100/s) are measured at once.
	m.Record(t0.Add(window * 4 / 5))
	at := t0.Add(window * 4 / 5)
	if got := AdaptiveGuardN(m.Rate(at), m.Events(at), manual); got != a.MinN {
		t.Fatalf("two events: n=%d, want min_n %d", got, a.MinN)
	}

	var burst WinchMeter
	burst.Record(t0)
	burst.Record(t0.Add(10 * time.Millisecond))
	bn := t0.Add(10 * time.Millisecond)
	if got, want := AdaptiveGuardN(burst.Rate(bn), burst.Events(bn), manual), min(a.MaxN, a.MinN+100/a.RateStep); got != want {
		t.Fatalf("100/s burst: n=%d, want %d", got, want)
	}

	// Rate grows with event density and clamps at max_n.
	var fast WinchMeter
	for i := 0; i < 1000; i++ {
		fast.Record(t0.Add(time.Duration(i) * window / 2000))
	}
	now := t0.Add(window / 2)
	if got := AdaptiveGuardN(fast.Rate(now), fast.Events(now), manual); got != a.MaxN {
		t.Fatalf("saturated: n=%d, want max_n %d", got, a.MaxN)
	}

	// Changing rate: events age out of the window, so the rate falls back.
	later := t0.Add(2 * window)
	if got := m.Events(later); got != 0 {
		t.Fatalf("aged out events = %d, want 0", got)
	}
	if got := AdaptiveGuardN(m.Rate(later), m.Events(later), manual); got != manual {
		t.Fatalf("after settle: n=%d, want manual %d", got, manual)
	}
}

func TestEffectiveGuardNPolicy(t *testing.T) {
	p := &Pane{ResizeConfig: DefaultResizeConfig()}
	p.ResizeConfig.WidthGuardN = 4
	p.adaptiveN = 6
	if got := p.EffectiveGuardN(); got != 4 {
		t.Fatalf("adaptive off: n=%d, want manual 4", got)
	}
	p.ResizeConfig.AdaptiveGuard = true
	if got := p.EffectiveGuardN(); got != 6 {
		t.Fatalf("adaptive on: n=%d, want 6", got)
	}
	p.adaptiveN = 0
	if got := p.EffectiveGuardN(); got != 4 {
		t.Fatalf("adaptive unmeasurable: n=%d, want manual 4", got)
	}
}

func TestGuardedColsUsesEffectiveN(t *testing.T) {
	p := &Pane{ResizeConfig: DefaultResizeConfig(), widthGuardActive: true}
	p.ResizeConfig.WidthGuardN = 2
	p.ResizeConfig.AdaptiveGuard = true
	p.adaptiveN = 5
	if got := p.guardedCols(80); got != 75 {
		t.Fatalf("guardedCols = %d, want 75", got)
	}
	p.widthGuardActive = false
	if got := p.guardedCols(80); got != 80 {
		t.Fatalf("inactive guardedCols = %d, want 80", got)
	}
	p.widthGuardActive = true
	p.MaxCols = 40
	if got := p.guardedCols(80); got != 35 {
		t.Fatalf("MaxCols guardedCols = %d, want 35", got)
	}
}

func TestAltScreenModeIsSpecced(t *testing.T) {
	cfg := DefaultResizeConfig()
	if cfg.AltScreen != SpeccedResizeModes.Modes["alt_screen"].Default {
		t.Fatal("alt_screen default differs from spec")
	}
	if !cfg.Toggle("alt_screen") || !cfg.AltScreen {
		t.Fatal("alt_screen toggle failed")
	}
}
