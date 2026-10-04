// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package gallery

import "ubunatic.com/loom"

// richTextEditDemo pairs the editor with a compact keycap hint row.
type richTextEditDemo struct {
	edit *loom.RichTextEdit
	view *loom.RichTextEdit
	area loom.Rect
}

func newRichTextEditDemo() *richTextEditDemo {
	blue := loom.ColorIndex(39)
	orange := loom.ColorIndex(208)
	view := loom.NewRichTextView(&loom.RichDocument{Lines: []loom.RichLine{
		{Spans: []loom.RichSpan{{Text: "View mode (read-only): ", Style: loom.Style{Dim: true}}, {Text: "https://ubunatic.com/loom", Link: "https://ubunatic.com/loom"}}},
	}})
	return &richTextEditDemo{view: view, edit: loom.NewRichTextEdit(&loom.RichDocument{Lines: []loom.RichLine{
		{Spans: []loom.RichSpan{{Text: "Welcome to ", Style: loom.Style{Bold: true}}, {Text: "Loom", Style: loom.Style{Bold: true, FG: orange}}, {Text: " RichText!", Style: loom.Style{Bold: true}}}},
		{},
		{Spans: []loom.RichSpan{{Text: "Try ", Style: loom.Style{}}, {Text: "bold", Style: loom.Style{Bold: true}}, {Text: ", ", Style: loom.Style{}}, {Text: "italic", Style: loom.Style{Italic: true}}, {Text: ", and ", Style: loom.Style{}}, {Text: "underline", Style: loom.Style{Underline: true}}, {Text: " formatting.", Style: loom.Style{}}}},
		{Spans: []loom.RichSpan{{Text: "Mention ", Style: loom.Style{}}, {Text: "@ada", Style: loom.Style{FG: blue, Bold: true}, PillData: &loom.RichPill{Kind: "mention", ID: "ada"}}, {Text: " or open ", Style: loom.Style{}}, {Text: "loom.dev", Link: "https://loom.dev"}, {Text: ".", Style: loom.Style{}}}},
		{Spans: []loom.RichSpan{{Text: "Drag to select; click B/I/U/S, #FG or #BG. Choose a color swatch.", Style: loom.Style{Dim: true}}}},
	}})}
}

func (w *richTextEditDemo) Draw(c *loom.Canvas, r loom.Rect) {
	w.area = r
	if r.W <= 0 || r.H <= 0 {
		return
	}
	viewHeight := 0
	if r.H >= 4 {
		viewHeight = 1
	}
	editHeight := max(0, r.H-1-viewHeight)
	w.edit.Draw(c, loom.Rect{X: r.X, Y: r.Y, W: r.W, H: editHeight})
	if viewHeight > 0 {
		w.view.Draw(c, loom.Rect{X: r.X, Y: r.Y + editHeight, W: r.W, H: viewHeight})
	}
	if r.H > 0 {
		c.WriteANSI(r.X, r.Y+r.H-1, "\x1b[2m[Ctrl+B/I/U] format  [Shift+←/→] select  Drag then click #FG/#BG for colors\x1b[0m")
	}
}

func (w *richTextEditDemo) ConsumeKey(key loom.KeyEvent) loom.EventResult {
	return w.edit.ConsumeKey(key)
}

func (w *richTextEditDemo) ConsumeMouse(mouse loom.MouseEvent) loom.EventResult {
	editHeight := w.area.H - 1
	if w.area.H >= 4 {
		editHeight--
		if mouse.Y == editHeight {
			mouse.Y = 0
			return w.view.ConsumeMouse(mouse)
		}
	}
	if mouse.Y >= editHeight {
		return loom.Ignored()
	}
	return w.edit.ConsumeMouse(mouse)
}
