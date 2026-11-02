package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

// TestANSIViewPTYSpaceAndCtrlFPagesDown checks that space and ctrl-f page
// down in `loom view` (issue 368). Each key gets a fresh view so one key's
// scroll cannot mask the other's.
func TestANSIViewPTYSpaceAndCtrlFPagesDown(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	bin := buildLoomBinary(t)
	ansiFile := filepath.Join(t.TempDir(), "test.ansi")
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = fmt.Sprintf("Line %03d", i+1)
	}
	if err := os.WriteFile(ansiFile, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct{ name, key string }{
		{"space", " "},
		{"ctrl-f", "\x06"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			view := ptytest.Start(t, 40, 10, bin, "view", ansiFile)
			view.WaitFor("Line 001", 3*time.Second)
			view.Send(tc.key)
			view.WaitForUpdateState(3*time.Second, func(screen []string) bool {
				text := strings.Join(screen, "\n")
				return !strings.Contains(text, "Line 001") && strings.Contains(text, "Line 01")
			})
			view.Send("q")
		})
	}
}
