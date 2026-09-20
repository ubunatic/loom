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
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"codeberg.org/ubunatic/loom"
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
	dir              string
	files            []os.DirEntry
	selected, offset int
	lines            []string
	kind             Kind
	metadata         string
}

// New opens dir and returns an interactive viewer widget.
func New(dir string) (*browser, error) { return newBrowser(dir) }

func newBrowser(dir string) (*browser, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("ansiviewer: read %s: %w", dir, err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	b := &browser{dir: dir, files: entries}
	if len(entries) > 0 {
		b.selectFile(0)
	}
	return b, nil
}

func (b *browser) selectFile(i int) {
	if i < 0 || i >= len(b.files) {
		return
	}
	b.selected = i
	b.offset = 0
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
	left := r.W / 3
	if left < 12 {
		left = 12
	}
	if left >= r.W {
		left = r.W - 1
	}
	if left < 1 {
		left = 1
	}
	c.Write(r.X, r.Y, "Files", loom.Style{Bold: true})
	c.Write(r.X+left, r.Y, "Preview", loom.Style{Bold: true})
	for i, f := range b.files {
		if i+1 >= r.H {
			break
		}
		s := "  " + f.Name()
		if i == b.selected {
			s = "> " + f.Name()
		}
		c.Write(r.X, r.Y+i+1, s, loom.Style{})
	}
	view := r.H - 1
	if view < 1 {
		return
	}
	for i := 0; i < view && b.offset+i < len(b.lines); i++ {
		writeANSI(c, loom.Rect{X: r.X + left + 1, Y: r.Y + i + 1, W: r.W - left - 1, H: 1}, b.lines[b.offset+i])
	}
}

func (b *browser) HandleKey(e loom.KeyEvent) bool {
	if e.Is("q", "esc", "ctrl-c") {
		return true
	}
	switch {
	case e.Is("up"):
		b.selectFile(b.selected - 1)
	case e.Is("down"):
		b.selectFile(b.selected + 1)
	case e.Is("pgup"):
		b.offset -= 10
		if b.offset < 0 {
			b.offset = 0
		}
	case e.Is("pgdn"):
		b.offset += 10
		if b.offset >= len(b.lines) {
			b.offset = max(0, len(b.lines)-1)
		}
	case e.Is("j"):
		b.offset++
		if b.offset >= len(b.lines) {
			b.offset = max(0, len(b.lines)-1)
		}
	case e.Is("k"):
		if b.offset > 0 {
			b.offset--
		}
	case e.Is("enter") && len(b.files) > 0 && b.files[b.selected].IsDir():
		if next, err := newBrowser(filepath.Join(b.dir, b.files[b.selected].Name())); err == nil {
			*b = *next
		}
	}
	return false
}
func (b *browser) HandleMouse(loom.MouseEvent) bool { return false }
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func writeANSI(c *loom.Canvas, r loom.Rect, input string) {
	style := loom.Style{}
	x, y := r.X, r.Y
	rs := []rune(input)
	for i := 0; i < len(rs) && y < r.Y+r.H; {
		if rs[i] == '\x1b' && i+1 < len(rs) && rs[i+1] == '[' {
			j := i + 2
			for j < len(rs) && rs[j] != 'm' {
				j++
			}
			if j < len(rs) {
				style = applySGR(style, string(rs[i+2:j]))
				i = j + 1
				continue
			}
		}
		if rs[i] == '\n' {
			y++
			x = r.X
			i++
			continue
		}
		w := loom.StringWidth(string(rs[i]))
		if w < 1 {
			i++
			continue
		}
		if x+w <= r.X+r.W {
			x += c.Write(x, y, string(rs[i]), style)
		}
		i++
	}
}
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
		case n == 39:
			style.FG = loom.ColorReset()
		case n == 38 && i+2 < len(parts) && parts[i+1] == "5":
			v, _ := strconv.Atoi(parts[i+2])
			style.FG = loom.ColorIndex(uint8(v))
			i += 2
		case n == 38 && i+4 < len(parts) && parts[i+1] == "2":
			r, _ := strconv.Atoi(parts[i+2])
			g, _ := strconv.Atoi(parts[i+3])
			b, _ := strconv.Atoi(parts[i+4])
			style.FG = loom.ColorRGB(uint8(r), uint8(g), uint8(b))
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
	return loom.RenderTo(out, b, cols, rows)
}

// Run starts the interactive viewer rooted at the optional directory argument.
func Run(args []string) error {
	fs := flag.NewFlagSet("ansiviewer", flag.ContinueOnError)
	record := fs.Duration("record", 0, "capture one snapshot after this delay")
	recordOut := fs.String("record-out", "ansiviewer.ansi", "snapshot output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}
	if *record > 0 {
		file, err := os.Create(*recordOut)
		if err != nil {
			return fmt.Errorf("ansiviewer: create recording: %w", err)
		}
		defer file.Close()
		return Record(context.Background(), file, dir, *record, 100, 30)
	}
	b, err := New(dir)
	if err != nil {
		return err
	}
	p, err := loom.New(1 << 16)
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Run(b)
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
