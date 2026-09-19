// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ptytest

import (
	"slices"
	"testing"
)

func TestVTPositioningAndErase(t *testing.T) {
	v := NewVT(12, 3)
	v.Write([]byte("\x1b[1;1Hhello world\x1b[2;3Hab\x1b[1;6H\x1b[K")) //nolint:errcheck
	want := []string{"hello", "  ab", ""}
	if got := v.Screen(); !slices.Equal(got, want) {
		t.Fatalf("screen = %q, want %q", got, want)
	}
}

func TestVTAutoWrapToggle(t *testing.T) {
	v := NewVT(4, 2)
	v.Write([]byte("abcdef")) //nolint:errcheck
	if got := v.Screen(); !slices.Equal(got, []string{"abcd", "ef"}) {
		t.Fatalf("wrap on: %q", got)
	}
	v = NewVT(4, 2)
	v.Write([]byte("\x1b[?7labcdef")) //nolint:errcheck
	if got := v.Screen(); !slices.Equal(got, []string{"abcf", ""}) {
		t.Fatalf("wrap off: %q", got)
	}
}

func TestVTWideRuneAndSplitWrites(t *testing.T) {
	v := NewVT(6, 1)
	// Split a multi-byte rune and a CSI across writes.
	for _, chunk := range []string{"a\xe4", "\xb8\xad\x1b[", "1;1Hz"} {
		v.Write([]byte(chunk)) //nolint:errcheck
	}
	if got := v.Screen()[0]; got != "z 中" && got != "z中" {
		t.Fatalf("row = %q", got)
	}
}

func TestVTFrameSnapshotsAtSynchronizedEnd(t *testing.T) {
	v := NewVT(5, 1)
	v.Write([]byte("\x1b[?2026h\x1b[1;1Hone\x1b[?2026l\x1b[?2026h\x1b[1;1Htwo\x1b[K\x1b[?2026l")) //nolint:errcheck
	if len(v.Frames) != 2 || v.Frames[0][0] != "one" || v.Frames[1][0] != "two" {
		t.Fatalf("frames = %q", v.Frames)
	}
}

func TestVTResizeKeepsTopLeft(t *testing.T) {
	v := NewVT(6, 2)
	v.Write([]byte("abcdef\x1b[2;1Hxyz")) //nolint:errcheck
	v.Resize(3, 1)
	if got := v.Screen(); !slices.Equal(got, []string{"abc"}) {
		t.Fatalf("screen = %q", got)
	}
}
