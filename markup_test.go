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
	for _, seed := range []string{"", "[bold]x[/]", `\[[[`, "[on #fff]a\nb", "[/]", "[not]", "[table]\n| a | b |\n|:-|-:|\n| [red]x[/] |\n[/table]"} {
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

func markupBoxLines(t *testing.T, src string) []string {
	t.Helper()
	doc, err := ParseMarkup(src)
	if err != nil {
		t.Fatalf("ParseMarkup(%q): %v", src, err)
	}
	var lines []string
	for _, line := range doc.Lines {
		var text strings.Builder
		for _, span := range line.Spans {
			text.WriteString(span.Text)
		}
		lines = append(lines, text.String())
	}
	return lines
}

func TestMarkupBoxStyles(t *testing.T) {
	tests := []struct {
		style string
		want  []string
	}{
		{"", []string{"┌────┐", "│ ab │", "│ c  │", "└────┘"}},
		{" style=normal", []string{"┌────┐", "│ ab │", "│ c  │", "└────┘"}},
		{" style=sharp", []string{"┌────┐", "│ ab │", "│ c  │", "└────┘"}},
		{" style=rounded", []string{"╭────╮", "│ ab │", "│ c  │", "╰────╯"}},
		{" style=double", []string{"╔════╗", "║ ab ║", "║ c  ║", "╚════╝"}},
		{" style=ascii", []string{"+----+", "| ab |", "| c  |", "+----+"}},
	}
	for _, tt := range tests {
		got := markupBoxLines(t, "[box"+tt.style+"]\nab\nc\n[/box]")
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("style %q = %q, want %q", tt.style, got, tt.want)
		}
	}
}

func TestMarkupBoxLayout(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
	}{
		{"surrounding text", "top\n[box]\nx\n[/box]\nend", []string{"top", "┌───┐", "│ x │", "└───┘", "end"}},
		{"pad", "[box pad=0]\nab\n[/box]", []string{"┌──┐", "│ab│", "└──┘"}},
		{"pad 2", "[box pad=2]\nab\n\n[/box]", []string{"┌──────┐", "│  ab  │", "│      │", "└──────┘"}},
		{"title", "[box title=\"Hi\"]\nabcdef\n[/box]", []string{"┌─ Hi ───┐", "│ abcdef │", "└────────┘"}},
		{"wide title", "[box style=rounded title=\"Status\"]\nx\n[/box]", []string{"╭─ Status ─╮", "│ x        │", "╰──────────╯"}},
		{"wide chars", "[box]\n日本\nab\n[/box]", []string{"┌──────┐", "│ 日本 │", "│ ab   │", "└──────┘"}},
		{"empty", "[box]\n[/box]", []string{"┌──┐", "└──┘"}},
		{"indented tags", "  [box]\nx\n  [/box]", []string{"┌───┐", "│ x │", "└───┘"}},
		{"nested", "[box]\n[box style=rounded]\nx\n[/box]\ny\n[/box]", []string{
			"┌───────┐",
			"│ ╭───╮ │",
			"│ │ x │ │",
			"│ ╰───╯ │",
			"│ y     │",
			"└───────┘",
		}},
	}
	for _, tt := range tests {
		if got := markupBoxLines(t, tt.src); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestMarkupBoxStyling(t *testing.T) {
	doc, err := ParseMarkup("[box border=\"bold cyan\" title=\"[red]T[/]\"]\n[italic]a[/]\n[/box]")
	if err != nil {
		t.Fatal(err)
	}
	border := Style{Bold: true, FG: ColorIndex(6)}
	want := []RichLine{
		{Spans: []RichSpan{{Text: "┌─ ", Style: border}, {Text: "T", Style: Style{FG: ColorIndex(1)}}, {Text: " ─┐", Style: border}}},
		{Spans: []RichSpan{{Text: "│", Style: border}, {Text: " "}, {Text: "a", Style: Style{Italic: true}}, {Text: "   "}, {Text: "│", Style: border}}},
		{Spans: []RichSpan{{Text: "└─────┘", Style: border}}},
	}
	if !reflect.DeepEqual(doc.Lines, want) {
		t.Fatalf("lines = %+v, want %+v", doc.Lines, want)
	}
	again, err := ParseMarkup(doc.ToMarkup())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(again.Lines, doc.Lines) {
		t.Fatalf("round trip = %+v, want %+v", again.Lines, doc.Lines)
	}
}

func TestMarkupBoxErrors(t *testing.T) {
	tests := []struct {
		src       string
		line, col int
		msg       string
	}{
		{"[box]\nx", 1, 1, "unclosed [box] block"},
		{"x\n[box]\n[box]\n[/box]", 2, 1, "unclosed [box] block"},
		{"x\n[/box]", 2, 1, "[/box] without open [box]"},
		{"[box style=fancy]\n[/box]", 1, 6, `unknown box style "fancy"`},
		{"[box color=red]\n[/box]", 1, 6, `unknown box option "color"`},
		{"[box rounded]\n[/box]", 1, 6, "needs key=value"},
		{"[box pad=-1]\n[/box]", 1, 6, "pad must be"},
		{"[box border=blink]\n[/box]", 1, 6, `unknown attribute "blink"`},
		{"[box title=\"[bold]x\"]\n[/box]", 1, 6, "title: unclosed tag"},
		{"[box title=\"x]\n[/box]", 1, 12, "unterminated quoted value"},
		{"[box pad=1\n[/box]", 1, 11, "unterminated box tag"},
		{"[box] text\n[/box]", 1, 6, "must stand alone"},
		{"a [box]x[/box]", 1, 4, `unknown attribute "box"`},
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

func TestMarkupTableStyles(t *testing.T) {
	src := "| a | bb |\n|---|---|\n| ccc | d |\n[/table]"
	tests := []struct {
		style string
		want  []string
	}{
		{"", []string{"┌─────┬────┐", "│ a   │ bb │", "├─────┼────┤", "│ ccc │ d  │", "└─────┴────┘"}},
		{" style=normal", []string{"┌─────┬────┐", "│ a   │ bb │", "├─────┼────┤", "│ ccc │ d  │", "└─────┴────┘"}},
		{" style=rounded", []string{"╭─────┬────╮", "│ a   │ bb │", "├─────┼────┤", "│ ccc │ d  │", "╰─────┴────╯"}},
		{" style=double", []string{"╔═════╦════╗", "║ a   ║ bb ║", "╠═════╬════╣", "║ ccc ║ d  ║", "╚═════╩════╝"}},
		{" style=ascii", []string{"+-----+----+", "| a   | bb |", "+-----+----+", "| ccc | d  |", "+-----+----+"}},
	}
	for _, tt := range tests {
		got := markupBoxLines(t, "[table"+tt.style+"]\n"+src)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("style %q = %q, want %q", tt.style, got, tt.want)
		}
	}
}

func TestMarkupTableLayout(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
	}{
		{"align", "[table]\n| L | C | R |\n|:--|:-:|--:|\n| abcd | abcd | abcd |\n| x | x | x |\n[/table]", []string{
			"┌──────┬──────┬──────┐",
			"│ L    │  C   │    R │",
			"├──────┼──────┼──────┤",
			"│ abcd │ abcd │ abcd │",
			"│ x    │  x   │    x │",
			"└──────┴──────┴──────┘",
		}},
		{"no header", "[table]\n| a | b |\n| cc | d |\n[/table]", []string{"┌────┬───┐", "│ a  │ b │", "│ cc │ d │", "└────┴───┘"}},
		{"pad 0", "[table pad=0]\n|a|b|\n[/table]", []string{"┌─┬─┐", "│a│b│", "└─┴─┘"}},
		{"wide chars and markup", "[table]\n| [bold red]日本[/] | x |\n| ab | [i]yz[/] |\n[/table]", []string{"┌──────┬────┐", "│ 日本 │ x  │", "│ ab   │ yz │", "└──────┴────┘"}},
		{"short rows and no trailing pipe", "[table]\n| a | b | c\n| d\n[/table]", []string{"┌───┬───┬───┐", "│ a │ b │ c │", "│ d │   │   │", "└───┴───┴───┘"}},
		{"escapes and blank lines", "[table]\n| a\\|b | \\[x\\] |\n\n| \\\\ | y |\n[/table]", []string{"┌─────┬─────┐", "│ a|b │ [x] │", "│ \\   │ y   │", "└─────┴─────┘"}},
		{"header only", "[table]\n| h |\n|---|\n[/table]", []string{"┌───┐", "│ h │", "├───┤", "└───┘"}},
		{"empty", "[table]\n[/table]", []string{"┌┐", "└┘"}},
		{"indented", "  [table]\n  | a |\n  [/table]", []string{"┌───┐", "│ a │", "└───┘"}},
		{"in box", "[box]\n[table]\n| a |\n[/table]\nz\n[/box]", []string{"┌───────┐", "│ ┌───┐ │", "│ │ a │ │", "│ └───┘ │", "│ z     │", "└───────┘"}},
	}
	for _, tt := range tests {
		if got := markupBoxLines(t, tt.src); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestMarkupTableStyling(t *testing.T) {
	doc, err := ParseMarkup("[table border=cyan pad=0]\n|[red]H[/]|\n|-|\n|[i]v[/]|\n[/table]")
	if err != nil {
		t.Fatal(err)
	}
	border := Style{FG: ColorIndex(6)}
	want := []RichLine{
		{Spans: []RichSpan{{Text: "┌─┐", Style: border}}},
		{Spans: []RichSpan{{Text: "│", Style: border}, {Text: "H", Style: Style{Bold: true, FG: ColorIndex(1)}}, {Text: "│", Style: border}}},
		{Spans: []RichSpan{{Text: "├─┤", Style: border}}},
		{Spans: []RichSpan{{Text: "│", Style: border}, {Text: "v", Style: Style{Italic: true}}, {Text: "│", Style: border}}},
		{Spans: []RichSpan{{Text: "└─┘", Style: border}}},
	}
	if !reflect.DeepEqual(doc.Lines, want) {
		t.Fatalf("lines = %+v, want %+v", doc.Lines, want)
	}
	again, err := ParseMarkup(doc.ToMarkup())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(again.Lines, doc.Lines) {
		t.Fatalf("round trip = %+v, want %+v", again.Lines, doc.Lines)
	}
	for _, header := range []string{"header=none", `header=""`} {
		doc, err := ParseMarkup("[table pad=0 " + header + "]\n|H|\n|-|\n[/table]")
		if err != nil {
			t.Fatal(err)
		}
		if got, want := doc.Lines[1].Spans, []RichSpan{{Text: "│H│"}}; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: header spans = %+v, want %+v", header, got, want)
		}
	}
	doc, err = ParseMarkup("[table pad=0 header=\"underline yellow\"]\n|H|\n|-|\n[/table]")
	if err != nil {
		t.Fatal(err)
	}
	if got := doc.Lines[1].Spans[1].Style; got != (Style{Underline: true, FG: ColorIndex(3)}) {
		t.Errorf("header style = %+v", got)
	}
}

func TestMarkupTableErrors(t *testing.T) {
	tests := []struct {
		src       string
		line, col int
		msg       string
	}{
		{"[table]\n| a |", 1, 1, "unclosed [table] block"},
		{"[/table]", 1, 1, "[/table] without open [table]"},
		{"[box]\n[/table]", 2, 1, "[/table] without open [table]"},
		{"[table]\n[/box]", 2, 1, "[/box] inside [table]"},
		{"[table]\n[box]\n[/box]\n[/table]", 2, 1, "[box] inside [table] is not supported"},
		{"[table]\n[table]\n[/table]", 2, 1, "[table] inside [table] is not supported"},
		{"[table]\n  a | b\n[/table]", 2, 3, "must start with |"},
		{"[table]\n|---|\n[/table]", 2, 1, "separator row must follow the first row"},
		{"[table]\n| a |\n| b |\n|---|\n[/table]", 4, 1, "separator row must follow the first row"},
		{"[table]\n| a |\n|---|\n|---|\n[/table]", 4, 1, "separator row must follow the first row"},
		{"[table]\n| a | [bold]b |\n[/table]", 2, 14, "unclosed tag [bold]"},
		{"[table]\n| [blink]a[/] |\n[/table]", 2, 4, `unknown attribute "blink"`},
		{"[table style=fancy]\n[/table]", 1, 8, `unknown table style "fancy"`},
		{"[table align=left]\n[/table]", 1, 8, `unknown table option "align"`},
		{"[table pad=x]\n[/table]", 1, 8, "pad must be"},
		{"[table header=blink]\n[/table]", 1, 8, `header: unknown attribute "blink"`},
		{"[table] x\n[/table]", 1, 8, "table tag must stand alone"},
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
