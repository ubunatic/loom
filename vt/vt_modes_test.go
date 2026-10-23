// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package vt_test

import (
	"testing"

	"ubunatic.com/loom/vt"
)

func TestVTModeTracking(t *testing.T) {
	v := vt.NewVT(80, 24)

	// Mouse 1000 + 1006 SGR
	v.Write([]byte("\x1b[?1000h\x1b[?1006h"))
	if v.MouseMode != 1000 || !v.MouseSGR {
		t.Fatalf("expected MouseMode=1000 and MouseSGR=true, got %d, %v", v.MouseMode, v.MouseSGR)
	}

	// Focus 1004
	v.Write([]byte("\x1b[?1004h"))
	if !v.FocusEvents {
		t.Fatalf("expected FocusEvents=true")
	}

	// Bracketed paste 2004
	v.Write([]byte("\x1b[?2004h"))
	if !v.BracketedPaste {
		t.Fatalf("expected BracketedPaste=true")
	}

	// Application cursor keys ?1
	v.Write([]byte("\x1b[?1h"))
	if !v.CursorKeysApp {
		t.Fatalf("expected CursorKeysApp=true")
	}

	// Keypad application \x1b=
	v.Write([]byte("\x1b="))
	if !v.KeypadApp {
		t.Fatalf("expected KeypadApp=true")
	}

	// Kitty keyboard push \x1b[>1u
	v.Write([]byte("\x1b[>1u"))
	if !v.KittyKeyboard {
		t.Fatalf("expected KittyKeyboard=true")
	}

	// Reset via disable sequences
	v.Write([]byte("\x1b[?1000l\x1b[?1006l\x1b[?1004l\x1b[?2004l\x1b[?1l\x1b>\x1b[<u"))
	if v.MouseMode != 0 {
		t.Errorf("expected MouseMode=0, got %d", v.MouseMode)
	}
	if v.MouseSGR {
		t.Errorf("expected MouseSGR=false")
	}
	if v.FocusEvents {
		t.Errorf("expected FocusEvents=false")
	}
	if v.BracketedPaste {
		t.Errorf("expected BracketedPaste=false")
	}
	if v.CursorKeysApp {
		t.Errorf("expected CursorKeysApp=false")
	}
	if v.KeypadApp {
		t.Errorf("expected KeypadApp=false")
	}
	if v.KittyKeyboard {
		t.Errorf("expected KittyKeyboard=false")
	}
}
