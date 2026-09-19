// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strings"
	"sync"
)

// Nav is a navigation signal propagated from a widget through the caller chain.
type Nav int

const (
	NavNone Nav = iota // normal exit: selection made or user cancelled
	NavHome            // user invoked :home — unwind to the application root
)

// Cmd is a command reachable from any widget's prompt by typing `:name` or `/name`.
type Cmd struct {
	Name  string       // command name, e.g. "help", "back", "home"
	Title string       // dim hint shown in the prompt: ":h[elp]  <Title>"
	Keys  []string     // optional direct key bindings (reserved for future use)
	Fn    func() error // nil for built-in commands
}

// cmdResult is the internal outcome of a cmdBar event.
type cmdResult int

const (
	cmdNone cmdResult = iota
	cmdBack           // caller should abort its widget (like Esc)
	cmdHome           // caller should signal NavHome
)

// cmdBar manages command-mode input for Choice and Table.
// Typing ':' or '/' as the first character activates command mode;
// Esc or backspace-to-empty deactivates it.
type cmdBar struct {
	global []Cmd
	local  []Cmd
	active bool
	query  string
	help   *Popup
}

// paneHelpRequest is installed only while a Pane event loop is active. It
// lets command bars nested below clipped composite widgets request a root
// overlay without needing to know about the widget tree above them.
var paneHelpRequest func([]Cmd)

func newCmdBar() *cmdBar {
	return &cmdBar{
		global: []Cmd{
			{Name: "help", Title: "show help for current view"},
			{Name: "back", Title: "go one level back"},
		},
	}
}

// AddLocal registers view-local commands shown in this widget's :help.
func (cb *cmdBar) AddLocal(cmds ...Cmd) {
	cb.local = append(cb.local, cmds...)
}

// IsActive reports whether the bar is in command mode.
func (cb *cmdBar) IsActive() bool { return cb.active }

// PromptParts returns the typed prefix and dim-hint strings for command mode.
// Returns ("", "") when not active.
func (cb *cmdBar) PromptParts() (prefix, hint string) {
	if !cb.active {
		return "", ""
	}
	prefix = ":" + cb.query
	cmd := cb.match()
	if cmd == nil {
		return prefix, ""
	}
	rest := strings.TrimPrefix(cmd.Name, strings.ToLower(cb.query))
	if rest != "" {
		hint = "[" + rest + "]"
	}
	hint += "  " + cmd.Title
	return prefix, hint
}

// HandleKey processes a keyboard event for the command bar.
// Returns (consumed, result): consumed=true means the event was handled here.
func (cb *cmdBar) HandleKey(e KeyEvent) (consumed bool, result cmdResult) {
	if !cb.active {
		if e.Text == ":" || e.Text == "/" {
			cb.active = true
			cb.query = ""
			return true, cmdNone
		}
		return false, cmdNone
	}

	switch e.Key {
	case "esc":
		cb.active = false
		cb.query = ""
		return true, cmdNone
	case "backspace":
		if len(cb.query) == 0 {
			cb.active = false
		} else {
			r := []rune(cb.query)
			cb.query = string(r[:len(r)-1])
		}
		return true, cmdNone
	case "enter", "tab":
		if cmd := cb.match(); cmd != nil {
			return true, cb.execute(cmd)
		}
		cb.active = false
		cb.query = ""
		return true, cmdNone
	}

	if e.Text != "" {
		cb.query += e.Text
		return true, cmdNone
	}
	// Arrow keys and other specials: consume without action in command mode.
	return true, cmdNone
}

func (cb *cmdBar) allCmds() []Cmd {
	out := make([]Cmd, 0, len(cb.global)+len(cb.local))
	out = append(out, cb.global...)
	out = append(out, cb.local...)
	return out
}

func (cb *cmdBar) match() *Cmd {
	q := strings.ToLower(cb.query)
	if q == "" {
		return nil
	}
	all := cb.allCmds()
	for i := range all {
		if all[i].Name == q {
			return &all[i]
		}
	}
	for i := range all {
		if strings.HasPrefix(all[i].Name, q) {
			return &all[i]
		}
	}
	return nil
}

func (cb *cmdBar) execute(cmd *Cmd) cmdResult {
	cb.active = false
	cb.query = ""
	switch cmd.Name {
	case "back":
		return cmdBack
	case "home":
		return cmdHome
	case "help":
		if nav := cb.showHelp(); nav == NavHome {
			return cmdHome
		}
		return cmdNone
	default:
		if cmd.Fn != nil {
			_ = cmd.Fn()
		}
		return cmdNone
	}
}

var (
	helpMu     sync.RWMutex
	helpRunner func(Widget, int) error
)

func (cb *cmdBar) showHelp() Nav {
	hw := newHelpWidget(cb.allCmds())
	helpMu.RLock()
	runner := helpRunner
	helpMu.RUnlock()
	if runner != nil {
		_ = runner(hw, hw.ContentHeight())
	} else if paneHelpRequest != nil {
		paneHelpRequest(cb.allCmds())
	} else {
		cb.help = NewPopup("Help", hw)
	}
	return NavNone
}

func (cb *cmdBar) drawHelp(cv *Canvas, r Rect) {
	if cb.help != nil {
		cb.help.Draw(cv, r)
	}
}

func (cb *cmdBar) handleHelp(e KeyEvent) bool {
	if cb.help == nil {
		return false
	}
	if cb.help.HandleKey(e) {
		cb.help = nil
	}
	return true
}

func (cb *cmdBar) handleHelpMouse(e MouseEvent) bool {
	if cb.help == nil {
		return false
	}
	if cb.help.HandleMouse(e) {
		cb.help = nil
	}
	return true
}

// helpWidget is a read-only command list that closes on any key except up/down.
type helpWidget struct {
	lines  []string
	scroll int
}

func newHelpWidget(cmds []Cmd) *helpWidget {
	lines := make([]string, len(cmds))
	for i, c := range cmds {
		keys := ""
		if len(c.Keys) > 0 {
			keys = "  [" + strings.Join(c.Keys, ",") + "]"
		}
		lines[i] = fmt.Sprintf("  :%-10s  %s%s", c.Name, c.Title, keys)
	}
	return &helpWidget{lines: lines}
}

func (hw *helpWidget) ContentHeight() int {
	n := len(hw.lines) + 1
	if n > 20 {
		n = 20
	}
	return n
}

func (hw *helpWidget) Draw(cv *Canvas, r Rect) {
	for row := 0; row < r.H-1; row++ {
		y := r.Y + row
		cv.PaintSurface(Rect{r.X, y, r.W, 1}, Reset)
		i := hw.scroll + row
		if i < len(hw.lines) {
			cv.Write(r.X, y, TruncateText(hw.lines[i], r.W, ""), Reset)
		}
	}
	promptY := r.Y + r.H - 1
	cv.PaintSurface(Rect{r.X, promptY, r.W, 1}, Reset)
	cv.Write(r.X, promptY, TruncateText("  press any key to close", r.W, ""), Style{Dim: true})
}

func (hw *helpWidget) HandleKey(e KeyEvent) bool {
	switch e.Key {
	case "up":
		if hw.scroll > 0 {
			hw.scroll--
		}
		return false
	case "down":
		if hw.scroll < len(hw.lines)-1 {
			hw.scroll++
		}
		return false
	}
	return true // any other key closes
}

func (hw *helpWidget) HandleMouse(_ MouseEvent) bool { return false }
