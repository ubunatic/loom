// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiviewer_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/internal/ptytest"
)

func buildViewer(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "ansiviewer")
	out, err := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/examples/ansiviewer").CombinedOutput()
	if err != nil {
		t.Fatalf("build ansiviewer: %v\n%s", err, out)
	}
	return bin
}

func TestViewerPTYShowsFilesAndQuits(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello from PTY\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := ptytest.Start(t, 100, 30, buildViewer(t), dir)
	s.WaitFor("hello.txt", 5*time.Second)
	if !strings.Contains(strings.Join(s.Screen(), "\n"), "Preview") {
		t.Fatal("viewer pane not rendered")
	}
	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("viewer did not exit: %v", err)
	}
}
