package main

import (
	"strings"
	"testing"
	"time"
)

func Test318EditPTYFocusedPopoverArrowAway(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, placement := range []string{"above", "below"} {
		t.Run(placement, func(t *testing.T) {
			s, _ := startEditPTY(t, "hello world\nhello world\nhello world\nhello world\n")
			toward, away, status := "\x1b[B", "\x1b[A", "Ln 1, Col 6"
			if placement == "above" {
				s.Send("\x1b[B\x1b[B")
				s.WaitFor("Ln 3, Col 1", 3*time.Second)
				status = "Ln 3, Col 6"
				toward, away = "\x1b[A", "\x1b[B"
			}
			s.Send("\x00")
			s.WaitFor("#FG", 3*time.Second)
			s.Send(toward + away)
			waitEditPTY(t, s, func() bool {
				screen := strings.Join(s.Screen(), "\n")
				return !strings.Contains(screen, "#FG") && strings.Contains(screen, status)
			})
			s.Send("\x00")
			s.WaitFor("#FG", 3*time.Second)
			// Open Box and verify both vertical arrows move its highlighted choice.
			s.Send(toward + strings.Repeat("\x1b[C", 6) + "\r")
			s.WaitFor("Rounded", 3*time.Second)
			s.WaitFor("Plain", 3*time.Second)
			s.WaitFor("Rounded", 3*time.Second)
			s.Send("\x1b[B")
			waitEditPTY(t, s, func() bool {
				screen := strings.Join(s.Screen(), "\n")
				return strings.Contains(screen, "Plain") && strings.Contains(screen, "Rounded") && strings.Contains(screen, "#FG")
			})
			s.Send("\x1b[A")
			waitEditPTY(t, s, func() bool {
				screen := strings.Join(s.Screen(), "\n")
				return strings.Contains(screen, "Plain") && strings.Contains(screen, "Rounded") && strings.Contains(screen, "#FG")
			})
		})
	}
}
