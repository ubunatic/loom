// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package ansiviewer is a small ANSI-aware directory browser example.
package ansiviewer

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"ubunatic.com/loom"
	"ubunatic.com/loom/examples/filebrowser/filebrowser"
)

// Kind describes the content presentation selected for a path.
type Kind string

const (
	KindText   Kind = "text"
	KindANSI   Kind = "ansi"
	KindBinary Kind = "binary"
	KindImage  Kind = "image"
)

// Classify identifies a file using its extension and a small content sniff.
func Classify(path string) Kind {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".ansi", ".ans":
		return KindANSI
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp":
		return KindImage
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return KindText
	}
	for _, v := range b {
		if v == 0 {
			return KindBinary
		}
	}
	if !utf8.Valid(b) {
		return KindBinary
	}
	return KindText
}

type browser struct {
	dir         string
	files       []os.DirEntry
	selected    int
	offset      int
	offsetX     int
	lines       []string
	kind        Kind
	ansiBuf     *loom.AnsiBuffer
	ruler       bool
	metadata    string
	previewView *loom.View
	navigation  *filebrowser.NavigationPane
	quit        bool
	focused     bool
}

func (b *browser) Focused() bool         { return b.focused }
func (b *browser) SetFocus(focused bool) { b.focused = focused }

type framedBrowser struct {
	frame   *loom.Frame
	astra   *astraToggle
	browser *browser
}

type astraToggle struct {
	enabled bool
}

func (a *astraToggle) BackgroundInterval() time.Duration {
	return loom.SpeccedBackground.RedrawInterval
}

func (a *astraToggle) DrawBackground(c *loom.Canvas, r loom.Rect) {
	a.DrawBackgroundAt(c, r, time.Now())
}

func (a *astraToggle) DrawBackgroundAt(c *loom.Canvas, r loom.Rect, now time.Time) {
	if a.enabled {
		loom.NewAstraBackground().DrawBackgroundAt(c, r, now)
	}
}

// filesMinWidth keeps short file names readable in narrow terminals.
const filesMinWidth = 16

func newFramedBrowser(b *browser, astra *astraToggle) *framedBrowser {
	theme := loom.Theme("julia256")
	border := loom.BoxBorder{
		TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
		Horizontal: "─", Vertical: "│", TitlePrefix: " ", TitleSuffix: " ",
	}
	frame := &loom.Frame{
		Gap:    1,
		Title:  "ANSI Viewer",
		Status: "↑↓ select  •  preview: hjkl pan/arrows, PgUp/PgDn  •  / filter  •  Enter open  •  Esc back  •  Tab preview  •  r ruler  •  a Astra  •  F10 Quit",
		Boxes: []loom.Box{
			{ID: "files", Title: "Files", FillHeight: true, MinWidth: filesMinWidth, Width: filesMinWidth, Height: 4, Border: border, Child: b.navigation},
			{ID: "viewer", Title: "Preview", FillHeight: true, Dynamic: true, MinWidth: 30, Height: 4, Border: border, Child: b},
		},
		Actions: []loom.FrameAction{
			{ID: "quit_q", Action: "quit", Key: "q"},
			{ID: "quit_ctrl_c", Action: "quit", Key: "ctrl-c"},
		},
		Style: theme.FrameStyle(),
	}
	frame.Boxes[0].Style = theme.BoxStyle()
	frame.Boxes[1].Style = theme.BoxStyle()
	if b.navigation != nil {
		b.navigation.ApplyTheme(theme)
	}
	return &framedBrowser{frame: frame, astra: astra, browser: b}
}

func (b *framedBrowser) Draw(c *loom.Canvas, r loom.Rect) {
	b.frame.Boxes[0].Title = "Files"
	b.frame.Boxes[1].Title = "Preview"
	// The file list takes a quarter of the width; the preview gets the rest.
	files := max(filesMinWidth, r.W/4)
	b.frame.Boxes[0].Width, b.frame.Boxes[0].MaxWidth = files, files
	if focused := b.frame.FocusedBox(); focused != nil {
		if focused.ID == "files" {
			b.frame.Boxes[0].Title = "▶ Files"
		} else {
			b.frame.Boxes[1].Title = "▶ Preview"
		}
	}
	b.frame.Draw(c, r)
}

func (b *framedBrowser) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if e.Is("r") {
		b.browser.ruler = !b.browser.ruler
		return loom.Handled()
	}
	if e.Is("a") && b.astra != nil {
		b.astra.enabled = !b.astra.enabled
		return loom.Handled()
	}
	if b.browser.navigation.Searching() || e.Is("/", "esc", "backspace") {
		result := b.browser.navigation.ConsumeKey(e)
		b.browser.syncSelection()
		if result.Quit || b.browser.quit {
			return loom.QuitResult()
		}
		if result.Consumed {
			return result
		}
	}
	result := b.frame.ConsumeKey(e)
	b.browser.syncSelection()
	if result.Quit || b.browser.quit {
		return loom.QuitResult()
	}
	return result
}

func (b *framedBrowser) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	result := b.frame.ConsumeMouse(e)
	b.browser.syncSelection()
	if result.Quit || b.browser.quit {
		return loom.QuitResult()
	}
	return result
}

// New opens dir and returns an interactive viewer widget.
func New(dir string) (*browser, error) { return newBrowser(dir) }

func newBrowser(dir string) (*browser, error) {
	return newBrowserSelection(dir, "")
}

func newBrowserSelection(dir, selectName string) (*browser, error) {
	b := &browser{dir: dir}
	var nav *filebrowser.NavigationPane
	nav, err := filebrowser.NewNavigationPane(dir, filebrowser.NavigationPaneOptions{
		OnSelection: func(entry loom.FileEntry) {
			b.selectPath(entry.Path)
		},
		OnOpen: func(directory loom.Directory) {
			b.dir = directory.Path
			b.syncSelection()
		},
		OnQuit: func() {
			b.quit = true
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ansiviewer: read %s: %w", dir, err)
	}
	nav.SetRoot("")
	b.navigation = nav
	b.dir = nav.Directory().Path
	if selectName != "" {
		for i, item := range nav.List().Items {
			if item.Name == selectName {
				nav.List().SelectIndex(i)
				break
			}
		}
	}
	b.syncSelection()
	return b, nil
}

func (b *browser) syncSelection() {
	if b.navigation == nil {
		return
	}
	entry, ok := b.navigation.Selected()
	if !ok {
		return
	}
	b.dir = b.navigation.Directory().Path
	b.selectPath(entry.Path)
}

func (b *browser) selectPath(path string) {
	if b.navigation != nil {
		b.dir = b.navigation.Directory().Path
	}
	entries, err := os.ReadDir(b.dir)
	if err != nil {
		return
	}
	b.files = entries
	for i, entry := range entries {
		if filepath.Join(b.dir, entry.Name()) == path {
			b.selectFile(i)
			return
		}
	}
	if filepath.Clean(path) == filepath.Clean(b.dir) || path == filepath.Dir(filepath.Clean(b.dir)) {
		b.lines, b.metadata, b.kind = []string{"directory", "Press Enter to open"}, "", KindText
		return
	}
	b.lines, b.metadata, b.kind = nil, "", KindText
}

func (b *browser) selectFile(i int) {
	if i < 0 || i >= len(b.files) {
		return
	}
	b.selected = i
	b.offset = 0
	b.offsetX = 0
	path := filepath.Join(b.dir, b.files[i].Name())
	info, err := b.files[i].Info()
	if err != nil {
		b.metadata = err.Error()
		return
	}
	b.kind = KindText
	b.lines = nil
	b.ansiBuf = nil
	b.metadata = fmt.Sprintf("%s  %d bytes  %s", b.files[i].Name(), info.Size(), Classify(path))
	if b.files[i].IsDir() {
		b.lines = []string{"directory", "Press Enter to open"}
		return
	}
	b.kind = Classify(path)
	if b.kind == KindBinary || b.kind == KindImage {
		b.lines = []string{"metadata only", b.metadata}
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		b.lines = []string{"read error: " + err.Error()}
		return
	}
	b.lines = strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if b.kind == KindANSI {
		b.ansiBuf, err = loom.ParseAnsiBuffer(string(data), 1, 1)
		if err != nil {
			b.lines = []string{"parse error: " + err.Error()}
			return
		}
	}
}

func (b *browser) Draw(c *loom.Canvas, r loom.Rect) {
	if b.ruler && r.W > 2 && r.H > 2 {
		loom.DrawRuler(c, r)
		r = loom.Rect{X: r.X + 1, Y: r.Y + 1, W: r.W - 2, H: r.H - 2}
	}
	if r.H < 1 || r.W < 1 {
		return
	}
	if b.previewView == nil {
		b.previewView = loom.NewView(b.lines)
	}
	b.previewView.Lines = b.lines
	b.previewView.Scroll = b.offset
	b.previewView.OffsetX = b.offsetX
	barCanvas := loom.NewCanvas(r.W, r.H)
	b.previewView.Draw(barCanvas, loom.Rect{W: r.W, H: r.H})
	b.offset = b.previewView.Scroll
	b.offsetX = b.previewView.OffsetX
	barVisible := false
	for y := 0; y < r.H; y++ {
		cell := barCanvas.Get(r.W-1, y)
		if (cell.Text == loom.SpeccedDefaults.Scrollbar.ForegroundChar && cell.Style == b.previewView.Scrollbar.Thumb) ||
			(cell.Text == loom.SpeccedDefaults.Scrollbar.BackgroundChar && cell.Style == b.previewView.Scrollbar.Track) {
			barVisible = true
			break
		}
	}
	contentRect := r
	if barVisible {
		contentRect.W--
	}
	if b.kind == KindANSI && b.ansiBuf != nil {
		for y := 0; y < contentRect.H; y++ {
			row := b.offset + y
			if row >= b.ansiBuf.Rows() {
				break
			}
			for x := 0; x < contentRect.W; x++ {
				col := b.offsetX + x
				if col >= b.ansiBuf.Cols() {
					break
				}
				c.Set(contentRect.X+x, contentRect.Y+y, b.ansiBuf.Get(col, row).ToCell())
			}
		}
	} else {
		for y := 0; y < contentRect.H; y++ {
			for x := 0; x < contentRect.W; x++ {
				c.Set(contentRect.X+x, contentRect.Y+y, barCanvas.Get(x, y))
			}
		}
	}
	if barVisible {
		for y := 0; y < r.H; y++ {
			c.Set(r.X+r.W-1, r.Y+y, barCanvas.Get(r.W-1, y))
		}
	}
}

func (b *browser) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if e.Is("q", "ctrl-c") {
		return loom.QuitResult()
	}
	if e.Is("esc", "backspace") {
		result := b.navigation.ConsumeKey(e)
		b.syncSelection()
		return result
	}
	if b.previewView == nil {
		b.previewView = loom.NewView(b.lines)
	}
	b.previewView.Lines = b.lines
	b.previewView.Scroll = b.offset
	b.previewView.OffsetX = b.offsetX
	result := b.previewView.ConsumeKey(e)
	b.offset = b.previewView.Scroll
	b.offsetX = b.previewView.OffsetX
	return result
}
func (b *browser) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	if b.previewView == nil {
		return loom.Ignored()
	}
	result := b.previewView.ConsumeMouse(e)
	b.offset = b.previewView.Scroll
	b.offsetX = b.previewView.OffsetX
	return result
}

// Render writes one deterministic headless frame from the real widget.
func Render(out io.Writer, dir string, cols, rows int) error {
	b, err := New(dir)
	if err != nil {
		return err
	}
	return loom.RenderTo(out, newFramedBrowser(b, nil), cols, rows)
}

// Run starts the interactive viewer rooted at the optional directory argument.
func Run(args []string) error { return run(args, os.Stdout) }

func run(args []string, output io.Writer) error {
	fs := flag.NewFlagSet("ansiviewer", flag.ContinueOnError)
	record := fs.Duration("record", 0, "capture one snapshot after this delay")
	recordOut := fs.String("record-out", "ansiviewer.ansi", "snapshot output path")
	fs.StringVar(recordOut, "o", "ansiviewer.ansi", "snapshot output path, or - for stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := "."
	selectName := ""
	if fs.NArg() > 0 && *record == 0 {
		dir = fs.Arg(0)
		info, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("ansiviewer: stat %s: %w", dir, err)
		}
		if !info.IsDir() {
			selectName = filepath.Base(dir)
			dir = filepath.Dir(dir)
		}
	}
	if *record > 0 {
		var out io.Writer = output
		var file *os.File
		var err error
		if *recordOut != "-" {
			file, err = os.Create(*recordOut)
			if err != nil {
				return fmt.Errorf("ansiviewer: create recording: %w", err)
			}
			defer file.Close()
			out = file
		}
		if fs.NArg() > 0 {
			return RecordCommand(context.Background(), out, *record, fs.Arg(0), fs.Args()[1:]...)
		}
		return Record(context.Background(), out, dir, *record, 100, 30)
	}
	b, err := newBrowserSelection(dir, selectName)
	if err != nil {
		return err
	}
	p, err := loom.New(1 << 16)
	if err != nil {
		return err
	}
	defer p.Close()
	p.Resizeable = true
	p.MaxCols = 0
	p.DisableDefaultQuit = true
	p.ResizeConfig.FullScreenBuffer = true
	astra := &astraToggle{}
	p.Background = astra
	view := newFramedBrowser(b, astra)
	return p.Run(view)
}

// Record waits delay, renders exactly one frame of the real viewer widget, and
// returns. The caller owns out; no child process or background goroutine leaks.
func Record(ctx context.Context, out io.Writer, dir string, delay time.Duration, cols, rows int) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}
	return Render(out, dir, cols, rows)
}

