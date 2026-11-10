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

// BenchmarkTableFilter types a filter on 100k rows. Each step narrows the
// match set (100k, 11k, 1.1k, 111 rows), so an incremental filter shows its
// gain here.
func BenchmarkTableFilter(b *testing.B) {
	cols := []loom.Column{{Header: "Value"}}
	rows := make([]loom.Row, 100000)
	for i := range rows {
		rows[i] = loom.Row{Cells: []string{fmt.Sprintf("Item %d", i)}}
	}
	table := loom.NewTable(cols, rows)
	bar := table.SearchBar()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bar.OnChange("")
		bar.OnChange("Item")
		bar.OnChange("Item 1")
		bar.OnChange("Item 12")
		bar.OnChange("Item 123")
	}
}
