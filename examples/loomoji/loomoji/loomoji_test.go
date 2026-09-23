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

	// Digits start a search while the grid is focused.
	p.HandleKey(loom.KeyEvent{Text: "3"})
	if p.group != grpFaces || p.query.Value() != "3" || p.gridFocus {
		t.Errorf("digit should start a search without changing category; group=%d query=%q focus=%v", p.group, p.query.Value(), p.gridFocus)
	}
	p.query.SetValue("")
	p.gridFocus = true
	p.refresh()

	// Brackets cycle categories.
	p.HandleKey(loom.KeyEvent{Text: "]"})
	if p.group != grpHands {
		t.Errorf("expected group %d (Hands), got %d", grpHands, p.group)
	}

	// Press ']' to advance to Animals
	p.HandleKey(loom.KeyEvent{Text: "]"})
	if p.group != grpAnimals {
		t.Errorf("expected group %d (Animals), got %d", grpAnimals, p.group)
	}

	// Press '[' to go back to Hands
	p.HandleKey(loom.KeyEvent{Text: "["})
	if p.group != grpHands {
		t.Errorf("expected group %d (Hands), got %d", grpHands, p.group)
	}

	// Select Box category directly
	p.selectCategory(grpBox)
	if p.group != grpBox {
		t.Errorf("expected group %d (Box), got %d", grpBox, p.group)
	}
	hasLightBox := false
	for _, idx := range p.items {
		if p.entries[idx].icon == "┌" {
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
		e := p.entries[idx]
		if !containsIgnoreCase(e.name, "box") && !containsIgnoreCase(e.icon, "box") {
			t.Errorf("item %q (%s) does not match query 'box'", e.name, e.icon)
		}
	}
}

func TestMouseInteraction(t *testing.T) {
	p := newPicker()
	canvas := loom.NewCanvas(60, 15)
	p.Draw(canvas, loom.Rect{X: 0, Y: 0, W: 50, H: 12})

	// Click category bar (y = p.categoryY)
	catY := p.categoryY
	// Click on 3rd category (Animals)
	// Calculate X position of 3rd category
	cx := 2
	for i := 0; i < 2; i++ {
		cx += loom.StringWidth(p.categories[i].icon) + 2
	}
	if p.HandleMouse(loom.MouseEvent{
		Action: loom.MousePress,
		Button: loom.MouseLeft,
		X:      cx,
		Y:      catY + 1,
	}) {
		t.Fatal("category click unexpectedly selected an entry")
	}
	if p.group != grpAnimals {
		t.Errorf("expected Animals after mouse click, got %d", p.group)
	}

	canvas = loom.NewCanvas(60, 15)
	p.Draw(canvas, loom.Rect{X: 0, Y: 0, W: 50, H: 12})
	want := p.entries[p.items[1]].icon
	if !p.HandleMouse(loom.MouseEvent{
		Action: loom.MousePress,
		Button: loom.MouseLeft,
		X:      2 + p.cellWidth + 1,
		Y:      p.searchY + 3,
	}) {
		t.Fatal("grid click should select and exit")
	}
	if p.chosen != want {
		t.Errorf("mouse selected %q, want %q", p.chosen, want)
	}
}

func TestCategoriesAreDisjoint(t *testing.T) {
	p := newPicker()
	seen := make(map[string]int)
	for _, e := range p.entries {
		if e.group == grpLines || e.group == grpBox {
			seen[e.icon]++
		}
	}
	for icon, count := range seen {
		if count != 1 {
			t.Errorf("line and box symbol %q occurs %d times", icon, count)
		}
	}
}

func TestGridCellsFitWideIcons(t *testing.T) {
	p := newPicker()
	p.entries = []entry{{icon: "👨‍👩‍👧‍👦", name: "family", group: grpFaces}, {icon: "😀", name: "face", group: grpFaces}}
	p.items = []int{0, 1}
	canvas := loom.NewCanvas(14, 1)
	p.drawGrid(canvas, loom.Rect{X: 0, Y: 0, W: 14, H: 1})
	if p.cellWidth != loom.StringWidth(p.entries[0].icon) {
		t.Fatalf("cell width = %d, want %d", p.cellWidth, loom.StringWidth(p.entries[0].icon))
	}
	if p.cols != 1 {
		t.Fatalf("columns = %d, want 1 when two two-cell icons do not fit side by side", p.cols)
	}
	if got := loom.StringWidth(canvas.Row(0)); got > 14 {
		t.Fatalf("rendered row width %d exceeds canvas width", got)
	}
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
