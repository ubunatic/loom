// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"path/filepath"
	"strings"
	"unicode/utf8"
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
	defs := SpeccedDefaults.RichTextEdit
	mnemonic, _ := utf8.DecodeRuneInString(defs.FileMenuMnemonic)
	bar.menu = NewMenuBar(Menu{
		Title:    defs.FileMenuTitle,
		Mnemonic: mnemonic,
		Items: []MenuItem{
			{Label: SpeccedDefaults.RichTextEdit.HotkeyOpenLabel, Shortcut: KeyCap(SpeccedDefaults.RichTextEdit.HotkeyOpenBinding), Action: func() { bar.menu.SetFocus(false); bar.edit.RequestOpen() }},
			{Label: SpeccedDefaults.RichTextEdit.HotkeyCloseLabel, Shortcut: KeyCap(SpeccedDefaults.RichTextEdit.HotkeyCloseBinding), Action: func() { bar.menu.SetFocus(false); bar.edit.RequestClose() }},
			{Label: defs.HotkeySaveLabel, Shortcut: KeyCap(SpeccedDefaults.RichTextEdit.HotkeySaveBinding), Action: func() { bar.save() }},
			{Label: defs.HotkeySaveAsLabel + "…", Action: func() { bar.saveAs() }},
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
	theme := b.edit.toolbarTheme()
	normal := Style{FG: theme.NormalFG.Color(), BG: theme.NormalBG.Color()}
	selected := Style{FG: theme.SelectedFG.Color(), BG: theme.SelectedBG.Color(), Bold: theme.SelectedBold}
	b.menu.Style = MenuStyle{Bar: normal, Normal: normal, Active: selected, Selected: selected,
		Border: Style{FG: theme.BorderFG.Color(), BG: theme.BorderBG.Color()}, Disabled: Style{FG: theme.PlaceholderFG.Color(), BG: theme.NormalBG.Color(), Dim: true}}
	b.menu.Draw(c, bounds)
	row := bounds.Y + bounds.H - 1
	titleWidth := 0
	if len(b.menu.titleRects) > 0 {
		titleWidth = b.menu.titleRects[0].W
	}
	if titleWidth == 0 {
		titleWidth = StringWidth(" " + SpeccedDefaults.RichTextEdit.FileMenuTitle + " ")
	}
	available := max(0, bounds.W-titleWidth)
	status := b.status()
	hints := ""
	if b.edit.BoxMode {
		hints = "[Box] Enter/Esc exits"
	}
	statusSeparator := strings.LastIndex(status, " · ")
	minimumStatusWidth := StringWidth(status)
	if statusSeparator >= 0 {
		minimumStatusWidth = StringWidth("…" + status[statusSeparator:])
	}
	minimumRowWidth := titleWidth + 1 + minimumStatusWidth + 1 + StringWidth(hints)
	if StringWidth(hints) > available || bounds.W < minimumRowWidth {
		hints = ""
	}
	hintWidth := StringWidth(hints)
	hintX := bounds.X + bounds.W - hintWidth
	statusStart := bounds.X + titleWidth
	statusWidth := max(0, hintX-statusStart-1)
	if statusWidth > 0 {
		text := " " + truncateRichTextFileBarStatus(status, statusWidth-1)
		style := Style{FG: theme.PlaceholderFG.Color(), BG: theme.NormalBG.Color()}
		c.Write(statusStart, row, text, style)
		if separator := strings.LastIndex(text, " · "); separator >= 0 {
			state := text[separator+len(" · "):]
			stateStyle := normal
			switch state {
			case DocStateUntitled.Label():
				stateStyle.FG, stateStyle.Dim = theme.PlaceholderFG.Color(), true
			case DocStateModified.Label():
				stateStyle.FG = theme.ModifiedFG.Color()
			case DocStateSaved.Label():
				stateStyle.FG = theme.SavedFG.Color()
			case DocStateError.Label():
				stateStyle.FG = theme.MediaErrorFG.Color()
			}
			c.Write(statusStart+StringWidth(text[:separator+len(" · ")]), row, state, stateStyle)
		}
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
	status := e.DocState().Label()
	return name + " · " + status
}

func truncateRichTextFileBarStatus(status string, width int) string {
	if StringWidth(status) <= width {
		return status
	}
	separator := strings.LastIndex(status, " · ")
	if separator < 0 {
		return TruncateText(status, width, "…")
	}
	suffix := status[separator:]
	suffixWidth := StringWidth(suffix)
	if suffixWidth >= width {
		return TruncateText(status, width, "…")
	}
	nameWidth := width - suffixWidth
	return TruncateText(status[:separator], nameWidth, "…") + suffix
}
