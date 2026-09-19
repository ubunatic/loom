package filebrowser

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
	"unicode"

	"codeberg.org/ubunatic/loom"
)

type browser struct {
	frame     *loom.Frame
	list      *loom.Choice
	details   *detailView
	dir       string
	paths     map[string]string
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
	dir, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}
	b := &browser{dir: dir, details: &detailView{View: loom.NewView(nil)}, openFile: launchFile}
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
	if err := b.open(dir); err != nil {
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

func (b *browser) open(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	items := make([]loom.Item, 0, len(entries)+1)
	paths := make(map[string]string, len(entries)+1)
	if parent := filepath.Dir(dir); parent != dir {
		items = append(items, loom.Item{Name: "..", Desc: "parent directory"})
		paths[".."] = parent
	}
	for _, entry := range entries {
		name := displayName(entry.Name())
		desc := ""
		if entry.IsDir() {
			desc = "<dir>"
		} else if entry.Type()&os.ModeSymlink != 0 {
			desc = "<link>"
		}
		items = append(items, loom.Item{Name: name, Desc: desc})
		paths[name] = filepath.Join(dir, entry.Name())
	}
	list := loom.NewChoice(items)
	list.Style = b.theme.ChoiceStyle()
	list.SelectOnlyOnClick = true
	list.Prompt = "filter> "
	list.Placeholder = "type to filter"
	list.OnSelect = func(item loom.Item) {
		path := paths[item.Name]
		info, err := os.Stat(path)
		if err != nil {
			b.notice = "Error: " + err.Error()
			return
		}
		if info.IsDir() {
			if err := b.open(path); err != nil {
				b.notice = "Error: " + err.Error()
			}
		} else if info.Mode().IsRegular() {
			if err := b.openFile(path); err != nil {
				b.notice = "Open failed: " + err.Error()
			} else {
				b.notice = "Opening with xdg-open"
			}
		} else {
			b.notice = "Cannot open this file type"
		}
	}
	b.dir, b.paths, b.list, b.notice = dir, paths, list, ""
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

// launchFile lets the desktop choose a viewer without holding up the TUI.
func launchFile(path string) error {
	cmd := exec.Command("xdg-open", path)
	cmd.Stdin = nil
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
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
		"Name: " + displayName(info.Name()),
		"Path: " + displayName(path),
		"Type: " + kind,
		fmt.Sprintf("Size: %d bytes", info.Size()),
		"Mode: " + info.Mode().String(),
		"Modified: " + info.ModTime().Format(time.RFC3339),
	}
	if info.Mode()&os.ModeSymlink != 0 {
		if target, err := os.Readlink(path); err == nil {
			lines = append(lines, "Target: "+displayName(target))
		}
	}
	return lines
}

func displayName(value string) string {
	for _, r := range value {
		if !unicode.IsPrint(r) {
			return strconv.QuoteToASCII(value)
		}
	}
	return value
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
