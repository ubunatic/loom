// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command loom-probe measures how the current terminal advances its cursor
// after representative emoji sequences.
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

type sample struct {
	name string
	text string
}

type result struct {
	sample sample
	row    int
	col    int
	err    error
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "loom-probe:", err)
		os.Exit(1)
	}
}

func run(input *os.File, output io.Writer) error {
	inFD := int(input.Fd())
	outFD, ok := output.(*os.File)
	if !ok || !term.IsTerminal(inFD) || !term.IsTerminal(int(outFD.Fd())) {
		return fmt.Errorf("stdin and stdout must be attached to a real terminal")
	}
	oldState, err := term.MakeRaw(inFD)
	if err != nil {
		return fmt.Errorf("enable raw terminal input: %w", err)
	}
	raw := true
	defer func() {
		if raw {
			term.Restore(inFD, oldState) //nolint:errcheck
		}
	}()

	inAlternateScreen := false
	defer func() {
		if inAlternateScreen {
			io.WriteString(output, "\x1b[?1049l") //nolint:errcheck
		}
	}()
	if _, err := io.WriteString(output, "\x1b[?1049h\x1b[2J"); err != nil {
		return fmt.Errorf("enter alternate screen: %w", err)
	}
	inAlternateScreen = true

	samples := []sample{
		{name: "ZWJ family", text: "👨‍👩‍👧‍👦"},
		{name: "DE flag", text: "🇩🇪"},
		{name: "plain emoji", text: "😀"},
	}
	results := make([]result, 0, len(samples))
	for i, s := range samples {
		labelRow := i*2 + 1
		textRow := labelRow + 1
		if _, err := fmt.Fprintf(output, "\x1b[%d;1H%s\x1b[%d;1H%s\x1b[6n", labelRow, s.name, textRow, s.text); err != nil {
			return fmt.Errorf("write %s sample: %w", s.name, err)
		}
		row, col, err := readCursorReport(inFD, time.Second)
		if err == nil && row != textRow {
			err = fmt.Errorf("cursor reported row %d, want %d", row, textRow)
		}
		results = append(results, result{sample: s, row: row, col: col, err: err})
	}
	if _, err := io.WriteString(output, "\x1b[?1049l"); err != nil {
		return fmt.Errorf("leave alternate screen: %w", err)
	}
	inAlternateScreen = false
	if err := term.Restore(inFD, oldState); err != nil {
		return fmt.Errorf("restore terminal: %w", err)
	}
	raw = false
	writeResults(output, results)
	return nil
}

func writeResults(output io.Writer, results []result) {
	fmt.Fprintln(output, "Terminal cursor advances (start column 1):")
	for _, r := range results {
		if r.err != nil {
			fmt.Fprintf(output, "  %s: unavailable (%v)\n", r.sample.name, r.err)
			continue
		}
		fmt.Fprintf(output, "  %s: %d columns (cursor column %d)\n", r.sample.name, cursorAdvance(1, r.col), r.col)
	}
}

func parseCursorReport(reply string) (int, int, error) {
	if !strings.HasPrefix(reply, "\x1b[") || !strings.HasSuffix(reply, "R") {
		return 0, 0, fmt.Errorf("invalid DSR reply %q", reply)
	}
	parts := strings.Split(reply[2:len(reply)-1], ";")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid DSR reply %q", reply)
	}
	row, rowErr := strconv.Atoi(parts[0])
	col, colErr := strconv.Atoi(parts[1])
	if rowErr != nil || colErr != nil || row < 1 || col < 1 {
		return 0, 0, fmt.Errorf("invalid DSR reply %q", reply)
	}
	return row, col, nil
}

func cursorAdvance(startColumn, endColumn int) int {
	return endColumn - startColumn
}
