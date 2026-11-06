// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

// chatPasteWidget models the host forwarding contract used by taia.Chat:
// transcript and status share a canvas with an inset, height-capped TextEntry.
type chatPasteWidget struct{ input *TextEntry }

func (w *chatPasteWidget) Draw(c *Canvas, r Rect) {
	c.PaintSurface(r, Style{FG: ColorIndex(7), BG: ColorIndex(4)})
	c.Write(r.X+1, r.Y, "Chat transcript", Style{Bold: true})
	c.Write(r.X+1, r.Y+1, "user: previous question", Style{})
	c.Write(r.X+1, r.Y+2, "assistant: previous answer", Style{})
	c.Write(r.X+1, r.Y+3, strings.Repeat("-", r.W-2), Style{})
	c.Write(r.X+1, r.Y+7, "status: ready", Style{})
	w.input.Draw(c, Rect{X: r.X + 2, Y: r.Y + 4, W: r.W - 4, H: 3})
}

func (w *chatPasteWidget) ConsumePaste(e PasteEvent) EventResult { return w.input.ConsumePaste(e) }
func (w *chatPasteWidget) ConsumeMouse(e MouseEvent) EventResult { return Ignored() }
func (w *chatPasteWidget) ConsumeKey(e KeyEvent) EventResult {
	if e.Key == "ctrl-q" {
		return QuitResult()
	}
	return w.input.ConsumeKey(e)
}

func TestTextEntryChatPastePreservesDisplay(t *testing.T) {
	for _, tc := range []struct {
		name, paste, want string
		selectAll         bool
	}{
		{name: "single line", paste: "hello", want: "draft hello"},
		{name: "multiline CRLF", paste: "first\r\nsecond\rthird", want: "draft first\nsecond\nthird"},
		{name: "wide and combining", paste: "界e\u0301🙂", want: "draft 界e\u0301🙂"},
		{name: "tabs and trailing newline", paste: "a\tb\n", want: "draft a    b\n"},
		{name: "wrap and scroll", paste: strings.Repeat("long line ", 30) + "\nlast", want: "draft " + strings.Repeat("long line ", 30) + "\nlast"},
		{name: "empty", want: "draft "},
		{name: "replace selection", paste: "replacement\nline", want: "replacement\nline", selectAll: true},
		// TextEntry drops control runes, retaining printable sequence suffixes.
		{name: "terminal controls", paste: "\x1b[2J\x1b[Hsafe\x03\x0c", want: "draft [2J[Hsafe"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var changes, submits int
			e := NewTextEntry(TextEntryOptions{MaxHeight: 3,
				OnChange: func(string) { changes++ },
				OnSubmit: func(string) EventResult { submits++; return QuitResult() },
			})
			e.SetFocus(true)
			e.SetValue("draft ")
			if tc.selectAll {
				entryKey(e, "ctrl-a")
			}
			w := &chatPasteWidget{input: e}
			// Nonzero placement catches accidental painting outside the input.
			r := Rect{X: 1, Y: 1, W: 40, H: 8}
			c := NewCanvas(44, 10)
			w.Draw(c, r)
			terminal := ptytest.NewVT(44, 10)
			var out strings.Builder
			c.Flush(&out, 1)
			terminal.Write([]byte(out.String()))
			before := terminal.Cells()
			result := DispatchPasteEvent(w, PasteEvent{Text: tc.paste})
			if result.Quit || result.Consumed != (tc.paste != "") {
				t.Fatalf("paste result = %+v", result)
			}
			wantChanges := 0
			if tc.paste != "" {
				wantChanges = 1
			}
			if e.Value() != tc.want || changes != wantChanges || submits != 0 {
				t.Fatalf("value=%q changes=%d submits=%d, want %q", e.Value(), changes, submits, tc.want)
			}
			// A fresh frame also catches hosts losing surrounding content on redraw.
			c.Clear()
			w.Draw(c, r)
			out.Reset()
			c.Flush(&out, 1)
			terminal.Write([]byte(out.String()))
			after := terminal.Cells()
			for y := range after {
				for x := range after[y] {
					if !(Rect{X: 3, Y: 5, W: 36, H: 3}).Contains(x, y) && after[y][x] != before[y][x] {
						t.Fatalf("surrounding cell (%d,%d) changed: %+v -> %+v", x, y, before[y][x], after[y][x])
					}
				}
			}
			if tc.paste != "" {
				entryKey(e, "ctrl-z")
				if e.Value() != "draft " {
					t.Fatalf("paste must undo as one edit, got %q", e.Value())
				}
			}
		})
	}
}

func TestPaneChatPastePreservesDisplay(t *testing.T) {
	t.Setenv("LOOM_CHAT_PASTE_HELPER", "1")
	s := ptytest.Start(t, 44, 10, os.Args[0], "-test.run=^TestPaneChatPasteHelper$")
	s.WaitFor("assistant: previous answer", 3*time.Second)
	before := s.Cells()
	// Split both framing and payload across writes; follow with a real key in
	// the same read as the end marker. Control bytes inside paste are not keys.
	s.SendRaw([]byte("\x1b[20"))
	s.SendRaw([]byte("0~first\r\n\x1b[2J\x1b[Hsec"))
	s.SendRaw([]byte("ond\x03\x0c\x1b[201~!"))
	s.WaitForUpdate("second!", 3*time.Second)
	rows := s.Screen()
	if strings.TrimSpace(rows[4]) != "first" || strings.TrimSpace(rows[5]) != "[2J[Hsecond!" {
		t.Fatalf("pasted multiline display = %q / %q", rows[4], rows[5])
	}
	after := s.Cells()
	for y := range after {
		for x := range after[y] {
			if !(Rect{X: 2, Y: 4, W: 40, H: 3}).Contains(x, y) && after[y][x] != before[y][x] {
				t.Fatalf("PTY surrounding cell (%d,%d) changed: %+v -> %+v", x, y, before[y][x], after[y][x])
			}
		}
	}
	s.SendRaw([]byte{0x11}) // ctrl-q, after observing the completed paste frame
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("chat paste helper: %v\n%s", err, s.Raw())
	}
}

func TestPaneChatPasteHelper(t *testing.T) {
	if os.Getenv("LOOM_CHAT_PASTE_HELPER") != "1" {
		return
	}
	p, err := New(10)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	p.SetScreenMode(ScreenAlt)
	e := NewTextEntry(TextEntryOptions{MaxHeight: 3})
	e.SetFocus(true)
	w := &chatPasteWidget{input: e}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.run(ctx, w, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if e.Value() != "first\n[2J[Hsecond!" {
		t.Fatalf("inserted text = %q, want first\\n[2J[Hsecond!", e.Value())
	}
}
