package loomtest

import (
	"strings"
	"testing"
)

func TestCompareANSIReport(t *testing.T) {
	tests := []struct {
		name string
		want string
		got  string
		diff bool
	}{
		{"bold diff", "\x1b[1mB\x1b[m", "B", true},
		{"dim diff", "\x1b[2mD\x1b[m", "D", true},
		{"same", "A", "A", false},
		{"wide char diff", "こ", "ん", true}, // Need to test that continuation cells display correctly
		{"wide vs ascii diff", "こ", "ab", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			report := compareANSIReport(tc.want, tc.got, 2, 1, false)
			hasDiff := !strings.Contains(report, "differences: 0\n")
			if hasDiff != tc.diff {
				t.Errorf("expected diff=%v, got report:\n%s", tc.diff, report)
			}
			t.Log(report)
		})
	}
}
