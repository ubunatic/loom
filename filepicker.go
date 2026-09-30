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
)

// FilePickerOptions configures a filesystem picker.
type FilePickerOptions struct {
	Mode FilePickerMode
	// Patterns are filepath.Match globs applied to file names; directories are never filtered.
	Patterns []string
	Style    ChoiceStyle
	OnSelect func(path string)
	OnCancel func()
}

// FilePicker is a keyboard and mouse navigable filesystem picker.
type FilePicker struct {
	options   FilePickerOptions
	directory Directory
	entries   []FileEntry
	items     map[string]FileEntry
	list      *Choice
	lastRect  Rect
	listRect  Rect
	done      bool
}

// NewFilePicker reads dir and creates a reusable file or directory picker.
func NewFilePicker(dir string, options FilePickerOptions) (*FilePicker, error) {
	if options.Mode > FilePickerDirectories {
		return nil, fmt.Errorf("loom: invalid file picker mode %d", options.Mode)
	}
	p := &FilePicker{options: options, list: NewChoice(nil), items: make(map[string]FileEntry)}
	p.list.Fuzzy = true
	p.list.SelectOnlyOnClick = true
	p.list.DoubleClickToActivate = true
	p.list.MouseTextOnly = true
	p.list.OnSelect = func(Item) { p.activate() }
	if options.Style != (ChoiceStyle{}) {
		p.list.Style = options.Style
	}
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
func (p *FilePicker) ContentHeight() int { return p.list.ContentHeight() + 1 }

// ApplyTheme updates the embedded list style.
func (p *FilePicker) ApplyTheme(theme ThemeColors) { p.list.ApplyTheme(theme) }

// Draw renders the current path and entry list.
func (p *FilePicker) Draw(c *Canvas, r Rect) {
	p.lastRect = r
	if r.W <= 0 || r.H <= 0 {
		return
	}
	c.Write(r.X, r.Y, DisplayPath(p.directory.Path), p.list.Style.Normal)
	p.listRect = Rect{X: r.X, Y: r.Y + 1, W: r.W, H: max(0, r.H-1)}
	if p.listRect.H > 0 {
		p.list.Draw(c, p.listRect)
	}
}

// HandleKey navigates, selects, or cancels the picker.
func (p *FilePicker) HandleKey(e KeyEvent) bool {
	if p.done {
		return false
	}
	if e.Is("esc") {
		p.done = true
		if p.options.OnCancel != nil {
			p.options.OnCancel()
		}
		return false
	}
	if e.Is("backspace") {
		if p.list.Query() != "" {
			return p.list.HandleKey(e)
		}
		p.parent()
		return false
	}
	if e.Is("enter") {
		return p.list.HandleKey(e)
	}
	return p.list.HandleKey(e)
}

// ConsumeKey reports picker navigation and selection as consumed events.
func (p *FilePicker) ConsumeKey(e KeyEvent) EventResult {
	before, dir, query := p.list.FilteredSel(), p.directory.Path, p.list.Query()
	p.HandleKey(e)
	return EventResult{Consumed: p.done || before != p.list.FilteredSel() || dir != p.directory.Path || query != p.list.Query() || e.Is("esc", "enter", "backspace", "up", "down", "pgup", "pgdown", "pgdn", "pagedown", "pageup") || e.Text != ""}
}

// HandleMouse forwards child-local coordinates to the embedded Choice.
func (p *FilePicker) HandleMouse(e MouseEvent) bool {
	if e.X < 0 || e.Y < 1 || e.X >= p.lastRect.W || e.Y >= p.lastRect.H {
		return false
	}
	e.X += p.listRect.X
	e.Y += p.listRect.Y
	quit := p.list.HandleMouse(e)
	return quit
}

// ConsumeMouse adapts mouse input to EventConsumer hosts.
func (p *FilePicker) ConsumeMouse(e MouseEvent) EventResult {
	before := p.list.FilteredSel()
	quit := p.HandleMouse(e)
	return EventResult{Consumed: before != p.list.FilteredSel() || (e.Action == MousePress && e.Y >= 1 && e.Y < p.lastRect.H), Quit: quit}
}

var _ Widget = (*FilePicker)(nil)
var _ EventConsumer = (*FilePicker)(nil)
var _ MouseConsumer = (*FilePicker)(nil)
var _ Themeable = (*FilePicker)(nil)
