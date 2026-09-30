// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestKeyHelpRendersLabeledBindingsInStableOrder(t *testing.T) {
	km := loom.NewKeyMapWithLabels(map[string][]string{
		"save":    {"ctrl-s", "s"},
		"quit":    {"q"},
		"ignored": {"i"},
	}, map[string]string{"save": "Save", "quit": "Quit"})

	canvas := loom.NewCanvas(40, 1)
	loom.NewKeyHelp(km).Draw(canvas, canvas.Bounds())
	got := strings.TrimRight(visibleKeyHelpRow(canvas.Row(0)), " ")
	if want := "q Quit · ctrl-s Save"; got != want {
		t.Fatalf("key help = %q, want %q", got, want)
	}
}

func TestKeyHelpTruncatesAtDisplayWidth(t *testing.T) {
	km := loom.NewKeyMapWithLabels(map[string][]string{"save": {"s"}}, map[string]string{"save": "保存 changes"})
	canvas := loom.NewCanvas(8, 1)
	loom.NewKeyHelp(km).Draw(canvas, canvas.Bounds())

	row := visibleKeyHelpRow(canvas.Row(0))
	if got := loom.StringWidth(row); got != 8 {
		t.Fatalf("rendered width = %d, want 8: %q", got, row)
	}
	if want := "s 保存 …"; row != want {
		t.Fatalf("truncated help = %q, want %q", row, want)
	}
}

func visibleKeyHelpRow(row string) string {
	var text strings.Builder
	for _, cell := range loom.ParseANSI(row) {
		if !cell.Continuation {
			text.WriteString(cell.Text)
		}
	}
	return text.String()
}
