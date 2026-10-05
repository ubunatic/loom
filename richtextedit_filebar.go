// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"path/filepath"
	"reflect"
)

type richTextEditFileBar struct {
	edit *RichTextEdit
	menu *MenuBar
}

func (e *RichTextEdit) ensureFileBar() *richTextEditFileBar {
	if e.fileBar != nil {
		return e.fileBar
	}
	bar := &richTextEditFileBar{edit: e}
	bar.menu = NewMenuBar(Menu{
		Title:    "File",
		Mnemonic: 'F',
		Items: []MenuItem{
			{Label: "Save", Shortcut: "Ctrl+S", Action: func() { bar.save() }},
			{Label: "Save as…", Shortcut: "Ctrl+Shift+S", Action: func() { bar.saveAs() }},
		},
	})
	bar.menu.Bottom = true
	e.fileBar = bar
	return bar
}

func (b *richTextEditFileBar) save() {
	b.menu.SetFocus(false)
	if err := b.edit.Save(); err != nil {
		b.edit.LastSaveError = err
	}
}

func (b *richTextEditFileBar) saveAs() {
	b.menu.SetFocus(false)
	if err := b.edit.openSavePicker(); err != nil {
		b.edit.LastSaveError = err
	}
}

func (b *richTextEditFileBar) ConsumeKey(key KeyEvent) EventResult {
	active := b.menu.Open || b.menu.Focused()
	result := b.menu.ConsumeKey(key)
	if !active {
		if key.Is("f7") {
			if b.edit.BoxMode {
				b.edit.toggleBoxMode()
			}
			b.edit.ViewMode = !b.edit.ViewMode
			return Handled()
		}
		return result
	}
	if !b.menu.Open {
		b.menu.SetFocus(false)
	}
	if result.Consumed {
		return result
	}
	return Handled()
}

func (b *richTextEditFileBar) Draw(c *Canvas, bounds Rect) {
	if c == nil || bounds.W <= 0 || bounds.H <= 0 {
		return
	}
	b.menu.Draw(c, bounds)
	row := bounds.Y + bounds.H - 1
	titleWidth := 0
	if len(b.menu.titleRects) > 0 {
		titleWidth = b.menu.titleRects[0].W
	}
	if titleWidth == 0 {
		titleWidth = StringWidth(" File ")
	}
	viewHint := "[F7] Edit"
	if b.edit.ViewMode {
		viewHint = "[F7] View"
	}
	hints := "[F10] File  Ctrl+S Save  Ctrl+Shift+S Save as  " + viewHint
	if b.edit.BoxMode {
		hints = "[Box mode] Esc exits  " + hints
	}
	available := max(0, bounds.W-titleWidth)
	hints = TruncateText(hints, available, "")
	hintWidth := StringWidth(hints)
	hintX := bounds.X + bounds.W - hintWidth
	statusStart := bounds.X + titleWidth
	statusWidth := max(0, hintX-statusStart-1)
	status := b.status()
	if statusWidth > 0 {
		c.Write(statusStart, row, " "+TruncateText(status, statusWidth-1, "…"), Style{Dim: true})
	}
	if hintWidth > 0 {
		c.Write(hintX, row, hints, Style{Dim: true})
	}
}

func (b *richTextEditFileBar) status() string {
	e := b.edit
	e.ensureDocument()
	name := "Untitled"
	if e.FilePath != "" {
		name = filepath.Base(e.FilePath)
	}
	status := "Saved"
	if e.LastSaveError != nil {
		status = "Error"
	} else if !reflect.DeepEqual(e.savedDocument, e.Document.Lines) {
		status = "Modified"
	} else if e.FilePath == "" {
		status = "Unsaved"
	}
	return name + " · " + status
}
