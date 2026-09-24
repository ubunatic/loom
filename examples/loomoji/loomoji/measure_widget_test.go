// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomoji

import (
	"os"
	"path/filepath"
	"reflect"
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

	w.setSelectedWidth(1)
	w.HandleKey(loom.KeyEvent{Key: "left"})
	if got := w.PageMeasurement(0).MeasuredWidth; got != 4 {
		t.Fatalf("left from width 1 = %d, want 4", got)
	}
	w.HandleKey(loom.KeyEvent{Key: "right"})
	if got := w.PageMeasurement(0).MeasuredWidth; got != 1 {
		t.Fatalf("right from width 4 = %d, want 1", got)
	}
	w.HandleKey(loom.KeyEvent{Key: "right"})
	if got := w.PageMeasurement(0).MeasuredWidth; got != 2 {
		t.Fatalf("right from width 1 = %d, want 2", got)
	}
	w.HandleKey(loom.KeyEvent{Text: "0"})
	if got := w.PageMeasurement(0).MeasuredWidth; got != 2 {
		t.Fatalf("numeric key changed selected width to %d, want 2", got)
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
	if got := store.Entries[page[0]]; !got.Answered || got.MeasuredWidth != 2 {
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
	w.setSelectedWidth(NewMeasurement(first).ComputedWidth + 1)

	canvas := loom.NewCanvas(120, 18)
	w.Draw(canvas, loom.Rect{X: 0, Y: 0, W: 120, H: 18})
	rows := make([]string, 18)
	for i := range rows {
		rows[i] = canvas.Row(i)
	}
	rendered := strings.Join(rows, "\n")
	for _, want := range []string{"Terminal emoji width measurement", "Row  Glyph", "|" + first + " |", "Computed", "VTE", "Answer", "←/→: width 1-4"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("measure widget rendering missing %q:\n%s", want, rendered)
		}
	}
}

func TestEvaluateVTEWidth(t *testing.T) {
	tests := []struct {
		glyph    string
		computed int
		wantVTE  int
	}{
		{glyph: "‼️", computed: 2, wantVTE: 2},
		{glyph: "⁉️", computed: 2, wantVTE: 2},
		{glyph: "⚠️", computed: 2, wantVTE: 2},
		{glyph: "☀️", computed: 2, wantVTE: 2},
		{glyph: "☝️", computed: 2, wantVTE: 2},
		{glyph: "🇩🇪", computed: 2, wantVTE: 2},
		{glyph: "👨‍👩‍👧", computed: 2, wantVTE: 2},
		{glyph: "✊", computed: 2, wantVTE: 2},
		{glyph: "😀", computed: 2, wantVTE: 2},
		{glyph: "←", computed: 1, wantVTE: 1},
		{glyph: "a", computed: 1, wantVTE: 1},
	}
	for _, tt := range tests {
		t.Run(tt.glyph, func(t *testing.T) {
			m := NewMeasurement(tt.glyph)
			if m.ComputedWidth != tt.computed {
				t.Errorf("NewMeasurement(%q).ComputedWidth = %d, want %d", tt.glyph, m.ComputedWidth, tt.computed)
			}
			if m.VTEWidth != tt.wantVTE {
				t.Errorf("NewMeasurement(%q).VTEWidth = %d, want %d", tt.glyph, m.VTEWidth, tt.wantVTE)
			}
			if got := EvaluateVTEWidth(tt.glyph); got != tt.wantVTE {
				t.Errorf("EvaluateVTEWidth(%q) = %d, want %d", tt.glyph, got, tt.wantVTE)
			}
		})
	}
}

func TestMeasureWidgetQuitLeavesUnconfirmedPageUnmeasured(t *testing.T) {
	dir := t.TempDir()
	store := NewMeasurementStore(TerminalProfile{})
	w := NewMeasureWidget(store, filepath.Join(dir, "measurements.json"), "")
	glyph := w.CurrentGlyph()
	w.HandleKey(loom.KeyEvent{Key: "right"})
	if !w.HandleKey(loom.KeyEvent{Text: "q"}) {
		t.Fatal("q did not request quit")
	}
	if _, exists := store.Entries[glyph]; exists {
		t.Fatalf("q saved unconfirmed row %q", glyph)
	}
}

func TestMeasureWidgetReviewModeFiltersAndSavesEdits(t *testing.T) {
	dir := t.TempDir()
	store := NewMeasurementStore(TerminalProfile{})
	glyphs := uniqueGlyphs()
	store.Set(Measurement{Glyph: glyphs[0], ComputedWidth: 1, MeasuredWidth: 1, Answered: true})
	store.Set(Measurement{Glyph: glyphs[1], ComputedWidth: 2, MeasuredWidth: 2, Answered: true, Comment: "recheck"})
	w := NewMeasureWidgetWithOptions(store, filepath.Join(dir, "measurements.json"), "", MeasureOptions{Filter: MeasureFilterWidth2})
	if got := w.PageGlyphs(); !reflect.DeepEqual(got, []string{glyphs[1]}) {
		t.Fatalf("review width-2 page = %#v, want %#v", got, []string{glyphs[1]})
	}
	w.HandleKey(loom.KeyEvent{Key: "left"})
	w.HandleKey(loom.KeyEvent{Key: "enter"})
	if got := store.Entries[glyphs[1]]; got.MeasuredWidth != 1 || !got.Answered {
		t.Fatalf("saved review edit = %#v, want answered width 1", got)
	}
	w.HandleKey(loom.KeyEvent{Key: "tab"})
	if w.filter != MeasureFilterWidth3 {
		t.Fatalf("Tab filter = %s, want %s", w.filter, MeasureFilterWidth3)
	}
}

func TestMeasureGlyphWithPadding(t *testing.T) {
	tests := []struct {
		name          string
		glyph         string
		computedWidth int
		measuredWidth int
		want          string
	}{
		{name: "matching widths", glyph: "👍", computedWidth: 2, measuredWidth: 2, want: "👍"},
		{name: "terminal draws wider", glyph: "☝️", computedWidth: 1, measuredWidth: 2, want: "☝️ "},
		{name: "chosen width is narrower", glyph: "👍", computedWidth: 2, measuredWidth: 1, want: "👍"},
		{name: "unsure", glyph: "👍", computedWidth: 2, measuredWidth: 0, want: "👍"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := measureGlyphWithPadding(tt.glyph, tt.computedWidth, tt.measuredWidth); got != tt.want {
				t.Errorf("measureGlyphWithPadding() = %q, want %q", got, tt.want)
			}
		})
	}
}
