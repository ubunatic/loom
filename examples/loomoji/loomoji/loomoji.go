// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package loomoji demonstrates a compact inline emoji picker.
package loomoji

import (
	"fmt"
	"os"
	"strings"

	"codeberg.org/ubunatic/loom"
	"github.com/spf13/cobra"
)

type emoji struct {
	icon  string
	name  string
	group int
}

var emojiList = []emoji{
	{"😀", "grinning face", 0}, {"😃", "grinning face with big eyes", 0}, {"😄", "grinning face smiling eyes", 0}, {"😁", "beaming face", 0},
	{"😆", "squinting face", 0}, {"😅", "grinning face sweat", 0}, {"😂", "face tears joy", 0}, {"🤣", "rolling floor laughing", 0},
	{"😊", "smiling face smiling eyes", 0}, {"😇", "smiling face halo", 0}, {"🙂", "slightly smiling face", 0}, {"🙃", "upside down face", 0},
	{"😉", "winking face", 0}, {"😍", "smiling face heart eyes", 0}, {"🥰", "smiling face hearts", 0}, {"😘", "face blowing kiss", 0},
	{"😋", "face savoring food", 0}, {"😛", "face tongue", 0}, {"🤔", "thinking face", 0}, {"🤗", "hugging face", 0},
	{"😎", "smiling face sunglasses", 0}, {"🥳", "partying face", 0}, {"😭", "loudly crying face", 0}, {"😴", "sleeping face", 0},
	{"👋", "waving hand", 1}, {"👌", "ok hand", 1}, {"✌️", "victory hand", 1}, {"🤞", "crossed fingers", 1},
	{"👍", "thumbs up", 1}, {"👎", "thumbs down", 1}, {"👏", "clapping hands", 1}, {"🙌", "raising hands", 1},
	{"💪", "flexed biceps", 1}, {"🙏", "folded hands", 1}, {"💅", "nail polish", 1}, {"🤝", "handshake", 1},
	{"🐶", "dog face", 2}, {"🐱", "cat face", 2}, {"🐭", "mouse face", 2}, {"🐹", "hamster face", 2},
	{"🐰", "rabbit face", 2}, {"🦊", "fox", 2}, {"🐻", "bear", 2}, {"🐼", "panda", 2},
	{"🐸", "frog", 2}, {"🦁", "lion", 2}, {"🐵", "monkey face", 2}, {"🐧", "penguin", 2},
	{"🍎", "red apple", 3}, {"🍕", "pizza", 3}, {"🍔", "hamburger", 3}, {"🍟", "french fries", 3},
	{"🌮", "taco", 3}, {"🍩", "doughnut", 3}, {"🍪", "cookie", 3}, {"🍰", "shortcake", 3},
	{"⚽", "soccer ball", 4}, {"🏀", "basketball", 4}, {"🎮", "video game", 4}, {"🎸", "guitar", 4},
	{"🚀", "rocket", 5}, {"🚗", "automobile", 5}, {"✈️", "airplane", 5}, {"🚲", "bicycle", 5},
	{"❤️", "red heart", 6}, {"💛", "yellow heart", 6}, {"💚", "green heart", 6}, {"💙", "blue heart", 6},
	{"✨", "sparkles", 6}, {"🔥", "fire", 6}, {"⭐", "star", 6}, {"🎉", "party popper", 6},
}

var categories = []string{"☺", "♙", "🐾", "🍴", "⚽", "✈", "♡"}

const (
	emojiCellWidth   = 3
	emojiGridColumns = 10
)

type picker struct {
	query     *loom.TextInput
	group     int
	items     []int
	index     int
	focused   bool
	chosen    string
	width     int
	searchY   int
	gridTop   int
	categoryY int
	cols      int
	viewRows  int
	gridStart int
}

func newPicker() *picker {
	p := &picker{query: loom.NewTextInput(""), focused: true}
	p.refresh()
	return p
}

func (p *picker) refresh() {
	p.items = p.items[:0]
	q := strings.ToLower(p.query.Value())
	for i, item := range emojiList {
		if item.group == p.group && (q == "" || strings.Contains(item.name, q) || strings.Contains(item.icon, q)) {
			p.items = append(p.items, i)
		}
	}
	p.index = min(p.index, max(0, len(p.items)-1))
}

func (p *picker) Draw(c *loom.Canvas, r loom.Rect) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	background := loom.ColorRGB(38, 40, 43)
	panel := loom.ColorRGB(31, 33, 36)
	accent := loom.ColorRGB(32, 151, 185)
	c.PaintSurface(r, loom.Style{BG: background})
	// Keep a quiet row above the control, matching the reference layout.
	p.width = r.W
	p.searchY = r.Y + 1
	searchRect := loom.Rect{X: r.X, Y: p.searchY, W: r.W, H: min(2, max(0, r.Y+r.H-p.searchY))}
	c.PaintSurface(searchRect, loom.Style{BG: panel})
	if r.H >= 3 {
		c.PaintSurface(loom.Rect{X: r.X + 1, Y: p.searchY, W: max(0, r.W-2), H: 1}, loom.Style{BG: loom.ColorRGB(67, 69, 72)})
		c.Write(r.X+2, p.searchY, "⌕", loom.Style{FG: accent, Bold: true})
		query := p.query.Value()
		if query == "" {
			query = "search…"
		}
		c.Write(r.X+5, p.searchY, query, loom.Style{FG: loom.ColorRGB(220, 220, 220)})
		if p.focused {
			caret := min(p.query.Caret(), len([]rune(p.query.Value())))
			c.CursorX = r.X + 5 + loom.StringWidth(string([]rune(p.query.Value())[:caret]))
			c.CursorY = p.searchY
		}
	}

	p.categoryY = r.Y + r.H - 2
	if p.categoryY < p.searchY+2 {
		p.categoryY = p.searchY + 2
	}
	if p.categoryY < r.Y+r.H {
		c.PaintSurface(loom.Rect{X: r.X, Y: p.categoryY, W: r.W, H: min(1, r.Y+r.H-p.categoryY)}, loom.Style{BG: panel})
		for i, icon := range categories {
			x := r.X + 1 + i*4
			if x >= r.X+r.W {
				break
			}
			style := loom.Style{FG: loom.ColorRGB(210, 210, 210)}
			if i == p.group {
				style = loom.Style{FG: loom.ColorRGB(255, 255, 255), BG: accent, Bold: true}
			}
			c.Write(x, p.categoryY, icon, style)
		}
	}
	p.gridTop = p.searchY + 2
	gridBottom := p.categoryY
	p.cols = min(emojiGridColumns, max(1, (r.W-2)/emojiCellWidth))
	p.viewRows = max(0, gridBottom-p.gridTop)
	startRow := 0
	selectedRow := p.index / p.cols
	if selectedRow >= p.viewRows && p.viewRows > 0 {
		startRow = selectedRow - p.viewRows + 1
	}
	start := startRow * p.cols
	p.gridStart = start
	visible := min(len(p.items)-start, p.cols*p.viewRows)
	for n := 0; n < visible; n++ {
		idx := start + n
		x := r.X + 1 + (n%p.cols)*emojiCellWidth
		y := p.gridTop + n/p.cols
		style := loom.Style{FG: loom.ColorRGB(244, 203, 69)}
		if idx == p.index {
			style = loom.Style{FG: loom.ColorRGB(255, 255, 255), BG: accent, Bold: true}
		}
		c.Write(x, y, emojiList[p.items[idx]].icon, style)
	}
	if p.categoryY > p.gridTop {
		if len(p.items) == 0 {
			c.Write(r.X+2, p.gridTop, "No matching emoji", loom.Style{FG: loom.ColorRGB(170, 170, 170), Dim: true})
		}
	}
	if r.H > 0 {
		footer := "↑↓←→ move   Enter insert   1–7 category   Ctrl-C quit"
		if len(p.items) > 0 {
			item := emojiList[p.items[p.index]]
			footer = fmt.Sprintf("%s  %s   •   %d / %d", item.icon, item.name, p.index+1, len(p.items))
		}
		c.Write(r.X+1, r.Y+r.H-1, footer, loom.Style{FG: loom.ColorRGB(175, 178, 181), Dim: true})
	}
}

func (p *picker) HandleKey(e loom.KeyEvent) bool {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	switch key {
	case "ctrl-c", "q":
		return true
	case "esc":
		if p.query.Value() != "" {
			p.query.SetValue("")
			p.refresh()
		} else {
			return true
		}
	case "left":
		if p.focused {
			p.query.HandleKey(e)
			p.refresh()
		} else if p.index > 0 {
			p.index--
		}
	case "right":
		if p.focused {
			p.query.HandleKey(e)
			p.refresh()
		} else if p.index+1 < len(p.items) {
			p.index++
		}
	case "up":
		p.focused = false
		p.index = max(0, p.index-p.cols)
	case "down":
		p.focused = false
		p.index = min(max(0, len(p.items)-1), p.index+p.cols)
	case "enter":
		if len(p.items) > 0 {
			p.chosen = emojiList[p.items[p.index]].icon
			return true
		}
	case "backspace":
		if p.focused {
			p.query.HandleKey(e)
			p.refresh()
		}
	default:
		if len(key) == 1 && key[0] >= '1' && key[0] <= '7' {
			p.group = int(key[0] - '1')
			p.index = 0
			p.refresh()
			return false
		}
		if e.Text != "" && p.focused {
			p.query.HandleKey(e)
			p.index = 0
			p.refresh()
		}
	}
	return false
}

func (p *picker) HandleMouse(e loom.MouseEvent) bool {
	if e.Action != loom.MousePress || e.Button != loom.MouseLeft {
		return false
	}
	// Pane translates mouse reports to canvas-local, zero-based coordinates
	// before dispatching them to widgets.
	x, y := e.X, e.Y
	if y == p.searchY {
		p.focused = true
		return false
	}
	if y == p.categoryY {
		if x < 1 || x >= p.width {
			return false
		}
		category := (x - 1) / 4
		if category >= 0 && category < len(categories) {
			p.group, p.index, p.focused = category, 0, false
			p.refresh()
		}
		return false
	}
	if y >= p.gridTop && y < p.categoryY {
		if x < 1 || x >= p.width {
			return false
		}
		col := (x - 1) / emojiCellWidth
		row := y - p.gridTop
		idx := p.gridStart + row*p.cols + col
		if col >= 0 && col < p.cols && idx >= 0 && idx < len(p.items) {
			p.index, p.focused = idx, false
			p.chosen = emojiList[p.items[idx]].icon
			return true
		}
	}
	return false
}

// NewWidget constructs the picker widget for embedding in another Loom app.
func NewWidget(args []string) (loom.Widget, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("loomoji: unexpected arguments: %v", args)
	}
	return newPicker(), nil
}

// Run launches the interactive emoji picker.
func Run(args []string) error {
	cmd := &cobra.Command{Use: "loomoji", Short: "Inline searchable emoji picker", Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true, RunE: func(_ *cobra.Command, _ []string) error {
		pane, err := loom.New(13)
		if err != nil {
			return err
		}
		defer pane.Close()
		pane.Resizeable = true
		pane.EnableMouseClicks()
		app := newPicker()
		err = pane.Run(app)
		pane.Close()
		if err != nil {
			return err
		}
		if app.chosen != "" {
			_, err = fmt.Fprint(os.Stdout, app.chosen)
			return err
		}
		return nil
	}}
	cmd.SetArgs(args)
	return cmd.Execute()
}
