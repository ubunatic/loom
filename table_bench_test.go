// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"fmt"
	"testing"
	"ubunatic.com/loom"
)

func BenchmarkTableDraw(b *testing.B) {
	cols := []loom.Column{
		{Header: "ID", Width: 10},
		{Header: "Name", Width: 20},
		{Header: "Value", Width: 30},
	}
	rows := make([]loom.Row, 10000)
	for i := 0; i < 10000; i++ {
		rows[i] = loom.Row{
			Key: fmt.Sprintf("row-%d", i),
			Cells: []string{
				fmt.Sprintf("%d", i),
				fmt.Sprintf("Item %d", i),
				fmt.Sprintf("Value %d", i*100),
			},
		}
	}
	table := loom.NewTable(cols, rows)
	canvas := loom.NewCanvas(80, 24)
	rect := loom.Rect{X: 0, Y: 0, W: 80, H: 24}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table.Draw(canvas, rect)
	}
}
