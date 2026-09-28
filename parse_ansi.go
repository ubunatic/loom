// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

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
// and cell mapping with continuation cells for wide characters with minimal allocations.
func ParseANSINew(s string) []Cell {
	if len(s) == 0 {
		return nil
	}

	cells := make([]Cell, 0, len(s))
	style := Style{}

	i := 0
	n := len(s)
	for i < n {
		b := s[i]

		// Check for CSI sequence: ESC [ ... m
		if b == '\x1b' && i+1 < n && s[i+1] == '[' {
			// Find the end of the sequence (terminated by a character in range @ to ~)
			j := i + 2
			for j < n && (s[j] < '@' || s[j] > '~') {
				j++
			}

			if j < n {
				if s[j] == 'm' {
					// SGR (Select Graphic Rendition) sequence
					params := s[i+2 : j]
					style = applySGRSequenceFast(style, params)
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
		if b == '\x1b' && i+2 < n && s[i+1] == '(' {
			i += 3
			continue
		}

		// Fast-path printable ASCII (0x20 .. 0x7e) when not followed by multi-byte UTF-8 combining marks/modifiers
		if b >= 0x20 && b <= 0x7e && (i+1 == n || s[i+1] < 0x80) {
			cells = append(cells, Cell{Text: s[i : i+1], Style: style})
			i++
			continue
		}

		// Skip control characters (< 0x20 or 0x7f)
		if b < 0x20 || b == 0x7f {
			i++
			continue
		}

		r, sz := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && sz == 1 {
			i++
			continue
		}

		j := i + sz
		if r >= 0x1F1E6 && r <= 0x1F1FF {
			if j < n {
				r2, sz2 := utf8.DecodeRuneInString(s[j:])
				if r2 >= 0x1F1E6 && r2 <= 0x1F1FF {
					j += sz2
				}
			}
		} else {
			for j < n {
				if s[j] == '\x1b' {
					break
				}
				nextR, nextSz := utf8.DecodeRuneInString(s[j:])
				if nextR == 0x200D {
					j += nextSz
					if j < n {
						_, zwjNextSz := utf8.DecodeRuneInString(s[j:])
						j += zwjNextSz
					}
					continue
				}
				if nextR == 0xFE0F || nextR == 0xFE0E || unicode.Is(unicode.Mn, nextR) || unicode.Is(unicode.Me, nextR) {
					j += nextSz
					continue
				}
				break
			}
		}

		cluster := s[i:j]
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

// parseSGRInt parses a non-negative decimal integer from s without allocating memory.
func parseSGRInt(s string) (int, bool) {
	if len(s) == 0 {
		return 0, false
	}
	v := 0
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < '0' || b > '9' {
			return 0, false
		}
		v = v*10 + int(b-'0')
		if v > 1000000 {
			return 0, false
		}
	}
	return v, true
}

// applySGRSequenceFast applies SGR parameters to a style without heap allocations.
func applySGRSequenceFast(style Style, params string) Style {
	if len(params) == 0 {
		return Style{}
	}

	var codes [32]int
	numCodes := 0

	start := 0
	n := len(params)
	for i := 0; i <= n; i++ {
		if i == n || params[i] == ';' {
			part := params[start:i]
			code := 0
			if len(part) > 0 {
				v, ok := parseSGRInt(part)
				if !ok {
					start = i + 1
					continue
				}
				code = v
			}
			if numCodes < len(codes) {
				codes[numCodes] = code
				numCodes++
			}
			start = i + 1
		}
	}

	for i := 0; i < numCodes; i++ {
		code := codes[i]
		switch {
		case code == 0:
			style = Style{}
		case code == 1:
			style.Bold = true
		case code == 2:
			style.Dim = true
		case code == 4:
			style.Underline = true
		case code == 22:
			style.Bold = false
			style.Dim = false
		case code == 24:
			style.Underline = false
		case code >= 30 && code <= 37:
			style.FG = ColorIndex(uint8(code - 30))
		case code >= 90 && code <= 97:
			style.FG = ColorIndex(uint8(code - 90 + 8))
		case code >= 40 && code <= 47:
			style.BG = ColorIndex(uint8(code - 40))
		case code >= 100 && code <= 107:
			style.BG = ColorIndex(uint8(code - 100 + 8))
		case code == 39:
			style.FG = ColorReset()
		case code == 49:
			style.BG = ColorReset()
		case code == 38 && i+2 < numCodes && codes[i+1] == 5:
			v := codes[i+2]
			if v >= 0 && v <= 255 {
				style.FG = ColorIndex(uint8(v))
			}
			i += 2
		case code == 48 && i+2 < numCodes && codes[i+1] == 5:
			v := codes[i+2]
			if v >= 0 && v <= 255 {
				style.BG = ColorIndex(uint8(v))
			}
			i += 2
		case code == 38 && i+4 < numCodes && codes[i+1] == 2:
			r, g, b := codes[i+2], codes[i+3], codes[i+4]
			if r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
				style.FG = ColorRGB(uint8(r), uint8(g), uint8(b))
			}
			i += 4
		case code == 48 && i+4 < numCodes && codes[i+1] == 2:
			r, g, b := codes[i+2], codes[i+3], codes[i+4]
			if r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
				style.BG = ColorRGB(uint8(r), uint8(g), uint8(b))
			}
			i += 4
		}
	}

	return style
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

	// Split by semicolon; an empty part (e.g. ESC[;1m or ESC[1;m) counts as 0
	parts := strings.Split(params, ";")
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		code := 0
		if p != "" {
			var err error
			code, err = strconv.Atoi(p)
			if err != nil {
				continue
			}
		}

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

		case code == 38 && i+2 < len(parts) && parts[i+1] == "5":
			// 256-color foreground: 38;5;n
			v, err := strconv.Atoi(parts[i+2])
			// Ignore out-of-range values (0-255 only)
			if err == nil && v >= 0 && v <= 255 {
				style.FG = ColorIndex(uint8(v))
			}
			i += 2

		case code == 48 && i+2 < len(parts) && parts[i+1] == "5":
			// 256-color background: 48;5;n
			v, err := strconv.Atoi(parts[i+2])
			// Ignore out-of-range values (0-255 only)
			if err == nil && v >= 0 && v <= 255 {
				style.BG = ColorIndex(uint8(v))
			}
			i += 2

		case code == 38 && i+4 < len(parts) && parts[i+1] == "2":
			// 24-bit RGB foreground: 38;2;r;g;b
			r, errR := strconv.Atoi(parts[i+2])
			g, errG := strconv.Atoi(parts[i+3])
			b, errB := strconv.Atoi(parts[i+4])
			// Ignore if any value is out of range (0-255 only)
			if errR == nil && errG == nil && errB == nil &&
				r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
				style.FG = ColorRGB(uint8(r), uint8(g), uint8(b))
			}
			i += 4

		case code == 48 && i+4 < len(parts) && parts[i+1] == "2":
			// 24-bit RGB background: 48;2;r;g;b
			r, errR := strconv.Atoi(parts[i+2])
			g, errG := strconv.Atoi(parts[i+3])
			b, errB := strconv.Atoi(parts[i+4])
			// Ignore if any value is out of range (0-255 only)
			if errR == nil && errG == nil && errB == nil &&
				r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
				style.BG = ColorRGB(uint8(r), uint8(g), uint8(b))
			}
			i += 4
		}
	}

	return style
}
