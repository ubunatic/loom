// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiviewer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestBrowserListsTextAndANSI(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "plain.txt"), []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "color.ansi"), []byte("\x1b[31mred\x1b[0m\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.files) != 2 || b.files[0].Name() != "color.ansi" {
		t.Fatalf("files = %v", b.files)
	}
	b.selectFile(0)
	c := loom.NewCanvas(30, 4)
	b.Draw(c, c.Bounds())
	if got := c.Row(0); got == "" {
		t.Fatal("empty rendered row")
	}
}

func TestANSIWriteClipsToBounds(t *testing.T) {
	c := loom.NewCanvas(5, 1)
	writeANSI(c, loom.Rect{X: 0, Y: 0, W: 5, H: 1}, "abcdef\x1b[31m!")
	if got := c.Row(0); got == "" {
		t.Fatal("empty row")
	}
}

func TestClassifyBinaryAndImage(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "data.bin")
	if err := os.WriteFile(bin, []byte{'x', 0, 'y'}, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Classify(bin); got != KindBinary {
		t.Fatalf("binary kind = %q", got)
	}
	if got := Classify(filepath.Join(dir, "photo.png")); got != KindImage {
		t.Fatalf("image kind = %q", got)
	}
}

func TestClassifyANSIByExtension(t *testing.T) {
	if got := Classify("sample.ansi"); got != KindANSI {
		t.Fatalf("ANSI kind = %q", got)
	}
}

func TestBrowserMetadataAndScroll(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "data.bin"), []byte{'x', 0}, 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.metadata, "binary") || len(b.lines) != 2 {
		t.Fatalf("metadata=%q lines=%v", b.metadata, b.lines)
	}
	for i := 0; i < 20; i++ {
		b.lines = append(b.lines, "line")
	}
	b.HandleKey(loom.KeyEvent{Key: "pgdn"})
	if b.offset == 0 {
		t.Fatal("pgdn did not scroll")
	}
}
