// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomoji

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestInitialGridFocusAndNavigation(t *testing.T) {
	p := newPicker()
	if !p.gridFocus {
		t.Fatalf("expected picker to start with grid focus, got false")
	}
	if len(p.items) == 0 {
		t.Fatalf("expected initial items to be populated")
	}
	if p.index != 0 {
		t.Errorf("expected initial index 0, got %d", p.index)
	}

	// Arrow right moves grid selection
	p.HandleKey(loom.KeyEvent{Key: "right"})
	if p.index != 1 {
		t.Errorf("expected index 1 after right arrow, got %d", p.index)
	}

	// Arrow left moves back
	p.HandleKey(loom.KeyEvent{Key: "left"})
	if p.index != 0 {
		t.Errorf("expected index 0 after left arrow, got %d", p.index)
	}
}

func TestCategorySwitching(t *testing.T) {
	p := newPicker()

	// Initial category is Faces (0)
	if p.group != grpFaces {
		t.Fatalf("expected initial group %d, got %d", grpFaces, p.group)
	}

	// Press '3' to jump to Animals (index 2)
	p.HandleKey(loom.KeyEvent{Text: "3"})
	if p.group != grpAnimals {
		t.Errorf("expected group %d (Animals), got %d", grpAnimals, p.group)
	}
	if len(p.items) == 0 {
		t.Errorf("expected non-empty items for Animals category")
	}

	// Press ']' to advance to Food
	p.HandleKey(loom.KeyEvent{Text: "]"})
	if p.group != grpFood {
		t.Errorf("expected group %d (Food), got %d", grpFood, p.group)
	}

	// Press '[' to go back to Animals
	p.HandleKey(loom.KeyEvent{Text: "["})
	if p.group != grpAnimals {
		t.Errorf("expected group %d (Animals), got %d", grpAnimals, p.group)
	}

	// Select Box category directly
	p.selectCategory(grpBox)
	if p.group != grpBox {
		t.Errorf("expected group %d (Box), got %d", grpBox, p.group)
	}
	hasLightBox := false
	for _, idx := range p.items {
		if entryList[idx].icon == "┌" {
			hasLightBox = true
			break
		}
	}
	if !hasLightBox {
		t.Errorf("expected Box category to contain '┌'")
	}
}

func TestSearchFiltering(t *testing.T) {
	p := newPicker()

	// Type text into search
	p.HandleKey(loom.KeyEvent{Text: "box"})
	if p.gridFocus {
		t.Errorf("expected search to gain focus after typing text")
	}
	if len(p.items) == 0 {
		t.Fatalf("expected search for 'box' to return results")
	}

	// Verify all returned items match "box" in name or icon
	for _, idx := range p.items {
		e := entryList[idx]
		if !containsIgnoreCase(e.name, "box") && !containsIgnoreCase(e.icon, "box") {
			t.Errorf("item %q (%s) does not match query 'box'", e.name, e.icon)
		}
	}
}

func TestMouseInteraction(t *testing.T) {
	p := newPicker()
	canvas := loom.NewCanvas(60, 15)
	p.Draw(canvas, canvas.Bounds())

	// Click category bar (y = p.categoryY)
	catY := p.categoryY
	// Click on 3rd category (Animals)
	// Calculate X position of 3rd category
	cx := 2
	for i := 0; i < 2; i++ {
		cx += loom.StringWidth(categories[i].icon) + 2
	}
	p.HandleMouse(loom.MouseEvent{
		Action: loom.MousePress,
		Button: loom.MouseLeft,
		X:      cx,
		Y:      catY,
	})
	if p.group != 2 {
		t.Errorf("expected group 2 after mouse click, got %d", p.group)
	}

	// Click on first grid item
	p.handleGridMouse(loom.MouseEvent{
		Action: loom.MousePress,
		Button: loom.MouseLeft,
		X:      2,
		Y:      0,
	})
	if p.chosen == "" {
		t.Errorf("expected item to be chosen on mouse click in grid")
	}
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
