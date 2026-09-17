// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPopupTitleFitsTerminalCells(t *testing.T) {
	for _, tc := range []struct {
		name  string
		title string
		width int
		want  string
	}{
		{name: "ascii", title: "Title", width: 10, want: "┌ Title ─┐"},
		{name: "accented fits despite byte length", title: "éé", width: 6, want: "┌ éé ┐"},
		{name: "accented clipped at cluster", title: "ééé", width: 6, want: "┌ ééé┐"},
		{name: "wide CJK clipped", title: "界界", width: 6, want: "┌ 界─┐"},
		{name: "emoji fits", title: "📖", width: 6, want: "┌ 📖 ┐"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			popup := NewPopup(tc.title, NewView(nil))
			popup.Width, popup.Height = tc.width, 3
			canvas := NewCanvas(tc.width, 3)
			popup.Draw(canvas, canvas.Bounds())
			got := strings.TrimSuffix(canvas.Row(0), "\x1b[0m")
			if !utf8.ValidString(got) || got != tc.want || StringWidth(got) != tc.width {
				t.Fatalf("top border = %q (valid=%v, cells=%d), want %q (%d cells)", got, utf8.ValidString(got), StringWidth(got), tc.want, tc.width)
			}
		})
	}
}
