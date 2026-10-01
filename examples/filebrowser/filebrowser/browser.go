package filebrowser

import (
	"fmt"
	"os"
	"strings"
	"time"

	"ubunatic.com/loom"
)

type browser struct {
	frame      *loom.Frame
	navigation *NavigationPane
	list       *loom.Choice
	details    *detailView
	dir        string
	notice     string
	themeName  string
	theme      loom.ThemeColors
	openFile   func(string) error
	background loom.AnimatedBackground
	quit       bool
}

type detailView struct {
	*loom.View
	focused bool
}

func (v *detailView) Focused() bool         { return v.focused }
func (v *detailView) SetFocus(focused bool) { v.focused = focused }

func newBrowser(path, themeName string, theme loom.ThemeColors) (*browser, error) {
	b := &browser{
		dir:        path,
		details:    &detailView{View: loom.NewView(nil)},
		openFile:   loom.OpenFile,
		background: loom.NewAstraBackground(),
	}
	border := loom.BoxBorder{
		TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
		Horizontal: "─", Vertical: "│", TitlePrefix: " ", TitleSuffix: " ",
	}
	navigation, err := NewNavigationPane(path, NavigationPaneOptions{
		Style:        theme.ChoiceStyle(),
		SelectParent: true,
		OnSelection:  func(entry loom.FileEntry) { b.updateDetails(entry.Path) },
		OnActivate:   b.activate,
		OnOpen: func(directory loom.Directory) {
			b.dir = directory.Path
			b.notice = ""
		},
		OnQuit: func() { b.quit = true },
	})
	if err != nil {
		return nil, err
	}
	navigation.SetRoot("") // Filebrowser allows navigating up without quitting on ESC.
	b.navigation = navigation
	b.list = navigation.List()
	b.dir = navigation.Directory().Path

	b.frame = &loom.Frame{
		Gap: 1, Breakpoint: 65,
		Status: "Tab pane  •  ↑↓ select  •  Enter open  •  F9 theme  •  F10 Quit",
		Boxes: []loom.Box{
			{ID: "files", Dynamic: true, FillHeight: true, MinWidth: 20, Height: 4, Border: border, Child: b.navigation},
			{ID: "metadata", Dynamic: true, FillHeight: true, MinWidth: 25, Height: 4, Border: border, Child: b.details},
		},
		Actions: []loom.FrameAction{
			{ID: "quit_f10", Action: "quit", Key: "f10"},
			{ID: "quit_ctrl_q", Action: "quit", Key: "ctrl-q"},
		},
	}
	b.applyNamedTheme(themeName, theme)
	if entry, ok := b.navigation.Selected(); ok {
		b.updateDetails(entry.Path)
	}
	return b, nil
}

func (b *browser) applyNamedTheme(name string, theme loom.ThemeColors) {
	b.themeName, b.theme = name, theme
	b.details.Style = theme.ChoiceStyle().Normal
	b.details.Scrollbar = theme.ScrollbarStyle()
	b.frame.Style = theme.FrameStyle()
	for i := range b.frame.Boxes {
		b.frame.Boxes[i].Style = theme.BoxStyle()
	}
	if b.navigation != nil {
		b.navigation.ApplyTheme(theme)
	}
}

// ApplyTheme updates the browser's theme from ThemeColors.
// If an exact match is found in SpeccedThemes, the name is updated;
// otherwise the current name is kept so F9 cycling stays valid.
func (b *browser) ApplyTheme(theme loom.ThemeColors) {
	// Find the name by matching the theme's colors
	for name, t := range loom.SpeccedThemes {
		if t == theme {
			b.applyNamedTheme(name, theme)
			return
		}
	}
	// If exact match not found, keep the current name and apply the theme
	b.applyNamedTheme(b.themeName, theme)
}

func (b *browser) applyTheme(name string, theme loom.ThemeColors) {
	b.applyNamedTheme(name, theme)
}

func (b *browser) cycleTheme() {
	names := themeNames()
	for i, name := range names {
		if name == b.themeName {
			next := names[(i+1)%len(names)]
			b.applyNamedTheme(next, loom.Theme(next))
			return
		}
	}
}

func (b *browser) activate(entry loom.FileEntry) {
	path := entry.Path
	if entry.Kind == loom.FileKindRegular {
		if err := b.openFile(path); err != nil {
			b.notice = "Open failed: " + err.Error()
		} else {
			b.notice = "Opening file"
		}
	} else if entry.Kind == loom.FileKindSymlink {
		info, err := os.Stat(path)
		if err == nil && info.Mode().IsRegular() {
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
	b.updateDetails(path)
}

func (b *browser) updateDetails(path string) {
	if path == "" {
		b.details.Lines = []string{"No matching file"}
		b.details.Scroll = 0
		return
	}
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
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		c.ComposeBackground(b.background, r, time.Now())
	}
	b.frame.Title = "Browse " + b.dir
	b.frame.Status = fmt.Sprintf("Tab pane  •  ↑↓ select  •  Enter open  •  F9 theme:%s  •  F10 Quit", b.themeName)
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

func (b *browser) syncDir() {
	if b.navigation != nil && b.dir != "" && b.dir != b.navigation.Directory().Path {
		_ = b.navigation.open(b.dir, "")
	}
}

func (b *browser) ConsumeKey(k loom.KeyEvent) loom.EventResult {
	key := k.Key
	if key == "" {
		key = k.Text
	}
	if key == "f10" || key == "ctrl-q" {
		return loom.QuitResult()
	}
	if key == "f9" {
		b.cycleTheme()
		return loom.Handled()
	}
	b.syncDir()
	if (k.Text != "" && b.navigation.Searching()) || key == "/" || key == "esc" || key == "backspace" {
		result := b.navigation.ConsumeKey(k)
		b.dir = b.navigation.Directory().Path
		if result.Quit || b.quit {
			return loom.QuitResult()
		}
		if result.Consumed {
			return result
		}
	}
	if key == "ctrl-c" || key == "ctrl-d" || key == "esc" {
		return loom.QuitResult()
	}
	if isBrowserListKey(key) {
		result := b.frame.ConsumeKey(k)
		b.dir = b.navigation.Directory().Path
		if b.quit {
			return loom.QuitResult()
		}
		return result
	}
	return loom.Ignored()
}

func isBrowserListKey(key string) bool {
	switch strings.ToLower(key) {
	case "up", "down", "left", "right", "pgup", "pgdown", "pgdn", "pageup", "pagedown", "home", "end", "backspace", "enter", "tab", "shift-tab":
		return true
	default:
		return false
	}
}

// PaneRequest declares the terminal capabilities required by the browser.
func (b *browser) PaneRequest() loom.PaneRequest {
	return loom.PaneRequest{Mouse: 1000, Resizeable: true, MaxCols: 0}
}

func (b *browser) ConsumeMouse(k loom.MouseEvent) loom.EventResult {
	result := b.frame.ConsumeMouse(k)
	b.dir = b.navigation.Directory().Path
	if b.quit {
		return loom.QuitResult()
	}
	return result
}
