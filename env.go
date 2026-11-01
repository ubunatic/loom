// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
	"strings"
	"sync/atomic"
)

type envSwitches struct {
	fastANSI bool
}

var envSnapshot atomic.Pointer[envSwitches]

func init() {
	ReloadEnvSwitches()
}

// ReloadEnvSwitches re-reads LOOM_FAST_ANSI. loom reads these once
// at process start, so a program or test that changes them at runtime must call
// ReloadEnvSwitches for the change to take effect. Tests should register
// t.Cleanup(loom.ReloadEnvSwitches) before t.Setenv, so the restore is seen.
func ReloadEnvSwitches() {
	s := &envSwitches{
		fastANSI: fastANSIFromEnv(),
	}
	envSnapshot.Store(s)
}

func env() *envSwitches {
	if s := envSnapshot.Load(); s != nil {
		return s
	}
	ReloadEnvSwitches()
	return envSnapshot.Load()
}

func fastANSIFromEnv() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("LOOM_FAST_ANSI")))
	return v != "0" && v != "false" && v != "off"
}
