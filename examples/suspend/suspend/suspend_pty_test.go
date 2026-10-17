// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package suspend_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

// TestSuspendEditsNoteInEditor runs the real binary with an $EDITOR that
// reads one line from the terminal: the line reaches the editor, not the pane,
// and the resumed pane shows the edited note.
func TestSuspendEditsNoteInEditor(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	dir := t.TempDir()
	bin := filepath.Join(dir, "suspend")
	if out, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/examples/suspend").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	editor := filepath.Join(dir, "editor.sh")
	script := "#!/bin/sh\nprintf 'EDITOR-READY\\n'\nread line\nprintf '%s\\n' \"$line\" > \"$1\"\n"
	if err := os.WriteFile(editor, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", editor)

	s := ptytest.Start(t, 60, 12, bin)
	s.WaitFor("Ctrl+G", 5*time.Second)
	s.Send("\x07")
	s.WaitFor("EDITOR-READY", 5*time.Second)
	s.Send("typed note\r")
	s.WaitFor("note updated", 5*time.Second)
	if screen := strings.Join(s.Screen(), "\n"); !strings.Contains(screen, "typed note") {
		t.Fatalf("edited note missing after resume:\n%s", screen)
	}
	s.Send("q")
	if err := s.Wait(5 * time.Second); err != nil {
		t.Fatalf("clean exit: %v", err)
	}
}
