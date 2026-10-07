// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"ubunatic.com/loom"
	"ubunatic.com/loom/gallery"
	"ubunatic.com/loom/spec"
)

type widgetCatalog struct {
	Widgets    []widgetEntry `yaml:"widgets"`
	Exclusions []struct {
		Name   string `yaml:"name"`
		Reason string `yaml:"reason"`
	} `yaml:"exclusions"`
}

type widgetEntry struct {
	Name         string   `yaml:"name"`
	Package      string   `yaml:"package"`
	Purpose      string   `yaml:"purpose"`
	Category     string   `yaml:"category"`
	Capabilities []string `yaml:"capabilities"`
	Example      string   `yaml:"example"`
	Source       string   `yaml:"source"`
	Docs         string   `yaml:"docs"`
	Ticket       string   `yaml:"ticket"`
}

func widgetsCommand() *cobra.Command {
	var show, list, debug bool
	var themeName string
	var width, height int
	command := &cobra.Command{
		Use:   "widgets [name] [-- <demo-args>]",
		Short: "List library widgets, show usage, or run live demos",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeWidgetNames(toComplete), cobra.ShellCompDirectiveNoFileComp
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if show {
				return nil
			}
			return cobra.MaximumNArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if show {
				if !loom.ThemeExists(themeName) {
					return fmt.Errorf("unknown theme %q (available: %s)", themeName, strings.Join(loom.ThemeNames(), ", "))
				}
				previousDebug := loom.Debug
				loom.Debug = debug
				defer func() { loom.Debug = previousDebug }()

				dash := cmd.ArgsLenAtDash()
				var names, demoArgs []string
				if dash != -1 {
					names = args[:dash]
					demoArgs = args[dash:]
				} else {
					names = args
				}
				return showWidgetDemos(names, demoArgs, themeName, width, height)
			}
			catalog, err := readWidgetCatalog()
			if err != nil {
				return err
			}
			if len(args) == 0 || list {
				categories := []string{"input", "display", "layout", "infra"}
				for _, cat := range categories {
					for _, entry := range catalog.Widgets {
						if entry.Category == cat {
							fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] — %s\n", entry.Name, entry.Category, entry.Purpose)
						}
					}
				}
				return nil
			}
			name := args[0]
			var matches []widgetEntry
			for _, entry := range catalog.Widgets {
				if name == entry.Name || name == strings.TrimPrefix(entry.Name, entry.Package+".") {
					matches = append(matches, entry)
				}
			}
			switch len(matches) {
			case 0:
				return fmt.Errorf("unknown widget %q", name)
			case 1:
				writeWidget(cmd.OutOrStdout(), matches[0])
				return nil
			}
			names := make([]string, len(matches))
			for i, entry := range matches {
				names[i] = entry.Name
			}
			return fmt.Errorf("widget %q is ambiguous: %s", name, strings.Join(names, ", "))
		},
	}
	command.Flags().BoolVar(&list, "list", false, "print the widget catalog and exit")
	command.Flags().BoolVar(&show, "show", false, "run live widget demos")
	command.Flags().BoolVar(&debug, "debug", false, "show debug cell outlines in live widget demos")
	command.Flags().StringVar(&themeName, "theme", loom.SpeccedDefaults.Editor.Theme, "gallery color theme")
	command.Flags().IntVarP(&width, "width", "W", 0, "maximum gallery width in columns")
	command.Flags().IntVarP(&height, "height", "H", 0, "gallery height in rows")
	_ = command.RegisterFlagCompletionFunc("show", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeWidgetNames(toComplete), cobra.ShellCompDirectiveNoFileComp
	})
	return command
}

func completeWidgetNames(prefix string) []string {
	catalog, err := readWidgetCatalog()
	if err != nil {
		return nil
	}
	available := make(map[string]bool)
	for _, name := range gallery.Names() {
		available[name] = true
	}
	var names []string
	for _, entry := range catalog.Widgets {
		short := strings.TrimPrefix(entry.Name, entry.Package+".")
		if !available[short] {
			continue
		}
		for _, name := range []string{short, entry.Name} {
			if strings.HasPrefix(name, prefix) {
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)
	return names
}

var runWidgetPane = func(widget loom.Widget, width, height int, altScreen bool) error {
	pane, err := loom.New(24)
	if err != nil {
		return err
	}
	if width > 0 || height > 0 {
		pane.InlineOnly = true
		pane.SetScreenMode(loom.ScreenInline)
	}
	if altScreen {
		pane.SetScreenMode(loom.ScreenAlt)
	}
	pane.SetMaxCols(galleryPaneMaxCols(width))
	if height > 0 {
		pane.Resize(height)
	}
	pane.EnableMouse()
	pane.DisableDefaultQuit = true
	if g, ok := widget.(*themedGallery); ok {
		g.setBackground = func(b loom.Background, onRedraw bool) {
			pane.Background = b
			pane.BackgroundOnRedraw = onRedraw
		}
	}
	defer pane.Close()
	return pane.Run(widget)
}

func galleryPaneMaxCols(width int) int { return width }

func showWidgetDemos(names []string, demoArgs []string, themeName string, width, height int) error {
	var widget loom.Widget
	if len(names) == 0 {
		if len(demoArgs) > 0 {
			return fmt.Errorf("demo arguments after -- can only be passed when showing a single widget demo")
		}
		widget = gallery.NewAll()
	} else if len(names) == 1 {
		var err error
		widget, err = gallery.NewWithArgs(names[0], demoArgs)
		if err != nil {
			return err
		}
		return runWidgetPane(newThemedGallery(widget, themeName), width, height, gallery.RequiresAltScreen(names...))
	} else {
		if len(demoArgs) > 0 {
			return fmt.Errorf("demo arguments after -- can only be passed when showing a single widget demo")
		}
		tabs := make([]loom.Tab, 0, len(names))
		for _, name := range names {
			child, err := gallery.New(name)
			if err != nil {
				return err
			}
			tabs = append(tabs, loom.Tab{Title: strings.TrimPrefix(name, "loom."), Widget: child})
		}
		tabWidget := loom.NewTabs(tabs...)
		tabWidget.Vertical = true
		tabWidget.ArrowSwitch = false
		tabWidget.SetKeys(loom.TabsKeys{Previous: "shift-tab", Next: "tab"})
		widget = tabWidget
	}
	return runWidgetPane(newThemedGallery(widget, themeName), width, height, gallery.RequiresAltScreen(names...))
}

type themedGallery struct {
	widget     loom.Widget
	themeName  string
	themeIndex int
	themes     []string
	hintBar    *loom.HintBar
	controls   *loom.HintBar
	hintRow    int
	controlRow int

	bgIndex int
	// setBackground applies a background to the running pane; nil in tests.
	setBackground func(loom.Background, bool)
}

var galleryBackgrounds = []struct {
	name     string
	onRedraw bool
	make     func() loom.Background
}{
	{"plain", false, func() loom.Background { return nil }},
	{"astra", false, func() loom.Background { return loom.NewAstraBackground() }},
	{"astra (on redraw)", true, func() loom.Background { return loom.NewAstraBackground() }},
}

func newThemedGallery(widget loom.Widget, themeName string) *themedGallery {
	themes := loom.ThemeNames()
	index := 0
	for i, name := range themes {
		if name == themeName {
			index = i
			break
		}
	}
	g := &themedGallery{widget: widget, themeName: themeName, themeIndex: index, themes: themes}
	g.applyTheme()
	return g
}

func (g *themedGallery) applyTheme() {
	if themeable, ok := loom.UnwrapWidget(g.widget).(loom.Themeable); ok {
		themeable.ApplyTheme(loom.Theme(g.themeName))
	}
}

func (g *themedGallery) Unwrap() loom.Widget { return g.widget }

func (g *themedGallery) Draw(c *loom.Canvas, r loom.Rect) {
	theme := loom.Theme(g.themeName)
	c.PaintSurface(r, loom.Style{FG: theme.NormalFG.Color(), BG: theme.NormalBG.Color()})
	content := r
	hintWidget, hasHints := g.widget.(interface{ HotkeyHint(int) string })
	structured, hasBar := g.widget.(interface{ HotkeyBar() *loom.HintBar })
	hasHints = hasHints || hasBar
	g.hintBar, g.controls = nil, nil
	content.H = max(0, r.H-1)
	if hasHints && r.H >= 2 {
		content.H--
	}
	g.widget.Draw(c, content)
	if r.H == 0 {
		return
	}
	g.controlRow = r.H - 1
	if hasHints && r.H >= 2 {
		g.hintRow = r.H - 2
		if hasBar {
			g.hintBar = structured.HotkeyBar()
			g.hintBar.ApplyTheme(theme)
			g.hintBar.Draw(c, loom.Rect{X: r.X, Y: r.Y + g.hintRow, W: r.W, H: 1})
		} else {
			style := theme.HintBarStyle().Label
			c.Write(r.X+1, r.Y+g.hintRow, loom.TruncateText(hintWidget.HotkeyHint(max(0, r.W-2)), max(0, r.W-2), ""), style)
		}
	}
	entry := func(key, binding, label, detail string) loom.HintEntry {
		return loom.HintEntry{Key: key, Binding: binding, Label: label, Detail: detail, Action: func() loom.EventResult { return g.ConsumeKey(loom.KeyEvent{Key: binding}) }}
	}
	g.controls = loom.NewHintBar(entry("F8", "f8", "BG", galleryBackgrounds[g.bgIndex].name), entry("F9", "f9", "Theme", g.themeName), entry("F10", "f10", "Quit", ""))
	g.controls.ApplyTheme(theme)
	g.controls.Draw(c, loom.Rect{X: r.X, Y: r.Y + g.controlRow, W: r.W, H: 1})
}

func (g *themedGallery) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	switch {
	case e.Is("f10", "ctrl-q"):
		return loom.QuitResult()
	case e.Is("f9"):
		g.themeIndex = (g.themeIndex + 1) % len(g.themes)
		g.themeName = g.themes[g.themeIndex]
		g.applyTheme()
		return loom.Handled()
	case e.Is("f8"):
		g.bgIndex = (g.bgIndex + 1) % len(galleryBackgrounds)
		if g.setBackground != nil {
			bg := galleryBackgrounds[g.bgIndex]
			g.setBackground(bg.make(), bg.onRedraw)
		}
		return loom.Handled()
	}
	return g.widget.ConsumeKey(e)
}

func (g *themedGallery) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	for _, row := range []struct {
		y   int
		bar *loom.HintBar
	}{{g.hintRow, g.hintBar}, {g.controlRow, g.controls}} {
		if row.bar != nil && e.Y == row.y {
			local := e
			local.Y = 0
			return row.bar.ConsumeMouse(local)
		}
	}
	return g.widget.ConsumeMouse(e)
}

func readWidgetCatalog() (widgetCatalog, error) {
	data, err := spec.WidgetsYAML()
	if err != nil {
		return widgetCatalog{}, fmt.Errorf("read embedded widget catalog: %w", err)
	}
	var catalog widgetCatalog
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return widgetCatalog{}, fmt.Errorf("decode embedded widget catalog: %w", err)
	}
	return catalog, nil
}

func writeWidget(out io.Writer, entry widgetEntry) {
	fmt.Fprintf(out, "%s [%s] — %s\n", entry.Name, entry.Category, entry.Purpose)
	if len(entry.Capabilities) > 0 {
		fmt.Fprintf(out, "Capabilities: %s\n", strings.Join(entry.Capabilities, "; "))
	}
	fmt.Fprintln(out, "Example:")
	for _, line := range strings.Split(entry.Example, "\n") {
		fmt.Fprintf(out, "  %s\n", line)
	}
	fmt.Fprintf(out, "Source: %s\n", entry.Source)
	if entry.Docs != "" {
		fmt.Fprintf(out, "Docs: %s\n", entry.Docs)
	}
	if entry.Ticket != "" {
		fmt.Fprintf(out, "Ticket: %s\n", entry.Ticket)
	}
}
