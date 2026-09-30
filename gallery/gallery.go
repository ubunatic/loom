// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package gallery provides live, sample-data demos for Loom widgets.
package gallery

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"codeberg.org/ubunatic/loom"
)

type constructor func() loom.Widget

var demos = map[string]constructor{
	"Choice": func() loom.Widget {
		choice := loom.NewChoice([]loom.Item{
			{Name: "filebrowser-widget", Desc: "Browser widget implementation"},
			{Name: "fuzzy-browser-widget", Desc: "Fuzzy matching example"},
			{Name: "frame-layout", Desc: "Frame composition"},
		})
		choice.Fuzzy = true
		for _, r := range "fbw" {
			choice.HandleKey(loom.KeyEvent{Text: string(r)})
		}
		return choice
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
	"PillCluster": func() loom.Widget {
		return loom.NewPillCluster(
			loom.ProviderPill{Name: "API", Symbol: "✓", State: loom.ProviderDone},
			loom.ProviderPill{Name: "Worker", Symbol: "…", State: loom.ProviderFetching},
			loom.ProviderPill{Name: "Cache", Symbol: "✓", State: loom.ProviderDone},
		)
	},
	"NumberInput": func() loom.Widget {
		value := 7.5
		input := loom.NewNumberInput(&value, 0, 10)
		input.Step = .5
		return input
	},
	"Paginator": func() loom.Widget {
		paginator := loom.NewPaginator(7)
		paginator.SetPage(2)
		return paginator
	},
	"Popup": func() loom.Widget {
		popup := loom.NewPopup("Gallery popup", loom.NewView([]string{"This overlay is a live widget.", "Press Esc to close it."}))
		popup.Width, popup.Height = 48, 7
		return popup
	},
	"ProgressBar": func() loom.Widget {
		bar := loom.NewProgressBar()
		bar.Options.Width = 18
		bar.Total = 24
		bar.ShowPercent = true
		bar.ShowCount = true
		bar.Unit = " files"
		bar.Set(16)
		return bar
	},
	"Spinner": func() loom.Widget {
		spinner := loom.NewSpinner("Syncing workspace")
		spinner.Start()
		return spinner
	},
	"Stopwatch": func() loom.Widget {
		watch := loom.NewStopwatch()
		watch.Formatter = func(time.Duration) string { return "02:37" }
		watch.Start()
		return watch
	},
	"Timer": func() loom.Widget {
		timer := loom.NewTimer(4*time.Minute + 12*time.Second)
		timer.Formatter = func(time.Duration) string { return "04:12" }
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
			[]loom.Column{{Header: "Task", Width: 24}, {Header: "Status", Width: 14}},
			[]loom.Row{{Cells: []string{"Compile", "done"}, Key: "compile"}, {Cells: []string{"Unit tests", "running"}, Key: "tests"}, {Cells: []string{"Package", "waiting"}, Key: "package"}},
		)
		return table
	},
	"Tabs": func() loom.Widget {
		return loom.NewTabs(
			loom.Tab{Title: "Overview", Widget: loom.NewView([]string{"Loom widget gallery", "Switch tabs with ← and →."})},
			loom.Tab{Title: "Details", Widget: loom.NewView([]string{"Tabs host any Loom widgets."})},
		)
	},
	"TextArea": func() loom.Widget {
		return &textAreaWidget{area: loom.NewTextArea("A multi-line editor\nwith sample content.\nUse the arrow keys to move.")}
	},
	"TextInput": func() loom.Widget {
		input := loom.NewTextInput("a long gallery value with 界 and 🙂")
		input.Prompt = "Name: "
		input.Mask = '•'
		return &textInputWidget{input: input}
	},
	"Toggle": func() loom.Widget {
		value := true
		return loom.NewToggle(&value)
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
	build, ok := demos[name]
	if !ok {
		return nil, fmt.Errorf("unknown widget demo %q", name)
	}
	return build(), nil
}

// NewAll constructs the complete gallery as a tabbed widget.
func NewAll() *loom.Tabs {
	tabs := make([]loom.Tab, 0, len(demos))
	for _, name := range Names() {
		tabs = append(tabs, loom.Tab{Title: name, Widget: demos[name]()})
	}
	return loom.NewTabs(tabs...)
}

type textInputWidget struct{ input *loom.TextInput }

func (w *textInputWidget) Draw(c *loom.Canvas, r loom.Rect) {
	c.PaintSurface(r, loom.Style{})
	w.input.Draw(c, r, true)
}
func (w *textInputWidget) HandleKey(e loom.KeyEvent) bool   { return w.input.HandleKey(e) }
func (w *textInputWidget) HandleMouse(loom.MouseEvent) bool { return false }

type textAreaWidget struct{ area *loom.TextArea }

func (w *textAreaWidget) Draw(c *loom.Canvas, r loom.Rect) { w.area.Draw(c, r, true) }
func (w *textAreaWidget) HandleKey(e loom.KeyEvent) bool   { return w.area.HandleKey(e) }
func (w *textAreaWidget) HandleMouse(loom.MouseEvent) bool { return false }
