package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLiveTmuxScriptRunsStepsAndMeasuresPause(t *testing.T) {
	if testing.Short() {
		t.Skip("real tmux integration test")
	}
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		t.Skip("tmux unavailable")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(root, "scripts", "live-tmux.sh")
	if _, err := os.Stat(script); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(t.TempDir(), "want.ansi")
	if err := os.WriteFile(want, []byte("ready\n"), 0600); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	outputDir := t.TempDir()
	cmd := exec.Command("bash", script, "-y", "2", "--wait", "ready", "--timeout", "2", "--output-dir", outputDir, "--", "bash", "-c", "printf 'ready\\n'; read -r line; printf 'done\\n'; read -r line")
	cmd.Env = append(os.Environ(), "PATH="+filepath.Dir(tmux)+":"+os.Getenv("PATH"))
	cmd.Stdin = strings.NewReader("pause 0.12\nsnap before\ndiff " + want + "\nkeys Enter\nwait done\nsnap after\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("live tmux script: %v\n%s", err, output)
	}
	if elapsed := time.Since(started); elapsed < 100*time.Millisecond {
		t.Fatalf("pause did not wait long enough: %v", elapsed)
	}
	for _, text := range []string{"ready", "done", "screens match"} {
		if !strings.Contains(string(output), text) {
			t.Errorf("output missing %q:\n%s", text, output)
		}
	}
	for _, name := range []string{"002-before.ansi", "006-after.ansi"} {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Errorf("snapshot %q missing: %v", name, err)
		}
	}

	wrong := filepath.Join(t.TempDir(), "wrong.ansi")
	if err := os.WriteFile(wrong, []byte("wrong\n"), 0600); err != nil {
		t.Fatal(err)
	}
	diff := exec.Command("bash", script, "-y", "2", "--wait", "ready", "--timeout", "2", "--", "bash", "-c", "printf 'ready\\n'; read -r line")
	diff.Stdin = strings.NewReader("diff " + wrong + "\n")
	diffOutput, err := diff.CombinedOutput()
	if err == nil || !strings.Contains(string(diffOutput), "screens differ") {
		t.Fatalf("diff accepted a mismatched pane: err=%v output=%s", err, diffOutput)
	}

	lastRowDir := t.TempDir()
	lastRow := exec.Command("bash", script, "-y", "4", "--wait", "bottom", "--timeout", "2", "--output-dir", lastRowDir, "--", "bash", "-c", "printf 'top\\nmiddle\\nthird\\nbottom\\n'; read -r line")
	lastRow.Stdin = strings.NewReader("snap last-row\n")
	lastRowOutput, err := lastRow.CombinedOutput()
	if err != nil || !strings.Contains(string(lastRowOutput), "bottom") {
		t.Fatalf("wait missed text on the last visible pane row: err=%v output=%s", err, lastRowOutput)
	}
}

func TestLiveTmuxTimeoutKillsItsServer(t *testing.T) {
	if testing.Short() {
		t.Skip("real tmux integration test")
	}
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux unavailable")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	cmd := exec.Command("bash", filepath.Join(root, "scripts", "live-tmux.sh"), "--wait", "missing text", "--timeout", "0.2", "--interval", "0.02", "--", "bash", "-c", "read -r line")
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "timed out waiting for pane text") {
		t.Fatalf("expected wait timeout, err=%v output=%s", err, output)
	}
	if elapsed := time.Since(started); elapsed < 150*time.Millisecond || elapsed > 2*time.Second {
		t.Fatalf("timeout did not honor its real duration: %v", elapsed)
	}
	var socket string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "tmux socket: ") {
			socket = strings.TrimPrefix(line, "tmux socket: ")
		}
	}
	if socket == "" {
		t.Fatalf("script did not report isolated socket: %s", output)
	}
	check := exec.Command("tmux", "-L", socket, "ls")
	if err := check.Run(); err == nil {
		t.Fatalf("tmux server %q survived timeout", socket)
	}
}
