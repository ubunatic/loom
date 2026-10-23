// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
	"strings"
	"unicode"

	"ubunatic.com/loom/ansi"
	"ubunatic.com/loom/measure"
)

// useFastANSI returns true unless LOOM_FAST_ANSI is set to "0", "false", or "off".
func useFastANSI() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("LOOM_FAST_ANSI")))
	return v != "0" && v != "false" && v != "off"
}

// ParseANSI decodes an ANSI-formatted string into a slice of Cells.
// By default, it uses the optimized zero-allocation ParseANSINew. If LOOM_FAST_ANSI
// is set to "0", "false", or "off", it falls back to ParseANSIOld.
func ParseANSI(s string) []Cell {
	if useFastANSI() {
		return ParseANSINew(s)
	}
	return ParseANSIOld(s)
}

// ParseANSINew is the optimized ANSI parser that performs cluster collection
// and cell mapping with continuation cells for wide characters.
func ParseANSINew(s string) []Cell {
	var cells []Cell
	style := Style{}

	rs := []rune(s)
	n := len(rs)
	for i := 0; i < n; {
		// Check for CSI sequence: ESC [ ... m
		if rs[i] == '\x1b' && i+1 < n && rs[i+1] == '[' {
			// Find the end of the sequence (terminated by a letter in range @ to ~)
			j := i + 2
			for j < n && (rs[j] < '@' || rs[j] > '~') {
				j++
			}

			if j < n {
				// We have a complete sequence
				if rs[j] == 'm' {
					// SGR (Select Graphic Rendition) sequence
					params := string(rs[i+2 : j])
					style = applySGRSequence(style, params)
					i = j + 1
					continue
				} else {
					// Non-SGR escape sequence (e.g., cursor movement); skip it
					i = j + 1
					continue
				}
			}
			// Incomplete sequence; skip the ESC and continue
			i++
			continue
		}

		// Check for character set designation: ESC ( ... (skip these)
		if rs[i] == '\x1b' && i+2 < n && rs[i+1] == '(' {
			i += 3
			continue
		}

		r := rs[i]
		if unicode.IsControl(r) {
			i++
			continue
		}

		// Collect base rune plus any subsequent combining marks / variation selectors / ZWJs
		j := i + 1
		if r >= 0x1F1E6 && r <= 0x1F1FF && j < n && rs[j] >= 0x1F1E6 && rs[j] <= 0x1F1FF {
			j++
		} else {
			for j < n {
				next := rs[j]
				if next == 0x200D {
					j++
					if j < n && measure.JoinsAfterZWJ(rs[j]) {
						j++
					}
					continue
				}
				if next == 0xFE0F || next == 0xFE0E || unicode.Is(unicode.Mn, next) || unicode.Is(unicode.Me, next) {
					j++
					continue
				}
				break
			}
		}

		cluster := string(rs[i:j])
		w := measure.StringWidth(cluster)
		if w > 0 {
			cells = append(cells, Cell{Text: cluster, Style: style})
			for k := 1; k < w; k++ {
				cells = append(cells, Cell{Style: style, Continuation: true})
			}
		}
		i = j
	}

	return cells
}

// ParseANSIOld is the legacy ANSI parser preserved for comparison and testing.
func ParseANSIOld(s string) []Cell {
	var cells []Cell
	style := Style{}

	rs := []rune(s)
	n := len(rs)
	for i := 0; i < n; {
		// Check for CSI sequence: ESC [ ... m
		if rs[i] == '\x1b' && i+1 < n && rs[i+1] == '[' {
			// Find the end of the sequence (terminated by a letter in range @ to ~)
			j := i + 2
			for j < n && (rs[j] < '@' || rs[j] > '~') {
				j++
			}

			if j < n {
				// We have a complete sequence
				if rs[j] == 'm' {
					// SGR (Select Graphic Rendition) sequence
					params := string(rs[i+2 : j])
					style = applySGRSequence(style, params)
					i = j + 1
					continue
				} else {
					// Non-SGR escape sequence (e.g., cursor movement); skip it
					i = j + 1
					continue
				}
			}
			// Incomplete sequence; skip the ESC and continue
			i++
			continue
		}

		// Check for character set designation: ESC ( ... (skip these)
		if rs[i] == '\x1b' && i+2 < n && rs[i+1] == '(' {
			i += 3
			continue
		}

		r := rs[i]
		if unicode.IsControl(r) {
			i++
			continue
		}

		j := i + 1
		if r >= 0x1F1E6 && r <= 0x1F1FF && j < n && rs[j] >= 0x1F1E6 && rs[j] <= 0x1F1FF {
			j++
		} else {
			for j < n {
				next := rs[j]
				if next == 0x200D {
					j++
					if j < n && measure.JoinsAfterZWJ(rs[j]) {
						j++
					}
					continue
				}
				if next == 0xFE0F || next == 0xFE0E || unicode.Is(unicode.Mn, next) || unicode.Is(unicode.Me, next) {
					j++
					continue
				}
				break
			}
		}

		cluster := string(rs[i:j])
		w := StringWidth(cluster)
		if w > 0 {
			cells = append(cells, Cell{Text: cluster, Style: style})
			for k := 1; k < w; k++ {
				cells = append(cells, Cell{Style: style, Continuation: true})
			}
		}
		i = j
	}

	return cells
}

// applySGRSequence applies SGR parameters to a style.
func applySGRSequence(style Style, params string) Style {
	return ansi.ParseSGR(style, params)
}
