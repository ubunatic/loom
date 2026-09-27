// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package usage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/internal/ptytest"
)

func TestUsageWidgetLayoutAndResize(t *testing.T) {
	collected := make(chan struct{}, 4)
	w, err := NewWidgetFromOptions(Options{
		CollectInterval: time.Hour,
		RedrawInterval:  time.Hour,
		Source: DataSourceFunc(func(_ context.Context, at time.Time) (Snapshot, error) {
			select {
			case collected <- struct{}{}:
			default:
			}
			return sampleSnapshot(at), nil
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.(*usageWidget).Close()
	widget := w.(*usageWidget)
	select {
	case <-collected:
	case <-time.After(time.Second):
		t.Fatal("initial asynchronous collection did not run")
	}
	waitSnapshot(t, widget)

	for _, size := range []struct{ width, height int }{{100, 18}, {48, 18}, {32, 20}, {76, 20}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			rows := loom.Render(widget, size.width, size.height)
			if len(rows) != size.height {
				t.Fatalf("rendered %d rows, want %d", len(rows), size.height)
			}
			text := strings.Join(rows, "\n")
			for _, want := range []string{"All Usage", "Load", "Claude", "CPU", "RAM"} {
				if !strings.Contains(text, want) {
					t.Errorf("render at %dx%d missing %q:\n%s", size.width, size.height, want, text)
				}
			}
			canvas := loom.NewCanvas(size.width, size.height)
			widget.Draw(canvas, loom.Rect{W: size.width, H: size.height})
			for y := 0; y < size.height; y++ {
				if got := loom.StringWidth(canvas.Row(y)); got != size.width {
					t.Errorf("row %d width = %d, want %d", y, got, size.width)
				}
			}
		})
	}
}

func TestUsageWidgetDrawAndTickStaySilent(t *testing.T) {
	w, err := NewWidgetFromOptions(Options{
		CollectInterval: time.Hour,
		RedrawInterval:  time.Hour,
		Source: DataSourceFunc(func(_ context.Context, at time.Time) (Snapshot, error) {
			return sampleSnapshot(at), nil
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.(*usageWidget).Close()
	widget := w.(*usageWidget)
	waitSnapshot(t, widget)
	widget.SetInvalidate(func() {})
	canvas := loom.NewCanvas(80, 20)
	output := captureOutput(t, func() {
		widget.Draw(canvas, loom.Rect{W: 80, H: 20})
		widget.Tick(time.Now())
	})
	if output != "" {
		t.Fatalf("hosted Draw/Tick wrote to process output: %q", output)
	}
}

func captureOutput(t *testing.T, run func()) string {
	t.Helper()
	oldOut, oldErr := os.Stdout, os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = writer, writer
	done := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(reader)
		done <- string(data)
	}()
	run()
	os.Stdout, os.Stderr = oldOut, oldErr
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output := <-done
	_ = reader.Close()
	return output
}

func waitSnapshot(t *testing.T, widget *usageWidget) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if !widget.snapshot().CapturedAt.IsZero() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("asynchronous source did not publish a snapshot")
}

func TestUsageWidgetInvalidatesAfterAsyncCollection(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	w, err := NewWidgetFromOptions(Options{
		CollectInterval: time.Hour,
		RedrawInterval:  time.Hour,
		Source: DataSourceFunc(func(ctx context.Context, at time.Time) (Snapshot, error) {
			if calls.Add(1) == 1 {
				close(started)
			}
			select {
			case <-ctx.Done():
				return Snapshot{}, ctx.Err()
			case <-release:
				return sampleSnapshot(at), nil
			}
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	widget := w.(*usageWidget)
	defer widget.Close()
	<-started
	invalidated := make(chan struct{}, 1)
	widget.SetInvalidate(func() {
		select {
		case invalidated <- struct{}{}:
		default:
		}
	})
	close(release)
	select {
	case <-invalidated:
	case <-time.After(time.Second):
		t.Fatal("collection completion did not request a pane redraw")
	}
}

func TestUsagePTYColorCellHelper(t *testing.T) {
	if os.Getenv("LOOM_USAGE_PTY_HELPER") == "" {
		return
	}
	source := DataSourceFunc(func(ctx context.Context, at time.Time) (Snapshot, error) {
		timer := time.NewTimer(250 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return Snapshot{}, ctx.Err()
		case <-timer.C:
			return sampleSnapshot(at), nil
		}
	})
	widget, err := NewWidgetFromOptions(Options{Source: source, CollectInterval: time.Hour, RedrawInterval: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	defer widget.(*usageWidget).Close()
	pane, err := loom.New(12)
	if err != nil {
		t.Fatal(err)
	}
	defer pane.Close()
	pane.Resizeable = true
	if err := pane.Run(widget); err != nil {
		t.Fatal(err)
	}
}

func TestUsagePTYColorCells(t *testing.T) {
	old, hadOld := os.LookupEnv("LOOM_USAGE_PTY_HELPER")
	if err := os.Setenv("LOOM_USAGE_PTY_HELPER", "1"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if hadOld {
			_ = os.Setenv("LOOM_USAGE_PTY_HELPER", old)
		} else {
			_ = os.Unsetenv("LOOM_USAGE_PTY_HELPER")
		}
	}()
	session := ptytest.Start(t, 100, 24, os.Args[0], "-test.run=^TestUsagePTYColorCellHelper$")
	session.WaitFor("Loom Usage", 3*time.Second)
	session.WaitFor("Claude", 3*time.Second)
	checkUsageColor(t, session)
	if err := os.WriteFile("/tmp/loom-103-usage.ansi", session.Raw(), 0600); err != nil {
		t.Fatalf("write ANSI snapshot: %v", err)
	}
	session.Send("q")
	if err := session.Wait(3 * time.Second); err != nil {
		t.Fatalf("usage pane did not exit after q: %v", err)
	}
}

func checkUsageColor(t *testing.T, session *ptytest.Session) {
	t.Helper()
	rows := session.Screen()
	cells := session.Cells()
	for y, row := range rows {
		if !strings.Contains(row, "All Usage") {
			continue
		}
		for x := 0; x+8 <= len(cells[y]); x++ {
			if cells[y][x].Rune != 'A' || cells[y][x+1].Rune != 'l' || cells[y][x+2].Rune != 'l' {
				continue
			}
			r, g, b, ok := cells[y][x].Style.FG.RGB()
			if !ok || r != 185 || g != 125 || b != 255 {
				t.Fatalf("All Usage title color = (%d,%d,%d), want (185,125,255)", r, g, b)
			}
			return
		}
	}
	t.Fatalf("All Usage title cell not found in PTY screen:\n%s", strings.Join(rows, "\n"))
}

func TestUsageViewFlagAndInteractiveSwitch(t *testing.T) {
	w, err := NewWidget([]string{"--view=plain", "--collect=1h", "--redraw=1h"})
	if err != nil {
		t.Fatal(err)
	}
	widget := w.(*usageWidget)
	defer widget.Close()
	waitSnapshot(t, widget)
	if widget.view != viewPlain {
		t.Fatalf("initial view = %q, want plain", widget.view)
	}
	plain := strings.Join(loom.Render(widget, 100, 18), "\n")
	for _, want := range []string{"All Usage", "Load", "Claude", "CPU", "plain view"} {
		if !strings.Contains(plain, want) {
			t.Errorf("plain view missing %q:\n%s", want, plain)
		}
	}
	if !strings.Contains(plain, "38;2;185;125;255m") {
		t.Fatal("plain All Usage panel lost its colored title")
	}
	if widget.HandleKey(loom.KeyEvent{Text: "v"}) {
		t.Fatal("view switch unexpectedly quit the widget")
	}
	if widget.view != viewLoom {
		t.Fatalf("view after v = %q, want Loom", widget.view)
	}
	loomView := strings.Join(loom.Render(widget, 100, 18), "\n")
	if !strings.Contains(loomView, "Claude") || strings.Contains(loomView, "plain view") {
		t.Fatalf("Loom view did not preserve shared usage data:\n%s", loomView)
	}
	widget.HandleKey(loom.KeyEvent{Text: "v"})
	if widget.view != viewPlain {
		t.Fatalf("view after second v = %q, want plain", widget.view)
	}
	if _, err := NewWidget([]string{"--view=unknown"}); err == nil {
		t.Fatal("unknown view was accepted")
	}
}

func TestPlainUsageViewFitsNarrowAndResizedRects(t *testing.T) {
	w, err := NewWidget([]string{"--view=plain", "--collect=1h", "--redraw=1h"})
	if err != nil {
		t.Fatal(err)
	}
	defer w.(*usageWidget).Close()
	widget := w.(*usageWidget)
	waitSnapshot(t, widget)
	for _, size := range []struct{ width, height int }{{100, 18}, {54, 20}, {30, 24}, {72, 16}} {
		rows := loom.Render(widget, size.width, size.height)
		if len(rows) != size.height {
			t.Fatalf("%dx%d rendered %d rows", size.width, size.height, len(rows))
		}
		for y, row := range rows {
			if got := loom.StringWidth(row); got > size.width {
				t.Errorf("%dx%d row %d width = %d", size.width, size.height, y, got)
			}
		}
		text := strings.Join(rows, "\n")
		if !strings.Contains(text, "All Usage") || !strings.Contains(text, "Load") {
			t.Errorf("%dx%d lost a panel:\n%s", size.width, size.height, text)
		}
	}
}
