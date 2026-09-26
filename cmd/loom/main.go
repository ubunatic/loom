// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command loom validates, measures, and views TUI assets.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/measure"
	"github.com/spf13/cobra"
)

func main() {
	if err := execute(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execute(args []string, out io.Writer) error {
	root := &cobra.Command{
		Use: "loom", Short: "Validate, measure, and view TUI assets",
		SilenceUsage: true, SilenceErrors: true,
	}
	root.SetOut(out)
	root.AddCommand(measureCommand(), evalCommand(), checkBoxCommand(), viewCommand())
	root.SetArgs(args)
	return root.Execute()
}

func measureCommand() *cobra.Command {
	return &cobra.Command{
		Use: "measure <file>", Short: "Measure visible line widths in terminal cells",
		Args: cobra.ExactArgs(1), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return measureFile(cmd.OutOrStdout(), args[0])
		},
	}
}

func evalCommand() *cobra.Command {
	return &cobra.Command{
		Use: "eval <file>", Short: "Evaluate terminal dimensions and row geometry",
		Args: cobra.ExactArgs(1), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return evaluateFile(cmd.OutOrStdout(), args[0])
		},
	}
}

func measureFile(out io.Writer, path string) error {
	data, err := readAsset(path)
	if err != nil {
		return err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	widths := make([]int, len(lines))
	maxWidth := 0
	trailing := false
	for i, line := range lines {
		widths[i] = measure.StringWidth(line)
		if widths[i] > maxWidth {
			maxWidth = widths[i]
		}
		plain := stripANSI(line)
		if strings.TrimRight(plain, " \t") != plain {
			trailing = true
		}
	}
	for i, width := range widths {
		marker := ""
		if width != maxWidth {
			marker = " (ragged)"
		}
		fmt.Fprintf(out, "line %d: %d columns%s\n", i+1, width, marker)
	}
	fmt.Fprintf(out, "bounding box: %d columns x %d lines\n", maxWidth, len(lines))
	if trailing {
		fmt.Fprintln(out, "warning: trailing whitespace detected")
	}
	return nil
}

func evaluateFile(out io.Writer, path string) error {
	data, err := readAsset(path)
	if err != nil {
		return err
	}
	text := strings.ReplaceAll(strings.ReplaceAll(string(data), "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	minWidth, maxWidth := -1, 0
	raggedRows, boxCount := 0, 0
	minX, minY := -1, -1
	maxX, maxY := -1, -1
	for y, line := range lines {
		width := measure.StringWidth(line)
		if minWidth < 0 || width < minWidth {
			minWidth = width
		}
		if width > maxWidth {
			maxWidth = width
		}
		plain := stripANSI(line)
		x := 0
		for _, r := range plain {
			if r == '┌' || r == '╔' {
				boxCount++
			}
			w := measure.RuneWidth(r)
			if w > 0 && !unicode.IsSpace(r) {
				if minX < 0 || x < minX {
					minX = x
				}
				if maxX < 0 || x+w-1 > maxX {
					maxX = x + w - 1
				}
				if minY < 0 || y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
			x += w
		}
	}
	for _, line := range lines {
		if measure.StringWidth(line) != maxWidth {
			raggedRows++
		}
	}
	fmt.Fprintf(out, "boxes: %d\nrows: %d\ncolumns: %d\n", boxCount, len(lines), maxWidth)
	fmt.Fprintf(out, "non-blank bounds: %s\n", boundsString(minX, minY, maxX, maxY))
	fmt.Fprintf(out, "ragged rows: %d\nmin line width: %d\nmax line width: %d\n", raggedRows, minWidth, maxWidth)
	return nil
}

func boundsString(minX, minY, maxX, maxY int) string {
	if minX < 0 {
		return "empty"
	}
	// Coordinates use one-based terminal columns and rows for editor-friendly output.
	return fmt.Sprintf("x=%d..%d y=%d..%d (%d x %d)", minX+1, maxX+1, minY+1, maxY+1, maxX-minX+1, maxY-minY+1)
}

func readAsset(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return data, nil
	}
	if pathErr, ok := err.(*os.PathError); ok {
		err = pathErr.Err
	}
	return nil, fmt.Errorf("%s: %w", path, err)
}

func stripANSI(text string) string {
	runes := []rune(text)
	var plain strings.Builder
	for i := 0; i < len(runes); i++ {
		if runes[i] != '\x1b' {
			plain.WriteRune(runes[i])
			continue
		}
		if i+1 >= len(runes) {
			break
		}
		i++
		switch runes[i] {
		case '[':
			for i+1 < len(runes) && (runes[i+1] < '@' || runes[i+1] > '~') {
				i++
			}
			if i+1 < len(runes) {
				i++
			}
		case ']', 'P', '^', '_':
			for i+1 < len(runes) && runes[i+1] != '\a' {
				if runes[i+1] == '\x1b' && i+2 < len(runes) && runes[i+2] == '\\' {
					i += 2
					break
				}
				i++
			}
			if i+1 < len(runes) && runes[i+1] == '\a' {
				i++
			}
		}
	}
	return plain.String()
}

func checkBoxCommand() *cobra.Command {
	return &cobra.Command{
		Use: "check-box <file...>", Short: "Validate alignment of Unicode box borders",
		Args: cobra.MinimumNArgs(1), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var failures []string
			for _, path := range args {
				data, err := readAsset(path)
				if err == nil {
					err = loom.ValidateAnsiBox(string(data))
				}
				if err != nil {
					if !strings.HasPrefix(err.Error(), path+":") {
						err = fmt.Errorf("%s: %w", path, err)
					}
					failures = append(failures, err.Error())
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s: ok\n", path)
			}
			if len(failures) > 0 {
				return fmt.Errorf("box validation failed:\n%s", strings.Join(failures, "\n"))
			}
			return nil
		},
	}
}

func viewCommand() *cobra.Command {
	return &cobra.Command{
		Use: "view <file>", Short: "Interactively view an ANSI art file",
		Args: cobra.ExactArgs(1), SilenceUsage: true,
		RunE: func(_ *cobra.Command, args []string) error {
			path := args[0]
			if strings.ToLower(filepath.Ext(path)) != ".ansi" {
				return fmt.Errorf("view supports .ansi files: %s", path)
			}
			data, err := readAsset(path)
			if err != nil {
				return err
			}
			buf, err := loom.ParseAnsiBuffer(string(data), 1, 1)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			pane, err := loom.New(24)
			if err != nil {
				return err
			}
			defer pane.Close()
			return pane.Run(&ansiView{buffer: buf})
		},
	}
}

type ansiView struct {
	buffer *loom.AnsiBuffer
	offset int
}

func (v *ansiView) Draw(canvas *loom.Canvas, rect loom.Rect) {
	for y := 0; y < rect.H; y++ {
		row := y + v.offset
		if row >= v.buffer.Rows() {
			break
		}
		for x := 0; x < rect.W && x < v.buffer.Cols(); x++ {
			canvas.Set(rect.X+x, rect.Y+y, v.buffer.Get(x, row).ToCell())
		}
	}
}

func (v *ansiView) HandleKey(event loom.KeyEvent) bool {
	switch {
	case event.Is("q", "f10", "ctrl-c"):
		return true
	case event.Is("up", "k"):
		if v.offset > 0 {
			v.offset--
		}
	case event.Is("down", "j"):
		if v.offset < v.buffer.Rows()-1 {
			v.offset++
		}
	case event.Is("pgup"):
		v.offset -= 10
		if v.offset < 0 {
			v.offset = 0
		}
	case event.Is("pgdn"):
		v.offset += 10
		if v.offset >= v.buffer.Rows() {
			v.offset = v.buffer.Rows() - 1
		}
	}
	return false
}

func (*ansiView) HandleMouse(loom.MouseEvent) bool { return false }
