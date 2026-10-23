// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build windows

package terminal

// DrainInput flushes the terminal input queue. On Windows this is a no-op.
func DrainInput(fd int) {}
