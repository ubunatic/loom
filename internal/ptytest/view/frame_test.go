// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package view

import (
	"strings"
	"testing"

	"ubunatic.com/loom/vt"
)

func TestFrameRender(t *testing.T) {
	child := vt.NewVT(10, 10)
	child.Write([]byte("\x1b[31mHello\x1b[0m")) //nolint:errcheck

	frame := &Frame{
		Child: child,
		Chrome: []string{
			"Chrome 1",
			"Chrome 2",
		},
	}

	render := frame.Render()

	if !strings.Contains(render, "Chrome 1") || !strings.Contains(render, "Chrome 2") {
		t.Errorf("expected render to contain chrome, got: %q", render)
	}

	// VT SGR output normalizes 31m to 0;31m
	if !strings.Contains(render, "\x1b[0;31mHello") {
		t.Errorf("expected render to contain styled child output, got: %q", render)
	}
}

func TestFrameRenderAltScreen(t *testing.T) {
	child := vt.NewVT(10, 10)
	child.Write([]byte("\x1b[?1049h\x1b[32mAlt\x1b[0m")) //nolint:errcheck

	frame := &Frame{
		Child: child,
		Chrome: []string{
			"Chrome 1",
			"Chrome 2",
		},
	}

	render := frame.Render()

	if strings.Contains(render, "Chrome 1") || strings.Contains(render, "Chrome 2") {
		t.Errorf("expected render to hide chrome in alt screen, got: %q", render)
	}

	if !strings.Contains(render, "\x1b[?1049h") {
		t.Errorf("expected render to switch to alt screen, got: %q", render)
	}

	if !strings.Contains(render, "\x1b[0;32mAlt") {
		t.Errorf("expected render to contain styled alt screen output, got: %q", render)
	}
}

func TestFrameRenderTrimsTrailingBlanks(t *testing.T) {
	child := vt.NewVT(10, 10)
	child.Write([]byte("\x1b[31mFoo\x1b[0m      ")) //nolint:errcheck

	frame := &Frame{
		Child: child,
	}

	render := frame.Render()

	if strings.Contains(render, "      ") {
		t.Errorf("expected render to trim trailing blanks, got: %q", render)
	}
}

// A wide character takes two terminal columns; its right-half cell must not
// be emitted as a space, or every later character on the row shifts right.
func TestFrameRenderWideCharacter(t *testing.T) {
	child := vt.NewVT(10, 3)
	child.Write([]byte("界x👍y")) //nolint:errcheck

	render := (&Frame{Child: child}).Render()

	if !strings.Contains(render, "界x👍y") {
		t.Errorf("expected wide characters without padding, got: %q", render)
	}
}
