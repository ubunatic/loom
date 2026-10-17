package loom

import (
	"strings"
	"testing"
	"unicode"

	"ubunatic.com/loom/measure"
)

// zwjControlCases pairs a ZWJ with a control payload that must never reach
// the terminal raw (issue: ZWJ joined the following rune unconditionally).
var zwjControlCases = []struct {
	name, text string
}{
	{"csi-clear", "a‍\x1b[2J"},
	{"osc52", "a‍\x1b]52;c;aGVsbG8=\a"},
	{"newline", "a‍\nb"},
	{"cr", "a‍\rb"},
	{"bel", "a‍\ab"},
	{"c1-csi", "a‍\u009b2J"},
	{"del", "a‍\x7fb"},
	{"bare-zwj-esc", "‍\x1b[2J"},
}

func hasControl(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func TestZWJControlNotJoinedParseANSI(t *testing.T) {
	for _, tc := range zwjControlCases {
		for name, parse := range map[string]func(string) []Cell{"ParseANSI": ParseANSI, "ParseANSIOld": ParseANSIOld} {
			for _, c := range parse(tc.text) {
				if hasControl(c.Text) {
					t.Errorf("%s(%s): cell text %q contains a control", name, tc.name, c.Text)
				}
			}
		}
	}
}

func TestZWJControlNotJoinedClipRow(t *testing.T) {
	for _, tc := range zwjControlCases {
		out := ClipRow(tc.text, 20)
		// ClipRow passes well-formed CSI through by design; strip those and
		// check that nothing else carries a control.
		plain := stripCSI(out)
		if hasControl(plain) {
			t.Errorf("ClipRow(%s) = %q leaks a control outside CSI", tc.name, out)
		}
		if tc.name == "osc52" && strings.Contains(out, "\x1b]") {
			t.Errorf("ClipRow(%s) = %q emits an OSC sequence", tc.name, out)
		}
	}
}

func stripCSI(s string) string {
	rs := []rune(s)
	var b strings.Builder
	for i := 0; i < len(rs); i++ {
		if rs[i] == 27 && i+1 < len(rs) && rs[i+1] == '[' {
			for i += 2; i < len(rs) && !(rs[i] >= '@' && rs[i] <= '~'); i++ {
			}
			continue
		}
		b.WriteRune(rs[i])
	}
	return b.String()
}

func TestZWJControlNotJoinedClusters(t *testing.T) {
	for _, tc := range zwjControlCases {
		for name, fn := range map[string]func(string) []string{"Clusters": measure.Clusters, "ClustersOld": measure.ClustersOld, "ClustersNew": measure.ClustersNew} {
			for _, cl := range fn(tc.text) {
				if hasControl(cl) {
					t.Errorf("%s(%s): cluster %q contains a control", name, tc.name, cl)
				}
			}
		}
	}
}

func TestZWJControlCanvasRow(t *testing.T) {
	for _, tc := range zwjControlCases {
		for _, write := range []string{"Write", "WriteANSI"} {
			c := NewCanvas(20, 1)
			if write == "Write" {
				c.Write(0, 0, tc.text, Style{})
			} else {
				c.WriteANSI(0, 0, tc.text)
			}
			if row := stripCSI(c.Row(0)); hasControl(row) {
				t.Errorf("Canvas.%s(%s).Row = %q leaks a control", write, tc.name, row)
			}
		}
	}
}

// Defense in depth: Row must not emit control bytes even when a cell's text
// bypassed Set's clustering (e.g. a future write path that stores raw text).
func TestCanvasRowDropsRawControlCellText(t *testing.T) {
	for _, text := range []string{"\x1b[2J", "a\u200d\x1b]52;c;x\a", "\n", "\r", "\a", "\u009b2J", "\x7f", "x\x00y"} {
		c := NewCanvas(4, 1)
		c.cells[0][0].Text = text // bypass Set on purpose
		row := c.Row(0)
		if hasControl(stripCSI(row)) || strings.Contains(row, "\x1b]") || strings.Contains(row, "\x1b[2J") {
			t.Errorf("Row with raw cell %q = %q emits control bytes", text, row)
		}
	}
}

// Even a cell set through Set with control bytes must not emit them.
func TestCanvasRowSanitizesCellText(t *testing.T) {
	for _, text := range []string{"\x1b[2J", "a‍\x1b]52;c;x\a", "\n", "\r", "\a", "\u009b", "\x7f", "x\x00y"} {
		c := NewCanvas(4, 1)
		c.Set(0, 0, Cell{Text: text})
		row := c.Row(0)
		if hasControl(stripCSI(row)) || strings.Contains(row, "\x1b]") || strings.Contains(row, "\x1b[2J") {
			t.Errorf("Row with cell %q = %q emits control bytes", text, row)
		}
		if got := measure.StringWidth(stripCSI(row)); got != 4 {
			t.Errorf("Row with cell %q has width %d, want 4", text, got)
		}
	}
}

func TestZWJEmojiStillOneCluster(t *testing.T) {
	for _, mode := range []string{"join", "split"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("LOOM_ZWJ", mode)
			for i, e := range []string{"👨‍👩‍👧", "🏳️‍🌈", "❤️‍🔥"} {
				cells := ParseANSI(e)
				if len(cells) == 0 || cells[0].Text != e {
					t.Errorf("ParseANSI(%q) first cell = %+v, want whole cluster", e, cells)
				}
				if cl := measure.Clusters(e); len(cl) != 1 || cl[0] != e {
					t.Errorf("Clusters(%q) = %q, want one cluster", e, cl)
				}
				if got := ClipRow(e, 10); got != e {
					t.Errorf("ClipRow(%q) = %q", e, got)
				}
				wantWidth := 2
				if mode == "split" {
					wantWidth = []int{6, 4, 4}[i]
				}
				if w := measure.StringWidth(e); w != wantWidth {
					t.Errorf("StringWidth(%q) in %s mode = %d, want %d", e, mode, w, wantWidth)
				}
			}
		})
	}
}
