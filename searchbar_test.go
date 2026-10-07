// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func TestSearchBarDrawingAndPlaceholder(t *testing.T) {
	sb := loom.NewSearchBar()
	sb.Prompt = "> "
	sb.Placeholder = "type to filter"
	sb.Focused = true

	cv := loom.NewCanvas(30, 1)
	sb.Draw(cv, loom.Rect{X: 0, Y: 0, W: 30, H: 1})

	row := cv.Row(0)
	if !strings.Contains(row, "> ") || !strings.Contains(row, "type to filter") {
		t.Fatalf("empty search bar output = %q, want prompt and placeholder", row)
	}
	if cv.CursorX != 2 || cv.CursorY != 0 {
		t.Fatalf("cursor = (%d, %d), want (2, 0)", cv.CursorX, cv.CursorY)
	}

	// Type query
	sb.ConsumeKey(loom.KeyEvent{Text: "foo"})
	cv = loom.NewCanvas(30, 1)
	sb.Draw(cv, loom.Rect{X: 0, Y: 0, W: 30, H: 1})

	row = searchBarPlainRow(cv)
	if !strings.Contains(row, "> foo") || strings.Contains(row, "type to filter") {
		t.Fatalf("queried search bar output = %q, want query without placeholder", row)
	}
	if cv.CursorX != 5 || cv.CursorY != 0 {
		t.Fatalf("cursor after typing = (%d, %d), want (5, 0)", cv.CursorX, cv.CursorY)
	}
}

func TestSearchBarControls(t *testing.T) {
	sb := loom.NewSearchBar()
	sb.Prompt = "> "
	sb.Query = "test"
	sb.Controls = "4/12"

	cv := loom.NewCanvas(30, 1)
	sb.Draw(cv, loom.Rect{X: 0, Y: 0, W: 30, H: 1})

	row := searchBarPlainRow(cv)
	if !strings.Contains(row, "> test") || !strings.Contains(row, "4/12") {
		t.Fatalf("search bar with controls = %q, want query and controls", row)
	}
}

func searchBarPlainRow(cv *loom.Canvas) string {
	var row strings.Builder
	for x := 0; x < cv.Cols(); x++ {
		row.WriteString(cv.Get(x, 0).Text)
	}
	return row.String()
}

func TestSearchBarKeyInputAndCallbacks(t *testing.T) {
	sb := loom.NewSearchBar()
	var changed []string
	sb.OnChange = func(q string) {
		changed = append(changed, q)
	}
	var submitted string
	sb.OnSubmit = func(q string) {
		submitted = q
	}
	var aborted bool
	sb.OnAbort = func() {
		aborted = true
	}

	sb.ConsumeKey(loom.KeyEvent{Text: "a"})
	sb.ConsumeKey(loom.KeyEvent{Text: "b"})
	sb.ConsumeKey(loom.KeyEvent{Key: "backspace"})

	if sb.Query != "a" {
		t.Fatalf("query = %q, want 'a'", sb.Query)
	}
	if len(changed) != 3 || changed[0] != "a" || changed[1] != "ab" || changed[2] != "a" {
		t.Fatalf("changed log = %v, want ['a', 'ab', 'a']", changed)
	}

	sb.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if submitted != "a" {
		t.Fatalf("submitted = %q, want 'a'", submitted)
	}

	sb.ConsumeKey(loom.KeyEvent{Key: "esc"})
	if !aborted {
		t.Fatalf("aborted = false, want true")
	}
}

func TestSearchBarCommandMode(t *testing.T) {
	sb := loom.NewSearchBar()
	sb.AddCmd(loom.Cmd{Name: "custom", Title: "run custom command"})

	res := sb.ConsumeKey(loom.KeyEvent{Text: ":"})
	if !res.Consumed || !sb.IsCommandMode() {
		t.Fatalf("command mode not active after ':'")
	}

	sb.ConsumeKey(loom.KeyEvent{Text: "c"})
	cv := loom.NewCanvas(40, 1)
	sb.Draw(cv, loom.Rect{X: 0, Y: 0, W: 40, H: 1})

	row := cv.Row(0)
	if !strings.Contains(row, ":c") || !strings.Contains(row, "custom") {
		t.Fatalf("command mode drawing = %q, want completion hint", row)
	}
}

func TestChoiceSearchBarIntegration(t *testing.T) {
	items := []loom.Item{
		{Name: "Apple"},
		{Name: "Banana"},
		{Name: "Cherry"},
	}
	choice := loom.NewChoice(items)
	if choice.SearchBar() == nil {
		t.Fatalf("choice.SearchBar() is nil")
	}

	choice.ConsumeKey(loom.KeyEvent{Text: "ban"})
	if choice.Query() != "ban" {
		t.Fatalf("choice.Query() = %q, want 'ban'", choice.Query())
	}
	if choice.SearchBar().Query != "ban" {
		t.Fatalf("choice.SearchBar().Query = %q, want 'ban'", choice.SearchBar().Query)
	}

	sel, ok := choice.Selected()
	if !ok || sel.Name != "Banana" {
		t.Fatalf("selected = %+v, want Banana", sel)
	}

	// Controlling Choice via returned SearchBar.SetQuery
	choice.SearchBar().SetQuery("che")
	if choice.Query() != "che" {
		t.Fatalf("choice.Query() after SetQuery = %q, want 'che'", choice.Query())
	}
	sel, ok = choice.Selected()
	if !ok || sel.Name != "Cherry" {
		t.Fatalf("selected after SetQuery = %+v, want Cherry", sel)
	}
}

func TestTableSearchBarIntegration(t *testing.T) {
	cols := []loom.Column{
		{Header: "Item"},
		{Header: "Qty"},
	}
	rows := []loom.Row{
		{Key: "a", Cells: []string{"Apple", "10"}},
		{Key: "b", Cells: []string{"Banana", "5"}},
		{Key: "c", Cells: []string{"Cherry", "20"}},
	}
	table := loom.NewTable(cols, rows)
	if table.SearchBar() == nil {
		t.Fatalf("table.SearchBar() is nil")
	}

	table.ConsumeKey(loom.KeyEvent{Text: "ban"})
	if table.SearchBar().Query != "ban" {
		t.Fatalf("table.SearchBar().Query = %q, want 'ban'", table.SearchBar().Query)
	}

	sel, ok := table.Selected()
	if !ok || sel.Name != "b" {
		t.Fatalf("selected = %+v, want b (Banana)", sel)
	}

	// Controlling Table via returned SearchBar.SetQuery
	table.SearchBar().SetQuery("che")
	sel, ok = table.Selected()
	if !ok || sel.Name != "c" {
		t.Fatalf("selected after SetQuery = %+v, want c (Cherry)", sel)
	}

	// SetRows preserves the active query filter
	newRows := []loom.Row{
		{Key: "a2", Cells: []string{"Apple Fresh", "12"}},
		{Key: "c2", Cells: []string{"Cherry Tart", "8"}},
	}
	table.SetRows(newRows)
	sel, ok = table.Selected()
	if !ok || sel.Name != "c2" {
		t.Fatalf("selected after SetRows = %+v, want c2 (Cherry Tart)", sel)
	}
}

func TestSearchBarContainerAndPromptStyleDistinction(t *testing.T) {
	sb := loom.NewSearchBar()
	sb.Prompt = "> "
	sb.Query = "test"
	sb.Style.Container = loom.Style{BG: loom.ColorIndex(234)}
	sb.Style.Prompt = loom.Style{FG: loom.ColorIndex(51)}
	sb.Style.Query = loom.Style{FG: loom.ColorIndex(250)}

	cv := loom.NewCanvas(20, 1)
	sb.Draw(cv, loom.Rect{X: 0, Y: 0, W: 20, H: 1})

	// Background of entire row should be container BG (234)
	for x := 0; x < 20; x++ {
		if got := cv.Get(x, 0).Style.BG; got != loom.ColorIndex(234) {
			t.Fatalf("cell %d BG = %+v, want ColorIndex(234)", x, got)
		}
	}
	// Prompt glyph FG should be 51
	if got := cv.Get(0, 0).Style.FG; got != loom.ColorIndex(51) {
		t.Fatalf("prompt FG = %+v, want ColorIndex(51)", got)
	}
}

func TestThemeSearchBarStyle(t *testing.T) {
	for _, themeName := range []string{"plain", "mc", "julia256"} {
		theme := loom.Theme(themeName)
		sbStyle := theme.SearchBarStyle()
		choiceStyle := theme.ChoiceStyle()
		tableStyle := theme.TableStyle()

		if choiceStyle.SearchBar.Container != sbStyle.Container {
			t.Errorf("theme %q: choiceStyle.SearchBar.Container mismatch", themeName)
		}
		if tableStyle.SearchBar.Container != sbStyle.Container {
			t.Errorf("theme %q: tableStyle.SearchBar.Container mismatch", themeName)
		}
	}
}
