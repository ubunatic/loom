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
	s.WaitFor("Subdued BG (theme)", 5*time.Second)
	cellForSample := func(label string) ptytest.Cell {
		t.Helper()
		for y, line := range s.Screen() {
			if x := strings.Index(line, label); x >= 0 {
				return s.Cell(x, y+1)
			}
		}
		t.Fatalf("sample label %q not found:\n%s", label, strings.Join(s.Screen(), "\n"))
		return ptytest.Cell{}
	}
	defaultCell := cellForSample("Default")
	if defaultCell.Style.FG != ptytest.ColorReset() || defaultCell.Style.BG != ptytest.ColorReset() || defaultCell.Style.Dim {
		t.Fatalf("Default PTY cell = %+v, want reset FG/BG without Dim", defaultCell)
	}
	faintCell := cellForSample("Faint text")
	if faintCell.Style.FG != ptytest.ColorReset() || faintCell.Style.BG != ptytest.ColorReset() || !faintCell.Style.Dim {
		t.Fatalf("Faint text PTY cell = %+v, want reset FG/BG with Dim", faintCell)
	}
	initialSubdued := cellForSample("Subdued BG (theme)")
	if initialSubdued.Style.FG != ptytest.ColorReset() || initialSubdued.Style.BG != ptytest.ColorReset() {
		t.Fatalf("plain theme sample = %+v, want reset FG/BG", initialSubdued)
	}

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
	s.WaitFor("Theme "+next, 5*time.Second)
	updatedSubdued := cellForSample("Subdued BG (theme)")
	if updatedSubdued.Style.FG != ptytest.ColorReset() || !ptyColorsMatch(updatedSubdued.Style.BG, loom.Theme(next).NormalBG.Color()) {
		t.Fatalf("%s theme sample = %+v, want reset FG and NormalBG", next, updatedSubdued)
	}
}

func ptyColorsMatch(got ptytest.Color, want loom.Color) bool {
	gr, gg, gb, gok := got.RGB()
	wr, wg, wb, wok := want.RGB()
	return gok && wok && gr == wr && gg == wg && gb == wb
}
