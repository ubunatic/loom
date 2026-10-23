// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package vt

import (
	"slices"
	"strings"
	"testing"
)

func TestVTNoWrapKeepsCursorOnRightmostCell(t *testing.T) {
	v := NewVT(3, 1)
	v.Write([]byte("\x1b[?7lab│")) //nolint:errcheck
	if x, _ := v.Cursor(); x != 2 {
		t.Fatalf("no-wrap cursor = %d, want last column 2", x)
	}
	v.Write([]byte("\x1b[K")) //nolint:errcheck
	if got := v.Screen()[0]; got != "ab" {
		t.Fatalf("EL at right edge = %q, want border erased", got)
	}
}

func TestVTPositioningAndErase(t *testing.T) {
	v := NewVT(12, 3)
	v.Write([]byte("\x1b[1;1Hhello world\x1b[2;3Hab\x1b[1;6H\x1b[K")) //nolint:errcheck
	want := []string{"hello", "  ab", ""}
	if got := v.Screen(); !slices.Equal(got, want) {
		t.Fatalf("screen = %q, want %q", got, want)
	}
}

func TestVTAutoWrapToggle(t *testing.T) {
	v := NewVT(4, 2)
	v.Write([]byte("abcdef")) //nolint:errcheck
	if got := v.Screen(); !slices.Equal(got, []string{"abcd", "ef"}) {
		t.Fatalf("wrap on: %q", got)
	}
	v = NewVT(4, 2)
	v.Write([]byte("\x1b[?7labcdef")) //nolint:errcheck
	if got := v.Screen(); !slices.Equal(got, []string{"abcf", ""}) {
		t.Fatalf("wrap off: %q", got)
	}
}

func TestVTWideRuneAndSplitWrites(t *testing.T) {
	v := NewVT(6, 1)
	for _, chunk := range []string{"a\xe4", "\xb8\xad\x1b[", "1;1Hz"} {
		v.Write([]byte(chunk)) //nolint:errcheck
	}
	if got := v.Screen()[0]; got != "z 中" && got != "z中" {
		t.Fatalf("row = %q", got)
	}
}

func TestVTScreenPreservesZWJCluster(t *testing.T) {
	const family = "👨\u200d👩\u200d👧"
	v := NewVT(8, 1)
	v.Write([]byte(family)) //nolint:errcheck
	if got := v.Screen()[0]; got != family {
		t.Fatalf("screen = %q, want ZWJ cluster %q", got, family)
	}
}

func TestVTFrameSnapshotsAtSynchronizedEnd(t *testing.T) {
	v := NewVT(5, 1)
	v.Write([]byte("\x1b[?2026h\x1b[1;1Hone\x1b[?2026l\x1b[?2026h\x1b[1;1Htwo\x1b[K\x1b[?2026l")) //nolint:errcheck
	if len(v.Frames) != 2 || v.Frames[0][0] != "one" || v.Frames[1][0] != "two" {
		t.Fatalf("frames = %q", v.Frames)
	}
}

func TestVTResizeKeepsTopLeft(t *testing.T) {
	v := NewVT(6, 2)
	v.Write([]byte("abcdef\x1b[2;1Hxyz")) //nolint:errcheck
	v.Resize(3, 1)
	if got := v.Screen(); !slices.Equal(got, []string{"abc"}) {
		t.Fatalf("screen = %q", got)
	}
}

func TestVTANSIPlainAndTrimmed(t *testing.T) {
	v := NewVT(10, 3)
	v.Write([]byte("hello     \r\nworld")) //nolint:errcheck
	got := v.ANSI()
	want := "hello\nworld\n"
	if got != want {
		t.Fatalf("VT.ANSI() = %q, want %q", got, want)
	}
}

func TestVTANSISGRStylesAndColors(t *testing.T) {
	v := NewVT(20, 2)
	// Line 1: Bold red text, reset, then 256-color green background text
	v.Write([]byte("\x1b[1;1H\x1b[1;31mRed\x1b[0m \x1b[48;5;46mGreenBG\x1b[0m")) //nolint:errcheck
	// Line 2: 24-bit truecolor text
	v.Write([]byte("\x1b[2;1H\x1b[38;2;10;20;30mRGB\x1b[0m")) //nolint:errcheck

	ansi := v.ANSI()
	if !strings.Contains(ansi, "\x1b[0;1;31mRed\x1b[0m") {
		t.Errorf("ANSI output does not contain bold red SGR: %q", ansi)
	}
	if !strings.Contains(ansi, "GreenBG") {
		t.Errorf("ANSI output does not contain GreenBG: %q", ansi)
	}
	if !strings.Contains(ansi, "\x1b[0;38;2;10;20;30mRGB\x1b[0m") {
		t.Errorf("ANSI output does not contain 24-bit RGB SGR: %q", ansi)
	}

	// Round-trip test: feed the ANSI output into a new VT and compare cells
	v2 := NewVT(20, 2)
	lines := strings.Split(strings.TrimSuffix(ansi, "\n"), "\n")
	v2.Write([]byte(strings.Join(lines, "\r\n"))) //nolint:errcheck
	for y := 0; y < 2; y++ {
		for x := 0; x < 20; x++ {
			c1 := v.Cell(x, y)
			c2 := v2.Cell(x, y)
			if c1.Rune != c2.Rune || c1.Text != c2.Text || c1.Style != c2.Style {
				t.Fatalf("cell mismatch at (%d,%d): want %+v, got %+v", x, y, c1, c2)
			}
		}
	}
}

func TestVTANSIAlternateScreen(t *testing.T) {
	v := NewVT(10, 2)
	v.Write([]byte("main screen")) //nolint:errcheck
	// Enter alternate screen (?1049h)
	v.Write([]byte("\x1b[?1049h\x1b[1;1Halt screen")) //nolint:errcheck
	if !v.AltScreen() {
		t.Fatal("expected AltScreen to be active")
	}

	ansi := v.ANSI()
	if !strings.HasPrefix(ansi, "alt screen\n") {
		t.Fatalf("VT.ANSI() in alt screen = %q, want starting with 'alt screen'", ansi)
	}
}

func TestVTANSIKeepsStyledTrailingSpace(t *testing.T) {
	v := NewVT(10, 1)
	// "A" followed by a space with reverse video (visible background)
	v.Write([]byte("\x1b[1;1HA\x1b[7m \x1b[0m")) //nolint:errcheck
	ansi := v.ANSI()
	if !strings.Contains(ansi, "A\x1b[0;7m \x1b[0m\n") {
		t.Fatalf("VT.ANSI() did not preserve styled space: %q", ansi)
	}
}

func TestVTANSITrimsTrailingBlankRows(t *testing.T) {
	v := NewVT(80, 24)
	v.Write([]byte("Line 1\r\nLine 2\r\nLine 3")) //nolint:errcheck
	ansi := v.ANSI()
	lines := strings.Split(strings.TrimSuffix(ansi, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("VT.ANSI() row count = %d, want 3; lines:\n%q", len(lines), lines)
	}
	want := "Line 1\nLine 2\nLine 3\n"
	if ansi != want {
		t.Fatalf("VT.ANSI() = %q, want %q", ansi, want)
	}
}

func TestVTRepaintRoundTripPlain(t *testing.T) {
	v := NewVT(20, 5)
	v.Write([]byte("\x1b[1;1HHello\x1b[3;5HWorld\x1b[5;2HEnd\x1b[2;7H")) //nolint:errcheck

	repaint := v.Repaint()
	v2 := NewVT(20, 5)
	v2.Write([]byte(repaint)) //nolint:errcheck

	if got, want := v2.Screen(), v.Screen(); !slices.Equal(got, want) {
		t.Fatalf("v2.Screen() = %q, want %q", got, want)
	}
	x1, y1 := v.Cursor()
	x2, y2 := v2.Cursor()
	if x1 != x2 || y1 != y2 {
		t.Fatalf("v2 cursor = (%d,%d), want (%d,%d)", x2, y2, x1, y1)
	}
	if v2.CursorVisible != v.CursorVisible {
		t.Fatalf("v2 cursorVisible = %v, want %v", v2.CursorVisible, v.CursorVisible)
	}
	if v2.AltScreen() != v.AltScreen() {
		t.Fatalf("v2 altScreen = %v, want %v", v2.AltScreen(), v.AltScreen())
	}
}

func TestVTRepaintRoundTripSGR(t *testing.T) {
	v := NewVT(25, 4)
	v.Write([]byte("\x1b[1;1H\x1b[1;31mBoldRed\x1b[0m \x1b[48;5;46mGreenBG\x1b[0m\r\n")) //nolint:errcheck
	v.Write([]byte("\x1b[38;2;10;20;30mRGB\x1b[0m \x1b[4;7mUnderRev\x1b[0m"))             //nolint:errcheck

	repaint := v.Repaint()
	v2 := NewVT(25, 4)
	v2.Write([]byte(repaint)) //nolint:errcheck

	for y := 0; y < 4; y++ {
		for x := 0; x < 25; x++ {
			c1 := v.Cell(x, y)
			c2 := v2.Cell(x, y)
			if c1 != c2 {
				t.Fatalf("cell mismatch at (%d,%d): want %+v, got %+v", x, y, c1, c2)
			}
		}
	}
}

func cellsMatch(c1, c2 Cell) bool {
	if c1.Style != c2.Style {
		return false
	}
	t1 := c1.Text
	if t1 == "" && c1.Rune == ' ' {
		t1 = " "
	}
	t2 := c2.Text
	if t2 == "" && c2.Rune == ' ' {
		t2 = " "
	}
	return c1.Rune == c2.Rune && t1 == t2
}

func TestVTRepaintRoundTripWideRunesAndZWJ(t *testing.T) {
	const family = "👨\u200d👩\u200d👧"
	v := NewVT(20, 3)
	v.Write([]byte("z中 " + family + " end\x1b[2;3H\x1b[7m \x1b[0m")) //nolint:errcheck

	repaint := v.Repaint()
	v2 := NewVT(20, 3)
	v2.Write([]byte(repaint)) //nolint:errcheck

	for y := 0; y < 3; y++ {
		for x := 0; x < 20; x++ {
			c1 := v.Cell(x, y)
			c2 := v2.Cell(x, y)
			if !cellsMatch(c1, c2) {
				t.Fatalf("cell mismatch at (%d,%d): want %+v, got %+v", x, y, c1, c2)
			}
		}
	}
}

func TestVTRepaintAltScreenSwitchDirections(t *testing.T) {
	vAlt := NewVT(20, 3)
	vAlt.Write([]byte("\x1b[?1049h\x1b[1;1HALT SCREEN CONTENT")) //nolint:errcheck

	vMain := NewVT(20, 3)
	vMain.Write([]byte("\x1b[1;1HMAIN SCREEN CONTENT")) //nolint:errcheck

	vOuter := NewVT(20, 3)

	// Switch to alt tab
	vOuter.Write([]byte(vAlt.Repaint())) //nolint:errcheck
	if !vOuter.AltScreen() {
		t.Fatal("expected vOuter to enter alt screen")
	}
	if got := vOuter.Screen()[0]; got != "ALT SCREEN CONTENT" {
		t.Fatalf("vOuter alt content = %q, want ALT SCREEN CONTENT", got)
	}

	// Switch to main tab
	vOuter.Write([]byte(vMain.Repaint())) //nolint:errcheck
	if vOuter.AltScreen() {
		t.Fatal("expected vOuter to return to main screen")
	}
	if got := vOuter.Screen()[0]; got != "MAIN SCREEN CONTENT" {
		t.Fatalf("vOuter main content = %q, want MAIN SCREEN CONTENT", got)
	}
}

func TestVTRepaintCursorVisibility(t *testing.T) {
	v := NewVT(10, 2)
	v.Write([]byte("\x1b[?25l\x1b[1;1HHiddenCursor")) //nolint:errcheck
	if v.CursorVisible {
		t.Fatal("expected cursor to be hidden")
	}

	repaint := v.Repaint()
	v2 := NewVT(10, 2)
	v2.Write([]byte(repaint)) //nolint:errcheck

	if v2.CursorVisible {
		t.Fatal("expected v2 cursor to be hidden after repaint")
	}
}

