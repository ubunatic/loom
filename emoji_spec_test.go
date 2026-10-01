// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"

	"ubunatic.com/loom/measure"
)

func TestEmbeddedEmojiSpecMatchesRuntime(t *testing.T) {
	if len(emojiYAML) == 0 {
		t.Fatal("emojiYAML is empty")
	}
	spec := measure.ActiveEmojiSpec()
	if spec.VS16DefaultWidth != 2 {
		t.Errorf("VS16DefaultWidth = %d, want 2", spec.VS16DefaultWidth)
	}
	if spec.ZWJDefaultWidth != 2 {
		t.Errorf("ZWJDefaultWidth = %d, want 2", spec.ZWJDefaultWidth)
	}
	if spec.FlagDefaultWidth != 2 {
		t.Errorf("FlagDefaultWidth = %d, want 2", spec.FlagDefaultWidth)
	}
	for _, o := range spec.Overrides {
		if strings.ContainsRune(o.Glyph, '\u200D') {
			continue // ZWJ widths are selected from the runtime terminal probe.
		}
		if got := measure.StringWidth(o.Glyph); got != o.Width {
			t.Errorf("measure.StringWidth(%q) = %d, want specced %d (%s)", o.Glyph, got, o.Width, o.Note)
		}
	}
}
