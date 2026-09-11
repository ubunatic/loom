// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import "codeberg.org/ubunatic/loom"

// The pane-driver widgets must satisfy Paneable so RunPane can drive them.
// RunPane itself opens /dev/tty and cannot run under `go test`; these
// compile-time assertions guard the contract instead.
var (
	_ loom.Paneable = (*loom.Choice)(nil)
	_ loom.Paneable = (*loom.Table)(nil)
	_ loom.Paneable = (*loom.View)(nil)
)
