package loom

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOverlayMouseModeLifecycle(t *testing.T) {
	for _, base := range []int{0, 1000, 1003} {
		t.Run(fmt.Sprint(base), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "terminal")
			out, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			defer out.Close()
			p := &Pane{tty: out, baseMouseMode: base}
			p.reconcileMouseMode()
			p.overlayMouseGrab = true
			p.reconcileMouseMode()
			want := base
			if want == 0 {
				want = 1000
			}
			if !p.mouse || p.mouseMode != want || p.baseMouseMode != base {
				t.Fatalf("overlay mode=%d base=%d, want %d/%d", p.mouseMode, p.baseMouseMode, want, base)
			}
			// Repeated frames must not toggle tracking while the overlay stays open.
			p.reconcileMouseMode()
			p.overlayMouseGrab = false
			p.reconcileMouseMode()
			if p.mouseMode != base || p.mouse != (base != 0) {
				t.Fatalf("closing overlay restored mode %d, want %d", p.mouseMode, base)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			raw := string(data)
			if strings.Count(raw, "\x1b[?1006h") != 1 {
				t.Fatalf("redundant mouse enable sequences: %q", raw)
			}
			if base == 0 && !strings.HasSuffix(raw, "\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l") {
				t.Fatalf("missing full mouse release: %q", raw)
			}
			if base != 0 && strings.Contains(raw, "\x1b[?1006l") {
				t.Fatalf("persistent grab was disabled: %q", raw)
			}
		})
	}
}

func TestOverlayMouseModeBaseToggle(t *testing.T) {
	out, err := os.CreateTemp(t.TempDir(), "terminal")
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	p := &Pane{tty: out, overlayMouseGrab: true}
	p.EnableMouse()
	p.DisableMouse()
	if !p.mouse || p.mouseMode != 1000 || p.baseMouseMode != 0 {
		t.Fatal("disabling global mouse interrupted overlay capture")
	}
	p.EnableMouse()
	p.overlayMouseGrab = false
	p.reconcileMouseMode()
	if p.mouseMode != 1003 {
		t.Fatal("closing overlay lost newly enabled global grab")
	}
	p.DisableMouse()
	if p.mouse {
		t.Fatal("global mouse remained enabled")
	}
}

func TestOverlayMouseRequests(t *testing.T) {
	for _, kind := range []string{"popup", "dialog", "menu", "menubar", "help", "popover"} {
		t.Run(kind, func(t *testing.T) {
			c := NewCanvas(80, 24)
			var w Widget
			switch kind {
			case "popup":
				w = NewPopup("Popup", NewRichTextEdit(nil))
			case "dialog":
				w = NewDialog("Dialog", "Body", "OK", "Cancel")
			case "menu":
				w = &Menu{Items: []MenuItem{{Label: "Item"}}}
			case "menubar":
				w = &MenuBar{Open: true, Menus: []Menu{{Title: "Menu", Items: []MenuItem{{Label: "Item"}}}}}
			case "help", "popover":
				e := NewRichTextEdit(nil)
				e.Draw(c, c.Bounds())
				key := "f1"
				if kind == "popover" {
					key = "ctrl-space"
				}
				e.ConsumeKey(KeyEvent{Key: key})
				w = e
			}
			c.Clear()
			w.Draw(c, c.Bounds())
			if !c.overlayMouseGrab {
				t.Fatal("visible overlay did not request mouse tracking")
			}
			if kind == "menu" {
				// Standalone Menu.Draw opens the menu; its host stops drawing it on close.
				w.ConsumeKey(KeyEvent{Key: "esc"})
			} else {
				w.ConsumeMouse(MouseEvent{X: 79, Y: 23, Action: MousePress, Button: MouseLeft})
			}
			c.Clear()
			if kind != "menu" {
				w.Draw(c, c.Bounds())
			}
			if c.overlayMouseGrab {
				t.Fatal("dismissed overlay retained mouse capture")
			}
		})
	}
}

func TestOverlayMouseStackAndClipping(t *testing.T) {
	c := NewCanvas(80, 24)
	a := NewPopup("First", NewRichTextEdit(nil))
	b := NewDialog("Second", "Body", "OK")
	draw := func() {
		c.Clear()
		paintClipped(c, c.Bounds(), func(local *Canvas) {
			a.Draw(local, local.Bounds())
			b.Draw(local, local.Bounds())
		})
	}
	draw()
	a.ConsumeKey(KeyEvent{Key: "esc"})
	draw()
	if !c.overlayMouseGrab {
		t.Fatal("closing one overlay released another overlay's grab")
	}
	b.ConsumeKey(KeyEvent{Key: "enter"})
	draw()
	if c.overlayMouseGrab {
		t.Fatal("last closed overlay retained grab")
	}
	b.Open = true
	sub := NewCanvas(40, 12)
	b.Draw(sub, sub.Bounds())
	c.Blit(sub.SubCanvas(sub.Bounds()), 0, 0)
	if !c.overlayMouseGrab {
		t.Fatal("blitted overlay lost its grab request")
	}
	c.Clear()
	c.Blit(sub, 100, 100)
	if c.overlayMouseGrab {
		t.Fatal("fully clipped overlay requested grab")
	}
}

func TestPopoverBackdropPreservesDocumentAndCursor(t *testing.T) {
	e := NewRichTextEdit(nil)
	c := NewCanvas(80, 24)
	e.Draw(c, c.Bounds())
	e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
	e.Draw(c, c.Bounds())
	before := e.Cursor
	if !e.ModalOpen() {
		t.Fatal("format popover does not capture host backdrop clicks")
	}
	e.ConsumeMouse(MouseEvent{X: 79, Y: 23, Action: MousePress, Button: MouseLeft})
	c.Clear()
	e.Draw(c, c.Bounds())
	if e.ModalOpen() || c.overlayMouseGrab || e.Cursor != before || e.IsModified() {
		t.Fatal("backdrop click did not dismiss without editing the background")
	}
}
