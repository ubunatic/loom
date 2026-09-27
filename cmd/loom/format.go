package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"codeberg.org/ubunatic/loom/measure"
	"github.com/spf13/cobra"
)

func formatCommand() *cobra.Command {
	var width int
	var trim, align, write bool
	cmd := &cobra.Command{Use: "format <file>", Short: "Normalize ANSI asset row widths", Args: cobra.ExactArgs(1), SilenceUsage: true}
	cmd.Flags().IntVar(&width, "width", 0, "target terminal-cell width for every row")
	cmd.Flags().BoolVarP(&write, "write", "w", false, "replace the file atomically")
	cmd.Flags().BoolVar(&trim, "trim-trailing", false, "remove trailing spaces and tabs")
	cmd.Flags().BoolVar(&align, "align-box", false, "align right edges of complete vertical box rows")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if width < 0 || (cmd.Flags().Changed("width") && width == 0) {
			return fmt.Errorf("width must be positive")
		}
		data, err := readAsset(args[0])
		if err != nil {
			return err
		}
		formatted, err := formatText(string(data), width, trim, align)
		if err != nil {
			return err
		}
		if write {
			return atomicWrite(args[0], []byte(formatted))
		}
		_, err = io.WriteString(cmd.OutOrStdout(), formatted)
		return err
	}
	return cmd
}

func formatText(text string, width int, trim, align bool) (string, error) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	finalNL := strings.HasSuffix(text, "\n")
	lines := strings.Split(text, "\n")
	if finalNL {
		lines = lines[:len(lines)-1]
	}
	if trim {
		for i := range lines {
			lines[i] = trimLine(lines[i])
		}
	}
	if align {
		max := 0
		for _, line := range lines {
			if left, right, ok := boxEdges(line); ok && right-left+1 > max {
				max = right - left + 1
			}
		}
		for i, line := range lines {
			if left, right, ok := boxEdges(line); ok {
				lines[i] = padBeforeLastEdge(line, max-(right-left+1))
			}
		}
	}
	if width > 0 {
		for i, line := range lines {
			lines[i] = fitANSIWidth(line, width)
		}
	}
	out := strings.Join(lines, "\n")
	if finalNL {
		out += "\n"
	}
	return out, nil
}

func trimLine(s string) string {
	// Keep trailing SGR sequences while removing whitespace immediately before them.
	end := len(s)
	start := end
	for start > 0 && s[start-1] == 'm' {
		esc := strings.LastIndex(s[:start], "\x1b[")
		if esc < 0 {
			break
		}
		start = esc
	}
	suffix := s[start:end]
	body := strings.TrimRight(s[:start], " \t")
	return body + suffix
}

func fitANSIWidth(s string, target int) string {
	var out strings.Builder
	width, active := 0, false
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] < '@' || s[j] > '~') {
				j++
			}
			if j < len(s) {
				j++
			}
			seq := s[i:j]
			out.WriteString(seq)
			if strings.HasSuffix(seq, "m") {
				if strings.Contains(seq, "[0m") || strings.Contains(seq, "[m") {
					active = false
				} else {
					active = true
				}
			}
			i = j
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if r >= 0x1F1E6 && r <= 0x1F1FF && i+size < len(s) {
			r2, size2 := utf8.DecodeRuneInString(s[i+size:])
			if r2 >= 0x1F1E6 && r2 <= 0x1F1FF {
				pair := s[i : i+size+size2]
				rw := measure.StringWidth(pair)
				if width+rw > target {
					break
				}
				out.WriteString(pair)
				width += rw
				i += size + size2
				continue
			}
		}
		rw := measure.RuneWidth(r)
		if width+rw > target {
			break
		}
		out.WriteRune(r)
		width += rw
		i += size
	}
	if active {
		out.WriteString("\x1b[0m")
	}
	if width < target {
		out.WriteString(strings.Repeat(" ", target-width))
	}
	return out.String()
}

func boxEdges(s string) (int, int, bool) {
	plain := []rune(stripANSI(s))
	left, right, col := -1, -1, 0
	for i := 0; i < len(plain); {
		r := plain[i]
		if r >= 0x1F1E6 && r <= 0x1F1FF && i+1 < len(plain) && plain[i+1] >= 0x1F1E6 && plain[i+1] <= 0x1F1FF {
			col += measure.StringWidth(string(plain[i : i+2]))
			i += 2
			continue
		}
		if isBoxVerticalRune(r) {
			if left < 0 {
				left = col
			}
			right = col
		}
		col += measure.RuneWidth(r)
		i++
	}
	return left, right, left >= 0 && right > left
}

func isBoxVerticalRune(r rune) bool { return r == '│' || r == '║' || r == '┃' || r == '|' }

func padBeforeLastEdge(s string, n int) string {
	if n <= 0 {
		return s
	}
	runes := []rune(s)
	last := -1
	for i, r := range runes {
		if isBoxVerticalRune(r) {
			last = i
		}
	}
	if last < 0 {
		return s
	}
	return string(runes[:last]) + strings.Repeat(" ", n) + string(runes[last:])
}

func atomicWrite(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".loom-format-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err = tmp.Chmod(info.Mode().Perm()); err == nil {
		_, err = tmp.Write(data)
	}
	if err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
