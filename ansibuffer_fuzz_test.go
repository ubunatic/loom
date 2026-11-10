package loom

import (
	"reflect"
	"testing"
)

func FuzzAnsiBuffer(f *testing.F) {
	for _, seed := range []string{
		"a",
		"\x1b[31ma",
		"\x1b[10;10Ha",
		"a\r\n\x1b[2J",
		"\x1b[4m\x1b[m",
		"\x1b[0m",
	} {
		f.Add(seed, 80, 24)
	}
	f.Fuzz(func(t *testing.T, content string, minCols, minRows int) {
		if minCols < 0 || minCols > 1000 || minRows < 0 || minRows > 1000 {
			return
		}

		buf, err := ParseAnsiBuffer(content, minCols, minRows)
		if err != nil {
			return
		}

		ser := buf.Serialize()

		buf2, err2 := ParseAnsiBuffer(ser, minCols, minRows)
		if err2 != nil {
			t.Fatalf("Serialize gave unparsable output %q: %v", ser, err2)
		}

		for y := 0; y < buf.Rows(); y++ {
			for x := 0; x < buf.Cols(); x++ {
				c1 := buf.Get(x, y)
				var c2 AnsiCell
				if y < buf2.Rows() && x < buf2.Cols() {
					c2 = buf2.Get(x, y)
				}
				// Serialize trims trailing blanks, so a blank ' ' may come
				// back as an unset cell; both are blank.
				if c1.IsBlank() && c2.IsBlank() {
					continue
				}
				if !reflect.DeepEqual(c1, c2) {
					t.Fatalf("Cell(%d, %d) mismatch: %#v vs %#v", x, y, c1, c2)
				}
			}
		}
	})
}
