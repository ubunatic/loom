// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package usage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom"
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
