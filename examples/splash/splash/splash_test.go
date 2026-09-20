// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package splash

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func stripANSI(s string) string {
	var b strings.Builder
	esc := false
	csi := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			esc = true
		case esc && r == '[':
			csi = true
			esc = false
		case csi && r >= 0x40 && r <= 0x7e:
			csi = false
		case esc:
			esc = false
		case !csi:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func TestSplashShowOnceOutput(t *testing.T) {
	var buf bytes.Buffer
	err := Execute(context.Background(), []string{"-w", "60", "-H", "12"}, &buf)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	out := stripANSI(buf.String())
	if !strings.Contains(out, "harnez usage") {
		t.Errorf("output missing title: %q", out)
	}
	if !strings.Contains(out, "fetching claude...") {
		t.Errorf("output missing step text: %q", out)
	}
	if !strings.Contains(out, "● mic  ✳ claude  ֍ codex  Λ agy") {
		t.Errorf("output missing status pills: %q", out)
	}
	if !strings.Contains(out, "Esc to skip") {
		t.Errorf("output missing footer: %q", out)
	}
}

func TestSplashRunWatchContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	err := runWatch(ctx, &buf)
	if err != nil && !errors.Is(err, context.Canceled) {
		if strings.Contains(err.Error(), "/dev/tty") || strings.Contains(err.Error(), "another pane") {
			t.Skipf("skipping test without TTY: %v", err)
		}
		t.Fatalf("runWatch failed: %v", err)
	}
}

func TestResolveTerminalDimensions(t *testing.T) {
	w, h := resolveTerminalDimensions(100, 40)
	if w != 100 || h != 40 {
		t.Fatalf("expected (100, 40), got (%d, %d)", w, h)
	}

	w, h = resolveTerminalDimensions(0, 0)
	if w <= 0 || h <= 0 {
		t.Fatalf("expected positive dimensions, got (%d, %d)", w, h)
	}
}

func TestInteractiveDestination(t *testing.T) {
	dest := newInteractiveDestination()
	if dest == nil {
		t.Fatal("expected non-nil destination widget")
	}
	canvas := loom.NewCanvas(60, 10)
	canvas.Clear()
	dest.Draw(canvas, canvas.Bounds())
	var b strings.Builder
	canvas.Flush(&b, 1)
	out := b.String()
	if !strings.Contains(out, "harnez usage") {
		t.Errorf("destination draw missing title: %q", out)
	}
	if !strings.Contains(out, "Initialized Providers") {
		t.Errorf("destination draw missing box header: %q", out)
	}
}
