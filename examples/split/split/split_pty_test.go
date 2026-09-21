package split

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"codeberg.org/ubunatic/loom/internal/ptytest"
)

func TestSplitPTYScrollbarDrag(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "split")
	if out, err := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/examples/split").CombinedOutput(); err != nil {
		t.Fatalf("build split: %v\n%s", err, out)
	}
	s := ptytest.Start(t, 80, 24, bin)
	s.WaitFor("Left line", 3*time.Second)
	before := strings.Join(s.Screen(), "\n")
	x, y := findThumb(t, s.Screen())
	t.Logf("dragging thumb at terminal column=%d row=%d", x, y)
	s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM", x, y)))
	for _, row := range []int{y + 3, y + 6, y + 9} {
		s.SendRaw([]byte(fmt.Sprintf("\x1b[<32;%d;%dM", x, row)))
		time.Sleep(30 * time.Millisecond)
	}
	s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dm", x, y+9)))
	time.Sleep(100 * time.Millisecond)
	after := strings.Join(s.Screen(), "\n")
	if firstLine(t, before) >= firstLine(t, after) {
		t.Fatalf("drag did not move first visible line\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if err := writePTYEvidence(t, s.Screen()); err != nil {
		t.Fatal(err)
	}
}

// firstLine returns the number of the first visible "Left line NN" row.
func firstLine(t *testing.T, screen string) int {
	t.Helper()
	var n int
	i := strings.Index(screen, "Left line ")
	if i < 0 {
		t.Fatalf("no Left line in screen:\n%s", screen)
	}
	if _, err := fmt.Sscanf(screen[i:], "Left line %d", &n); err != nil {
		t.Fatalf("parse line number: %v", err)
	}
	return n
}

func findThumb(t *testing.T, screen []string) (int, int) {
	t.Helper()
	for y, row := range screen {
		if x := strings.Index(row, "▓"); x >= 0 {
			// SGR coordinates count cells, not bytes.
			return utf8.RuneCountInString(row[:x]) + 1, y + 1
		}
	}
	t.Fatal("scrollbar thumb not found")
	return 0, 0
}

func writePTYEvidence(t *testing.T, screen []string) error {
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		return nil
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return fmt.Errorf("go.mod not found")
		}
		root = parent
	}
	dir := filepath.Join(root, "docs", "progress", "092")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "M3-pty-drag.ansi"), []byte(strings.Join(screen, "\n")+"\n"), 0o644)
}
