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
	for _, state := range []struct {
		name string
		keys []string
	}{
		{name: "fg-active"},
		{name: "bg-active", keys: []string{"down"}},
		{name: "bg-rug", keys: []string{"down", "left", "left", "space"}},
	} {
		t.Run(state.name, func(t *testing.T) {
			picker := loom.NewColorPicker(loom.ColorIndex(4), loom.ColorIndex(15))
			picker.ApplyTheme(loom.Theme("julia256"))
			for _, key := range state.keys {
				picker.ConsumeKey(loom.KeyEvent{Key: key})
			}
			loomtest.CompareWidget(t, "testdata/ColorPicker-"+state.name+".ansi", width, height, func() string {
				canvas := loom.NewCanvas(width, height)
				picker.Draw(canvas, canvas.Bounds())
				var rendered strings.Builder
				for y := 0; y < height; y++ {
					if y > 0 {
						rendered.WriteByte('\n')
					}
					rendered.WriteString(canvas.Row(y))
				}
				return rendered.String() + "\n"
			})
		})
	}
}
