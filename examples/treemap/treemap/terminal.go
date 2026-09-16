// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// Terminal geometry is demo plumbing: the graph renderer accepts explicit
// dimensions, while this app must clamp them to the live output terminal.
func terminalSize(out *os.File) (width, height int, ok bool) {
	if !term.IsTerminal(int(out.Fd())) {
		return 0, 0, false
	}
	cols, rows, err := term.GetSize(int(out.Fd()))
	if err != nil || cols < 1 || rows < 1 {
		return 0, 0, false
	}
	return cols, rows, true
}

// resolveDimensions fills unset dimensions from the terminal, or from an
// 80x24 fallback when output is redirected. Explicit dimensions are clamped
// to the terminal to prevent row wrapping from scrambling the grid.
func resolveDimensions(width, height int) (int, int) {
	cols, rows, ok := terminalSize(os.Stdout)
	if !ok {
		if width <= 0 {
			width = 80
		}
		if height <= 0 {
			height = 24
		}
		return width, height
	}
	w, h, widthClamped, heightClamped := clampDimensions(width, height, cols, rows)
	if widthClamped {
		fmt.Fprintf(os.Stderr, "treemap: --width %d exceeds the terminal's %d columns; clamping to avoid line-wrap corruption\n", width, cols)
	}
	if heightClamped {
		fmt.Fprintf(os.Stderr, "treemap: --height %d exceeds the usable terminal height %d; clamping\n", height, h)
	}
	return w, h
}

// clampDimensions leaves one terminal row free for the cursor and chooses
// defaults that leave another row for the shell prompt.
func clampDimensions(width, height, cols, rows int) (w, h int, widthClamped, heightClamped bool) {
	switch {
	case width <= 0:
		w = cols
	case width > cols:
		w = cols
		widthClamped = true
	default:
		w = width
	}
	maxHeight := rows - 1
	if maxHeight < 1 {
		maxHeight = 1
	}
	switch {
	case height <= 0:
		h = rows - 2
		if h < 1 {
			h = 1
		}
	case height > maxHeight:
		h = maxHeight
		heightClamped = true
	default:
		h = height
	}
	return w, h, widthClamped, heightClamped
}
