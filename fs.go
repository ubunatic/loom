package loom

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// FileKind identifies the filesystem kind reported by a directory entry.
type FileKind uint8

const (
	// FileKindRegular is a regular file.
	FileKindRegular FileKind = iota
	// FileKindDirectory is a directory.
	FileKindDirectory
	// FileKindSymlink is a symbolic link. Its target is not followed while reading.
	FileKindSymlink
	// FileKindSpecial is any other filesystem object.
	FileKindSpecial
)

// FileEntry describes one entry in a Directory.
type FileEntry struct {
	Name     string
	Path     string
	Kind     FileKind
	IsParent bool
}

// DisplayName returns the entry name escaped for safe terminal display.
func (e FileEntry) DisplayName() string { return QuoteUnprintable(e.Name) }

// Directory is a filesystem directory and its entries.
type Directory struct {
	Path    string
	Entries []FileEntry
}

// DirectoryOptions controls which navigation entries ReadDirectory returns.
type DirectoryOptions struct {
	IncludeParent bool
}

// ReadDirectory reads path without following symlinks merely to classify entries.
func ReadDirectory(path string, opts DirectoryOptions) (Directory, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return Directory{}, fmt.Errorf("resolve directory %q: %w", path, err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return Directory{}, fmt.Errorf("stat directory %q: %w", absPath, err)
	}
	if !info.IsDir() {
		return Directory{}, fmt.Errorf("%s is not a directory", absPath)
	}
	osEntries, err := os.ReadDir(absPath)
	if err != nil {
		return Directory{}, fmt.Errorf("read directory %q: %w", absPath, err)
	}

	dir := Directory{Path: absPath, Entries: make([]FileEntry, 0, len(osEntries)+1)}
	if parent, ok := dir.Parent(); opts.IncludeParent && ok {
		dir.Entries = append(dir.Entries, FileEntry{
			Name: "..", Path: parent, Kind: FileKindDirectory, IsParent: true,
		})
	}
	for _, entry := range osEntries {
		dir.Entries = append(dir.Entries, FileEntry{
			Name: entry.Name(), Path: dir.EntryPath(entry.Name()), Kind: fileKind(entry),
		})
	}
	return dir, nil
}

// Parent returns the parent directory and reports false at a volume root.
func (d Directory) Parent() (string, bool) {
	parent := filepath.Dir(d.Path)
	return parent, parent != d.Path
}

// EntryPath resolves a child name relative to the directory.
func (d Directory) EntryPath(name string) string { return filepath.Join(d.Path, name) }

// EntryMap maps raw entry names to their entries.
func (d Directory) EntryMap() map[string]FileEntry {
	entries := make(map[string]FileEntry, len(d.Entries))
	for _, entry := range d.Entries {
		entries[entry.Name] = entry
	}
	return entries
}

func fileKind(entry os.DirEntry) FileKind {
	if entry.Type()&os.ModeSymlink != 0 {
		return FileKindSymlink
	}
	if entry.IsDir() {
		return FileKindDirectory
	}
	if entry.Type().IsRegular() {
		return FileKindRegular
	}
	return FileKindSpecial
}

// DisplayPath formats a path for safe terminal display.
func DisplayPath(path string) string { return QuoteUnprintable(path) }

// QuoteUnprintable returns value unchanged when every rune is printable and an
// ASCII-escaped, quoted representation otherwise.
func QuoteUnprintable(value string) string {
	if !utf8.ValidString(value) {
		return strconv.QuoteToASCII(value)
	}
	for _, r := range value {
		if !unicode.IsPrint(r) {
			return strconv.QuoteToASCII(value)
		}
	}
	return value
}
