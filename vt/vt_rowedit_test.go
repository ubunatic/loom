package vt

import (
	"reflect"
	"testing"
)

func TestRowEditing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		cx, cy   int
	}{
		{
			name:     "ich_default",
			input:    "abcdef\x1b[3D\x1b[@",
			expected: []string{"abc def", "", ""},
			cx:       3, cy: 0,
		},
		{
			name:     "ich_large_n",
			input:    "abcdef\x1b[3D\x1b[20@",
			expected: []string{"abc", "", ""},
			cx:       3, cy: 0,
		},
		{
			name:     "ich_wide",
			input:    "a中b\x1b[3D\x1b[@",
			expected: []string{"a   b", "", ""},
			cx:       1, cy: 0,
		},
		{
			// A wide character shifted into the last two cells stays intact.
			name:     "ich_wide_fits_at_end",
			input:    "abcdefg中\x1b[1;1H\x1b[@",
			expected: []string{" abcdefg中", "", ""},
			cx:       0, cy: 0,
		},
		{
			name:     "dch_default",
			input:    "abcdef\x1b[4D\x1b[P",
			expected: []string{"abdef", "", ""},
			cx:       2, cy: 0,
		},
		{
			name:     "dch_large_n",
			input:    "abcdef\x1b[4D\x1b[20P",
			expected: []string{"ab", "", ""},
			cx:       2, cy: 0,
		},
		{
			name:     "dch_wide",
			input:    "a中b\x1b[3D\x1b[P",
			expected: []string{"a b", "", ""},
			cx:       1, cy: 0,
		},
		{
			name:     "ech_default",
			input:    "abcdef\x1b[4D\x1b[X",
			expected: []string{"ab def", "", ""},
			cx:       2, cy: 0,
		},
		{
			name:     "ech_large_n",
			input:    "abcdef\x1b[4D\x1b[20X",
			expected: []string{"ab", "", ""},
			cx:       2, cy: 0,
		},
		{
			// Erasing the left half of a wide character blanks both halves.
			name:     "ech_wide_left_half",
			input:    "a中b\x1b[3D\x1b[X",
			expected: []string{"a  b", "", ""},
			cx:       1, cy: 0,
		},
		{
			// Erasing the right half of a wide character blanks both halves.
			name:     "ech_wide_right_half",
			input:    "a中b\x1b[2D\x1b[X",
			expected: []string{"a  b", "", ""},
			cx:       2, cy: 0,
		},
		{
			name:     "bs_wrap",
			input:    "abcdefghij\b\b",
			expected: []string{"abcdefghij", "", ""},
			cx:       7, cy: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := NewVT(10, 3)
			v.Write([]byte(tc.input))
			screen := v.Screen()
			if !reflect.DeepEqual(screen, tc.expected) {
				t.Errorf("expected screen %q, got %q", tc.expected, screen)
			}
			x, y := v.Cursor()
			if x != tc.cx || y != tc.cy {
				t.Errorf("expected cursor %d,%d, got %d,%d", tc.cx, tc.cy, x, y)
			}
		})
	}
}
