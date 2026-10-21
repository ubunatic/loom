package main

import (
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

func Test357EditPTYEnterAppliesColorPickerSwatch(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s, _ := startEditPTY(t, "hello\n")
	s.Send("\x00\x1b[C\x1b[C\x1b[C\x1b[C\r") // Open bar, focus 🎨, open picker.
	s.WaitFor("Preview", 3*time.Second)
	s.Send("\x1b[C\rX") // Choose palette index 0 with Enter, then type.
	waitEditPTY(t, s, func() bool {
		text := strings.Join(s.Screen(), "\n")
		return strings.Contains(text, "X") && !strings.Contains(text, "Preview")
	})
	for _, row := range s.Cells() {
		for _, cell := range row {
			if cell.Text == "X" {
				if cell.Style.FG.Type != ptytest.ColorIndexed || cell.Style.FG.Index != 0 {
					t.Fatalf("typed X foreground = %+v, want indexed color 0", cell.Style.FG)
				}
				return
			}
		}
	}
	t.Fatalf("typed X not found on screen:\n%s", strings.Join(s.Screen(), "\n"))
}
