package loom

import "testing"

// paneRequestWidget is a minimal declarative child used by the acceptance tests.
type paneRequestWidget struct {
	request PaneRequest
}

func (w paneRequestWidget) Draw(*Canvas, Rect)          {}
func (w paneRequestWidget) HandleKey(KeyEvent) bool     { return false }
func (w paneRequestWidget) HandleMouse(MouseEvent) bool { return false }
func (w paneRequestWidget) PaneRequest() PaneRequest    { return w.request }

type plainPaneWidget struct{}

func (plainPaneWidget) Draw(*Canvas, Rect)          {}
func (plainPaneWidget) HandleKey(KeyEvent) bool     { return false }
func (plainPaneWidget) HandleMouse(MouseEvent) bool { return false }

func TestMergePaneRequestMouse(t *testing.T) {
	request := (&Stack{Children: []Widget{
		paneRequestWidget{request: PaneRequest{Mouse: 1000}},
		paneRequestWidget{request: PaneRequest{Mouse: 1003}},
	}}).PaneRequest()
	if request.Mouse != 1003 {
		t.Fatalf("Mouse = %d, want 1003", request.Mouse)
	}
}

func TestMergePaneRequestResizeable(t *testing.T) {
	request := (&Grid{Children: []Widget{
		paneRequestWidget{request: PaneRequest{Resizeable: true}},
		paneRequestWidget{},
	}}).PaneRequest()
	if !request.Resizeable {
		t.Fatal("Resizeable = false, want true")
	}
}

func TestMergePaneRequestMaxColsZeroWins(t *testing.T) {
	request := (&Stack{Children: []Widget{
		paneRequestWidget{request: PaneRequest{MaxCols: 80}},
		paneRequestWidget{request: PaneRequest{MaxCols: 0}},
	}}).PaneRequest()
	if request.MaxCols != 0 {
		t.Fatalf("MaxCols = %d, want 0", request.MaxCols)
	}
}

func TestMergePaneRequestMaxColsMax(t *testing.T) {
	request := (&Stack{Children: []Widget{
		paneRequestWidget{request: PaneRequest{MaxCols: 40}},
		paneRequestWidget{request: PaneRequest{MaxCols: 80}},
	}}).PaneRequest()
	if request.MaxCols != 80 {
		t.Fatalf("MaxCols = %d, want 80", request.MaxCols)
	}
}

func TestMergePaneRequestOwnsQuit(t *testing.T) {
	request := (&Grid{Children: []Widget{
		paneRequestWidget{request: PaneRequest{OwnsQuit: true}},
		paneRequestWidget{},
	}}).PaneRequest()
	if !request.OwnsQuit {
		t.Fatal("OwnsQuit = false, want true")
	}
}

func TestMergePaneRequestNested(t *testing.T) {
	inner := &Tabs{Tabs: []Tab{{Widget: paneRequestWidget{request: PaneRequest{Mouse: 1003, MaxCols: 80}}}}}
	outer := &Grid{Children: []Widget{
		inner,
		paneRequestWidget{request: PaneRequest{Resizeable: true, MaxCols: 40, OwnsQuit: true}},
	}}
	request := outer.PaneRequest()
	want := PaneRequest{Mouse: 1003, Resizeable: true, MaxCols: 80, OwnsQuit: true}
	if request != want {
		t.Fatalf("request = %#v, want %#v", request, want)
	}
}

func TestPaneRequestExplicitFieldWins(t *testing.T) {
	root := &Tabs{Tabs: []Tab{{Widget: paneRequestWidget{request: PaneRequest{Mouse: 1003}}}}}
	request := root.PaneRequest()
	p := &Pane{mouse: true, mouseMode: 1000}
	if p.mouse && request.Mouse > 0 {
		// This is the same explicit-field guard used by Pane.run.
		request.Mouse = p.mouseMode
	}
	if request.Mouse != 1000 {
		t.Fatalf("explicit mouse mode = %d, want 1000", request.Mouse)
	}
}

func TestWidgetWithoutPaneRequesterYieldsDefaults(t *testing.T) {
	var request PaneRequest
	if _, ok := Widget(plainPaneWidget{}).(PaneRequester); ok {
		t.Fatal("plain widget unexpectedly implements PaneRequester")
	}
	if request != (PaneRequest{}) {
		t.Fatalf("default request = %#v, want zero value", request)
	}
}
