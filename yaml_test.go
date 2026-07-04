// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestBuildWidget(t *testing.T) {
	yamlStr := `
pane:
  height: 10
  max_width: 60
  grid: |
    +---+
    |P  |
    +-+-+
    |C|I|
    +-+-+
    |S  |
    +-+-+
  elements:
    P:
      type: input
      prompt: "🔍 Emoji Search: "
      placeholder: "Type to filter..."
    C:
      type: choice
      source: "static:git,make,find"
    I:
      type: view
      source: "static:info line 1,info line 2"
    S:
      type: notif
`

	widget, cfg, err := loom.BuildWidget(strings.NewReader(yamlStr))
	if err != nil {
		t.Fatalf("BuildWidget failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.MaxWidth() != 60 {
		t.Errorf("cfg.MaxWidth() = %d, want 60", cfg.MaxWidth())
	}

	if widget == nil {
		t.Fatal("expected non-nil widget root")
	}

	// The root widget should be a vertical stack because we have 3 rows in the grid.
	stack, ok := widget.(*loom.Stack)
	if !ok {
		t.Fatalf("expected root widget to be *loom.Stack, got %T", widget)
	}

	if len(stack.Children) != 3 {
		t.Fatalf("expected 3 vertical rows, got %d", len(stack.Children))
	}
}

func TestParseASCIIGrid(t *testing.T) {
	gridStr := `
+---+
|P  |
+-+-+
|C|I|
+-+-+
`
	widgets := map[string]loom.Widget{
		"P": loom.NewChoice(nil),
		"C": loom.NewChoice(nil),
		"I": loom.NewView(nil),
	}

	widget, err := loom.ParseASCIIGrid(gridStr, widgets)
	if err != nil {
		t.Fatalf("ParseASCIIGrid failed: %v", err)
	}

	stack, ok := widget.(*loom.Stack)
	if !ok {
		t.Fatalf("expected root stack, got %T", widget)
	}

	if len(stack.Children) != 2 {
		t.Fatalf("expected 2 vertical rows, got %d", len(stack.Children))
	}

	// First row should contain only P (not wrapped in a horizontal stack since it's alone)
	_, isChoice := stack.Children[0].(*loom.Choice)
	if !isChoice {
		t.Errorf("expected row 0 child to be *loom.Choice, got %T", stack.Children[0])
	}

	// Second row should be a Horizontal stack containing C and I
	row2Stack, ok := stack.Children[1].(*loom.Stack)
	if !ok {
		t.Fatalf("expected row 1 child to be *loom.Stack, got %T", stack.Children[1])
	}

	if len(row2Stack.Children) != 2 {
		t.Errorf("expected 2 children in row 1, got %d", len(row2Stack.Children))
	}
}

func TestValidateYAML(t *testing.T) {
	validYaml := `
pane:
  height: 5
  grid: |
    +---+
    |P  |
    +---+
  elements:
    P:
      type: notif
`
	if err := loom.ValidateYAML(strings.NewReader(validYaml)); err != nil {
		t.Errorf("expected valid YAML to pass validation, got error: %v", err)
	}

	invalidYaml := `
pane:
  height: -5
  grid: |
    +---+
    |P  |
    +---+
  elements:
    P:
      type: notif
`
	if err := loom.ValidateYAML(strings.NewReader(invalidYaml)); err == nil {
		t.Error("expected invalid YAML (height <= 0) to fail validation")
	}
}

func TestStaticContent(t *testing.T) {
	yamlStr := `
pane:
  height: 5
  grid: |
    +---+
    |A  |
    +-+-+
    |B  |
    +---+
  elements:
    A:
      type: view
      static: |
        line 1
        line 2
    B:
      type: choice
      static:
        - option 1
        - option 2
`
	widget, _, err := loom.BuildWidget(strings.NewReader(yamlStr))
	if err != nil {
		t.Fatalf("failed parsing static lists/blocks: %v", err)
	}

	stack, ok := widget.(*loom.Stack)
	if !ok {
		t.Fatalf("expected stack, got %T", widget)
	}
	if len(stack.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(stack.Children))
	}
}

func TestCursorPositioningAndFocusPropagation(t *testing.T) {
	yamlStr := `
app:
  root: main
  height: 10
views:
  - name: main
    grid: |
      +---+
      |A|B|
      +---+
    elements:
      A:
        type: choice
        prompt: "🔍 "
        cursor: start
      B:
        type: choice
        prompt: "filter> "
        cursor: end
`
	widget, _, err := loom.BuildWidget(strings.NewReader(yamlStr))
	if err != nil {
		t.Fatalf("failed parsing multi-view: %v", err)
	}

	router, ok := widget.(*loom.Router)
	if !ok {
		t.Fatalf("expected *loom.Router, got %T", widget)
	}

	// Active view is "main" which wraps the grid
	canvas := loom.NewCanvas(50, 10)
	router.Draw(canvas, canvas.Bounds())

	// Grid should focus index 0 (which is element A).
	// A is a choice widget, and it should get focused.
	// Since A has cursor: start, and prompt "🔍 " (visual width 3),
	// canvas.CursorX should be 3.
	if canvas.CursorX != 3 {
		t.Errorf("expected CursorX = 3, got %d", canvas.CursorX)
	}

	// Now if we change focus to index 1 (element B) using key events or directly.
	// B is the second child in the horizontal stack (or grid).
	// Let's verify that B is focused when we press "tab".
	router.HandleKey(loom.KeyEvent{Key: "tab"})

	// Redraw to update focus propagation and canvas cursor
	canvas.Clear()
	router.Draw(canvas, canvas.Bounds())

	// B should be focused now. B has cursor: end.
	// In the grid layout, B gets the second half of the width: [25, 50).
	// Its drawW is cellW = r.W / Cols = 50 / 2 = 25.
	// For B, r.X is 25.
	// cursorX is r.X + drawW - 1 = 25 + 25 - 1 = 49.
	if canvas.CursorX != 49 {
		t.Errorf("expected CursorX = 49 for B, got %d", canvas.CursorX)
	}
}

func TestRouterDynamicResize(t *testing.T) {
	yamlStr := `
app:
  name: testapp
  root: main
  min_height: 5
  max_height: 12
views:
  - name: main
    grid: |
      +---+
      |M  |
      +---+
    elements:
      M:
        type: choice
        static:
          - "one"
          - "two"
  - name: help
    grid: |
      +---+
      |H  |
      +---+
    elements:
      H:
        type: view
        static: |-
          1
          2
          3
          4
          5
          6
          7
          8
          9
          10
`
	widget, _, err := loom.BuildWidget(strings.NewReader(yamlStr))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	router, ok := widget.(*loom.Router)
	if !ok {
		t.Fatalf("expected *loom.Router, got %T", widget)
	}

	// 1. Initial view "main" content height is 3 (choice with 2 items), min_height is 5
	if h := router.ContentHeight(); h != 5 {
		t.Errorf("expected main view target height = 5 (min_height), got %d", h)
	}

	// 2. Transition to "help" view. Content height is 10 (10 lines), within min/max [5, 12]
	router.RouteTo("help")
	if h := router.ContentHeight(); h != 10 {
		t.Errorf("expected help view target height = 10, got %d", h)
	}
}

func TestRouterNavigation(t *testing.T) {
	yamlStr := `
app:
  root: main
  height: 8
views:
  - name: main
    grid: |
      +---+
      |L  |
      +---+
    elements:
      L:
        type: choice
        source: "static:list,grid,help"
    on_key:
      "esc": exit
  - name: list
    grid: |
      +---+
      |V  |
      +---+
    elements:
      V:
        type: view
        static: "list view"
    on_key:
      "esc": back
  - name: grid
    grid: |
      +---+
      |V  |
      +---+
    elements:
      V:
        type: view
        static: "grid view"
    on_key:
      "esc": back
  - name: help
    grid: |
      +---+
      |V  |
      +---+
    elements:
      V:
        type: view
        static: "help view"
    on_key:
      "esc": back
`
	widget, _, err := loom.BuildWidget(strings.NewReader(yamlStr))
	if err != nil {
		t.Fatalf("BuildWidget failed: %v", err)
	}

	router, ok := widget.(*loom.Router)
	if !ok {
		t.Fatalf("expected *loom.Router, got %T", widget)
	}

	// Start on main
	if router.Current() != "main" {
		t.Fatalf("expected current view = main, got %s", router.Current())
	}

	// Press Enter to select the first item ("list") — triggers OnSelect → RouteTo
	router.HandleKey(loom.KeyEvent{Key: "enter"})
	if router.Current() != "list" {
		t.Errorf("expected current view = list after Enter, got %s", router.Current())
	}

	// Navigate down to "grid" and press Enter
	router.GoBack()
	router.HandleKey(loom.KeyEvent{Key: "down"})
	router.HandleKey(loom.KeyEvent{Key: "enter"})
	if router.Current() != "grid" {
		t.Errorf("expected current view = grid after navigating down + Enter, got %s", router.Current())
	}

	// GoBack should return to main
	router.GoBack()
	if router.Current() != "main" {
		t.Errorf("expected current view = main after GoBack, got %s", router.Current())
	}
}
