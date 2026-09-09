// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestTruncateTextBudgets(t *testing.T) {
	for _, tc := range []struct {
		text, marker string
		width        int
		want         string
	}{
		{"long", "...", -1, ""},
		{"long", "...", 0, ""},
		{"long", "...", 1, "."},
		{"long", "...", 2, ".."},
		{"long", "...", 3, "..."},
		{"long", "...", 4, "long"},
		{"123456", "...", 5, "12..."},
		{"e\u0301中ABC", ".", 4, "e\u0301中."},
		{"中中中", ".", 4, "中."},
		{"ABC", "中", 1, "A"},
		{"\x1b[1m99%\x1b[0m", "...", 3, "99%"},
		{"\x1b[31m100%\x1b[0m", ".", 3, "10."},
		{"\u0301A", "...", 1, "A"},
		{"⣿⣀█░", "", 2, "⣿⣀"},
	} {
		t.Run(tc.text, func(t *testing.T) {
			got := TruncateText(tc.text, tc.width, tc.marker)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			cells, err := oracleCells(got)
			if err != nil || len(cells) > max(0, tc.width) {
				t.Fatalf("independent width check: %v %v", cells, err)
			}
		})
	}
}
