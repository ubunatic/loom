// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"ubunatic.com/loom"
)

type editFocus int

const (
	focusEditor editFocus = iota
	focusBrowser
	focusSearch
)

type searchMatch struct {
	line   int
	start  int
	length int
}

func editCommand() *cobra.Command {
	var configPath string
	var themeName string
	var mouseGrab bool
	var altScreen bool

	cmd := &cobra.Command{
		Use:   "edit <file>",
		Short: "Interactively edit a rich text or ANSI file",
		Args:  cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			cfg, err := loom.LoadEditorConfig(configPath)
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("theme") {
				if !loom.ThemeExists(themeName) {
					return fmt.Errorf("unknown theme %q", themeName)
				}
				cfg.Theme = themeName
			}
			if cmd.Flags().Changed("mousegrab") {
				cfg.MouseGrab = mouseGrab
			}
			if cmd.Flags().Changed("altscreen") {
				cfg.AltScreen = altScreen
			}

			edit, err := loom.NewRichTextEditFromFile(path)
			if err != nil {
				return err
			}

			pane, err := loom.New(24)
			if err != nil {
				return err
			}
			defer pane.Close()

			configureEditPane(pane, cfg)
			view, err := newEditView(edit, path, cfg)
			if err != nil {
				return err
			}
			return pane.Run(view)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to editor settings YAML file")
	cmd.Flags().StringVarP(&themeName, "theme", "t", "", "color theme name")
	cmd.Flags().BoolVarP(&mouseGrab, "mousegrab", "m", false, "enable mouse tracking/capture")
	cmd.Flags().BoolVarP(&altScreen, "altscreen", "a", true, "use alternate screen buffer")

	return cmd
}

func configureEditPane(pane *loom.Pane, cfg loom.EditorConfig) {
	pane.MaxCols = 0
	pane.Resizeable = true
	pane.DisableGlobalF10Quit = true
	if cfg.MouseGrab {
		pane.EnableMouse()
	}
	if cfg.AltScreen {
		pane.SetScreenMode(loom.ScreenAlt)
	} else {
		pane.SetScreenMode(loom.ScreenInline)
	}
}

type editView struct {
	edit            *loom.RichTextEdit
	filePath        string
	config          loom.EditorConfig
	focused         editFocus
	showSidePanel   bool
	filePicker      *loom.FilePicker
	showSearch      bool
	searchBar       *loom.SearchBar
	searchQuery     string
	regexMode       bool
	searchMatches   []searchMatch
	searchIndex     int
	searchErr       string
	unsavedDialog   *loom.Dialog
	pendingOpenPath string
	shouldQuit      bool
	lastRect        loom.Rect
	editorRect      loom.Rect
	browserRect     loom.Rect
	searchRect      loom.Rect
}

func newEditView(edit *loom.RichTextEdit, path string, cfg loom.EditorConfig) (*editView, error) {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		dir, _ = filepath.Abs(".")
	}
	picker, err := loom.NewFilePicker(dir, loom.FilePickerOptions{
		Mode: loom.FilePickerFiles,
	})
	if err != nil {
		return nil, err
	}

	sb := loom.NewSearchBar()
	sb.Prompt = "Find: "
	sb.Placeholder = "type query..."

	v := &editView{
		edit:          edit,
		filePath:      path,
		config:        cfg,
		focused:       focusEditor,
		showSidePanel: false, // initial toggle state
		filePicker:    picker,
		searchBar:     sb,
	}

	if theme := loom.Theme(cfg.Theme); cfg.Theme != "" {
		v.edit.ApplyTheme(theme)
		v.filePicker.ApplyTheme(theme)
	}

	return v, nil
}

func (v *editView) openFile(path string) error {
	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		return err
	}
	v.edit = edit
	v.filePath = path
	if theme := loom.Theme(v.config.Theme); v.config.Theme != "" {
		v.edit.ApplyTheme(theme)
	}
	v.updateSearchMatches()
	return nil
}

func (v *editView) hotkeyBar() *loom.HintBar {
	theme := loom.Theme(v.config.Theme)
	style := theme.HintBarStyle()

	entries := []loom.HintEntry{
		{Key: "F2", Binding: "f2", Label: "Files", Action: func() loom.EventResult {
			v.toggleSidePanel()
			return loom.Handled()
		}},
		{Key: "F3/^F", Binding: "f3", Label: "Search", Action: func() loom.EventResult {
			v.toggleSearch()
			return loom.Handled()
		}},
		{Key: "^S", Binding: "ctrl-s", Label: "Save", Action: func() loom.EventResult {
			_ = v.edit.Save()
			return loom.Handled()
		}},
		{Key: "^Shift+S", Binding: "ctrl-shift-s", Label: "Save as", Action: func() loom.EventResult {
			return v.edit.ConsumeKey(loom.KeyEvent{Key: "ctrl-shift-s"})
		}},
		{Key: "F5", Binding: "f5", Label: "Box", Action: func() loom.EventResult {
			return v.edit.ConsumeKey(loom.KeyEvent{Key: "f5"})
		}},
		{Key: "F10", Binding: "f10", Label: "Quit", Action: func() loom.EventResult {
			return v.handleQuit()
		}},
	}

	bar := &loom.HintBar{Entries: entries, Style: style}
	return bar
}

func (v *editView) handleQuit() loom.EventResult {
	if !v.edit.IsModified() {
		v.shouldQuit = true
		return loom.QuitResult()
	}
	name := "Untitled"
	if v.filePath != "" {
		name = filepath.Base(v.filePath)
	}
	dialog := loom.NewDialog("Save changes?", fmt.Sprintf("Save changes to %s before closing?", name), "Save", "Discard", "Cancel")
	dialog.OnSelect = func(button string) {
		switch button {
		case "Save":
			if err := v.edit.Save(); err == nil {
				if v.pendingOpenPath != "" {
					_ = v.openFile(v.pendingOpenPath)
					v.pendingOpenPath = ""
				} else {
					v.shouldQuit = true
				}
			}
		case "Discard":
			if v.pendingOpenPath != "" {
				_ = v.openFile(v.pendingOpenPath)
				v.pendingOpenPath = ""
			} else {
				v.shouldQuit = true
			}
		case "Cancel":
			v.pendingOpenPath = ""
		}
	}
	v.unsavedDialog = dialog
	return loom.Handled()
}

func (v *editView) toggleSidePanel() {
	v.showSidePanel = !v.showSidePanel
	if v.showSidePanel {
		v.focused = focusBrowser
	} else if v.focused == focusBrowser {
		v.focused = focusEditor
	}
}

func (v *editView) toggleSearch() {
	v.showSearch = !v.showSearch
	if v.showSearch {
		v.focused = focusSearch
		v.searchBar.SetFocused(true)
	} else {
		if v.focused == focusSearch {
			v.focused = focusEditor
		}
		v.edit.ClearSelection()
	}
}

func (v *editView) updateSearchMatches() {
	v.searchMatches = nil
	v.searchErr = ""
	if v.searchQuery == "" {
		v.edit.ClearSelection()
		return
	}

	lines := v.edit.Document.Lines
	if v.regexMode {
		re, err := regexp.Compile("(?i)" + v.searchQuery)
		if err != nil {
			v.searchErr = "Invalid regex"
			v.edit.ClearSelection()
			return
		}
		for lineIdx, line := range lines {
			plainText := richLinePlainText(line)
			locs := re.FindAllStringIndex(plainText, -1)
			for _, loc := range locs {
				startRune := len([]rune(plainText[:loc[0]]))
				matchRunes := len([]rune(plainText[loc[0]:loc[1]]))
				v.searchMatches = append(v.searchMatches, searchMatch{
					line:   lineIdx,
					start:  startRune,
					length: matchRunes,
				})
			}
		}
	} else {
		queryLower := strings.ToLower(v.searchQuery)
		for lineIdx, line := range lines {
			plainText := richLinePlainText(line)
			textLower := strings.ToLower(plainText)
			startIdx := 0
			for {
				idx := strings.Index(textLower[startIdx:], queryLower)
				if idx < 0 {
					break
				}
				matchByteStart := startIdx + idx
				matchByteEnd := matchByteStart + len(queryLower)
				startRune := len([]rune(plainText[:matchByteStart]))
				matchRunes := len([]rune(plainText[matchByteStart:matchByteEnd]))
				v.searchMatches = append(v.searchMatches, searchMatch{
					line:   lineIdx,
					start:  startRune,
					length: matchRunes,
				})
				startIdx = matchByteStart + max(1, len(queryLower))
			}
		}
	}

	if len(v.searchMatches) > 0 {
		if v.searchIndex >= len(v.searchMatches) {
			v.searchIndex = 0
		}
		v.highlightSearchMatch()
	} else {
		v.edit.ClearSelection()
	}
}

func (v *editView) highlightSearchMatch() {
	if len(v.searchMatches) == 0 {
		return
	}
	m := v.searchMatches[v.searchIndex]
	from := loom.RichPosition{Line: m.line, Offset: m.start}
	to := loom.RichPosition{Line: m.line, Offset: m.start + m.length}
	v.edit.SetSelection(from, to)
	v.edit.Cursor = to
}

func (v *editView) Draw(canvas *loom.Canvas, rect loom.Rect) {
	v.lastRect = rect
	if rect.W <= 0 || rect.H <= 0 {
		return
	}

	theme := loom.Theme(v.config.Theme)
	boxStyle := theme.BoxStyle()
	normalStyle := theme.ChoiceStyle().Normal
	dimStyle := loom.Style{Dim: true, FG: normalStyle.FG, BG: normalStyle.BG}

	// 1. Render Outer Box Frame with title "loom edit"
	canvas.PaintDefaultSurface(rect, normalStyle)
	boxBorder := "┌─ loom edit " + strings.Repeat("─", max(0, rect.W-14)) + "┐"
	if rect.H > 0 {
		canvas.WriteDefault(rect.X, rect.Y, loom.TruncateText(boxBorder, rect.W, ""), boxStyle.Border)
	}
	if rect.H > 1 {
		canvas.WriteDefault(rect.X, rect.Y+rect.H-1, "└"+strings.Repeat("─", max(0, rect.W-2))+"┘", boxStyle.Border)
	}
	for y := rect.Y + 1; y < rect.Y+rect.H-1; y++ {
		canvas.Set(rect.X, y, loom.Cell{Text: "│", Style: boxStyle.Border})
		canvas.Set(rect.X+rect.W-1, y, loom.Cell{Text: "│", Style: boxStyle.Border})
	}

	innerW := max(0, rect.W-2)
	innerX := rect.X + 1

	// 2. Line Y=1: Header filebar
	fileName := "Untitled"
	if v.filePath != "" {
		fileName = filepath.Base(v.filePath)
	}
	mouseStr := "off"
	if v.config.MouseGrab {
		mouseStr = "on"
	}
	rightStatus := fmt.Sprintf("theme: %-8s   mouse: %s", v.config.Theme, mouseStr)
	headerText := fmt.Sprintf(" File    %-48s %s", fileName, rightStatus)
	if rect.H > 2 {
		canvas.WriteDefault(innerX, rect.Y+1, loom.TruncateText(headerText, innerW, ""), normalStyle)
	}

	// Line Y=2: Divider
	if rect.H > 3 {
		canvas.WriteDefault(innerX, rect.Y+2, strings.Repeat("─", innerW), boxStyle.Border)
	}

	// 3. Line Y=H-3: Divider, Line Y=H-2: Status, Line Y=H-1: Hotkeys
	if rect.H >= 6 {
		// Y=H-3
		canvas.WriteDefault(innerX, rect.Y+rect.H-4, strings.Repeat("─", innerW), boxStyle.Border)

		// Y=H-2
		altStr := "off"
		if v.config.AltScreen {
			altStr = "on"
		}
		absPath := v.filePath
		if abs, err := filepath.Abs(v.filePath); err == nil {
			absPath = abs
		}
		curLine := v.edit.Cursor.Line + 1
		curCol := v.edit.Cursor.Offset + 1
		posStr := fmt.Sprintf("Ln %d, Col %d", curLine, curCol)
		statusRight := fmt.Sprintf("%s  |  altscreen: %s", posStr, altStr)
		rightW := len(statusRight)
		leftW := max(0, innerW-rightW-2)
		statusLine := fmt.Sprintf(" %-*s %s", leftW, loom.TruncateText(absPath, leftW, "..."), statusRight)
		canvas.WriteDefault(innerX, rect.Y+rect.H-3, loom.TruncateText(statusLine, innerW, ""), normalStyle)

		// Y=H-1
		v.hotkeyBar().Draw(canvas, loom.Rect{X: innerX, Y: rect.Y + rect.H - 2, W: innerW, H: 1})
	}

	// 4. Middle Content Area: Y = rect.Y + 3 to rect.Y + rect.H - 5
	contentY := rect.Y + 3
	contentH := max(0, rect.H-6)

	if contentH > 0 && innerW > 0 {
		if v.showSidePanel {
			browserW := min(24, innerW/3)
			if browserW < 10 {
				browserW = innerW / 2
			}
			v.browserRect = loom.Rect{X: innerX, Y: contentY, W: browserW, H: contentH}

			// Render Browser Panel
			v.filePicker.Draw(canvas, v.browserRect)

			// Vertical separator line
			sepX := innerX + browserW
			if sepX < innerX+innerW {
				for y := contentY; y < contentY+contentH; y++ {
					canvas.Set(sepX, y, loom.Cell{Text: "│", Style: boxStyle.Border})
				}
			}

			editorX := sepX + 1
			editorW := max(0, innerX+innerW-editorX)
			v.editorRect = loom.Rect{X: editorX, Y: contentY, W: editorW, H: contentH}
		} else {
			v.browserRect = loom.Rect{}
			v.editorRect = loom.Rect{X: innerX, Y: contentY, W: innerW, H: contentH}
		}

		v.edit.Draw(canvas, v.editorRect)

		// Search Panel Overlay
		if v.showSearch {
			sWidth := min(48, max(24, v.editorRect.W-2))
			sHeight := min(7, contentH)
			sX := v.editorRect.X + v.editorRect.W - sWidth
			sY := v.editorRect.Y
			v.searchRect = loom.Rect{X: sX, Y: sY, W: sWidth, H: sHeight}

			if sWidth > 4 && sHeight > 2 {
				// Search box background and border
				canvas.PaintDefaultSurface(v.searchRect, normalStyle)
				canvas.WriteDefault(sX, sY, "┌─ Search  F3 / ^F "+strings.Repeat("─", max(0, sWidth-20))+"┐", boxStyle.Border)
				canvas.WriteDefault(sX, sY+sHeight-1, "└"+strings.Repeat("─", max(0, sWidth-2))+"┘", boxStyle.Border)
				for y := sY + 1; y < sY+sHeight-1; y++ {
					canvas.Set(sX, y, loom.Cell{Text: "│", Style: boxStyle.Border})
					canvas.Set(sX+sWidth-1, y, loom.Cell{Text: "│", Style: boxStyle.Border})
				}

				// Find input line
				v.searchBar.MaxWidth = sWidth - 4
				v.searchBar.Draw(canvas, loom.Rect{X: sX + 2, Y: sY + 1, W: sWidth - 4, H: 1})

				// Mode toggle line: [ Normal * ]  [ Regex ]
				normStr := "[ Normal * ]"
				regStr := "[ Regex ]"
				if v.regexMode {
					normStr = "[ Normal ]"
					regStr = "[ Regex * ]"
				}
				if sHeight > 3 {
					canvas.WriteDefault(sX+2, sY+2, normStr+"  "+regStr, normalStyle)
				}

				// Match counter line
				if sHeight > 4 {
					matchStr := "No matches"
					if v.searchErr != "" {
						matchStr = v.searchErr
					} else if len(v.searchMatches) > 0 {
						matchStr = fmt.Sprintf("%d / %d matches", v.searchIndex+1, len(v.searchMatches))
					}
					canvas.WriteDefault(sX+2, sY+3, matchStr, dimStyle)
				}

				// Nav hints
				if sHeight > 5 {
					canvas.WriteDefault(sX+2, sY+4, "Enter: next   S-Enter: prev", dimStyle)
				}
			}
		}
	}

	// Unsaved Changes Dialog Overlay
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

	if key.Is("f2") {
		v.toggleSidePanel()
		return loom.Handled()
	}

	if key.Is("f3", "ctrl-f") {
		v.toggleSearch()
		return loom.Handled()
	}

	// Tab focus switching
	if key.Is("tab") && !v.showSearch && (v.showSidePanel || v.focused == focusBrowser) {
		if v.focused == focusBrowser {
			v.focused = focusEditor
		} else {
			v.focused = focusBrowser
		}
		return loom.Handled()
	}

	if v.showSearch && v.focused == focusSearch {
		if key.Is("esc") {
			v.toggleSearch()
			return loom.Handled()
		}
		if key.Is("enter") {
			if len(v.searchMatches) > 0 {
				v.searchIndex = (v.searchIndex + 1) % len(v.searchMatches)
				v.highlightSearchMatch()
			}
			return loom.Handled()
		}
		if key.Is("shift-enter") {
			if len(v.searchMatches) > 0 {
				v.searchIndex = (v.searchIndex - 1 + len(v.searchMatches)) % len(v.searchMatches)
				v.highlightSearchMatch()
			}
			return loom.Handled()
		}
		if key.Is("tab", "ctrl-r") {
			v.regexMode = !v.regexMode
			v.searchIndex = 0
			v.updateSearchMatches()
			return loom.Handled()
		}
		// Search mode toggle or input
		res := v.searchBar.ConsumeKey(key)
		if v.searchBar.Query != v.searchQuery {
			v.searchQuery = v.searchBar.Query
			v.searchIndex = 0
			v.updateSearchMatches()
		}
		return res
	}

	if v.showSidePanel && v.focused == focusBrowser {
		if key.Is("enter") {
			if entry, ok := v.filePicker.Selected(); ok && entry.Kind == loom.FileKindRegular {
				if v.edit.IsModified() {
					v.pendingOpenPath = entry.Path
					_ = v.handleQuit()
				} else {
					_ = v.openFile(entry.Path)
				}
				return loom.Handled()
			}
		}
		return v.filePicker.ConsumeKey(key)
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

func richLinePlainText(line loom.RichLine) string {
	var sb strings.Builder
	for _, span := range line.Spans {
		sb.WriteString(span.Text)
	}
	return sb.String()
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

	if v.lastRect.H > 2 && mouse.Y == v.lastRect.H-2 {
		mouseLocal := mouse
		mouseLocal.Y = 0
		res := v.hotkeyBar().ConsumeMouse(mouseLocal)
		if v.shouldQuit {
			return loom.QuitResult()
		}
		return res
	}

	// Click in side panel
	if v.showSidePanel && v.browserRect.Contains(mouse.X, mouse.Y) {
		v.focused = focusBrowser
		mouseLocal := mouse
		mouseLocal.X -= v.browserRect.X
		mouseLocal.Y -= v.browserRect.Y
		return v.filePicker.ConsumeMouse(mouseLocal)
	}

	// Click in editor area
	if v.editorRect.Contains(mouse.X, mouse.Y) {
		v.focused = focusEditor
		mouseLocal := mouse
		mouseLocal.X -= v.editorRect.X
		mouseLocal.Y -= v.editorRect.Y
		return v.edit.ConsumeMouse(mouseLocal)
	}

	return loom.Ignored()
}
