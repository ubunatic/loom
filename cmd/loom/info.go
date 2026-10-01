// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"ubunatic.com/loom"
	"ubunatic.com/loom/measure"
)

func infoCommand() *cobra.Command {
	var watch bool
	cmd := &cobra.Command{
		Use: "info", Short: "Report terminal size and detected capabilities",
		Long: "Report the loom version, terminal size and environment, detected colour profile, Unicode sample width, and graphics path. With --watch, continuously show terminal size and resolved 0-based pointer coordinates alongside the original 1-based SGR mouse report; move the pointer across the panel to compare cell hit-testing. Press q or Esc to quit.",
		Args: cobra.NoArgs, SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if watch {
				pane, err := loom.New(1 << 16)
				if err != nil {
					return err
				}
				defer pane.Close()
				pane.MaxCols = 0
				pane.Resizeable = true
				return pane.Run(&infoWatch{})
			}
			cols, rows, err := loom.TerminalSize()
			if err != nil {
				cols, rows = 0, 0
			}
			return writeInfo(cmd.OutOrStdout(), cols, rows)
		},
	}
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch terminal changes and probe mouse coordinates")
	return cmd
}

func writeInfo(out interface{ Write([]byte) (int, error) }, cols, rows int) error {
	profile := colorProfileName(loom.DetectColorProfile())
	size := "unknown (pixels: unknown)"
	if cols > 0 && rows > 0 {
		size = fmt.Sprintf("%d x %d cells (pixels: unknown)", cols, rows)
	}
	_, err := fmt.Fprintf(out,
		"loom version: %s\nterminal size: %s\nTERM=%s\nCOLORTERM=%s\nTERM_PROGRAM=%s\ncolour depth: %s\nmouse tracking: SGR 1003 (requested by --watch)\ntrue colour: %t\nUnicode sample width: %d cells\ngraphics protocol: %s\n",
		loom.Version, size, envOrUnknown("TERM"), envOrUnknown("COLORTERM"), envOrUnknown("TERM_PROGRAM"),
		profile, loom.DetectColorProfile() == loom.ColorProfileTrueColor, measure.StringWidth("界🙂"), measure.DetectRenderPath())
	return err
}

func colorProfileName(profile loom.ColorProfile) string {
	switch profile {
	case loom.ColorProfileTrueColor:
		return "truecolor"
	case loom.ColorProfile256:
		return "256"
	case loom.ColorProfile16:
		return "16"
	default:
		return "none"
	}
}

func envOrUnknown(name string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return "unknown"
}

type infoWatch struct {
	x, y       int
	rawX, rawY int
	button     string
	hasPointer bool
}

func (w *infoWatch) Draw(canvas *loom.Canvas, rect loom.Rect) {
	cols, rows, err := loom.TerminalSize()
	if err != nil {
		cols, rows = rect.W, rect.H
	}
	lines := []string{
		"loom info --watch",
		fmt.Sprintf("terminal size: %d x %d cells", cols, rows),
		"mouse tracking: SGR 1003",
		fmt.Sprintf("pointer: x=%d y=%d raw=%d,%d button=%s", w.x, w.y, w.rawX, w.rawY, w.button),
		fmt.Sprintf("colour depth: %s  TERM=%s", colorProfileName(loom.DetectColorProfile()), envOrUnknown("TERM")),
		fmt.Sprintf("COLORTERM=%s TERM_PROGRAM=%s", envOrUnknown("COLORTERM"), envOrUnknown("TERM_PROGRAM")),
		"Move the mouse; q or Esc quits.",
	}
	for y, line := range lines {
		if y >= rect.H {
			break
		}
		canvas.Write(rect.X, rect.Y+y, line, loom.Style{})
	}
	if w.hasPointer && w.x >= 0 && w.x < rect.W && w.y >= 0 && w.y < rect.H {
		canvas.Set(rect.X+w.x, rect.Y+w.y, loom.Cell{Text: "+"})
	}
}

func (w *infoWatch) ConsumeKey(event loom.KeyEvent) loom.EventResult {
	if event.Is("q", "esc", "ctrl-c") {
		return loom.QuitResult()
	}
	return loom.Ignored()
}

func (w *infoWatch) ConsumeMouse(event loom.MouseEvent) loom.EventResult {
	w.x, w.y = event.X, event.Y
	w.rawX, w.rawY = event.RawX, event.RawY
	switch event.Button {
	case loom.MouseLeft:
		w.button = "left"
	case loom.MouseMiddle:
		w.button = "middle"
	case loom.MouseRight:
		w.button = "right"
	default:
		w.button = "none"
	}
	w.hasPointer = true
	return loom.Handled()
}

func (*infoWatch) PaneRequest() loom.PaneRequest {
	return loom.PaneRequest{Mouse: 1003, Resizeable: true, MaxCols: 0, OwnsQuit: true}
}

func (*infoWatch) TickInterval() time.Duration { return time.Second }

func (*infoWatch) Tick(time.Time) {}
