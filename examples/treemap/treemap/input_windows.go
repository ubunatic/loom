// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build windows

package treemap

import "context"

func watchForQuitKey(context.CancelFunc) func() {
	return func() {}
}
