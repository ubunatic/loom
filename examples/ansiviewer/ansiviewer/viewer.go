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
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/examples/filebrowser/filebrowser"
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

func (b *framedBrowser) HandleKey(e loom.KeyEvent) bool {
	if e.Is("r") {
		b.browser.ruler = !b.browser.ruler
		return false
	}
	if e.Is("a") && b.astra != nil {
		b.astra.enabled = !b.astra.enabled
		return false
	}
	quit := b.frame.HandleKey(e)
	b.browser.syncSelection()
	return quit || b.browser.quit
}

func (b *framedBrowser) ConsumeKey(e loom.KeyEvent) (quit, consumed bool) {
	quit, consumed = b.browser.navigation.ConsumeKey(e)
	if consumed {
		b.browser.syncSelection()
		return quit || b.browser.quit, true
	}
	return false, false
}

func (b *framedBrowser) HandleMouse(e loom.MouseEvent) bool {
	quit := b.frame.HandleMouse(e)
	b.browser.syncSelection()
	return quit || b.browser.quit
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
}

func (b *browser) Draw(c *loom.Canvas, r loom.Rect) {
	if b.ruler && r.W > 2 && r.H > 2 {
		drawRuler(c, r)
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
	start := min(b.offset, len(b.lines))
	end := min(len(b.lines), b.offset+r.H)
	visibleLines := b.lines[start:end]
	if b.kind == KindANSI {
		writeANSI(c, contentRect, strings.Join(visibleLines, "\n"), b.offsetX)
	} else {
		for i, line := range visibleLines {
			writeANSI(c, loom.Rect{X: contentRect.X, Y: contentRect.Y + i, W: contentRect.W, H: 1}, line, b.offsetX)
		}
	}
	if barVisible {
		for y := 0; y < r.H; y++ {
			c.Set(r.X+r.W-1, r.Y+y, barCanvas.Get(r.W-1, y))
		}
	}
}

func (b *browser) HandleKey(e loom.KeyEvent) bool {
	if e.Is("q", "ctrl-c") {
		return true
	}
	if e.Is("esc", "backspace") {
		quit := b.navigation.HandleKey(e)
		b.syncSelection()
		return quit
	}
	if b.previewView == nil {
		b.previewView = loom.NewView(b.lines)
	}
	b.previewView.Lines = b.lines
	b.previewView.Scroll = b.offset
	b.previewView.OffsetX = b.offsetX
	if b.previewView.HandleKey(e) {
		return true
	}
	b.offset = b.previewView.Scroll
	b.offsetX = b.previewView.OffsetX
	return false
}
func (b *browser) HandleMouse(e loom.MouseEvent) bool {
	if b.previewView == nil {
		return false
	}
	quit := b.previewView.HandleMouse(e)
	b.offset = b.previewView.Scroll
	b.offsetX = b.previewView.OffsetX
	return quit
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func writeANSI(c *loom.Canvas, r loom.Rect, input string, offsetX ...int) {
	baseStyle := initialANSIStyle(input)
	style := baseStyle
	claimANSIArea(c, r, baseStyle)
	horizontalOffset := 0
	if len(offsetX) > 0 {
		horizontalOffset = max(0, offsetX[0])
	}
	x, y := r.X-horizontalOffset, r.Y
	rs := []rune(input)
	for i := 0; i < len(rs) && y < r.Y+r.H; {
		if rs[i] == '\x1b' && i+1 < len(rs) && rs[i+1] == '[' {
			j := i + 2
			for j < len(rs) && (rs[j] < '@' || rs[j] > '~') {
				j++
			}
			if j < len(rs) {
				params := string(rs[i+2 : j])
				switch rs[j] {
				case 'm':
					style = applySGR(style, params)
				case 'C':
					x += csiNumber(params, 1)
				case 'H', 'f':
					row, col := csiPosition(params)
					y = r.Y + row - 1
					x = r.X + col - 1
				case 'G':
					x = r.X + csiNumber(params, 1) - 1
				case 'd':
					y = r.Y + csiNumber(params, 1) - 1
				case 'J':
					if params == "2" || params == "3" {
						c.ClearRect(r)
						claimANSIArea(c, r, baseStyle)
					}
				}
				i = j + 1
				continue
			}
		}
		if rs[i] == '\x1b' && i+2 < len(rs) && rs[i+1] == '(' {
			i += 3
			continue
		}
		if rs[i] == '\r' {
			x = r.X
			i++
			continue
		}
		if rs[i] == '\n' {
			y++
			x = r.X
			i++
			continue
		}
		if rs[i] < 0x20 || rs[i] == 0x7f {
			i++
			continue
		}
		end := i + 1
		for end < len(rs) && rs[end] != '\x1b' && rs[end] != '\r' && rs[end] != '\n' && rs[end] >= 0x20 && rs[end] != 0x7f {
			end++
		}
		for _, cell := range loom.ParseANSI(string(rs[i:end])) {
			if cell.Continuation {
				continue
			}
			w := loom.StringWidth(cell.Text)
			if w < 1 {
				continue
			}
			if x >= r.X && x+w <= r.X+r.W {
				x += c.Write(x, y, cell.Text, style)
			} else {
				x += w
			}
		}
		i = end
	}
}

// claimANSIArea makes the replay canvas the sole painter for the preview
// rectangle, including cells that the stream leaves blank. The cells remain
// transparent to the container's already-painted surface because their
// styles use the reset background.
func claimANSIArea(c *loom.Canvas, r loom.Rect, style loom.Style) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			c.PaintForeground(x, y, loom.Cell{Style: style})
		}
	}
}

// initialANSIStyle returns the first style that establishes an explicit ANSI
// background. Cursor-positioning spaces before that point are terminal
// housekeeping, not the recording's panel surface.
func initialANSIStyle(input string) loom.Style {
	style := loom.Style{}
	rs := []rune(input)
	for i := 0; i < len(rs); i++ {
		if rs[i] == '\x1b' && i+1 < len(rs) && rs[i+1] == '[' {
			j := i + 2
			for j < len(rs) && (rs[j] < '@' || rs[j] > '~') {
				j++
			}
			if j < len(rs) && rs[j] == 'm' {
				params := string(rs[i+2 : j])
				style = applySGR(style, params)
				// mc first uses 40m while clearing terminal state; its
				// actual panel surface is the explicit 256-color background.
				if strings.Contains(params, "48;5;") || strings.Contains(params, "48;2;") {
					return style
				}
				i = j
				continue
			}
		}
	}
	return style
}

func csiNumber(params string, fallback int) int {
	n, err := strconv.Atoi(params)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func csiPosition(params string) (int, int) {
	parts := strings.Split(params, ";")
	row, col := 1, 1
	if len(parts) > 0 && parts[0] != "" {
		row = csiNumber(parts[0], 1)
	}
	if len(parts) > 1 && parts[1] != "" {
		col = csiNumber(parts[1], 1)
	}
	return row, col
}

// applySGR parses and applies SGR (Select Graphic Rendition) parameters to a style.
// Note: This function is a custom parser for backward compatibility. In loom 034+,
// this could be replaced with loom.ParseANSI for SGR handling, but the parent
// writeANSI() function handles cursor positioning (H, C, G, d) and character set
// designation (ESC () alongside SGR, requiring a refactor to separate these concerns.
// For now, applySGR remains local to maintain the existing rectangle-bounded
// rendering with integrated cursor replay.
func applySGR(style loom.Style, params string) loom.Style {
	parts := strings.Split(params, ";")
	for i := 0; i < len(parts); i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			continue
		}
		switch {
		case n == 0:
			style = loom.Style{}
		case n == 1:
			style.Bold = true
		case n == 4:
			style.Underline = true
		case n >= 30 && n <= 37:
			style.FG = loom.ColorIndex(uint8(n - 30))
		case n >= 90 && n <= 97:
			style.FG = loom.ColorIndex(uint8(n - 90 + 8))
		case n >= 40 && n <= 47:
			style.BG = loom.ColorIndex(uint8(n - 40))
		case n >= 100 && n <= 107:
			style.BG = loom.ColorIndex(uint8(n - 100 + 8))
		case n == 39:
			style.FG = loom.ColorReset()
		case n == 49:
			style.BG = loom.ColorReset()
		case n == 38 && i+2 < len(parts) && parts[i+1] == "5":
			v, _ := strconv.Atoi(parts[i+2])
			style.FG = loom.ColorIndex(uint8(v))
			i += 2
		case n == 48 && i+2 < len(parts) && parts[i+1] == "5":
			v, _ := strconv.Atoi(parts[i+2])
			style.BG = loom.ColorIndex(uint8(v))
			i += 2
		case n == 38 && i+4 < len(parts) && parts[i+1] == "2":
			r, _ := strconv.Atoi(parts[i+2])
			g, _ := strconv.Atoi(parts[i+3])
			b, _ := strconv.Atoi(parts[i+4])
			style.FG = loom.ColorRGB(uint8(r), uint8(g), uint8(b))
			i += 4
		case n == 48 && i+4 < len(parts) && parts[i+1] == "2":
			r, _ := strconv.Atoi(parts[i+2])
			g, _ := strconv.Atoi(parts[i+3])
			b, _ := strconv.Atoi(parts[i+4])
			style.BG = loom.ColorRGB(uint8(r), uint8(g), uint8(b))
			i += 4
		}
	}
	return style
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

// drawRuler frames r with a 1-cell ruler on the terminal's default background,
// so cells a snapshot leaves unpainted show through next to it.
func drawRuler(c *loom.Canvas, r loom.Rect) {
	dim := loom.Style{FG: loom.ColorIndex(244), BG: loom.ColorReset()}
	hot := loom.Style{FG: loom.ColorIndex(214), BG: loom.ColorReset()}
	// PaintForeground keeps the reset background instead of inheriting the box's.
	put := func(x, y int, text string, style loom.Style) {
		c.PaintForeground(x, y, loom.Cell{Text: text, Style: style})
	}
	for x := 0; x < r.W; x++ {
		text, style := "·", dim
		switch {
		case x == 0 || x == r.W-1:
			text = " "
		case x%10 == 0:
			text, style = strconv.Itoa(x/10%10), hot
		case x%5 == 0:
			text = "┊"
		}
		put(r.X+x, r.Y, text, style)
		put(r.X+x, r.Y+r.H-1, text, style)
	}
	for y := 1; y < r.H-1; y++ {
		style := dim
		if y%5 == 0 {
			style = hot
		}
		put(r.X, r.Y+y, strconv.Itoa(y%10), style)
		put(r.X+r.W-1, r.Y+y, "│", dim)
	}
}
