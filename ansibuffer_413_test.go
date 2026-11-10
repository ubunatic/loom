package loom

import (
	"testing"
	"strings"
)

func Test413_ParseAnsiBuffer(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		check   func(t *testing.T, b *AnsiBuffer)
		minRows int
		minCols int
	}{
		{
			name:    "CSI A B D relative moves",
			input:   "X\x1b[B\x1b[1DY\x1b[1A\x1b[1CZ",
			minRows: 1,
			minCols: 1,
			check: func(t *testing.T, b *AnsiBuffer) {
				if b.Rows() < 2 {
					t.Fatalf("expected at least 2 rows, got %d", b.Rows())
				}
				if b.Cols() < 2 {
					t.Fatalf("expected at least 2 cols, got %d", b.Cols())
				}
				if r := b.Get(0, 0).Rune; r != 'X' {
					t.Errorf("expected X at 0,0, got %q", r)
				}
				if r := b.Get(0, 1).Rune; r != 'Y' {
					t.Errorf("expected Y at 0,1 (x=0, y=1), got %q", r)
				}
				if r := b.Get(2, 0).Rune; r != 'Z' {
					t.Errorf("expected Z at 2,0 (x=2, y=0), got %q", r)
				}
			},
		},
		{
			name:    "Prescan bounds check B",
			input:   "\x1b[10BXX",
			minRows: 1,
			minCols: 1,
			check: func(t *testing.T, b *AnsiBuffer) {
				if b.Rows() < 11 {
					t.Fatalf("expected 11 rows due to 10B, got %d", b.Rows())
				}
				if r := b.Get(0, 10).Rune; r != 'X' {
					t.Errorf("expected X at 0,10, got %q", r)
				}
			},
		},
		{
			name:    "Colon truecolor SGR 38:2::r:g:b",
			input:   "\x1b[38:2::255:128:64mA",
			minRows: 1,
			minCols: 1,
			check: func(t *testing.T, b *AnsiBuffer) {
				c := b.Get(0, 0)
				if c.FG != ColorRGB(255, 128, 64) {
					t.Errorf("expected fg #ff8040, got %v", c.FG)
				}
			},
		},
		{
			name:    "Colon truecolor SGR 48:2:r:g:b",
			input:   "\x1b[48:2:10:20:30mA",
			minRows: 1,
			minCols: 1,
			check: func(t *testing.T, b *AnsiBuffer) {
				c := b.Get(0, 0)
				if c.BG != ColorRGB(10, 20, 30) {
					t.Errorf("expected bg #0a141e, got %v", c.BG)
				}
			},
		},
		{
			name:    "Wide cell continuation no extra space",
			input:   "字A",
			minRows: 1,
			minCols: 1,
			check: func(t *testing.T, b *AnsiBuffer) {
				if b.Cols() < 3 {
					t.Fatalf("expected at least 3 cols, got %d", b.Cols())
				}
				if r := b.Get(0, 0).Rune; r != '字' {
					t.Errorf("expected 字 at 0,0, got %q", r)
				}
				if r := b.Get(1, 0).Rune; r != 0 {
					t.Errorf("expected continuation cell (0) at 1,0, got %q", r)
				}
				if r := b.Get(2, 0).Rune; r != 'A' {
					t.Errorf("expected A at 2,0, got %q", r)
				}
				ser := b.Serialize()
				if strings.Contains(ser, "字 A") {
					t.Errorf("Serialize emitted extra space for wide char: %q", ser)
				}
				if ser != "字A" {
					t.Errorf("expected Serialize to be '字A', got %q", ser)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := ParseAnsiBuffer(tc.input, tc.minCols, tc.minRows)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tc.check(t, b)
		})
	}
}
