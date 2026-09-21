package filebrowser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/internal/ptytest"
)

func TestFilebrowser093PTYClick(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha.txt", "beta.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(t.TempDir(), "filebrowser")
	if out, err := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/examples/filebrowser").CombinedOutput(); err != nil {
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
