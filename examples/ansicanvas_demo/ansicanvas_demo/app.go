// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package ansicanvas_demo demonstrates embedding loom.AnsiEditor directly
// in a clean, minimal Loom frame for 2D ANSI graphic text editing.
package ansicanvas_demo

import (
	_ "embed"
	"fmt"

	"codeberg.org/ubunatic/loom"
	"github.com/spf13/cobra"
)

//go:embed loom-logo.ansi
var defaultLogoANSI []byte

// DemoApp embeds loom.AnsiEditor inside a minimal Loom Frame.
type DemoApp struct {
	frame  *loom.Frame
	editor *loom.AnsiEditor
	quit   bool
}

// NewDemoApp creates a new DemoApp for the given ANSI buffer.
func NewDemoApp(buf *loom.AnsiBuffer) *DemoApp {
	if buf == nil {
		if len(defaultLogoANSI) > 0 {
			parsed, err := loom.ParseAnsiBuffer(string(defaultLogoANSI), 54, 14)
			if err == nil {
				buf = parsed
			}
		}
		if buf == nil {
			buf = loom.NewAnsiBuffer(60, 18)
		}
		buf.SetModified(false)
	}

	editor := loom.NewAnsiEditor(buf)
	editor.ActiveFG = loom.ColorIndex(15)

	frame := &loom.Frame{
		Title: "🎨 ANSI Canvas Demo",
		Boxes: []loom.Box{
			{
				ID:       "canvas",
				Title:    "Graphic Cell Buffer",
				Width:    64,
				Height:   20,
				Dynamic:  true,
				MinWidth: 30,
				Child:    editor,
			},
		},
		Actions: []loom.FrameAction{
			{ID: "quit", Action: "quit", Key: "F10"},
			{ID: "quit_q", Action: "quit", Key: "q"},
		},
	}

	return &DemoApp{
		frame:  frame,
		editor: editor,
	}
}

// Editor returns the embedded AnsiEditor widget.
func (a *DemoApp) Editor() *loom.AnsiEditor {
	return a.editor
}

// Draw renders the frame with live cursor coordinates in the status footer.
func (a *DemoApp) Draw(c *loom.Canvas, r loom.Rect) {
	curX, curY := a.editor.Cursor()
	a.frame.Status = fmt.Sprintf("Cursor: (%d,%d) • Mode: %s • F10 Quit • q Quit", curX, curY, a.editor.EditMode)
	a.frame.Draw(c, r)
}

// ConsumeKey handles key events and routes them to the editor and frame actions.
func (a *DemoApp) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if e.Is("f10", "ctrl-q") {
		a.quit = true
		return loom.QuitResult()
	}
	if e.Key == "q" || (e.Text == "q" && e.Key == "") {
		// q quits if at start or quit key pressed
		a.quit = true
		return loom.QuitResult()
	}
	res := a.editor.ConsumeKey(e)
	if res.Consumed {
		return res
	}
	if a.frame.HandleKey(e) {
		a.quit = true
		return loom.QuitResult()
	}
	return loom.Ignored()
}

// HandleKey implements loom.Widget.
func (a *DemoApp) HandleKey(e loom.KeyEvent) bool {
	res := a.ConsumeKey(e)
	return res.Quit
}

// ConsumeMouse handles mouse events and routes them to the editor.
func (a *DemoApp) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	return a.editor.ConsumeMouse(e)
}

// HandleMouse implements loom.Widget.
func (a *DemoApp) HandleMouse(e loom.MouseEvent) bool {
	res := a.ConsumeMouse(e)
	return res.Quit
}

// PaneRequest declares terminal capabilities.
func (a *DemoApp) PaneRequest() loom.PaneRequest {
	return loom.PaneRequest{
		Mouse:      1003,
		Resizeable: true,
		OwnsQuit:   true,
	}
}

// NewWidget builds the demo app widget for in-process hosting.
func NewWidget(args []string) (loom.Widget, error) {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	var buf *loom.AnsiBuffer
	var err error
	if path != "" {
		buf, err = loom.LoadAnsiBuffer(path, 60, 18)
		if err != nil {
			return nil, err
		}
	}
	return NewDemoApp(buf), nil
}

// Run executes the ansicanvas_demo command.
func Run(args []string) error {
	var cursorFX bool
	cmd := &cobra.Command{
		Use:           "ansicanvas_demo [file.ansi]",
		Short:         "Minimal demonstration of embedding loom.AnsiEditor",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, cmdArgs []string) error {
			path := ""
			if len(cmdArgs) > 0 {
				path = cmdArgs[0]
			}
			var buf *loom.AnsiBuffer
			var err error
			if path != "" {
				buf, err = loom.LoadAnsiBuffer(path, 60, 18)
				if err != nil {
					return err
				}
			}
			app := NewDemoApp(buf)
			pane, err := loom.New(1 << 16)
			if err != nil {
				return err
			}
			defer pane.Close()
			pane.Resizeable = true
			pane.DisableDefaultQuit = true
			if cursorFX {
				pane.EnableCursorStarTrail()
			} else {
				pane.EnableMouse()
			}
			return pane.Run(app)
		},
	}
	cmd.Flags().BoolVar(&cursorFX, "cursor-fx", false, "show the cursor proximity glow and star trail")
	cmd.SetArgs(args)
	return cmd.Execute()
}
