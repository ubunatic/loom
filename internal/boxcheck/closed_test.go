// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package boxcheck

import (
	"strings"
	"testing"
)

func TestClosed(t *testing.T) {
	for _, tc := range []struct{ name, text, want string }{
		{"missing corner", "┌───┐\n│abc│\n└─── ", "missing bottom-right"},
		{"missing border", "┌───┐\n│abc│\n", "missing bottom border"},
		{"double border with color", "\x1b[31m╔═══╗\x1b[0m\n║abc║\n╚═══ ", "missing bottom-right"},
		{"closed", "┌───┐\n│a+b│\n└───┘", ""},
		{"misaligned closed", "┌───┐\n│abc│\n└──┘", ""},
		{"partial diagram", "│ab│ ┌──┐\n│cd│ └──┘\n│ef│", ""},
		{"cursor recording", "┌───┐\n\x1b[3;1H│abc│", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Closed(tc.text)
			if tc.want == "" && err != nil || tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("Closed() = %v, want %q", err, tc.want)
			}
		})
	}
}
