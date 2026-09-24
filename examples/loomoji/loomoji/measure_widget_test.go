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

func TestMeasureWidgetAnswersCommentsReviewAndIncrementalSave(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "measurements.json")
	textPath := filepath.Join(dir, "measurements.txt")
	store := NewMeasurementStore(TerminalProfile{Term: "xterm", TermProgram: "test"})
	first := uniqueGlyphs()[0]
	store.Set(Measurement{Glyph: first, ComputedWidth: 2, MeasuredWidth: 2, Answered: true})
	w := NewMeasureWidget(store, jsonPath, textPath)
	if w.CurrentGlyph() != uniqueGlyphs()[1] {
		t.Fatalf("CurrentGlyph() = %q, want second unique glyph %q", w.CurrentGlyph(), uniqueGlyphs()[1])
	}
	if got, want := w.Remaining(), len(uniqueGlyphs())-1; got != want {
		t.Fatalf("Remaining() = %d, want %d", got, want)
	}

	glyph := w.CurrentGlyph()
	w.HandleKey(loom.KeyEvent{Text: "c"})
	w.HandleKey(loom.KeyEvent{Text: "manual observation"})
	w.HandleKey(loom.KeyEvent{Key: "enter"})
	if got := store.Entries[glyph]; got.Comment != "manual observation" || got.Answered {
		t.Fatalf("comment-only record = %#v, want saved comment with unanswered width", got)
	}
	if w.CurrentGlyph() != glyph {
		t.Fatalf("comment edit moved to %q, want to remain on %q", w.CurrentGlyph(), glyph)
	}
	w.HandleKey(loom.KeyEvent{Text: "1"})
	if got := store.Entries[glyph]; got.MeasuredWidth != 1 || !got.Answered || got.Comment != "manual observation" {
		t.Fatalf("answer after comment = %#v", got)
	}
	secondGlyph := w.CurrentGlyph()
	if secondGlyph == "" || secondGlyph == glyph {
		t.Fatalf("answer did not advance to a new glyph: %q", secondGlyph)
	}
	w.HandleKey(loom.KeyEvent{Key: "left"})
	if w.CurrentGlyph() != glyph {
		t.Fatalf("left moved to %q, want previous glyph %q", w.CurrentGlyph(), glyph)
	}
	w.HandleKey(loom.KeyEvent{Text: "2"})
	if got := store.Entries[glyph]; got.MeasuredWidth != 2 || got.Comment != "manual observation" {
		t.Errorf("corrected answer = %#v", got)
	}
	w.HandleKey(loom.KeyEvent{Key: "right"})
	w.HandleKey(loom.KeyEvent{Text: "?"})
	if got := store.Entries[secondGlyph]; got.MeasuredWidth != 0 || !got.Answered {
		t.Errorf("unsure answer = %#v", got)
	}

	loaded, err := LoadMeasurementStore(jsonPath)
	if err != nil {
		t.Fatalf("LoadMeasurementStore() = %v", err)
	}
	if got := loaded.Entries[glyph]; got.MeasuredWidth != 2 || got.Comment != "manual observation" {
		t.Errorf("incrementally saved answer = %#v", got)
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

func TestMeasureWidgetDrawShowsObservedGlyphAndWidthReferences(t *testing.T) {
	w := NewMeasureWidget(NewMeasurementStore(TerminalProfile{}), "", "")
	canvas := loom.NewCanvas(90, 16)
	w.Draw(canvas, loom.Rect{X: 0, Y: 0, W: 90, H: 16})
	rows := strings.Join([]string{
		canvas.Row(0), canvas.Row(1), canvas.Row(2), canvas.Row(3),
		canvas.Row(4), canvas.Row(5), canvas.Row(6), canvas.Row(7),
		canvas.Row(8), canvas.Row(9), canvas.Row(10), canvas.Row(11),
		canvas.Row(12), canvas.Row(13), canvas.Row(14), canvas.Row(15),
	}, "\n")
	for _, want := range []string{"Terminal emoji width measurement", "Glyph 1 /", "Observed: |", "Reference: 1 column |x|", "2 columns |xx|"} {
		if !strings.Contains(rows, want) {
			t.Errorf("measure widget rendering missing %q:\n%s", want, rows)
		}
	}
}

func TestMeasureWidgetSkipsAnsweredUnsureGlyphs(t *testing.T) {
	store := NewMeasurementStore(TerminalProfile{})
	glyph := uniqueGlyphs()[0]
	store.Set(Measurement{Glyph: glyph, MeasuredWidth: 0, Answered: true})
	w := NewMeasureWidget(store, "", "")
	if w.CurrentGlyph() == glyph {
		t.Errorf("widget queued explicitly answered unsure glyph %q", glyph)
	}
	if got := store.Unmeasured([]string{glyph}); len(got) != 0 {
		t.Errorf("Unmeasured() included answered unsure glyph: %v", got)
	}
}
