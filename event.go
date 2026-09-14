// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// KeyEvent is a decoded keyboard event.
// Key names match common terminal conventions; Text carries printable input.
type KeyEvent struct {
	Key string // "up","down","left","right","home","end","delete","enter","esc",
	//            "backspace","tab","shift-tab","ctrl-b","ctrl-c","ctrl-d",
	//            "ctrl-f","ctrl-q","ctrl-u","ctrl-w",
	//            or "" for plain text
	Text string // typed printable text (Key == "" when Text != "")
}

// MouseAction classifies a mouse event.
type MouseAction int

const (
	MousePress      MouseAction = iota
	MouseRelease                // button released
	MouseHover                  // motion with no button held
	MouseDrag                   // motion with button held
	MouseScrollUp               // wheel up
	MouseScrollDown             // wheel down
)

// MouseButton identifies which button triggered the event.
type MouseButton int

const (
	MouseLeft   MouseButton = iota // left / primary
	MouseMiddle                    // middle / scroll-click
	MouseRight                     // right / secondary
	MouseNone                      // hover or scroll (no button)
)

// MouseEvent is a decoded SGR (1006) mouse report.
type MouseEvent struct {
	Action MouseAction
	Button MouseButton
	X, Y   int // 1-based terminal column and row
}

// DecodeKey converts raw bytes from a /dev/tty read into a KeyEvent.
// It handles both normal cursor keys (\x1b[A) and application cursor keys
// (\x1bOA) because ZSH ZLE enables DECCKM and zle -I does not reset it.
// See docs/TuiInput.md §3.
func DecodeKey(b []byte) KeyEvent {
	if len(b) == 0 {
		return KeyEvent{}
	}
	switch {
	case b[0] == 3:
		return KeyEvent{Key: "ctrl-c"}
	case b[0] == 2:
		return KeyEvent{Key: "ctrl-b"}
	case b[0] == 4:
		return KeyEvent{Key: "ctrl-d"}
	case b[0] == 6:
		return KeyEvent{Key: "ctrl-f"}
	case b[0] == 9:
		return KeyEvent{Key: "tab"}
	case b[0] == 10 || b[0] == 13:
		return KeyEvent{Key: "enter"}
	case b[0] == 17:
		return KeyEvent{Key: "ctrl-q"}
	case b[0] == 21:
		return KeyEvent{Key: "ctrl-u"}
	case b[0] == 23:
		return KeyEvent{Key: "ctrl-w"}
	case b[0] == 127 || b[0] == 8:
		return KeyEvent{Key: "backspace"}
	case b[0] == 27:
		if len(b) == 1 {
			return KeyEvent{Key: "esc"}
		}
		// \x1b[ (CSI) and \x1bO (application cursor) share the same final byte.
		if len(b) >= 3 && (b[1] == '[' || b[1] == 'O') {
			switch b[2] {
			case 'Z':
				return KeyEvent{Key: "shift-tab"}
			case 'A':
				return KeyEvent{Key: "up"}
			case 'B':
				return KeyEvent{Key: "down"}
			case 'C':
				return KeyEvent{Key: "right"}
			case 'D':
				return KeyEvent{Key: "left"}
			case 'H':
				return KeyEvent{Key: "home"}
			case 'F':
				return KeyEvent{Key: "end"}
			}
			// Numeric tilde sequences: \x1b[1~ home, \x1b[3~ delete, \x1b[4~ end,
			// \x1b[5~ pgup, \x1b[6~ pgdown, \x1b[7~ home, \x1b[8~ end (xterm/linux console variants).
			if len(b) >= 4 && b[len(b)-1] == '~' {
				switch b[2] {
				case '1', '7':
					return KeyEvent{Key: "home"}
				case '3':
					return KeyEvent{Key: "delete"}
				case '4', '8':
					return KeyEvent{Key: "end"}
				case '5':
					return KeyEvent{Key: "pgup"}
				case '6':
					return KeyEvent{Key: "pgdown"}
				}
			}
		}
		return KeyEvent{} // unhandled escape sequence
	default:
		out := make([]byte, 0, len(b))
		for _, c := range b {
			if c >= 32 && c <= 126 {
				out = append(out, c)
			}
		}
		return KeyEvent{Text: string(out)}
	}
}

// DecodeMouse parses an SGR (\x1b[<…M or \x1b[<…m) mouse report.
// Returns (event, true) on success, (zero, false) if b is not an SGR mouse report.
func DecodeMouse(b []byte) (MouseEvent, bool) {
	// SGR format: \x1b [ < Cb ; Cx ; Cy M|m
	if len(b) < 6 || b[0] != 27 || b[1] != '[' || b[2] != '<' {
		return MouseEvent{}, false
	}
	s := string(b[3:])
	term := s[len(s)-1]
	if term != 'M' && term != 'm' {
		return MouseEvent{}, false
	}
	s = s[:len(s)-1]
	var cb, cx, cy int
	if n, _ := parseInts(s, &cb, &cx, &cy); n != 3 {
		return MouseEvent{}, false
	}
	btn := cb & 3
	motion := cb&32 != 0
	scroll := cb&64 != 0

	var action MouseAction
	var button MouseButton
	switch {
	case scroll && btn == 0:
		return MouseEvent{Action: MouseScrollUp, Button: MouseNone, X: cx, Y: cy}, true
	case scroll && btn == 1:
		return MouseEvent{Action: MouseScrollDown, Button: MouseNone, X: cx, Y: cy}, true
	case motion:
		if btn == 3 {
			action = MouseHover
			button = MouseNone
		} else {
			action = MouseDrag
			button = MouseButton(btn)
		}
	case term == 'M':
		action = MousePress
		button = MouseButton(btn)
	default:
		action = MouseRelease
		button = MouseButton(btn)
	}
	return MouseEvent{Action: action, Button: button, X: cx, Y: cy}, true
}

// scanMouse extracts the leading SGR mouse report from b, returning the decoded
// event and the number of bytes consumed. ok is false if b does not start with a
// complete SGR mouse report. It exists to drain a buffer holding several reports:
// \x1b[?1003h any-motion tracking floods multiple reports per read, so a single
// os.File.Read often returns "\x1b[<…M\x1b[<…M\x1b[<…M". Callers loop on the
// remaining bytes to dispatch each report in order.
func scanMouse(b []byte) (MouseEvent, int, bool) {
	if len(b) < 6 || b[0] != 27 || b[1] != '[' || b[2] != '<' {
		return MouseEvent{}, 0, false
	}
	end := -1
	for i := 3; i < len(b); i++ {
		if b[i] == 'M' || b[i] == 'm' {
			end = i
			break
		}
	}
	if end < 0 {
		return MouseEvent{}, 0, false
	}
	ev, ok := DecodeMouse(b[:end+1])
	if !ok {
		return MouseEvent{}, 0, false
	}
	return ev, end + 1, true
}

// parseInts parses semicolon-separated ints into dst, returning how many were filled.
func parseInts(s string, dst ...*int) (int, error) {
	n := 0
	for _, p := range dst {
		semi := len(s)
		for i, c := range s {
			if c == ';' {
				semi = i
				break
			}
		}
		v := 0
		for _, c := range s[:semi] {
			if c < '0' || c > '9' {
				return n, nil
			}
			v = v*10 + int(c-'0')
		}
		*p = v
		n++
		if semi == len(s) {
			break
		}
		s = s[semi+1:]
	}
	return n, nil
}
