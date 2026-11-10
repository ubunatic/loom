package vt

import "testing"

// TestSplitEveryByteMatchesWholeWrite checks that a stream written in two
// parts, split at any byte, gives the same screen and cursor as one write
// (issue 477: a combining mark, ZWJ tail or VS16 that starts a write must
// extend the cell before the cursor).
func TestSplitEveryByteMatchesWholeWrite(t *testing.T) {
	streams := map[string]string{
		"decomposed":   "Qéx",
		"sgr-cup-wide": "\x1b[31mA\x1b[2;3H中é\x1b[0m!",
		"zwj":          "a👨‍🌾b",
		"vs16":         "❤️z",
		"skin":         "👍🏽k",
		"crlf-mark":    "ab\r\ńc",
		"wrap-mark":    "abcdéf",
		"scroll":       "1\r\n2\r\n3\r\né",
		"utf8-tail":    "0\xc2\xa9",
	}
	for name, s := range streams {
		whole := NewVT(5, 3)
		whole.Write([]byte(s))
		ws := whole.Screen()
		wx, wy := whole.Cursor()
		for i := 1; i < len(s); i++ {
			v := NewVT(5, 3)
			v.Write([]byte(s[:i]))
			v.Write([]byte(s[i:]))
			x, y := v.Cursor()
			got := v.Screen()
			same := x == wx && y == wy
			for k := range ws {
				same = same && ws[k] == got[k]
			}
			if !same {
				t.Errorf("%s split %d: %q cursor %d,%d; whole write %q cursor %d,%d", name, i, got, x, y, ws, wx, wy)
			}
		}
		v := NewVT(5, 3)
		for i := 0; i < len(s); i++ {
			v.Write([]byte{s[i]})
		}
		x, y := v.Cursor()
		got := v.Screen()
		same := x == wx && y == wy
		for k := range ws {
			same = same && ws[k] == got[k]
		}
		if !same {
			t.Errorf("%s byte by byte: %q cursor %d,%d; whole write %q cursor %d,%d", name, got, x, y, ws, wx, wy)
		}
	}
}
