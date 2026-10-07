package loom

import (
	"bytes"
	"testing"
)

func TestCursorShapeStateWritesChangesAndRestores(t *testing.T) {
	var output bytes.Buffer
	var state cursorShapeState
	if err := state.apply(&output, CursorShapeBar); err != nil {
		t.Fatal(err)
	}
	if err := state.apply(&output, CursorShapeBar); err != nil {
		t.Fatal(err)
	}
	if err := state.apply(&output, CursorShapeBlock); err != nil {
		t.Fatal(err)
	}
	if err := state.restore(&output); err != nil {
		t.Fatal(err)
	}
	if err := state.restore(&output); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "\x1b[6 q\x1b[2 q\x1b[0 q"; got != want {
		t.Fatalf("cursor shape sequences = %q, want %q", got, want)
	}
}

func TestCursorShapeStateRejectsUnknownShape(t *testing.T) {
	var output bytes.Buffer
	var state cursorShapeState
	if err := state.apply(&output, CursorShape(255)); err == nil {
		t.Fatal("unknown cursor shape was accepted")
	}
	if output.Len() != 0 || state.set {
		t.Fatalf("invalid shape changed output/state: %q %+v", output.String(), state)
	}
}

func TestRichTextEditReportsCursorShapeByMode(t *testing.T) {
	e := richEditorLines("x")
	if got := e.CursorShape(); got != CursorShapeBar {
		t.Fatalf("normal cursor shape = %v, want bar", got)
	}
	e.BoxMode = true
	if got := e.CursorShape(); got != CursorShapeBlock {
		t.Fatalf("box mode cursor shape = %v, want block", got)
	}
}

type cursorShapeTestWrapper struct{ child Widget }

func (w cursorShapeTestWrapper) Draw(c *Canvas, r Rect) { w.child.Draw(c, r) }
func (w cursorShapeTestWrapper) ConsumeKey(e KeyEvent) EventResult {
	return w.child.ConsumeKey(e)
}
func (w cursorShapeTestWrapper) ConsumeMouse(e MouseEvent) EventResult {
	return w.child.ConsumeMouse(e)
}

func TestNestedTabsCursorShapeReachesPaneOutput(t *testing.T) {
	edit := richEditorLines("text")
	tabs := NewTabs(Tab{Title: "Editor", Widget: edit})
	root := cursorShapeTestWrapper{child: tabs}
	canvas := NewCanvas(50, 8)
	var output bytes.Buffer
	pane := &Pane{}

	root.Draw(canvas, canvas.Bounds())
	pane.applyCanvasCursorShape(&output, canvas)
	if result := root.ConsumeKey(KeyEvent{Key: "ctrl-d"}); !result.Consumed || !edit.BoxMode {
		t.Fatalf("^D in nested editor = %+v, BoxMode=%v", result, edit.BoxMode)
	}
	canvas.Clear()
	root.Draw(canvas, canvas.Bounds())
	pane.applyCanvasCursorShape(&output, canvas)
	if err := pane.cursorShape.restore(&output); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "\x1b[6 q\x1b[2 q\x1b[0 q"; got != want {
		t.Fatalf("nested cursor shape sequences = %q, want %q", got, want)
	}
}
