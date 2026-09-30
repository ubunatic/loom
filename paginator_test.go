// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func paginatorText(c *Canvas, n int) string {
	text := ""
	for x := 0; x < n && x < c.cols; x++ {
		if !c.cells[0][x].Continuation {
			text += c.cells[0][x].Text
		}
	}
	return text
}

func TestPaginatorNavigationAndCallback(t *testing.T) {
	p := NewPaginator(7)
	var changes []int
	p.OnChange = func(page int) { changes = append(changes, page) }
	if p.Page != 0 || p.Pages != 7 {
		t.Fatalf("initial state = %d/%d, want 0/7", p.Page, p.Pages)
	}
	p.HandleKey(KeyEvent{Key: "pgdn"})
	p.HandleKey(KeyEvent{Key: "pgup"})
	p.HandleKey(KeyEvent{Key: "pgup"})
	if p.Page != 0 {
		t.Fatalf("page = %d, want lower bound 0", p.Page)
	}
	p.SetPage(100)
	if p.Page != 6 {
		t.Fatalf("page = %d, want upper bound 6", p.Page)
	}
	if got := changes; len(got) != 3 || got[0] != 1 || got[1] != 0 || got[2] != 6 {
		t.Fatalf("OnChange pages = %v, want [1 0 6]", got)
	}
}

func TestPaginatorAcceptsPageDownAliases(t *testing.T) {
	for _, key := range []string{"pgdown", "pgdn", "pagedown"} {
		p := NewPaginator(3)
		result := p.ConsumeKey(KeyEvent{Key: key})
		if !result.Consumed || p.Page != 1 {
			t.Errorf("key %q: consumed=%v page=%d, want consumed and page 1", key, result.Consumed, p.Page)
		}
	}
	decoded := DecodeKey([]byte("\x1b[6~"))
	p := NewPaginator(3)
	if decoded.Key != "pgdown" || !p.ConsumeKey(decoded).Consumed || p.Page != 1 {
		t.Errorf("decoded page-down key %q did not advance paginator to page 1", decoded.Key)
	}
}

func TestPaginatorDotAndNumericRendering(t *testing.T) {
	p := NewPaginator(3)
	c := NewCanvas(20, 1)
	p.Draw(c, Rect{W: 20, H: 1})
	if got := paginatorText(c, 5); got != "● ○ ○" {
		t.Fatalf("dot rendering = %q, want leading selected dot", got)
	}
	p.Style = PaginatorNumeric
	p.SetPage(1)
	c = NewCanvas(20, 1)
	p.Draw(c, Rect{W: 20, H: 1})
	if got := paginatorText(c, 3); got != "2/3" {
		t.Fatalf("numeric rendering = %q, want 2/3", got)
	}
}

func TestPaginatorMouseSelectsPageWithLocalCoordinates(t *testing.T) {
	p := NewPaginator(4)
	c := NewCanvas(12, 1)
	p.Draw(c, Rect{W: 8, H: 1})
	e := MouseEvent{Action: MousePress, Button: MouseLeft, X: 4, Y: 0}
	if result := p.ConsumeMouse(e); !result.Consumed {
		t.Fatal("click on a page indicator was not consumed")
	}
	if p.Page != 2 {
		t.Fatalf("clicked page = %d, want 2", p.Page)
	}
}
