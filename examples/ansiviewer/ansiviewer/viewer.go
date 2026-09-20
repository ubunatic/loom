// Package ansiviewer is a small ANSI-aware directory browser example.
package ansiviewer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"codeberg.org/ubunatic/loom"
)

type browser struct {
	dir      string
	files    []os.DirEntry
	selected int
	text     []string
}

// New opens dir and returns an interactive viewer widget.
func New(dir string) (*browser, error) { return newBrowser(dir) }

func newBrowser(dir string) (*browser, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("ansiviewer: read %s: %w", dir, err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	return &browser{dir: dir, files: entries}, nil
}

func (b *browser) selectFile(i int) {
	if i < 0 || i >= len(b.files) {
		return
	}
	b.selected = i
	data, err := os.ReadFile(filepath.Join(b.dir, b.files[i].Name()))
	if err != nil {
		b.text = []string{"read error: " + err.Error()}
		return
	}
	b.text = strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
}

func (b *browser) Draw(c *loom.Canvas, r loom.Rect) {
	left := r.W / 3
	if left < 12 {
		left = 12
	}
	c.Write(r.X, r.Y, "Files", loom.Style{Bold: true})
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
	for y, line := range b.text {
		if y >= r.H {
			break
		}
		writeANSI(c, loom.Rect{X: r.X + left + 1, Y: r.Y + y, W: r.W - left - 1, H: r.H - y}, line)
	}
}

func (b *browser) HandleKey(e loom.KeyEvent) bool {
	if e.Is("q", "esc", "ctrl-c") {
		return true
	}
	if e.Is("up") && b.selected > 0 {
		b.selectFile(b.selected - 1)
	}
	if e.Is("down") && b.selected+1 < len(b.files) {
		b.selectFile(b.selected + 1)
	}
	if e.Is("enter") && len(b.files) > 0 && b.files[b.selected].IsDir() {
		if next, err := newBrowser(filepath.Join(b.dir, b.files[b.selected].Name())); err == nil {
			*b = *next
		}
	}
	return false
}
func (b *browser) HandleMouse(loom.MouseEvent) bool { return false }

func writeANSI(c *loom.Canvas, r loom.Rect, input string) {
	style := loom.Style{}
	x := r.X
	y := r.Y
	for i := 0; i < len(input) && y < r.Y+r.H; {
		if input[i] == '\x1b' && i+1 < len(input) && input[i+1] == '[' {
			j := i + 2
			for j < len(input) && input[j] != 'm' {
				j++
			}
			if j < len(input) {
				if strings.Contains(input[i+2:j], "31") {
					style.FG = loom.ColorIndex(1)
				}
				if input[i+2:j] == "0" {
					style = loom.Style{}
				}
				i = j + 1
				continue
			}
		}
		if input[i] == '\n' {
			y++
			x = r.X
			i++
			continue
		}
		if x < r.X+r.W {
			x += c.Write(x, y, string(input[i]), style)
		}
		i++
	}
}
