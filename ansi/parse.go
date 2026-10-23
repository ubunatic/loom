// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansi

import "errors"

// ParseSGR applies semicolon-separated SGR parameters (without ESC[ and m) to style.
// Empty parameters reset the style; unsupported and invalid codes are ignored.
func ParseSGR(style Style, params string) Style {
	if params == "" {
		return Style{}
	}

	for pos := 0; pos <= len(params); {
		var parts [5]string
		var ends [5]int
		count, next := 0, pos
		for count < len(parts) && next <= len(params) {
			end := next
			for end < len(params) && params[end] != ';' {
				end++
			}
			parts[count], ends[count] = params[next:end], end+1
			count++
			next = end + 1
		}
		pos = ends[0]
		code, err := sgrInt(parts[0])
		if err != nil && parts[0] != "" {
			continue
		}

		switch {
		case code == 0:
			style = Style{}

		case code == 1:
			style.Bold = true

		case code == 2:
			style.Dim = true

		case code == 3:
			style.Italic = true

		case code == 4:
			style.Underline = true

		case code == 7:
			style.Invert = true

		case code == 9:
			style.Strike = true

		case code == 22:
			style.Bold = false
			style.Dim = false

		case code == 23:
			style.Italic = false

		case code == 24:
			style.Underline = false

		case code == 27:
			style.Invert = false

		case code == 29:
			style.Strike = false

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

		case code == 38 || code == 48:
			color, consumed, ok := parseExtendedColor(parts, count)
			pos = ends[consumed-1]
			if ok {
				if code == 38 {
					style.FG = color
				} else {
					style.BG = color
				}
			}

		}
	}

	return style
}

// parseExtendedColor consumes a complete indexed or RGB color, even when its
// values are invalid. Incomplete groups leave following parameters untouched.
func parseExtendedColor(parts [5]string, count int) (Color, int, bool) {
	if count >= 3 && parts[1] == "5" {
		v, err := sgrInt(parts[2])
		return ColorIndex(uint8(v)), 3, err == nil && v >= 0 && v <= 255
	}
	if count >= 5 && parts[1] == "2" {
		r, errR := sgrInt(parts[2])
		g, errG := sgrInt(parts[3])
		b, errB := sgrInt(parts[4])
		ok := errR == nil && errG == nil && errB == nil && r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255
		return ColorRGB(uint8(r), uint8(g), uint8(b)), 5, ok
	}
	return Color{}, 1, false
}

// sgrInt accepts the same signed decimal syntax as strconv.Atoi.
// The shared error avoids allocations for malformed input.
func sgrInt(s string) (int, error) {
	if s == "" {
		return 0, errParameter
	}
	negative, pos := false, 0
	if s[0] == '-' || s[0] == '+' {
		negative, pos = s[0] == '-', 1
	}
	if pos == len(s) {
		return 0, errParameter
	}
	limit := uint(^uint(0) >> 1)
	if negative {
		limit++
	}
	var n uint
	for ; pos < len(s); pos++ {
		c := s[pos]
		if c < '0' || c > '9' {
			return 0, errParameter
		}
		digit := uint(c - '0')
		if n > (limit-digit)/10 {
			return 0, errParameter
		}
		n = n*10 + digit
	}
	if negative {
		return -int(n), nil
	}
	return int(n), nil
}

var errParameter = errors.New("invalid SGR parameter")
