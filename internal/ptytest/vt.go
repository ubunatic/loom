// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package ptytest runs real binaries on pseudo-terminals and mirrors their
// output through the public loomtest virtual terminal.
package ptytest

import "ubunatic.com/loom/loomtest"

type VT = loomtest.VT
type Cell = loomtest.Cell
type Style = loomtest.Style
type Color = loomtest.Color
type ColorType = loomtest.ColorType

const (
	ColorDefault = loomtest.ColorDefault
	ColorIndexed = loomtest.ColorIndexed
	ColorDirect  = loomtest.ColorDirect
)

func NewVT(cols, rows int) *VT     { return loomtest.NewVT(cols, rows) }
func ColorReset() Color            { return loomtest.ColorReset() }
func ColorIndex(n uint8) Color     { return loomtest.ColorIndex(n) }
func ColorRGB(r, g, b uint8) Color { return loomtest.ColorRGB(r, g, b) }
