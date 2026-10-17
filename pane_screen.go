// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strings"
)

// Resize adjusts the pane height dynamically.
// If the height increases and overflows the terminal screen, it scrolls the terminal up
// to reserve the required space. If it decreases, it clears the abandoned lines.
func (p *Pane) Resize(newHeight int) {
	if newHeight < 1 {
		newHeight = 1
	}
	cols, termRows := termSize(p.fd)
	if cols < 1 {
		cols = 80
	}
	p.wantRows = newHeight
	if p.fullActive || p.altActive {
		return // the layout owns the rows; wantRows applies when inline again
	}
	if newHeight > termRows {
		newHeight = termRows
	}

	if newHeight == p.rows {
		return
	}

	if newHeight > p.rows {
		// Growing: check for overflow at the bottom of the screen
		overflow := (p.startRow + newHeight - 1) - termRows
		if overflow > 0 {
			// Scroll up to reserve space
			p.tty.WriteString(fmt.Sprintf("\x1b[%dS", overflow)) //nolint:errcheck
			p.startRow -= overflow
			if p.startRow < 1 {
				p.startRow = 1
			}
		}
	} else {
		// The next frame clears the abandoned rows together with its UI output.
		p.staleRows += p.rows - newHeight
	}

	p.rows = newHeight
}

// SetMaxCols sets the canvas width cap. A value of zero keeps the full
// terminal width, even when the widget requests a narrower default.
func (p *Pane) SetMaxCols(maxCols int) {
	p.MaxCols = maxCols
	p.maxColsExplicit = true
}

func (p *Pane) applyRequestedMaxCols(maxCols int) {
	if !p.maxColsExplicit && (p.MaxCols == 0 || p.MaxCols == DefaultMaxCols) {
		p.MaxCols = maxCols
	}
}

// winchBounds recomputes pane placement after a terminal window resize. Given
// the current top row and height and the new terminal row count, it clamps the
// height to fit the terminal and lifts the top row if the
// pane would overflow the new bottom. It is pure so the math can be unit-tested.
func winchBounds(startRow, rows, termRows int) (newStartRow, newRows int) {
	newRows = rows
	if newRows > termRows {
		newRows = termRows
	}
	if newRows < 1 {
		newRows = 1
	}
	newStartRow = startRow
	if overflow := (newStartRow + newRows - 1) - termRows; overflow > 0 {
		newStartRow -= overflow
	}
	if newStartRow < 1 {
		newStartRow = 1
	}
	return newStartRow, newRows
}

// applyWinch re-queries the terminal size after a SIGWINCH and updates the
// pane bounds and canvas width so the next frame reflows. Rendering and stale
// row clearing remain a single operation in the caller.
func (p *Pane) applyWinch(cols *int) {
	newCols, termRows := termSize(p.fd)
	if newCols < 1 {
		newCols = 80
	}

	// Clamp from the desired height, not the currently-clamped one, so the pane
	// grows back toward wantRows when the window is enlarged again (a tiny window
	// must not permanently pin the pane at one row).
	newStartRow, newRows := winchBounds(p.startRow, p.wantRows, termRows)
	if p.fullActive {
		newStartRow = 1
		newRows = max(1, termRows)
	}
	if p.altActive {
		newStartRow, newRows = 1, max(1, termRows)
	}

	if p.ResizeConfig.OutOfBandClear && p.tty != nil {
		top := p.startRow
		if newStartRow < top {
			top = newStartRow
		}
		var b strings.Builder
		for row := top; row <= termRows; row++ {
			fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", row)
		}
		p.tty.WriteString(b.String()) //nolint:errcheck
	}

	p.startRow, p.rows = newStartRow, newRows
	p.cols = newCols

	*cols = p.guardedCols(newCols)
}

// refreshStartupSize closes the gap between New's size read and signal
// registration, including time spent probing the terminal. No frame exists
// yet, so updating inline bounds needs no erasure or scrollback changes.
func (p *Pane) refreshStartupSize() {
	cols, rows := termSize(p.fd)
	wantRows := p.wantRows
	if wantRows < 1 {
		wantRows = p.rows
	}
	p.startRow, p.rows = winchBounds(p.startRow, wantRows, rows)
	if p.fullActive {
		p.startRow, p.rows = 1, max(1, rows)
	}
	if p.altActive {
		p.startRow, p.rows = 1, max(1, rows)
	}
	p.cols = cols
}

// ScreenMode is where the pane draws: a few rows below the prompt, the whole
// terminal, or the whole alternate screen.
type ScreenMode int

const (
	ScreenInline ScreenMode = iota // reserved rows below the prompt
	ScreenFull                     // full terminal from row 1 on the primary screen
	ScreenAlt                      // full terminal on the alternate screen
)

// SetScreenMode selects the layout by setting the resize modes it consists of.
// The next frame performs the switch. Auto full screen, when enabled, can still
// promote an inline pane; disable it in ResizeConfig to force ScreenInline.
func (p *Pane) SetScreenMode(m ScreenMode) {
	p.ResizeConfig.FullScreenBuffer = m == ScreenFull
	p.ResizeConfig.AltScreen = m == ScreenAlt
}

// Screen reports the layout currently in force, including an automatic
// promotion to full screen.
func (p *Pane) Screen() ScreenMode {
	switch {
	case p.altActive:
		return ScreenAlt
	case p.fullActive:
		return ScreenFull
	}
	return ScreenInline
}

// AutoFullscreenReasons reports which detectors (width, height) currently ask
// for the full-screen layout. Both are false when detection is off or the pane
// is InlineOnly. It is for status displays; the pane acts on it by itself.
func (p *Pane) AutoFullscreenReasons() (byWidth, byHeight bool) {
	cfg := p.ResizeConfig
	if !cfg.AutoFullscreen || p.InlineOnly {
		return false, false
	}
	termCols, termRows := termSize(p.fd)
	w, h := cfg, cfg
	w.FullByHeight, h.FullByWidth = false, false
	return w.QuasiFullscreen(p.MaxCols, p.wantRows, termCols, termRows),
		h.QuasiFullscreen(p.MaxCols, p.wantRows, termCols, termRows)
}

// wantScreen derives the layout the config asks for right now. full applies to
// the primary screen only; it is ignored while alt is wanted. Auto full screen
// compares the wanted size (not the clamped one) with the terminal size.
func (p *Pane) wantScreen() (full, alt bool) {
	cfg := p.ResizeConfig
	byWidth, byHeight := p.AutoFullscreenReasons()
	auto := byWidth || byHeight
	return cfg.FullScreenBuffer || auto, cfg.AltScreen || (auto && cfg.FullAlt)
}

// changeScreen moves between layouts. Leaving alt first restores the primary
// bounds, the primary full/inline change happens on the primary screen, and
// entering alt comes last so it saves the final primary bounds.
func (p *Pane) changeScreen(full, alt bool) {
	if p.altActive && !alt {
		p.switchAltScreen(false)
	}
	if !alt && full != p.fullActive {
		p.switchFull(full)
	}
	if alt && !p.altActive {
		p.switchAltScreen(true)
	}
}

// switchFull changes the primary-screen layout between inline and full screen.
// The old frame is erased row by row; the next frame paints the new one.
func (p *Pane) switchFull(on bool) {
	if p.tty != nil {
		var b strings.Builder
		for i := 0; i < p.rows+p.staleRows; i++ {
			fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", p.startRow+i)
		}
		p.tty.WriteString(b.String()) //nolint:errcheck
	}
	p.staleRows = 0
	if on {
		p.inlineStart = p.startRow
	} else {
		p.startRow = p.inlineStart
	}
	p.fullActive = on
}

// switchAltScreen enters or leaves the alternate screen buffer. The primary
// screen's pane bounds are kept and restored on leave.
func (p *Pane) switchAltScreen(on bool) {
	if p.tty == nil || on == p.altActive {
		return
	}
	if p.kittyKeys {
		p.tty.WriteString(kittyKeyboardPop)        //nolint:errcheck
		defer p.tty.WriteString(kittyKeyboardPush) //nolint:errcheck
	}
	if on {
		p.savedStart, p.savedRows = p.startRow, p.rows
		// Erase the frame already drawn on the primary screen, or it stays in
		// the scrollback after leaving; park the cursor where the pane started
		// so leaving restores it there.
		var b strings.Builder
		for i := 0; i < p.rows+p.staleRows; i++ {
			fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", p.startRow+i)
		}
		p.staleRows = 0
		fmt.Fprintf(&b, "\x1b[%d;1H\x1b[?1049h\x1b[2J", p.startRow)
		p.tty.WriteString(b.String()) //nolint:errcheck
	} else {
		p.tty.WriteString("\x1b[?1049l") //nolint:errcheck
		p.startRow, p.rows = p.savedStart, p.savedRows
		if p.ResizeConfig.FullLeakGuard {
			// The terminal may have reflowed or resized the primary screen while
			// the alt screen was up; start from clean pane rows.
			var b strings.Builder
			for i := 0; i < p.rows; i++ {
				fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", p.startRow+i)
			}
			p.tty.WriteString(b.String()) //nolint:errcheck
		}
	}
	p.altActive = on
}
