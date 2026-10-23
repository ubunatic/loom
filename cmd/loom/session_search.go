// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"ubunatic.com/loom"
	"ubunatic.com/loom/vt"
)

type sessionMatch struct {
	row, start, end int
}

func findSessionMatches(rows [][]vt.Cell, query string) []sessionMatch {
	if query == "" {
		return nil
	}
	caseSensitive := false
	for _, r := range query {
		if unicode.IsUpper(r) {
			caseSensitive = true
			break
		}
	}
	want := []rune(query)
	var matches []sessionMatch
	for y, row := range rows {
		runes, columns := sessionRowRunes(row)
		for start := 0; start+len(want) <= len(runes); start++ {
			candidate := string(runes[start : start+len(want)])
			equal := candidate == query
			if !caseSensitive {
				equal = strings.EqualFold(candidate, query)
			}
			if equal {
				end := columns[start+len(want)-1] + 1
				lastCell := columns[start+len(want)-1]
				if lastCell+1 < len(row) && row[lastCell+1].Rune == 0 && row[lastCell+1].Text == "" {
					end++
				}
				matches = append(matches, sessionMatch{row: y, start: columns[start], end: end})
				start += len(want) - 1
			}
		}
	}
	return matches
}

func sessionRowRunes(row []vt.Cell) ([]rune, []int) {
	var runes []rune
	var columns []int
	for x, cell := range row {
		text := cell.Text
		if text == "" && cell.Rune != 0 {
			text = string(cell.Rune)
		}
		for _, r := range text {
			runes = append(runes, r)
			columns = append(columns, x)
		}
	}
	return runes, columns
}

func sessionSearchControls(defs loom.SessionDefaults, query string, matchIndex, matches int) string {
	search := defs.Search
	status := ""
	if matches > 0 && matchIndex >= 0 {
		status = fmt.Sprintf("%d/%d", matchIndex+1, matches)
	} else if query != "" {
		status = search.NoMatch
	}
	keyCaps := func(bindings []string) string {
		caps := make([]string, len(bindings))
		for i, binding := range bindings {
			caps[i] = loom.KeyCap(binding)
		}
		return strings.Join(caps, "/")
	}
	controls := []string{keyCaps(search.Bindings.Next) + " " + search.NextHint, keyCaps(search.Bindings.Previous) + " " + search.PreviousHint, keyCaps(search.Bindings.Close) + " " + search.CloseHint}
	if status != "" {
		controls = append([]string{status}, controls...)
	}
	return strings.Join(controls, "  ")
}

func drawSessionSearchBar(out io.Writer, cols, row int, searchBar *loom.SearchBar, controls string) {
	if cols <= 0 {
		return
	}
	searchBar.Controls = controls
	canvas := loom.NewCanvas(cols, 1)
	searchBar.Draw(canvas, loom.Rect{W: cols, H: 1})
	var rendered strings.Builder
	canvas.Flush(&rendered, row)
	_, _ = io.WriteString(out, rendered.String())
}

func drawSessionSearchView(out io.Writer, rows [][]vt.Cell, screenRows, cols int, selected *sessionMatch) {
	if screenRows <= 0 || cols <= 0 {
		return
	}
	start := max(0, len(rows)-screenRows)
	if selected != nil {
		start = max(0, selected.row-screenRows/2)
		start = min(start, max(0, len(rows)-screenRows))
	}
	var output strings.Builder
	for y := 0; y < screenRows; y++ {
		fmt.Fprintf(&output, "\x1b[%d;1H\x1b[2K", y+1)
		index := start + y
		if index >= len(rows) {
			continue
		}
		row := rows[index]
		activeStyle := vt.Style{}
		for x := 0; x < min(cols, len(row)); x++ {
			cell := row[x]
			style := cell.Style
			if selected != nil && selected.row == index && x >= selected.start && x < selected.end {
				style.Reverse = true
			}
			if style != activeStyle {
				output.WriteString(style.SGR())
				activeStyle = style
			}
			switch {
			case cell.Text != "":
				output.WriteString(cell.Text)
			case cell.Rune != 0:
				output.WriteRune(cell.Rune)
			default:
				output.WriteByte(' ')
			}
		}
		output.WriteString("\x1b[0m")
	}
	_, _ = out.Write([]byte(output.String()))
}
