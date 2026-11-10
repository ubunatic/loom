package measure

import "testing"

// TestWidthTargets pins the widths decided in 481 (Jules sweep): VTE/Tilix,
// foot and kitty draw these as listed. A ZWJ that joins nothing adds no
// column, in either LOOM_ZWJ mode.
func TestWidthTargets(t *testing.T) {
	cases := []struct {
		in       string
		width    int
		clusters int
	}{
		{"️", 0, 1},
		{"a️", 1, 1},
		{"#️", 2, 1},
		{"❤️", 2, 1},
		{"👍🏽", 2, 1},
		{"a👍🏽b", 4, 3},
		{"क्ष", 2, 1},
		{"क्षa", 3, 2},
		{"ক্ষ", 2, 1},
		{"नमस्ते", 4, 3},
		{"0‍", 1, 1},
		{"‍", 0, 1},
		{"0\x7f‍", 1, 1},
	}
	for _, mode := range []int32{zwjSplit, zwjJoin} {
		old := detectedZWJMode.Swap(mode)
		for _, c := range cases {
			if got := StringWidth(c.in); got != c.width {
				t.Errorf("mode %d: StringWidth(%q) = %d, want %d", mode, c.in, got, c.width)
			}
			if got := len(Clusters(c.in)); got != c.clusters {
				t.Errorf("mode %d: len(Clusters(%q)) = %d, want %d", mode, c.in, got, c.clusters)
			}
		}
		detectedZWJMode.Store(old)
	}
}
