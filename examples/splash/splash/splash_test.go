// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package splash

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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

	err := runWatch(ctx, Options{Watch: true})
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

func TestNewWidgetOptions(t *testing.T) {
	w, err := NewWidget([]string{"--watch"})
	if err != nil {
		t.Fatalf("NewWidget(--watch) failed: %v", err)
	}
	if w == nil {
		t.Fatal("NewWidget returned nil")
	}

	_, err = NewWidget([]string{"extra"})
	if err == nil {
		t.Fatal("expected error on unexpected arguments")
	}
}

func TestSplashWidgetLifecycleAndTicks(t *testing.T) {
	widget, err := NewWidget([]string{"--watch"})
	if err != nil {
		t.Fatalf("NewWidget failed: %v", err)
	}
	app, ok := widget.(*splashApp)
	if !ok {
		t.Fatalf("expected *splashApp, got %T", widget)
	}
	defer app.Close()

	requester, ok := widget.(loom.PaneRequester)
	if !ok {
		t.Fatalf("expected loom.PaneRequester, got %T", widget)
	}
	req := requester.PaneRequest()
	if !req.Resizeable {
		t.Errorf("expected Resizeable=true, got %v", req.Resizeable)
	}

	ticker, ok := widget.(loom.Ticker)
	if !ok {
		t.Fatalf("expected loom.Ticker, got %T", widget)
	}
	if ticker.TickInterval() != 80*time.Millisecond {
		t.Errorf("expected 80ms interval, got %v", ticker.TickInterval())
	}

	// First tick starts controller
	ticker.Tick(time.Now())

	// Dismiss splash via key
	app.ConsumeKey(loom.KeyEvent{Key: "esc"})

	// Advance ticks to finish transition
	ticker.Tick(time.Now())
	canvas := loom.NewCanvas(60, 12)
	app.Draw(canvas, canvas.Bounds())
	ticker.Tick(time.Now())

	if !app.active {
		t.Error("expected splash to transition to destination widget after dismissal")
	}

	// Key in active destination: 'q' should quit
	quit := app.ConsumeKey(loom.KeyEvent{Key: "q"}).Quit
	if !quit {
		t.Error("expected 'q' in active destination to signal quit")
	}
}

func TestSplashHostedInTabsOnChildQuit(t *testing.T) {
	widget, err := NewWidget([]string{"--watch"})
	if err != nil {
		t.Fatalf("NewWidget failed: %v", err)
	}
	app := widget.(*splashApp)
	defer app.Close()

	var childQuitReported bool
	tabs := loom.NewTabs(loom.Tab{Title: "Splash", Widget: app})
	tabs.OnChildQuit = func(i int) bool {
		childQuitReported = true
		return false // stay alive in host
	}

	// Dismiss splash and transition to destination
	app.Tick(time.Now())
	app.ConsumeKey(loom.KeyEvent{Key: "esc"})
	app.Tick(time.Now())
	canvas := loom.NewCanvas(60, 12)
	tabs.Draw(canvas, canvas.Bounds())
	app.Tick(time.Now())

	// Send 'q' key through Tabs
	quit := tabs.ConsumeKey(loom.KeyEvent{Key: "q"}).Quit
	if quit {
		t.Error("expected Tabs.ConsumeKey to return false when OnChildQuit contains quit")
	}
	if !childQuitReported {
		t.Error("expected Tabs.OnChildQuit to be called when child destination quits")
	}
}
