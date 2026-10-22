// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

func fmtCommand() *cobra.Command {
	var write bool
	cmd := &cobra.Command{
		Use: "fmt <file.ansi>", Short: "Canonicalize ANSI encoding without changing the screen",
		Args: cobra.ExactArgs(1), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := readAsset(args[0])
			if err != nil {
				return err
			}
			formatted := []byte(normalizeANSIEncoding(string(data)))
			if write {
				return atomicWrite(args[0], formatted)
			}
			_, err = io.WriteString(cmd.OutOrStdout(), string(formatted))
			return err
		},
	}
	cmd.Flags().BoolVarP(&write, "write", "w", false, "replace the file atomically")
	return cmd
}

type fmtStyle struct {
	fgSet, bgSet bool
	reverse      bool
	underline    bool
	unknown      bool
}

func normalizeANSIEncoding(text string) string {
	var out strings.Builder
	style := fmtStyle{}
	tailStart := -1
	tailControls := ""
	lastReset := false
	commitTail := func() {
		out.WriteString(tailControls)
		tailStart = -1
		tailControls = ""
	}
	trimTail := func() {
		if tailStart < 0 {
			return
		}
		content := out.String()
		out.Reset()
		out.WriteString(content[:tailStart])
		out.WriteString(tailControls)
		tailStart = -1
		tailControls = ""
	}
	for i := 0; i < len(text); {
		if text[i] == '\x1b' {
			j := i + 1
			if j < len(text) && text[j] == '[' {
				j++
				for j < len(text) && (text[j] < '@' || text[j] > '~') {
					j++
				}
				if j < len(text) {
					j++
				}
				sequence := text[i:j]
				if strings.HasSuffix(sequence, "m") {
					params := strings.TrimSuffix(strings.TrimPrefix(sequence, "\x1b["), "m")
					canonical, values, ok := canonicalSGR(params)
					if ok {
						applyFmtSGR(&style, values)
						if len(values) == 1 && values[0] == 0 && lastReset {
							i = j
							continue
						}
						sequence = "\x1b[" + canonical + "m"
					} else {
						style.unknown = true
					}
					if tailStart >= 0 {
						tailControls += sequence
					} else {
						out.WriteString(sequence)
					}
					lastReset = len(values) == 1 && values[0] == 0
					i = j
					continue
				}
			} else {
				for j < len(text) && text[j] != '\a' && !(text[j] == '\x1b' && j+1 < len(text) && text[j+1] == '\\') {
					j++
				}
				if j < len(text) {
					if text[j] == '\x1b' {
						j += 2
					} else {
						j++
					}
				}
			}
			commitTail()
			out.WriteString(text[i:j])
			i = j
			lastReset = false
			continue
		}
		if text[i] == '\n' {
			trimTail()
			out.WriteByte('\n')
			i++
			lastReset = false
			continue
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == ' ' && fmtTrimSafe(style) {
			if tailStart < 0 {
				tailStart = out.Len()
			}
			out.WriteByte(' ')
			i += size
			lastReset = false
			continue
		}
		commitTail()
		out.WriteString(text[i : i+size])
		i += size
		lastReset = false
	}
	trimTail()
	return out.String()
}

func canonicalSGR(params string) (string, []int, bool) {
	values := []int{0}
	if params != "" {
		parts := strings.Split(params, ";")
		values = make([]int, len(parts))
		for i, part := range parts {
			if part == "" {
				values[i] = 0
				continue
			}
			n, err := strconv.Atoi(part)
			if err != nil || n < 0 {
				return params, nil, false
			}
			values[i] = n
		}
	}
	var tokens [][]int
	hasReset, fgCount, bgCount := false, 0, 0
	for i := 0; i < len(values); i++ {
		code := values[i]
		if code == 38 || code == 48 {
			if i+2 >= len(values) || (values[i+1] != 5 && values[i+1] != 2) {
				return params, nil, false
			}
			length := 2
			if values[i+1] == 2 {
				length = 4
			}
			if i+length >= len(values) {
				return params, nil, false
			}
			group := append([]int(nil), values[i:i+length+1]...)
			if group[1] == 5 && group[2] < 16 {
				base := 30
				if code == 48 {
					base = 40
				}
				if group[2] >= 8 {
					base += 60
				}
				group = []int{base + group[2]%8}
			}
			if code == 38 {
				fgCount++
			} else {
				bgCount++
			}
			tokens = append(tokens, group)
			i += length
			continue
		}
		if code == 0 || code == 22 || (code >= 23 && code <= 29) || code == 39 || code == 49 {
			hasReset = true
		}
		if code >= 30 && code <= 37 || code >= 90 && code <= 97 {
			fgCount++
		}
		if code >= 40 && code <= 47 || code >= 100 && code <= 107 {
			bgCount++
		}
		tokens = append(tokens, []int{code})
	}
	if !hasReset && fgCount <= 1 && bgCount <= 1 {
		for i := range tokens {
			for j := i + 1; j < len(tokens); j++ {
				if tokens[j][0] < tokens[i][0] {
					tokens[i], tokens[j] = tokens[j], tokens[i]
				}
			}
		}
	}
	flat := make([]int, 0, len(values))
	for _, token := range tokens {
		if len(flat) > 0 && len(token) == 1 && flat[len(flat)-1] == token[0] {
			continue
		}
		flat = append(flat, token...)
	}
	parts := make([]string, len(flat))
	for i, value := range flat {
		parts[i] = strconv.Itoa(value)
	}
	return strings.Join(parts, ";"), values, true
}

func applyFmtSGR(style *fmtStyle, values []int) {
	for i := 0; i < len(values); i++ {
		code := values[i]
		switch {
		case code == 0:
			*style = fmtStyle{}
		case code == 4 || code == 21:
			style.underline = true
		case code == 24:
			style.underline = false
		case code == 7:
			style.reverse = true
		case code == 27:
			style.reverse = false
		case code == 1 || code == 2 || code == 3 || code == 5 || code == 6 || code == 8 || code == 9 || code == 22 || code == 23 || code == 25 || code == 28 || code == 29 || code == 51 || code == 52 || code == 53 || code == 55:
			// Keep trailing spaces carrying text attributes. Their glyph is
			// blank, but the screen cell still has a different style.
			style.unknown = true
		case code >= 30 && code <= 37 || code >= 90 && code <= 97:
			style.fgSet = true
		case code == 39:
			style.fgSet = false
		case code >= 40 && code <= 47 || code >= 100 && code <= 107:
			style.bgSet = true
		case code == 49:
			style.bgSet = false
		case (code == 38 || code == 48) && i+2 < len(values) && values[i+1] == 5:
			if code == 38 {
				style.fgSet = true
			} else {
				style.bgSet = true
			}
			i += 2
		case (code == 38 || code == 48) && i+4 < len(values) && values[i+1] == 2:
			if code == 38 {
				style.fgSet = true
			} else {
				style.bgSet = true
			}
			i += 4
		default:
			style.unknown = true
		}
	}
}

func fmtTrimSafe(style fmtStyle) bool {
	return !style.fgSet && !style.bgSet && !style.reverse && !style.underline && !style.unknown
}
