// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// Widget is the core interface every loom UI element must satisfy.
//
// Pane drives a single root Widget. Composite widgets (Stack, Grid, Popup)
// implement Widget and delegate to their children.
//
// Lifecycle per frame:
//  1. Pane clears the Canvas and calls root.Draw(canvas, canvas.Bounds()).
//  2. Pane reads one event (key or mouse) from the terminal.
//  3. Pane calls root.HandleKey or root.HandleMouse.
//  4. If the handler returns quit=true, Pane.Run returns.
type Widget interface {
	// Draw renders the widget into region r of canvas c.
	// r is guaranteed to be within c.Bounds(). Draw must not write outside r.
	Draw(c *Canvas, r Rect)

	// HandleKey processes a keyboard event.
	// Returns quit=true to signal that the event loop should stop.
	HandleKey(e KeyEvent) (quit bool)

	// HandleMouse processes a mouse event.
	// Returns quit=true to signal that the event loop should stop.
	HandleMouse(e MouseEvent) (quit bool)
}

// Focusable is an optional interface for widgets that track input focus.
// Pane uses it to highlight the active widget when multiple widgets are present.
type Focusable interface {
	Widget
	Focused() bool
	SetFocus(bool)
}

// Item is a named entry used by Choice and similar list widgets.
type Item struct {
	Name string
	Desc string
}

// ContentHeighter is an optional interface for widgets that can estimate their content height.
type ContentHeighter interface {
	ContentHeight() int
}
