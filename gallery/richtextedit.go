// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package gallery

import (
	"ubunatic.com/loom"
)

// richTextEditDemo pairs the editor with a read-only link preview.
type richTextEditDemo struct {
	edit     *loom.RichTextEdit
	view     *loom.RichTextEdit
	editRect loom.Rect
	viewH    int
}

func newRichTextEditDemo() *richTextEditDemo {
	blue := loom.ColorIndex(39)
	orange := loom.ColorIndex(208)
	view := loom.NewRichTextView(&loom.RichDocument{Lines: []loom.RichLine{
		{Spans: []loom.RichSpan{{Text: "View mode (read-only): ", Style: loom.Style{Dim: true}}, {Text: "https://ubunatic.com/loom", Link: "https://ubunatic.com/loom"}}},
	}})
	edit := loom.NewRichTextEdit(&loom.RichDocument{Lines: []loom.RichLine{
		{Spans: []loom.RichSpan{{Text: "Welcome to ", Style: loom.Style{Bold: true}}, {Text: "Loom", Style: loom.Style{Bold: true, FG: orange}}, {Text: " RichText!", Style: loom.Style{Bold: true}}}},
		{},
		{Spans: []loom.RichSpan{{Text: "Try ", Style: loom.Style{}}, {Text: "bold", Style: loom.Style{Bold: true}}, {Text: ", ", Style: loom.Style{}}, {Text: "italic", Style: loom.Style{Italic: true}}, {Text: ", and ", Style: loom.Style{}}, {Text: "underline", Style: loom.Style{Underline: true}}, {Text: " formatting.", Style: loom.Style{}}}},
		{Spans: []loom.RichSpan{{Text: "Mention ", Style: loom.Style{}}, {Text: "@ada", Style: loom.Style{FG: blue, Bold: true}, PillData: &loom.RichPill{Kind: "mention", ID: "ada"}}, {Text: " or open ", Style: loom.Style{}}, {Text: "loom.dev", Link: "https://loom.dev"}, {Text: ".", Style: loom.Style{}}}},
		{Spans: []loom.RichSpan{{Text: "Drag to select; click B/I/U/S, #FG, #BG or Box. Choose a color swatch.", Style: loom.Style{Dim: true}}}},
	}})
	edit.ShowFileBar = true
	return &richTextEditDemo{view: view, edit: edit}
}

func (w *richTextEditDemo) Draw(c *loom.Canvas, r loom.Rect) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	viewHeight := 0
	if r.H >= 4 {
		viewHeight = 1
	}
	w.viewH = viewHeight
	editHeight := max(0, r.H-viewHeight)
	w.editRect = loom.Rect{X: r.X, Y: r.Y + viewHeight, W: r.W, H: editHeight}
	if viewHeight > 0 {
		w.view.Draw(c, loom.Rect{X: r.X, Y: r.Y, W: r.W, H: viewHeight})
	}
	w.edit.Draw(c, w.editRect)
}

func (w *richTextEditDemo) ConsumeKey(key loom.KeyEvent) loom.EventResult {
	return w.edit.ConsumeKey(key)
}

func (w *richTextEditDemo) ConsumeMouse(mouse loom.MouseEvent) loom.EventResult {
	if w.viewH > 0 && mouse.Y < w.viewH {
		return w.view.ConsumeMouse(mouse)
	}
	if mouse.Y < w.viewH || mouse.Y >= w.viewH+w.editRect.H {
		return loom.Ignored()
	}
	mouse.Y -= w.viewH
	return w.edit.ConsumeMouse(mouse)
}
