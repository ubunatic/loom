package editor

import (
	"strings"
	"testing"
	"time"
)

func Test356EditPTYPlainArrowCollapsesUnfocusedPopoverSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, tc := range []struct {
		name, arrow, want string
	}{
		{"left", "\x1b[D", "Xhello world"},
		{"right", "\x1b[C", "helloX world"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := startEditPTY(t, "hello world\n")
			s.Send("\x1b[1;2C\x1b[1;2C\x1b[1;2C\x1b[1;2C\x1b[1;2C")
			s.WaitFor("🎨", 3*time.Second)
			s.Send(tc.arrow)
			s.Send("X")
			waitEditPTY(t, s, func() bool {
				text := strings.Join(s.Screen(), "\n")
				return strings.Contains(text, tc.want) && !strings.Contains(text, "🎨")
			})
		})
	}
}
