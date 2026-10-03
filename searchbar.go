// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// SearchBarStyle controls the visual appearance of a SearchBar widget.
type SearchBarStyle struct {
	Container   Style // full search bar container background and foreground
	Prompt      Style // prompt prefix (e.g. "> " or "⌕ ")
	Query       Style // typed query text
	Placeholder Style // hint shown when Query is empty
	Controls    Style // right-aligned controls/hints
}

// DefaultSearchBarStyle returns a minimal monochrome style, derived from the plain theme.
func DefaultSearchBarStyle() SearchBarStyle {
	return Theme("plain").SearchBarStyle()
}

// SearchBar is a reusable search and filter input widget.
// It manages prompt rendering, query input, cursor positioning, placeholder hints,
// right-aligned controls, and command-mode (: / /) interactions.
type SearchBar struct {
	Prompt      string         // prompt prefix, default "> "
	Placeholder string         // dim hint shown when Query is empty; "" disables
	Query       string         // current typed filter text
	Controls    string         // right-aligned controls/status label
	Focused     bool           // whether search bar is focused (shows cursor)
	CursorAlign string         // "start" or "end", default "start"
	Style       SearchBarStyle // visual styling
	MaxWidth    int            // optional max render width

	OnChange func(query string) // called when Query changes
	OnSubmit func(query string) // called on Enter
	OnAbort  func()             // called on Esc / quit keys

	cmd     *cmdBar
	cmdNav  Nav
	aborted bool
	done    bool
	drawn   bool
	lastRect Rect
}

// NewSearchBar creates a ready-to-use SearchBar with default style and prompt.
func NewSearchBar() *SearchBar {
	sb := &SearchBar{
		Prompt:      SpeccedDefaults.SearchBar.Prompt,
		Placeholder: SpeccedDefaults.SearchBar.Placeholder,
		Focused:     true,
		Style:       DefaultSearchBarStyle(),
	}
	sb.cmd = newCmdBar()
	return sb
}

// Nav returns the navigation signal set by a prompt command (:home).
func (s *SearchBar) Nav() Nav { return s.cmdNav }

// AddCmd registers a view-local command accessible via ':name' in this widget.
func (s *SearchBar) AddCmd(cmd Cmd) {
	if s.cmd == nil {
		s.cmd = newCmdBar()
	}
	s.cmd.AddLocal(cmd)
}

// IsCommandMode reports whether command mode is currently active.
func (s *SearchBar) IsCommandMode() bool { return s.cmd != nil && s.cmd.IsActive() }

// SetQuery sets the query string, updating state and firing OnChange if changed.
func (s *SearchBar) SetQuery(q string) {
	if s.Query != q {
		s.Query = q
		if s.OnChange != nil {
			s.OnChange(q)
		}
	}
}

// Clear resets the query string to empty.
func (s *SearchBar) Clear() {
	s.SetQuery("")
}

// SetFocused updates focus state.
func (s *SearchBar) SetFocused(focused bool) {
	s.Focused = focused
}

// Draw renders the search bar into region r.
func (s *SearchBar) Draw(cv *Canvas, r Rect) {
	s.drawn = true
	s.lastRect = r
	if r.H <= 0 || r.W <= 0 {
		return
	}

	drawW := r.W
	if s.MaxWidth > 0 && drawW > s.MaxWidth {
		drawW = s.MaxWidth
	}

	surfaceStyle := s.Style.Container
	if surfaceStyle == (Style{}) {
		surfaceStyle = s.Style.Prompt
	}
	cv.PaintDefaultSurface(Rect{r.X, r.Y, drawW, 1}, surfaceStyle)

	promptStyle := s.Style.Prompt
	if promptStyle == (Style{}) {
		promptStyle = surfaceStyle
	}

	if s.cmd != nil {
		if prefix, hint := s.cmd.PromptParts(); prefix != "" {
			// Command mode: ":typed[completion]  dim title"
			n := cv.WriteDefault(r.X, r.Y, prefix, promptStyle)
			if hint != "" {
				cv.WriteDefault(r.X+n, r.Y, hint, Style{Dim: true})
			}
			if s.Focused {
				cv.CursorX = r.X + n
				cv.CursorY = r.Y
			}
			return
		}
	}

	// Normal mode: prompt + query or placeholder
	promptText := s.Prompt
	if promptText == "" {
		promptText = "> "
	}
	n := cv.WriteDefault(r.X, r.Y, promptText, promptStyle)

	if s.Query == "" && s.Placeholder != "" {
		hint := s.Style.Placeholder
		if hint == (Style{}) {
			hint = promptStyle
		}
		hint.Dim = true
		cv.WriteDefault(r.X+n, r.Y, TruncateText(s.Placeholder, max(0, drawW-n), ""), hint)
	} else if s.Query != "" {
		queryStyle := s.Style.Query
		if queryStyle == (Style{}) {
			queryStyle = promptStyle
		}
		cv.WriteDefault(r.X+n, r.Y, s.Query, queryStyle)
	}

	if s.Controls != "" {
		ctrlW := StringWidth(s.Controls)
		if drawW > ctrlW+StringWidth(promptText) {
			ctrlStyle := s.Style.Controls
			if ctrlStyle == (Style{}) {
				ctrlStyle = Style{Dim: true}
			}
			cv.WriteDefault(r.X+drawW-ctrlW, r.Y, s.Controls, ctrlStyle)
		}
	}

	if s.Focused {
		cursorX := r.X + StringWidth(promptText) + StringWidth(s.Query)
		if s.CursorAlign == "end" {
			cursorX = r.X + drawW - 1
		}
		if cursorX >= r.X && cursorX < r.X+drawW {
			cv.CursorX = cursorX
			cv.CursorY = r.Y
		}
	}
}

// ConsumeKey processes keyboard events for SearchBar.
func (s *SearchBar) ConsumeKey(e KeyEvent) EventResult {
	if s.cmd != nil {
		if s.cmd.handleHelp(e) {
			return Handled()
		}
		if consumed, result := s.cmd.ConsumeKey(e); consumed {
			switch result {
			case cmdBack:
				s.aborted = true
				s.done = true
				if s.OnAbort != nil {
					s.OnAbort()
				}
				return QuitResult()
			case cmdHome:
				s.cmdNav = NavHome
				s.aborted = true
				s.done = true
				if s.OnAbort != nil {
					s.OnAbort()
				}
				return QuitResult()
			}
			return Handled()
		}
	}

	switch e.Key {
	case "esc", "ctrl-c", "ctrl-d", "ctrl-q":
		s.aborted = true
		s.done = true
		if s.OnAbort != nil {
			s.OnAbort()
		}
		return QuitResult()
	case "enter":
		s.done = true
		if s.OnSubmit != nil {
			s.OnSubmit(s.Query)
		}
		return DoneResult()
	case "backspace":
		if len(s.Query) > 0 {
			runes := []rune(s.Query)
			s.SetQuery(string(runes[:len(runes)-1]))
			return Handled()
		}
		return Handled()
	default:
		if e.Text != "" {
			s.SetQuery(s.Query + e.Text)
			return Handled()
		}
		return Ignored()
	}
}

// ConsumeMouse handles mouse interactions for SearchBar.
func (s *SearchBar) ConsumeMouse(e MouseEvent) EventResult {
	if s.cmd != nil && s.cmd.handleHelpMouse(e) {
		return Handled()
	}
	if e.Action == MousePress && e.Button == MouseLeft {
		s.Focused = true
		return Handled()
	}
	return Ignored()
}
