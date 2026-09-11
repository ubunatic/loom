// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"codeberg.org/ubunatic/loom/layout"
	"codeberg.org/ubunatic/loom/measure"
)

// AlignBox is a container widget that positions its child according to horizontal
// and vertical alignment rules within the allocated Canvas rectangle.
type AlignBox struct {
	Child  Widget
	HAlign layout.Align
	VAlign layout.Align
	Width  int // explicit content width; 0 derives from child
	Height int // explicit content height; 0 derives from child
}

// Center is a convenience type for an AlignBox configured to center its child.
type Center = AlignBox

// NewCenter creates an AlignBox widget that centers its child both horizontally
// and vertically.
func NewCenter(child Widget) *AlignBox {
	return &AlignBox{
		Child:  child,
		HAlign: layout.AlignCenter,
		VAlign: layout.AlignCenter,
	}
}

// NewAlign creates an AlignBox widget with specific horizontal and vertical alignments.
func NewAlign(child Widget, hAlign, vAlign layout.Align) *AlignBox {
	return &AlignBox{
		Child:  child,
		HAlign: hAlign,
		VAlign: vAlign,
	}
}

// Draw renders Child within r, offset by the computed alignment margins.
// Out-of-bounds or zero space is handled safely without drawing outside r.
func (a *AlignBox) Draw(c *Canvas, r Rect) {
	if a.Child == nil || r.W <= 0 || r.H <= 0 {
		return
	}

	targetW := a.Width
	if targetW <= 0 {
		if widther, ok := a.Child.(ContentWidther); ok {
			targetW = widther.ContentWidth()
		}
	}
	if targetW <= 0 || targetW > r.W {
		targetW = r.W
	}

	targetH := a.Height
	if targetH <= 0 {
		if heighter, ok := a.Child.(ContentHeighter); ok {
			targetH = heighter.ContentHeight()
		}
	}
	if targetH <= 0 || targetH > r.H {
		targetH = r.H
	}

	offsetX := layout.AlignOffset(r.W, targetW, a.HAlign)
	offsetY := layout.AlignOffset(r.H, targetH, a.VAlign)

	childRect := Rect{
		X: r.X + offsetX,
		Y: r.Y + offsetY,
		W: targetW,
		H: targetH,
	}

	a.Child.Draw(c, childRect)
}

// HandleKey delegates key events to the child widget.
func (a *AlignBox) HandleKey(e KeyEvent) (quit bool) {
	if a.Child != nil {
		return a.Child.HandleKey(e)
	}
	return false
}

// HandleMouse delegates mouse events to the child widget.
func (a *AlignBox) HandleMouse(e MouseEvent) (quit bool) {
	if a.Child != nil {
		return a.Child.HandleMouse(e)
	}
	return false
}

// ContentWidth estimates the child width.
func (a *AlignBox) ContentWidth() int {
	if a.Width > 0 {
		return a.Width
	}
	if widther, ok := a.Child.(ContentWidther); ok {
		return widther.ContentWidth()
	}
	return 0
}

// ContentHeight estimates the child height.
func (a *AlignBox) ContentHeight() int {
	if a.Height > 0 {
		return a.Height
	}
	if heighter, ok := a.Child.(ContentHeighter); ok {
		return heighter.ContentHeight()
	}
	return 0
}

// Measure implements Measurer.
func (a *AlignBox) Measure(width int) measure.Size {
	if measurer, ok := a.Child.(Measurer); ok {
		sz := measurer.Measure(width)
		if a.Width > 0 {
			sz.Width = a.Width
		}
		if a.Height > 0 {
			sz.Height = a.Height
		}
		return sz
	}
	return measure.Size{
		Width:  a.ContentWidth(),
		Height: a.ContentHeight(),
	}
}
