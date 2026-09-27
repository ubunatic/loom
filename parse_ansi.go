// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
	"strings"
	"unicode"

	"codeberg.org/ubunatic/loom/measure"
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
					if j < n {
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
					if j < n {
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
	// Empty parameter list (bare ESC[m) counts as code 0 (reset)
	if params == "" {
		return Style{}
	}

	var args [16]int
	var valid [16]bool
	nArgs := 0

	start := 0
	for i := 0; i <= len(params); i++ {
		if i == len(params) || params[i] == ';' {
			if nArgs < len(args) {
				if i == start {
					// empty param (e.g. ESC[;1m or ESC[1;m) counts as 0
					args[nArgs] = 0
					valid[nArgs] = true
				} else {
					val := 0
					isValid := true
					for k := start; k < i; k++ {
						c := params[k]
						if c >= '0' && c <= '9' {
							val = val*10 + int(c-'0')
						} else {
							isValid = false
							break
						}
					}
					if isValid {
						args[nArgs] = val
						valid[nArgs] = true
					}
				}
				nArgs++
			}
			start = i + 1
		}
	}

	for i := 0; i < nArgs; i++ {
		if !valid[i] {
			continue
		}
		code := args[i]

		switch {
		case code == 0:
			// Reset all attributes
			style = Style{}

		case code == 1:
			// Bold
			style.Bold = true

		case code == 2:
			// Dim
			style.Dim = true

		case code == 4:
			// Underline
			style.Underline = true

		case code == 22:
			// Bold and dim off
			style.Bold = false
			style.Dim = false

		case code == 24:
			// Underline off
			style.Underline = false

		case code >= 30 && code <= 37:
			// 16-color foreground (30-37)
			style.FG = ColorIndex(uint8(code - 30))

		case code >= 90 && code <= 97:
			// Bright 16-color foreground (90-97)
			style.FG = ColorIndex(uint8(code - 90 + 8))

		case code >= 40 && code <= 47:
			// 16-color background (40-47)
			style.BG = ColorIndex(uint8(code - 40))

		case code >= 100 && code <= 107:
			// Bright 16-color background (100-107)
			style.BG = ColorIndex(uint8(code - 100 + 8))

		case code == 39:
			// Reset foreground to default
			style.FG = ColorReset()

		case code == 49:
			// Reset background to default
			style.BG = ColorReset()

		case code == 38 && i+2 < nArgs && valid[i+1] && args[i+1] == 5:
			// 256-color foreground: 38;5;n
			if valid[i+2] {
				v := args[i+2]
				if v >= 0 && v <= 255 {
					style.FG = ColorIndex(uint8(v))
				}
			}
			i += 2

		case code == 48 && i+2 < nArgs && valid[i+1] && args[i+1] == 5:
			// 256-color background: 48;5;n
			if valid[i+2] {
				v := args[i+2]
				if v >= 0 && v <= 255 {
					style.BG = ColorIndex(uint8(v))
				}
			}
			i += 2

		case code == 38 && i+4 < nArgs && valid[i+1] && args[i+1] == 2:
			// 24-bit RGB foreground: 38;2;r;g;b
			if valid[i+2] && valid[i+3] && valid[i+4] {
				r, g, b := args[i+2], args[i+3], args[i+4]
				if r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
					style.FG = ColorRGB(uint8(r), uint8(g), uint8(b))
				}
			}
			i += 4

		case code == 48 && i+4 < nArgs && valid[i+1] && args[i+1] == 2:
			// 24-bit RGB background: 48;2;r;g;b
			if valid[i+2] && valid[i+3] && valid[i+4] {
				r, g, b := args[i+2], args[i+3], args[i+4]
				if r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
					style.BG = ColorRGB(uint8(r), uint8(g), uint8(b))
				}
			}
			i += 4
		}
	}

	return style
}
