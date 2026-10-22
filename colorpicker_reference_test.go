// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"ubunatic.com/loom"
	"ubunatic.com/loom/loomtest"
)

func TestColorPickerReference(t *testing.T) {
	const width, height = 36, 6
	loomtest.CompareANSI(t, "docs/data/ColorPicker.ansi", width, height, func() string {
		canvas := loom.NewCanvas(width, height)
		loom.NewColorPicker(loom.ColorIndex(4), loom.ColorIndex(15)).Draw(canvas, canvas.Bounds())
		var rendered strings.Builder
		for y := 0; y < height; y++ {
			if y > 0 {
				rendered.WriteByte('\n')
			}
			rendered.WriteString(canvas.Row(y))
		}
		return rendered.String() + "\n"
	})
}
