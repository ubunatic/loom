// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansi

import (
	"fmt"
	"strings"
	"testing"
)

// ── A3: Style.ANSI and color escape sequences ─────────────────────────────────

func TestStyleResetANSI(t *testing.T) {
	// Reset: full attribute reset, then default fg (39) and default bg (49).
	got := Reset.ANSI()
	want := "\x1b[0m\x1b[39m\x1b[49m"
	if got != want {
		t.Errorf("Reset.ANSI() = %q, want %q", got, want)
	}
}

func TestStyleAttributesANSI(t *testing.T) {
	s := Style{Bold: true, Underline: true, Dim: true, Italic: true, Strike: true, Invert: true}
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
	c := ColorIndex(202)
	s := Style{FG: c}
	if !strings.Contains(s.ANSI(), "\x1b[38;5;202m") {
		t.Errorf("indexed FG missing 38;5;202: %q", s.ANSI())
	}
	bg := Style{BG: c}
	if !strings.Contains(bg.ANSI(), "\x1b[48;5;202m") {
		t.Errorf("indexed BG missing 48;5;202: %q", bg.ANSI())
	}
}

func TestColorRGBSequences(t *testing.T) {
	s := Style{FG: ColorRGB(10, 20, 30)}
	if !strings.Contains(s.ANSI(), "\x1b[38;2;10;20;30m") {
		t.Errorf("RGB FG missing 38;2;10;20;30: %q", s.ANSI())
	}
	bg := Style{BG: ColorRGB(1, 2, 3)}
	if !strings.Contains(bg.ANSI(), "\x1b[48;2;1;2;3m") {
		t.Errorf("RGB BG missing 48;2;1;2;3: %q", bg.ANSI())
	}
}

func TestColorResetSequences(t *testing.T) {
	fg := Style{FG: ColorReset()}.ANSI()
	if !strings.Contains(fg, "\x1b[39m") {
		t.Errorf("default FG should emit 39m: %q", fg)
	}
	bg := Style{BG: ColorReset()}.ANSI()
	if !strings.Contains(bg, "\x1b[49m") {
		t.Errorf("default BG should emit 49m: %q", bg)
	}
}

func TestDetectColorProfile(t *testing.T) {
	tests := []struct {
		name, loom, noColor, colorTerm, term string
		want                                 ColorProfile
	}{
		{name: "override truecolor", loom: "truecolor", noColor: "1", want: ColorProfileTrueColor},
		{name: "override none", loom: "none", colorTerm: "truecolor", want: ColorProfileNone},
		{name: "no color", noColor: "1", colorTerm: "truecolor", want: ColorProfileNone},
		{name: "colorterm", colorTerm: "24bit", term: "xterm", want: ColorProfileTrueColor},
		{name: "256 term", term: "screen-256color", want: ColorProfile256},
		{name: "basic term", term: "vt100", want: ColorProfile16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectColorProfileFrom(tt.loom, tt.noColor, tt.colorTerm, tt.term); got != tt.want {
				t.Fatalf("profile = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStyleANSIDownsamplesRGB(t *testing.T) {
	tests := []struct {
		name    string
		profile ColorProfile
		color   Color
		want    string
	}{
		{name: "truecolor", profile: ColorProfileTrueColor, color: ColorRGB(255, 0, 0), want: "\x1b[38;2;255;0;0m"},
		{name: "256 exact", profile: ColorProfile256, color: ColorRGB(255, 0, 0), want: "\x1b[38;5;9m"},
		{name: "16 exact", profile: ColorProfile16, color: ColorRGB(255, 0, 0), want: "\x1b[91m"},
		{name: "none", profile: ColorProfileNone, color: ColorRGB(255, 0, 0), want: "\x1b[39m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (Style{FG: tt.color}).ANSIFor(tt.profile); !strings.Contains(got, tt.want) {
				t.Fatalf("ANSI = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAppendAllPaletteEntries(t *testing.T) {
	for n := 0; n < 256; n++ {
		c := ColorIndex(uint8(n))
		r, g, b, ok := c.RGB()
		wr, wg, wb := computeXterm256RGB(uint8(n))
		if !ok || r != wr || g != wg || b != wb {
			t.Fatalf("palette %d: RGB mismatch", n)
		}
		for _, profile := range []ColorProfile{ColorProfileNone, ColorProfile16, ColorProfile256, ColorProfileTrueColor} {
			resolved := c.forProfile(profile)
			if profile == ColorProfile16 && resolved != nearestColor(r, g, b, 16) {
				t.Fatalf("palette %d: 16-color mismatch", n)
			}
			style := Style{FG: c, BG: c, Bold: true, Dim: true, Italic: true, Underline: true, Invert: true, Strike: true}
			want := "\x1b[0m\x1b[1m\x1b[2m\x1b[3m\x1b[4m\x1b[7m\x1b[9m" + referenceColor(resolved, profile, false) + referenceColor(resolved, profile, true)
			got := style.AppendANSIFor([]byte("prefix"), profile)
			if string(got) != "prefix"+want || style.ANSIFor(profile) != want {
				t.Fatalf("index=%d profile=%d: %q want %q", n, profile, got, want)
			}
		}
	}
}

func referenceColor(c Color, profile ColorProfile, bg bool) string {
	base := 30
	if bg {
		base = 40
	}
	if c.mode == colorReset {
		return fmt.Sprintf("\x1b[%dm", base+9)
	}
	if profile == ColorProfile16 {
		n := int(c.index)
		if n >= 8 {
			base += 60
			n -= 8
		}
		return fmt.Sprintf("\x1b[%dm", base+n)
	}
	if c.mode == colorIndex {
		return fmt.Sprintf("\x1b[%d;5;%dm", base+8, c.index)
	}
	return fmt.Sprintf("\x1b[%d;2;%d;%d;%dm", base+8, c.r, c.g, c.b)
}

func TestAppendRGBAndDefaults(t *testing.T) {
	for _, c := range []Color{ColorReset(), ColorRGB(0, 0, 0), ColorRGB(255, 255, 255), ColorRGB(10, 20, 30)} {
		for _, profile := range []ColorProfile{ColorProfileNone, ColorProfile16, ColorProfile256, ColorProfileTrueColor} {
			for _, bg := range []bool{false, true} {
				var got []byte
				if bg {
					got = c.AppendBG(nil, profile)
				} else {
					got = c.AppendFG(nil, profile)
				}
				want := referenceColor(c.forProfile(profile), profile, bg)
				if string(got) != want {
					t.Fatalf("got %q want %q", got, want)
				}
			}
		}
	}
	if _, _, _, ok := ColorReset().RGB(); ok {
		t.Fatal("reset has RGB")
	}
	if _, ok := ColorRGB(0, 0, 0).Index(); ok {
		t.Fatal("RGB has index")
	}
	if n, ok := ColorIndex(255).Index(); !ok || n != 255 {
		t.Fatal("lost index")
	}
}

func TestParseSGR(t *testing.T) {
	all := Style{Bold: true, Dim: true, Italic: true, Underline: true, Invert: true, Strike: true}
	for _, tc := range []struct {
		params        string
		initial, want Style
	}{
		{"", all, Style{}}, {"0", all, Style{}},
		{"1;2;3;4;7;9", Style{}, all},
		{"22;23;24;27;29", all, Style{}},
		{"31;104", Style{}, Style{FG: ColorIndex(1), BG: ColorIndex(12)}},
		{"97;47", Style{}, Style{FG: ColorIndex(15), BG: ColorIndex(7)}},
		{"38;5;255;48;5;0", Style{}, Style{FG: ColorIndex(255), BG: ColorIndex(0)}},
		{"38;2;0;128;255;48;2;255;0;1", Style{}, Style{FG: ColorRGB(0, 128, 255), BG: ColorRGB(255, 0, 1)}},
		{"39;49", Style{FG: ColorIndex(1), BG: ColorIndex(2)}, Style{}},
		{"garbage;999999999999999999999999999999;999", all, all},
		{"38;5;256;48;2;-1;0;0", all, all},
		{"38;5;", all, all},
		{"38;2;;0;0", all, all},
		{"1;", Style{}, Style{}}, {";1", all, Style{Bold: true}},
		{"38;2;1", Style{}, Style{Dim: true, Bold: true}},
		{"+1;-2", Style{}, Style{Bold: true}},
	} {
		if got := ParseSGR(tc.initial, tc.params); got != tc.want {
			t.Errorf("%q: got %+v want %+v", tc.params, got, tc.want)
		}
	}
}

func TestZeroAlloc(t *testing.T) {
	var buffer [128]byte
	style := Style{FG: ColorRGB(37, 99, 201), BG: ColorIndex(255), Bold: true, Underline: true}
	for _, profile := range []ColorProfile{ColorProfileNone, ColorProfile16, ColorProfile256, ColorProfileTrueColor} {
		if n := testing.AllocsPerRun(100, func() {
			b := style.AppendANSIFor(buffer[:0], profile)
			b = style.FG.AppendFG(b, profile)
			_ = style.BG.AppendBG(b, profile)
		}); n != 0 {
			t.Fatalf("profile %d: %v allocations", profile, n)
		}
	}
	if string(style.AppendANSI(nil)) != style.ANSI() {
		t.Fatal("AppendANSI mismatch")
	}
	if n := testing.AllocsPerRun(100, func() { _ = ParseSGR(style, "38;2;1;2;3;48;5;255;bogus;9999999999999999999999999999") }); n != 0 {
		t.Fatalf("parsing: %v allocations", n)
	}
}

func TestDetectEnvironment(t *testing.T) {
	for _, name := range []string{"LOOMCOLOR", "NO_COLOR", "COLORTERM", "TERM"} {
		t.Setenv(name, "")
	}
	if DetectColorProfile() != ColorProfileTrueColor {
		t.Fatal("empty environment")
	}
	t.Setenv("TERM", "dumb")
	if DetectColorProfile() != ColorProfileNone {
		t.Fatal("dumb terminal")
	}
	for _, value := range []string{" 256COLOR ", "basic", "off", "24bit", "unknown"} {
		t.Setenv("LOOMCOLOR", value)
		if DetectColorProfile() != DetectColorProfileFrom(value, "", "", "dumb") {
			t.Fatal(value)
		}
	}
}
