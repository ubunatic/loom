// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package suspend shows Pane.Suspend: Ctrl+G hands the terminal to $EDITOR
// and the pane resumes with the edited note.
package suspend

import (
	"flag"
	"os"
	"os/exec"
	"strings"

	"ubunatic.com/loom"
)

// App is a note viewer whose note is edited in an external editor.
type App struct {
	pane   *loom.Pane
	note   string
	status string
}

// Draw renders the note and the key help.
func (a *App) Draw(c *loom.Canvas, r loom.Rect) {
	c.Write(r.X, r.Y, "Ctrl+G edit in $EDITOR   q quit", loom.Style{Bold: true})
	lines := strings.Split(strings.TrimRight(a.note, "\n"), "\n")
	for i, line := range lines {
		if i+1 >= r.H-1 {
			break
		}
		c.Write(r.X, r.Y+1+i, line, loom.Style{})
	}
	if a.status != "" && r.H > 1 {
		c.Write(r.X, r.Y+r.H-1, a.status, loom.Style{Dim: true})
	}
}

// ConsumeMouse ignores mouse input.
func (a *App) ConsumeMouse(loom.MouseEvent) loom.EventResult { return loom.Ignored() }

// ConsumeKey opens the editor on Ctrl+G and quits on q.
func (a *App) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	switch {
	case e.Is("ctrl-g"):
		a.edit()
		return loom.Consumed()
	case e.Is("q"):
		return loom.QuitResult()
	}
	return loom.Ignored()
}

func (a *App) edit() {
	f, err := os.CreateTemp("", "loom-note-*.txt")
	if err != nil {
		a.status = err.Error()
		return
	}
	defer os.Remove(f.Name())
	_, _ = f.WriteString(a.note)
	_ = f.Close()
	err = a.pane.Suspend(func() error { return runEditor(f.Name()) })
	if err != nil {
		a.status = "editor: " + err.Error()
		return
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		a.status = err.Error()
		return
	}
	a.note, a.status = string(data), "note updated"
}

// runEditor runs $VISUAL or $EDITOR (default vi) on the controlling terminal.
func runEditor(path string) error {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer tty.Close()
	// sh -c lets $EDITOR carry arguments, e.g. "code --wait".
	cmd := exec.Command("sh", "-c", editor+` "$1"`, "sh", path)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = tty, tty, tty
	return cmd.Run()
}

// Run starts the example.
func Run(args []string) error {
	flags := flag.NewFlagSet("suspend", flag.ContinueOnError)
	if err := flags.Parse(args); err != nil {
		return err
	}
	pane, err := loom.New(6)
	if err != nil {
		return err
	}
	defer pane.Close()
	app := &App{pane: pane, note: "Press Ctrl+G to edit this note.\n"}
	return pane.Run(app)
}
