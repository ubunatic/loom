// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// YamlSetting defines a configurable setting row inside a settings element.
type YamlSetting struct {
	Name    string   `yaml:"name"`
	Kind    string   `yaml:"kind"`
	Options []string `yaml:"options"`
	Default string   `yaml:"default"`
}

// YamlElement defines a single configurable component in the YAML schema.
type YamlElement struct {
	Type        string    `yaml:"type"`
	Prompt      string    `yaml:"prompt"`
	Placeholder string    `yaml:"placeholder"`
	Cursor      string    `yaml:"cursor"`
	Unicode     string    `yaml:"unicode"`
	MaxW        int       `yaml:"max_width"`
	Controls    string    `yaml:"controls"`
	Source      string    `yaml:"source"`
	Static      yaml.Node `yaml:"static"`
	OnChange    string    `yaml:"on_change"`
	StyleName   string    `yaml:"style"`
}

// YamlApp defines global TUI properties.
type YamlApp struct {
	Name   string `yaml:"name"`
	Mode   string `yaml:"mode"`
	Root   string `yaml:"root"`
	Height int    `yaml:"height"`
	MinH   int    `yaml:"min_height"`
	MaxH   int    `yaml:"max_height"`
	MaxW   int    `yaml:"max_width"`
}

// YamlView defines the layout and bindings for a single TUI screen pane.
type YamlView struct {
	Name     string                 `yaml:"name"`
	Title    string                 `yaml:"title"`
	Type     string                 `yaml:"type"`
	Height   int                    `yaml:"height"`
	MinH     int                    `yaml:"min_height"`
	MaxH     int                    `yaml:"max_height"`
	MaxW     int                    `yaml:"max_width"`
	Grid     string                 `yaml:"grid"`
	Elements map[string]YamlElement `yaml:"elements"`
	Order    []string               `yaml:"order"`
	Frame    *Frame                 `yaml:"frame"`
	// OnKey is retained for legacy documents; routing still uses Widget.HandleKey.
	OnKey map[string]string `yaml:"on_key"`
}

// YamlConfig is the top-level container for the Loom YAML configuration.
type YamlConfig struct {
	App   YamlApp    `yaml:"app"`
	Pane  *YamlView  `yaml:"pane"`  // legacy fallback
	View  *YamlView  `yaml:"view"`  // single view
	Views []YamlView `yaml:"views"` // multiple views
}

// RootView resolves the entrypoint layout configuration.
func (c *YamlConfig) RootView() (*YamlView, error) {
	forms := 0
	if c.View != nil {
		forms++
	}
	if c.Pane != nil {
		forms++
	}
	if c.Views != nil {
		forms++
	}
	if forms != 1 {
		return nil, fmt.Errorf("root: declare exactly one of pane, view, views")
	}
	if c.View != nil {
		if c.App.Root != "" && c.App.Root != c.View.Name && !(c.View.Name == "" && c.App.Root == "main") {
			return nil, fmt.Errorf("app.root: unknown view %q", c.App.Root)
		}
		return c.View, nil
	}
	if c.Pane != nil {
		if c.App.Root != "" && c.App.Root != c.Pane.Name && !(c.Pane.Name == "" && c.App.Root == "main") {
			return nil, fmt.Errorf("app.root: unknown pane %q", c.App.Root)
		}
		return c.Pane, nil
	}
	if len(c.Views) > 0 {
		rootName := c.App.Root
		if rootName == "" {
			rootName = "main"
		}
		for i := range c.Views {
			if c.Views[i].Name == rootName {
				return &c.Views[i], nil
			}
		}
		if c.App.Root != "" {
			return nil, fmt.Errorf("app.root: unknown view %q", c.App.Root)
		}
		return &c.Views[0], nil
	}
	return nil, fmt.Errorf("no views defined")
}

// Height returns the target pane height given the estimated content height.
func (c *YamlConfig) Height(contentH int) int {
	if c.App.Height != 0 {
		return c.App.Height
	}
	root, err := c.RootView()
	if err == nil && root.Height != 0 {
		return root.Height
	}

	minH := c.App.MinH
	if root != nil && root.MinH != 0 {
		minH = root.MinH
	}
	maxH := c.App.MaxH
	if root != nil && root.MaxH != 0 {
		maxH = root.MaxH
	}

	if minH <= 0 {
		minH = 3 // default minimum fallback
	}
	if maxH <= 0 {
		maxH = 20 // default maximum fallback
	}
	if minH > maxH {
		minH = maxH
	}

	h := contentH
	if h <= 0 {
		h = minH
	}
	if h < minH {
		h = minH
	}
	if h > maxH {
		h = maxH
	}
	return h
}

// MaxWidth returns the target max width.
func (c *YamlConfig) MaxWidth() int {
	if c.App.MaxW > 0 {
		return c.App.MaxW
	}
	if root, err := c.RootView(); err == nil && root.MaxW > 0 {
		return root.MaxW
	}
	return 0
}

// Validate checks that all views, layout grids, and elements are correctly configured.
func (c *YamlConfig) Validate() error {
	_, err := c.build()
	return err
}

// ValidateYAML decodes and validates a Loom YAML layout configuration stream.
func ValidateYAML(r io.Reader) error {
	_, _, err := BuildWidget(r)
	return err
}

// ParseYAMLFile reads a .loom.yaml file and constructs the widget tree and Pane.
func ParseYAMLFile(path string) (Widget, *Pane, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	return ParseYAML(file)
}

// BuildWidgetFile reads a .loom.yaml file and constructs the widget tree without opening a terminal.
func BuildWidgetFile(path string) (Widget, *YamlConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	return BuildWidget(file)
}

// Router manages screen transitions between multiple views defined in the layout configuration.
type Router struct {
	config  *YamlConfig
	pane    *Pane
	views   map[string]Widget
	current string
	history []string
}

// RouteTo switches the active view to name and adds current to history.
func (r *Router) RouteTo(name string) {
	if _, ok := r.views[name]; ok {
		r.history = append(r.history, r.current)
		r.current = name
	}
}

// Current returns the name of the active view.
func (r *Router) Current() string { return r.current }

// ViewNames returns the names of all routable views, sorted for stable output.
func (r *Router) ViewNames() []string {
	names := make([]string, 0, len(r.views))
	for name := range r.views {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Deeplink navigates through a colon-separated path of view names, pushing each
// onto the history so Esc/back returns through every step. Empty segments are
// skipped. It returns an error naming the offending segment (and listing the
// available views) when a segment does not match a known view.
//
// Example: "list:settings" lands on the settings view with [root, list] in
// history, so two Esc presses walk back to the root.
func (r *Router) Deeplink(path string) error {
	for _, name := range strings.Split(path, ":") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := r.views[name]; !ok {
			return fmt.Errorf("loom: unknown view %q (have: %s)", name, strings.Join(r.ViewNames(), ", "))
		}
		r.RouteTo(name)
	}
	return nil
}

// GoBack pops the last view from history and sets it active. Returns true if successful.
func (r *Router) GoBack() bool {
	if len(r.history) > 0 {
		r.current = r.history[len(r.history)-1]
		r.history = r.history[:len(r.history)-1]
		return true
	}
	return false
}

// Draw renders the active view.
func (r *Router) Draw(c *Canvas, rect Rect) {
	if w, ok := r.views[r.current]; ok {
		w.Draw(c, rect)
	}
}

// currentView returns the active YamlView configuration.
func (r *Router) currentView() *YamlView {
	for i := range r.config.Views {
		if r.config.Views[i].Name == r.current {
			return &r.config.Views[i]
		}
	}
	return nil
}

// ContentHeight estimates the required height for the active view, respecting view/app constraints.
func (r *Router) ContentHeight() int {
	contentH := 0
	if w, ok := r.views[r.current]; ok {
		if ch, ok := w.(ContentHeighter); ok {
			contentH = ch.ContentHeight()
		}
	}

	view := r.currentView()
	if view != nil && view.Height != 0 {
		return view.Height
	}

	minH := r.config.App.MinH
	if view != nil && view.MinH != 0 {
		minH = view.MinH
	}
	maxH := r.config.App.MaxH
	if view != nil && view.MaxH != 0 {
		maxH = view.MaxH
	}

	if minH <= 0 {
		minH = 3
	}
	if maxH <= 0 {
		maxH = 20
	}
	if minH > maxH {
		minH = maxH
	}

	h := contentH
	if h <= 0 {
		h = minH
	}
	if h < minH {
		h = minH
	}
	if h > maxH {
		h = maxH
	}
	return h
}

// HandleKey processes key events, returning quit=true on unhandled ESC.
func (r *Router) HandleKey(e KeyEvent) (quit bool) {
	// If active widget handles key and returns quit, we pop view or ignore
	if w, ok := r.views[r.current]; ok {
		if quit := w.HandleKey(e); quit {
			if r.GoBack() {
				return false
			}
			return true
		}
		return false
	}

	if e.Key == "esc" {
		if r.GoBack() {
			return false
		}
		return true
	}
	return false
}

// HandleMouse forwards mouse events to the active view.
func (r *Router) HandleMouse(e MouseEvent) (quit bool) {
	if w, ok := r.views[r.current]; ok {
		return w.HandleMouse(e)
	}
	return false
}

// BuildWidget parses YAML configuration and constructs the widget tree and Router
// without opening a terminal. Safe to call in tests and headless environments.
func BuildWidget(r io.Reader) (Widget, *YamlConfig, error) {
	var cfg YamlConfig
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, nil, fmt.Errorf("loom: decode yaml: %w", err)
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, nil, fmt.Errorf("loom: decode yaml: %w", err)
		}
		return nil, nil, fmt.Errorf("loom: expected one YAML document")
	}
	w, err := cfg.build()
	if err != nil {
		return nil, nil, fmt.Errorf("loom: %w", err)
	}
	return w, &cfg, nil
}

func (c *YamlConfig) build() (Widget, error) {
	root, err := c.RootView()
	if err != nil {
		return nil, err
	}
	if err := validateDimensions("app", c.App.Height, c.App.MinH, c.App.MaxH, c.App.MaxW); err != nil {
		return nil, err
	}
	if len(c.Views) == 0 {
		if root.Frame != nil && (c.App.Height <= 0 || c.App.MaxW <= 0) {
			return nil, fmt.Errorf("app: frame documents require positive height and max_width")
		}
		return compileView(root, nil)
	}
	router := &Router{config: c, views: make(map[string]Widget), current: root.Name}
	for i := range c.Views {
		v := &c.Views[i]
		if v.Frame != nil && (c.App.Height <= 0 || c.App.MaxW <= 0) {
			return nil, fmt.Errorf("app: frame documents require positive height and max_width")
		}
		if v.Name == "" {
			return nil, fmt.Errorf("views[%d].name: required", i)
		}
		if _, exists := router.views[v.Name]; exists {
			return nil, fmt.Errorf("views[%d].name: duplicate %q", i, v.Name)
		}
		w, err := compileView(v, router)
		if err != nil {
			return nil, fmt.Errorf("views[%d]: %w", i, err)
		}
		router.views[v.Name] = w
	}
	return router, nil
}

func validateDimensions(path string, height, minH, maxH, maxW int) error {
	if height < 0 || minH < 0 || maxH < 0 || maxW < 0 {
		return fmt.Errorf("%s: dimensions cannot be negative", path)
	}
	if minH > 0 && maxH > 0 && minH > maxH {
		return fmt.Errorf("%s: min_height exceeds max_height", path)
	}
	return nil
}

func compileView(v *YamlView, router *Router) (Widget, error) {
	if err := validateDimensions("view", v.Height, v.MinH, v.MaxH, v.MaxW); err != nil {
		return nil, err
	}
	if v.Frame != nil {
		if len(v.Elements) != 0 || v.Grid != "" || len(v.Order) != 0 {
			return nil, fmt.Errorf("view.frame: cannot combine with elements, grid or order")
		}
		if err := v.Frame.validate(); err != nil {
			return nil, err
		}
		return v.Frame, nil
	}
	keys := make([]string, 0, len(v.Elements))
	for key := range v.Elements {
		keys = append(keys, key)
	}
	sort.Strings(keys) // Legacy maps have a stable lexical fallback; order is explicit.
	widgets := make(map[string]Widget, len(keys))
	for _, key := range keys {
		w, err := compileWidget(v.Elements[key])
		if err != nil {
			return nil, fmt.Errorf("elements.%s: %w", key, err)
		}
		if choice, ok := w.(*Choice); ok && router != nil && v.Elements[key].Type == "choice" {
			choice.OnSelect = func(item Item) { router.RouteTo(item.Name) }
		}
		widgets[key] = w
	}
	if v.Grid != "" {
		if len(v.Order) != 0 {
			return nil, fmt.Errorf("view: grid and order are mutually exclusive")
		}
		return ParseASCIIGrid(v.Grid, widgets)
	}
	if len(v.Order) > 0 {
		seen := make(map[string]bool)
		for i, key := range v.Order {
			if widgets[key] == nil {
				return nil, fmt.Errorf("order[%d]: unknown element %q", i, key)
			}
			if seen[key] {
				return nil, fmt.Errorf("order[%d]: duplicate element %q", i, key)
			}
			seen[key] = true
		}
		if len(seen) != len(keys) {
			return nil, fmt.Errorf("order: must include every element")
		}
		keys = v.Order
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("view: no elements or frame declared")
	}
	children := make([]Widget, 0, len(keys))
	for _, key := range keys {
		children = append(children, widgets[key])
	}
	if len(children) == 1 {
		return children[0], nil
	}
	return NewStack(Vertical, children...), nil
}

// ParseYAML reads YAML configuration from a reader and builds the TUI layout with view routing.
func ParseYAML(r io.Reader) (Widget, *Pane, error) {
	widget, cfg, err := BuildWidget(r)
	if err != nil {
		return nil, nil, err
	}

	// Estimate content height for initial pane sizing
	contentH := 0
	if ch, ok := widget.(ContentHeighter); ok {
		contentH = ch.ContentHeight()
	}

	// Initialize Pane with dynamic height
	paneH := cfg.Height(contentH)
	if Debug {
		root, _ := cfg.RootView()
		minH := cfg.App.MinH
		if root != nil && root.MinH != 0 {
			minH = root.MinH
		}
		maxH := cfg.App.MaxH
		if root != nil && root.MaxH != 0 {
			maxH = root.MaxH
		}
		fmt.Fprintf(os.Stderr, "loom: [debug] layout: %s (max_w: %d)\n", cfg.App.Name, cfg.MaxWidth())
		fmt.Fprintf(os.Stderr, "loom: [debug] height: min: %d, max: %d\n", minH, maxH)
		fmt.Fprintf(os.Stderr, "loom: [debug] height: content: %d -> pane: %d\n", contentH, paneH)
	}
	pane, err := New(paneH)
	if err != nil {
		return nil, nil, err
	}
	pane.MaxCols = cfg.MaxWidth()
	pane.Resizeable = true

	if router, ok := widget.(*Router); ok {
		router.pane = pane
	}
	return widget, pane, nil
}

func configureChoice(choice *Choice, elem YamlElement) {
	if elem.Prompt != "" {
		choice.Prompt = elem.Prompt
	}
	choice.CursorAlign = elem.Cursor
	choice.Placeholder = elem.Placeholder
	choice.Controls = elem.Controls
	choice.MaxWidth = elem.MaxW
}

func compileWidget(elem YamlElement) (Widget, error) {
	switch elem.Type {
	case "input":
		// Choice widget acts as the search/input filter bar when no list items are present.
		choice := NewChoice(nil)
		configureChoice(choice, elem)
		return choice, nil

	case "choice":
		var items []Item
		if elem.Static.Kind != 0 {
			var str string
			if err := elem.Static.Decode(&str); err == nil {
				parts := strings.Split(str, "\n")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						items = append(items, Item{Name: p})
					}
				}
			} else {
				var slice []string
				if err := elem.Static.Decode(&slice); err == nil {
					for _, s := range slice {
						items = append(items, Item{Name: s})
					}
				}
			}
		} else if strings.HasPrefix(elem.Source, "static:") {
			parts := strings.Split(strings.TrimPrefix(elem.Source, "static:"), ",")
			for _, p := range parts {
				items = append(items, Item{Name: p})
			}
		}
		choice := NewChoice(items)
		configureChoice(choice, elem)
		return choice, nil

	case "view":
		var lines []string
		if elem.Static.Kind != 0 {
			var str string
			if err := elem.Static.Decode(&str); err == nil {
				lines = strings.Split(str, "\n")
			} else {
				var slice []string
				if err := elem.Static.Decode(&slice); err == nil {
					lines = slice
				}
			}
		} else if strings.HasPrefix(elem.Source, "static:") {
			lines = strings.Split(strings.TrimPrefix(elem.Source, "static:"), ",")
		}
		return NewView(lines), nil

	case "settings":
		var settings []Setting
		if elem.Static.Kind == yaml.SequenceNode {
			var ySettings []YamlSetting
			if err := elem.Static.Decode(&ySettings); err == nil {
				for _, ys := range ySettings {
					setting := Setting{
						Label: ys.Name,
					}
					// Options or default pointers require local variables
					// allocated per loop iteration.
					nameVal := ys.Name
					defaultVal := ys.Default
					optionsVal := ys.Options

					switch ys.Kind {
					case "bool":
						setting.Kind = KindBool
						boolVal := defaultVal == "true"
						setting.Bool = &boolVal
					case "string":
						setting.Kind = KindString
						setting.Str = &defaultVal
					case "choice":
						setting.Kind = KindChoice
						setting.Options = optionsVal
						idx := 0
						for i, opt := range optionsVal {
							if opt == defaultVal {
								idx = i
								break
							}
						}
						setting.Index = &idx
					}
					_ = nameVal // prevent unused check
					settings = append(settings, setting)
				}
			}
		}
		return NewSettings(settings), nil

	case "notif":
		return &Notif{}, nil

	default:
		return nil, fmt.Errorf("unsupported element type: %s", elem.Type)
	}
}

// ParseASCIIGrid parses the ASCII visual layout grid and maps widget keys to layout stacks.
func ParseASCIIGrid(gridStr string, widgets map[string]Widget) (Widget, error) {
	lines := strings.Split(gridStr, "\n")
	// Clean empty lines at start/end
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty grid layout")
	}

	// 1. Identify horizontal boundary rows
	var yCoords []int
	for y, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// A line is a horizontal border boundary if it contains only +, -, |, and spaces, and has at least one -
		isBorder := true
		hasDash := false
		for _, r := range trimmed {
			if r == '-' {
				hasDash = true
			} else if r != '+' && r != '|' && r != ' ' {
				isBorder = false
				break
			}
		}
		if isBorder && hasDash {
			yCoords = append(yCoords, y)
		}
	}

	// 2. Identify vertical boundary columns
	xCoordsMap := make(map[int]bool)
	for _, line := range lines {
		for x, r := range line {
			if r == '+' || r == '|' {
				xCoordsMap[x] = true
			}
		}
	}
	var xCoords []int
	for x := range xCoordsMap {
		xCoords = append(xCoords, x)
	}
	sort.Ints(xCoords)

	if len(yCoords) < 2 || len(xCoords) < 2 {
		return nil, fmt.Errorf("invalid grid boundaries: horizontal boundaries=%d, vertical boundaries=%d", len(yCoords), len(xCoords))
	}

	var rowWidgets []Widget

	// 3. Process each row segment
	for r := 0; r < len(yCoords)-1; r++ {
		yStart := yCoords[r]
		yEnd := yCoords[r+1]

		// Find cell columns in this row segment
		var colWidgets []Widget
		cStart := 0

		for cNext := 1; cNext < len(xCoords); cNext++ {
			xBorder := xCoords[cNext]
			// Check if xBorder has a vertical line in all content rows of this segment
			isBorder := true
			for y := yStart + 1; y < yEnd; y++ {
				if y >= len(lines) || xBorder >= len(lines[y]) {
					isBorder = false
					break
				}
				char := lines[y][xBorder]
				if char != '|' && char != '+' {
					isBorder = false
					break
				}
			}

			// If it's a border or the last column, we have a cell
			if isBorder || cNext == len(xCoords)-1 {
				xStart := xCoords[cStart]
				xEnd := xBorder

				// Find the character inside this cell
				key := ""
				for y := yStart + 1; y < yEnd; y++ {
					if y >= len(lines) {
						continue
					}
					for x := xStart + 1; x < xEnd; x++ {
						if x >= len(lines[y]) {
							continue
						}
						char := lines[y][x]
						if char != ' ' && char != '|' && char != '+' && char != '-' {
							key = string(char)
							break
						}
					}
					if key != "" {
						break
					}
				}

				if w, ok := widgets[key]; ok {
					colWidgets = append(colWidgets, w)
				} else if key != "" {
					return nil, fmt.Errorf("grid: unknown element %q at row %d", key, yStart+1)
				}
				cStart = cNext
			}
		}

		if len(colWidgets) == 1 {
			rowWidgets = append(rowWidgets, colWidgets[0])
		} else if len(colWidgets) > 1 {
			rowWidgets = append(rowWidgets, NewStack(Horizontal, colWidgets...))
		}
	}

	if len(rowWidgets) == 0 {
		return nil, fmt.Errorf("no widgets mapped in layout grid")
	}

	if len(rowWidgets) == 1 {
		return rowWidgets[0], nil
	}

	return NewStack(Vertical, rowWidgets...), nil
}
