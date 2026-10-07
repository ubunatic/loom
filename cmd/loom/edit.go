// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

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
		Use:          "edit <file>",
		Short:        "Interactively edit a rich text or ANSI file",
		Args:         cobra.ExactArgs(1),
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
			view.attach(pane)
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
	edit          *loom.RichTextEdit
	filePath      string
	config        loom.EditorConfig
	focused       editFocus
	showSidePanel bool
	filePicker    *loom.FilePicker
	showSearch    bool
	searchBar     *loom.SearchBar
	searchQuery   string
	regexMode     bool
	searchMatches []searchMatch
	searchIndex   int
	searchErr     string
	unsavedDialog *loom.Dialog
	pendingAction func()
	quit          func()
	lastRect      loom.Rect
	editorRect    loom.Rect
	browserRect   loom.Rect
	searchRect    loom.Rect
	searchModes   *loom.HintBar
	statusMessage string
}

func newEditView(edit *loom.RichTextEdit, path string, cfg loom.EditorConfig) (*editView, error) {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		dir, _ = filepath.Abs(".")
	}

	v := &editView{
		edit:          edit,
		filePath:      path,
		config:        cfg,
		focused:       focusEditor,
		showSidePanel: false, // initial toggle state
	}

	v.bindEdit()

	picker, err := loom.NewFilePicker(dir, loom.FilePickerOptions{
		Mode: loom.FilePickerFiles,
		OnSelect: func(selectedPath string) {
			if v.filePicker != nil {
				v.filePicker.Reset()
			}
			v.guardUnsaved(func() { v.reportError(v.openFile(selectedPath)) })
		},
		OnCancel: func() {
			if v.filePicker != nil {
				v.filePicker.Reset()
			}
			v.showSidePanel = false
			v.focused = focusEditor
		},
	})
	if err != nil {
		return nil, err
	}
	v.filePicker = picker

	sb := loom.NewSearchBar()
	sb.Prompt = loom.SpeccedDefaults.Editor.SearchPrompt
	sb.Placeholder = loom.SpeccedDefaults.Editor.SearchPlaceholder
	v.searchBar = sb

	if theme := loom.Theme(cfg.Theme); cfg.Theme != "" {
		v.edit.ApplyTheme(theme)
		v.filePicker.ApplyTheme(theme)
	}

	return v, nil
}

// bindEdit subscribes the host status line to the editor's document state,
// which is the single source of the Saved/Modified indicator.
func (v *editView) bindEdit() {
	v.edit.OnStateChange = func(state loom.DocState) {
		switch state {
		case loom.DocStateSaved:
			v.statusMessage = "Saved: " + v.edit.FilePath
			v.filePath = v.edit.FilePath
		case loom.DocStateModified:
			v.statusMessage = ""
		}
	}
	v.edit.OnOpenRequest = v.showFileBrowser
	v.edit.OnCloseRequest = func() { v.guardUnsaved(v.closeDocument) }
}

// showFileBrowser opens the side panel and focuses it; unlike F2 it never hides it.
func (v *editView) showFileBrowser() {
	v.showSidePanel = true
	v.focused = focusBrowser
}

// closeDocument resets to an empty Untitled document. Callers guard unsaved changes first.
func (v *editView) closeDocument() {
	v.edit.Close()
	v.filePath = ""
	v.statusMessage = ""
	v.updateSearchMatches()
}

// guardUnsaved runs action now when the buffer is clean, otherwise after the
// user chose Save or Discard in the unsaved-changes dialog.
func (v *editView) guardUnsaved(action func()) {
	if !v.edit.IsModified() {
		action()
		return
	}
	v.pendingAction = action
	_ = v.showUnsavedDialog()
}

func (v *editView) openFile(path string) error {
	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		return err
	}
	v.edit = edit
	v.filePath = path
	v.bindEdit()
	if theme := loom.Theme(v.config.Theme); v.config.Theme != "" {
		v.edit.ApplyTheme(theme)
	}
	v.updateSearchMatches()
	return nil
}

func (v *editView) hotkeyBar() *loom.HintBar {
	theme := loom.Theme(v.config.Theme)
	style := theme.HintBarStyle()
	editorDefs := loom.SpeccedDefaults.Editor
	richDefs := loom.SpeccedDefaults.RichTextEdit

	entries := []loom.HintEntry{
		{Binding: editorDefs.HotkeyFilesBinding, Label: editorDefs.HotkeyFilesLabel, Action: func() loom.EventResult {
			v.toggleSidePanel()
			return loom.Handled()
		}},
		{Key: editorDefs.HotkeySearchKey, Binding: editorDefs.HotkeySearchBinding, Label: editorDefs.HotkeySearchLabel, Action: func() loom.EventResult {
			v.toggleSearch()
			return loom.Handled()
		}},
		{Binding: richDefs.HotkeySaveBinding, Label: richDefs.HotkeySaveLabel, Action: func() loom.EventResult {
			v.reportError(v.edit.Save())
			return loom.Handled()
		}},
		{Binding: richDefs.HotkeySaveAsBinding, Label: richDefs.HotkeySaveAsLabel, Action: func() loom.EventResult {
			return v.edit.ConsumeKey(loom.KeyEvent{Key: richDefs.HotkeySaveAsBinding})
		}},
		{Binding: editorDefs.HotkeyBoxBinding, Label: editorDefs.HotkeyBoxLabel, Action: func() loom.EventResult {
			return v.edit.ConsumeKey(loom.KeyEvent{Key: editorDefs.HotkeyBoxBinding})
		}},
		{Binding: editorDefs.HotkeyScreenshotBinding, Label: editorDefs.HotkeyScreenshotLabel, DropPriority: 2, Action: func() loom.EventResult {
			v.reportError(v.captureScreenshot())
			return loom.Handled()
		}},
		// No Binding: the Pane owns the quit keys; this entry only serves the mouse.
		{Key: loom.KeyCap(richDefs.HotkeyQuitBinding), Label: richDefs.HotkeyQuitLabel, Action: func() loom.EventResult {
			if v.closeRequest(loom.CloseReasonQuitKey) == loom.CloseAllow {
				v.doQuit()
			}
			return loom.Handled()
		}},
	}

	bar := &loom.HintBar{Entries: entries, Style: style}
	return bar
}

// attach hands close control to the pane: quit keys and signals ask the
// view first, so unsaved changes always get the Save/Discard/Cancel dialog.
func (v *editView) attach(pane *loom.Pane) {
	v.quit = pane.Quit
	pane.OnCloseRequest = v.closeRequest
}

func (v *editView) doQuit() {
	if v.quit != nil {
		v.quit()
	}
}

// closeRequest allows closing a clean buffer and vetoes it behind the unsaved
// changes dialog otherwise.
func (v *editView) closeRequest(loom.CloseReason) loom.CloseDecision {
	if !v.edit.IsModified() {
		return loom.CloseAllow
	}
	v.guardUnsaved(v.doQuit)
	return loom.CloseVeto
}

func (v *editView) showUnsavedDialog() loom.EventResult {
	name := "Untitled"
	if v.filePath != "" {
		name = filepath.Base(v.filePath)
	}
	dialog := loom.NewDialog("Save changes?", fmt.Sprintf("Save changes to %s before closing?", name), "Save", "Discard", "Cancel")
	dialog.OnSelect = func(button string) {
		action := v.pendingAction
		v.pendingAction = nil
		switch button {
		case "Save":
			if err := v.edit.Save(); err != nil {
				v.reportError(err)
			} else if action != nil && !v.edit.IsModified() {
				// An untitled buffer opens the save picker and stays modified; keep it.
				action()
			}
		case "Discard":
			if action != nil {
				action()
			}
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
		v.edit.Highlights = nil
	}
}

func (v *editView) updateSearchMatches() {
	v.searchMatches = nil
	v.searchErr = ""
	if v.searchQuery == "" {
		v.edit.Highlights = nil
		return
	}

	lines := v.edit.Document.Lines
	if v.regexMode {
		re, err := regexp.Compile("(?i)" + v.searchQuery)
		if err != nil {
			v.searchErr = "Invalid regex"
			v.edit.Highlights = nil
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
		queryRunes := len([]rune(v.searchQuery))
		for lineIdx, line := range lines {
			plainText := richLinePlainText(line)
			for _, startRune := range findRuneMatches(plainText, v.searchQuery) {
				v.searchMatches = append(v.searchMatches, searchMatch{
					line:   lineIdx,
					start:  startRune,
					length: queryRunes,
				})
			}
		}
	}

	if len(v.searchMatches) > 0 {
		if v.searchIndex >= len(v.searchMatches) {
			v.searchIndex = 0
		}
		v.highlightSearchMatch()
	} else {
		v.edit.Highlights = nil
	}
}

func (v *editView) highlightSearchMatch() {
	if len(v.searchMatches) == 0 {
		return
	}
	m := v.searchMatches[v.searchIndex]
	from := loom.RichPosition{Line: m.line, Offset: m.start}
	to := loom.RichPosition{Line: m.line, Offset: m.start + m.length}
	v.edit.Highlights = []loom.RichTextHighlight{{From: from, To: to, Style: v.edit.SelectionStyle}}
	if v.focused == focusSearch {
		v.edit.Cursor = to
	}
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

	outer := loom.Box{Title: "loom edit", Style: boxStyle, Border: loom.BoxBorder{TitlePrefix: "─ ", TitleSuffix: " "}}
	outer.Draw(canvas, rect)
	v.editorRect, v.browserRect, v.searchRect = loom.Rect{}, loom.Rect{}, loom.Rect{}
	v.edit.SetFocus(v.focused == focusEditor)
	v.searchBar.SetFocused(v.showSearch && v.focused == focusSearch)

	innerW := max(0, rect.W-2)
	innerX := rect.X + 1

	// 2. Line Y=1: App-level header (theme, mouse). The document name, state
	// and File menu belong to the RichTextEdit file bar.
	mouseStr := "off"
	if v.config.MouseGrab {
		mouseStr = "on"
	}
	if rect.H > 2 {
		headerText := fmt.Sprintf(" theme: %-8s   mouse: %s", v.config.Theme, mouseStr)
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
		if v.statusMessage != "" {
			absPath = v.statusMessage
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
			sWidth := min(48, v.editorRect.W)
			sHeight := min(7, contentH)
			sX := max(v.editorRect.X, v.editorRect.X+v.editorRect.W-sWidth)
			sY := v.editorRect.Y
			v.searchRect = loom.Rect{X: sX, Y: sY, W: sWidth, H: sHeight}

			if sWidth > 4 && sHeight > 2 {
				searchBox := loom.Box{Title: "Search  F3 / ^F", Style: boxStyle}
				searchBox.Draw(canvas, v.searchRect)

				// Find input line
				v.searchBar.MaxWidth = max(1, sWidth-4)
				v.searchBar.Draw(canvas, loom.Rect{X: sX + 2, Y: sY + 1, W: max(1, sWidth-4), H: 1})

				if sHeight > 3 {
					v.searchModes = v.modeBar()
					v.searchModes.Draw(canvas, loom.Rect{X: sX + 1, Y: sY + 2, W: sWidth - 2, H: 1})
				}

				// Match counter line
				if sHeight > 4 {
					matchStr := "No matches"
					if v.searchErr != "" {
						matchStr = v.searchErr
					} else if len(v.searchMatches) > 0 {
						matchStr = fmt.Sprintf("%d / %d matches", v.searchIndex+1, len(v.searchMatches))
					}
					canvas.WriteDefault(sX+2, sY+3, loom.TruncateText(matchStr, sWidth-4, ""), dimStyle)
				}

				// Nav hints
				if sHeight > 5 {
					canvas.WriteDefault(sX+2, sY+4, loom.TruncateText("Enter: next   S-Enter: prev", sWidth-4, ""), dimStyle)
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
	if key.Is(loom.SpeccedDefaults.Editor.HotkeyScreenshotBinding) {
		v.reportError(v.captureScreenshot())
		return loom.Handled()
	}

	if v.unsavedDialog != nil && v.unsavedDialog.Open {
		_ = v.unsavedDialog.ConsumeKey(key)
		if !v.unsavedDialog.Open {
			v.unsavedDialog = nil
		}
		return loom.Handled()
	}

	editorDefs := loom.SpeccedDefaults.Editor

	// Save works whichever panel has focus; the editor reports the new state.
	if key.Is(loom.SpeccedDefaults.RichTextEdit.HotkeySaveBinding) {
		v.reportError(v.edit.Save())
		return loom.Handled()
	}

	// Open and Close work whichever panel has focus; the editor owns the actions.
	if key.Is(loom.SpeccedDefaults.RichTextEdit.HotkeyOpenBinding) {
		v.edit.RequestOpen()
		return loom.Handled()
	}
	if key.Is(loom.SpeccedDefaults.RichTextEdit.HotkeyCloseBinding) {
		v.edit.RequestClose()
		return loom.Handled()
	}

	if key.Is(editorDefs.HotkeyFilesBinding) {
		v.toggleSidePanel()
		return loom.Handled()
	}

	if key.Is(editorDefs.HotkeySearchBinding, "ctrl-f") {
		v.toggleSearch()
		return loom.Handled()
	}

	if key.Is("tab", "shift-tab") && (v.showSearch || v.showSidePanel) {
		focuses := []editFocus{focusEditor}
		if v.showSidePanel {
			focuses = append(focuses, focusBrowser)
		}
		if v.showSearch {
			focuses = append(focuses, focusSearch)
		}
		for i, focus := range focuses {
			if focus == v.focused {
				step := 1
				if key.Is("shift-tab") {
					step = -1
				}
				v.focused = focuses[(i+step+len(focuses))%len(focuses)]
				break
			}
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
		if key.Is("ctrl-r") {
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
		if res.Quit {
			// The search bar aborts on quit keys; leave them to the Pane.
			return loom.Ignored()
		}
		return res
	}

	if v.showSidePanel && v.focused == focusBrowser {
		res := v.filePicker.ConsumeKey(key)
		if res.Quit {
			return loom.Ignored()
		}
		return res
	}

	if res := v.hotkeyBar().ConsumeKey(key); res.Consumed {
		return res
	}

	res := v.edit.ConsumeKey(key)
	if res.Quit {
		return loom.Ignored()
	}
	if v.showSearch && res.Consumed {
		v.updateSearchMatches()
	}
	return res
}

func (v *editView) reportError(err error) {
	if err != nil {
		v.statusMessage = err.Error()
	}
}

func (v *editView) modeBar() *loom.HintBar {
	normal, regex := "Normal", "Regex"
	if v.regexMode {
		regex += " *"
	} else {
		normal += " *"
	}
	bar := loom.NewHintBar(
		loom.HintEntry{Key: "", Label: normal, Action: func() loom.EventResult {
			v.regexMode = false
			v.searchIndex = 0
			v.updateSearchMatches()
			return loom.Handled()
		}},
		loom.HintEntry{Key: "^R", Binding: "ctrl-r", Label: regex, Action: func() loom.EventResult {
			v.regexMode = true
			v.searchIndex = 0
			v.updateSearchMatches()
			return loom.Handled()
		}},
	)
	bar.ApplyTheme(loom.Theme(v.config.Theme))
	return bar
}

func (v *editView) captureScreenshot() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, "Pictures", "Screenshots")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	num := 1
	for _, entry := range entries {
		prefix, _, ok := strings.Cut(entry.Name(), "-loom-edit-")
		if ok {
			if n, err := strconv.Atoi(prefix); err == nil && n >= num {
				num = n + 1
			}
		}
	}
	details := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, filepath.Base(v.filePath))
	if details == "" || details == "." {
		details = "document"
	}
	w, h := v.lastRect.W, v.lastRect.H
	if w <= 0 || h <= 0 {
		w, h = 100, 24
	}
	content := strings.Join(loom.Render(v, w, h), "\n") + "\n"
	for {
		path := filepath.Join(dir, fmt.Sprintf("%02d-loom-edit-%s.ansi", num, details))
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if os.IsExist(err) {
			num++
			continue
		}
		if err != nil {
			return err
		}
		_, writeErr := file.WriteString(content)
		closeErr := file.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		v.statusMessage = "Saved screenshot: " + path
		return nil
	}
}

func findRuneMatches(line string, query string) []int {
	lineRunes := []rune(line)
	queryRunes := []rune(query)
	if len(queryRunes) == 0 || len(lineRunes) < len(queryRunes) {
		return nil
	}
	var matches []int
	for i := 0; i <= len(lineRunes)-len(queryRunes); i++ {
		match := true
		for j := 0; j < len(queryRunes); j++ {
			r1 := unicode.ToLower(lineRunes[i+j])
			r2 := unicode.ToLower(queryRunes[j])
			if r1 != r2 {
				match = false
				break
			}
		}
		if match {
			matches = append(matches, i)
		}
	}
	return matches
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
		if !v.unsavedDialog.Open {
			v.unsavedDialog = nil
		}
		return loom.Handled()
	}

	if v.lastRect.H > 2 && mouse.Y == v.lastRect.Y+v.lastRect.H-2 {
		mouseLocal := mouse
		mouseLocal.Y = 0
		mouseLocal.X -= v.lastRect.X + 1
		return v.hotkeyBar().ConsumeMouse(mouseLocal)
	}

	// Click in search panel
	if v.showSearch && v.searchRect.Contains(mouse.X, mouse.Y) {
		if mouse.Action == loom.MousePress {
			v.focused = focusSearch
			v.searchBar.SetFocused(true)
		}
		if mouse.Y == v.searchRect.Y+2 && v.searchModes != nil {
			local := mouse
			local.X -= v.searchRect.X + 1
			local.Y = 0
			_ = v.searchModes.ConsumeMouse(local)
			return loom.Handled()
		}
		if mouse.Y != v.searchRect.Y+1 {
			return loom.Handled()
		}

		mouseLocal := mouse
		mouseLocal.X -= v.searchRect.X + 2
		mouseLocal.Y -= v.searchRect.Y + 1
		_ = v.searchBar.ConsumeMouse(mouseLocal)
		return loom.Handled()
	}

	// Click in side panel
	if v.showSidePanel && v.browserRect.Contains(mouse.X, mouse.Y) {
		if mouse.Action == loom.MousePress {
			v.focused = focusBrowser
		}
		mouseLocal := mouse
		mouseLocal.X -= v.browserRect.X
		mouseLocal.Y -= v.browserRect.Y
		return v.filePicker.ConsumeMouse(mouseLocal)
	}

	// Click in editor area
	if v.editorRect.Contains(mouse.X, mouse.Y) {
		if mouse.Action == loom.MousePress {
			v.focused = focusEditor
		}
		mouseLocal := mouse
		mouseLocal.X -= v.editorRect.X
		mouseLocal.Y -= v.editorRect.Y
		return v.edit.ConsumeMouse(mouseLocal)
	}

	return loom.Ignored()
}
