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
			wantR, wantG, wantB, wantOK := loom.Theme("julia256").Selection.Color().RGB()
			gotR, gotG, gotB, gotOK := cell.Style.BG.RGB()
			if cell.Text == "B" && wantOK && gotOK && gotR == wantR && gotG == wantG && gotB == wantB {
				return true
			}
		}
	}
	return false
}

func Test349EditPTYCtrlSpaceFocusesFirstPopoverAction(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("COLORTERM", "truecolor")
	for _, placement := range []string{"above", "below"} {
		t.Run(placement, func(t *testing.T) {
			s, _ := startEditPTY(t, "hello world\nhello world\nhello world\nhello world\n")
			if placement == "above" {
				s.Send("\x1b[B\x1b[B")
				s.WaitFor("Ln 3, Col 1", 3*time.Second)
			}
			s.Send("\x00")
			s.WaitFor("🎨", 3*time.Second)
			waitEditPTY(t, s, func() bool { return editPTYPopoverBFocused(s) })
			s.Send("b")
			waitEditPTY(t, s, func() bool { return strings.Contains(strings.Join(s.Screen(), "\n"), "Modified") })
		})
	}
}
