// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package measure exposes Loom's terminal-cell measurement policy to layout
// planners and applications without introducing a second width contract.
package measure

import (
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// useFastMeasure returns true unless LOOM_FAST_MEASURE is set to "0", "false", or "off".
func useFastMeasure() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("LOOM_FAST_MEASURE")))
	return v != "0" && v != "false" && v != "off"
}

// Size describes the visible terminal-cell dimensions of text lines.
type Size struct {
	Width  int
	Height int
}

// StringWidth returns the visible terminal-cell width of text according to
// Loom's rendering policy.
func StringWidth(text string) int {
	if useFastMeasure() {
		return StringWidthNew(text)
	}
	return StringWidthOld(text)
}

// StringWidthOld is the legacy StringWidth preserved for fallback and comparison.
func StringWidthOld(text string) int {
	w := 0
	for _, r := range plainTerminalTextOld(text) {
		w += RuneWidth(r)
	}
	return w
}

// StringWidthNew is the zero-allocation fast path for StringWidth.
func StringWidthNew(text string) int {
	if text == "" {
		return 0
	}
	// Fast path for plain printable ASCII
	isPureASCII := true
	for i := 0; i < len(text); i++ {
		b := text[i]
		if b < 0x20 || b > 0x7e {
			isPureASCII = false
			break
		}
	}
	if isPureASCII {
		return len(text)
	}

	w := 0
	i := 0
	n := len(text)
	for i < n {
		b := text[i]
		if b == 27 { // ESC
			i++
			if i >= n {
				break
			}
			switch text[i] {
			case '[':
				i++
				for i < n && (text[i] < '@' || text[i] > '~') {
					i++
				}
				if i < n {
					i++
				}
			case ']', 'P', '^', '_':
				i++
				for i < n {
					if text[i] == 7 {
						i++
						break
					}
					if text[i] == 27 && i+1 < n && text[i+1] == '\\' {
						i += 2
						break
					}
					i++
				}
			}
			continue
		}

		r, sz := utf8.DecodeRuneInString(text[i:])
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			i += sz
			continue
		}

		w += RuneWidth(r)
		i += sz
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
	if useFastMeasure() {
		return ClustersNew(text)
	}
	return ClustersOld(text)
}

// ClustersOld is the legacy Clusters implementation preserved for comparison.
func ClustersOld(text string) []string {
	var result []string
	for _, r := range plainTerminalTextOld(text) {
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

// ClustersNew is the zero-allocation cluster slice scanner.
func ClustersNew(text string) []string {
	if text == "" {
		return nil
	}

	// Fast path for pure printable ASCII
	isPureASCII := true
	for i := 0; i < len(text); i++ {
		b := text[i]
		if b < 0x20 || b > 0x7e {
			isPureASCII = false
			break
		}
	}
	if isPureASCII {
		res := make([]string, len(text))
		for i := 0; i < len(text); i++ {
			res[i] = text[i : i+1]
		}
		return res
	}

	var result []string
	plain := plainTerminalTextNew(text)
	if plain == "" {
		return nil
	}

	var clusterStart int
	var lastLen int
	var hasCluster bool

	i := 0
	n := len(plain)
	for i < n {
		r, sz := utf8.DecodeRuneInString(plain[i:])
		rw := RuneWidth(r)
		if rw == 0 {
			if hasCluster {
				// Extend previous cluster
				lastLen += sz
				result[len(result)-1] = plain[clusterStart : clusterStart+lastLen]
			}
		} else {
			clusterStart = i
			lastLen = sz
			hasCluster = true
			result = append(result, plain[clusterStart:clusterStart+sz])
		}
		i += sz
	}
	return result
}

func plainTerminalText(s string) string {
	if useFastMeasure() {
		return plainTerminalTextNew(s)
	}
	return plainTerminalTextOld(s)
}

func plainTerminalTextOld(s string) string {
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

func plainTerminalTextNew(s string) string {
	if s == "" {
		return ""
	}

	// Fast path for pure ASCII printable
	isPureASCII := true
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 0x20 || b > 0x7e {
			isPureASCII = false
			break
		}
	}
	if isPureASCII {
		return s
	}

	// Check if any filtering is needed at all
	hasEscOrControl := false
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == 27 || b < 0x20 || b == 0x7f {
			hasEscOrControl = true
			break
		}
	}

	if !hasEscOrControl {
		// Verify no utf8 control/Cf runes
		clean := true
		for _, r := range s {
			if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
				clean = false
				break
			}
		}
		if clean {
			return s
		}
	}

	var b strings.Builder
	b.Grow(len(s))

	i := 0
	n := len(s)
	for i < n {
		byteVal := s[i]
		if byteVal == 27 {
			i++
			if i >= n {
				break
			}
			switch s[i] {
			case '[':
				i++
				for i < n && (s[i] < '@' || s[i] > '~') {
					i++
				}
				if i < n {
					i++
				}
			case ']', 'P', '^', '_':
				i++
				for i < n {
					if s[i] == 7 {
						i++
						break
					}
					if s[i] == 27 && i+1 < n && s[i+1] == '\\' {
						i += 2
						break
					}
					i++
				}
			}
			continue
		}

		r, sz := utf8.DecodeRuneInString(s[i:])
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			i += sz
			continue
		}

		b.WriteRune(r)
		i += sz
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
	return Fit(strings.Join(clusters, ""), width-StringWidth(end)) + end
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
	start := Fit(marker, width)
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

// Fit keeps the longest prefix of text that fits within width terminal cells.
// It strips terminal controls and never splits a combining cluster or a wide
// glyph. A non-positive width returns an empty string.
func Fit(text string, width int) string {
	if width <= 0 {
		return ""
	}
	return fitClusters(Clusters(text), width)
}

// Pad fits text to at most width terminal cells and pads the remaining cells.
// When right is true, padding is placed before the text. Text is normalized by
// the same terminal-cell policy as Fit.
func Pad(text string, width int, right bool) string {
	if width <= 0 {
		return ""
	}
	text = Fit(text, width)
	padding := strings.Repeat(" ", width-StringWidth(text))
	if right {
		return padding + text
	}
	return text + padding
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
	if useFastMeasure() {
		return plainTerminalLinesNew(s)
	}
	return plainTerminalLinesOld(s)
}

func plainTerminalLinesOld(s string) []string {
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

func plainTerminalLinesNew(s string) []string {
	if s == "" {
		return []string{""}
	}

	// Fast path for plain ASCII
	isPureASCII := true
	for i := 0; i < len(s); i++ {
		b := s[i]
		if (b < 0x20 && b != '\n') || b > 0x7e {
			isPureASCII = false
			break
		}
	}
	if isPureASCII {
		return strings.Split(s, "\n")
	}

	var lines []string
	var current strings.Builder
	current.Grow(len(s))

	i := 0
	n := len(s)
	for i < n {
		b := s[i]
		if b == 27 {
			i++
			if i >= n {
				break
			}
			switch s[i] {
			case '[':
				i++
				for i < n && (s[i] < '@' || s[i] > '~') {
					i++
				}
				if i < n {
					i++
				}
			case ']', 'P', '^', '_':
				i++
				for i < n {
					if s[i] == 7 {
						i++
						break
					}
					if s[i] == 27 && i+1 < n && s[i+1] == '\\' {
						i += 2
						break
					}
					i++
				}
			}
			continue
		}

		if b == '\n' {
			lines = append(lines, current.String())
			current.Reset()
			i++
			continue
		}

		r, sz := utf8.DecodeRuneInString(s[i:])
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			i += sz
			continue
		}

		current.WriteRune(r)
		i += sz
	}
	lines = append(lines, current.String())
	return lines
}
