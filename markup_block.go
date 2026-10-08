// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// .tui block tags ([box], [table], [image]) stand alone on their line and carry
// key=value options. The option syntax is shared by all block kinds.

// markupBlockNames lists the block tags that ParseMarkup recognizes.
var markupBlockNames = []string{"box", "table", "image"}

// markupOption is one key=value option of a block tag. Col is the 1-based
// column of the key.
type markupOption struct {
	key, value string
	col        int
}

// markupBlockTag reports whether line is a block tag. It returns the block
// name and "open" or "close", or two empty strings.
func markupBlockTag(line string) (name, kind string) {
	trimmed := strings.TrimSpace(line)
	for _, name := range markupBlockNames {
		if trimmed == "[/"+name+"]" {
			return name, "close"
		}
		if rest, ok := strings.CutPrefix(trimmed, "["+name); ok && rest != "" && (rest[0] == ' ' || rest[0] == ']') {
			return name, "open"
		}
	}
	return "", ""
}

// parseMarkupBlockOptions parses the key=value options of an opening block
// tag line such as `[box style=rounded title="Status"]`.
func parseMarkupBlockOptions(line, name string, lineNo int) ([]markupOption, error) {
	col := func(offset int) int { return utf8.RuneCountInString(line[:offset]) + 1 }
	fail := func(offset int, format string, args ...any) error {
		return &MarkupError{lineNo, col(offset), fmt.Sprintf(format, args...)}
	}
	var options []markupOption
	i := strings.Index(line, "["+name) + len(name) + 1
	for {
		for i < len(line) && line[i] == ' ' {
			i++
		}
		if i >= len(line) {
			return nil, fail(len(line), "unterminated %s tag", name)
		}
		if line[i] == ']' {
			if strings.TrimSpace(line[i+1:]) != "" {
				return nil, fail(i+1, "%s tag must stand alone on its line", name)
			}
			return options, nil
		}
		keyStart := i
		eq := strings.IndexByte(line[i:], '=')
		if eq < 0 {
			return nil, fail(keyStart, "%s option needs key=value", name)
		}
		key := strings.ToLower(line[i : i+eq])
		i += eq + 1
		var value string
		if i < len(line) && line[i] == '"' {
			end := strings.IndexByte(line[i+1:], '"')
			if end < 0 {
				return nil, fail(i, "unterminated quoted value")
			}
			value = line[i+1 : i+1+end]
			i += end + 2
		} else {
			end := strings.IndexAny(line[i:], " ]")
			if end < 0 {
				end = len(line) - i
			}
			value = line[i : i+end]
			i += end
		}
		options = append(options, markupOption{key: key, value: value, col: col(keyStart)})
	}
}

// parseMarkupBorderStyle parses a block style option; "normal" is an alias
// for "sharp".
func parseMarkupBorderStyle(value string) (BoxBorderStyle, error) {
	name := strings.ToLower(value)
	if name == "normal" {
		name = "sharp"
	}
	return ParseBoxBorderStyle(name)
}
