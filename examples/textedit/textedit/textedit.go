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
	ew.app.editor.Draw(c, r, focused)
	ew.app.applySyntaxHighlighting(c, r)
}

func (ew *editorWidget) HandleKey(e loom.KeyEvent) bool {
	prevVal := ew.app.editor.Value()
	handled := ew.app.editor.HandleKey(e)
	if ew.app.editor.Value() != prevVal {
		ew.app.modified = true
	}
	return handled
}

func (ew *editorWidget) HandleMouse(e loom.MouseEvent) bool { return false }

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
	c.Write(r.X, r.Y, "Files", loom.Style{Bold: true})
	for i, entry := range fb.files {
		if i+1 >= r.H {
			break
		}
		prefix := "  "
		if i == fb.selected {
			prefix = "> "
		}
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		style := loom.Style{}
		if i == fb.selected {
			style = loom.Style{Bold: true, FG: loom.ColorIndex(4)} // Cyan
		} else if entry.IsDir() {
			style = loom.Style{FG: loom.ColorIndex(4)}
		}
		c.Write(r.X, r.Y+i+1, prefix+name, style)
	}
}

func (fb *fileBrowserWidget) HandleKey(e loom.KeyEvent) bool {
	switch {
	case e.Is("up", "k"):
		if fb.selected > 0 {
			fb.selected--
		}
		return true
	case e.Is("down", "j"):
		if fb.selected < len(fb.files)-1 {
			fb.selected++
		}
		return true
	case e.Is("enter"):
		if len(fb.files) == 0 {
			return true
		}
		sel := fb.files[fb.selected]
		target := filepath.Join(fb.dir, sel.Name())
		if sel.IsDir() {
			fb.dir = target
			fb.reload()
		} else {
			fb.app.OpenFile(target)
		}
		return true
	}
	return false
}

func (fb *fileBrowserWidget) HandleMouse(e loom.MouseEvent) bool {
	if e.Action == loom.MousePress && e.Button == loom.MouseLeft && e.Y > 0 {
		idx := e.Y - 1
		if idx >= 0 && idx < len(fb.files) {
			fb.selected = idx
			sel := fb.files[fb.selected]
			if !sel.IsDir() {
				fb.app.OpenFile(filepath.Join(fb.dir, sel.Name()))
			}
			return true
		}
	}
	return false
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
	t.input.Prompt = "$ "
	return t
}

func (tw *terminalWidget) Draw(c *loom.Canvas, r loom.Rect) {
	c.PaintSurface(r, loom.Style{})
	c.Write(r.X, r.Y, "Terminal Output", loom.Style{Bold: true, Dim: true})
	maxLogs := r.H - 2
	if maxLogs < 0 {
		maxLogs = 0
	}
	startLog := max(0, len(tw.logs)-maxLogs)
	visibleLogs := tw.logs[startLog:]
	for i, logLine := range visibleLogs {
		c.Write(r.X, r.Y+1+i, logLine, loom.Style{FG: loom.ColorIndex(2)}) // Green logs
	}

	if r.H > 1 {
		inputRect := loom.Rect{X: r.X, Y: r.Y + r.H - 1, W: r.W, H: 1}
		focused := tw.app.activeFocus == focusTerminal
		tw.input.Draw(c, inputRect, focused)
	}
}

func (tw *terminalWidget) HandleKey(e loom.KeyEvent) bool {
	if e.Is("enter") {
		cmd := tw.input.Value()
		tw.input.SetValue("")
		if strings.TrimSpace(cmd) != "" {
			tw.execCommand(strings.TrimSpace(cmd))
		}
		return true
	}
	return tw.input.HandleKey(e)
}

func (tw *terminalWidget) HandleMouse(e loom.MouseEvent) bool { return false }

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

// NewApp initializes and returns a new TextEditApp.
func NewApp() *TextEditApp {
	app := &TextEditApp{
		activePath:  "demo.go",
		mruList:     []string{"demo.go", "README.md", "go.mod"},
		showBrowser: true,
		activeFocus: focusEditor,
		statusMsg:   "Ready. C-s: Save • C-S-s: SaveAs • C-o: Open • C-c/C-v/C-x: Clip • C-b: Sidebar",
	}

	initialContent := `package main

import "fmt"

// Welcome to Loom Text Editor
func main() {
	fmt.Println("Hello, Loom!")
}`
	app.editor = loom.NewTextArea(initialContent)
	app.editorW = &editorWidget{app: app}
	app.fileBrowser = newFileBrowser(".", app)
	app.terminal = newTerminal(app)

	app.vSplit = loom.NewSplit(app.editorW, app.terminal)
	app.vSplit.Orientation = loom.Vertical
	app.vSplit.Ratio = 0.7
	app.vSplit.MinFirst = 5
	app.vSplit.MinSecond = 3

	app.hSplit = loom.NewSplit(app.fileBrowser, app.vSplit)
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
			{ID: "quit", Action: "quit", Key: "q"},
		},
	}

	return app
}

// OpenFile opens path into the editor buffer and updates MRU list.
func (app *TextEditApp) OpenFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		app.statusMsg = "Error reading " + path + ": " + err.Error()
		return
	}
	app.activePath = path
	app.editor.SetValue(string(data))
	app.modified = false
	app.touchMRU(path)
	app.statusMsg = "Opened " + path
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
	app.modified = false
	app.touchMRU(path)
	app.statusMsg = "Saved " + path
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

func (app *TextEditApp) Draw(c *loom.Canvas, r loom.Rect) {
	modStr := ""
	if app.modified {
		modStr = " *"
	}
	mruStr := strings.Join(app.mruList, ", ")
	if len(mruStr) > 40 {
		mruStr = mruStr[:37] + "..."
	}

	focusName := "Editor"
	switch app.activeFocus {
	case focusBrowser:
		focusName = "Sidebar"
	case focusTerminal:
		focusName = "Terminal"
	}

	app.frame.Boxes[0].Title = fmt.Sprintf("File: %s%s  | MRU: [%s]  | Focus: %s", app.activePath, modStr, mruStr, focusName)
	app.frame.Status = app.statusMsg
	app.frame.Draw(c, r)
}

func (app *TextEditApp) applySyntaxHighlighting(c *loom.Canvas, r loom.Rect) {
	// Re-style keywords/comments/strings on top of the rendered editor area
	val := app.editor.Value()
	lines := strings.Split(val, "\n")
	keywords := map[string]bool{
		"package": true, "import": true, "func": true, "return": true,
		"if": true, "else": true, "for": true, "range": true, "var": true,
		"type": true, "struct": true, "const": true, "def": true, "class": true,
	}

	for lineIdx, line := range lines {
		y := r.Y + lineIdx
		if y >= r.Y+r.H {
			break
		}
		words := strings.Fields(line)
		for _, w := range words {
			if keywords[w] {
				// Highlight keyword
				idx := strings.Index(line, w)
				if idx >= 0 && r.X+idx < r.X+r.W {
					c.Write(r.X+idx, y, w, loom.Style{Bold: true, FG: loom.ColorIndex(4)}) // Cyan bold keyword
				}
			}
		}
		// Comment highlight
		if cIdx := strings.Index(line, "//"); cIdx >= 0 {
			if r.X+cIdx < r.X+r.W {
				c.Write(r.X+cIdx, y, line[cIdx:], loom.Style{FG: loom.ColorIndex(8), Dim: true}) // Gray comment
			}
		}
	}
}

func (app *TextEditApp) HandleKey(e loom.KeyEvent) bool {
	// Global Keybindings
	switch {
	case e.Is("ctrl-s", "ctrl-S"):
		app.SaveFile(app.activePath)
		return true
	case e.Is("ctrl-shift-s"):
		saveAsPath := app.activePath + ".bak"
		app.SaveFile(saveAsPath)
		return true
	case e.Is("ctrl-o"):
		// Open next file from MRU list
		if len(app.mruList) > 1 {
			nextPath := app.mruList[1]
			app.OpenFile(nextPath)
		} else {
			app.statusMsg = "MRU list empty"
		}
		return true
	case e.Is("ctrl-c"):
		val := app.editor.Value()
		lines := strings.Split(val, "\n")
		row, _ := app.editor.Caret()
		if row >= 0 && row < len(lines) {
			app.clipboard = lines[row]
			app.statusMsg = "Copied line to clipboard"
		}
		return true
	case e.Is("ctrl-v"):
		if app.clipboard != "" {
			for _, ch := range app.clipboard {
				app.editor.HandleKey(loom.KeyEvent{Text: string(ch)})
			}
			app.modified = true
			app.statusMsg = "Pasted from clipboard"
		}
		return true
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
		return true
	case e.Is("ctrl-b"):
		app.showBrowser = !app.showBrowser
		if app.showBrowser {
			app.hSplit.SetRatio(0.25)
			app.statusMsg = "Sidebar opened"
		} else {
			app.hSplit.SetRatio(0.0)
			app.statusMsg = "Sidebar collapsed"
		}
		return true
	case e.Is("tab", "f6"):
		app.activeFocus = (app.activeFocus + 1) % 3
		app.statusMsg = fmt.Sprintf("Switched focus to %v", app.activeFocus)
		return true
	}

	// Dispatch to active focus area
	switch app.activeFocus {
	case focusBrowser:
		if fbHandled := app.fileBrowser.HandleKey(e); fbHandled {
			return true
		}
	case focusTerminal:
		if termHandled := app.terminal.HandleKey(e); termHandled {
			return true
		}
	case focusEditor:
		if edHandled := app.editorW.HandleKey(e); edHandled {
			return true
		}
	}

	return app.frame.HandleKey(e)
}

func (app *TextEditApp) HandleMouse(e loom.MouseEvent) bool {
	return app.frame.HandleMouse(e)
}

// PaneRequest declares terminal requirements.
func (app *TextEditApp) PaneRequest() loom.PaneRequest {
	return loom.PaneRequest{
		Mouse:      1003,
		Resizeable: true,
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
	pane.EnableMouse()
	return pane.Run(app)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
