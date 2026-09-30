package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/internal/ptytest"
)

func buildLoomBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "loom")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	return bin
}

func TestInfoPTYReportAndWatch(t *testing.T) {
	bin := buildLoomBinary(t)
	info := ptytest.Start(t, 80, 24, bin, "info")
	info.WaitFor("loom version", time.Second*3)
	infoText := strings.Join(info.Screen(), "\n")
	if !strings.Contains(infoText, "80 x 24") {
		t.Fatalf("info report lacks PTY size: %q", infoText)
	}

	watch := ptytest.Start(t, 40, 12, bin, "info", "--watch")
	watch.WaitFor("pointer", time.Second*3)
	watch.SendRaw([]byte("\x1b[<35;8;5M"))
	// The alternate screen starts at terminal row 1, so raw SGR row 5 maps
	// directly to zero-based canvas row 4. Inline panes previously reserved
	// an earlier row below the prompt, making this expected value stale.
	watch.WaitFor("x=7 y=4 raw=8,5", time.Second*3)
	cells := watch.Cells()
	if got := cells[4][7].Rune; got != '+' {
		t.Fatalf("crosshair at (7,4) = %q, want '+'", got)
	}
	watch.Resize(52, 16)
	watch.WaitFor("52 x 16", time.Second*3)
	watch.Send("q")
}
