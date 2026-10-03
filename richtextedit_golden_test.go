// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
	"strings"
	"testing"

	"ubunatic.com/loom/measure"
)

func TestRichTextEditGoldenMockup(t *testing.T) {
	data, err := os.ReadFile("docs/data/richtext-widget-v1.ansi")
	if err != nil {
		t.Fatal(err)
	}
	fixture := strings.TrimSuffix(string(data), "\n")
	rows := strings.Split(fixture, "\n")
	width := 0
	for _, row := range rows {
		width = max(width, measure.StringWidth(row))
	}

	want := NewCanvas(width, len(rows))
	for y, row := range rows {
		want.WriteANSI(0, y, row)
	}
	doc := &RichDocument{}
	doc.FromANSI(fixture)
	got := NewCanvas(width, len(rows))
	edit := NewRichTextEdit(doc)
	edit.SetFocus(false)
	edit.Draw(got, got.Bounds())

	for y := range rows {
		for x := 0; x < width; x++ {
			if actual, expected := got.Get(x, y), want.Get(x, y); actual != expected {
				t.Fatalf("cell (%d,%d) = %#v, want %#v", x, y, actual, expected)
			}
		}
	}
}
