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

	"github.com/spf13/cobra"
	"ubunatic.com/loom"
	"ubunatic.com/loom/measure"
)

func main() {
	if err := execute(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execute(args []string, out io.Writer) error {
	root := &cobra.Command{
		Use: "loom", Version: loom.Version, Short: "Inspect terminal capabilities and validate TUI assets",
		SilenceUsage: true, SilenceErrors: true,
	}
	root.SetOut(out)
	root.AddCommand(infoCommand(), measureCommand(), evalCommand(), checkBoxCommand(), viewCommand(), editCommand(), formatCommand(), frameCommand(), widgetsCommand())
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
	var annotate bool
	cmd := &cobra.Command{
		Use: "eval <file>", Short: "Evaluate terminal dimensions and row geometry",
		Args: cobra.ExactArgs(1), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return evaluateFile(cmd.OutOrStdout(), args[0], annotate)
		},
	}
	cmd.Flags().BoolVarP(&annotate, "annotate", "a", false, "print plain text with diagnostic markers")
	return cmd
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

func evaluateFile(out io.Writer, path string, annotate bool) error {
	data, err := readAsset(path)
	if err != nil {
		return err
	}
	text := strings.ReplaceAll(strings.ReplaceAll(string(data), "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	if annotate {
		var diagnostics []annotation
		maxWidth := 0
		widths := make([]int, len(lines))
		for i, line := range lines {
			widths[i] = measure.StringWidth(line)
			if widths[i] > maxWidth {
				maxWidth = widths[i]
			}
		}
		for i, line := range lines {
			if widths[i] != maxWidth {
				diagnostics = append(diagnostics, annotation{row: i, message: fmt.Sprintf("line %d: ragged width (%d columns, expected %d)", i+1, widths[i], maxWidth)})
			}
			plain := stripANSI(line)
			if strings.TrimRight(plain, " \t") != plain {
				diagnostics = append(diagnostics, annotation{row: i, message: fmt.Sprintf("line %d: trailing whitespace", i+1)})
			}
		}
		return writeAnnotated(out, text, diagnostics)
	}
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
		rs := []rune(plain)
		for i := 0; i < len(rs); {
			r := rs[i]
			w := measure.RuneWidth(r)
			if r >= 0x1F1E6 && r <= 0x1F1FF && i+1 < len(rs) && rs[i+1] >= 0x1F1E6 && rs[i+1] <= 0x1F1FF {
				w = measure.StringWidth(string(rs[i : i+2]))
				i++
			}
			if r == '┌' || r == '╔' {
				boxCount++
			}
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
			i++
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
	var annotate bool
	cmd := &cobra.Command{
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
					if annotate && data != nil && strings.HasPrefix(err.Error(), "boxed line ") {
						var line, left, right, width, wantLeft, wantRight, wantWidth, previous int
						if _, scanErr := fmt.Sscanf(err.Error(), "boxed line %d: boundaries %d..%d width %d, want %d..%d width %d from line %d", &line, &left, &right, &width, &wantLeft, &wantRight, &wantWidth, &previous); scanErr == nil {
							detail := fmt.Sprintf("line %d: box width mismatch (%d columns, expected %d from line %d)", line, width, wantWidth, previous)
							fmt.Fprintf(cmd.OutOrStdout(), "%s:\n", path)
							annotated := writeAnnotated(cmd.OutOrStdout(), string(data), []annotation{{row: line - 1, message: detail}})
							if annotated != nil {
								return annotated
							}
						}
					}
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
	cmd.Flags().BoolVarP(&annotate, "annotate", "a", false, "print plain text with diagnostic markers")
	return cmd
}

type annotation struct {
	row     int
	message string
}

func writeAnnotated(out io.Writer, text string, diagnostics []annotation) error {
	buffer, err := loom.ParseAnsiBuffer(text, 1, 1)
	if err != nil {
		return err
	}
	grid := strings.Split(buffer.PlainText(), "\n")
	byRow := make(map[int][]int)
	for i, diagnostic := range diagnostics {
		byRow[diagnostic.row] = append(byRow[diagnostic.row], i+1)
	}
	for row, line := range grid {
		if _, err := fmt.Fprint(out, line); err != nil {
			return err
		}
		if indices := byRow[row]; len(indices) > 0 {
			padding := buffer.Cols() - measure.StringWidth(line)
			if padding < 0 {
				padding = 0
			}
			if _, err := fmt.Fprint(out, strings.Repeat(" ", padding), "  <-- "); err != nil {
				return err
			}
			for i, index := range indices {
				if i > 0 {
					if _, err := fmt.Fprint(out, ", "); err != nil {
						return err
					}
				}
				if _, err := fmt.Fprint(out, index); err != nil {
					return err
				}
			}
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
	}
	for i, diagnostic := range diagnostics {
		if _, err := fmt.Fprintf(out, "%d: %s\n", i+1, diagnostic.message); err != nil {
			return err
		}
	}
	return nil
}

func viewCommand() *cobra.Command {
	var plain bool
	cmd := &cobra.Command{
		Use: "view <file>", Short: "Interactively view an ANSI art file",
		Args: cobra.ExactArgs(1), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			if plain {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), buf.PlainText())
				return err
			}
			pane, err := loom.New(24)
			if err != nil {
				return err
			}
			defer pane.Close()
			configureViewPane(pane)
			return pane.Run(&ansiView{buffer: buf})
		},
	}
	cmd.Flags().BoolVarP(&plain, "plain", "p", false, "Print parsed ANSI art as plain text")
	return cmd
}

func configureViewPane(pane *loom.Pane) {
	pane.MaxCols = 0
	pane.Resizeable = true
}

func editCommand() *cobra.Command {
	return &cobra.Command{
		Use: "edit <file>", Short: "Interactively edit a rich text or ANSI file",
		Args: cobra.ExactArgs(1), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			edit, err := loom.NewRichTextEditFromFile(path)
			if err != nil {
				return err
			}
			pane, err := loom.New(24)
			if err != nil {
				return err
			}
			defer pane.Close()
			configureEditPane(pane)
			return pane.Run(&editView{edit: edit})
		},
	}
}

func configureEditPane(pane *loom.Pane) {
	pane.MaxCols = 0
	pane.Resizeable = true
	pane.DisableGlobalF10Quit = true
}

type editView struct {
	edit          *loom.RichTextEdit
	unsavedDialog *loom.Dialog
	shouldQuit    bool
	lastRect      loom.Rect
}

func (v *editView) hotkeyBar() *loom.HintBar {
	bar := v.edit.HotkeyBar()
	quitEntry := loom.HintEntry{
		Key:     "F10",
		Binding: "f10",
		Label:   "Quit",
		Action: func() loom.EventResult {
			return v.handleQuit()
		},
	}
	bar.Entries = append(bar.Entries, quitEntry)
	return bar
}

func (v *editView) handleQuit() loom.EventResult {
	if !v.edit.IsModified() {
		v.shouldQuit = true
		return loom.QuitResult()
	}
	name := "Untitled"
	if v.edit.FilePath != "" {
		name = filepath.Base(v.edit.FilePath)
	}
	dialog := loom.NewDialog("Save changes?", fmt.Sprintf("Save changes to %s before closing?", name), "Save", "Discard", "Cancel")
	dialog.OnSelect = func(button string) {
		switch button {
		case "Save":
			if err := v.edit.Save(); err == nil {
				v.shouldQuit = true
			}
		case "Discard":
			v.shouldQuit = true
		}
	}
	v.unsavedDialog = dialog
	return loom.Handled()
}

func (v *editView) Draw(canvas *loom.Canvas, rect loom.Rect) {
	v.lastRect = rect
	if rect.W <= 0 || rect.H <= 0 {
		return
	}
	editorRect := rect
	if rect.H > 1 {
		editorRect.H = rect.H - 1
		barRect := loom.Rect{X: rect.X, Y: rect.Y + rect.H - 1, W: rect.W, H: 1}
		v.hotkeyBar().Draw(canvas, barRect)
	}
	v.edit.Draw(canvas, editorRect)
	if v.unsavedDialog != nil && v.unsavedDialog.Open {
		v.unsavedDialog.Draw(canvas, rect)
	}
}

func (v *editView) ConsumeKey(key loom.KeyEvent) loom.EventResult {
	if v.unsavedDialog != nil && v.unsavedDialog.Open {
		_ = v.unsavedDialog.ConsumeKey(key)
		if v.shouldQuit {
			return loom.QuitResult()
		}
		if !v.unsavedDialog.Open {
			v.unsavedDialog = nil
		}
		return loom.Handled()
	}
	if key.Is("f10") {
		res := v.handleQuit()
		if v.shouldQuit {
			return loom.QuitResult()
		}
		return res
	}
	if res := v.hotkeyBar().ConsumeKey(key); res.Consumed {
		if v.shouldQuit {
			return loom.QuitResult()
		}
		return res
	}
	res := v.edit.ConsumeKey(key)
	if v.shouldQuit {
		return loom.QuitResult()
	}
	return res
}

func (v *editView) ConsumeMouse(mouse loom.MouseEvent) loom.EventResult {
	if v.unsavedDialog != nil && v.unsavedDialog.Open {
		_ = v.unsavedDialog.ConsumeMouse(mouse)
		if v.shouldQuit {
			return loom.QuitResult()
		}
		if !v.unsavedDialog.Open {
			v.unsavedDialog = nil
		}
		return loom.Handled()
	}
	if v.lastRect.H > 1 && mouse.Y == v.lastRect.H-1 {
		mouseLocal := mouse
		mouseLocal.Y = 0
		res := v.hotkeyBar().ConsumeMouse(mouseLocal)
		if v.shouldQuit {
			return loom.QuitResult()
		}
		return res
	}
	res := v.edit.ConsumeMouse(mouse)
	if v.shouldQuit {
		return loom.QuitResult()
	}
	return res
}

type ansiView struct {
	buffer       *loom.AnsiBuffer
	offsetX      int
	offsetY      int
	viewportCols int
	viewportRows int
}

func (v *ansiView) Draw(canvas *loom.Canvas, rect loom.Rect) {
	v.viewportCols = rect.W
	v.viewportRows = rect.H
	v.clampOffsets()
	for y := 0; y < rect.H; y++ {
		row := y + v.offsetY
		if row >= v.buffer.Rows() {
			break
		}
		for x := 0; x < rect.W; x++ {
			col := x + v.offsetX
			if col >= v.buffer.Cols() {
				break
			}
			canvas.Set(rect.X+x, rect.Y+y, v.buffer.Get(col, row).ToCell())
		}
	}
}

func (v *ansiView) clampOffsets() {
	cols, rows := v.viewportCols, v.viewportRows
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	maxX := v.buffer.Cols() - cols
	if maxX < 0 {
		maxX = 0
	}
	maxY := v.buffer.Rows() - rows
	if maxY < 0 {
		maxY = 0
	}
	if v.offsetX < 0 {
		v.offsetX = 0
	} else if v.offsetX > maxX {
		v.offsetX = maxX
	}
	if v.offsetY < 0 {
		v.offsetY = 0
	} else if v.offsetY > maxY {
		v.offsetY = maxY
	}
}

func (v *ansiView) ConsumeKey(event loom.KeyEvent) loom.EventResult {
	panStep := loom.SpeccedDefaults.Pane.ViewPanStep
	switch {
	case event.Is("q", "f10", "ctrl-c"):
		return loom.QuitResult()
	case event.Is("shift-left", "["):
		v.offsetX -= panStep
	case event.Is("shift-right", "]"):
		v.offsetX += panStep
	case event.Is("left", "h"):
		v.offsetX--
	case event.Is("right", "l"):
		v.offsetX++
	case event.Is("home"):
		v.offsetX = 0
	case event.Is("end"):
		v.offsetX = v.buffer.Cols()
	case event.Is("up", "k"):
		v.offsetY--
	case event.Is("down", "j"):
		v.offsetY++
	case event.Is("pgup"):
		v.offsetY -= panStep
	case event.Is("pgdown", "pgdn", "pagedown"):
		v.offsetY += panStep
	}
	v.clampOffsets()
	return loom.Ignored()
}

func (*ansiView) ConsumeMouse(loom.MouseEvent) loom.EventResult { return loom.Ignored() }
