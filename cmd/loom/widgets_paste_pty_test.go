// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

func TestTextEntryGalleryBracketedPastePTY(t *testing.T) {
	bin := buildLoomBinary(t)
	for _, tc := range []struct {
		name   string
		pieces []string
	}{
		{"complete", []string{"\x1b[200~first line\nsecond line\x1b[201~"}},
		{"fragmented", []string{"\x1b[20", "0~first line\n", "second line\x1b[20", "1~"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gallery := ptytest.Start(t, 100, 24, bin, "widgets", "--show", "TextEntry")
			gallery.WaitFor("/topic Draft:", 3*time.Second)
			if !bytes.Contains(gallery.Raw(), []byte("\x1b[?2004h")) {
				t.Fatal("gallery did not enable bracketed paste")
			}
			for _, piece := range tc.pieces {
				gallery.SendRaw([]byte(piece))
			}
			gallery.WaitFor("second line", 3*time.Second)
			screen := strings.Join(gallery.Screen(), "\n")
			// The gallery submit callback clears the entry. Keeping the original
			// text and both lines therefore also proves no submit was delivered.
			for _, want := range []string{"/topic Draft:", "first line", "second line"} {
				if !strings.Contains(screen, want) {
					t.Fatalf("paste lost %q or submitted the entry: %q", want, screen)
				}
			}
		})
	}
}

func TestTextEntryGalleryCtrlCAndParentRoutingPTY(t *testing.T) {
	bin := buildLoomBinary(t)
	gallery := ptytest.Start(t, 100, 24, bin, "widgets", "--show", "TextEntry")
	gallery.WaitFor("/topic Draft:", 3*time.Second)

	// Ctrl+A selects the draft. Ctrl+C must preserve it rather than clear it.
	gallery.SendRaw([]byte{0x01, 0x03})
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		return strings.Contains(strings.Join(rows, "\n"), "/topic Draft:")
	})
	if got := strings.Join(gallery.Screen(), "\n"); !strings.Contains(got, "/topic Draft:") || !strings.Contains(got, "F2: placeholder") {
		t.Fatalf("selected Ctrl+C changed gallery display: %q", got)
	}
	// Right collapses the selection. The gallery parent then clears nonempty
	// input; an empty Ctrl+C remains unhandled, so a following paste proves the
	// actual top-level pane did not assume that it should exit.
	gallery.SendRaw([]byte("\x1b[C\x03"))
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		joined := strings.Join(rows, "\n")
		return !strings.Contains(joined, "/topic Draft:") && strings.Contains(joined, "F2: placeholder")
	})
	gallery.SendRaw([]byte{0x03})
	gallery.SendRaw([]byte("\x1b[200~after\nclear\x1b[2J\x1b[201~"))
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		joined := strings.Join(rows, "\n")
		return strings.Contains(joined, "after") && strings.Contains(joined, "clear") && strings.Contains(joined, "F2: placeholder")
	})
}
