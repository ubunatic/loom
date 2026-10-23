package editor

import (
	"testing"
	"time"

	"ubunatic.com/loom"
)

func Test317EditSelectionTabBeforePanels(t *testing.T) {
	for _, key := range []string{"tab", "ctrl-i"} {
		t.Run(key, func(t *testing.T) {
			e := loom.NewRichTextEdit(&loom.RichDocument{Lines: []loom.RichLine{{Spans: []loom.RichSpan{{Text: "hello world"}}}}})
			v, err := newEditView(e, "", loom.EditorConfig{Theme: "plain"})
			if err != nil {
				t.Fatal(err)
			}
			v.showSidePanel, v.showSearch = true, true
			v.focused = focusEditor
			v.Draw(loom.NewCanvas(100, 24), loom.Rect{W: 100, H: 24})
			e.SetSelection(loom.RichPosition{}, loom.RichPosition{Offset: 5})
			// Exercise selection routing even when the toolbar is suppressed.
			e.ShowPopover = false
			if result := v.ConsumeKey(loom.KeyEvent{Key: key}); !result.Consumed || v.focused != focusEditor || !e.Document.Lines[0].Spans[0].Style.Italic {
				t.Fatal("selected Tab was routed to panels")
			}
			e.ClearSelection()
			for _, want := range []editFocus{focusBrowser, focusSearch, focusEditor} {
				v.ConsumeKey(loom.KeyEvent{Key: "tab"})
				if v.focused != want {
					t.Fatalf("plain Tab focus=%v, want %v", v.focused, want)
				}
			}
		})
	}
}

func Test317EditPTYFocusedPopoverHotkeyTogglesItalic(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, placement := range []string{"above", "below"} {
		t.Run(placement, func(t *testing.T) {
			s, _ := startEditPTY(t, "hello world\nhello world\nhello world\nhello world\n")
			row := 1
			if placement == "above" {
				s.Send("\x1b[B\x1b[B")
				s.WaitFor("Ln 3, Col 1", 3*time.Second)
				row = 3
			}
			s.Send("\x00")
			s.WaitFor("🎨", 3*time.Second)
			s.Send("i")
			waitEditPTY(t, s, func() bool { return s.Cell(1, row).Style.Italic })
			// Applying the first hotkey closes the popover and clears its
			// selection. Select the word again and reopen before toggling off.
			s.Send("\x1b[1;2D\x1b[1;2D\x1b[1;2D\x1b[1;2D\x1b[1;2D\x00")
			s.WaitFor("🎨", 3*time.Second)
			s.Send("i")
			waitEditPTY(t, s, func() bool { return !s.Cell(1, row).Style.Italic })
		})
	}
}
