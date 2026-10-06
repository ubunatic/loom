// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package gallery

import (
	"ubunatic.com/loom"
)

// richTextEditDemo shows the editor with its File bar; F7 toggles View mode.
type richTextEditDemo struct {
	edit *loom.RichTextEdit
}

func newRichTextEditDemo() *richTextEditDemo {
	blue := loom.ColorIndex(39)
	orange := loom.ColorIndex(208)
	doc := &loom.RichDocument{Lines: []loom.RichLine{
		{Spans: []loom.RichSpan{{Text: "Welcome to ", Style: loom.Style{Bold: true}}, {Text: "Loom", Style: loom.Style{Bold: true, FG: orange}}, {Text: " RichText!", Style: loom.Style{Bold: true}}}},
		{},
		{Spans: []loom.RichSpan{{Text: "Try ", Style: loom.Style{}}, {Text: "bold", Style: loom.Style{Bold: true}}, {Text: ", ", Style: loom.Style{}}, {Text: "italic", Style: loom.Style{Italic: true}}, {Text: ", and ", Style: loom.Style{}}, {Text: "underline", Style: loom.Style{Underline: true}}, {Text: " formatting.", Style: loom.Style{}}}},
		{Spans: []loom.RichSpan{{Text: "Mention ", Style: loom.Style{}}, {Text: "@ada", Style: loom.Style{FG: blue, Bold: true}, PillData: &loom.RichPill{Kind: "mention", ID: "ada"}}, {Text: " or open ", Style: loom.Style{}}, {Text: "loom.dev", Link: "https://loom.dev"}, {Text: ".", Style: loom.Style{}}}},
		{Spans: []loom.RichSpan{{Text: "Drag to select; click B/I/U/S, #FG, #BG or Box. Choose a color swatch.", Style: loom.Style{Dim: true}}}},
	}}
	return newRichTextEditDemoWithDoc(doc)
}

func newRichTextEditDemoWithDoc(doc *loom.RichDocument) *richTextEditDemo {
	edit := loom.NewRichTextEdit(doc)
	edit.ShowFileBar = true
	return &richTextEditDemo{edit: edit}
}

func (w *richTextEditDemo) Draw(c *loom.Canvas, r loom.Rect) { w.edit.Draw(c, r) }

func (w *richTextEditDemo) ConsumeKey(key loom.KeyEvent) loom.EventResult {
	return w.edit.ConsumeKey(key)
}

func (w *richTextEditDemo) ConsumeMouse(mouse loom.MouseEvent) loom.EventResult {
	return w.edit.ConsumeMouse(mouse)
}

func (w *richTextEditDemo) HotkeyHint(width int) string { return w.edit.HotkeyHint(width) }

func (w *richTextEditDemo) HotkeyBar() *loom.HintBar { return w.edit.HotkeyBar() }

func (w *richTextEditDemo) ApplyTheme(theme loom.ThemeColors) {
	w.edit.ApplyTheme(theme)
}
