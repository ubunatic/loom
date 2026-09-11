// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

// DefaultSpinnerFrames are the standard 10-frame braille rotating spinner glyphs.
var DefaultSpinnerFrames = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// HarnezSpinnerFrames is an alternative braille sequence matching the Harnez reference target.
var HarnezSpinnerFrames = []rune{'⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏', '⠋'}

// SpinnerGlyph returns the braille spinner rune for a given frame index from DefaultSpinnerFrames.
func SpinnerGlyph(frame int) rune {
	return SpinnerGlyphFrom(DefaultSpinnerFrames, frame)
}

// SpinnerString returns the braille spinner rune as a string.
func SpinnerString(frame int) string {
	return string(SpinnerGlyph(frame))
}

// SpinnerGlyphFrom returns the spinner rune from custom frames for a given frame index.
func SpinnerGlyphFrom(frames []rune, frame int) rune {
	if len(frames) == 0 {
		return ' '
	}
	n := len(frames)
	idx := ((frame % n) + n) % n
	return frames[idx]
}
