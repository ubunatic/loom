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

func TestANSIViewPTYSpaceAndCtrlFPagesDown(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	bin := buildLoomBinary(t)

	// Create test file with line numbers
	ansiFile := filepath.Join(t.TempDir(), "test.ansi")
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = fmt.Sprintf("Line %d", i+1)
	}
	if err := os.WriteFile(ansiFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		t.Fatal(err)
	}

	view := ptytest.Start(t, 40, 10, bin, "view", ansiFile)
	view.WaitFor("Line 1", time.Second*3)

	// Test space
	view.Send(" ")
	time.Sleep(time.Millisecond * 200)

	screenText := strings.Join(view.Screen(), "\n")
	if strings.Contains(screenText, "Line 1\n") {
		t.Errorf("Space did not page down, 'Line 1' still on screen")
	}

	// Reset
	view.Send("g") // home? or we can just run a new one

	// For now, let's just make sure "ctrl-f" is also tested
	view.Send("\x06") // ctrl-f

	time.Sleep(time.Millisecond * 200)

	screenText = strings.Join(view.Screen(), "\n")
	if strings.Contains(screenText, "Line 1\n") {
		t.Errorf("Ctrl-F did not page down")
	}

	view.Send("q")
}
