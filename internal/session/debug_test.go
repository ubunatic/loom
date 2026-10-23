// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later
//go:build !windows

package session

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

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
