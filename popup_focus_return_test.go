package loom

import (
	"testing"
)

func TestPopupFocusReturn(t *testing.T) {
	tests := []struct {
		name string
		closeFunc func(t *testing.T, s *Stack, f *Frame, p *Popup)
	}{
		{"esc", func(t *testing.T, s *Stack, f *Frame, p *Popup) {
			if s != nil {
				s.ConsumeKey(KeyEvent{Key: "esc"})
			} else if f != nil {
				f.ConsumeKey(KeyEvent{Key: "esc"})
			}
		}},
		{"outside click", func(t *testing.T, s *Stack, f *Frame, p *Popup) {
			p.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: -10, Y: -10})
		}},
		{"programmatic", func(t *testing.T, s *Stack, f *Frame, p *Popup) {
			p.Open = false
		}},
	}

	for _, tt := range tests {
		t.Run("Stack_"+tt.name, func(t *testing.T) {
			input := NewTextEntry(TextEntryOptions{})
			popupInput := NewTextEntry(TextEntryOptions{})
			popup := NewPopup("Test", popupInput)
			stack := NewStack(Vertical, input, popup)

			stack.ConsumeKey(KeyEvent{Key: "tab"}) // Focus popup
			tt.closeFunc(t, stack, nil, popup) // Close it

			stack.ConsumeKey(KeyEvent{Key: "x"})
			if input.Value() != "x" {
				t.Errorf("got input value %q, want 'x'", input.Value())
			}
		})

		t.Run("Frame_"+tt.name, func(t *testing.T) {
			input := NewTextEntry(TextEntryOptions{})
			popupInput := NewTextEntry(TextEntryOptions{})
			popup := NewPopup("Test", popupInput)
			frame := &Frame{Boxes: []Box{{Child: input}, {Child: popup}}}

			frame.ConsumeKey(KeyEvent{Key: "tab"}) // Focus popup
			tt.closeFunc(t, nil, frame, popup) // Close it

			frame.ConsumeKey(KeyEvent{Key: "x"})
			if input.Value() != "x" {
				t.Errorf("got input value %q, want 'x'", input.Value())
			}
		})
	}
}
