// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later
//go:build !windows

package session

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"ubunatic.com/loom/vt"
)

// TestDebugRedrawSignalsForegroundGroupWithoutResizing pins one repair to one
// SIGWINCH. Each PTY size change raises its own SIGWINCH, so a size toggle lets a
// child's trap run twice (REPAIRED:1 then :2) under load.
func TestDebugRedrawSignalsForegroundGroupWithoutResizing(t *testing.T) {
	master, slave, err := openSessionPTY(80, 24)
	if err != nil {
		t.Skipf("PTY unavailable: %v", err)
	}
	t.Cleanup(func() { master.Close() })
	cmd := exec.Command("sleep", "30")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = slave, slave, slave
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := cmd.Start(); err != nil {
		slave.Close()
		t.Fatalf("start child: %v", err)
	}
	slave.Close()
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	tab := &sessionTab{master: master, child: cmd, vt: vt.NewVT(80, 24)}

	resizes := 0
	origResize, origSignal := debugResizePTY, debugSignalGroup
	t.Cleanup(func() { debugResizePTY, debugSignalGroup = origResize, origSignal })
	debugResizePTY = func(f *os.File, cols, rows int) {
		resizes++
		setPTYSize(f, cols, rows)
	}
	var signaled []int
	debugSignalGroup = func(pgid int) error {
		signaled = append(signaled, pgid)
		return unix.Kill(-pgid, unix.SIGWINCH)
	}

	redrawTabChild(tab)

	if resizes != 0 {
		t.Fatalf("repair resized the PTY %d times; each resize raises its own SIGWINCH", resizes)
	}
	if len(signaled) != 1 || signaled[0] != cmd.Process.Pid {
		t.Fatalf("repair signaled groups %v, want one signal to child group %d", signaled, cmd.Process.Pid)
	}
}

func TestDebugRingWrapAndChunkReconstruction(t *testing.T) {
	ring := newDebugRing()
	ring.Write([]byte("first"))
	ring.Write([]byte(" second"))
	var rebuilt []byte
	for _, chunk := range ring.Chunks() {
		rebuilt = append(rebuilt, chunk.Data...)
	}
	if !bytes.Equal(rebuilt, []byte("first second")) {
		t.Fatalf("chunks reconstructed %q", rebuilt)
	}
	ring.Write(bytes.Repeat([]byte{'x'}, debugRingSize+3))
	got := ring.Bytes()
	if len(got) != debugRingSize || !bytes.Equal(got, bytes.Repeat([]byte{'x'}, debugRingSize)) {
		t.Fatal("oversized ring write did not retain newest capacity bytes")
	}
	if len(ring.Chunks()) != 1 || ring.Chunks()[0].Offset != int64(12+3) {
		t.Fatalf("chunks after wrap: %#v", ring.Chunks())
	}
}

func TestDebugBundleRotationAndContents(t *testing.T) {
	dir := t.TempDir()
	ring := newDebugRing()
	ring.Write([]byte("raw-chunk"))
	base := time.Unix(300, 0)
	for i := 0; i < debugBundleLimit+1; i++ {
		if err := writeDebugBundle(dir, 2, string(debugLeak), ring, "before", "after", base.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != debugBundleLimit*3 {
		t.Fatalf("files = %d, want %d", len(entries), debugBundleLimit*3)
	}
	var logs []string
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".log" {
			logs = append(logs, entry.Name())
		}
	}
	if len(logs) != debugBundleLimit || strings.Contains(logs[0], "000500") {
		t.Fatalf("unexpected retained logs: %v", logs)
	}
	logPath := filepath.Join(dir, logs[0])
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(logBytes, []byte("raw-chunk")) || !bytes.Contains(logBytes, []byte(string(debugLeak))) {
		t.Fatalf("log missing diagnostics: %s", logBytes)
	}
}

func TestDebugBundleDirectoryErrors(t *testing.T) {
	file := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(file, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeDebugBundle(file, 1, string(debugLeak), newDebugRing(), "", "", time.Now()); err == nil {
		t.Fatal("expected directory creation error")
	}
}

func TestDebugStatusBarIndicator(t *testing.T) {
	tab := &sessionTab{title: "bash", vt: vt.NewVT(80, 23), debugAlert: true}
	bar, _, _ := sessionBarContent([]*sessionTab{tab}, 0, 99, false)
	if !strings.Contains(bar, "[!]") {
		t.Fatalf("warning indicator missing from bar: %q", bar)
	}
}

func TestDebugDetectorControls(t *testing.T) {
	base := time.Unix(100, 0)
	tests := []struct {
		name  string
		parts [][]byte
		want  debugAnomaly
	}{
		{"split-csi", [][]byte{[]byte("\x1b[38;"), []byte("2;1;2;3mhello")}, ""},
		{"split-osc", [][]byte{[]byte("\x1b]0;title"), []byte("\ahello")}, ""},
		{"split-utf8", [][]byte{{0xe2, 0x82}, {0xac}}, ""},
		{"cancel", [][]byte{[]byte("\x1b[123"), []byte{0x18}}, ""},
		{"valid-ansi", [][]byte{[]byte("\x1b[31mred\x1b[0m")}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &debugDetector{}
			for i, part := range tt.parts {
				if got := d.Observe(part, base.Add(time.Duration(i)*time.Millisecond)); got != tt.want {
					t.Fatalf("Observe() = %q, want %q", got, tt.want)
				}
			}
			if len(d.pending) != 0 {
				t.Fatalf("pending bytes remain: %q", d.pending)
			}
		})
	}
}

func TestDebugDetectorStallAndOverlong(t *testing.T) {
	base := time.Unix(200, 0)
	d := &debugDetector{}
	if got := d.Observe([]byte("\x1b[38;2;"), base); got != "" {
		t.Fatalf("initial anomaly = %q", got)
	}
	if got := d.Stalled(base.Add(debugStallTimeout)); got != "" {
		t.Fatalf("stalled at boundary: %q", got)
	}
	if got := d.Stalled(base.Add(debugStallTimeout + time.Nanosecond)); got != debugIncomplete {
		t.Fatalf("stalled after boundary = %q", got)
	}
	d = &debugDetector{}
	if got := d.Observe([]byte("\x1b]"+strings.Repeat("x", debugControlLimit+1)), base); got != debugOverlong {
		t.Fatalf("overlong = %q", got)
	}
}

func TestDebugSuspiciousText(t *testing.T) {
	for _, text := range []string{"38;2;12;34;56m", "1;2;3H", "prefix 38;2;1;2;3;"} {
		if !suspiciousDebugText(text) {
			t.Errorf("expected suspicious fragment in %q", text)
		}
	}
	for _, text := range []string{"normal terminal text", "38 is a number; 2 is another", "temperature 38;2 degrees"} {
		if suspiciousDebugText(text) {
			t.Errorf("normal text flagged: %q", text)
		}
	}
}
