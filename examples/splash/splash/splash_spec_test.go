// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package splash

import (
	"testing"
	"time"

	"ubunatic.com/loom"
)

// TestSplashTickIntervalUsesSpec overrides spinner.tick_interval. A return to the
// former 80ms literal fails here.
func TestSplashTickIntervalUsesSpec(t *testing.T) {
	saved := loom.SpeccedDefaults
	t.Cleanup(func() { loom.SpeccedDefaults = saved })
	loom.SpeccedDefaults.Spinner.TickInterval = 41 * time.Millisecond

	app := &splashApp{}
	if got := app.TickInterval(); got != 41*time.Millisecond {
		t.Fatalf("TickInterval() = %v, want spec spinner.tick_interval", got)
	}
}
