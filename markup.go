// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// .tui bracket markup (issue 308):
//
//	[bold red on #202020]text[/]   open a style frame, close the innermost
//	[/bold red on #202020]         close the innermost frame, which must match
//	\[ \] \\ [[                    literal '[', ']', '\', '['
//
// Attributes: bold|b, dim, italic|i, underline|u, strike|s, invert|reverse,
// "not <attr>" to clear an inherited flag, a color for the foreground and
// "on <color>" for the background. Colors: default, a name from
// spec/markup.yaml, color(0-255), #rgb or #rrggbb. Tags must close on the line
// where they start, so every line parses on its own. Box blocks are in
// markup_box.go.

//go:embed spec/markup.yaml
var markupSpecYAML []byte

var markupColorIndex, markupColorName = func() (map[string]uint8, map[uint8]string) {
	var spec struct {
		Colors yaml.Node `yaml:"colors"`
	}
	if err := yaml.Unmarshal(markupSpecYAML, &spec); err != nil {
		panic(fmt.Sprintf("loom: parse spec/markup.yaml: %v", err))
	}
	byName, byIndex := map[string]uint8{}, map[uint8]string{}
	pairs := spec.Colors.Content
	for i := 0; i+1 < len(pairs); i += 2 {
		var index uint8
		if err := pairs[i+1].Decode(&index); err != nil {
			panic(fmt.Sprintf("loom: parse spec/markup.yaml: color %q: %v", pairs[i].Value, err))
		}
		byName[pairs[i].Value] = index
		if _, ok := byIndex[index]; !ok {
			byIndex[index] = pairs[i].Value
		}
	}
	return byName, byIndex
}()

// MarkupError reports a .tui markup syntax error. Line and Col are 1-based;
// Col counts runes.
type MarkupError struct {
	Line, Col int
	Msg       string
}

func (e *MarkupError) Error() string {
	return fmt.Sprintf("markup %d:%d: %s", e.Line, e.Col, e.Msg)
}

type markupFrame struct {
	tag   string
	style Style
}

// ParseMarkup parses .tui bracket markup into a RichDocument. Parsing is
// strict: unknown attributes, malformed or mismatched tags and tags left open
// at the end of a line return a *MarkupError. CRLF and CR become LF, and one
// trailing newline is dropped, as in FromANSI.
func ParseMarkup(src string) (*RichDocument, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src = strings.ReplaceAll(src, "\r", "\n")
	src = strings.TrimSuffix(src, "\n")
	doc := &RichDocument{}
	var boxes []*markupBox
	target := func() *RichDocument {
		if len(boxes) == 0 {
			return doc
		}
		return boxes[len(boxes)-1].doc
	}
	for lineIndex, line := range strings.Split(src, "\n") {
		lineNo := lineIndex + 1
		switch markupBoxTag(line) {
		case "open":
			box, err := parseMarkupBoxTag(line, lineNo)
			if err != nil {
				return nil, err
			}
			boxes = append(boxes, box)
			continue
		case "close":
			if len(boxes) == 0 {
				return nil, &MarkupError{lineNo, 1, "[/box] without open [box]"}
			}
			box := boxes[len(boxes)-1]
			boxes = boxes[:len(boxes)-1]
			parent := target()
			parent.Lines = append(parent.Lines, box.render()...)
			continue
		}
		into := target()
		into.Lines = append(into.Lines, RichLine{})
		if err := parseMarkupLine(into, line, lineNo); err != nil {
			return nil, err
		}
	}
	if len(boxes) > 0 {
		return nil, &MarkupError{boxes[len(boxes)-1].lineNo, 1, "unclosed [box] block"}
	}
	doc.normalize()
	return doc, nil
}

func parseMarkupLine(doc *RichDocument, line string, lineNo int) error {
	var stack []markupFrame
	var text strings.Builder
	style := func() Style {
		if len(stack) == 0 {
			return Style{}
		}
		return stack[len(stack)-1].style
	}
	flush := func() {
		if text.Len() > 0 {
			appendRichText(doc, text.String(), style())
			text.Reset()
		}
	}
	col := func(offset int) int { return utf8.RuneCountInString(line[:offset]) + 1 }
	for i := 0; i < len(line); {
		c := line[i]
		switch {
		case c == '\\' && i+1 < len(line) && strings.IndexByte(`[]\`, line[i+1]) >= 0:
			text.WriteByte(line[i+1])
			i += 2
		case c == '[' && i+1 < len(line) && line[i+1] == '[':
			text.WriteByte('[')
			i += 2
		case c == '[':
			end := strings.IndexByte(line[i:], ']')
			if end < 0 {
				return &MarkupError{lineNo, col(i), "unterminated tag"}
			}
			body := line[i+1 : i+end]
			flush()
			if closing, ok := strings.CutPrefix(body, "/"); ok {
				if len(stack) == 0 {
					return &MarkupError{lineNo, col(i), "closing tag without open tag"}
				}
				if closing != "" {
					tag, _, err := applyMarkupTag(Style{}, closing)
					if err != nil {
						return &MarkupError{lineNo, col(i + 2), err.Error()}
					}
					if open := stack[len(stack)-1].tag; tag != open {
						return &MarkupError{lineNo, col(i), fmt.Sprintf("closing tag [/%s] does not match [%s]", closing, open)}
					}
				}
				stack = stack[:len(stack)-1]
			} else {
				tag, next, err := applyMarkupTag(style(), body)
				if err != nil {
					return &MarkupError{lineNo, col(i + 1), err.Error()}
				}
				stack = append(stack, markupFrame{tag: tag, style: next})
			}
			i += end + 1
		default:
			_, size := utf8.DecodeRuneInString(line[i:])
			text.WriteString(line[i : i+size])
			i += size
		}
	}
	flush()
	if len(stack) > 0 {
		return &MarkupError{lineNo, col(len(line)), fmt.Sprintf("unclosed tag [%s] at end of line", stack[len(stack)-1].tag)}
	}
	return nil
}

// applyMarkupTag applies the attributes of one tag body to base. It returns
// the normalized tag text used to match closing tags.
func applyMarkupTag(base Style, body string) (string, Style, error) {
	words := strings.Fields(strings.ToLower(body))
	if len(words) == 0 {
		return "", base, fmt.Errorf("empty tag")
	}
	style := base
	for i := 0; i < len(words); i++ {
		word := words[i]
		switch word {
		case "not":
			if i+1 >= len(words) {
				return "", base, fmt.Errorf(`"not" needs an attribute`)
			}
			i++
			flag := markupFlag(&style, words[i])
			if flag == nil {
				return "", base, fmt.Errorf("unknown attribute %q after \"not\"", words[i])
			}
			*flag = false
		case "on":
			if i+1 >= len(words) {
				return "", base, fmt.Errorf(`"on" needs a color`)
			}
			i++
			color, err := parseMarkupColor(words[i])
			if err != nil {
				return "", base, err
			}
			style.BG = color
		default:
			if flag := markupFlag(&style, word); flag != nil {
				*flag = true
				continue
			}
			color, err := parseMarkupColor(word)
			if err != nil && !strings.HasPrefix(word, "#") && !strings.HasPrefix(word, "color(") {
				return "", base, fmt.Errorf("unknown attribute %q", word)
			}
			if err != nil {
				return "", base, err
			}
			style.FG = color
		}
	}
	return strings.Join(words, " "), style, nil
}

func markupFlag(style *Style, word string) *bool {
	switch word {
	case "bold", "b":
		return &style.Bold
	case "dim":
		return &style.Dim
	case "italic", "i":
		return &style.Italic
	case "underline", "u":
		return &style.Underline
	case "strike", "s":
		return &style.Strike
	case "invert", "reverse":
		return &style.Invert
	}
	return nil
}

func parseMarkupColor(word string) (Color, error) {
	if word == "default" {
		return ColorReset(), nil
	}
	if index, ok := markupColorIndex[word]; ok {
		return ColorIndex(index), nil
	}
	if arg, ok := strings.CutPrefix(word, "color("); ok {
		if arg, ok = strings.CutSuffix(arg, ")"); ok {
			n, err := strconv.ParseUint(arg, 10, 8)
			if err != nil {
				return Color{}, fmt.Errorf("color index %q must be 0-255", arg)
			}
			return ColorIndex(uint8(n)), nil
		}
	}
	if hex, ok := strings.CutPrefix(word, "#"); ok {
		if len(hex) == 3 {
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		n, err := strconv.ParseUint(hex, 16, 32)
		if len(hex) != 6 || err != nil {
			return Color{}, fmt.Errorf("invalid hex color %q", word)
		}
		return ColorRGB(uint8(n>>16), uint8(n>>8), uint8(n)), nil
	}
	return Color{}, fmt.Errorf("unknown color %q", word)
}

// ToMarkup serializes d as .tui bracket markup. Each styled span becomes one
// [attrs]text[/] group; default-styled text stays bare. The result ends with a
// newline, which ParseMarkup drops again. Pill, link and code metadata are not
// part of the v1 format; their visible text is kept.
func (d RichDocument) ToMarkup() string {
	var out strings.Builder
	for lineIndex, line := range d.Lines {
		for _, span := range line.Spans {
			if span.Text == "" {
				continue
			}
			tag := markupTag(span.Style)
			if tag != "" {
				out.WriteString("[" + tag + "]")
			}
			out.WriteString(escapeMarkup(span.Text))
			if tag != "" {
				out.WriteString("[/]")
			}
		}
		if lineIndex+1 < len(d.Lines) {
			out.WriteByte('\n')
		}
	}
	out.WriteByte('\n')
	return out.String()
}

func markupTag(style Style) string {
	var words []string
	for _, attr := range []struct {
		on   bool
		name string
	}{
		{style.Bold, "bold"}, {style.Dim, "dim"}, {style.Italic, "italic"},
		{style.Underline, "underline"}, {style.Strike, "strike"}, {style.Invert, "invert"},
	} {
		if attr.on {
			words = append(words, attr.name)
		}
	}
	if style.FG != ColorReset() {
		words = append(words, markupColor(style.FG))
	}
	if style.BG != ColorReset() {
		words = append(words, "on", markupColor(style.BG))
	}
	return strings.Join(words, " ")
}

func markupColor(c Color) string {
	if c.mode == colorRGB {
		return fmt.Sprintf("#%02x%02x%02x", c.r, c.g, c.b)
	}
	if name, ok := markupColorName[c.index]; ok {
		return name
	}
	return fmt.Sprintf("color(%d)", c.index)
}

var markupEscaper = strings.NewReplacer(`\`, `\\`, `[`, `\[`, `]`, `\]`)

func escapeMarkup(text string) string { return markupEscaper.Replace(text) }
