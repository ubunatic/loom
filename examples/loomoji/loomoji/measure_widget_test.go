// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomoji

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestMeasureWidgetStagesAndConfirmsPages(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "measurements.json")
	textPath := filepath.Join(dir, "measurements.txt")
	store := NewMeasurementStore(TerminalProfile{Term: "xterm", TermProgram: "test"})
	w := NewMeasureWidget(store, jsonPath, textPath)
	page := w.PageGlyphs()
	if len(page) != measurePageSize {
		t.Fatalf("PageGlyphs() has %d rows, want %d", len(page), measurePageSize)
	}
	if got, want := w.CurrentGlyph(), page[0]; got != want {
		t.Fatalf("CurrentGlyph() = %q, want %q", got, want)
	}
	if got, want := w.PageMeasurement(0).MeasuredWidth, NewMeasurement(page[0]).ComputedWidth; got != want {
		t.Fatalf("initial row answer = %d, want computed width %d", got, want)
	}

	w.HandleKey(loom.KeyEvent{Text: "0"})
	firstToggle := 1
	if NewMeasurement(page[0]).ComputedWidth == 1 {
		firstToggle = 2
	}
	if got := w.PageMeasurement(0).MeasuredWidth; got != firstToggle {
		t.Fatalf("first 0 toggle = %d, want %d", got, firstToggle)
	}
	w.HandleKey(loom.KeyEvent{Text: "0"})
	secondToggle := 1
	if firstToggle == 1 {
		secondToggle = 2
	}
	if got := w.PageMeasurement(0).MeasuredWidth; got != secondToggle {
		t.Fatalf("second 0 toggle = %d, want %d", got, secondToggle)
	}
	w.HandleKey(loom.KeyEvent{Key: "down"})
	w.HandleKey(loom.KeyEvent{Text: "?"})
	if got := w.PageMeasurement(1); got.MeasuredWidth != 0 || got.Answered {
		t.Fatalf("staged unsure row = %#v, want unconfirmed unsure answer", got)
	}
	w.HandleKey(loom.KeyEvent{Text: "c"})
	w.HandleKey(loom.KeyEvent{Text: "manual observation"})
	w.HandleKey(loom.KeyEvent{Key: "enter"})
	if got := w.PageMeasurement(1).Comment; got != "manual observation" {
		t.Fatalf("staged comment = %q, want manual observation", got)
	}
	if len(store.Entries) != 0 {
		t.Fatalf("unconfirmed page saved entries: %#v", store.Entries)
	}

	w.HandleKey(loom.KeyEvent{Key: "enter"})
	if got, want := w.page, 1; got != want {
		t.Fatalf("page after confirmation = %d, want %d", got, want)
	}
	if got := store.Entries[page[0]]; !got.Answered || got.MeasuredWidth != secondToggle {
		t.Fatalf("confirmed first row = %#v", got)
	}
	if got := store.Entries[page[1]]; !got.Answered || got.MeasuredWidth != 0 || got.Comment != "manual observation" {
		t.Fatalf("confirmed unsure row = %#v", got)
	}
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("page confirmation did not save JSON: %v", err)
	}
	if report, err := os.ReadFile(textPath); err != nil {
		t.Fatalf("read report: %v", err)
	} else if !strings.Contains(string(report), "manual observation") {
		t.Errorf("incremental text report does not contain comment: %s", report)
	}
	if w.Err() != nil {
		t.Errorf("widget save error = %v", w.Err())
	}
}

func TestMeasureWidgetPageNavigationAndDraw(t *testing.T) {
	w := NewMeasureWidget(NewMeasurementStore(TerminalProfile{}), "", "")
	first := w.CurrentGlyph()
	w.HandleKey(loom.KeyEvent{Key: "pgdown"})
	if w.page != 1 {
		t.Fatalf("PgDn page = %d, want 1", w.page)
	}
	w.HandleKey(loom.KeyEvent{Key: "pgup"})
	if got := w.CurrentGlyph(); got != first {
		t.Fatalf("PgUp current glyph = %q, want %q", got, first)
	}

	canvas := loom.NewCanvas(120, 18)
	w.Draw(canvas, loom.Rect{X: 0, Y: 0, W: 120, H: 18})
	rows := make([]string, 18)
	for i := range rows {
		rows[i] = canvas.Row(i)
	}
	rendered := strings.Join(rows, "\n")
	for _, want := range []string{"Terminal emoji width measurement", "Row  Glyph", "|" + first + "|", "Computed", "Answer", "0-9: toggle"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("measure widget rendering missing %q:\n%s", want, rendered)
		}
	}
}

func TestMeasureWidgetQuitLeavesUnconfirmedPageUnmeasured(t *testing.T) {
	dir := t.TempDir()
	store := NewMeasurementStore(TerminalProfile{})
	w := NewMeasureWidget(store, filepath.Join(dir, "measurements.json"), "")
	glyph := w.CurrentGlyph()
	w.HandleKey(loom.KeyEvent{Text: "0"})
	if !w.HandleKey(loom.KeyEvent{Text: "q"}) {
		t.Fatal("q did not request quit")
	}
	if _, exists := store.Entries[glyph]; exists {
		t.Fatalf("q saved unconfirmed row %q", glyph)
	}
}
