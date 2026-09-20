// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom"
)

type testWidget struct {
	mu          sync.Mutex
	drawCount   int
	drawnTexts  []string
	text        string
	keyCount    int
	mouseCount  int
	lastKey     loom.KeyEvent
	lastMouse   loom.MouseEvent
	returnKey   bool
	returnMouse bool
}

func newTestWidget(text string) *testWidget {
	return &testWidget{text: text}
}

func (w *testWidget) Draw(c *loom.Canvas, rect loom.Rect) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.drawCount++
	w.drawnTexts = append(w.drawnTexts, w.text)
	if w.text != "" {
		for y := rect.Y; y < rect.Y+rect.H; y++ {
			for x, r := range w.text {
				if rect.X+x < rect.X+rect.W {
					c.Set(rect.X+x, y, loom.Cell{Text: string(r)})
				}
			}
		}
	}
}

func (w *testWidget) HandleKey(e loom.KeyEvent) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.keyCount++
	w.lastKey = e
	return w.returnKey
}

func (w *testWidget) HandleMouse(e loom.MouseEvent) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.mouseCount++
	w.lastMouse = e
	return w.returnMouse
}

func TestRenderTo(t *testing.T) {
	t.Run("valid render", func(t *testing.T) {
		var buf bytes.Buffer
		view := loom.NewView([]string{"line 1", "line 2"})
		err := loom.RenderTo(&buf, view, 20, 2)
		if err != nil {
			t.Fatalf("RenderTo failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "line 1") || !strings.Contains(out, "line 2") {
			t.Errorf("output missing expected lines: %q", out)
		}
	})

	t.Run("nil writer", func(t *testing.T) {
		view := loom.NewView([]string{"test"})
		err := loom.RenderTo(nil, view, 20, 2)
		if err == nil {
			t.Fatal("expected error with nil writer")
		}
	})

	t.Run("nil widget", func(t *testing.T) {
		var buf bytes.Buffer
		err := loom.RenderTo(&buf, nil, 20, 2)
		if err == nil {
			t.Fatal("expected error with nil widget")
		}
	})

	t.Run("invalid dimensions", func(t *testing.T) {
		var buf bytes.Buffer
		view := loom.NewView([]string{"test"})
		if err := loom.RenderTo(&buf, view, 0, 5); err == nil {
			t.Fatal("expected error with zero width")
		}
		if err := loom.RenderTo(&buf, view, 5, 0); err == nil {
			t.Fatal("expected error with zero height")
		}
		if err := loom.RenderTo(&buf, view, -1, 5); err == nil {
			t.Fatal("expected error with negative width")
		}
	})
}

func TestStartupConfigValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("nil pane", func(t *testing.T) {
		var p *loom.Pane
		err := p.RunStartup(ctx, loom.StartupConfig{
			Splash: loom.NewSplashController(loom.SplashConfig{}),
			Next:   newTestWidget("next"),
		})
		if err == nil || !strings.Contains(err.Error(), "nil pane") {
			t.Fatalf("expected nil pane error, got %v", err)
		}
	})

	t.Run("nil splash controller", func(t *testing.T) {
		skipWithoutTTY(t)
		p, err := loom.New(5)
		if err != nil {
			t.Fatalf("New pane failed: %v", err)
		}
		defer p.Close()

		err = p.RunStartup(ctx, loom.StartupConfig{
			Next: newTestWidget("next"),
		})
		if err == nil || !strings.Contains(err.Error(), "splash controller required") {
			t.Fatalf("expected nil splash controller error, got %v", err)
		}
	})

	t.Run("nil next destination widget", func(t *testing.T) {
		skipWithoutTTY(t)
		p, err := loom.New(5)
		if err != nil {
			t.Fatalf("New pane failed: %v", err)
		}
		defer p.Close()

		err = p.RunStartup(ctx, loom.StartupConfig{
			Splash: loom.NewSplashController(loom.SplashConfig{}),
		})
		if err == nil || !strings.Contains(err.Error(), "destination widget required") {
			t.Fatalf("expected nil destination error, got %v", err)
		}
	})

	t.Run("already cancelled context", func(t *testing.T) {
		skipWithoutTTY(t)
		p, err := loom.New(5)
		if err != nil {
			t.Fatalf("New pane failed: %v", err)
		}
		defer p.Close()

		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		err = p.RunStartup(cancelledCtx, loom.StartupConfig{
			Splash: loom.NewSplashController(loom.SplashConfig{}),
			Next:   newTestWidget("next"),
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})
}

func TestStartupLifecycleProgressionAndHandover(t *testing.T) {
	skipWithoutTTY(t)
	p, err := loom.New(5)
	if err != nil {
		t.Fatalf("New pane failed: %v", err)
	}
	defer p.Close()

	sc := loom.NewSplashController(loom.SplashConfig{
		Title: "Test Splash",
		Tasks: []loom.ProviderTask{
			{Name: "task1", Symbol: "1", Duration: 10 * time.Millisecond},
		},
		TickInterval: 10 * time.Millisecond,
		HoldDuration: 10 * time.Millisecond,
	})

	next := newTestWidget("next-content")
	transitionCalled := false
	transitionFunc := loom.TransitionFunc(func(ctx context.Context, from, to loom.Widget) error {
		transitionCalled = true
		if to != next {
			t.Errorf("transition target mismatch: got %v, want %v", to, next)
		}
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cadence := loom.Cadence{
		Collect: 10 * time.Millisecond,
		Redraw:  10 * time.Millisecond,
	}

	// We start RunStartup in a goroutine and cancel after next widget is drawn or short time
	errCh := make(chan error, 1)
	go func() {
		errCh <- p.RunStartup(ctx, loom.StartupConfig{
			Splash:     sc,
			Next:       next,
			Cadence:    cadence,
			Transition: transitionFunc,
		})
	}()

	// Wait for controller completion
	select {
	case <-sc.Done():
	case <-time.After(1 * time.Second):
		t.Fatal("splash controller did not complete in time")
	}

	// Wait for next widget to be rendered
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		next.mu.Lock()
		count := next.drawCount
		next.mu.Unlock()
		if count > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	next.mu.Lock()
	count := next.drawCount
	next.mu.Unlock()
	if count == 0 {
		t.Errorf("next widget drawCount = 0, expected > 0 after transition")
	}

	if !transitionCalled {
		t.Errorf("transition callback was not called")
	}

	// Clean shutdown via context cancel
	cancel()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("RunStartup unexpected error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("RunStartup did not exit after context cancel")
	}
}

func TestStartupTransitionError(t *testing.T) {
	skipWithoutTTY(t)
	p, err := loom.New(5)
	if err != nil {
		t.Fatalf("New pane failed: %v", err)
	}
	defer p.Close()

	sc := loom.NewSplashController(loom.SplashConfig{
		Tasks:        []loom.ProviderTask{},
		TickInterval: 5 * time.Millisecond,
		HoldDuration: 5 * time.Millisecond,
	})

	expectedErr := errors.New("transition failed")
	transitionFunc := loom.TransitionFunc(func(ctx context.Context, from, to loom.Widget) error {
		return expectedErr
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = p.RunStartup(ctx, loom.StartupConfig{
		Splash:     sc,
		Next:       newTestWidget("next"),
		Cadence:    loom.Cadence{Collect: 5 * time.Millisecond, Redraw: 5 * time.Millisecond},
		Transition: transitionFunc,
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestStartupDismissalImmediateHandover(t *testing.T) {
	skipWithoutTTY(t)
	p, err := loom.New(5)
	if err != nil {
		t.Fatalf("New pane failed: %v", err)
	}
	defer p.Close()

	sc := loom.NewSplashController(loom.SplashConfig{
		Tasks: []loom.ProviderTask{
			{Name: "slow", Duration: 1 * time.Second},
		},
		TickInterval: 10 * time.Millisecond,
	})

	next := newTestWidget("next-content")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- p.RunStartup(ctx, loom.StartupConfig{
			Splash:  sc,
			Next:    next,
			Cadence: loom.Cadence{Collect: 10 * time.Millisecond, Redraw: 10 * time.Millisecond},
		})
	}()

	// Dismiss via key or direct call
	time.Sleep(20 * time.Millisecond)
	sc.Dismiss()

	// Wait for next widget to be rendered
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		next.mu.Lock()
		count := next.drawCount
		next.mu.Unlock()
		if count > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	next.mu.Lock()
	count := next.drawCount
	next.mu.Unlock()
	if count == 0 {
		t.Errorf("next widget was not drawn after dismissal")
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("RunStartup returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("RunStartup did not exit after context cancel")
	}
}

func skipWithoutTTY(t *testing.T) {
	t.Helper()
	tty, err := os.Open("/dev/tty")
	if err != nil {
		t.Skipf("requires /dev/tty: %v", err)
	}
	_ = tty.Close()
}
