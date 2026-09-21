package filebrowser

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"codeberg.org/ubunatic/loom"
)

type browser struct {
	frame     *loom.Frame
	list      *loom.Choice
	details   *detailView
	dir       string
	paths     map[string]string
	entries   map[string]loom.FileEntry
	notice    string
	themeName string
	theme     loom.ThemeColors
	openFile  func(string) error
}

type detailView struct {
	*loom.View
	focused bool
}

func (v *detailView) Focused() bool         { return v.focused }
func (v *detailView) SetFocus(focused bool) { v.focused = focused }

func newBrowser(path, themeName string, theme loom.ThemeColors) (*browser, error) {
	b := &browser{dir: path, details: &detailView{View: loom.NewView(nil)}, openFile: loom.OpenFile}
	border := loom.BoxBorder{
		TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
		Horizontal: "─", Vertical: "│", TitlePrefix: " ", TitleSuffix: " ",
	}
	b.frame = &loom.Frame{
		Gap: 1, Breakpoint: 65,
		Status: "Tab pane  •  ↑↓ select  •  Enter open  •  F9 theme  •  F10/^Q quit",
		Boxes: []loom.Box{
			{ID: "files", Dynamic: true, FillHeight: true, MinWidth: 20, Height: 18, Border: border},
			{ID: "metadata", Dynamic: true, FillHeight: true, MinWidth: 25, Height: 18, Border: border, Child: b.details},
		},
		Actions: []loom.FrameAction{
			{ID: "quit_f10", Action: "quit", Key: "f10"},
			{ID: "quit_ctrl_q", Action: "quit", Key: "ctrl-q"},
		},
	}
	b.applyTheme(themeName, theme)
	if err := b.open(path, ""); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *browser) applyTheme(name string, theme loom.ThemeColors) {
	b.themeName, b.theme = name, theme
	b.details.Style = theme.ChoiceStyle().Normal
	b.details.Scrollbar = theme.ScrollbarStyle()
	b.frame.Style = theme.FrameStyle()
	for i := range b.frame.Boxes {
		b.frame.Boxes[i].Style = theme.BoxStyle()
	}
	if b.list != nil {
		b.list.Style = theme.ChoiceStyle()
	}
}

// ApplyTheme updates the browser's theme from ThemeColors.
// If an exact match is found in SpeccedThemes, the name is updated;
// otherwise the current name is kept so F9 cycling stays valid.
func (b *browser) ApplyTheme(theme loom.ThemeColors) {
	// Find the name by matching the theme's colors
	for name, t := range loom.SpeccedThemes {
		if t == theme {
			b.applyTheme(name, theme)
			return
		}
	}
	// If exact match not found, keep the current name and apply the theme
	b.applyTheme(b.themeName, theme)
}

func (b *browser) cycleTheme() {
	names := themeNames()
	for i, name := range names {
		if name == b.themeName {
			next := names[(i+1)%len(names)]
			b.applyTheme(next, loom.Theme(next))
			return
		}
	}
}

func (b *browser) open(dir, selectName string) error {
	directory, err := loom.ReadDirectory(dir, loom.DirectoryOptions{IncludeParent: true})
	if err != nil {
		return err
	}
	items := make([]loom.Item, 0, len(directory.Entries))
	paths := make(map[string]string, len(directory.Entries))
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
		paths[name] = entry.Path
		entries[name] = entry
	}
	list := loom.NewChoice(items)
	list.Style = b.theme.ChoiceStyle()
	list.SelectOnlyOnClick = true
	list.Prompt = "filter> "
	list.Placeholder = "type to filter"
	list.OnSelect = func(item loom.Item) {
		entry, ok := entries[item.Name]
		if !ok {
			b.notice = "Error: entry disappeared"
			return
		}
		path := entry.Path
		if entry.Kind == loom.FileKindDirectory {
			nextSelection := ""
			if item.Name == ".." {
				nextSelection = loom.QuoteUnprintable(filepath.Base(dir))
			}
			if err := b.open(path, nextSelection); err != nil {
				b.notice = "Error: " + err.Error()
			}
		} else if entry.Kind == loom.FileKindRegular {
			if err := b.openFile(path); err != nil {
				b.notice = "Open failed: " + err.Error()
			} else {
				b.notice = "Opening file"
			}
		} else if entry.Kind == loom.FileKindSymlink {
			info, err := os.Stat(path)
			if err == nil && info.IsDir() {
				if err := b.open(path, ""); err != nil {
					b.notice = "Error: " + err.Error()
				}
			} else if err == nil && info.Mode().IsRegular() {
				if err := b.openFile(path); err != nil {
					b.notice = "Open failed: " + err.Error()
				} else {
					b.notice = "Opening file"
				}
			} else {
				b.notice = "Cannot open this file type"
			}
		} else {
			b.notice = "Cannot open this file type"
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
	b.dir, b.paths, b.entries, b.list, b.notice = directory.Path, paths, entries, list, ""
	b.frame.Boxes[0].Child = list
	b.updateDetails()
	return nil
}

func (b *browser) updateDetails() {
	item, ok := b.list.Selected()
	if !ok {
		b.details.Lines = []string{"No matching file"}
		b.details.Scroll = 0
		return
	}
	path := b.paths[item.Name]
	lines := metadata(path)
	if b.notice != "" {
		lines = append([]string{b.notice, ""}, lines...)
	}
	if !equalLines(lines, b.details.Lines) {
		b.details.Lines = lines
		b.details.Scroll = 0
	}
}

func equalLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func metadata(path string) []string {
	info, err := os.Lstat(path)
	if err != nil {
		return []string{"Error: " + err.Error()}
	}
	kind := "file"
	if info.IsDir() {
		kind = "directory"
	} else if info.Mode()&os.ModeSymlink != 0 {
		kind = "symbolic link"
	} else if !info.Mode().IsRegular() {
		kind = "special file"
	}
	lines := []string{
		"Name: " + loom.QuoteUnprintable(info.Name()),
		"Path: " + loom.DisplayPath(path),
		"Type: " + kind,
		fmt.Sprintf("Size: %d bytes", info.Size()),
		"Mode: " + info.Mode().String(),
		"Modified: " + info.ModTime().Format(time.RFC3339),
	}
	if info.Mode()&os.ModeSymlink != 0 {
		if target, err := os.Readlink(path); err == nil {
			lines = append(lines, "Target: "+loom.DisplayPath(target))
		}
	}
	return lines
}

func (b *browser) Draw(c *loom.Canvas, r loom.Rect) {
	b.frame.Title = "Browse " + b.dir
	b.frame.Status = fmt.Sprintf("Tab pane  •  ↑↓ select  •  Enter open  •  F9 theme:%s  •  F10/^Q quit", b.themeName)
	b.frame.Boxes[0].Title = "Files"
	b.frame.Boxes[1].Title = "Metadata"
	if focused := b.frame.FocusedBox(); focused != nil {
		if focused.ID == "files" {
			b.frame.Boxes[0].Title = "▶ Files"
		} else {
			b.frame.Boxes[1].Title = "▶ Metadata"
		}
	}
	b.frame.Draw(c, r)
}

func (b *browser) HandleKey(k loom.KeyEvent) bool {
	if k.Key == "f9" {
		b.cycleTheme()
		return false
	}
	quit := b.frame.HandleKey(k)
	b.updateDetails()
	return quit
}

func (b *browser) HandleMouse(k loom.MouseEvent) bool {
	quit := b.frame.HandleMouse(k)
	b.updateDetails()
	return quit
}
