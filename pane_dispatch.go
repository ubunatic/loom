// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
)

func (p *Pane) dispatchKey(root Widget, ke KeyEvent) bool {
	if p.globalF10Quit(ke) {
		return p.requestClose(CloseReasonQuitKey)
	}
	if ke.Is("shift-f12") {
		p.DebugMode = !p.DebugMode
		return false
	}
	res := DispatchKeyEvent(root, ke)
	if res.Consumed {
		return res.Quit
	}
	if reason, ok := p.fallbackClose(ke); ok {
		return p.requestClose(reason)
	}
	return false
}

func (p *Pane) globalF10Quit(ke KeyEvent) bool {
	return !p.DisableGlobalF10Quit && strings.EqualFold(ke.Name(), "f10")
}

func (p *Pane) dispatchMouse(root Widget, me MouseEvent) EventResult {
	if p.help != nil && p.help.Open {
		root = p.help
	}
	if result, captured := p.modalMouse.Dispatch(root, me); captured {
		if p.help != nil && !p.help.Open {
			p.help = nil
		}
		return result
	}
	return DispatchMouseEvent(root, me)
}

func (p *Pane) handleHelpKey(e KeyEvent) (quit, handled bool) {
	if p.globalF10Quit(e) {
		return p.requestClose(CloseReasonQuitKey), true
	}
	if p.help == nil {
		return false, false
	}
	if e.Key == "esc" {
		p.help = nil
		return false, true
	}
	result := p.help.ConsumeKey(e)
	if result.Quit || !p.help.Open {
		p.help = nil
	}
	return false, true
}

var defaultQuitKeyMap = func() map[string]bool {
	m := make(map[string]bool, len(SpeccedDefaults.QuitKeys))
	for _, k := range SpeccedDefaults.QuitKeys {
		m[strings.ToLower(k)] = true
	}
	return m
}()

// fallbackClose reports whether an unhandled key asks the pane to close and
// why: the spec quit_keys, Esc when escape_quits or EscapeQuits is set, and
// Ctrl-C as an interrupt. It does not call OnCloseRequest.
func (p *Pane) fallbackClose(ke KeyEvent) (CloseReason, bool) {
	if p.DisableDefaultQuit {
		return 0, false
	}
	name := strings.ToLower(ke.Name())
	switch {
	case name == "ctrl-c":
		return CloseReasonInterrupt, true
	case defaultQuitKeyMap[name]:
		return CloseReasonQuitKey, true
	case name == "esc" && (p.EscapeQuits || SpeccedDefaults.EscapeQuits):
		return CloseReasonQuitKey, true
	}
	return 0, false
}

// handleKeyFallback reports whether an unhandled key should close the pane
// after OnCloseRequest has been consulted.
func (p *Pane) handleKeyFallback(ke KeyEvent, roots ...Widget) bool {
	if len(roots) > 0 {
		if res := DispatchKeyEvent(roots[0], ke); res.Consumed {
			return false
		}
	}
	reason, ok := p.fallbackClose(ke)
	return ok && p.requestClose(reason)
}
