// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
	"strings"
	"testing"
)

func TestRichDocumentFromANSIAndPlainText(t *testing.T) {
	var doc RichDocument
	doc.FromANSI("A\x1b[1;3;4;9;7;38;5;202m界e\u0301\x1b[22;23;24;29;27m!\r\n\nend\n")
	if got, want := doc.ToPlainText(), "A界e\u0301!\n\nend"; got != want {
		t.Fatalf("plain text = %q, want %q", got, want)
	}
	if len(doc.Lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(doc.Lines))
	}
	styled := doc.Lines[0].Spans[1]
	if styled.Text != "界e\u0301" || !styled.Style.Bold || !styled.Style.Italic || !styled.Style.Underline || !styled.Style.Strike || !styled.Style.Invert || styled.Style.FG != ColorIndex(202) {
		t.Fatalf("styled span = %#v", styled)
	}
	if len(doc.Lines[1].Spans) != 0 {
		t.Fatalf("empty line has spans: %#v", doc.Lines[1])
	}
}

func TestRichDocumentANSIStyleRoundTrip(t *testing.T) {
	input := "plain \x1b[1;2;3;4;7;9;38;2;12;34;56;48;5;23mrich\x1b[0m\nlast"
	var original, parsed RichDocument
	original.FromANSI(input)
	parsed.FromANSI(original.ToANSI())
	if got, want := parsed.ToPlainText(), original.ToPlainText(); got != want {
		t.Fatalf("plain text round trip = %q, want %q", got, want)
	}
	if len(parsed.Lines[0].Spans) != 2 || parsed.Lines[0].Spans[1].Style != original.Lines[0].Spans[1].Style {
		t.Fatalf("styled spans did not round trip: %#v", parsed.Lines[0].Spans)
	}
	if !strings.HasSuffix(original.ToANSI(), "\x1b[0m") {
		t.Fatal("ANSI output must end with a style reset")
	}
}

func TestRichDocumentConsumesControlsAndNormalizesLines(t *testing.T) {
	var doc RichDocument
	doc.FromANSI("one\x1b[2Atwo\x1b]0;title\a\rthree\x1b[31")
	if got, want := doc.ToPlainText(), "onetwo\nthree"; got != want {
		t.Fatalf("plain text = %q, want %q", got, want)
	}
}

func TestRichDocumentPreservesPillMetadataInMemory(t *testing.T) {
	doc := RichDocument{Lines: []RichLine{{Spans: []RichSpan{
		{Text: "assign "},
		{Text: "@team-core", Style: Style{Bold: true}, PillData: &RichPill{Kind: "mention", ID: "team-core", Metadata: map[string]string{"team": "core"}}},
	}}}}
	if got := doc.ToPlainText(); got != "assign @team-core" {
		t.Fatalf("plain text = %q", got)
	}
	if got := doc.ToANSI(); !strings.Contains(got, "@team-core") {
		t.Fatalf("ANSI output lost pill label: %q", got)
	}
	if doc.Lines[0].Spans[1].PillData.Metadata["team"] != "core" {
		t.Fatal("pill metadata changed")
	}
}

func TestRichDocumentFromANSITrailingNewlineSymmetric(t *testing.T) {
	// Single line with trailing SGR and newline
	var doc1 RichDocument
	doc1.FromANSI("hello\x1b[0m\n")
	if len(doc1.Lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(doc1.Lines))
	}
	if got, want := doc1.ToPlainText(), "hello"; got != want {
		t.Fatalf("doc1 plain text = %q, want %q", got, want)
	}

	// Document ending in an empty line (hello\n)
	var doc2 RichDocument
	doc2.FromANSI("hello\n\x1b[0m\n")
	if len(doc2.Lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(doc2.Lines))
	}
	if got, want := doc2.ToPlainText(), "hello\n"; got != want {
		t.Fatalf("doc2 plain text = %q, want %q", got, want)
	}
}

func TestRichDocumentImportsMockups(t *testing.T) {
	for _, name := range []string{"richtext-widget-v1.ansi", "richtext-widget-v2.ansi", "richtext-widget-v3.ansi"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile("docs/data/" + name)
			if err != nil {
				t.Fatal(err)
			}
			var doc RichDocument
			doc.FromANSI(string(data))
			if !strings.Contains(doc.ToPlainText(), "RichTextEdit:") {
				t.Fatalf("mockup heading missing from plain text: %q", doc.ToPlainText())
			}
			if len(doc.Lines) == 0 || len(doc.ToANSI()) == 0 {
				t.Fatal("mockup did not import and serialize")
			}
		})
	}
}
