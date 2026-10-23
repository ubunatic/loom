package editor

import (
	"strings"
	"testing"
	"time"
)

func Test318EditPTYFocusedPopoverArrowAway(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, placement := range []string{"above", "below"} {
		t.Run(placement, func(t *testing.T) {
			s, _ := startEditPTY(t, "hello world\nhello world\nhello world\nhello world\n")
			away, status := "\x1b[A", "Ln 1, Col 6"
			if placement == "above" {
				s.Send("\x1b[B\x1b[B")
				s.WaitFor("Ln 3, Col 1", 3*time.Second)
				status = "Ln 3, Col 6"
				away = "\x1b[B"
			}
			s.Send("\x00")
			s.WaitFor("🎨", 3*time.Second)
			s.Send(away)
			waitEditPTY(t, s, func() bool {
				screen := strings.Join(s.Screen(), "\n")
				return !strings.Contains(screen, "🎨") && strings.Contains(screen, status)
			})
			s.Send("\x00")
			s.WaitFor("🎨", 3*time.Second)
			// Open Box and verify both vertical arrows move its highlighted choice.
			s.Send(strings.Repeat("\x1b[C", 5) + "\r")
			s.WaitFor("Rounded", 3*time.Second)
			s.WaitFor("Plain", 3*time.Second)
			s.WaitFor("Rounded", 3*time.Second)
			s.Send("\x1b[B")
			waitEditPTY(t, s, func() bool {
				screen := strings.Join(s.Screen(), "\n")
				return strings.Contains(screen, "Plain") && strings.Contains(screen, "Rounded") && strings.Contains(screen, "🎨")
			})
			s.Send("\x1b[A")
			waitEditPTY(t, s, func() bool {
				screen := strings.Join(s.Screen(), "\n")
				return strings.Contains(screen, "Plain") && strings.Contains(screen, "Rounded") && strings.Contains(screen, "🎨")
			})
		})
	}
}

func Test358EditPTYFocusedEntryArrowClosesAndClearsSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, placement := range []string{"above", "below"} {
		t.Run(placement, func(t *testing.T) {
			s, _ := startEditPTY(t, "hello world\nhello world\nhello world\nhello world\n")
			entry := "\x1b[A"
			if placement == "above" {
				s.Send("\x1b[B\x1b[B")
				s.WaitFor("Ln 3, Col 1", 3*time.Second)
			} else {
				entry = "\x1b[B"
			}
			s.Send("\x00")
			s.WaitFor("🎨", 3*time.Second)
			s.Send(entry + "X")
			waitEditPTY(t, s, func() bool {
				screen := strings.Join(s.Screen(), "\n")
				return !strings.Contains(screen, "🎨") && strings.Contains(screen, "helloX world")
			})
		})
	}
}
