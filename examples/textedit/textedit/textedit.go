// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package textedit provides a full-featured text editor example application in loom,
// featuring keybindings (C-s, C-S-s, C-o, C-c, C-v, C-x), MRU file list,
// syntax highlighting, collapsible filebrowser, and an embedded terminal.
package textedit

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/measure"
	"codeberg.org/ubunatic/loom/syntax"
	"github.com/spf13/cobra"
)

type focusArea int

const (
	focusEditor focusArea = iota
	focusBrowser
	focusTerminal
)

// TextEditApp is the root text editor application widget.
type TextEditApp struct {
	frame       *loom.Frame
	hSplit      *loom.Split // divides filebrowser and editor+terminal
	vSplit      *loom.Split // divides editor and terminal
	editor      *loom.TextArea
	editorW     *editorWidget
	fileBrowser *fileBrowserWidget
	sidebar     *sidebarWidget
	terminal    *terminalWidget

	activePath  string
	modified    bool
	mruList     []string
	clipboard   string
	showBrowser bool
	activeFocus focusArea
	statusMsg   string
}

type editorWidget struct {
	app *TextEditApp
}

func (ew *editorWidget) Draw(c *loom.Canvas, r loom.Rect) {
	focused := ew.app.activeFocus == focusEditor
	textRect := loom.Rect{X: r.X + 5, Y: r.Y, W: max(0, r.W-5), H: r.H}
	ew.app.editor.Draw(c, textRect, focused)

	vis := ew.app.editor.VisibleLines()
	scroll := ew.app.editor.Scroll()

	foldStarts := make(map[int]int)
	if nav, ok := ew.app.editor.Highlighter().(syntax.Navigator); ok {
		for _, f := range nav.Folds() {
			foldStarts[f[0]] = f[1]
		}
	}

	for i := 0; i < r.H; i++ {
		visIdx := scroll + i
		if visIdx < len(vis) {
			docLine := vis[visIdx]
			icon := " "
			if end, isFold := foldStarts[docLine]; isFold && end > docLine {
				if ew.app.editor.IsFolded(docLine) {
					icon = "▸"
				} else {
					icon = "▾"
				}
			}
			gutter := fmt.Sprintf("%2d%s│ ", docLine+1, icon)
			c.Write(r.X, r.Y+i, gutter, loom.Style{Dim: true, FG: loom.ColorIndex(8)})
		} else {
			c.Write(r.X, r.Y+i, "   │ ", loom.Style{Dim: true, FG: loom.ColorIndex(8)})
		}
	}
}

func (ew *editorWidget) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	prevVal := ew.app.editor.Value()
	result := ew.app.editor.ConsumeKey(e)
	if ew.app.editor.Value() != prevVal {
		ew.app.modified = true
	}
	return result
}

func (ew *editorWidget) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	if e.Action == loom.MousePress && e.Button == loom.MouseLeft && e.X < 5 {
		vis := ew.app.editor.VisibleLines()
		visIdx := ew.app.editor.Scroll() + e.Y
		if visIdx >= 0 && visIdx < len(vis) {
			docLine := vis[visIdx]
			if nav, ok := ew.app.editor.Highlighter().(syntax.Navigator); ok {
				for _, f := range nav.Folds() {
					if f[0] == docLine {
						ew.app.editor.ToggleFold(f[0], f[1])
						return loom.QuitResult()
					}
				}
			}
		}
	}
	return loom.Ignored()
}

// SidebarMode defines the active tab in the left sidebar.
type SidebarMode int

const (
	SidebarFiles SidebarMode = iota
	SidebarOutline
)

type sidebarWidget struct {
	app     *TextEditApp
	browser *fileBrowserWidget
	outline *outlineWidget
	mode    SidebarMode
}

func newSidebar(dir string, app *TextEditApp) *sidebarWidget {
	return &sidebarWidget{
		app:     app,
		browser: newFileBrowser(dir, app),
		outline: &outlineWidget{app: app},
		mode:    SidebarFiles,
	}
}

func (sb *sidebarWidget) Draw(c *loom.Canvas, r loom.Rect) {
	c.PaintSurface(r, loom.Style{})
	if r.H <= 0 || r.W <= 0 {
		return
	}
	filesStyle := loom.Style{FG: loom.ColorIndex(8), Dim: true}
	outlineStyle := loom.Style{FG: loom.ColorIndex(8), Dim: true}
	if sb.mode == SidebarFiles {
		filesStyle = loom.Style{Bold: true, FG: loom.ColorIndex(15), BG: loom.ColorIndex(238)}
	} else {
		outlineStyle = loom.Style{Bold: true, FG: loom.ColorIndex(15), BG: loom.ColorIndex(238)}
	}
	c.Write(r.X, r.Y, "📁 Explorer", filesStyle)
	c.Write(r.X+12, r.Y, "📋 Outline", outlineStyle)

	contentRect := loom.Rect{X: r.X, Y: r.Y + 1, W: r.W, H: max(0, r.H-1)}
	if sb.mode == SidebarFiles {
		sb.browser.DrawEntries(c, contentRect)
	} else {
		sb.outline.Draw(c, contentRect)
	}
}

func (sb *sidebarWidget) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if e.Is("ctrl-t", "alt-o", "o", "f") {
		if sb.mode == SidebarFiles {
			sb.mode = SidebarOutline
		} else {
			sb.mode = SidebarFiles
		}
		return loom.QuitResult()
	}
	if sb.mode == SidebarFiles {
		return sb.browser.ConsumeKey(e)
	}
	return sb.outline.ConsumeKey(e)
}

func (sb *sidebarWidget) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	if e.Action == loom.MousePress && e.Button == loom.MouseLeft && e.Y == 0 {
		if e.X < 12 {
			sb.mode = SidebarFiles
		} else {
			sb.mode = SidebarOutline
		}
		return loom.QuitResult()
	}
	childEv := e
	childEv.Y = e.Y - 1
	if childEv.Y < 0 {
		return loom.Ignored()
	}
	if sb.mode == SidebarFiles {
		return sb.browser.ConsumeMouse(e)
	}
	return sb.outline.ConsumeMouse(childEv)
}

type outlineWidget struct {
	app         *TextEditApp
	symbols     []syntax.Symbol
	selected    int
	doubleClick *loom.DoubleClickRecognizer
}

func (ow *outlineWidget) refresh() {
	if nav, ok := ow.app.editor.Highlighter().(syntax.Navigator); ok {
		ow.symbols = nav.Symbols()
	} else {
		ow.symbols = nil
	}
	if ow.selected >= len(ow.symbols) {
		ow.selected = max(0, len(ow.symbols)-1)
	}
}

func (ow *outlineWidget) Draw(c *loom.Canvas, r loom.Rect) {
	ow.refresh()
	if len(ow.symbols) == 0 {
		c.Write(r.X, r.Y, "  No symbols", loom.Style{Dim: true, FG: loom.ColorIndex(8)})
		return
	}
	for i, sym := range ow.symbols {
		if i >= r.H {
			break
		}
		icon := "▪ "
		switch sym.Kind {
		case "function", "method":
			icon = "🔧 "
		case "type":
			icon = "🔷 "
		case "heading":
			icon = "📄 "
		case "key":
			icon = "🔑 "
		}
		prefix := "  "
		if i == ow.selected {
			prefix = "▸ "
		}
		style := loom.Style{}
		if i == ow.selected {
			style = loom.Style{Bold: true, FG: loom.ColorIndex(15), BG: loom.ColorIndex(24)}
		}
		line := loom.TruncateText(prefix+icon+sym.Name, r.W, "")
		if i == ow.selected {
			line += strings.Repeat(" ", max(0, r.W-measure.StringWidth(line)))
		}
		c.Write(r.X, r.Y+i, line, style)
	}
}

func (ow *outlineWidget) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	ow.refresh()
	switch {
	case e.Is("up", "k"):
		if ow.selected > 0 {
			ow.selected--
		}
		return loom.QuitResult()
	case e.Is("down", "j"):
		if ow.selected < len(ow.symbols)-1 {
			ow.selected++
		}
		return loom.QuitResult()
	case e.Is("enter"):
		if len(ow.symbols) > 0 && ow.selected >= 0 && ow.selected < len(ow.symbols) {
			sym := ow.symbols[ow.selected]
			ow.app.editor.SetCaret(sym.Line, sym.Column)
			ow.app.activeFocus = focusEditor
			ow.app.statusMsg = fmt.Sprintf("Jumped to %s (line %d)", sym.Name, sym.Line+1)
		}
		return loom.QuitResult()
	}
	return loom.Ignored()
}

func (ow *outlineWidget) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	ow.refresh()
	if ow.doubleClick == nil {
		ow.doubleClick = loom.NewDoubleClickRecognizer(nil)
	}
	target := ""
	if e.Y >= 0 && e.Y < len(ow.symbols) {
		sym := ow.symbols[e.Y]
		target = fmt.Sprintf("%d:%d:%s", sym.Line, sym.Column, sym.Name)
	}
	doubleClick := ow.doubleClick.Handle(e, target)
	if e.Action == loom.MousePress && e.Button == loom.MouseLeft && target != "" {
		ow.selected = e.Y
		if doubleClick {
			sym := ow.symbols[ow.selected]
			ow.app.editor.SetCaret(sym.Line, sym.Column)
			ow.app.activeFocus = focusEditor
			ow.app.statusMsg = fmt.Sprintf("Jumped to %s (line %d)", sym.Name, sym.Line+1)
		}
		return loom.QuitResult()
	}
	return loom.Ignored()
}

type fileBrowserWidget struct {
	dir      string
	files    []os.DirEntry
	selected int
	app      *TextEditApp
}

func newFileBrowser(dir string, app *TextEditApp) *fileBrowserWidget {
	fb := &fileBrowserWidget{dir: dir, app: app}
	fb.reload()
	return fb
}

func (fb *fileBrowserWidget) reload() {
	entries, err := os.ReadDir(fb.dir)
	if err != nil {
		fb.files = nil
		return
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})
	fb.files = entries
	if fb.selected >= len(fb.files) {
		fb.selected = max(0, len(fb.files)-1)
	}
}

func (fb *fileBrowserWidget) Draw(c *loom.Canvas, r loom.Rect) {
	c.PaintSurface(r, loom.Style{})
	c.Write(r.X, r.Y, "📁 Explorer", loom.Style{Bold: true, FG: loom.ColorIndex(7)})
	fb.DrawEntries(c, loom.Rect{X: r.X, Y: r.Y + 1, W: r.W, H: max(0, r.H-1)})
}

func (fb *fileBrowserWidget) DrawEntries(c *loom.Canvas, r loom.Rect) {
	for i, entry := range fb.files {
		if i >= r.H {
			break
		}
		prefix := "  "
		name := entry.Name()
		if entry.IsDir() {
			prefix = "▸ "
		}
		if i == fb.selected {
			prefix = "▸ "
		}
		style := loom.Style{}
		if i == fb.selected {
			style = loom.Style{Bold: true, FG: loom.ColorIndex(15), BG: loom.ColorIndex(24)}
		} else if entry.IsDir() {
			style = loom.Style{FG: loom.ColorIndex(4)}
		}
		line := loom.TruncateText(prefix+name, r.W, "")
		if i == fb.selected {
			line += strings.Repeat(" ", max(0, r.W-measure.StringWidth(line)))
		}
		c.Write(r.X, r.Y+i, line, style)
	}
}

func (fb *fileBrowserWidget) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	switch {
	case e.Is("up", "k"):
		if fb.selected > 0 {
			fb.selected--
		}
		return loom.QuitResult()
	case e.Is("down", "j"):
		if fb.selected < len(fb.files)-1 {
			fb.selected++
		}
		return loom.QuitResult()
	case e.Is("backspace", "esc"):
		parent := filepath.Dir(filepath.Clean(fb.dir))
		if parent != filepath.Clean(fb.dir) {
			fb.dir = parent
			fb.reload()
		}
		return loom.QuitResult()
	case e.Is("enter"):
		if len(fb.files) == 0 {
			return loom.QuitResult()
		}
		sel := fb.files[fb.selected]
		target := filepath.Join(fb.dir, sel.Name())
		if sel.IsDir() {
			fb.dir = target
			fb.reload()
		} else {
			fb.app.OpenFile(target)
		}
		return loom.QuitResult()
	}
	return loom.Ignored()
}

func (fb *fileBrowserWidget) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	if e.Action == loom.MousePress && e.Button == loom.MouseLeft && e.Y > 0 {
		idx := e.Y - 1
		if idx >= 0 && idx < len(fb.files) {
			fb.selected = idx
			sel := fb.files[fb.selected]
			if !sel.IsDir() {
				fb.app.OpenFile(filepath.Join(fb.dir, sel.Name()))
			}
			return loom.QuitResult()
		}
	}
	return loom.Ignored()
}

type terminalWidget struct {
	logs  []string
	input *loom.TextInput
	app   *TextEditApp
}

func newTerminal(app *TextEditApp) *terminalWidget {
	t := &terminalWidget{
		logs:  []string{"Loom Embedded Terminal v1.0", "Type 'help' for commands."},
		input: loom.NewTextInput(""),
		app:   app,
	}
	t.input.Prompt = "❯ "
	return t
}

func (tw *terminalWidget) Draw(c *loom.Canvas, r loom.Rect) {
	c.PaintSurface(r, loom.Style{})
	c.Write(r.X, r.Y, "💻 Terminal", loom.Style{Bold: true, FG: loom.ColorIndex(7)})
	maxLogs := r.H - 2
	if maxLogs < 0 {
		maxLogs = 0
	}
	startLog := max(0, len(tw.logs)-maxLogs)
	visibleLogs := tw.logs[startLog:]
	for i, logLine := range visibleLogs {
		style := loom.Style{FG: loom.ColorIndex(7)}
		if strings.HasPrefix(logLine, "$ ") {
			style = loom.Style{FG: loom.ColorIndex(2), Bold: true}
		} else if strings.HasPrefix(logLine, "Loom Embedded") || strings.HasPrefix(logLine, "Type ") {
			style = loom.Style{FG: loom.ColorIndex(8), Dim: true}
		}
		c.Write(r.X, r.Y+1+i, logLine, style)
	}

	if r.H > 1 {
		inputRect := loom.Rect{X: r.X, Y: r.Y + r.H - 1, W: r.W, H: 1}
		focused := tw.app.activeFocus == focusTerminal
		c.Write(inputRect.X, inputRect.Y, "❯ ", loom.Style{FG: loom.ColorIndex(2), Bold: true})
		inputRect.X += 2
		inputRect.W = max(0, inputRect.W-2)
		tw.input.Prompt = ""
		tw.input.Draw(c, inputRect, focused)
	}
}

func (tw *terminalWidget) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if e.Is("enter") {
		cmd := tw.input.Value()
		tw.input.SetValue("")
		if strings.TrimSpace(cmd) != "" {
			tw.execCommand(strings.TrimSpace(cmd))
		}
		return loom.QuitResult()
	}
	return tw.input.ConsumeKey(e)
}

func (tw *terminalWidget) ConsumeMouse(e loom.MouseEvent) loom.EventResult { return loom.Ignored() }

func (tw *terminalWidget) execCommand(cmd string) {
	tw.logs = append(tw.logs, "$ "+cmd)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "help":
		tw.logs = append(tw.logs, "Available commands: help, ls, cat <file>, date, echo <msg>, clear")
	case "clear":
		tw.logs = nil
	case "date":
		tw.logs = append(tw.logs, time.Now().Format(time.RFC1123))
	case "echo":
		tw.logs = append(tw.logs, strings.Join(parts[1:], " "))
	case "ls":
		entries, err := os.ReadDir(".")
		if err != nil {
			tw.logs = append(tw.logs, "ls error: "+err.Error())
		} else {
			var names []string
			for _, e := range entries {
				names = append(names, e.Name())
			}
			tw.logs = append(tw.logs, strings.Join(names, "  "))
		}
	case "cat":
		if len(parts) < 2 {
			tw.logs = append(tw.logs, "cat: missing filename")
		} else {
			data, err := os.ReadFile(parts[1])
			if err != nil {
				tw.logs = append(tw.logs, "cat: "+err.Error())
			} else {
				lines := strings.Split(string(data), "\n")
				if len(lines) > 5 {
					lines = append(lines[:5], fmt.Sprintf("... (%d more lines)", len(lines)-5))
				}
				tw.logs = append(tw.logs, lines...)
			}
		}
	default:
		tw.logs = append(tw.logs, fmt.Sprintf("Command executed: %s", parts[0]))
	}
}

func engineForPath(path string) syntax.Engine {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return syntax.NewLexicalEngine("go")
	case ".md":
		return syntax.NewLexicalEngine("markdown")
	case ".json":
		return syntax.NewLexicalEngine("json")
	case ".yaml", ".yml":
		return syntax.NewLexicalEngine("yaml")
	default:
		return nil
	}
}

// NewApp initializes and returns a new TextEditApp.
func NewApp() *TextEditApp {
	app := &TextEditApp{
		activePath:  "demo.go",
		mruList:     []string{"demo.go", "README.md", "go.mod"},
		showBrowser: true,
		activeFocus: focusEditor,
		statusMsg:   "Ready",
	}

	initialContent := `package main

import "fmt"

// Welcome to Loom Text Editor
func main() {
	fmt.Println("Hello, Loom!")
}`
	app.editor = loom.NewTextArea(initialContent)
	app.editor.SetHighlighter(engineForPath(app.activePath))
	app.editorW = &editorWidget{app: app}
	app.sidebar = newSidebar(".", app)
	app.fileBrowser = app.sidebar.browser
	app.terminal = newTerminal(app)

	app.vSplit = loom.NewSplit(app.editorW, app.terminal)
	app.vSplit.Orientation = loom.Vertical
	app.vSplit.Ratio = 0.7
	app.vSplit.MinFirst = 5
	app.vSplit.MinSecond = 3

	app.hSplit = loom.NewSplit(app.sidebar, app.vSplit)
	app.hSplit.Orientation = loom.Horizontal
	app.hSplit.Ratio = 0.25
	app.hSplit.MinFirst = 12
	app.hSplit.MinSecond = 20

	border := loom.BoxBorder{
		TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
		Horizontal: "─", Vertical: "│", TitlePrefix: " ", TitleSuffix: " ",
	}

	app.frame = &loom.Frame{
		Title: "Text Editor",
		Boxes: []loom.Box{
			{ID: "main", Title: "Text Editor Workspace", Dynamic: true, FillHeight: true, Border: border, Child: app.hSplit},
		},
		Actions: []loom.FrameAction{
			{ID: "save", Action: "save", Key: "ctrl-s", Hint: "C-s: Save"},
			{ID: "open", Action: "open", Key: "ctrl-o", Hint: "C-o: Open"},
			{ID: "fold", Action: "fold", Key: "f2", Hint: "F2: Fold"},
			{ID: "clip", Action: "clip", Key: "ctrl-c", Hint: "C-c/v/x: Clip"},
			{ID: "sidebar", Action: "sidebar", Key: "ctrl-b", Hint: "C-b: Sidebar"},
			{ID: "focus", Action: "focus", Key: "tab", Hint: "Tab: Focus"},
			{ID: "quit_f10", Action: "quit", Key: "f10", Hint: "F10 Quit"},
			{ID: "quit_ctrl_q", Action: "quit", Key: "ctrl-q"},
		},
		ControlSeparator: " • ",
	}

	return app
}

// OpenFile opens path into the editor buffer and updates MRU list.
func (app *TextEditApp) OpenFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		app.statusMsg = "Error reading " + filepath.Base(path) + ": " + err.Error()
		return
	}
	app.activePath = path
	app.editor.SetValue(string(data))
	app.editor.SetHighlighter(engineForPath(path))
	app.modified = false
	app.touchMRU(path)
	app.statusMsg = "Opened " + filepath.Base(path)
}

// SaveFile writes active editor contents to disk.
func (app *TextEditApp) SaveFile(path string) {
	if path == "" {
		path = app.activePath
	}
	if path == "" {
		path = "untitled.txt"
	}
	err := os.WriteFile(path, []byte(app.editor.Value()), 0644)
	if err != nil {
		app.statusMsg = "Save error: " + err.Error()
		return
	}
	app.activePath = path
	app.editor.SetHighlighter(engineForPath(path))
	app.modified = false
	app.touchMRU(path)
	app.statusMsg = "Saved " + filepath.Base(path)
}

func (app *TextEditApp) touchMRU(path string) {
	newMRU := []string{path}
	for _, p := range app.mruList {
		if p != path {
			newMRU = append(newMRU, p)
		}
	}
	if len(newMRU) > 10 {
		newMRU = newMRU[:10]
	}
	app.mruList = newMRU
}

// SidebarMode returns the active sidebar tab.
func (app *TextEditApp) SidebarMode() SidebarMode {
	if app.sidebar != nil {
		return app.sidebar.mode
	}
	return SidebarFiles
}

// SetSidebarMode sets the active sidebar tab.
func (app *TextEditApp) SetSidebarMode(mode SidebarMode) {
	if app.sidebar != nil {
		app.sidebar.mode = mode
	}
}

// ToggleSidebarMode toggles between Explorer and Outline tabs.
func (app *TextEditApp) ToggleSidebarMode() {
	if app.sidebar != nil {
		if app.sidebar.mode == SidebarFiles {
			app.sidebar.mode = SidebarOutline
		} else {
			app.sidebar.mode = SidebarFiles
		}
	}
}

// Outline returns the outline widget.
func (app *TextEditApp) Outline() *outlineWidget {
	if app.sidebar != nil {
		return app.sidebar.outline
	}
	return nil
}

// ToggleFold toggles folding on the given start and end lines.
func (app *TextEditApp) ToggleFold(startLine, endLine int) {
	app.editor.ToggleFold(startLine, endLine)
}

func (app *TextEditApp) toggleFoldAtCaret() {
	row, _ := app.editor.Caret()
	if nav, ok := app.editor.Highlighter().(syntax.Navigator); ok {
		for _, fold := range nav.Folds() {
			start, end := fold[0], fold[1]
			if row == start || (row > start && row <= end) {
				app.editor.ToggleFold(start, end)
				if app.editor.IsFolded(start) {
					app.statusMsg = fmt.Sprintf("Folded lines %d-%d", start+1, end+1)
				} else {
					app.statusMsg = fmt.Sprintf("Unfolded lines %d-%d", start+1, end+1)
				}
				return
			}
		}
	}
	app.statusMsg = "No foldable block at caret"
}

func (app *TextEditApp) Draw(c *loom.Canvas, r loom.Rect) {
	focusName := "Editor"
	switch app.activeFocus {
	case focusBrowser:
		focusName = "Sidebar"
	case focusTerminal:
		focusName = "Terminal"
	}

	app.frame.Boxes[0].Title = ""
	app.frame.Status = ""
	app.frame.Draw(c, r)
	if r.H == 0 || r.W == 0 {
		return
	}
	// Header badges are drawn over the frame title row.
	c.Fill(loom.Rect{X: r.X, Y: r.Y, W: r.W, H: 1}, loom.Cell{Text: " ", Style: loom.Style{FG: loom.ColorIndex(6)}})
	x := r.X + 2
	x += c.Write(x, r.Y, "Loom TextEdit", loom.Style{FG: loom.ColorIndex(7), Bold: true}) + 2
	file := filepath.Base(app.activePath)
	if file == "." || file == "" {
		file = "untitled"
	}
	mark := ""
	if app.modified {
		mark = " ●"
	}

	crumb := "📁 " + file
	if nav, ok := app.editor.Highlighter().(syntax.Navigator); ok {
		row, col := app.editor.Caret()
		crumbs := nav.Breadcrumb(row, col)
		for _, c := range crumbs {
			crumb += " › 🔧 " + c
		}
	}
	crumb += mark

	x += c.Write(x, r.Y, "["+crumb+"]", loom.Style{FG: loom.ColorIndex(7), Bold: true, BG: loom.ColorIndex(238)}) + 2
	x += c.Write(x, r.Y, "[Focus: ", loom.Style{FG: loom.ColorIndex(7), BG: loom.ColorIndex(238)})
	x += c.Write(x, r.Y, focusName, loom.Style{FG: loom.ColorIndex(14), Bold: true, BG: loom.ColorIndex(238)})
	x += c.Write(x, r.Y, "]", loom.Style{FG: loom.ColorIndex(7), BG: loom.ColorIndex(238)}) + 2
	c.Write(x, r.Y, "[v1.0]", loom.Style{FG: loom.ColorIndex(8), Dim: true})
	if r.H > 1 {
		app.drawStatus(c, loom.Rect{X: r.X, Y: r.Y + r.H - 1, W: r.W, H: 1})
	}
}

func (app *TextEditApp) drawStatus(c *loom.Canvas, r loom.Rect) {
	bg := loom.Style{BG: loom.ColorIndex(236)}
	c.Fill(r, loom.Cell{Text: " ", Style: bg})
	x := r.X
	label := "Ready"
	if app.statusMsg != "" {
		label = app.statusMsg
	}
	if app.modified {
		label = "Modified"
	}
	pill := "● " + label
	x += c.Write(x, r.Y, loom.TruncateText(" "+pill+" ", min(r.W, 22), ""), loom.Style{FG: loom.ColorIndex(2), BG: loom.ColorIndex(236), Bold: true}) + 1
	keys := [][2]string{{"C-s", "Save"}, {"C-o", "Open"}, {"F2", "Fold"}, {"C-c/v", "Clip"}, {"Tab", "Focus"}, {"C-b", "Tree"}, {"F10", "Quit"}}
	for _, key := range keys {
		text := " " + key[0] + " " + key[1] + " "
		if x+len([]rune(text)) > r.X+r.W {
			break
		}
		c.Write(x, r.Y, text, loom.Style{FG: loom.ColorIndex(15), BG: loom.ColorIndex(239), Bold: true})
		x += len([]rune(text)) + 1
	}
}

func (app *TextEditApp) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	// F10 and Ctrl-Q always exit immediately, regardless of focus.
	if e.Is("f10", "F10", "ctrl-q", "ctrl-Q") {
		return loom.QuitResult()
	}

	// Global Keybindings
	switch {
	case e.Is("f2", "F2"):
		app.toggleFoldAtCaret()
		return loom.Ignored()
	case e.Is("ctrl-s", "ctrl-S"):
		app.SaveFile(app.activePath)
		return loom.Ignored()
	case e.Is("ctrl-shift-s"):
		saveAsPath := app.activePath + ".bak"
		app.SaveFile(saveAsPath)
		return loom.Ignored()
	case e.Is("ctrl-o"):
		// Open next file from MRU list
		if len(app.mruList) > 1 {
			nextPath := app.mruList[1]
			app.OpenFile(nextPath)
		} else {
			app.statusMsg = "MRU list empty"
		}
		return loom.Ignored()
	case e.Is("ctrl-c"):
		val := app.editor.Value()
		lines := strings.Split(val, "\n")
		row, _ := app.editor.Caret()
		if row >= 0 && row < len(lines) {
			app.clipboard = lines[row]
			app.statusMsg = "Copied line to clipboard"
		}
		return loom.Ignored()
	case e.Is("ctrl-v"):
		if app.clipboard != "" {
			for _, ch := range app.clipboard {
				app.editor.ConsumeKey(loom.KeyEvent{Text: string(ch)})
			}
			app.modified = true
			app.statusMsg = "Pasted from clipboard"
		}
		return loom.Ignored()
	case e.Is("ctrl-x"):
		val := app.editor.Value()
		lines := strings.Split(val, "\n")
		row, _ := app.editor.Caret()
		if row >= 0 && row < len(lines) {
			app.clipboard = lines[row]
			lines = append(lines[:row], lines[row+1:]...)
			app.editor.SetValue(strings.Join(lines, "\n"))
			app.modified = true
			app.statusMsg = "Cut line to clipboard"
		}
		return loom.Ignored()
	case e.Is("ctrl-b"):
		app.showBrowser = !app.showBrowser
		if app.showBrowser {
			app.hSplit.SetRatio(0.25)
			app.statusMsg = "Sidebar opened"
		} else {
			app.hSplit.SetRatio(0.0)
			app.statusMsg = "Sidebar collapsed"
		}
		return loom.Ignored()
	case e.Is("tab", "f6"):
		app.activeFocus = (app.activeFocus + 1) % 3
		app.statusMsg = fmt.Sprintf("Switched focus to %v", app.activeFocus)
		return loom.Ignored()
	}

	// Dispatch to active focus area
	switch app.activeFocus {
	case focusBrowser:
		if result := app.sidebar.ConsumeKey(e); result.Consumed {
			return result
		}
	case focusTerminal:
		if result := app.terminal.ConsumeKey(e); result.Consumed {
			return result
		}
	case focusEditor:
		if result := app.editorW.ConsumeKey(e); result.Consumed {
			return result
		}
	}

	// The frame has its own focus tree, which is not the same as activeFocus.
	// Forwarding an unhandled key would send it to a second widget and turn that
	// widget's consumed result into quit.
	return loom.Ignored()
}

func (app *TextEditApp) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	app.frame.ConsumeMouse(e)
	return loom.Ignored()
}

// PaneRequest declares terminal requirements.
func (app *TextEditApp) PaneRequest() loom.PaneRequest {
	return loom.PaneRequest{
		Mouse:      1003,
		Resizeable: true,
		OwnsQuit:   true,
	}
}

// NewWidget builds the textedit root widget for in-process hosting.
func NewWidget(args []string) (loom.Widget, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("textedit: unexpected arguments: %v", args)
	}
	return NewApp(), nil
}

// Run executes the textedit interactive example command.
func Run(args []string) error {
	cmd := &cobra.Command{
		Use:           "textedit",
		Short:         "Multi-pane text editor with keybindings, MRU, filebrowser, and embedded terminal",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run()
		},
	}
	cmd.SetArgs(args)
	return cmd.Execute()
}

func run() error {
	app := NewApp()
	pane, err := loom.New(1 << 16)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.DisableDefaultQuit = true
	pane.EnableMouse()
	return pane.Run(app)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
