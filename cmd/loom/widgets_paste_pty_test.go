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
	// input without arming the app-level quit confirmation.
	gallery.SendRaw([]byte("\x1b[C\x03"))
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		joined := strings.Join(rows, "\n")
		return !strings.Contains(joined, "/topic Draft:") && strings.Contains(joined, "F2: placeholder")
	})
	gallery.SendRaw([]byte{0x03})
	gallery.WaitForUpdate("Press Ctrl+C again to quit", 3*time.Second)

	// Typing cancels confirmation and remains usable. A Ctrl+C that clears this
	// new value must not count as the first press of the quit sequence.
	gallery.Send("x")
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		joined := strings.Join(rows, "\n")
		return strings.Contains(joined, "x") && !strings.Contains(joined, "Press Ctrl+C again to quit")
	})
	gallery.SendRaw([]byte{0x03})
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		joined := strings.Join(rows, "\n")
		return !strings.Contains(joined, "x") && !strings.Contains(joined, "Press Ctrl+C again to quit") && strings.Contains(joined, "F2: placeholder")
	})
	// Keep a control byte inside the bracketed payload: it must be treated as
	// pasted text rather than as a terminal clear command.
	gallery.SendRaw([]byte("\x1b[200~after\nclear\x1b[2J\x1b[201~"))
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		joined := strings.Join(rows, "\n")
		return strings.Contains(joined, "after") && strings.Contains(joined, "clear") && strings.Contains(joined, "F2: placeholder") && strings.Contains(joined, "F10")
	})
	gallery.SendRaw([]byte{0x03})
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		joined := strings.Join(rows, "\n")
		return !strings.Contains(joined, "after") && !strings.Contains(joined, "clear") && strings.Contains(joined, "F2: placeholder")
	})
	gallery.SendRaw([]byte{0x03})
	gallery.WaitForUpdate("Press Ctrl+C again to quit", 3*time.Second)
	gallery.SendRaw([]byte{0x03})
	if err := gallery.Wait(3 * time.Second); err != nil {
		t.Fatalf("gallery exit status = %v, want success", err)
	}
}

func TestTextEntryGalleryCtrlUPTY(t *testing.T) {
	bin := buildLoomBinary(t)
	gallery := ptytest.Start(t, 100, 24, bin, "widgets", "--show", "TextEntry")
	gallery.WaitFor("/topic Draft:", 3*time.Second)
	gallery.SendRaw([]byte{0x01}) // select the initial draft
	gallery.SendRaw([]byte("\x1b[200~aa\nbbbb\x1b[201~"))
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		joined := strings.Join(rows, "\n")
		return strings.Contains(joined, "aa") && strings.Contains(joined, "bbbb")
	})
	endX, endY := gallery.VT().Cursor()
	if endX != 6 {
		t.Fatalf("multiline end cursor x = %d, want prompt plus bbbb (6); screen: %q", endX, gallery.Text())
	}
	gallery.SendRaw([]byte("\x1b[H")) // Home: start of the current hard line
	gallery.WaitForUpdateState(3*time.Second, func([]string) bool {
		x, y := gallery.VT().Cursor()
		return x == 2 && y == endY
	})
	gallery.SendRaw([]byte{0x15})
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		x, y := gallery.VT().Cursor()
		joined := strings.Join(rows, "\n")
		return x == 4 && y == endY-1 && strings.Contains(joined, "aa") && strings.Contains(joined, "bbbb")
	})
	if !strings.Contains(gallery.Text(), "aa") || !strings.Contains(gallery.Text(), "bbbb") {
		t.Fatalf("hard-start navigation did not retain both lines: %q", gallery.Text())
	}

	gallery.SendRaw([]byte{0x01})
	gallery.SendRaw([]byte("\x1b[200~abcdef\nxyzz\x1b[201~"))
	gallery.WaitForUpdate("xyzz", 3*time.Second)
	gallery.SendRaw([]byte("\x1b[D\x1b[D"))
	gallery.WaitForUpdateState(3*time.Second, func([]string) bool {
		x, _ := gallery.VT().Cursor()
		return x == 4
	})
	insideX, insideY := gallery.VT().Cursor()
	if insideX != 4 {
		t.Fatalf("cursor before interior Ctrl+U = (%d,%d), want x=4: %q", insideX, insideY, gallery.Text())
	}
	gallery.SendRaw([]byte{0x15})
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		x, y := gallery.VT().Cursor()
		joined := strings.Join(rows, "\n")
		return x == 2 && y == insideY && strings.Contains(joined, "abcdef") && strings.Contains(joined, "zz") && !strings.Contains(joined, "xyzz")
	})
	// Place the caret at the second soft boundary in a known-width wrapped line.
	gallery.SendRaw([]byte{0x01})
	wrapped := strings.Repeat("x", 220)
	gallery.SendRaw([]byte("\x1b[200~" + wrapped + "\x1b[201~"))
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		joined := strings.Join(rows, "\n")
		return strings.Contains(joined, strings.Repeat("x", 97)) && strings.Contains(joined, strings.Repeat("x", 26))
	})
	gallery.SendRaw([]byte(strings.Repeat("\x1b[D", 26)))
	gallery.WaitForUpdateState(3*time.Second, func([]string) bool {
		x, _ := gallery.VT().Cursor()
		return x == 2
	})
	gallery.SendRaw([]byte{0x15})
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		x, _ := gallery.VT().Cursor()
		joined := strings.Join(rows, "\n")
		return x == 2 && strings.Contains(joined, strings.Repeat("x", 97)) && strings.Contains(joined, strings.Repeat("x", 26)) && !strings.Contains(joined, wrapped)
	})
	gallery.SendRaw([]byte{0x15})
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		x, _ := gallery.VT().Cursor()
		joined := strings.Join(rows, "\n")
		return x == 2 && strings.Contains(joined, strings.Repeat("x", 26)) && !strings.Contains(joined, strings.Repeat("x", 97))
	})
	gallery.Send("x")
	gallery.WaitForUpdate("x", 3*time.Second)
	gallery.SendRaw([]byte{0x03})
	gallery.WaitForUpdateState(3*time.Second, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "x")
	})
	gallery.SendRaw([]byte{0x03})
	gallery.WaitForUpdate("Press Ctrl+C again to quit", 3*time.Second)
	gallery.SendRaw([]byte{0x03})
	if err := gallery.Wait(3 * time.Second); err != nil {
		t.Fatalf("exit: %v", err)
	}
}
