// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"path/filepath"
)

// FilePickerMode controls which entries a FilePicker can select.
type FilePickerMode uint8

const (
	// FilePickerFiles selects regular files and uses directories for navigation.
	FilePickerFiles FilePickerMode = iota
	// FilePickerDirectories selects directories and hides files.
	FilePickerDirectories
	// FilePickerSave selects a file name in the current directory as a save destination.
	FilePickerSave
)

// FilePickerOptions configures a filesystem picker.
type FilePickerOptions struct {
	Mode FilePickerMode
	// FileName seeds the destination name in save mode.
	FileName string
	// FocusFileName starts save mode with the destination name field focused.
	FocusFileName bool
	// Patterns are filepath.Match globs applied to file names; directories are never filtered.
	Patterns []string
	Style    ChoiceStyle
	OnSelect func(path string)
	OnCancel func()
	// OnSave validates or writes a save-mode destination. An error keeps the
	// picker open so the caller can report the problem and the user can retry.
	OnSave func(path string) error
}

// FilePicker is a keyboard and mouse navigable filesystem picker.
type FilePicker struct {
	options   FilePickerOptions
	directory Directory
	entries   []FileEntry
	items     map[string]FileEntry
	list      *Choice
	fileName  *TextInput
	nameFocus bool
	lastRect  Rect
	listRect  Rect
	done      bool
}

// NewFilePicker reads dir and creates a reusable file or directory picker.
func NewFilePicker(dir string, options FilePickerOptions) (*FilePicker, error) {
	if options.Mode > FilePickerSave {
		return nil, fmt.Errorf("loom: invalid file picker mode %d", options.Mode)
	}
	p := &FilePicker{options: options, list: NewChoice(nil), items: make(map[string]FileEntry)}
	p.list.Fuzzy = true
	p.list.SelectOnlyOnClick = true
	p.list.DoubleClickToActivate = true
	p.list.MouseTextOnly = true
	p.list.OnSelect = func(Item) { p.activate() }
	if options.Mode == FilePickerSave {
		p.fileName = NewTextInput(options.FileName)
		p.fileName.Prompt = "Name: "
		p.nameFocus = options.FocusFileName
	}
	if options.Style != (ChoiceStyle{}) {
		p.list.Style = options.Style
	}
	p.updateFocus()
	if err := p.open(dir, ""); err != nil {
		return nil, err
	}
	return p, nil
}

// Directory returns the directory currently being browsed.
func (p *FilePicker) Directory() Directory { return p.directory }

// Entries returns the selectable or navigable entries currently shown.
func (p *FilePicker) Entries() []FileEntry { return append([]FileEntry(nil), p.entries...) }

// List returns the underlying Choice for host customization.
func (p *FilePicker) List() *Choice { return p.list }

// FileName returns the destination name entered in save mode.
func (p *FilePicker) FileName() string {
	if p.fileName == nil {
		return ""
	}
	return p.fileName.Value()
}

// Selected returns the current directory entry, if one is selected.
func (p *FilePicker) Selected() (FileEntry, bool) {
	item, ok := p.list.Selected()
	if !ok {
		return FileEntry{}, false
	}
	e, ok := p.items[item.Name]
	return e, ok
}

func (p *FilePicker) open(path, selectName string) error {
	dir, err := ReadDirectory(path, DirectoryOptions{IncludeParent: true})
	if err != nil {
		return fmt.Errorf("loom: open file picker directory: %w", err)
	}
	entries := make([]FileEntry, 0, len(dir.Entries))
	items := make([]Item, 0, len(dir.Entries))
	byName := make(map[string]FileEntry)
	for _, entry := range dir.Entries {
		if p.options.Mode == FilePickerDirectories && entry.Kind != FileKindDirectory {
			continue
		}
		if entry.Kind != FileKindDirectory && entry.Kind != FileKindSymlink && !p.matches(entry.Name) {
			continue
		}
		entries = append(entries, entry)
		desc := ""
		if entry.IsParent {
			desc = "parent directory"
		} else if entry.Kind == FileKindDirectory {
			desc = "<dir>"
		} else if entry.Kind == FileKindSymlink {
			desc = "<link>"
		}
		items = append(items, Item{Name: entry.DisplayName(), Desc: desc})
		byName[entry.DisplayName()] = entry
	}
	p.directory, p.entries, p.items = dir, entries, byName
	p.list.SetItems(items)
	if selectName != "" {
		for i, entry := range entries {
			if entry.Name == selectName {
				p.list.SelectIndex(i)
				break
			}
		}
	} else if p.options.Mode != FilePickerDirectories && len(entries) > 1 && entries[0].IsParent {
		p.list.SelectIndex(1)
	}
	return nil
}

func (p *FilePicker) matches(name string) bool {
	if len(p.options.Patterns) == 0 {
		return true
	}
	for _, pattern := range p.options.Patterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

func (p *FilePicker) activate() {
	entry, ok := p.Selected()
	if !ok {
		return
	}
	if entry.Kind == FileKindDirectory {
		if p.options.Mode == FilePickerDirectories {
			p.done = true
			if p.options.OnSelect != nil {
				p.options.OnSelect(entry.Path)
			}
			return
		}
		selectName := ""
		if entry.IsParent {
			selectName = filepath.Base(p.directory.Path)
		}
		_ = p.open(entry.Path, selectName)
		return
	}
	if p.options.Mode == FilePickerSave && entry.Kind == FileKindRegular {
		p.fileName.SetValue(entry.Name)
		p.nameFocus = true
		return
	}
	if entry.Kind == FileKindSymlink || entry.Kind != FileKindRegular || p.options.Mode == FilePickerDirectories {
		return
	}
	p.done = true
	if p.options.OnSelect != nil {
		p.options.OnSelect(entry.Path)
	}
}

func (p *FilePicker) parent() {
	if parent, ok := p.directory.Parent(); ok {
		_ = p.open(parent, filepath.Base(p.directory.Path))
	}
}

// ContentHeight reports the picker content height for a Viewport host.
func (p *FilePicker) ContentHeight() int {
	height := p.list.ContentHeight() + 1
	if p.fileName != nil {
		height += 2
	}
	return height
}

// ApplyTheme updates the embedded list style.
func (p *FilePicker) ApplyTheme(theme ThemeColors) { p.list.ApplyTheme(theme) }

// Draw renders the current path and entry list.
func (p *FilePicker) Draw(c *Canvas, r Rect) {
	p.lastRect = r
	if r.W <= 0 || r.H <= 0 {
		return
	}
	c.Write(r.X, r.Y, DisplayPath(p.directory.Path), p.list.Style.Normal)
	listHeight := max(0, r.H-1)
	if p.fileName != nil && listHeight > 0 {
		listHeight--
		p.fileName.Draw(c, Rect{X: r.X, Y: r.Y + r.H - 1, W: r.W, H: 1}, p.nameFocus)
		if listHeight > 0 {
			listHeight--
			for x := r.X; x < r.X+r.W; x++ {
				c.Set(x, r.Y+1+listHeight, Cell{Text: "─", Style: p.list.Style.Normal})
			}
		}
	}
	p.listRect = Rect{X: r.X, Y: r.Y + 1, W: r.W, H: listHeight}
	if p.listRect.H > 0 {
		p.list.Draw(c, p.listRect)
	}
}

// ConsumeKey navigates, selects, or cancels the picker.
func (p *FilePicker) ConsumeKey(e KeyEvent) EventResult {
	if p.done {
		return Ignored()
	}
	if e.Is("esc") {
		p.done = true
		if p.options.OnCancel != nil {
			p.options.OnCancel()
		}
		return Handled()
	}
	if e.Is("backspace") {
		if p.fileName != nil && p.nameFocus {
			return p.fileName.ConsumeKey(e)
		}
		if p.list.Query() != "" {
			return p.list.ConsumeKey(e)
		}
		p.parent()
		return Handled()
	}
	if e.Is("enter") {
		if p.fileName != nil && p.nameFocus {
			name := p.fileName.Value()
			if name == "" {
				return Handled()
			}
			path := filepath.Join(p.directory.Path, name)
			if p.options.OnSave != nil {
				if err := p.options.OnSave(path); err != nil {
					return Handled()
				}
			} else if p.options.OnSelect != nil {
				p.options.OnSelect(path)
			}
			p.done = true
			return Handled()
		}
		return p.list.ConsumeKey(e)
	}
	if p.fileName != nil && e.Is("tab", "shift-tab") {
		p.nameFocus = !p.nameFocus
		p.updateFocus()
		return Handled()
	}
	if p.fileName != nil && p.nameFocus {
		return p.fileName.ConsumeKey(e)
	}
	return p.list.ConsumeKey(e)
}

// ConsumeMouse forwards child-local coordinates to the embedded Choice.
func (p *FilePicker) ConsumeMouse(e MouseEvent) EventResult {
	if e.X < 0 || e.Y < 1 || e.X >= p.lastRect.W || e.Y >= p.lastRect.H {
		return Ignored()
	}
	if p.fileName != nil && e.Y == p.lastRect.H-1 {
		p.nameFocus = true
		p.updateFocus()
		return Handled()
	}
	p.nameFocus = false
	p.updateFocus()
	e.X -= p.listRect.X - p.lastRect.X
	e.Y -= p.listRect.Y - p.lastRect.Y
	return p.list.ConsumeMouse(e)
}

func (p *FilePicker) updateFocus() {
	p.list.SetFocus(!p.nameFocus)
}

var _ Widget = (*FilePicker)(nil)
var _ EventConsumer = (*FilePicker)(nil)
var _ MouseConsumer = (*FilePicker)(nil)
var _ Themeable = (*FilePicker)(nil)
