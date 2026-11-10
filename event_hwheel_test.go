// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestDecodeMouse_SGRHorizontalWheel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDesc string
	}{
		{"64 (up)", "\x1b[<64;1;1M", "wheel/scroll action"},
		{"64 + shift", "\x1b[<68;1;1M", "wheel/scroll action"},
		{"64 + alt", "\x1b[<72;1;1M", "wheel/scroll action"},
		{"64 + ctrl", "\x1b[<80;1;1M", "wheel/scroll action"},
		{"65 (down)", "\x1b[<65;1;1M", "wheel/scroll action"},
		{"65 + shift", "\x1b[<69;1;1M", "wheel/scroll action"},
		{"65 + alt", "\x1b[<73;1;1M", "wheel/scroll action"},
		{"65 + ctrl", "\x1b[<81;1;1M", "wheel/scroll action"},
		{"66 (left)", "\x1b[<66;1;1M", "wheel/scroll action"},
		{"66 + shift", "\x1b[<70;1;1M", "wheel/scroll action"},
		{"66 + alt", "\x1b[<74;1;1M", "wheel/scroll action"},
		{"66 + ctrl", "\x1b[<82;1;1M", "wheel/scroll action"},
		{"67 (right)", "\x1b[<67;1;1M", "wheel/scroll action"},
		{"67 + shift", "\x1b[<71;1;1M", "wheel/scroll action"},
		{"67 + alt", "\x1b[<75;1;1M", "wheel/scroll action"},
		{"67 + ctrl", "\x1b[<83;1;1M", "wheel/scroll action"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, ok := DecodeMouse([]byte(tt.input))
			if !ok {
				t.Fatalf("DecodeMouse(%q) = false", tt.input)
			}
			if ev.Action == MousePress {
				t.Errorf("DecodeMouse(%q) = %v; want not a MousePress (should be %s)", tt.input, ev, tt.wantDesc)
			}
			if ev.Action != MouseScrollUp && ev.Action != MouseScrollDown {
			    t.Errorf("DecodeMouse(%q) = %v; want a wheel/scroll action", tt.input, ev)
			}
		})
	}
}

type hwheelRecordingWidget struct {
	presses []MouseEvent
	done chan struct{}
}

func (w *hwheelRecordingWidget) Draw(*Canvas, Rect) {}
func (w *hwheelRecordingWidget) ConsumeKey(KeyEvent) EventResult { return Ignored() }
func (w *hwheelRecordingWidget) ConsumeMouse(me MouseEvent) EventResult {
	if me.Action == MousePress {
		w.presses = append(w.presses, me)
	}
	// Exit when receiving release for 66 so test can finish
	if me.Action == MouseRelease && me.Button == 2 {
		select {
		case <-w.done:
		default:
			close(w.done)
		}
		return QuitResult()
	}
	return Handled()
}

func TestPaneHorizontalWheelRoute(t *testing.T) {
	r, wr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	p := &Pane{
		tty: r,
		fd: int(r.Fd()),
		ResizeConfig: ResizeConfig{ResizeHandling: false},
	}
	p.EnableMouseClicks()

	w := &hwheelRecordingWidget{done: make(chan struct{})}

	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()

	go func() {
		wr.Write([]byte("\x1b[<66;45;15M"))
		wr.Write([]byte("\x1b[<66;45;15m")) // release
	}()

	go func() {
	    select {
	    case <-w.done:
	    case <-ctx.Done():
	    }
	    wr.Close()
	}()

	err = p.run(ctx, w, nil, nil, nil)
	if err != nil && err != context.Canceled && err.Error() != "loom: input: EOF" && err.Error() != "loom: input: read |0: file already closed" {
		t.Fatalf("run returned %v", err)
	}
	r.Close()

	if len(w.presses) > 0 {
		t.Errorf("horizontal wheel reached widget as press: %v", w.presses)
	}
}
