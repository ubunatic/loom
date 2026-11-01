// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package measure

import (
	"os"
	"strings"
	"sync/atomic"
)

// envSwitches is the process-wide snapshot of the environment switches that
// measure consults on hot paths. It is read once, at package init, so width
// calls never touch the process environment.
type envSwitches struct {
	fastMeasure        bool
	zwjOverrideSet     bool
	zwjOverride        int32      // zwjSplit or zwjJoin; valid when zwjOverrideSet
	detectedRenderPath RenderPath // DetectRenderPath result; used when no SetRenderPath override
}

var envSnapshot atomic.Pointer[envSwitches]

func init() {
	ReloadEnvSwitches()
}

// ReloadEnvSwitches re-reads LOOM_FAST_MEASURE, LOOM_ZWJ and LOOM_RENDER_PATH
// and the terminal variables used by DetectRenderPath. Measure reads these once
// at process start, so a program or test that changes them at runtime must call
// ReloadEnvSwitches for the change to take effect. Tests should register
// t.Cleanup(measure.ReloadEnvSwitches) before t.Setenv, so the restore is seen.
func ReloadEnvSwitches() {
	s := &envSwitches{
		fastMeasure:        fastMeasureFromEnv(),
		detectedRenderPath: DetectRenderPath(),
	}
	if mode, ok := parseZWJMode(os.Getenv("LOOM_ZWJ")); ok {
		s.zwjOverrideSet = true
		s.zwjOverride = mode
	}
	envSnapshot.Store(s)
}

// env returns the current snapshot, loading it on first use if the package
// variables are initialised before init has run.
func env() *envSwitches {
	if s := envSnapshot.Load(); s != nil {
		return s
	}
	ReloadEnvSwitches()
	return envSnapshot.Load()
}

func fastMeasureFromEnv() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("LOOM_FAST_MEASURE")))
	return v != "0" && v != "false" && v != "off"
}
