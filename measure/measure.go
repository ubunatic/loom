// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package measure exposes Loom's terminal-cell measurement policy to layout
// planners and applications without introducing a second width contract.
package measure

import (
	"strings"
	"unicode"
)

// Size describes the visible terminal-cell dimensions of text lines.
type Size struct {
	Width  int
	Height int
}

// StringWidth returns the visible terminal-cell width of text according to
// Loom's rendering policy.
func StringWidth(text string) int {
	w := 0
	for _, r := range plainTerminalText(text) {
		w += RuneWidth(r)
	}
	return w
}

// RuneWidth returns the visual column width of a single rune.
func RuneWidth(r rune) int {
	if unicode.IsControl(r) || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Cf, r) {
		return 0
	}
	if r >= 0x1100 && r <= 0x115f || r >= 0x2e80 && r <= 0xa4cf && r != 0x303f || r >= 0xac00 && r <= 0xd7a3 || r >= 0xf900 && r <= 0xfaff || r >= 0xfe10 && r <= 0xfe19 || r >= 0xfe30 && r <= 0xfe6f || r >= 0xff01 && r <= 0xff60 || r >= 0xffe0 && r <= 0xffe6 || r >= 0x20000 && r <= 0x3fffd {
		return 2
	}
	if r >= 0x1f000 && r <= 0x1faff {
		return 2
	}
	return 1
}

// Clusters returns printable base-rune clusters with combining marks attached.
// Emoji ZWJ sequences are intentionally not treated as one cluster.
func Clusters(text string) []string {
	var result []string
	for _, r := range plainTerminalText(text) {
		if RuneWidth(r) == 0 {
			if len(result) > 0 {
				result[len(result)-1] += string(r)
			}
		} else {
			result = append(result, string(r))
		}
	}
	return result
}

func plainTerminalText(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == 27 {
			i++
			if i >= len(runes) {
				break
			}
			switch runes[i] {
			case '[':
				for i++; i < len(runes); i++ {
					if runes[i] >= '@' && runes[i] <= '~' {
						break
					}
				}
			case ']', 'P', '^', '_':
				for i++; i < len(runes); i++ {
					if runes[i] == 7 {
						break
					}
					if runes[i] == 27 && i+1 < len(runes) && runes[i+1] == '\\' {
						i++
						break
					}
				}
			}
			continue
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Lines returns the visible dimensions of a multi-line string. An empty
// string is one empty line.
func Lines(text string) Size {
	lines := plainTerminalLines(text)
	result := Size{Height: len(lines)}
	for _, line := range lines {
		if width := StringWidth(line); width > result.Width {
			result.Width = width
		}
	}
	return result
}

// Truncate fits text to a terminal-cell budget. Terminal styling is stripped;
// combining marks stay with their base and wide glyphs are never split.
func Truncate(text string, width int, marker string) string {
	if width <= 0 {
		return ""
	}
	clusters := Clusters(text)
	if StringWidth(strings.Join(clusters, "")) <= width {
		return strings.Join(clusters, "")
	}
	end := fitClusters(Clusters(marker), width)
	return fitClusters(clusters, width-StringWidth(end)) + end
}

// TruncateLeft fits text to a terminal-cell budget while keeping its trailing
// content. Terminal styling is stripped and wide glyphs are never split.
func TruncateLeft(text string, width int, marker string) string {
	if width <= 0 {
		return ""
	}
	clusters := Clusters(text)
	plain := strings.Join(clusters, "")
	if StringWidth(plain) <= width {
		return plain
	}
	start := fitClusters(Clusters(marker), width)
	remain := width - StringWidth(start)
	if remain <= 0 {
		return start
	}
	var trailing []string
	used := 0
	for i := len(clusters) - 1; i >= 0; i-- {
		w := StringWidth(clusters[i])
		if used+w > remain {
			break
		}
		trailing = append([]string{clusters[i]}, trailing...)
		used += w
	}
	return start + strings.Join(trailing, "")
}

func fitClusters(clusters []string, width int) string {
	var result strings.Builder
	for _, cluster := range clusters {
		cells := StringWidth(cluster)
		if cells > width {
			break
		}
		result.WriteString(cluster)
		width -= cells
	}
	return result.String()
}

func plainTerminalLines(s string) []string {
	lines := []string{""}
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == 27 {
			i++
			if i >= len(runes) {
				break
			}
			switch runes[i] {
			case '[':
				for i++; i < len(runes); i++ {
					if runes[i] >= '@' && runes[i] <= '~' {
						break
					}
				}
			case ']', 'P', '^', '_':
				for i++; i < len(runes); i++ {
					if runes[i] == 7 {
						break
					}
					if runes[i] == 27 && i+1 < len(runes) && runes[i+1] == '\\' {
						i++
						break
					}
				}
			}
			continue
		}
		if r == '\n' {
			lines = append(lines, "")
			continue
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			continue
		}
		lines[len(lines)-1] += string(r)
	}
	return lines
}
