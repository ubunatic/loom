// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
)

func frameRow(c *Canvas, y, w int) string {
	var b strings.Builder
	for x := 0; x < w; x++ {
		b.WriteString(c.Get(x, y).Text)
	}
	return b.String()
}

func TestTextEntryBoxFrame(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Prompt: "❯ ", Frame: TextEntryFrame{Style: TextEntryFrameBox, Title: "chat", Hint: "F1", Footer: "ok"}})
	e.SetValue("hi")
	if got := e.Measure(20); got.Height != 3 || got.Width != 20 {
		t.Fatalf("Measure = %+v, want 20x3", got)
	}
	c := NewCanvas(20, 3)
	e.Draw(c, Rect{W: 20, H: 3})
	if got := frameRow(c, 0, 20); !strings.HasPrefix(got, "╭─ chat ─") || !strings.HasSuffix(got, "─ F1 ─╮") {
		t.Fatalf("top = %q", got)
	}
	if got := frameRow(c, 1, 20); !strings.HasPrefix(got, "│ ❯ hi") || !strings.HasSuffix(got, " │") {
		t.Fatalf("middle = %q", got)
	}
	if got := frameRow(c, 2, 20); !strings.HasPrefix(got, "╰─") || !strings.HasSuffix(got, "─ ok ─╯") {
		t.Fatalf("bottom = %q", got)
	}
	if !c.Get(0, 0).Style.Dim || !c.Get(3, 0).Style.Dim {
		t.Fatal("border and title default to dim")
	}
	if row, col := e.Cursor(); row != 1 || col != 6 {
		t.Fatalf("cursor = %d,%d, want 1,6 (inside the frame)", row, col)
	}
}

func TestTextEntrySimpleFrame(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Prompt: "❯ ", Frame: TextEntryFrame{Style: TextEntryFrameSimple, Title: "t", Hint: "h", Footer: "f"}})
	e.SetValue("hi")
	c := NewCanvas(16, 3)
	e.Draw(c, Rect{W: 16, H: 3})
	if got := frameRow(c, 0, 16); !strings.HasPrefix(got, "── t ─") || !strings.HasSuffix(got, "─ h ──") {
		t.Fatalf("top = %q", got)
	}
	if got := frameRow(c, 1, 16); !strings.HasPrefix(got, "❯ hi") {
		t.Fatalf("middle = %q", got)
	}
	if got := frameRow(c, 2, 16); !strings.HasSuffix(got, "─ f ──") {
		t.Fatalf("bottom = %q", got)
	}
}

func TestTextEntryFrameColors(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Frame: TextEntryFrame{Style: TextEntryFrameSimple, BorderColor: "#ff0000", LabelColor: "7", Title: "x"}})
	c := NewCanvas(12, 3)
	e.Draw(c, Rect{W: 12, H: 3})
	if got := c.Get(0, 0).Style; got.Dim || got.FG != ColorRGB(255, 0, 0) {
		t.Fatalf("border style = %+v", got)
	}
	if got := c.Get(3, 0).Style; got.FG != ColorIndex(7) {
		t.Fatalf("label style = %+v", got)
	}
	theme := DefaultTheme()
	e.SetFrame(TextEntryFrame{Style: TextEntryFrameSimple, BorderColor: "status_fg"})
	e.ApplyTheme(theme)
	if got := e.frameColor("status_fg"); got.Dim || got.FG != theme.StatusFG.Color() {
		t.Fatalf("theme color = %+v, want %+v", got, theme.StatusFG.Color())
	}
	if got := e.frameColor("nonsense"); !got.Dim {
		t.Fatal("unknown names fall back to dim")
	}
}

func TestTextEntryFrameTooSmallDrawsContentOnly(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Frame: TextEntryFrame{Style: TextEntryFrameBox}})
	e.SetValue("ab")
	c := NewCanvas(10, 2)
	e.Draw(c, Rect{W: 10, H: 2})
	if got := frameRow(c, 0, 10); !strings.HasPrefix(got, "ab") {
		t.Fatalf("row = %q, want plain text", got)
	}
}

func TestTextEntryHeightBounds(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	e.SetValue("a\nb\nc\nd\ne\nf\ng\nh\ni\nj")
	c := NewCanvas(20, 20) // auto max: 30% of 20 rows
	e.Draw(c, Rect{W: 20, H: 20})
	if got := e.Measure(20).Height; got != 6 {
		t.Fatalf("auto height = %d, want 6", got)
	}
	e.SetHeights(0, 0, -1)
	if got := e.Measure(20).Height; got != 10 {
		t.Fatalf("unbounded height = %d, want 10", got)
	}
	e.SetHeights(0, 0, 3)
	if got := e.Measure(20).Height; got != 3 {
		t.Fatalf("max 3 = %d", got)
	}
	e.SetValue("")
	e.SetHeights(2, 0, 4)
	if got := e.Measure(20).Height; got != 2 {
		t.Fatalf("default height 2 = %d", got)
	}
	e.SetHeights(0, 3, 0)
	if got := e.Measure(20).Height; got != 3 {
		t.Fatalf("min height 3 = %d", got)
	}
	e.SetHeights(0, 5, 2)
	if got := e.Measure(20).Height; got != 5 {
		t.Fatalf("min wins over max = %d", got)
	}
}

func TestTextEntrySendButtonSubmitsOnClick(t *testing.T) {
	var got []string
	e := NewTextEntry(TextEntryOptions{SendButton: true, SendLabel: " go ", Frame: TextEntryFrame{Style: TextEntryFrameBox},
		OnSubmit: func(v string) EventResult { got = append(got, v); return Handled() }})
	e.SetValue("hello")
	c := NewCanvas(20, 3)
	e.Draw(c, Rect{W: 20, H: 3})
	if e.sendRect.W != 4 {
		t.Fatalf("send rect = %+v", e.sendRect)
	}
	if text := c.Get(e.sendRect.X+1, e.sendRect.Y).Text; text != "g" {
		t.Fatalf("button cell = %q, want g", text)
	}
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: e.sendRect.X + 1, Y: e.sendRect.Y})
	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf("submitted %q, want [hello]", got)
	}
	// A press on the frame does not move the caret.
	e.SetCursorOffset(5)
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 0, Y: 0})
	if e.CursorOffset() != 5 {
		t.Fatalf("caret moved to %d by a border press", e.CursorOffset())
	}
	// A press on the text lands at the shifted cell: x=2 is the first text column.
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 3, Y: 1})
	if e.CursorOffset() != 1 {
		t.Fatalf("caret = %d, want 1", e.CursorOffset())
	}
}

func TestTextEntrySendButtonOffByDefault(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	e.Draw(NewCanvas(10, 1), Rect{W: 10, H: 1})
	if e.sendRect.W != 0 || e.buttonWidth() != 0 {
		t.Fatal("no send button unless asked for")
	}
}

func TestTextEntryScrollbarSitsAboveSendButton(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{MaxHeight: 3, SendButton: true, SendLabel: " go "})
	e.SetValue("a\nb\nc\nd\ne\nf")
	c := NewCanvas(20, 3)
	e.Draw(c, Rect{W: 20, H: 3})
	if !e.bar || !e.barAbove || e.layoutWidth != 16 {
		t.Fatalf("bar=%v above=%v text width=%d, want a bar above the button and an unreduced width 16", e.bar, e.barAbove, e.layoutWidth)
	}
	x := 16 + 2 // middle of the 4-cell button
	if e.barX != x || e.barH != 2 {
		t.Fatalf("bar column %d rows %d, want %d and 2", e.barX, e.barH, x)
	}
	if c.Get(x, 0).Text == " " && c.Get(x, 0).Style == (Style{}) {
		t.Fatal("no bar cell above the button")
	}
	if c.Get(x, 2).Text != "o" && c.Get(x, 2).Text != "g" {
		t.Fatalf("button row cell = %q, want the label", c.Get(x, 2).Text)
	}
	// Dragging the bar scrolls; a click on the button still sends.
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: x, Y: 0})
	e.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: x, Y: 0})
	if e.scroll == 3 {
		t.Fatal("a track click above the button did not scroll")
	}
	// A one-row entry has no row above the button: the bar stays beside the text.
	one := NewTextEntry(TextEntryOptions{MaxHeight: 1, SendButton: true, SendLabel: " go "})
	one.SetValue("a\nb\nc")
	one.Draw(NewCanvas(20, 1), Rect{W: 20, H: 1})
	if one.barAbove {
		t.Fatal("one row: the bar must stay beside the text")
	}
}

func TestTextEntrySetPromptReindentsAndRewraps(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Prompt: "> ", MaxHeight: -1})
	e.SetValue("aaaa bbbb cccc")
	c := NewCanvas(12, 3)
	e.Draw(c, Rect{W: 12, H: 3})
	if got := frameRow(c, 0, 12); !strings.HasPrefix(got, "> aaaa") {
		t.Fatalf("row 0 = %q", got)
	}
	e.SetPrompt("‼︎ !! ")
	if e.Prompt() != "‼︎ !! " {
		t.Fatalf("Prompt() = %q", e.Prompt())
	}
	c = NewCanvas(12, 3)
	e.Draw(c, Rect{W: 12, H: 3})
	w := StringWidth("‼︎ !! ")
	if got := c.Get(w, 0).Text; got != "a" {
		t.Fatalf("text starts at column %d with %q, want a", w, got)
	}
	if row, col := e.Cursor(); row < 0 || col < w {
		t.Fatalf("cursor = %d,%d, want it right of the prompt", row, col)
	}
	if e.Measure(12).Height < 2 {
		t.Fatal("a wider prompt must wrap the text onto more rows")
	}
}
