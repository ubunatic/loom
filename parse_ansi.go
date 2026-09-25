// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
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

// ParseANSINew is the optimized zero-allocation ANSI parser that performs in-place
// zero-copy string scanning over s and pre-allocates cell slice capacity.
func ParseANSINew(s string) []Cell {
	if s == "" {
		return nil
	}
	cells := make([]Cell, 0, len(s))
	style := Style{}

	i := 0
	n := len(s)
	for i < n {
		b := s[i]

		// Fast path for printable ASCII characters (0x20 to 0x7e, excluding ESC 0x1b)
		// that are not followed by a non-ASCII byte (which could be a combining mark).
		if b >= 0x20 && b <= 0x7e && (i+1 == n || s[i+1] < 0x80) {
			cells = append(cells, Cell{Text: s[i : i+1], Style: style})
			i++
			continue
		}

		// Check for CSI sequence: ESC [ ... m
		if b == '\x1b' && i+1 < n && s[i+1] == '[' {
			// Find the end of the sequence (terminated by a letter in range @ to ~)
			j := i + 2
			for j < n && (s[j] < '@' || s[j] > '~') {
				j++
			}

			if j < n {
				// We have a complete sequence
				if s[j] == 'm' {
					// SGR (Select Graphic Rendition) sequence
					params := s[i+2 : j]
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
		if b == '\x1b' && i+2 < n && s[i+1] == '(' {
			i += 3
			continue
		}

		// Decode UTF-8 rune starting at s[i:]
		r, sz := utf8.DecodeRuneInString(s[i:])
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			i += sz
			continue
		}

		w := measure.RuneWidth(r)
		if w == 0 {
			// Standalone combining mark without base
			i += sz
			continue
		}

		// Collect base rune plus any subsequent zero-width combining marks into a cluster
		j := i + sz
		for j < n {
			nextR, nextSz := utf8.DecodeRuneInString(s[j:])
			if unicode.Is(unicode.Mn, nextR) || unicode.Is(unicode.Me, nextR) {
				j += nextSz
			} else {
				break
			}
		}

		cluster := s[i:j]
		if w == 2 {
			// Wide character: add lead cell and continuation cell
			cells = append(cells, Cell{Text: cluster, Style: style})
			cells = append(cells, Cell{Style: style, Continuation: true})
		} else {
			// Normal width character (w == 1)
			cells = append(cells, Cell{Text: cluster, Style: style})
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
	for i := 0; i < len(rs); {
		// Check for CSI sequence: ESC [ ... m
		if rs[i] == '\x1b' && i+1 < len(rs) && rs[i+1] == '[' {
			// Find the end of the sequence (terminated by a letter in range @ to ~)
			j := i + 2
			for j < len(rs) && (rs[j] < '@' || rs[j] > '~') {
				j++
			}

			if j < len(rs) {
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
		if rs[i] == '\x1b' && i+2 < len(rs) && rs[i+1] == '(' {
			i += 3
			continue
		}

		// Regular character: add to cells
		// Use text clusters to handle combining marks properly
		clusters := measure.Clusters(string(rs[i:]))
		if len(clusters) > 0 {
			cluster := clusters[0]
			w := StringWidth(cluster)

			if w == 2 {
				// Wide character: add lead cell and continuation cell
				cells = append(cells, Cell{Text: cluster, Style: style})
				cells = append(cells, Cell{Style: style, Continuation: true})
			} else if w == 1 {
				// Normal width character
				cells = append(cells, Cell{Text: cluster, Style: style})
			} else if w == 0 {
				// Zero-width (combining mark or similar); skip
				// (combining marks are already handled as part of clusters)
			}

			// Advance past the cluster
			i += utf8.RuneCountInString(cluster)
		} else {
			i++
		}
	}

	return cells
}

// parseSGRDecimal parses a non-negative decimal integer from s without heap allocation.
func parseSGRDecimal(s string) (int, bool) {
	if s == "" {
		return 0, true
	}
	n := 0
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < '0' || b > '9' {
			return 0, false
		}
		n = n*10 + int(b-'0')
		if n > 1000000 {
			n = 1000000
		}
	}
	return n, true
}

// applySGRSequence applies SGR parameters to a style using zero-allocation stack buffering.
func applySGRSequence(style Style, params string) Style {
	// Empty parameter list (bare ESC[m) counts as code 0 (reset)
	if params == "" {
		return Style{}
	}

	var buf [16]int
	codes := buf[:0]

	start := 0
	pLen := len(params)
	for i := 0; i <= pLen; i++ {
		if i == pLen || params[i] == ';' {
			sub := params[start:i]
			if v, ok := parseSGRDecimal(sub); ok {
				codes = append(codes, v)
			} else {
				codes = append(codes, -1) // marker for malformed numeric token
			}
			start = i + 1
		}
	}

	for i := 0; i < len(codes); i++ {
		code := codes[i]
		if code < 0 {
			continue
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

		case code == 3:
			// Italic (not supported in loom.Style, skip)

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

		case code == 38 && i+2 < len(codes) && codes[i+1] == 5:
			// 256-color foreground: 38;5;n
			v := codes[i+2]
			if v >= 0 && v <= 255 {
				style.FG = ColorIndex(uint8(v))
			}
			i += 2

		case code == 48 && i+2 < len(codes) && codes[i+1] == 5:
			// 256-color background: 48;5;n
			v := codes[i+2]
			if v >= 0 && v <= 255 {
				style.BG = ColorIndex(uint8(v))
			}
			i += 2

		case code == 38 && i+4 < len(codes) && codes[i+1] == 2:
			// 24-bit RGB foreground: 38;2;r;g;b
			r, g, b := codes[i+2], codes[i+3], codes[i+4]
			if r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
				style.FG = ColorRGB(uint8(r), uint8(g), uint8(b))
			}
			i += 4

		case code == 48 && i+4 < len(codes) && codes[i+1] == 2:
			// 24-bit RGB background: 48;2;r;g;b
			r, g, b := codes[i+2], codes[i+3], codes[i+4]
			if r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
				style.BG = ColorRGB(uint8(r), uint8(g), uint8(b))
			}
			i += 4
		}
	}

	return style
}
