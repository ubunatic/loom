// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"os"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

// ── A3: Style.ANSI and color escape sequences ─────────────────────────────────

func TestStyleResetANSI(t *testing.T) {
	// Reset: full attribute reset, then default fg (39) and default bg (49).
	got := loom.Reset.ANSI()
	want := "\x1b[0m\x1b[39m\x1b[49m"
	if got != want {
		t.Errorf("Reset.ANSI() = %q, want %q", got, want)
	}
}

func TestStyleAppendANSI(t *testing.T) {
	s := loom.Style{
		FG:        loom.ColorRGB(120, 200, 50),
		BG:        loom.ColorIndex(42),
		Bold:      true,
		Underline: true,
	}

	buf := s.AppendANSI(nil)
	got := string(buf)
	want := s.ANSI()

	if got != want {
		t.Errorf("AppendANSI() = %q, want %q", got, want)
	}
}

func TestStyleAttributesANSI(t *testing.T) {
	s := loom.Style{Bold: true, Underline: true, Dim: true}
	got := s.ANSI()
	// Always starts with a full reset, then bold, underline, dim, then colors.
	want := "\x1b[0m\x1b[1m\x1b[4m\x1b[2m\x1b[39m\x1b[49m"
	if got != want {
		t.Errorf("attr ANSI() = %q, want %q", got, want)
	}
	// Sanity: order of attributes is bold → underline → dim.
	if strings.Index(got, "\x1b[1m") > strings.Index(got, "\x1b[4m") {
		t.Error("bold should precede underline")
	}
	if strings.Index(got, "\x1b[4m") > strings.Index(got, "\x1b[2m") {
		t.Error("underline should precede dim")
	}
}

func TestColorIndexSequences(t *testing.T) {
	c := loom.ColorIndex(202)
	s := loom.Style{FG: c}
	if !strings.Contains(s.ANSI(), "\x1b[38;5;202m") {
		t.Errorf("indexed FG missing 38;5;202: %q", s.ANSI())
	}
	bg := loom.Style{BG: c}
	if !strings.Contains(bg.ANSI(), "\x1b[48;5;202m") {
		t.Errorf("indexed BG missing 48;5;202: %q", bg.ANSI())
	}
}

func TestColorRGBSequences(t *testing.T) {
	s := loom.Style{FG: loom.ColorRGB(10, 20, 30)}
	if !strings.Contains(s.ANSI(), "\x1b[38;2;10;20;30m") {
		t.Errorf("RGB FG missing 38;2;10;20;30: %q", s.ANSI())
	}
	bg := loom.Style{BG: loom.ColorRGB(1, 2, 3)}
	if !strings.Contains(bg.ANSI(), "\x1b[48;2;1;2;3m") {
		t.Errorf("RGB BG missing 48;2;1;2;3: %q", bg.ANSI())
	}
}

func TestColorResetSequences(t *testing.T) {
	fg := loom.Style{FG: loom.ColorReset()}.ANSI()
	if !strings.Contains(fg, "\x1b[39m") {
		t.Errorf("default FG should emit 39m: %q", fg)
	}
	bg := loom.Style{BG: loom.ColorReset()}.ANSI()
	if !strings.Contains(bg, "\x1b[49m") {
		t.Errorf("default BG should emit 49m: %q", bg)
	}
}

func TestFastANSIEnvVarToggleParity(t *testing.T) {
	s := loom.Style{
		FG:        loom.ColorRGB(100, 150, 200),
		BG:        loom.ColorIndex(50),
		Bold:      true,
		Underline: true,
	}
	c := loom.NewCanvas(10, 1)
	c.Set(0, 0, loom.Cell{Text: "X", Style: s})

	os.Unsetenv("LOOM_FAST_ANSI")
	fastStyle := s.ANSI()
	fastRow := c.Row(0)

	os.Setenv("LOOM_FAST_ANSI", "0")
	slowStyle := s.ANSI()
	slowRow := c.Row(0)

	os.Unsetenv("LOOM_FAST_ANSI")

	if fastStyle != slowStyle {
		t.Errorf("Style.ANSI mismatch fast=%q vs slow=%q", fastStyle, slowStyle)
	}
	if fastRow != slowRow {
		t.Errorf("Canvas.Row mismatch fast=%q vs slow=%q", fastRow, slowRow)
	}
}

// ── Benchmarks ──────────────────────────────────────────────────────────────

func BenchmarkStyleANSI(b *testing.B) {
	s := loom.Style{
		FG:        loom.ColorRGB(120, 200, 50),
		BG:        loom.ColorIndex(42),
		Bold:      true,
		Underline: true,
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.ANSI()
	}
}

func BenchmarkStyleAppendANSI(b *testing.B) {
	s := loom.Style{
		FG:        loom.ColorRGB(120, 200, 50),
		BG:        loom.ColorIndex(42),
		Bold:      true,
		Underline: true,
	}
	var buf [64]byte
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.AppendANSI(buf[:0])
	}
}
