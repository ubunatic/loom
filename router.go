// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"sort"
	"strings"
)

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

// ConsumeKey processes key events, returning quit=true on unhandled ESC.
func (r *Router) ConsumeKey(e KeyEvent) (quit EventResult) {
	// If active widget handles key and returns quit, we pop view or ignore
	if w, ok := r.views[r.current]; ok {
		if result := w.ConsumeKey(e); result.Consumed {
			if result.Quit {
				if r.GoBack() {
					return Handled()
				}
				return QuitResult()
			}
			return result
		}
		return Ignored()
	}

	if e.Key == "esc" {
		if r.GoBack() {
			return Ignored()
		}
		return QuitResult()
	}
	return Ignored()
}

// ConsumeMouse forwards mouse events to the active view.
func (r *Router) ConsumeMouse(e MouseEvent) (quit EventResult) {
	if w, ok := r.views[r.current]; ok {
		return w.ConsumeMouse(e)
	}
	return Ignored()
}
