// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package boxcheck provides optional checks for ANSI artwork.
package boxcheck

import (
	"fmt"
	"strings"

	"ubunatic.com/loom/measure"
)

// Closed detects incomplete standalone Unicode box outlines. Ambiguous partial
// diagrams and cursor-addressed recordings are left to other checks.
func Closed(text string) error {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		line, ok := plain(lines[i])
		if !ok {
			continue
		}
		trim := strings.TrimSpace(line)
		r := []rune(trim)
		if len(r) < 3 || !strings.ContainsRune("┌╔", r[0]) || !strings.ContainsRune("┐╗", r[len(r)-1]) {
			continue
		}
		left := measure.StringWidth(line[:strings.Index(line, trim)])
		bottomLeft, bottomRight := '└', '┘'
		if r[0] == '╔' {
			bottomLeft, bottomRight = '╚', '╝'
		}
		for j := i + 1; j < len(lines); j++ {
			row, known := plain(lines[j])
			if !known {
				break
			}
			cells := columns(row)
			if left >= len(cells) {
				break
			}
			if cells[left] == bottomLeft {
				if !strings.ContainsRune(row, bottomRight) {
					return fmt.Errorf("unclosed box at line %d: missing bottom-right corner on line %d", i+1, j+1)
				}
				i = j
				break
			}
			if cells[left] != '│' && cells[left] != '║' && cells[left] != '├' && cells[left] != '╟' {
				break
			}
			if j == len(lines)-1 || (j == len(lines)-2 && lines[j+1] == "") {
				return fmt.Errorf("unclosed box at line %d: missing bottom border", i+1)
			}
		}
	}
	return nil
}

// plain removes SGR sequences; cursor movement makes a line indeterminate.
func plain(s string) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '\x1b' {
			b.WriteByte(s[i])
			i++
			continue
		}
		if i+1 >= len(s) || s[i+1] != '[' {
			return "", false
		}
		j := i + 2
		for j < len(s) && (s[j] < '@' || s[j] > '~') {
			j++
		}
		if j == len(s) || s[j] != 'm' {
			return "", false
		}
		i = j + 1
	}
	return b.String(), true
}

func columns(s string) []rune {
	var cells []rune
	for _, r := range s {
		w := measure.RuneWidth(r)
		for n := 0; n < w; n++ {
			cells = append(cells, r)
		}
	}
	return cells
}
