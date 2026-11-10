package vt

import (
	"testing"
)

func FuzzVTWrite(f *testing.F) {
	for _, seed := range []string{
		"a",
		"\x1b[31ma",
		"\x1b[H",
		"a\r\n\x1b[2J",
		"x\x1b[4m\x1b[m",
		"\x1b[?25h",
		"a\x1b[10;10H",
	} {
		f.Add([]byte(seed), 0)
	}
	f.Fuzz(func(t *testing.T, b []byte, split int) {
		term1 := NewVT(80, 24)
		term1.Write(b)

		screen1 := term1.Screen()
		cursorX1, cursorY1 := term1.Cursor()

		if cursorX1 < 0 || cursorX1 > 80 {
			t.Fatalf("cursorX out of bounds: %d", cursorX1)
		}
		if cursorY1 < 0 || cursorY1 > 24 {
			t.Fatalf("cursorY out of bounds: %d", cursorY1)
		}

		if split < 0 || split > len(b) {
			return
		}

		term2 := NewVT(80, 24)
		term2.Write(b[:split])
		term2.Write(b[split:])

		screen2 := term2.Screen()
		cursorX2, cursorY2 := term2.Cursor()

		if cursorX2 < 0 || cursorX2 > 80 {
			t.Fatalf("cursorX2 out of bounds: %d", cursorX2)
		}
		if cursorY2 < 0 || cursorY2 > 24 {
			t.Fatalf("cursorY2 out of bounds: %d", cursorY2)
		}

		if cursorX1 != cursorX2 || cursorY1 != cursorY2 {
			// #477 mark combining mark breakages
			t.Fatalf("477 cursor mismatch at split=%d: %d,%d vs %d,%d", split, cursorX1, cursorY1, cursorX2, cursorY2)
		}

		if !stringsEqual(screen1, screen2) {
			t.Fatalf("477 screen mismatch at split=%d: %q vs %q", split, screen1, screen2)
		}
	})
}

func stringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
