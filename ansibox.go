// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strings"

	"ubunatic.com/loom/measure"
)

// ValidateAnsiBox checks that rows which form a Unicode box have consistent
// visual widths and aligned left and right borders. Text without box drawing
// characters is accepted unchanged.
func ValidateAnsiBox(text string) error {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	type edge struct {
		line, left, right, width int
		kind, first, last        rune
	}
	var framed []edge
	for i, line := range lines {
		// Cursor motion changes where cells land on screen, so only that
		// physical line is outside the scope of this line-oriented check.
		if hasNonSGRANSI(line) {
			continue
		}
		plain := stripANSIForBox(line)
		left, right := -1, -1
		col := 0
		rs := []rune(plain)
		for j := 0; j < len(rs); {
			r := rs[j]
			if r >= 0x1F1E6 && r <= 0x1F1FF && j+1 < len(rs) && rs[j+1] >= 0x1F1E6 && rs[j+1] <= 0x1F1FF {
				col += measure.StringWidth(string(rs[j : j+2]))
				j += 2
				continue
			}
			if isBoxVertical(r) || isBoxCorner(r) {
				if left < 0 {
					left = col
				}
				right = col
			}
			col += measure.RuneWidth(r)
			j++
		}
		trimmed := strings.TrimSpace(plain)
		if left >= 0 && len(trimmed) > 1 && isBoxEdge(firstRune(trimmed)) && isBoxEdge(lastRune(trimmed)) {
			kind := boxRowKind(firstRune(trimmed), lastRune(trimmed))
			framed = append(framed, edge{line: i + 1, left: left, right: right, width: col, kind: kind, first: firstRune(trimmed), last: lastRune(trimmed)})
		}
	}
	if len(framed) < 2 {
		return nil
	}
	var openLine int
	for i := 0; i < len(framed); i++ {
		row := framed[i]
		if i > 0 && row.line != framed[i-1].line+1 {
			if openLine > 0 {
				foundBottom := false
				for g := framed[i-1].line; g < row.line-1 && g < len(lines); g++ {
					prevLinePlain := stripANSIForBox(lines[g])
					if strings.ContainsRune(prevLinePlain, '└') || strings.ContainsRune(prevLinePlain, '╚') || strings.ContainsRune(prevLinePlain, '╰') || strings.ContainsRune(prevLinePlain, '+') {
						foundBottom = true
						break
					}
				}
				if !foundBottom {
					return fmt.Errorf("boxed line %d: unclosed box from line %d", framed[i-1].line, openLine)
				}
			}
			openLine = 0
		}
		if row.kind == 't' {
			if openLine > 0 {
				return fmt.Errorf("boxed line %d: unclosed box from line %d", row.line-1, openLine)
			}
			openLine = row.line
		} else if row.kind == 'b' {
			if openLine == 0 {
				startLine := row.line
				for j := i - 1; j >= 0 && framed[j].line == framed[j+1].line-1; j-- {
					startLine = framed[j].line
				}

				if startLine > 1 {
					foundTop := false
					for g := startLine-2; g >= 0 && g < len(lines); g-- {
						prevLinePlain := stripANSIForBox(lines[g])
						if strings.ContainsRune(prevLinePlain, '┌') || strings.ContainsRune(prevLinePlain, '╔') || strings.ContainsRune(prevLinePlain, '╭') || strings.ContainsRune(prevLinePlain, '+') {
							foundTop = true
							break
						}
						if strings.TrimSpace(prevLinePlain) == "" {
							break
						}
					}
					if !foundTop {
						return fmt.Errorf("boxed line %d: unclosed box from line %d", row.line, startLine)
					}
				} else {
					return fmt.Errorf("boxed line %d: unclosed box from line %d", row.line, startLine)
				}
			}
			openLine = 0
		}
	}
	if openLine > 0 {
		foundBottom := false
		for g := framed[len(framed)-1].line; g < len(lines); g++ {
			prevLinePlain := stripANSIForBox(lines[g])
			if strings.ContainsRune(prevLinePlain, '└') || strings.ContainsRune(prevLinePlain, '╚') || strings.ContainsRune(prevLinePlain, '╰') || strings.ContainsRune(prevLinePlain, '+') {
				foundBottom = true
				break
			}
		}
		if !foundBottom {
			return fmt.Errorf("boxed line %d: unclosed box from line %d", framed[len(framed)-1].line, openLine)
		}
	}
	// Compare rows within a box. A bottom row followed by another top row starts
	// a new box, even when there is no blank line between the two. A nested box
	// that closes or opens at either end of a row (a dialog inside a grid cell)
	// also ends the comparison, because its corner moves that row's edge.
	for i := 1; i < len(framed); i++ {
		base, row := framed[i-1], framed[i]
		if row.line != base.line+1 || base.kind == 'b' || row.kind == 't' {
			continue
		}
		if isBoxBottomLeft(base.first) || isBoxBottomRight(base.last) || isBoxTopLeft(row.first) || isBoxTopRight(row.last) {
			continue
		}
		if row.left != base.left || row.right != base.right || row.width != base.width {
			return fmt.Errorf("boxed line %d: boundaries %d..%d width %d, want %d..%d width %d from line %d", row.line, row.left, row.right, row.width, base.left, base.right, base.width, base.line)
		}
	}
	return nil
}

func boxRowKind(first, last rune) rune {
	if isBoxTopLeft(first) && isBoxTopRight(last) {
		return 't'
	}
	if isBoxBottomLeft(first) && isBoxBottomRight(last) {
		return 'b'
	}
	return 'm'
}

func isBoxTopLeft(r rune) bool { return r == '┌' || r == '╔' }

func isBoxTopRight(r rune) bool { return r == '┐' || r == '╗' }

func isBoxBottomLeft(r rune) bool { return r == '└' || r == '╚' }

func isBoxBottomRight(r rune) bool { return r == '┘' || r == '╝' }

func hasNonSGRANSI(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != 0x1b {
			continue
		}
		if i+1 >= len(s) || s[i+1] != '[' {
			return true
		}
		j := i + 2
		for j < len(s) && (s[j] < '@' || s[j] > '~') {
			j++
		}
		if j >= len(s) || s[j] != 'm' {
			return true
		}
		i = j
	}
	return false
}

func isBoxVertical(r rune) bool { return r == '│' || r == '║' || r == '┃' || r == '|' }

func isBoxCorner(r rune) bool {
	return strings.ContainsRune("┌┐└┘╔╗╚╝╭╮╰╯├┤┬┴┼╟╢╤╧╪", r)
}

func isBoxEdge(r rune) bool { return isBoxVertical(r) || isBoxCorner(r) }

func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}

func lastRune(s string) rune {
	var last rune
	for _, r := range s {
		last = r
	}
	return last
}

func stripANSIForBox(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] != 0x1b {
			b.WriteByte(s[i])
			i++
			continue
		}
		i++
		if i >= len(s) {
			break
		}
		if s[i] == '[' {
			i++
			for i < len(s) && (s[i] < '@' || s[i] > '~') {
				i++
			}
			if i < len(s) {
				i++
			}
		} else {
			i++
		}
	}
	return b.String()
}
