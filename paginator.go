// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "fmt"

// PaginatorStyle selects the page indicator presentation.
type PaginatorStyle int

const (
	PaginatorDots PaginatorStyle = iota
	PaginatorNumeric
)

// Paginator provides page navigation and reports page changes to its host.
// Page is zero-based; Pages is clamped to at least one.
type Paginator struct {
	Page     int
	Pages    int
	Style    PaginatorStyle
	OnChange func(page int)

	drawn bool
	width int
}

// NewPaginator creates a paginator with the requested number of pages.
func NewPaginator(pages int) *Paginator {
	p := &Paginator{Style: PaginatorDots}
	p.SetPages(pages)
	return p
}

// SetPages updates the page count and clamps the current page.
func (p *Paginator) SetPages(pages int) {
	if pages < 1 {
		pages = 1
	}
	p.Pages = pages
	if p.Page < 0 {
		p.SetPage(0)
	} else if p.Page >= pages {
		p.SetPage(pages - 1)
	}
}

// SetPage selects a zero-based page, clamping it to the available range.
func (p *Paginator) SetPage(page int) {
	if p.Pages < 1 {
		p.Pages = 1
	}
	if page < 0 {
		page = 0
	}
	if page >= p.Pages {
		page = p.Pages - 1
	}
	if p.Page == page {
		return
	}
	p.Page = page
	if p.OnChange != nil {
		p.OnChange(page)
	}
}

// ConsumeKey handles page up and page down without requesting application quit.
func (p *Paginator) ConsumeKey(e KeyEvent) EventResult {
	switch e.Key {
	case "pgup", "pageup":
		p.SetPage(p.Page - 1)
		return Handled()
	case "pgdown", "pgdn", "pagedown":
		p.SetPage(p.Page + 1)
		return Handled()
	default:
		return Ignored()
	}
}

// ConsumeMouse selects a page when a dot is clicked. Coordinates are 0-based
// and local to the widget rectangle.
func (p *Paginator) ConsumeMouse(e MouseEvent) EventResult {
	if !p.drawn || e.Action != MousePress || e.Button != MouseLeft || e.Y != 0 || e.X < 0 || e.X >= p.width {
		return Ignored()
	}
	if p.Style == PaginatorDots {
		index := e.X / 2
		if e.X%2 == 1 || index >= p.Pages {
			return Ignored()
		}
		p.SetPage(index)
		return Handled()
	}
	if p.Style == PaginatorNumeric {
		if e.X < p.width/2 {
			p.SetPage(p.Page - 1)
		} else {
			p.SetPage(p.Page + 1)
		}
		return Handled()
	}
	return Ignored()
}

// Draw renders the page indicators from the top-left of r.
func (p *Paginator) Draw(c *Canvas, r Rect) {
	p.drawn = true
	p.width = r.W
	if r.W <= 0 || r.H <= 0 {
		return
	}
	text := p.indicator()
	c.Write(r.X, r.Y, text, Style{})
}

func (p *Paginator) indicator() string {
	if p.Pages < 1 {
		p.Pages = 1
	}
	if p.Page < 0 {
		p.Page = 0
	}
	if p.Page >= p.Pages {
		p.Page = p.Pages - 1
	}
	if p.Style == PaginatorNumeric {
		return fmt.Sprintf("%d/%d", p.Page+1, p.Pages)
	}
	text := ""
	for i := 0; i < p.Pages; i++ {
		if i > 0 {
			text += " "
		}
		if i == p.Page {
			text += "●"
		} else {
			text += "○"
		}
	}
	return text
}
