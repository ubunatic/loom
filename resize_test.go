// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"testing"
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
