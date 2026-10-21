package main

import (
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
)

func editPTYPopoverBFocused(s *ptytest.Session) bool {
	for _, row := range s.Cells() {
		hasColorPicker := false
		for _, cell := range row {
			if cell.Text == "🎨" {
				hasColorPicker = true
				break
			}
		}
		if !hasColorPicker {
			continue
		}
		for _, cell := range row {
			if cell.Text == "B" && cell.Style.BG.Type == ptytest.ColorIndexed && cell.Style.BG.Index == uint8(loom.SpeccedDefaults.RichTextEdit.PopoverFocusBG) {
				return true
			}
		}
	}
	return false
}

func Test349EditPTYEitherVerticalArrowFocusesPopover(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("COLORTERM", "truecolor")
	for _, placement := range []string{"above", "below"} {
		for _, key := range []string{"\x1b[A", "\x1b[B"} {
			t.Run(placement+"/"+key, func(t *testing.T) {
				s, _ := startEditPTY(t, "hello world\nhello world\nhello world\nhello world\n")
				status := "Ln 1, Col 6"
				if placement == "above" {
					s.Send("\x1b[B\x1b[B")
					s.WaitFor("Ln 3, Col 1", 3*time.Second)
					status = "Ln 3, Col 6"
				}
				s.Send("\x00")
				s.WaitFor("🎨", 3*time.Second)
				s.Send(key)
				waitEditPTY(t, s, func() bool { return editPTYPopoverBFocused(s) })
				if !strings.Contains(strings.Join(s.Screen(), "\n"), status) {
					t.Fatal("vertical arrow moved the editor cursor instead of focusing the toolbar")
				}
			})
		}
	}
}
