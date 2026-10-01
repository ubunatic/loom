// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strconv"
	"strings"
)

// NumberInput edits a bounded floating-point value. Left and right step the
// value; Enter starts inline editing, Enter commits, and Esc cancels.
type NumberInput struct {
	Value  *float64
	Min    float64
	Max    float64
	Step   float64
	Format string
	// FixedWidth pads the complete control to a stable display width. Zero keeps the natural width.
	FixedWidth int
	// Align controls value alignment inside FixedWidth; the default is left.
	Align Align

	editor   *TextInput
	orig     float64
	err      string
	lastRect Rect
}

// NewNumberInput creates a bounded numeric input. A zero Step defaults to 1.
func NewNumberInput(value *float64, min, max float64) *NumberInput {
	return &NumberInput{Value: value, Min: min, Max: max}
}

// Draw renders the current value, or the active editor when focused.
func (n *NumberInput) Draw(c *Canvas, r Rect) {
	n.lastRect = r
	c.PaintSurface(r, Style{})
	if n.editor != nil {
		n.editor.Draw(c, r, true)
		return
	}
	c.Write(r.X, r.Y, TruncateText(n.renderedText(), r.W, ""), Style{})
}

func (n *NumberInput) renderedText() string {
	if n == nil || n.Value == nil {
		return ""
	}
	text := n.String()
	if n.FixedWidth > 0 {
		value := n.format(*n.Value)
		padding := strings.Repeat(" ", max(0, n.FixedWidth-StringWidth(value)-4))
		if n.Align == AlignRight {
			text = "◂ " + padding + value + " ▸"
		} else {
			text = "◂ " + value + padding + " ▸"
		}
	}
	return text
}

// String returns the formatted current value with step indicators.
func (n *NumberInput) String() string {
	if n == nil || n.Value == nil {
		return ""
	}
	return "◂ " + n.format(*n.Value) + " ▸"
}

// Error returns validation feedback from the most recent failed commit.
func (n *NumberInput) Error() string {
	if n == nil {
		return ""
	}
	return n.err
}

// Editing reports whether inline editing is active.
func (n *NumberInput) Editing() bool { return n != nil && n.editor != nil }

// ConsumeKey steps, edits, commits, or cancels the value.
func (n *NumberInput) ConsumeKey(e KeyEvent) (quit EventResult) {
	if n == nil {
		return Ignored()
	}
	if n.editor != nil {
		if n.handleEditKey(e) {
			return Handled()
		}
		return Ignored()
	}
	switch e.Key {
	case "enter":
		n.beginEdit()
	case "left", "minus":
		n.StepBy(-1)
	case "right", "plus":
		n.StepBy(1)
	default:
		if e.Text == "+" {
			n.StepBy(1)
		}
		if e.Text == "-" {
			n.StepBy(-1)
		}
	}
	return Ignored()
}

// ConsumeMouse handles mouse scroll wheel and click stepping.
func (n *NumberInput) ConsumeMouse(e MouseEvent) EventResult {
	if n == nil || n.Value == nil || n.editor != nil {
		return Ignored()
	}
	if n.lastRect.W > 0 && (e.X < 0 || e.X >= n.lastRect.W) {
		return Ignored()
	}
	if n.lastRect.H > 0 && (e.Y < 0 || e.Y >= n.lastRect.H) {
		return Ignored()
	}
	switch e.Action {
	case MouseScrollUp:
		n.StepBy(1)
		return Handled()
	case MouseScrollDown:
		n.StepBy(-1)
		return Handled()
	case MousePress:
		if e.Button != MouseLeft || e.Y != 0 {
			return Ignored()
		}
		width := StringWidth(n.renderedText())
		if n.lastRect.W > 0 && width > n.lastRect.W {
			width = n.lastRect.W
		}
		if width <= 0 || e.X < 0 || e.X >= width {
			return Ignored()
		}
		if e.X < 2 {
			n.StepBy(-1)
			return Handled()
		}
		if e.X >= width-2 {
			n.StepBy(1)
			return Handled()
		}
		n.beginEdit()
		return Handled()
	}
	return Ignored()
}

// StepBy changes the value by direction times Step and clamps it to the bounds.
func (n *NumberInput) StepBy(direction float64) {
	if n == nil || n.Value == nil {
		return
	}
	if n.Min > n.Max {
		*n.Value = n.Min
		return
	}
	step := n.Step
	if step == 0 {
		step = 1
	}
	value := *n.Value + direction*step
	if value < n.Min {
		value = n.Min
	} else if value > n.Max {
		value = n.Max
	}
	*n.Value = value
}

func (n *NumberInput) format(value float64) string {
	format := n.Format
	if format == "" || strings.Count(format, "%") != 1 {
		format = "%g"
	}
	return fmt.Sprintf(format, value)
}

func (n *NumberInput) beginEdit() {
	if n.Value == nil {
		return
	}
	n.orig = *n.Value
	n.editor = NewTextInput(n.format(*n.Value))
	n.err = ""
}

func (n *NumberInput) handleEditKey(e KeyEvent) bool {
	switch e.Key {
	case "enter":
		if n.editor == nil {
			n.err = "invalid number"
			return false
		}
		value, err := strconv.ParseFloat(n.editor.Value(), 64)
		if err != nil {
			n.err = "invalid number"
			return false
		}
		if n.Min > n.Max {
			value = n.Min
		} else if value < n.Min {
			n.err = fmt.Sprintf("%s is below the minimum %s", n.editor.Value(), n.format(n.Min))
			return false
		} else if value > n.Max {
			n.err = fmt.Sprintf("%s is above the maximum %s", n.editor.Value(), n.format(n.Max))
			return false
		}
		if n.Value != nil {
			*n.Value = value
		}
		n.endEdit()
	case "esc", "ctrl-c":
		if n.Value != nil {
			*n.Value = n.orig
		}
		n.endEdit()
	default:
		if n.editor == nil {
			return false
		}
		if e.Text == "-" {
			if n.Min < 0 && n.editor.Caret() == 0 && !strings.Contains(n.editor.Value(), "-") {
				n.editor.ConsumeKey(e)
				n.err = ""
			}
			return false
		}
		if e.Text == "." && !strings.Contains(n.editor.Value(), ".") {
			n.editor.ConsumeKey(e)
			n.err = ""
			return false
		}
		if e.Text != "" {
			valid := true
			invalidDot := false
			for _, r := range e.Text {
				if r < '0' || r > '9' {
					valid = false
					invalidDot = r == '.'
					break
				}
			}
			if valid {
				n.editor.ConsumeKey(e)
				n.err = ""
			} else if invalidDot || strings.Contains(e.Text, ".") {
				n.err = "invalid number"
			}
			return false
		}
		n.editor.ConsumeKey(e)
		n.err = ""
	}
	return false
}

func (n *NumberInput) endEdit() { n.editor = nil; n.err = "" }
