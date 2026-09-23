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

// ParseANSINew is the optimized zero-allocation ANSI parser that performs in-place
// rune cluster scanning over []rune slices and pre-allocates cell slice capacity.
func ParseANSINew(s string) []Cell {
	if s == "" {
		return nil
	}
	cells := make([]Cell, 0, len(s))
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
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			i++
			continue
		}

		w := measure.RuneWidth(r)
		if w == 0 {
			// Standalone combining mark without base
			i++
			continue
		}

		// Collect base rune plus any subsequent zero-width combining marks into a cluster
		j := i + 1
		for j < n && (unicode.Is(unicode.Mn, rs[j]) || unicode.Is(unicode.Me, rs[j])) {
			j++
		}

		cluster := string(rs[i:j])
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

// applySGRSequence applies SGR parameters to a style.
func applySGRSequence(style Style, params string) Style {
	// Empty parameter list (bare ESC[m) counts as code 0 (reset)
	if params == "" {
		return Style{}
	}

	parts := strings.Split(params, ";")

	for i := 0; i < len(parts); i++ {
		// Empty parts (e.g., ESC[;1m) are treated as 0
		if parts[i] == "" {
			parts[i] = "0"
		}

		code, err := strconv.Atoi(parts[i])
		if err != nil {
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
