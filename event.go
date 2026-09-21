// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// KeyEvent is a decoded keyboard event.
// Key names match common terminal conventions; Text carries printable input.
type KeyEvent struct {
	Key string // "up","down","left","right","home","end","delete","enter","esc",
	//            "backspace","tab","shift-tab","shift-up","shift-down","shift-left","shift-right",
	//            "ctrl-up","ctrl-down","ctrl-left","ctrl-right","alt-up","alt-down",
	//            "f1","f2","f3","f4","f5","f6","f7","f8","f9","f10","f11","f12",
	//            "ctrl-b","ctrl-c","ctrl-d","ctrl-f","ctrl-q","ctrl-u","ctrl-w",
	//            or "" for plain text
	Text string // typed printable text (Key == "" when Text != "")
}

// Name returns the special-key name or printable text carried by the event.
func (e KeyEvent) Name() string {
	if e.Key != "" {
		return e.Key
	}
	return e.Text
}

// Is reports whether the event name matches one of the provided key identifiers.
func (e KeyEvent) Is(keys ...string) bool {
	name := e.Name()
	for _, key := range keys {
		if name == key {
			return true
		}
	}
	return false
}

// Rune returns the first printable rune in the event text, or zero otherwise.
func (e KeyEvent) Rune() rune {
	for _, r := range e.Text {
		if unicode.IsPrint(r) {
			return r
		}
		return 0
	}
	return 0
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
	case b[0] >= 1 && b[0] <= 26 && b[0] != 2 && b[0] != 3 && b[0] != 4 && b[0] != 6 && b[0] != 8 && b[0] != 9 && b[0] != 10 && b[0] != 13 && b[0] != 17 && b[0] != 21 && b[0] != 23:
		return KeyEvent{Key: "ctrl-" + string(rune('a'+b[0]-1))}
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
		// CSI 1;<mod><A|B|C|D> (xterm modified cursor keys)
		// mod: 2=Shift, 3=Alt, 4=Shift+Alt, 5=Ctrl, 6=Ctrl+Shift
		if len(b) >= 6 && b[1] == '[' && b[2] == '1' && b[3] == ';' {
			var prefix string
			switch b[4] {
			case '2':
				prefix = "shift-"
			case '3':
				prefix = "alt-"
			case '4':
				prefix = "shift-alt-"
			case '5':
				prefix = "ctrl-"
			case '6':
				prefix = "ctrl-shift-"
			case '7':
				prefix = "ctrl-alt-"
			case '8':
				prefix = "ctrl-shift-alt-"
			}
			if prefix != "" {
				switch b[5] {
				case 'A':
					return KeyEvent{Key: prefix + "up"}
				case 'B':
					return KeyEvent{Key: prefix + "down"}
				case 'C':
					return KeyEvent{Key: prefix + "right"}
				case 'D':
					return KeyEvent{Key: prefix + "left"}
				}
			}
		}
		// \x1b[ (CSI) and \x1bO (application cursor) share the same final byte.
		if b[1] != '[' && b[1] != 'O' {
			return KeyEvent{Key: "alt-" + string(b[1:])}
		}
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
			case 'a':
				return KeyEvent{Key: "shift-up"}
			case 'b':
				return KeyEvent{Key: "shift-down"}
			case 'c':
				return KeyEvent{Key: "shift-right"}
			case 'd':
				return KeyEvent{Key: "shift-left"}
			case 'P':
				return KeyEvent{Key: "f1"}
			case 'Q':
				return KeyEvent{Key: "f2"}
			case 'R':
				return KeyEvent{Key: "f3"}
			case 'S':
				return KeyEvent{Key: "f4"}
			case 'H':
				return KeyEvent{Key: "home"}
			case 'F':
				return KeyEvent{Key: "end"}
			}
			// Numeric tilde sequences: \x1b[1~ home, \x1b[3~ delete, \x1b[4~ end,
			// \x1b[5~ pgup, \x1b[6~ pgdown, \x1b[7~ home, \x1b[8~ end,
			// and function keys \x1b[15~ .. \x1b[24~ (f5..f12).
			if len(b) >= 4 && b[len(b)-1] == '~' {
				seq := string(b[2 : len(b)-1])
				if strings.Contains(seq, ";") {
					seq = strings.SplitN(seq, ";", 2)[0]
				}
				switch seq {
				case "1", "7":
					return KeyEvent{Key: "home"}
				case "3":
					return KeyEvent{Key: "delete"}
				case "2":
					return KeyEvent{Key: "insert"}
				case "4", "8":
					return KeyEvent{Key: "end"}
				case "5":
					return KeyEvent{Key: "pgup"}
				case "6":
					return KeyEvent{Key: "pgdown"}
				case "11":
					return KeyEvent{Key: "f1"}
				case "12":
					return KeyEvent{Key: "f2"}
				case "13":
					return KeyEvent{Key: "f3"}
				case "14":
					return KeyEvent{Key: "f4"}
				case "15":
					return KeyEvent{Key: "f5"}
				case "17":
					return KeyEvent{Key: "f6"}
				case "18":
					return KeyEvent{Key: "f7"}
				case "19":
					return KeyEvent{Key: "f8"}
				case "20":
					return KeyEvent{Key: "f9"}
				case "21":
					return KeyEvent{Key: "f10"}
				case "23":
					return KeyEvent{Key: "f11"}
				case "24":
					return KeyEvent{Key: "f12"}
				}
			}
		}
		return KeyEvent{} // unhandled escape sequence
	default:
		if utf8.Valid(b) {
			for _, r := range string(b) {
				if !unicode.IsPrint(r) {
					return KeyEvent{}
				}
			}
			return KeyEvent{Text: string(b)}
		}
		return KeyEvent{}
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

// scanKey extracts the leading key event from b, the non-mouse counterpart to
// scanMouse: a single tty read (or several coalesced reads) can carry more
// than one complete key sequence, and a fast terminal can split one escape
// sequence across two reads. ok is false when b ends with bytes that could
// still be the unfinished prefix of a longer escape sequence (e.g. a lone
// "\x1b" or "\x1b[1") — the caller should hold those bytes as pending and
// wait for more data (or time out and treat a standalone ESC as a plain
// "esc" keypress; see Pane.run) instead of dispatching a wrong decode. When
// ok is true, used is the number of bytes consumed for the returned event.
func scanKey(b []byte) (KeyEvent, int, bool) {
	if len(b) == 0 {
		return KeyEvent{}, 0, true
	}
	switch b[0] {
	case 3, 2, 4, 6, 9, 10, 13, 17, 21, 23, 127, 8:
		return DecodeKey(b[:1]), 1, true
	case 27:
		return scanEscapeKey(b)
	default:
		// A run of plain text/UTF-8 bytes up to the next control byte or ESC,
		// decoded together as DecodeKey's default branch already does.
		end := 1
		for end < len(b) && !isKeyControlByte(b[end]) {
			end++
		}
		return DecodeKey(b[:end]), end, true
	}
}

// isKeyControlByte reports whether c is one of the single-byte control keys
// or ESC that scanKey/DecodeKey special-case, i.e. a byte that must not be
// folded into a plain-text run.
func isKeyControlByte(c byte) bool {
	switch c {
	case 3, 2, 4, 6, 9, 10, 13, 17, 21, 23, 127, 8, 27:
		return true
	default:
		return false
	}
}

// scanEscapeKey handles the b[0] == 27 (ESC) case for scanKey, mirroring the
// escape-sequence forms DecodeKey recognizes so it can report how many bytes
// each one consumes (or that more bytes are needed to tell).
func scanEscapeKey(b []byte) (KeyEvent, int, bool) {
	if len(b) == 1 {
		return KeyEvent{}, 0, false // could be a standalone ESC, or the start of a sequence
	}
	if b[1] != '[' && b[1] != 'O' {
		if b[1] >= 0x80 {
			if !utf8.FullRune(b[1:]) {
				return KeyEvent{}, 0, false
			}
			_, n := utf8.DecodeRune(b[1:])
			return DecodeKey(b[:1+n]), 1 + n, true
		}
		return DecodeKey(b[:2]), 2, true
	}
	if len(b) == 2 {
		return KeyEvent{}, 0, false
	}
	if b[1] == '[' && b[2] == '1' {
		if len(b) == 3 {
			return KeyEvent{}, 0, false
		}
		if b[3] == ';' {
			// Modified cursor: ESC [ 1 ; <mod> <letter>
			if len(b) < 6 {
				return KeyEvent{}, 0, false
			}
			return DecodeKey(b[:6]), 6, true
		}
		// Not modified-cursor; falls through to the tilde-number scan below
		// (e.g. "\x1b[15~").
	}
	switch b[2] {
	case 'Z', 'A', 'B', 'C', 'D', 'a', 'b', 'c', 'd', 'P', 'Q', 'R', 'S', 'H', 'F':
		return DecodeKey(b[:3]), 3, true
	}
	if b[2] >= '0' && b[2] <= '9' {
		i := 2
		for i < len(b) && b[i] >= '0' && b[i] <= '9' {
			i++
		}
		if i == len(b) {
			return KeyEvent{}, 0, false // digits ran out; a '~' may still follow
		}
		if b[i] == '~' {
			return DecodeKey(b[:i+1]), i + 1, true
		}
		// Not a tilde sequence after all; consume what was scanned so an
		// unrecognized form cannot stall the loop.
		return KeyEvent{}, i, true
	}
	// Unrecognized CSI/SS3 form; consume the 3 bytes seen so far.
	return KeyEvent{}, 3, true
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
