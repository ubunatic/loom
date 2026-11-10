package vtdiff

import (
	"strings"
	"testing"

	upstream "github.com/charmbracelet/x/vt"
	loom "ubunatic.com/loom/vt"
)

func upstreamText(u *upstream.Emulator, cols, rows int) []string {
	out := make([]string, rows)
	for y := 0; y < rows; y++ {
		var b strings.Builder
		for x := 0; x < cols; x++ {
			c := u.CellAt(x, y)
			if c != nil && c.Content != "" && c.Content != "\x00" {
				b.WriteString(c.Content)
				if c.Width > 1 {
					x += c.Width - 1
				}
			} else {
				b.WriteString(" ")
			}
		}
		out[y] = strings.TrimRight(b.String(), " ")
	}
	return out
}

func compare(t *testing.T, s string) {
	cols, rows := 10, 4

	runScenario := func(t *testing.T, prefix string, writeFunc func(l *loom.VT, u *upstream.Emulator)) {
		l := loom.NewVT(cols, rows)
		u := upstream.NewEmulator(cols, rows)
		writeFunc(l, u)

		lScreen := l.Screen()
		uScreen := upstreamText(u, cols, rows)
		lx, ly := l.Cursor()
		uPos := u.CursorPosition()
		ux, uy := uPos.X, uPos.Y

		diff := false
		for i := 0; i < rows; i++ {
			if lScreen[i] != uScreen[i] {
				diff = true
			}
		}
		if lx != ux || ly != uy {
			diff = true
		}

		if diff {
			t.Errorf("%s:\nLoom:     cursor(%d,%d) screen:%#v\nUpstream: cursor(%d,%d) screen:%#v", prefix, lx, ly, lScreen, ux, uy, uScreen)
		}
	}

	t.Run("whole", func(t *testing.T) {
		runScenario(t, "whole", func(l *loom.VT, u *upstream.Emulator) {
			l.Write([]byte(s))
			u.Write([]byte(s))
		})
	})

	t.Run("split", func(t *testing.T) {
		for i := 1; i < len(s); i++ {
			runScenario(t, "split", func(l *loom.VT, u *upstream.Emulator) {
				l.Write([]byte(s[:i]))
				l.Write([]byte(s[i:]))
				u.Write([]byte(s[:i]))
				u.Write([]byte(s[i:]))
			})
		}
	})

	t.Run("byte-by-byte", func(t *testing.T) {
		runScenario(t, "byte-by-byte", func(l *loom.VT, u *upstream.Emulator) {
			for i := 0; i < len(s); i++ {
				l.Write([]byte{s[i]})
				u.Write([]byte{s[i]})
			}
		})
	})
}

func TestVTDiff(t *testing.T) {
	streams := map[string]string{
		"plain text": "hello world",
		"CR/LF": "hello\r\nworld",
		"wrap at the last column": "1234567890X",
		"scroll": "1\n2\n3\n4\n5",
		"SGR colours": "\x1b[31mred\x1b[0m",
		"CUP/CUU/CUD/CUF/CUB": "A\x1b[2;2HB\x1b[1AC\x1b[1BD\x1b[1CE\x1b[1DF",
		"erase in line and display": "ABC\x1b[2J\x1b[HXYZ\x1b[1K\x1b[2K",
		"insert/delete chars and lines": "A\x1b[1@B\x1b[1PC\x1b[1LD\x1b[1ME",
		"scroll regions": "\x1b[2;3r1\n2\n3\n4",
		"tabs": "a\tb",
		"wide CJK": "中国",
		"combining marks": "e\u0301",
		"ZWJ emoji": "👨‍🌾",
		"VS16": "❤️",
		"skin tones": "👍🏽",
		"incomplete UTF-8": "a\xc2",
		"alt screen": "main\x1b[?1049halt",

		// 8 more to reach 25
		"CUP wrap": "\x1b[5;10Hwrap",
		"erase display 0": "123\x1b[J",
		"erase display 1": "123\x1b[1J",
		"erase line 0": "123\x1b[K",
		"erase line 1": "123\x1b[1K",
		"reverse wrap": "abcdefghij\b\b",
		"tab stop": "\x1b[H\x1b[g\t\x1b[g\t",
		"DEC line drawing": "\x1b(0lqqk\x1b(B",
	}
	for name, s := range streams {
		t.Run(name, func(t *testing.T) {
			compare(t, s)
		})
	}
}
