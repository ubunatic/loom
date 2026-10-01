// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package gallery provides live, sample-data demos for Loom widgets.
package gallery

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strings"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/media"
)

type constructor func() loom.Widget

var demos = map[string]constructor{
	"Chart": func() loom.Widget {
		return &loom.Chart{Series: []loom.ChartSeries{
			{Name: "Requests", Values: []float64{12, 18, 14, 26, 22, 31, 27}},
			{Name: "Errors", Values: []float64{2, 4, 3, 6, 5, 8, 4}},
		}}
	},
	"DatePicker": func() loom.Widget {
		selected := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
		picker := loom.NewDatePicker(&selected)
		picker.Now = func() time.Time { return time.Date(2024, time.January, 20, 0, 0, 0, 0, time.UTC) }
		return picker
	},
	"Choice": func() loom.Widget {
		choice := loom.NewChoice([]loom.Item{
			{Name: "filebrowser-widget", Desc: "Browser widget implementation"},
			{Name: "fuzzy-browser-widget", Desc: "Fuzzy matching example"},
			{Name: "frame-layout", Desc: "Frame composition"},
		})
		choice.Fuzzy = true
		for _, r := range "fbw" {
			choice.ConsumeKey(loom.KeyEvent{Text: string(r)})
		}
		return choice
	},
	"Dialog": func() loom.Widget {
		dialog := loom.NewDialog("Save changes", "Keep your edits before closing?", "Discard", "Save")
		dialog.Width, dialog.Height = 50, 7
		return &dialogDemo{dialog: dialog}
	},
	"FilePicker": func() loom.Widget {
		picker, err := loom.NewFilePicker(".", loom.FilePickerOptions{Mode: loom.FilePickerFiles, Patterns: []string{"*.go", "*.md"}})
		if err != nil {
			return loom.NewView([]string{"FilePicker demo unavailable", err.Error()})
		}
		return picker
	},
	"Form": func() loom.Widget {
		name := loom.NewTextInput("Ada Lovelace")
		role := loom.NewTextInput("Engineer")
		return loom.NewForm([]loom.FormField{
			{Label: "Name", Widget: name, Required: true, Help: "Your display name"},
			{Label: "Role", Widget: role},
		})
	},
	"KeyHelp": func() loom.Widget {
		keymap := loom.NewKeyMapWithLabels(map[string][]string{
			"back": {"esc"},
			"next": {"j", "down"},
			"open": {"enter"},
		}, map[string]string{
			"back": "Back",
			"next": "Next",
			"open": "Open",
		})
		return loom.NewKeyHelp(keymap)
	},
	"MenuBar": func() loom.Widget {
		checked := true
		bar := loom.NewMenuBar(
			loom.Menu{Title: "File", Mnemonic: 'F', Items: []loom.MenuItem{
				{Label: "Open", Shortcut: "Ctrl+O"},
				{Label: "Save", Shortcut: "Ctrl+S"},
				{Label: "Autosave", Checked: &checked},
				{Label: "Recent files", Submenu: []loom.MenuItem{{Label: "notes.txt"}, {Label: "project.go"}}},
			}},
			loom.Menu{Title: "Edit", Mnemonic: 'E', Items: []loom.MenuItem{{Label: "Undo", Shortcut: "Ctrl+Z"}, {Label: "Redo", Shortcut: "Ctrl+Y"}}},
			loom.Menu{Title: "Help", Mnemonic: 'H', Items: []loom.MenuItem{{Label: "Keyboard shortcuts", Shortcut: "F1"}}},
		)
		bar.Open = true
		return bar
	},
	"Media": func() loom.Widget {
		img := image.NewRGBA(image.Rect(0, 0, 64, 32))
		for y := 0; y < 32; y++ {
			for x := 0; x < 64; x++ {
				img.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 7), B: 180, A: 255})
			}
		}
		widget, err := media.NewImage(img, media.ModeHalfblock)
		if err != nil {
			return loom.NewView([]string{"Media demo unavailable", err.Error()})
		}
		return widget
	},
	"PillCluster": func() loom.Widget {
		return loom.NewPillCluster(
			loom.ProviderPill{Name: "API", Symbol: "✓", State: loom.ProviderDone},
			loom.ProviderPill{Name: "Worker", Symbol: "…", State: loom.ProviderFetching},
			loom.ProviderPill{Name: "Cache", Symbol: "✓", State: loom.ProviderDone},
		)
	},
	"NumberInput": func() loom.Widget {
		value := 7.5
		input := loom.NewNumberInput(&value, -100, 100)
		input.Step = .5
		input.Format, input.FixedWidth, input.Align = "%.2f", 11, loom.AlignRight
		return input
	},
	"Paginator": func() loom.Widget {
		paginator := loom.NewPaginator(7)
		paginator.SetPage(2)
		return paginator
	},
	"PaintCanvas": func() loom.Widget {
		canvas := &loom.PaintCanvas{Controls: true}
		canvas.Draw(loom.NewCanvas(32, 8), loom.Rect{W: 32, H: 8})
		canvas.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 4, Y: 2})
		canvas.ConsumeMouse(loom.MouseEvent{Action: loom.MouseDrag, Button: loom.MouseLeft, X: 25, Y: 5})
		canvas.ConsumeMouse(loom.MouseEvent{Action: loom.MouseRelease, Button: loom.MouseLeft, X: 25, Y: 5})
		return canvas
	},
	"Popup": func() loom.Widget {
		popup := loom.NewPopup("Gallery popup", loom.NewView([]string{"This overlay is a live widget.", "Press Esc to close it."}))
		popup.Width, popup.Height = 48, 7
		return &popupDemo{popup: popup}
	},
	"ProgressBar": func() loom.Widget {
		bar := loom.NewProgressBar()
		bar.Options.Width = 18
		bar.ShowPercent = true
		bar.ShowCount = true
		bar.Unit = " files"
		return &progressDemo{ProgressBar: bar}
	},
	"Spinner": func() loom.Widget {
		spinner := loom.NewSpinner("Syncing workspace")
		spinner.Start()
		return spinner
	},
	"Stopwatch": func() loom.Widget {
		watch := loom.NewStopwatch()
		watch.Controls = true
		watch.Start()
		return watch
	},
	"Timer": func() loom.Widget {
		timer := loom.NewTimer(4*time.Minute + 12*time.Second)
		timer.Controls = true
		timer.Start()
		return timer
	},
	"Tree": func() loom.Widget {
		return loom.NewTree([]*loom.TreeNode{
			{ID: "src", Label: "src", Expanded: true, Children: []*loom.TreeNode{
				{ID: "app", Label: "app.go"},
				{ID: "tree", Label: "tree.go"},
			}},
			{ID: "docs", Label: "docs", Children: []*loom.TreeNode{{ID: "widgets", Label: "Widgets.md"}}},
			{ID: "go.mod", Label: "go.mod"},
		})
	},
	"Table": func() loom.Widget {
		table := loom.NewTable(
			[]loom.Column{{Header: "Task", Width: 18}, {Header: "Status", Width: 12}, {Header: "Owner", Width: 12}},
			[]loom.Row{{Cells: []string{"Compile", "done", "Ada"}, Key: "compile"}, {Cells: []string{"Unit tests", "running", "Lin"}, Key: "tests"}, {Cells: []string{"Package", "waiting", "Sam"}, Key: "package"}},
		)
		table.CellCursor = true
		table.FrozenCols = 1
		table.Controls = "←/→ cell  ↑/↓ row"
		return table
	},
	"Tabs": func() loom.Widget {
		tabs := loom.NewTabs(
			loom.Tab{Title: "Overview", Widget: loom.NewView([]string{"Loom widget gallery", "Click a tab or use Tab / Shift-Tab."})},
			loom.Tab{Title: "Details", Widget: loom.NewView([]string{"Tabs host any Loom widgets."})},
		)
		tabs.Vertical = true
		tabs.ArrowSwitch = false
		tabs.SetKeys(loom.TabsKeys{Previous: "shift-tab", Next: "tab"})
		return tabs
	},
	"TextArea": func() loom.Widget {
		return &textAreaWidget{area: loom.NewTextArea("A multi-line editor\nwith sample content.\nUse the arrow keys to move.")}
	},
	"TextInput": func() loom.Widget {
		input := loom.NewTextInput("Ada Lovelace")
		placeholder := loom.NewTextInput("")
		placeholder.Placeholder = "Type a value…"
		masked := loom.NewTextInput("correct horse")
		masked.Mask = '•'
		form := loom.NewForm([]loom.FormField{
			{Label: "Name:", Widget: input},
			{Label: "Placeholder:", Widget: placeholder},
			{Label: "Masked:", Widget: masked},
		})
		form.AdvanceOnEnter = true
		return form
	},
	"Toggle": func() loom.Widget {
		value := true
		toggle := loom.NewToggle(&value)
		toggle.Label = "Notifications"
		return toggle
	},
	"Viewport": func() loom.Widget {
		lines := make([]string, 30)
		for i := range lines {
			lines[i] = fmt.Sprintf("%02d  scrollable child content", i+1)
		}
		return loom.NewViewport(loom.NewView(lines))
	},
}

// Names returns the available demo names in sorted order.
func Names() []string {
	names := make([]string, 0, len(demos))
	for name := range demos {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// New constructs one demo by catalog widget name, accepting qualified names
// such as "loom.TextInput" as well as short names.
func New(name string) (loom.Widget, error) {
	name = strings.TrimPrefix(name, "loom.")
	if strings.EqualFold(name, "all") {
		return newAllDemo(), nil
	}
	build, ok := demos[name]
	if !ok {
		return nil, fmt.Errorf("unknown widget demo %q", name)
	}
	return build(), nil
}

// NewAll constructs the complete gallery as a tabbed widget.
func NewAll() *loom.Tabs {
	tabs := make([]loom.Tab, 0, len(demos)+1)
	tabs = append(tabs, loom.Tab{Title: "All", Widget: newAllDemo()})
	for _, name := range Names() {
		tabs = append(tabs, loom.Tab{Title: name, Widget: demos[name]()})
	}
	all := loom.NewTabs(tabs...)
	all.Vertical = true
	all.ArrowSwitch = false
	all.SetKeys(loom.TabsKeys{Previous: "shift-tab", Next: "tab"})
	return all
}

type buttonDemo struct {
	label   string
	clicked int
	focused bool
}

func (b *buttonDemo) Focused() bool   { return b.focused }
func (b *buttonDemo) SetFocus(f bool) { b.focused = f }

func (b *buttonDemo) Draw(c *loom.Canvas, r loom.Rect) {
	c.PaintSurface(r, loom.Style{})
	text := fmt.Sprintf("[ %s ]", b.label)
	if b.clicked > 0 {
		text = fmt.Sprintf("[ %s (%d) ]", b.label, b.clicked)
	}
	style := loom.Style{}
	if b.focused {
		style = loom.Style{Bold: true}
	}
	c.Write(r.X, r.Y, text, style)
}

func (b *buttonDemo) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if e.Key == "enter" || e.Key == "space" || e.Text == " " {
		b.clicked++
		return loom.Handled()
	}
	return loom.Ignored()
}

func (b *buttonDemo) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	text := fmt.Sprintf("[ %s ]", b.label)
	if b.clicked > 0 {
		text = fmt.Sprintf("[ %s (%d) ]", b.label, b.clicked)
	}
	width := loom.StringWidth(text)
	if e.Action == loom.MousePress && e.Button == loom.MouseLeft && e.Y == 0 && e.X >= 0 && e.X < width {
		b.clicked++
		return loom.Handled()
	}
	return loom.Ignored()
}

type checkboxDemo struct {
	label   string
	checked bool
	focused bool
}

func (cb *checkboxDemo) Focused() bool   { return cb.focused }
func (cb *checkboxDemo) SetFocus(f bool) { cb.focused = f }

func (cb *checkboxDemo) Draw(c *loom.Canvas, r loom.Rect) {
	c.PaintSurface(r, loom.Style{})
	mark := "[ ]"
	if cb.checked {
		mark = "[x]"
	}
	text := mark + " " + cb.label
	style := loom.Style{}
	if cb.focused {
		style = loom.Style{Bold: true}
	}
	c.Write(r.X, r.Y, text, style)
}

func (cb *checkboxDemo) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if e.Key == "enter" || e.Key == "space" || e.Text == " " {
		cb.checked = !cb.checked
		return loom.Handled()
	}
	return loom.Ignored()
}

func (cb *checkboxDemo) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	width := loom.StringWidth("[x] " + cb.label)
	if e.Action == loom.MousePress && e.Button == loom.MouseLeft && e.Y == 0 && e.X >= 0 && e.X < width {
		cb.checked = !cb.checked
		return loom.Handled()
	}
	return loom.Ignored()
}

type badgeDemo struct {
	label   string
	status  string
	focused bool
}

func (b *badgeDemo) Focused() bool   { return b.focused }
func (b *badgeDemo) SetFocus(f bool) { b.focused = f }

func (b *badgeDemo) Draw(c *loom.Canvas, r loom.Rect) {
	c.PaintSurface(r, loom.Style{})
	text := fmt.Sprintf("[%s: %s ●]", b.label, b.status)
	style := loom.Style{}
	if b.focused {
		style = loom.Style{Bold: true}
	}
	c.Write(r.X, r.Y, text, style)
}

func (b *badgeDemo) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if e.Key == "enter" || e.Key == "space" || e.Text == " " {
		if b.status == "Active" {
			b.status = "Idle"
		} else {
			b.status = "Active"
		}
		return loom.Handled()
	}
	return loom.Ignored()
}

func (b *badgeDemo) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	width := loom.StringWidth(fmt.Sprintf("[%s: %s ●]", b.label, b.status))
	if e.Action == loom.MousePress && e.Button == loom.MouseLeft && e.Y == 0 && e.X >= 0 && e.X < width {
		if b.status == "Active" {
			b.status = "Idle"
		} else {
			b.status = "Active"
		}
		return loom.Handled()
	}
	return loom.Ignored()
}

func newAllDemo() *loom.Grid {
	button := &buttonDemo{label: "Click Me"}
	toggleVal := true
	toggle := loom.NewToggle(&toggleVal)
	toggle.Label = "Toggle"
	checkbox := &checkboxDemo{label: "Checkbox", checked: true}

	numVal := 42.0
	numInput := loom.NewNumberInput(&numVal, -100, 100)
	numInput.Step = 1.0
	numInput.Format = "%.2f"
	numInput.FixedWidth = 10
	numInput.Align = loom.AlignRight

	badge := &badgeDemo{label: "Badge", status: "Active"}
	pillCluster := loom.NewPillCluster(
		loom.ProviderPill{Name: "API", Symbol: "✓", State: loom.ProviderDone},
		loom.ProviderPill{Name: "Job", Symbol: "…", State: loom.ProviderFetching},
		loom.ProviderPill{Name: "DB", Symbol: "✓", State: loom.ProviderDone},
	)

	bar := loom.NewProgressBar()
	bar.Options.Width = 12
	bar.ShowPercent = true
	bar.ShowCount = false
	progressBar := &progressDemo{ProgressBar: bar}

	sparkline := &loom.Sparkline{
		Values: []float64{12, 18, 14, 26, 22, 31, 27, 35, 42, 38},
		Width:  12,
	}

	spinner := loom.NewSpinner("Syncing")
	spinner.Start()

	watch := loom.NewStopwatch()
	watch.Start()

	timer := loom.NewTimer(4*time.Minute + 12*time.Second)
	timer.Start()

	paginator := loom.NewPaginator(5)
	paginator.SetPage(2)

	return loom.NewGrid(3,
		button, toggle, checkbox,
		numInput, badge, pillCluster,
		progressBar, sparkline, spinner,
		watch, timer, paginator,
	)
}

type textAreaWidget struct{ area *loom.TextArea }

// progressDemo drives a determinate bar through fill, dim, hidden, and restart.
// It retains its own ticking hook rather than exposing the determinate child
// through Unwrap, since a determinate ProgressBar deliberately does not tick.
type progressDemo struct {
	*loom.ProgressBar
	phase      int
	invalidate func()
}

func (p *progressDemo) TickInterval() time.Duration {
	return loom.SpeccedDefaults.ProgressBar.DemoInterval
}

func (p *progressDemo) SetInvalidate(fn func()) {
	p.invalidate = fn
	p.ProgressBar.SetInvalidate(fn)
}

func (p *progressDemo) Tick(time.Time) {
	switch p.phase {
	case 0:
		if p.Value() < p.Total {
			p.Set(p.Value() + loom.SpeccedDefaults.ProgressBar.DemoStep)
			return
		}
		p.phase = 1
	case 1:
		p.phase = 2
	case 2:
		p.phase = 0
		p.Reset()
		return
	}
	if p.invalidate != nil {
		p.invalidate()
	}
}

func (p *progressDemo) Draw(c *loom.Canvas, r loom.Rect) {
	if p.phase == 2 {
		c.PaintSurface(r, loom.Style{})
		return
	}
	p.ProgressBar.Draw(c, r)
	if p.phase == 1 && r.H > 0 {
		for x := r.X; x < r.X+r.W; x++ {
			cell := c.Get(x, r.Y)
			cell.Style.Dim = true
			c.Set(x, r.Y, cell)
		}
	}
}

func (w *textAreaWidget) Draw(c *loom.Canvas, r loom.Rect)              { w.area.Draw(c, r, true) }
func (w *textAreaWidget) ConsumeKey(e loom.KeyEvent) loom.EventResult   { return w.area.ConsumeKey(e) }
func (w *textAreaWidget) ConsumeMouse(loom.MouseEvent) loom.EventResult { return loom.Ignored() }

// The gallery wrappers reopen their sample after it has been dismissed.
// All active input is forwarded as the original EventResult value.
type popupDemo struct{ popup *loom.Popup }

func (w *popupDemo) ApplyTheme(theme loom.ThemeColors) { w.popup.ApplyTheme(theme) }

func (w *popupDemo) Draw(c *loom.Canvas, r loom.Rect) { w.popup.Draw(c, r) }
func (w *popupDemo) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if !w.popup.Open && e.Key == "enter" {
		w.popup.Open = true
		return loom.Handled()
	}
	return w.popup.ConsumeKey(e)
}
func (w *popupDemo) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	return w.popup.ConsumeMouse(e)
}

type dialogDemo struct{ dialog *loom.Dialog }

func (w *dialogDemo) Unwrap() loom.Widget { return w.dialog }

func (w *dialogDemo) ApplyTheme(theme loom.ThemeColors) { w.dialog.ApplyTheme(theme) }

func (w *dialogDemo) Draw(c *loom.Canvas, r loom.Rect) { w.dialog.Draw(c, r) }
func (w *dialogDemo) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if !w.dialog.Open && e.Key == "enter" {
		w.dialog.Open = true
		return loom.Handled()
	}
	return w.dialog.ConsumeKey(e)
}
func (w *dialogDemo) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	return w.dialog.ConsumeMouse(e)
}
