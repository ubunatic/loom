// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"fmt"
	"testing"
	"ubunatic.com/loom"
)

func TestTableIncrementalFilter(t *testing.T) {
	cols := []loom.Column{{Header: "Value"}}
	rows := make([]loom.Row, 1000)
	for i := 0; i < 1000; i++ {
		rows[i] = loom.Row{Key: fmt.Sprintf("k%d", i), Cells: []string{fmt.Sprintf("Item %d", i)}}
	}

	// Create two identical tables
	tableInc := loom.NewTable(cols, rows)
	tableFull := loom.NewTable(cols, rows)

	steps := []string{"I", "It", "Ite", "Item", "Item ", "Item 1", "Item 12", "Item 1", "Item "}

	for _, step := range steps {
		// Just use OnChange to set query directly
		tableInc.SearchBar().OnChange(step)

		// For the full table, force a full refilter by resetting Rows
		tableFull.SetRows(rows)
		tableFull.SearchBar().OnChange(step)

		hInc := tableInc.ContentHeight()
		hFull := tableFull.ContentHeight()

		if hInc != hFull {
			t.Errorf("Mismatch at step %q: heights %d vs %d", step, hInc, hFull)
		}
	}
}

func BenchmarkTableFilter(b *testing.B) {
	cols := []loom.Column{{Header: "Value"}}
	rows := make([]loom.Row, 100000)
	for i := 0; i < 100000; i++ {
		rows[i] = loom.Row{Cells: []string{fmt.Sprintf("Item %d", i)}}
	}

	table := loom.NewTable(cols, rows)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset back to base
		table.SetRows(rows)
		table.SearchBar().OnChange("I")
		table.SearchBar().OnChange("It")
		table.SearchBar().OnChange("Ite")
		table.SearchBar().OnChange("Item")
	}
}
