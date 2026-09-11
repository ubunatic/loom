// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"testing"

	"codeberg.org/ubunatic/loom/measure"
)

func TestSpinnerFrames(t *testing.T) {
	for i, frame := range DefaultSpinnerFrames {
		g := SpinnerGlyph(i)
		if g != frame {
			t.Errorf("SpinnerGlyph(%d) = %c, want %c", i, g, frame)
		}
		if w := measure.RuneWidth(g); w != 1 {
			t.Errorf("RuneWidth(%c) = %d, want 1", g, w)
		}
	}

	// Negative index wraps around cleanly
	g := SpinnerGlyph(-1)
	if g != DefaultSpinnerFrames[len(DefaultSpinnerFrames)-1] {
		t.Errorf("SpinnerGlyph(-1) = %c, want %c", g, DefaultSpinnerFrames[len(DefaultSpinnerFrames)-1])
	}
}

func TestHarnezSpinnerFrames(t *testing.T) {
	for i, frame := range HarnezSpinnerFrames {
		g := SpinnerGlyphFrom(HarnezSpinnerFrames, i)
		if g != frame {
			t.Errorf("SpinnerGlyphFrom(%d) = %c, want %c", i, g, frame)
		}
		if w := measure.RuneWidth(g); w != 1 {
			t.Errorf("RuneWidth(%c) = %d, want 1", g, w)
		}
	}
}
