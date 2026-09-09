// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package graph provides dependency-free renderers for single-row,
// fixed-width terminal graphs: determinate fill bars and rolling sparklines.
//
// The rendering primitives are adapted from Harnez's internal/rograph package
// (commit 01e59b331c9d85699e55fbbb946c579e19901083).
package graph

// MaxWidth is the default maximum render width, in characters/glyphs, for
// single-row graph bars and sparklines.
const MaxWidth = 10

// RenderProgressBar generates an ANSI/Unicode progress bar of the given
// character width: a "[filled empty]" bar using '█' for the filled portion
// and '░' for the empty portion, scaled to usedPercent (clamped to [0, 100]).
// The boundary character where fill transitions from filled to empty renders
// at eighth-block precision (▏▎▍▌▋▊▉█) rather than snapping to fully filled
// or empty.
//
// If width < 1, it clamps to 1.
func RenderProgressBar(usedPercent float64, width int) string {
	if width < 1 {
		width = 1
	}
	return RenderBar(usedPercent, BarOptions{Width: width, SubChar: true})
}
