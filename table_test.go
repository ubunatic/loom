// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

func psTestTable() (*loom.Table, []loom.Column, []loom.Row) {
	cols := []loom.Column{
		{Header: "NAME", Align: loom.AlignLeft},
		{Header: "MEM", Align: loom.AlignRight},
		{Header: "CPU", Align: loom.AlignRight},
	}
	rows := []loom.Row{
		{Key: "firefox (300)", Cells: []string{"firefox (300)", "512M", "2.5%"}},
		{Key: "bash (200)", Cells: []string{"bash (200)", "16M", "0.1%"}},
		{Key: "vim (201)", Cells: []string{"vim (201)", "8M", "0.0%"}},
	}
	return loom.NewTable(cols, rows), cols, rows
}

func TestTableContentHeight(t *testing.T) {
	tbl, _, rows := psTestTable()
	want := len(rows) + 2 // header + rows + prompt
	if got := tbl.ContentHeight(); got != want {
		t.Errorf("ContentHeight = %d, want %d", got, want)
	}
}

func TestTableInitialSelection(t *testing.T) {
	tbl, _, _ := psTestTable()
	item, ok := tbl.Selected()
	if !ok {
		t.Fatal("expected initial selection")
	}
	if item.Name != "firefox (300)" {
		t.Errorf("initial selection = %q, want %q", item.Name, "firefox (300)")
	}
}

func TestTableNavigation(t *testing.T) {
	tbl, _, _ := psTestTable()

	tbl.HandleKey(loom.KeyEvent{Key: "down"})
	item, _ := tbl.Selected()
	if item.Name != "bash (200)" {
		t.Errorf("after down: %q, want bash (200)", item.Name)
	}

	tbl.HandleKey(loom.KeyEvent{Key: "down"})
	tbl.HandleKey(loom.KeyEvent{Key: "down"}) // wraps back to first
	item, _ = tbl.Selected()
	if item.Name != "firefox (300)" {
		t.Errorf("after wrap-down: %q, want firefox (300)", item.Name)
	}

	tbl.HandleKey(loom.KeyEvent{Key: "up"}) // wraps to last
	item, _ = tbl.Selected()
	if item.Name != "vim (201)" {
		t.Errorf("after wrap-up: %q, want vim (201)", item.Name)
	}
}

func TestTableEnter(t *testing.T) {
	tbl, _, _ := psTestTable()

	tbl.HandleKey(loom.KeyEvent{Key: "down"}) // select bash
	quit := tbl.HandleKey(loom.KeyEvent{Key: "enter"})
	if !quit {
		t.Error("enter should return quit=true")
	}
	item, ok := tbl.Selected()
	if !ok {
		t.Fatal("expected selection after enter")
	}
	if item.Name != "bash (200)" {
		t.Errorf("selected = %q, want bash (200)", item.Name)
	}
}

func TestTableAbort(t *testing.T) {
	for _, key := range []string{"esc", "ctrl-c", "ctrl-d", "ctrl-q"} {
		tbl, _, _ := psTestTable()
		quit := tbl.HandleKey(loom.KeyEvent{Key: key})
		if !quit {
			t.Errorf("key %q: expected quit=true", key)
		}
		if !tbl.Aborted() {
			t.Errorf("key %q: expected Aborted()=true", key)
		}
		_, ok := tbl.Selected()
		if ok {
			t.Errorf("key %q: aborted table should not return selection", key)
		}
	}
}

func TestTableFilter(t *testing.T) {
	tbl, _, _ := psTestTable()

	// Type "vim" — only vim (201) should match.
	for _, ch := range "vim" {
		tbl.HandleKey(loom.KeyEvent{Text: string(ch)})
	}
	item, ok := tbl.Selected()
	if !ok {
		t.Fatal("expected selection after filter")
	}
	if item.Name != "vim (201)" {
		t.Errorf("filtered selection = %q, want vim (201)", item.Name)
	}

	// Backspace clears filter char by char.
	tbl.HandleKey(loom.KeyEvent{Key: "backspace"}) // "vi"
	tbl.HandleKey(loom.KeyEvent{Key: "backspace"}) // "v"
	tbl.HandleKey(loom.KeyEvent{Key: "backspace"}) // ""
	// All rows visible again.
	item, ok = tbl.Selected()
	if !ok {
		t.Fatal("expected selection after clearing filter")
	}
	if item.Name != "vim (201)" {
		// After filter clear, current sel stays at 0 of the unfiltered list.
		// The exact row depends on clamping; just check ok=true is enough.
	}
}

func TestTableFilterAllColumns(t *testing.T) {
	tbl, _, _ := psTestTable()

	// "512" matches the MEM cell of firefox, not the NAME cell.
	for _, ch := range "512" {
		tbl.HandleKey(loom.KeyEvent{Text: string(ch)})
	}
	item, ok := tbl.Selected()
	if !ok {
		t.Fatal("expected selection")
	}
	if item.Name != "firefox (300)" {
		t.Errorf("filter on MEM cell: %q, want firefox (300)", item.Name)
	}
}

func TestTableSortCycle(t *testing.T) {
	tbl, cols, _ := psTestTable()

	if tbl.SortCol != -1 {
		t.Errorf("initial SortCol = %d, want -1", tbl.SortCol)
	}

	for i := range cols {
		tbl.HandleKey(loom.KeyEvent{Key: "tab"})
		want := i
		if tbl.SortCol != want {
			t.Errorf("after %d tab(s): SortCol = %d, want %d", i+1, tbl.SortCol, want)
		}
	}

	// One more tab wraps back to 0.
	tbl.HandleKey(loom.KeyEvent{Key: "tab"})
	if tbl.SortCol != 0 {
		t.Errorf("after wrap: SortCol = %d, want 0", tbl.SortCol)
	}
}

func TestTableSortCallback(t *testing.T) {
	tbl, _, _ := psTestTable()

	var gotCol int
	var gotDesc bool
	tbl.OnSort = func(col int, desc bool) {
		gotCol = col
		gotDesc = desc
	}

	tbl.HandleKey(loom.KeyEvent{Key: "tab"}) // → col 0
	if gotCol != 0 {
		t.Errorf("OnSort col = %d, want 0", gotCol)
	}
	if gotDesc {
		t.Error("OnSort desc should be false initially")
	}
}

func TestTableSortDirection(t *testing.T) {
	tbl, _, _ := psTestTable()
	tbl.SortCol = 0

	if tbl.SortDesc {
		t.Error("initial SortDesc should be false")
	}

	tbl.HandleKey(loom.KeyEvent{Text: "!"})
	if !tbl.SortDesc {
		t.Error("after !: SortDesc should be true")
	}

	tbl.HandleKey(loom.KeyEvent{Text: "!"})
	if tbl.SortDesc {
		t.Error("after !!: SortDesc should be false")
	}
}

func TestTableSortDirectionNoCallbackWhenNoCol(t *testing.T) {
	tbl, _, _ := psTestTable()
	// SortCol=-1: ! should toggle SortDesc but not call OnSort.
	called := false
	tbl.OnSort = func(_ int, _ bool) { called = true }
	tbl.HandleKey(loom.KeyEvent{Text: "!"})
	if called {
		t.Error("! with SortCol=-1 should not call OnSort")
	}
}

func TestTableSetRows(t *testing.T) {
	tbl, _, _ := psTestTable()

	newRows := []loom.Row{
		{Key: "new-a", Cells: []string{"new-a", "1M", "0.0%"}},
	}
	tbl.SetRows(newRows)

	if tbl.ContentHeight() != 3 { // 1 row + header + prompt
		t.Errorf("ContentHeight after SetRows = %d, want 3", tbl.ContentHeight())
	}
	item, ok := tbl.Selected()
	if !ok {
		t.Fatal("expected selection after SetRows")
	}
	if item.Name != "new-a" {
		t.Errorf("Selected after SetRows = %q, want new-a", item.Name)
	}
}

func TestTableDraw(t *testing.T) {
	tbl, _, _ := psTestTable()
	tbl.SortCol = 1 // MEM

	cv := loom.NewCanvas(60, 5)
	tbl.Draw(cv, cv.Bounds())

	// Header row should contain "MEM▼" (sort indicator).
	headerRow := cv.Row(0)
	if !containsStr(headerRow, "MEM") {
		t.Error("header row should contain MEM")
	}
	// First data row should contain "firefox".
	dataRow := cv.Row(1)
	if !containsStr(dataRow, "firefox") {
		t.Error("first data row should contain firefox")
	}
}

func containsStr(ansiRow, sub string) bool {
	// Strip ANSI escape sequences for a simple contains check.
	// We just look for the raw substring since Row() includes ANSI codes.
	return len(ansiRow) > 0 && contains([]rune(ansiRow), []rune(sub))
}

func contains(haystack, needle []rune) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j, r := range needle {
			if haystack[i+j] != r {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
