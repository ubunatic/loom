// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
)

func TestWidgetsColorPickerGalleryComparisonPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	build := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build loom binary: %v\n%s", err, output)
	}
	s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "--theme", "plain", "ColorPicker")
	s.WaitFor("Theme surface (BG)", 5*time.Second)
	cellForSample := func(label string) ptytest.Cell {
		t.Helper()
		for y, line := range s.Screen() {
			if at := strings.Index(line, label); at >= 0 {
				return s.Cell(loom.StringWidth(line[:at]), y+1)
			}
		}
		t.Fatalf("sample label %q not found:\n%s", label, strings.Join(s.Screen(), "\n"))
		return ptytest.Cell{}
	}
	labelPosition := func(rows []string, label string) (int, int, bool) {
		for y, line := range rows {
			if at := strings.Index(line, label); at >= 0 {
				return loom.StringWidth(line[:at]), y, true
			}
		}
		return 0, 0, false
	}
	defaultCell := cellForSample("Default")
	if defaultCell.Style.FG != ptytest.ColorReset() || defaultCell.Style.BG != ptytest.ColorReset() || defaultCell.Style.Dim {
		t.Fatalf("Default PTY cell = %+v, want reset FG/BG without Dim", defaultCell)
	}
	faintCell := cellForSample("Faint text")
	if faintCell.Style.FG != ptytest.ColorReset() || faintCell.Style.BG != ptytest.ColorReset() || !faintCell.Style.Dim {
		t.Fatalf("Faint text PTY cell = %+v, want reset FG/BG with Dim", faintCell)
	}
	initialSubdued := cellForSample("Theme surface (BG)")
	if initialSubdued.Style.FG != ptytest.ColorReset() || initialSubdued.Style.BG != ptytest.ColorReset() {
		t.Fatalf("plain theme sample = %+v, want reset FG/BG", initialSubdued)
	}
	palette := loom.SpeccedDefaults.ColorPicker.Palette
	initialSelected := -1
	for i, color := range palette {
		if color == 4 {
			initialSelected = i
			break
		}
	}
	if initialSelected < 0 || initialSelected+1 >= len(palette) {
		t.Fatalf("palette %v cannot support the expected Right navigation", palette)
	}
	selectedColor := loom.ColorIndex(uint8(palette[initialSelected+1]))
	cursorColumn := loom.SpeccedDefaults.ColorPicker.LeadingPadding + (initialSelected+2)*loom.SpeccedDefaults.ColorPicker.SwatchWidth
	pickerState := func(rows []string, selectionColor loom.Color) bool {
		labelX, labelY, ok := labelPosition(rows, "Preview")
		if !ok {
			return false
		}
		cursorX := labelX - loom.SpeccedDefaults.ColorPicker.PreviewPadding + cursorColumn
		cursorY := labelY - 5
		cells := s.Cells()
		if cursorY < 0 || cursorY >= len(cells) || cursorX < 0 || cursorX >= len(cells[cursorY]) {
			return false
		}
		cursor := cells[cursorY][cursorX]
		preview := cells[labelY][labelX]
		return cursor.Text == loom.SpeccedDefaults.ColorPicker.Glyphs.EdgeLower &&
			ptyColorMatches(cursor.Style.FG, selectionColor) && ptyColorMatches(preview.Style.FG, selectedColor)
	}
	s.Send("\x1b[C") // Right moves the picker cursor one palette entry.
	s.WaitForUpdateState(5*time.Second, func(rows []string) bool {
		return pickerState(rows, loom.Theme("plain").Selection.Color())
	})
	s.Send(strings.Repeat("\x1b[C", len(palette)-initialSelected-2) + "\x1b[C\r") // Reach Faint and apply it.
	s.WaitForUpdateState(5*time.Second, func(rows []string) bool {
		return strings.Contains(strings.Join(rows, "\n"), "Faint: [x]")
	})
	s.WaitForUpdateState(5*time.Second, func(rows []string) bool {
		labelX, labelY, ok := labelPosition(rows, "Preview")
		return ok && s.Cells()[labelY][labelX].Style.Dim
	})

	names := loom.ThemeNames()
	next := ""
	for i, name := range names {
		if name == "plain" {
			next = names[(i+1)%len(names)]
			break
		}
	}
	if next == "" {
		t.Fatal("plain theme is missing")
	}
	s.Send("\x1b[20~") // F9
	s.WaitForUpdateState(5*time.Second, func(rows []string) bool {
		return strings.Contains(strings.Join(rows, "\n"), "Theme "+next) && strings.Contains(strings.Join(rows, "\n"), "Faint: [x]")
	})
	updatedSubdued := cellForSample("Theme surface (BG)")
	if updatedSubdued.Style.FG != ptytest.ColorReset() || !ptyColorsMatch(updatedSubdued.Style.BG, loom.Theme(next).NormalBG.Color()) {
		t.Fatalf("%s theme sample = %+v, want reset FG and NormalBG", next, updatedSubdued)
	}
}

func ptyColorMatches(got ptytest.Color, want loom.Color) bool {
	if want == loom.ColorReset() {
		return got == ptytest.ColorReset()
	}
	if index, ok := want.Index(); ok {
		return got.Type == ptytest.ColorIndexed && got.Index == index
	}
	return ptyColorsMatch(got, want)
}

func ptyColorsMatch(got ptytest.Color, want loom.Color) bool {
	gr, gg, gb, gok := got.RGB()
	wr, wg, wb, wok := want.RGB()
	return gok && wok && gr == wr && gg == wg && gb == wb
}
