// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"bytes"
	"testing"
	"ubunatic.com/loom"
)

func BenchmarkParseAnsiBuffer(b *testing.B) {
	canvas := loom.NewCanvas(200, 60)
	rect := loom.Rect{X: 0, Y: 0, W: 200, H: 60}
	style := loom.Style{}
	canvas.DrawBorder(rect, loom.BoxBorderStyleDouble, style)

	var buf bytes.Buffer
	buf.Grow(200 * 60 * 10)
	canvas.Flush(&buf, 0)

	content := buf.String()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loom.ParseAnsiBuffer(content, 200, 60)
	}
}
