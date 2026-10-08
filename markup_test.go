// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func markupSpans(t *testing.T, src string) []RichSpan {
	t.Helper()
	doc, err := ParseMarkup(src)
	if err != nil {
		t.Fatalf("ParseMarkup(%q): %v", src, err)
	}
	if len(doc.Lines) != 1 {
		t.Fatalf("ParseMarkup(%q) lines = %d, want 1", src, len(doc.Lines))
	}
	return doc.Lines[0].Spans
}

func TestMarkupParseAttributesAndColors(t *testing.T) {
	tests := []struct {
		src  string
		want Style
	}{
		{"[bold]x[/]", Style{Bold: true}},
		{"[b i u s]x[/]", Style{Bold: true, Italic: true, Underline: true, Strike: true}},
		{"[dim invert]x[/]", Style{Dim: true, Invert: true}},
		{"[reverse]x[/]", Style{Invert: true}},
		{"[red]x[/]", Style{FG: ColorIndex(1)}},
		{"[bright_cyan]x[/]", Style{FG: ColorIndex(14)}},
		{"[GREY]x[/]", Style{FG: ColorIndex(8)}},
		{"[color(208)]x[/]", Style{FG: ColorIndex(208)}},
		{"[#ff8000]x[/]", Style{FG: ColorRGB(0xff, 0x80, 0)}},
		{"[#F80]x[/]", Style{FG: ColorRGB(0xff, 0x88, 0)}},
		{"[on blue]x[/]", Style{BG: ColorIndex(4)}},
		{"[bold red on #202020]x[/]", Style{Bold: true, FG: ColorIndex(1), BG: ColorRGB(0x20, 0x20, 0x20)}},
		{"[default on default]x[/]", Style{}},
	}
	for _, tt := range tests {
		spans := markupSpans(t, tt.src)
		if len(spans) != 1 || spans[0].Text != "x" || spans[0].Style != tt.want {
			t.Errorf("ParseMarkup(%q) = %+v, want style %+v", tt.src, spans, tt.want)
		}
	}
}

func TestMarkupParseNestingAndNot(t *testing.T) {
	spans := markupSpans(t, "a[bold red]b[italic]c[not bold on blue]d[/]e[/]f[/bold red]g")
	want := []RichSpan{
		{Text: "a"},
		{Text: "b", Style: Style{Bold: true, FG: ColorIndex(1)}},
		{Text: "c", Style: Style{Bold: true, Italic: true, FG: ColorIndex(1)}},
		{Text: "d", Style: Style{Italic: true, FG: ColorIndex(1), BG: ColorIndex(4)}},
		{Text: "e", Style: Style{Bold: true, Italic: true, FG: ColorIndex(1)}},
		{Text: "f", Style: Style{Bold: true, FG: ColorIndex(1)}},
		{Text: "g"},
	}
	if !reflect.DeepEqual(spans, want) {
		t.Fatalf("spans = %+v\nwant  %+v", spans, want)
	}
}

func TestMarkupParseEscapes(t *testing.T) {
	tests := map[string]string{
		`\[bold\]`:    "[bold]",
		`[[bold]`:     "[bold]",
		`a\\b`:        `a\b`,
		`C:\dir`:      `C:\dir`,
		`a]b`:         "a]b",
		`tail\`:       `tail\`,
		"[red]\\[[/]": "[",
	}
	for src, want := range tests {
		doc, err := ParseMarkup(src)
		if err != nil {
			t.Errorf("ParseMarkup(%q): %v", src, err)
			continue
		}
		if got := doc.ToPlainText(); got != want {
			t.Errorf("ParseMarkup(%q) text = %q, want %q", src, got, want)
		}
	}
}

func TestMarkupParseLines(t *testing.T) {
	doc, err := ParseMarkup("[bold]a[/]\r\n\rb\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Lines) != 3 || doc.ToPlainText() != "a\n\nb" {
		t.Fatalf("lines = %+v", doc.Lines)
	}
	empty, err := ParseMarkup("")
	if err != nil || len(empty.Lines) != 1 || len(empty.Lines[0].Spans) != 0 {
		t.Fatalf("empty = %+v, %v", empty, err)
	}
}

func TestMarkupParseErrors(t *testing.T) {
	tests := []struct {
		src       string
		line, col int
		msg       string
	}{
		{"ab[bold", 1, 3, "unterminated tag"},
		{"x\n[]y", 2, 2, "empty tag"},
		{"[blink]x[/]", 1, 2, `unknown attribute "blink"`},
		{"[on]x[/]", 1, 2, `"on" needs a color`},
		{"[on nope]x[/]", 1, 2, `unknown color "nope"`},
		{"[not red]x[/]", 1, 2, `unknown attribute "red" after "not"`},
		{"[color(256)]x[/]", 1, 2, "must be 0-255"},
		{"[#12345g]x[/]", 1, 2, "invalid hex color"},
		{"äö[/]", 1, 3, "closing tag without open tag"},
		{"[bold]x[/red]", 1, 8, "does not match [bold]"},
		{"ok\n[bold]open", 2, 11, "unclosed tag [bold]"},
	}
	for _, tt := range tests {
		_, err := ParseMarkup(tt.src)
		var me *MarkupError
		if !errors.As(err, &me) {
			t.Errorf("ParseMarkup(%q) err = %v, want *MarkupError", tt.src, err)
			continue
		}
		if me.Line != tt.line || me.Col != tt.col || !strings.Contains(me.Msg, tt.msg) {
			t.Errorf("ParseMarkup(%q) = %v, want %d:%d %q", tt.src, err, tt.line, tt.col, tt.msg)
		}
	}
}

func TestMarkupSerialize(t *testing.T) {
	doc := RichDocument{Lines: []RichLine{
		{Spans: []RichSpan{
			{Text: "plain [x] \\ "},
			{Text: "hot", Style: Style{Bold: true, Italic: true, FG: ColorIndex(9), BG: ColorRGB(1, 2, 3)}},
			{Text: "idx", Style: Style{FG: ColorIndex(200), Strike: true}},
		}},
		{},
		{Spans: []RichSpan{{Text: "all", Style: Style{Dim: true, Underline: true, Invert: true, BG: ColorIndex(8)}}}},
	}}
	want := `plain \[x\] \\ [bold italic bright_red on #010203]hot[/][strike color(200)]idx[/]` +
		"\n\n[dim underline invert on bright_black]all[/]\n"
	if got := doc.ToMarkup(); got != want {
		t.Fatalf("ToMarkup =\n%s\nwant\n%s", got, want)
	}
}

func TestMarkupRoundTrip(t *testing.T) {
	canonical := []string{
		"\n",
		"\n\n",
		"plain\n",
		"[bold red]a[/] b [on #102030]\\[c\\][/]\n",
		"[italic]x[/][italic underline]y[/]\n\n[color(99)]z[/] \\\\\n",
	}
	for _, src := range canonical {
		doc, err := ParseMarkup(src)
		if err != nil {
			t.Fatalf("ParseMarkup(%q): %v", src, err)
		}
		if got := doc.ToMarkup(); got != src {
			t.Errorf("markup round trip %q -> %q", src, got)
		}
	}
	nested := "[bold][red]a[on blue]b[/][/]c[/]"
	doc, err := ParseMarkup(nested)
	if err != nil {
		t.Fatal(err)
	}
	again, err := ParseMarkup(doc.ToMarkup())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(doc, again) {
		t.Fatalf("doc round trip:\n%+v\n%+v", doc, again)
	}
}

func TestMarkupMatchesANSIRoundTrip(t *testing.T) {
	var src RichDocument
	src.FromANSI("\x1b[1;31mred\x1b[0m mid \x1b[3;48;2;1;2;3mbg[\x1b[0m\n\x1b[38;5;200mx\x1b[0m")
	doc, err := ParseMarkup(src.ToMarkup())
	if err != nil {
		t.Fatalf("ParseMarkup(%q): %v", src.ToMarkup(), err)
	}
	if got, want := doc.ToANSI(), src.ToANSI(); got != want {
		t.Fatalf("ToANSI via markup = %q, want %q", got, want)
	}
	direct, err := ParseMarkup("[bold red]hi[/]")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := direct.ToANSI(), (Style{Bold: true, FG: ColorIndex(1)}).ANSI()+"hi\x1b[0m"; got != want {
		t.Fatalf("ToANSI = %q, want %q", got, want)
	}
}

func FuzzMarkupParse(f *testing.F) {
	for _, seed := range []string{"", "[bold]x[/]", `\[[[`, "[on #fff]a\nb", "[/]", "[not]"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, src string) {
		doc, err := ParseMarkup(src)
		if err != nil {
			return
		}
		again, err := ParseMarkup(doc.ToMarkup())
		if err != nil {
			t.Fatalf("reparse %q: %v", doc.ToMarkup(), err)
		}
		if again.ToMarkup() != doc.ToMarkup() {
			t.Fatalf("not stable: %q vs %q", doc.ToMarkup(), again.ToMarkup())
		}
	})
}
