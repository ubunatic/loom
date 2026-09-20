// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "codeberg.org/ubunatic/loom/measure"

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

	// HandleMouse processes a mouse event. e.X and e.Y are canvas-absolute,
	// 0-based coordinates, with the canvas origin at its top-left corner.
	// Returns quit=true to signal that the event loop should stop.
	HandleMouse(e MouseEvent) (quit bool)
}

// PaneRequest declares terminal capabilities a widget needs from its host
// before the pane starts processing input.
// Zero values request no special capability; Mouse is a terminal mouse mode
// (1000 for clicks or 1003 for all tracking), and MaxCols 0 means unbounded.
type PaneRequest struct {
	Mouse      int
	Resizeable bool
	MaxCols    int
	OwnsQuit   bool
}

// PaneRequester optionally lets a widget declare terminal capabilities to its
// host pane. Pane reads the request once before starting its event loop.
type PaneRequester interface {
	PaneRequest() PaneRequest
}

func mergePaneRequest(dst *PaneRequest, src PaneRequest) {
	if src.Mouse > dst.Mouse {
		dst.Mouse = src.Mouse
	}
	dst.Resizeable = dst.Resizeable || src.Resizeable
	if dst.MaxCols == 0 || src.MaxCols == 0 {
		dst.MaxCols = 0
	} else if src.MaxCols > dst.MaxCols {
		dst.MaxCols = src.MaxCols
	}
	dst.OwnsQuit = dst.OwnsQuit || src.OwnsQuit
}

// Focusable is an optional interface for widgets that track input focus.
// Pane uses it to highlight the active widget when multiple widgets are present.
type Focusable interface {
	Widget
	Focused() bool
	SetFocus(bool)
}

// KeyConsumer optionally lets a widget report that it handled a key. The
// signal allows hosted composites to give children first refusal while still
// preserving Widget.HandleKey's historical quit-only API.
type KeyConsumer interface {
	ConsumeKey(e KeyEvent) (quit, consumed bool)
}

// FocusContainer is a focusable composite that can move focus within itself.
// FocusNext and FocusPrevious return false when focus is already at the
// corresponding boundary, allowing a parent container to continue traversal.
type FocusContainer interface {
	Focusable
	FocusNext() bool
	FocusPrevious() bool
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

// WidthHeighter estimates preferred height at an allocated terminal-cell width.
// Resizable panes prefer this over ContentHeighter for responsive roots.
type WidthHeighter interface {
	HeightForWidth(width int) int
}

// ContentWidther estimates preferred width from visible widget content.
type ContentWidther interface {
	ContentWidth() int
}

// Measurer reports a preferred size at an optional parent-supplied width.
type Measurer interface {
	Measure(width int) measure.Size
}
