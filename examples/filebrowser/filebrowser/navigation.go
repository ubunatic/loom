package filebrowser

import (
	"fmt"
	"path/filepath"

	"codeberg.org/ubunatic/loom"
)

// NavigationPaneOptions configures a reusable filesystem navigation pane.
type NavigationPaneOptions struct {
	// OnSelection is called when the current selection changes.
	OnSelection func(loom.FileEntry)
	// OnActivate is called when the user confirms a non-directory entry.
	OnActivate func(loom.FileEntry)
	// OnOpen is called after a directory has been opened successfully.
	OnOpen func(loom.Directory)
	// OnQuit is called when ESC is pressed at the filesystem root.
	OnQuit func()
	// Style overrides the default Choice appearance when non-zero.
	Style loom.ChoiceStyle
}

// NavigationPane provides a filterable, keyboard- and mouse-driven directory list.
// Hosts can inspect Directory and Selected to update their own preview or metadata.
type NavigationPane struct {
	rootDir   string
	directory loom.Directory
	entries   map[string]loom.FileEntry
	list      *loom.Choice
	options   NavigationPaneOptions
	lastRect  loom.Rect
}

// NewNavigationPane reads dir and creates a navigation pane with an initial selection.
func NewNavigationPane(dir string, options NavigationPaneOptions) (*NavigationPane, error) {
	pane := &NavigationPane{options: options}
	if err := pane.open(dir, ""); err != nil {
		return nil, err
	}
	pane.rootDir = pane.directory.Path
	pane.notifySelection()
	return pane, nil
}

// Directory returns the current directory and its entries.
func (p *NavigationPane) Directory() loom.Directory { return p.directory }

// List returns the underlying Choice for styling and host integration.
func (p *NavigationPane) List() *loom.Choice { return p.list }

// Selected returns the currently selected filesystem entry, if any.
func (p *NavigationPane) Selected() (loom.FileEntry, bool) {
	item, ok := p.list.Selected()
	if !ok {
		return loom.FileEntry{}, false
	}
	entry, ok := p.entries[item.Name]
	return entry, ok
}

func (p *NavigationPane) notifySelection() {
	if p.options.OnSelection != nil {
		if entry, ok := p.Selected(); ok {
			p.options.OnSelection(entry)
		}
	}
}

func (p *NavigationPane) open(dir, selectName string) error {
	directory, err := loom.ReadDirectory(dir, loom.DirectoryOptions{IncludeParent: true})
	if err != nil {
		return fmt.Errorf("filebrowser: open navigation directory: %w", err)
	}
	items := make([]loom.Item, 0, len(directory.Entries))
	entries := make(map[string]loom.FileEntry, len(directory.Entries))
	for _, entry := range directory.Entries {
		name := entry.DisplayName()
		desc := ""
		if entry.IsParent {
			desc = "parent directory"
		} else if entry.Kind == loom.FileKindDirectory {
			desc = "<dir>"
		} else if entry.Kind == loom.FileKindSymlink {
			desc = "<link>"
		}
		items = append(items, loom.Item{Name: name, Desc: desc})
		entries[name] = entry
	}
	list := loom.NewChoice(items)
	list.SelectOnlyOnClick = true
	list.MouseTextOnly = true
	list.Prompt = "filter> "
	list.Placeholder = "type to filter"
	if p.options.Style != (loom.ChoiceStyle{}) {
		list.Style = p.options.Style
	}
	list.OnSelect = func(item loom.Item) {
		entry, ok := entries[item.Name]
		if !ok {
			return
		}
		if entry.Kind == loom.FileKindDirectory {
			name := ""
			if entry.IsParent {
				name = filepath.Base(directory.Path)
			}
			if err := p.open(entry.Path, name); err == nil && p.options.OnOpen != nil {
				p.options.OnOpen(p.directory)
			}
			return
		}
		if p.options.OnActivate != nil {
			p.options.OnActivate(entry)
		}
	}
	for i, item := range items {
		if item.Name == selectName {
			for range i {
				list.HandleKey(loom.KeyEvent{Key: "down"})
			}
			break
		}
	}
	if selectName == "" && len(items) > 1 && items[0].Name == ".." {
		list.HandleKey(loom.KeyEvent{Key: "down"})
	}
	p.directory, p.entries, p.list = directory, entries, list
	p.notifySelection()
	return nil
}

// Draw renders the navigation list within r.
func (p *NavigationPane) Draw(c *loom.Canvas, r loom.Rect) {
	p.lastRect = r
	p.list.Draw(c, r)
}

// HandleKey processes navigation, filtering, directory opening, and selection.
func (p *NavigationPane) HandleKey(e loom.KeyEvent) bool {
	if e.Is("esc") {
		if p.directory.Path == p.rootDir {
			if p.options.OnQuit != nil {
				p.options.OnQuit()
			}
			return true
		}
		parent, ok := p.directory.Parent()
		if !ok || parent == p.directory.Path {
			if p.options.OnQuit != nil {
				p.options.OnQuit()
			}
			return true
		}
		if err := p.open(parent, filepath.Base(p.directory.Path)); err != nil {
			return false
		}
		if p.options.OnOpen != nil {
			p.options.OnOpen(p.directory)
		}
		return false
	}
	if e.Is("enter") {
		item, ok := p.list.Selected()
		if !ok {
			return false
		}
		entry, ok := p.entries[item.Name]
		if !ok {
			return false
		}
		if entry.Kind == loom.FileKindDirectory {
			name := ""
			if entry.IsParent {
				name = filepath.Base(p.directory.Path)
			}
			if err := p.open(entry.Path, name); err != nil {
				return false
			}
			if p.options.OnOpen != nil {
				p.options.OnOpen(p.directory)
			}
			return false
		}
		if p.options.OnActivate != nil {
			p.options.OnActivate(entry)
		}
		return false
	}
	quit := p.list.HandleKey(e)
	p.notifySelection()
	return quit
}

// HandleMouse processes mouse selection and forwards quit requests.
func (p *NavigationPane) HandleMouse(e loom.MouseEvent) bool {
	e.X += p.lastRect.X
	e.Y += p.lastRect.Y
	quit := p.list.HandleMouse(e)
	p.notifySelection()
	return quit
}

// ConsumeKey reports whether a key belongs to the navigation pane's host-level
// navigation contract. ESC is consumed for both parent navigation and root quit.
func (p *NavigationPane) ConsumeKey(e loom.KeyEvent) (quit, consumed bool) {
	if !e.Is("esc") {
		return false, false
	}
	return p.HandleKey(e), true
}
