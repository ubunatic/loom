// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"unicode/utf8"
)

// RichDocument is a flow-based document. Lines preserve empty lines, including
// a final empty line after a trailing newline.
type RichDocument struct {
	Lines []RichLine
}

// RichLine is an ordered sequence of styled spans on one logical line.
type RichLine struct {
	Spans []RichSpan
}

// RichSpan is text with a visual style and optional semantic metadata.
// PillData marks an atomic token; its Text is the label shown to the user.
type RichSpan struct {
	Text     string
	Style    Style
	PillData *RichPill
	Link     string
	Code     bool
}

// RichPill stores application-defined data for an atomic inline token.
type RichPill struct {
	Kind     string
	ID       string
	Metadata map[string]string
}

// FromANSI replaces d with the styled text represented by an ANSI SGR stream.
// Non-SGR escape sequences are consumed and never interpreted as terminal
// commands. CRLF and bare CR line endings are normalized to LF.
func (d *RichDocument) FromANSI(input string) {
	if d == nil {
		return
	}
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")
	d.Lines = []RichLine{{}}
	style := Style{}
	for i := 0; i < len(input); {
		if input[i] == '\x1b' {
			if end, params, ok := richTextCSI(input, i); ok {
				if input[end-1] == 'm' {
					style = applySGRSequence(style, params)
				}
				i = end
				continue
			}
			if i+1 < len(input) && input[i+1] == ']' {
				i = skipRichTextOSC(input, i+2)
			} else {
				i++
			}
			continue
		}
		r := input[i]
		if r == '\n' {
			d.Lines = append(d.Lines, RichLine{})
			i++
			continue
		}
		if r < 0x80 {
			if r >= 0x20 && r != 0x7f {
				appendRichText(d, string(r), style)
			}
			i++
			continue
		}
		_, size := utf8.DecodeRuneInString(input[i:])
		if size < 1 {
			size = 1
		}
		appendRichText(d, input[i:i+size], style)
		i += size
	}
}

// ToANSI serializes visible text and SGR styling, ending with a full reset.
// Pill labels are serialized as text; semantic metadata remains in the document.
func (d RichDocument) ToANSI() string {
	var out strings.Builder
	for lineIndex, line := range d.Lines {
		for _, span := range line.Spans {
			out.WriteString(span.Style.ANSI())
			out.WriteString(span.Text)
		}
		if lineIndex+1 < len(d.Lines) {
			out.WriteByte('\n')
		}
	}
	out.WriteString("\x1b[0m")
	return out.String()
}

// ToPlainText returns visible text, with logical lines joined by LF.
func (d RichDocument) ToPlainText() string {
	lines := make([]string, len(d.Lines))
	for i, line := range d.Lines {
		var text strings.Builder
		for _, span := range line.Spans {
			text.WriteString(span.Text)
		}
		lines[i] = text.String()
	}
	return strings.Join(lines, "\n")
}

func appendRichText(d *RichDocument, text string, style Style) {
	line := &d.Lines[len(d.Lines)-1]
	if n := len(line.Spans); n > 0 && line.Spans[n-1].Style == style && line.Spans[n-1].PillData == nil && line.Spans[n-1].Link == "" && !line.Spans[n-1].Code {
		line.Spans[n-1].Text += text
		return
	}
	line.Spans = append(line.Spans, RichSpan{Text: text, Style: style})
}

// richTextCSI returns the end offset and parameters of a complete CSI.
func richTextCSI(input string, start int) (int, string, bool) {
	if start+1 >= len(input) || input[start+1] != '[' {
		return 0, "", false
	}
	for i := start + 2; i < len(input); i++ {
		if input[i] >= '@' && input[i] <= '~' {
			return i + 1, input[start+2 : i], true
		}
	}
	return len(input), "", true
}

func skipRichTextOSC(input string, start int) int {
	for i := start; i < len(input); i++ {
		if input[i] == '\a' {
			return i + 1
		}
		if input[i] == '\x1b' && i+1 < len(input) && input[i+1] == '\\' {
			return i + 2
		}
	}
	return len(input)
}
