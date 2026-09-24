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
	for _, c := range ClustersOld(text) {
		w += ClusterWidth(c)
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
		if unicode.IsControl(r) || (unicode.Is(unicode.Cf, r) && !isFormatRune(r)) {
			i += sz
			continue
		}

		// Check for regional indicator pair (flag)
		if isRegionalIndicator(r) {
			nextIdx := i + sz
			if nextIdx < n {
				r2, sz2 := utf8.DecodeRuneInString(text[nextIdx:])
				if isRegionalIndicator(r2) {
					w += ActiveEmojiSpec().FlagDefaultWidth
					i = nextIdx + sz2
					continue
				}
			}
		}

		// Check for base rune + modifiers / ZWJ / VS16
		start := i
		clusterWidth := RuneWidth(r)
		hasVS16 := false
		hasVS15 := false
		hasZWJ := false

		i += sz
		for i < n {
			if text[i] == 27 {
				break
			}
			nr, nsz := utf8.DecodeRuneInString(text[i:])
			if nr == 0x200D {
				hasZWJ = true
				i += nsz
				if i < n {
					_, jsz := utf8.DecodeRuneInString(text[i:])
					i += jsz
				}
				continue
			}
			if nr == 0xFE0F {
				hasVS16 = true
				i += nsz
				continue
			}
			if nr == 0xFE0E {
				hasVS15 = true
				i += nsz
				continue
			}
			if unicode.Is(unicode.Mn, nr) || unicode.Is(unicode.Me, nr) || (unicode.Is(unicode.Cf, nr) && !isFormatRune(nr)) {
				i += nsz
				continue
			}
			break
		}

		if gw, ok := getGlyphOverride(text[start:i]); ok {
			w += gw
		} else if hasZWJ {
			w += ActiveEmojiSpec().ZWJDefaultWidth
		} else if hasVS16 {
			w += ActiveEmojiSpec().VS16DefaultWidth
		} else if hasVS15 {
			w += 1
		} else {
			w += clusterWidth
		}
	}
	return w
}

func isFormatRune(r rune) bool {
	return r == 0xFE0F || r == 0xFE0E || r == 0x200D || r == 0x200C
}

func isRegionalIndicator(r rune) bool {
	return r >= 0x1F1E6 && r <= 0x1F1FF
}

// ClusterWidth returns the visual column width of a single text cluster.
func ClusterWidth(cluster string) int {
	if cluster == "" {
		return 0
	}
	if w, ok := getGlyphOverride(cluster); ok {
		return w
	}
	rs := []rune(cluster)
	if len(rs) == 0 {
		return 0
	}
	// Flag sequence: two regional indicator symbols (U+1F1E6..U+1F1FF)
	if len(rs) == 2 && isRegionalIndicator(rs[0]) && isRegionalIndicator(rs[1]) {
		return ActiveEmojiSpec().FlagDefaultWidth
	}
	// ZWJ sequences: rendered as single joined emoji in modern terminals
	if strings.ContainsRune(cluster, '\u200D') {
		return ActiveEmojiSpec().ZWJDefaultWidth
	}
	// VS16 (Variation Selector-16) emoji presentation: width 2
	if strings.ContainsRune(cluster, '\uFE0F') {
		return ActiveEmojiSpec().VS16DefaultWidth
	}
	// VS15 (Variation Selector-15) text presentation: width 1
	if strings.ContainsRune(cluster, '\uFE0E') {
		return 1
	}
	return RuneWidth(rs[0])
}

// RuneWidth returns the visual column width of a single rune.
func RuneWidth(r rune) int {
	if unicode.IsControl(r) || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) {
		return 0
	}
	if w, ok := getRuneOverride(r); ok {
		return w
	}
	if unicode.Is(unicode.Cf, r) {
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
func Clusters(text string) []string {
	if useFastMeasure() {
		return ClustersNew(text)
	}
	return ClustersOld(text)
}

// ClustersOld is the legacy Clusters implementation preserved for comparison.
func ClustersOld(text string) []string {
	plain := plainTerminalTextOld(text)
	if plain == "" {
		return nil
	}
	var result []string
	rs := []rune(plain)
	n := len(rs)
	for i := 0; i < n; {
		if (unicode.Is(unicode.Mn, rs[i]) || unicode.Is(unicode.Me, rs[i])) && !isFormatRune(rs[i]) {
			if len(result) > 0 {
				result[len(result)-1] += string(rs[i])
			}
			i++
			continue
		}
		if isRegionalIndicator(rs[i]) && i+1 < n && isRegionalIndicator(rs[i+1]) {
			result = append(result, string(rs[i:i+2]))
			i += 2
			continue
		}
		start := i
		i++
		for i < n {
			r := rs[i]
			if r == 0x200D {
				i++
				if i < n {
					i++
				}
				continue
			}
			if r == 0xFE0F || r == 0xFE0E || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || (unicode.Is(unicode.Cf, r) && !isFormatRune(r)) {
				i++
				continue
			}
			break
		}
		result = append(result, string(rs[start:i]))
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

	plain := plainTerminalTextNew(text)
	if plain == "" {
		return nil
	}

	var result []string
	rs := []rune(plain)
	n := len(rs)
	for i := 0; i < n; {
		if (unicode.Is(unicode.Mn, rs[i]) || unicode.Is(unicode.Me, rs[i])) && !isFormatRune(rs[i]) {
			if len(result) > 0 {
				result[len(result)-1] += string(rs[i])
			}
			i++
			continue
		}
		if isRegionalIndicator(rs[i]) && i+1 < n && isRegionalIndicator(rs[i+1]) {
			result = append(result, string(rs[i:i+2]))
			i += 2
			continue
		}
		start := i
		i++
		for i < n {
			r := rs[i]
			if r == 0x200D {
				i++
				if i < n {
					i++
				}
				continue
			}
			if r == 0xFE0F || r == 0xFE0E || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || (unicode.Is(unicode.Cf, r) && !isFormatRune(r)) {
				i++
				continue
			}
			break
		}
		result = append(result, string(rs[start:i]))
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
		if unicode.IsControl(r) || (unicode.Is(unicode.Cf, r) && !isFormatRune(r)) {
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
			if unicode.IsControl(r) || (unicode.Is(unicode.Cf, r) && !isFormatRune(r)) {
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
		if unicode.IsControl(r) || (unicode.Is(unicode.Cf, r) && !isFormatRune(r)) {
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
	return start + tailClusters(clusters, remain)
}

func tailClusters(clusters []string, width int) string {
	var result []string
	current := 0
	for i := len(clusters) - 1; i >= 0; i-- {
		w := StringWidth(clusters[i])
		if current+w > width {
			break
		}
		result = append([]string{clusters[i]}, result...)
		current += w
	}
	return strings.Join(result, "")
}

func fitClusters(clusters []string, width int) string {
	var result []string
	current := 0
	for _, c := range clusters {
		w := StringWidth(c)
		if current+w > width {
			break
		}
		result = append(result, c)
		current += w
	}
	return strings.Join(result, "")
}

// Fit returns the longest prefix of text that fits within width terminal cells.
// Terminal styling is stripped; combining marks and wide glyphs are never split.
func Fit(text string, width int) string {
	if width <= 0 {
		return ""
	}
	clusters := Clusters(text)
	return fitClusters(clusters, width)
}

// Pad extends text to width terminal cells using spaces.
func Pad(text string, width int, right bool) string {
	w := StringWidth(text)
	if w >= width {
		return text
	}
	padding := strings.Repeat(" ", width-w)
	if right {
		return padding + text
	}
	return text + padding
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
		if unicode.IsControl(r) || (unicode.Is(unicode.Cf, r) && !isFormatRune(r)) {
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
		if unicode.IsControl(r) || (unicode.Is(unicode.Cf, r) && !isFormatRune(r)) {
			i += sz
			continue
		}

		current.WriteRune(r)
		i += sz
	}
	lines = append(lines, current.String())
	return lines
}
