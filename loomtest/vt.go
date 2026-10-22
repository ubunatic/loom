// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomtest

import "ubunatic.com/loom/vt"

// VT is a minimal terminal emulator. It is not safe for concurrent use.
type VT = vt.VT

// NewVT returns a blank VT of the given size with auto-wrap enabled.
func NewVT(cols, rows int) *VT { return vt.NewVT(cols, rows) }
