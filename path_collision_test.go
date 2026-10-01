package loom

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaseInsensitivePathCollisions(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"Dialog.ansi", "dialog.ansi", "one/readme.md", "two/README.md"} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	paths := []string{"Dialog.ansi", "dialog.ansi", "one/readme.md", "two/README.md"}
	got := caseInsensitivePathCollisions(paths)
	want := [][2]string{{"Dialog.ansi", "dialog.ansi"}}
	if len(got) != len(want) {
		t.Fatalf("collisions = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("collision[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNoCaseInsensitiveTrackedPathCollisions(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is unavailable")
	}
	if _, err := os.Stat(".git"); err != nil {
		t.Skip(".git directory is unavailable")
	}

	cmd := exec.Command("git", "ls-files", "-z")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	var paths []string
	for _, path := range strings.Split(string(output), "\x00") {
		if path != "" {
			paths = append(paths, path)
		}
	}
	if collisions := caseInsensitivePathCollisions(paths); len(collisions) != 0 {
		t.Errorf("tracked paths collide case-insensitively:")
		for _, pair := range collisions {
			t.Errorf("  %q and %q", pair[0], pair[1])
		}
	}
}
