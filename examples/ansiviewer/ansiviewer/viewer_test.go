// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiviewer

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom"
	"golang.org/x/sys/unix"
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
	c := loom.NewCanvas(24, 2)
	c.Write(0, 0, "LEFT", loom.Style{})
	c.Write(18, 0, "RIGHT", loom.Style{})
	before := c.Row(0)
	writeANSI(c, loom.Rect{X: 4, Y: 0, W: 6, H: 1}, "界\x1b[38;2;1;2;3mCJK long text")
	after := c.Row(0)
	if !strings.Contains(after, "LEFT") || !strings.Contains(after, "RIGHT") {
		t.Fatalf("outside cells changed: before=%q after=%q", before, after)
	}
	if !strings.Contains(after, "界") || !strings.Contains(after, "CJK") {
		t.Fatalf("bounded content missing: %q", after)
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

func TestRecordWritesOneSnapshotAfterDelay(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "record.txt"), []byte("recorded\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Record(context.Background(), &out, dir, time.Millisecond, 40, 8); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "record.txt") {
		t.Fatalf("recording does not show file: %q", out.String())
	}
}

func TestRecordCommandReapsSelfExitingChild(t *testing.T) {
	var out bytes.Buffer
	if err := RecordCommand(context.Background(), &out, time.Second, "sh", "-c", "printf '\\033[2Jself-exit'"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "self-exit") {
		t.Fatalf("capture = %q", out.String())
	}
}

func TestRecordCommandStopsLongLivedChild(t *testing.T) {
	var out bytes.Buffer
	if err := RecordCommand(context.Background(), &out, 20*time.Millisecond, "sh", "-c", "printf long-lived; trap 'exit 0' TERM; while :; do sleep 1; done"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "long-lived") {
		t.Fatalf("capture = %q", out.String())
	}
}

func TestRecordLongLivedKeepsPreTeardownScreen(t *testing.T) {
	var out bytes.Buffer
	if err := RecordCommand(context.Background(), &out, 40*time.Millisecond, "sh", "-c", "printf 'harnez usage'; trap 'exit 0' TERM; while :; do sleep 1; done"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "harnez usage") {
		t.Fatalf("pre-teardown screen = %q", out.String())
	}
}

func TestRunRecordWritesToStdout(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--record", "1ms", "-o", "-", "--", "sh", "-c", "printf cli"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "cli") {
		t.Fatalf("CLI recording = %q", out.String())
	}
}

func TestRecordCommandWritesRenderedScreenOnly(t *testing.T) {
	var out bytes.Buffer
	if err := RecordCommand(context.Background(), &out, time.Second, "sh", "-c", "printf '\\033[2J\\033[31mred\\033[0m'"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "red") {
		t.Fatalf("screen = %q", out.String())
	}
	if !strings.Contains(out.String(), "\x1b[31m") {
		t.Fatalf("ANSI styling was discarded: %q", out.String())
	}
}

func TestRecordCommandStripsTerminalStateButKeepsANSIStyles(t *testing.T) {
	var out bytes.Buffer
	command := "printf '\033[?1049h\033]0;title\033\\\r\033[31mred\033[0m\033[?1049l'"
	if err := RecordCommand(context.Background(), &out, time.Second, "sh", "-c", command); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "\r\x1b[31mred\x1b[0m" {
		t.Fatalf("recording = %q, want preserved carriage return and color", got)
	}
}

func TestRecordCommandPreservesUsageANSIFixture(t *testing.T) {
	testRecordFixture(t, "harnez-usage.ansi")
}

func TestRecordCommandPreservesMCANSIFixtures(t *testing.T) {
	for _, name := range []string{"mc-julia256.ansi", "mc-mc46.ansi"} {
		t.Run(name, func(t *testing.T) {
			testRecordFixture(t, name)
		})
	}
}

func testRecordFixture(t *testing.T, name string) {
	t.Helper()
	fixture := filepath.Join("..", "..", "..", "docs", "data", name)
	want, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := RecordCommand(context.Background(), &out, time.Second, "cat", fixture); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), want) {
		t.Fatalf("recorded %s differs: got %d bytes, want %d", name, out.Len(), len(want))
	}
	if !bytes.Contains(out.Bytes(), []byte("\x1b[")) {
		t.Fatalf("recorded %s lost ANSI control sequences", name)
	}
}

func TestOpenRecorderPTYUsesRequestedGeometry(t *testing.T) {
	master, slave, err := openRecorderPTY(115, 37)
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()
	defer slave.Close()

	ws, err := unix.IoctlGetWinsize(int(slave.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Col != 115 || ws.Row != 37 {
		t.Fatalf("recorder PTY geometry = %dx%d, want 115x37", ws.Col, ws.Row)
	}
}
