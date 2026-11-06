// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"unicode"
)

func (e *TextEntry) toggleBold() EventResult {
	e.typingRun = false
	if !e.hasSelection() {
		before := e.snapshot()
		e.typingBold = !e.typingBold
		e.pushUndo(before)
		e.redo = nil
		return Consumed()
	}
	lo, hi := e.selectionBounds()
	allBold := true
	anyText := false
	for i := lo; i < hi; i++ {
		if e.units[i] == "\n" {
			continue
		}
		anyText = true
		allBold = allBold && e.bold[i]
	}
	if anyText {
		e.change(false, func() {
			for i := lo; i < hi; i++ {
				if e.units[i] != "\n" {
					e.bold[i] = !allBold
				}
			}
		})
	}
	return Consumed()
}

// MarkdownValue returns the plain entry text with bold spans marked using
// Markdown strong delimiters. Literal Markdown punctuation is backslash
// escaped. Existing newlines are preserved, and each hard line closes bold
// spans. Leading and trailing whitespace in a bold span remains outside the
// delimiters because Markdown does not render whitespace-only emphasis.
func (e *TextEntry) MarkdownValue() string {
	var out strings.Builder
	for start := 0; start < len(e.units); {
		end := start
		for end < len(e.units) && e.units[end] != "\n" {
			end++
		}
		writeMarkdownLine(&out, e.units[start:end], e.bold[start:end])
		if end < len(e.units) {
			out.WriteString(e.units[end])
		}
		start = end + 1
	}
	return out.String()
}

func writeMarkdownLine(out *strings.Builder, units []string, bold []bool) {
	for i := 0; i < len(units); {
		if !bold[i] {
			writeMarkdownEscaped(out, units[i])
			i++
			continue
		}
		end := i + 1
		for end < len(units) && bold[end] {
			end++
		}
		coreStart, coreEnd := i, end
		for coreStart < coreEnd && markdownWhitespace(units[coreStart]) {
			coreStart++
		}
		for coreEnd > coreStart && markdownWhitespace(units[coreEnd-1]) {
			coreEnd--
		}
		for j := i; j < coreStart; j++ {
			writeMarkdownEscaped(out, units[j])
		}
		if coreStart < coreEnd {
			out.WriteString("**")
			for j := coreStart; j < coreEnd; j++ {
				writeMarkdownEscaped(out, units[j])
			}
			out.WriteString("**")
		}
		for j := coreEnd; j < end; j++ {
			writeMarkdownEscaped(out, units[j])
		}
		i = end
	}
}

func markdownWhitespace(unit string) bool {
	return strings.TrimFunc(unit, unicode.IsSpace) == ""
}

func writeMarkdownEscaped(out *strings.Builder, text string) {
	for _, r := range text {
		if r >= '!' && r <= '~' && strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", r) {
			out.WriteByte('\\')
		}
		out.WriteRune(r)
	}
}
