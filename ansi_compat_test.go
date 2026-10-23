// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"
	"ubunatic.com/loom"
	"ubunatic.com/loom/ansi"
)

func TestANSITypeAliases(t *testing.T) {
	var color ansi.Color = loom.ColorRGB(1, 2, 3)
	var style ansi.Style = loom.Style{FG: color, Bold: true}
	var profile ansi.ColorProfile = loom.ColorProfileTrueColor
	var rootStyle loom.Style = style
	if rootStyle.ANSIFor(profile) != style.ANSI() {
		t.Fatal("aliases changed styling")
	}
	canvas := loom.NewCanvas(3, 1)
	canvas.ColorProfile = profile
	canvas.Set(0, 0, loom.Cell{Text: "a", Style: style})
	canvas.Set(1, 0, loom.Cell{Text: "b", Style: style})
	canvas.Set(2, 0, loom.Cell{Text: "c"})
	want := style.ANSI() + "ab" + ansi.Reset.ANSI() + "c\x1b[0m"
	if got := canvas.Row(0); got != want {
		t.Fatalf("row = %q, want %q", got, want)
	}
}
