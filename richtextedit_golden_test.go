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

func TestRichTextEditPopoverControlsMatchV2Mockup(t *testing.T) {
	data, err := os.ReadFile("docs/data/richtext-widget-v2.ansi")
	if err != nil {
		t.Fatal(err)
	}
	fixture := &RichDocument{}
	fixture.FromANSI(strings.TrimSuffix(string(data), "\n"))
	mockupText := fixture.ToPlainText()
	for _, label := range []string{"B", "I", "U", "#FG", "#BG", "Link"} {
		if !strings.Contains(mockupText, label) {
			t.Fatalf("v2 mockup is missing %q control", label)
		}
	}

	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "select me"}}}}}
	edit := NewRichTextEdit(doc)
	edit.SetSelection(RichPosition{Offset: 0}, RichPosition{Offset: 6})
	canvas := NewCanvas(48, 5)
	edit.Draw(canvas, Rect{W: 40, H: 5})
	if len(edit.popoverButtons) != len(richPopoverLabels) {
		t.Fatalf("rendered toolbar has %d controls, want %d", len(edit.popoverButtons), len(richPopoverLabels))
	}
	var rendered strings.Builder
	first, last := edit.popoverButtons[0].rect, edit.popoverButtons[len(edit.popoverButtons)-1].rect
	for x := first.X - 1; x <= last.X+last.W; x++ {
		cell := canvas.Get(x, first.Y)
		if cell.Text == "" {
			rendered.WriteByte(' ')
		} else {
			rendered.WriteString(cell.Text)
		}
	}
	for _, label := range richPopoverLabels {
		if !strings.Contains(rendered.String(), label) {
			t.Fatalf("rendered toolbar %q is missing %q", rendered.String(), label)
		}
	}
}
