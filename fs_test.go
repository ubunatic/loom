package loom

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDirectoryNavigationAndKinds(t *testing.T) {
	root := t.TempDir()
	dirPath := filepath.Join(root, "child")
	if err := os.Mkdir(dirPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dirPath, "file.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dirPath, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("sub", filepath.Join(dirPath, "link")); err != nil {
		t.Fatal(err)
	}

	dir, err := ReadDirectory(dirPath, DirectoryOptions{IncludeParent: true})
	if err != nil {
		t.Fatal(err)
	}
	parent, ok := dir.Parent()
	if !ok || parent != root || dir.Path != dirPath {
		t.Fatalf("directory = %+v, parent = %q, ok = %v", dir, parent, ok)
	}
	entries := dir.EntryMap()
	tests := []struct {
		name string
		kind FileKind
	}{
		{name: "..", kind: FileKindDirectory},
		{name: "file.txt", kind: FileKindRegular},
		{name: "link", kind: FileKindSymlink},
		{name: "sub", kind: FileKindDirectory},
	}
	for _, test := range tests {
		entry, found := entries[test.name]
		if !found || entry.Kind != test.kind {
			t.Errorf("entry %q = %+v, found = %v; want kind %v", test.name, entry, found, test.kind)
		}
		if test.name != ".." && entry.Path != dir.EntryPath(test.name) {
			t.Errorf("entry %q path = %q", test.name, entry.Path)
		}
	}
	if !entries[".."].IsParent || entries[".."].Path != root {
		t.Fatalf("parent entry = %+v", entries[".."])
	}
}

func TestReadDirectoryRejectsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadDirectory(path, DirectoryOptions{}); err == nil {
		t.Fatal("ReadDirectory accepted a regular file")
	}
}

func TestQuoteUnprintable(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "printable path/é", want: "printable path/é"},
		{value: "line\nbreak", want: `"line\nbreak"`},
		{value: "escape\x1b", want: `"escape\x1b"`},
		{value: string([]byte{'b', 'a', 'd', 0xff}), want: `"bad\xff"`},
	}
	for _, test := range tests {
		if got := QuoteUnprintable(test.value); got != test.want {
			t.Errorf("QuoteUnprintable(%q) = %q, want %q", test.value, got, test.want)
		}
		if got := DisplayPath(test.value); got != test.want {
			t.Errorf("DisplayPath(%q) = %q, want %q", test.value, got, test.want)
		}
	}
}
