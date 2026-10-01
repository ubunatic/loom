package filebrowser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

func TestFilebrowser093PTYClick(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha.txt", "beta.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(t.TempDir(), "filebrowser")
	if out, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/examples/filebrowser").CombinedOutput(); err != nil {
		t.Fatalf("build filebrowser: %v\n%s", err, out)
	}
	s := ptytest.Start(t, 100, 30, bin, "--theme", "plain", dir)
	s.WaitFor("alpha.txt", 5*time.Second)
	screen := s.Screen()
	if strings.Contains(strings.Join(screen, "\n"), "Name: alpha.txt") {
		t.Fatal("alpha.txt was selected before the text click")
	}
	row, col := findPTYText(t, screen, "alpha.txt")
	// SGR mouse coordinates are 1-based; the screen strings are cell-oriented.
	s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM", col+1, row+1)))
	time.Sleep(100 * time.Millisecond)
	if !strings.Contains(strings.Join(s.Screen(), "\n"), "Name: alpha.txt") {
		t.Fatal("text click did not select alpha.txt")
	}
	row, _ = findPTYText(t, s.Screen(), "alpha.txt")
	s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM", col+25, row+1)))
	time.Sleep(100 * time.Millisecond)
	if !strings.Contains(strings.Join(s.Screen(), "\n"), "Name: alpha.txt") {
		t.Fatal("whitespace click changed selection")
	}
	if os.Getenv("LOOM_EVIDENCE") == "1" {
		root := findFilebrowserRepoRoot(t)
		path := filepath.Join(root, "docs", "progress", "093", "M3-pty-click.ansi")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(strings.Join(s.Screen(), "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFilebrowserPTYDoubleClickOpensFileAndEntersDirectory(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha.txt", "beta.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	child := filepath.Join(dir, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(child, "nested.txt"), []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}
	openedFile := filepath.Join(t.TempDir(), "opened-files")
	openStubDir := t.TempDir()
	openStub := filepath.Join(openStubDir, "xdg-open")
	if err := os.WriteFile(openStub, []byte("#!/bin/sh\nprintf '%s\\n' \"$1\" >> \"$LOOM_TEST_OPENED_FILE\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LOOM_TEST_OPENED_FILE", openedFile)
	t.Setenv("PATH", openStubDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	bin := filepath.Join(t.TempDir(), "filebrowser")
	if out, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/examples/filebrowser").CombinedOutput(); err != nil {
		t.Fatalf("build filebrowser: %v\n%s", err, out)
	}
	s := ptytest.Start(t, 100, 30, bin, "--theme", "plain", dir)
	s.WaitFor("alpha.txt", 5*time.Second)

	screen := s.Screen()
	row, col := findPTYText(t, screen, "alpha.txt")
	sendPTYClick(s, row, col)
	s.WaitFor("Name: alpha.txt", 2*time.Second)
	time.Sleep(150 * time.Millisecond)
	screenText := strings.Join(s.Screen(), "\n")
	if strings.Contains(screenText, "Opening file") {
		t.Fatal("a single file click opened the file")
	}
	if _, err := os.Stat(openedFile); !os.IsNotExist(err) {
		t.Fatalf("a single file click invoked xdg-open; marker stat error = %v", err)
	}

	sendPTYClick(s, row, col)
	s.WaitFor("Opening file", 2*time.Second)
	openedPath := filepath.Join(dir, "alpha.txt")
	waitPTYFileContains(t, openedFile, openedPath, 2*time.Second)
	opened, err := os.ReadFile(openedFile)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(opened), openedPath); got != 1 {
		t.Fatalf("double-click invoked the opener %d times, want exactly once; marker=%q", got, opened)
	}

	screen = s.Screen()
	row, col = findPTYText(t, screen, "child")
	sendPTYClick(s, row, col)
	s.WaitFor("Name: child", 2*time.Second)
	time.Sleep(150 * time.Millisecond)
	if strings.Contains(strings.Join(s.Screen(), "\n"), "nested.txt") {
		t.Fatal("a single directory click entered the directory")
	}

	sendPTYClick(s, row, col)
	s.WaitFor("nested.txt", 2*time.Second)
}

func sendPTYClick(s *ptytest.Session, row, col int) {
	press := fmt.Sprintf("\x1b[<0;%d;%dM", col+1, row+1)
	release := fmt.Sprintf("\x1b[<0;%d;%dm", col+1, row+1)
	s.SendRaw([]byte(press))
	s.SendRaw([]byte(release))
}

func waitPTYFileContains(t *testing.T, path, want string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		contents, err := os.ReadFile(path)
		if err == nil && strings.Contains(string(contents), want) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	contents, err := os.ReadFile(path)
	t.Fatalf("timed out waiting for %q in opener marker %q: contents=%q err=%v", want, path, contents, err)
}

func TestFilebrowserPTYArrowMovesOneItem(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha.txt", "beta.txt", "gamma.txt", "delta.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(t.TempDir(), "filebrowser")
	if out, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/examples/filebrowser").CombinedOutput(); err != nil {
		t.Fatalf("build filebrowser: %v\n%s", err, out)
	}
	s := ptytest.Start(t, 100, 30, bin, "--theme", "plain", dir)
	s.WaitFor("▶ ..", 5*time.Second)
	s.SendRaw([]byte("\x1b[B"))
	time.Sleep(150 * time.Millisecond)
	screen := strings.Join(s.Screen(), "\n")
	if !strings.Contains(screen, "Name: alpha.txt") || !strings.Contains(screen, "▶ alpha.txt") {
		t.Fatalf("one down arrow did not select alpha.txt; screen:\n%s", screen)
	}
}

func findPTYText(t *testing.T, screen []string, text string) (row, col int) {
	t.Helper()
	for y, line := range screen {
		if x := strings.Index(line, text); x >= 0 {
			return y, len([]rune(line[:x]))
		}
	}
	t.Fatalf("text %q not found on screen", text)
	return 0, 0
}
