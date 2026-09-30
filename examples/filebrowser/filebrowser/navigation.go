package filebrowser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	// SelectParent keeps the parent-directory row selected when a directory opens.
	// By default, the first non-parent entry is selected.
	SelectParent bool
	// Style overrides the default Choice appearance when non-zero.
	Style loom.ChoiceStyle
	// TypeToSearch sends every typed character to the filter, as before "/" gating.
	// By default, typing filters only after "/", so the app's own letter keys work.
	TypeToSearch bool
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
	searching bool
}

// NewNavigationPane reads dir and creates a navigation pane with an initial selection.
func NewNavigationPane(dir string, options NavigationPaneOptions) (*NavigationPane, error) {
	pane := &NavigationPane{
		options: options,
		list:    loom.NewChoice(nil),
	}
	pane.list.SelectOnlyOnClick = true
	pane.list.DoubleClickToActivate = true
	pane.list.MouseTextOnly = true
	pane.list.Prompt = "filter> "
	pane.list.Placeholder = "type to filter"
	if pane.options.Style != (loom.ChoiceStyle{}) {
		pane.list.Style = pane.options.Style
	}
	pane.list.OnSelect = func(item loom.Item) {
		entry, ok := pane.entries[item.Name]
		if !ok {
			return
		}
		if entry.Kind == loom.FileKindDirectory {
			name := ""
			if entry.IsParent {
				name = filepath.Base(pane.directory.Path)
			}
			if err := pane.open(entry.Path, name); err == nil {
				if pane.options.OnOpen != nil {
					pane.options.OnOpen(pane.directory)
				}
			}
			return
		}
		if entry.Kind == loom.FileKindSymlink {
			if info, err := os.Stat(entry.Path); err == nil && info.IsDir() {
				if err := pane.open(entry.Path, ""); err == nil {
					if pane.options.OnOpen != nil {
						pane.options.OnOpen(pane.directory)
					}
				}
				return
			}
		}
		if pane.options.OnActivate != nil {
			pane.options.OnActivate(entry)
		}
	}
	if err := pane.open(dir, ""); err != nil {
		return nil, err
	}
	pane.rootDir = pane.directory.Path
	pane.notifySelection()
	return pane, nil
}

// Directory returns the current directory and its entries.
func (p *NavigationPane) Directory() loom.Directory { return p.directory }

// SetRoot sets the directory where ESC requests a quit from the host.
// Passing "" disables the root quit trigger.
func (p *NavigationPane) SetRoot(dir string) {
	if dir == "" {
		p.rootDir = ""
		return
	}
	p.rootDir = filepath.Clean(dir)
}

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

// Focused reports whether the underlying choice list has input focus.
func (p *NavigationPane) Focused() bool {
	return p.list != nil && p.list.Focused()
}

// SetFocus enables or disables input focus on the underlying choice list.
func (p *NavigationPane) SetFocus(focused bool) {
	if p.list != nil {
		p.list.SetFocus(focused)
	}
}

// ContentHeight estimates the required height of the navigation list.
func (p *NavigationPane) ContentHeight() int {
	if p.list != nil {
		return p.list.ContentHeight()
	}
	return 1
}

// ApplyTheme updates the Choice style from the theme.
func (p *NavigationPane) ApplyTheme(theme loom.ThemeColors) {
	if p.list != nil {
		p.list.ApplyTheme(theme)
	}
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
	p.directory, p.entries = directory, entries
	p.list.SetItems(items)
	if selectName != "" {
		for i, item := range items {
			if item.Name == selectName {
				p.list.SelectIndex(i)
				break
			}
		}
	} else if !p.options.SelectParent && len(items) > 1 && items[0].Name == ".." {
		p.list.SelectIndex(1)
	}
	p.notifySelection()
	return nil
}

// Draw renders the navigation list within r.
func (p *NavigationPane) Draw(c *loom.Canvas, r loom.Rect) {
	p.lastRect = r
	if !p.options.TypeToSearch && !p.searching {
		focused := p.list.Focused()
		p.list.SetFocus(false)
		p.list.Draw(c, r)
		promptY := r.Y + r.H - 1
		if p.list.PromptTop {
			promptY = r.Y
		}
		c.Write(r.X, promptY, strings.Repeat(" ", r.W), p.list.Style.Placeholder)
		c.Write(r.X, promptY, "/ search", p.list.Style.Placeholder)
		p.list.SetFocus(focused)
		return
	}
	p.list.Draw(c, r)
}

// ConsumeKey processes navigation, filtering, directory opening, and selection.
func (p *NavigationPane) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if handled, quit := p.handleSearchKey(e); handled {
		return loom.EventResult{Consumed: true, Quit: quit}
	}
	// Outside search, leave printable text to the host's bindings.
	if !p.Searching() && e.Key == "" && e.Text != "" {
		return loom.Ignored()
	}
	if e.Is("esc") {
		if p.rootDir != "" && p.directory.Path == p.rootDir {
			if p.options.OnQuit != nil {
				p.options.OnQuit()
			}
			return loom.QuitResult()
		}
		parent, ok := p.directory.Parent()
		if !ok || parent == p.directory.Path {
			if p.rootDir != "" {
				if p.options.OnQuit != nil {
					p.options.OnQuit()
				}
				return loom.QuitResult()
			}
			return loom.Handled()
		}
		if err := p.open(parent, filepath.Base(p.directory.Path)); err != nil {
			return loom.Ignored()
		}
		if p.options.OnOpen != nil {
			p.options.OnOpen(p.directory)
		}
		return loom.Handled()
	}
	if e.Is("backspace") {
		if p.list.Query() != "" {
			result := p.list.ConsumeKey(e)
			p.notifySelection()
			return result
		}
		parent, ok := p.directory.Parent()
		if !ok || parent == p.directory.Path {
			return loom.Ignored()
		}
		if err := p.open(parent, filepath.Base(p.directory.Path)); err != nil {
			return loom.Ignored()
		}
		if p.options.OnOpen != nil {
			p.options.OnOpen(p.directory)
		}
		return loom.Handled()
	}
	if e.Is("enter") {
		return p.list.ConsumeKey(e)
	}
	result := p.list.ConsumeKey(e)
	p.notifySelection()
	return result
}

// ConsumeMouse processes mouse selection and forwards quit requests.
func (p *NavigationPane) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	result := p.list.ConsumeMouse(e)
	p.notifySelection()
	return result
}

// Searching reports whether typed text currently goes to the filter.
func (p *NavigationPane) Searching() bool { return p.searching || p.options.TypeToSearch }

func (p *NavigationPane) startsSearch(e loom.KeyEvent) bool {
	return !p.options.TypeToSearch && !p.searching && e.Is("/")
}

// handleSearchKey gates the filter behind "/": outside search, typed text is
// left to the app; inside, Esc clears the query and Enter keeps it.
func (p *NavigationPane) handleSearchKey(e loom.KeyEvent) (handled, quit bool) {
	switch {
	case p.options.TypeToSearch:
		return false, false
	case p.startsSearch(e):
		p.searching = true
		return true, false
	case e.Is("esc") && p.list.Query() != "":
		for p.list.Query() != "" {
			p.list.ConsumeKey(loom.KeyEvent{Key: "backspace"})
		}
		p.searching = false
		p.notifySelection()
		return true, false
	case !p.searching:
		return false, false
	case e.Is("esc"):
		for p.list.Query() != "" {
			p.list.ConsumeKey(loom.KeyEvent{Key: "backspace"})
		}
		p.searching = false
		p.notifySelection()
		return true, false
	case e.Is("enter"):
		p.searching = false
		return false, false
	case e.Is("backspace") && p.list.Query() == "":
		// An empty search closes, and backspace navigates up as usual.
		p.searching = false
	}
	return false, false
}
