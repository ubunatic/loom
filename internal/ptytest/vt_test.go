// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ptytest

import (
	"slices"
	"testing"
)

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
	// Split a multi-byte rune and a CSI across writes.
	for _, chunk := range []string{"a\xe4", "\xb8\xad\x1b[", "1;1Hz"} {
		v.Write([]byte(chunk)) //nolint:errcheck
	}
	if got := v.Screen()[0]; got != "z 中" && got != "z中" {
		t.Fatalf("row = %q", got)
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

func TestSGRReset(t *testing.T) {
	v := NewVT(10, 1)
	v.Write([]byte("\x1b[1;31mhi\x1b[0mthere\x1b[1;32m!\x1b[m?")) //nolint:errcheck

	// 'h', 'i' should be Bold and Red (index 1)
	for x := 0; x < 2; x++ {
		c := v.Cell(x, 0)
		if !c.Style.Bold || c.Style.FG != ColorIndex(1) {
			t.Errorf("cell(%d,0) = %+v, want bold+fg1", x, c)
		}
	}
	// 'there' should be default style
	for x := 2; x < 7; x++ {
		c := v.Cell(x, 0)
		if c.Style != (Style{}) {
			t.Errorf("cell(%d,0) = %+v, want default style", x, c)
		}
	}
	// '!' should be bold and green (index 2)
	c := v.Cell(7, 0)
	if !c.Style.Bold || c.Style.FG != ColorIndex(2) {
		t.Errorf("cell(7,0) = %+v, want bold+fg2", c)
	}
	// '?' reset via \x1b[m
	c = v.Cell(8, 0)
	if c.Style != (Style{}) {
		t.Errorf("cell(8,0) = %+v, want default style after \\x1b[m", c)
	}
}

func TestSGR16AndBrightColors(t *testing.T) {
	v := NewVT(16, 4)
	// Standard FG (30-37)
	for i := 0; i < 8; i++ {
		seq := string([]byte{0x1b, '[', '3', byte('0' + i), 'm', 'F'})
		v.Write([]byte(seq)) //nolint:errcheck
	}
	// Bright FG (90-97)
	for i := 0; i < 8; i++ {
		seq := string([]byte{0x1b, '[', '9', byte('0' + i), 'm', 'B'})
		v.Write([]byte(seq)) //nolint:errcheck
	}
	// Next line: standard BG (40-47)
	v.Write([]byte("\r\n")) //nolint:errcheck
	for i := 0; i < 8; i++ {
		seq := string([]byte{0x1b, '[', '4', byte('0' + i), 'm', 'G'})
		v.Write([]byte(seq)) //nolint:errcheck
	}
	// Bright BG (100-107)
	for i := 0; i < 8; i++ {
		v.Write([]byte("\x1b[10" + string(rune('0'+i)) + "mH")) //nolint:errcheck
	}

	for i := 0; i < 8; i++ {
		c := v.Cell(i, 0)
		if c.Style.FG != ColorIndex(uint8(i)) {
			t.Errorf("cell(%d,0) FG = %v, want index %d", i, c.Style.FG, i)
		}
	}
	for i := 0; i < 8; i++ {
		c := v.Cell(8+i, 0)
		if c.Style.FG != ColorIndex(uint8(8+i)) {
			t.Errorf("cell(%d,0) FG = %v, want index %d", 8+i, c.Style.FG, 8+i)
		}
	}
	for i := 0; i < 8; i++ {
		c := v.Cell(i, 1)
		if c.Style.BG != ColorIndex(uint8(i)) {
			t.Errorf("cell(%d,1) BG = %v, want index %d", i, c.Style.BG, i)
		}
	}
	for i := 0; i < 8; i++ {
		c := v.Cell(8+i, 1)
		if c.Style.BG != ColorIndex(uint8(8+i)) {
			t.Errorf("cell(%d,1) BG = %v, want index %d", 8+i, c.Style.BG, 8+i)
		}
	}

	// Test 39/49 reset
	v.Write([]byte("\x1b[39m\x1b[49m\r\nZ")) //nolint:errcheck
	c := v.Cell(0, 2)
	if c.Style.FG.Type != ColorDefault || c.Style.BG.Type != ColorDefault {
		t.Errorf("cell(0,2) style after 39/49 = %+v, want default FG/BG", c.Style)
	}
}

func TestSGR256Colors(t *testing.T) {
	v := NewVT(10, 1)
	v.Write([]byte("\x1b[38;5;123;48;5;201mX")) //nolint:errcheck
	c := v.Cell(0, 0)
	if c.Style.FG != ColorIndex(123) {
		t.Errorf("FG = %v, want index(123)", c.Style.FG)
	}
	if c.Style.BG != ColorIndex(201) {
		t.Errorf("BG = %v, want index(201)", c.Style.BG)
	}
	r, g, b, ok := c.Style.FG.RGB()
	if !ok || (r == 0 && g == 0 && b == 0) {
		t.Errorf("FG.RGB() = (%d,%d,%d,%v), want non-zero resolved RGB", r, g, b, ok)
	}
}

func TestSGRTruecolor(t *testing.T) {
	v := NewVT(10, 1)
	v.Write([]byte("\x1b[38;2;12;34;56;48;2;78;90;120mX")) //nolint:errcheck
	c := v.Cell(0, 0)
	if c.Style.FG != ColorRGB(12, 34, 56) {
		t.Errorf("FG = %v, want rgb(12,34,56)", c.Style.FG)
	}
	if c.Style.BG != ColorRGB(78, 90, 120) {
		t.Errorf("BG = %v, want rgb(78,90,120)", c.Style.BG)
	}
	r, g, b, ok := c.Style.FG.RGB()
	if !ok || r != 12 || g != 34 || b != 56 {
		t.Errorf("FG.RGB() = (%d,%d,%d,%v), want (12,34,56,true)", r, g, b, ok)
	}
}

func TestSGRCombinedParams(t *testing.T) {
	v := NewVT(10, 1)
	v.Write([]byte("\x1b[1;3;4;38;2;10;20;30;48;5;42;7mC")) //nolint:errcheck
	c := v.Cell(0, 0)
	want := Style{
		Bold:      true,
		Italic:    true,
		Underline: true,
		FG:        ColorRGB(10, 20, 30),
		BG:        ColorIndex(42),
		Reverse:   true,
	}
	if c.Style != want {
		t.Errorf("cell style = %+v, want %+v", c.Style, want)
	}
}

func TestSGRReverse(t *testing.T) {
	v := NewVT(10, 1)
	v.Write([]byte("\x1b[7mRev\x1b[27mNorm")) //nolint:errcheck
	for x := 0; x < 3; x++ {
		c := v.Cell(x, 0)
		if !c.Style.Reverse {
			t.Errorf("cell(%d,0).Reverse = false, want true", x)
		}
	}
	for x := 3; x < 7; x++ {
		c := v.Cell(x, 0)
		if c.Style.Reverse {
			t.Errorf("cell(%d,0).Reverse = true, want false", x)
		}
	}
}

func TestSGRWideRune(t *testing.T) {
	v := NewVT(10, 1)
	v.Write([]byte("\x1b[38;2;255;0;0;48;2;0;255;0m🚀\x1b[0m")) //nolint:errcheck
	left := v.Cell(0, 0)
	right := v.Cell(1, 0)

	wantStyle := Style{
		FG: ColorRGB(255, 0, 0),
		BG: ColorRGB(0, 255, 0),
	}
	if left.Rune != '🚀' || left.Style != wantStyle {
		t.Errorf("left cell = %+v, want rune='🚀' style=%+v", left, wantStyle)
	}
	if right.Rune != 0 || right.Style != wantStyle {
		t.Errorf("right cell = %+v, want rune=0 style=%+v", right, wantStyle)
	}
}

func TestSGREraseWithBG(t *testing.T) {
	v := NewVT(6, 2)
	// Set blue BG, write "ab", erase rest of line
	v.Write([]byte("\x1b[44mab\x1b[K")) //nolint:errcheck

	blueStyle := Style{BG: ColorIndex(4)}
	for x := 2; x < 6; x++ {
		c := v.Cell(x, 0)
		if c.Rune != ' ' || c.Style != blueStyle {
			t.Errorf("erased cell(%d,0) = %+v, want ' ' with blue BG", x, c)
		}
	}

	// Test eraseDisplay with truecolor BG
	v.Write([]byte("\x1b[48;2;10;20;30m\x1b[2J")) //nolint:errcheck
	tcStyle := Style{BG: ColorRGB(10, 20, 30)}
	for y := 0; y < 2; y++ {
		for x := 0; x < 6; x++ {
			c := v.Cell(x, y)
			if c.Rune != ' ' || c.Style != tcStyle {
				t.Fatalf("cell(%d,%d) = %+v, want ' ' with truecolor BG", x, y, c)
			}
		}
	}
}

func TestSGRSplitAcrossWrites(t *testing.T) {
	v := NewVT(10, 1)
	chunks := []string{
		"\x1b[38;2;255;",
		"128;64mSplit",
		"\x1b[48;",
		"5;99m256",
	}
	for _, chunk := range chunks {
		v.Write([]byte(chunk)) //nolint:errcheck
	}

	splitCell := v.Cell(0, 0)
	if splitCell.Rune != 'S' || splitCell.Style.FG != ColorRGB(255, 128, 64) {
		t.Errorf("cell(0,0) = %+v, want 'S' with FG rgb(255,128,64)", splitCell)
	}

	c256 := v.Cell(5, 0)
	if c256.Rune != '2' || c256.Style.FG != ColorRGB(255, 128, 64) || c256.Style.BG != ColorIndex(99) {
		t.Errorf("cell(5,0) = %+v, want '2' with FG rgb(255,128,64) and BG index(99)", c256)
	}
}

func TestSGRFrameSnapshots(t *testing.T) {
	v := NewVT(5, 1)
	v.Write([]byte("\x1b[?2026h\x1b[1;1H\x1b[31mA\x1b[?2026l\x1b[?2026h\x1b[1;1H\x1b[32mB\x1b[?2026l")) //nolint:errcheck
	if len(v.CellFrames) != 2 {
		t.Fatalf("len(CellFrames) = %d, want 2", len(v.CellFrames))
	}
	frame0 := v.CellFrames[0]
	frame1 := v.CellFrames[1]
	if frame0[0][0].Rune != 'A' || frame0[0][0].Style.FG != ColorIndex(1) {
		t.Errorf("frame 0 cell(0,0) = %+v, want 'A' with red FG", frame0[0][0])
	}
	if frame1[0][0].Rune != 'B' || frame1[0][0].Style.FG != ColorIndex(2) {
		t.Errorf("frame 1 cell(0,0) = %+v, want 'B' with green FG", frame1[0][0])
	}
}
