// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package apitest

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "accept reviewed exported API snapshots")

type listedPackage struct {
	ImportPath string
	Name       string
}

// TestPublicAPI covers every importable package, including experimental examples.
// go doc elides unexported implementation details but retains signatures, struct
// fields, interfaces, constant declarations and the documentation contract.
func TestPublicAPI(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = root
	data, err := cmd.Output()
	if err != nil {
		t.Fatalf("list packages: %v", err)
	}
	dir := filepath.Join(root, "testdata", "api")
	if *update {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	seen := make(map[string]bool)
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	for {
		var pkg listedPackage
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if pkg.Name == "main" || strings.Contains(pkg.ImportPath, "/internal/") {
			continue
		}
		relative := strings.TrimPrefix(pkg.ImportPath, "ubunatic.com/loom")
		name := "loom"
		if relative != "" {
			name = strings.ReplaceAll(strings.TrimPrefix(relative, "/"), "/", "__")
		}
		name += ".txt"
		seen[name] = true
		t.Run(strings.TrimPrefix(pkg.ImportPath, "ubunatic.com/"), func(t *testing.T) {
			doc := exec.Command("go", "doc", "-all", pkg.ImportPath)
			doc.Dir = root
			got, err := doc.Output()
			if err != nil {
				t.Fatalf("go doc: %v", err)
			}
			path := filepath.Join(dir, name)
			if *update {
				if err := os.WriteFile(path, got, 0644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("missing API snapshot %s: %v; see docs/APIStability.md", name, err)
			}
			if string(got) != string(want) {
				t.Errorf("exported API differs from testdata/api/%s; review the API diff, then run go test ./internal/apitest -update", name)
				gotLines, wantLines := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
				for i := range min(len(gotLines), len(wantLines)) {
					if gotLines[i] != wantLines[i] {
						t.Logf("first difference at line %d: got %q, want %q", i+1, gotLines[i], wantLines[i])
						break
					}
				}
			}
		})
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") || seen[entry.Name()] {
			continue
		}
		if *update {
			if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
				t.Fatal(err)
			}
		} else {
			t.Errorf("stale API snapshot %s: package removed or made internal; review before -update", entry.Name())
		}
	}
}
