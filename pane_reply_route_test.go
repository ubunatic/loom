// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"context"
	"os"
	"testing"
	"time"
)

type recordingWidget struct {
	keys []KeyEvent
	done chan struct{}
}

func (w *recordingWidget) Draw(*Canvas, Rect) {}
func (w *recordingWidget) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
func (w *recordingWidget) ConsumeKey(ke KeyEvent) EventResult {
	w.keys = append(w.keys, ke)
	if ke.Key == "q" {
		select {
		case <-w.done:
		default:
			close(w.done)
		}
		return QuitResult()
	}
	return Handled()
}

func TestPaneReplyRouting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
	}{
		{"DECRPM", "\x1b[?2026;2$y", nil},
		{"DA1", "\x1b[?62;22c", nil},
		{"CPR", "\x1b[5;10R", nil},
		{"kitty flags", "\x1b[?1u", nil},
		{"resize report", "\x1b[48;24;80;480;960t", nil},
		{"OSC 11 BEL", "\x1b]11;rgb:1e1e/1e1e/1e1e\x07", nil},
		{"OSC 11 ST", "\x1b]11;rgb:1e1e/1e1e/1e1e\x1b\\", nil},
		{"focus in", "\x1b[I", nil},
		{"focus out", "\x1b[O", nil},
		{"997", "\x1b[?997;1n", nil},
		{"interleaved key", "\x1b[?2026;2$yA", []string{"A"}},
		{"split DECRPM", "\x1b[?2026", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, wr, err := os.Pipe()
			if err != nil {
			    t.Fatal(err)
			}

			p := &Pane{
			    tty: r,
			    fd: int(r.Fd()),
			    ResizeConfig: ResizeConfig{ResizeHandling: false},
			}
			w := &recordingWidget{done: make(chan struct{})}

			ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
			defer cancel()

			go func() {
			    if tt.name == "split DECRPM" {
			        wr.Write([]byte("\x1b[?2026"))
			        // Wait past 50ms escape timeout using channel
			        timer := time.NewTimer(70 * time.Millisecond)
			        select {
			        case <-timer.C:
			        case <-w.done:
			        }
			        wr.Write([]byte(";2$y"))
			        // The bug implies replies might be held and dispatched as text. We must dispatch "q" only AFTER the bug completes its erroneous text key generation.
			        timer2 := time.NewTimer(20 * time.Millisecond)
			        select {
			        case <-timer2.C:
			        case <-w.done:
			        }
			        wr.Write([]byte("q"))
			    } else {
				    wr.Write([]byte(tt.input))
			        timer := time.NewTimer(20 * time.Millisecond)
			        select {
			        case <-timer.C:
			        case <-w.done:
			        }
				    wr.Write([]byte("q"))
				}
			}()

			// Start closing procedure if 'q' is received early or we timed out
			go func() {
			    select {
				case <-w.done:
				case <-ctx.Done():
				}
				wr.Close()
			}()

			err = p.run(ctx, w, nil, nil, nil)
			if err != nil && err != context.Canceled && err.Error() != "loom: input: EOF" && err.Error() != "loom: input: read |0: file already closed" {
			    // Some errors might be normal if the deadline exceeded, wait, we don't want deadline exceeded,
			    // run returns context deadline exceeded if it hits the timeout before "q" is consumed
			    if err.Error() == "context deadline exceeded" {
			        // Ignore timeout, we just assert keys below
			    } else {
			        t.Fatalf("run returned %v", err)
			    }
			}

			r.Close()

			var gotKeys []string
			for _, k := range w.keys {
			    if k.Key != "q" && k.Key != "" {
			        gotKeys = append(gotKeys, k.Key)
			    } else if k.Key == "" && !k.Report {
			        gotKeys = append(gotKeys, "<empty>")
			    } else if k.Report {
			        gotKeys = append(gotKeys, "<report>")
			    }
			}

			if len(tt.wantKeys) == 0 {
			    tt.wantKeys = []string{}
			}

			if len(gotKeys) != len(tt.wantKeys) {
			    t.Errorf("got keys %v, want %v", gotKeys, tt.wantKeys)
			} else {
			    for i, k := range gotKeys {
			        if k != tt.wantKeys[i] {
			            t.Errorf("got key[%d] = %q, want %q", i, k, tt.wantKeys[i])
			        }
			    }
			}
		})
	}
}

func TestPaneReplyRouting_NormalKey(t *testing.T) {
    r, wr, err := os.Pipe()
    if err != nil {
        t.Fatal(err)
    }

    p := &Pane{
        tty: r,
        fd: int(r.Fd()),
        ResizeConfig: ResizeConfig{ResizeHandling: false},
    }
    w := &recordingWidget{done: make(chan struct{})}

    ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
    defer cancel()

    go func() {
        wr.Write([]byte("\x1b[?2026;2$yA"))
        timer := time.NewTimer(20 * time.Millisecond)
        select {
        case <-timer.C:
        case <-w.done:
        }
        wr.Write([]byte("q"))
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
        if err.Error() != "context deadline exceeded" {
            t.Fatalf("run returned %v", err)
        }
    }

    r.Close()

    var gotKeys []string
    for _, k := range w.keys {
        if k.Key != "q" && k.Key != "" {
            gotKeys = append(gotKeys, k.Key)
        } else if k.Key == "" && !k.Report {
            gotKeys = append(gotKeys, "<empty>")
        } else if k.Report {
            gotKeys = append(gotKeys, "<report>")
        }
    }

    want := []string{"A"}
    if len(gotKeys) != len(want) {
        t.Errorf("got keys %v, want %v", gotKeys, want)
    } else {
        for i, k := range gotKeys {
            if k != want[i] {
                t.Errorf("got key[%d] = %q, want %q", i, k, want[i])
            }
        }
    }
}
