// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"ubunatic.com/loom"
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

func TestStyleAttributesANSI(t *testing.T) {
	s := loom.Style{Bold: true, Underline: true, Dim: true, Italic: true, Strike: true, Invert: true}
	got := s.ANSI()
	// Attributes have a stable order, followed by colors.
	want := "\x1b[0m\x1b[1m\x1b[2m\x1b[3m\x1b[4m\x1b[7m\x1b[9m\x1b[39m\x1b[49m"
	if got != want {
		t.Errorf("attr ANSI() = %q, want %q", got, want)
	}
	if strings.Index(got, "\x1b[3m") > strings.Index(got, "\x1b[4m") || strings.Index(got, "\x1b[7m") > strings.Index(got, "\x1b[9m") {
		t.Error("style attributes should use stable order")
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

func TestDetectColorProfile(t *testing.T) {
	tests := []struct {
		name, loom, noColor, colorTerm, term string
		want                                 loom.ColorProfile
	}{
		{name: "override truecolor", loom: "truecolor", noColor: "1", want: loom.ColorProfileTrueColor},
		{name: "override none", loom: "none", colorTerm: "truecolor", want: loom.ColorProfileNone},
		{name: "no color", noColor: "1", colorTerm: "truecolor", want: loom.ColorProfileNone},
		{name: "colorterm", colorTerm: "24bit", term: "xterm", want: loom.ColorProfileTrueColor},
		{name: "256 term", term: "screen-256color", want: loom.ColorProfile256},
		{name: "basic term", term: "vt100", want: loom.ColorProfile16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loom.DetectColorProfileFrom(tt.loom, tt.noColor, tt.colorTerm, tt.term); got != tt.want {
				t.Fatalf("profile = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStyleANSIDownsamplesRGB(t *testing.T) {
	tests := []struct {
		name    string
		profile loom.ColorProfile
		color   loom.Color
		want    string
	}{
		{name: "truecolor", profile: loom.ColorProfileTrueColor, color: loom.ColorRGB(255, 0, 0), want: "\x1b[38;2;255;0;0m"},
		{name: "256 exact", profile: loom.ColorProfile256, color: loom.ColorRGB(255, 0, 0), want: "\x1b[38;5;9m"},
		{name: "16 exact", profile: loom.ColorProfile16, color: loom.ColorRGB(255, 0, 0), want: "\x1b[91m"},
		{name: "none", profile: loom.ColorProfileNone, color: loom.ColorRGB(255, 0, 0), want: "\x1b[39m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (loom.Style{FG: tt.color}).ANSIFor(tt.profile); !strings.Contains(got, tt.want) {
				t.Fatalf("ANSI = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCanvasRowUsesCapturedColorProfile(t *testing.T) {
	c := loom.NewCanvas(1, 1)
	c.ColorProfile = loom.ColorProfile256
	c.Set(0, 0, loom.Cell{Text: "x", Style: loom.Style{FG: loom.ColorRGB(255, 0, 0)}})
	if got := c.Row(0); !strings.Contains(got, "\x1b[38;5;9m") {
		t.Fatalf("row = %q, want 256-color red", got)
	}
}
