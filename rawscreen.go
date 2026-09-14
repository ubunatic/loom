// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"io"
	"os"
	"strings"

	"codeberg.org/ubunatic/loom/measure"
	"golang.org/x/term"
)

// WriteRows writes rows to out as plain, sequential, newline-terminated
// lines, at wherever the cursor already is -- like any ordinary command's
// output. This is the right primitive for a one-shot "print once and exit"
// path: it must never touch cursor position or full-screen state, or it
// will clobber whatever the shell (prompt, prior scrollback) already put on
// screen above it.
//
// When out is a real terminal, each row is still clipped (ClipRow) to the
// terminal's live width first, so a single over-wide row can't wrap and
// misalign -- but that's the only terminal-state concern a one-shot print
// has. Anything more (hiding the cursor, disabling auto-wrap, absolute
// positioning) belongs to RawScreen, and only to a genuine redraw loop; see
// RawScreen's doc comment for why applying that to a one-shot print is
// itself a bug, not extra safety.
func WriteRows(out io.Writer, rows []string) error {
	cols := 0
	if file, ok := out.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		if c, _, err := term.GetSize(int(file.Fd())); err == nil && c > 0 {
			cols = c
		}
	}
	for _, row := range rows {
		if cols > 0 {
			row = ClipRow(row, cols)
		}
		if _, err := fmt.Fprintln(out, row); err != nil {
			return err
		}
	}
	return nil
}

// RawScreen is the "final render loop" for a persistent, full-screen redraw
// loop (a `--watch`-style command redrawing in place on an interval, like
// `watch(1)` or htop) whose content needs inline per-cell ANSI color/style
// that a loom.View would strip (View collapses every line to a single
// Style; see view.go). Renderers like graph.RenderTreemap must never embed
// terminal-state control sequences themselves -- only content and SGR
// color/style codes -- and RawScreen is where that terminal safety belongs
// instead: it is the one place responsible for keeping the terminal clean
// and capping over-wide content, so every raw-ANSI redraw loop gets it for
// free instead of re-deriving it.
//
// RawScreen always draws relative to the physical top-left of the terminal
// (row 1), taking over the full visible screen on every Draw -- the same
// convention `watch(1)`/htop use. That is deliberately NOT anchored to
// wherever the cursor happened to be when Open was called (contrast with
// Pane, which reserves a fixed region below the shell cursor and anchors
// all its drawing there -- see pane.go's startRow). Do not use RawScreen for
// a one-shot "print once and exit" path: it WILL overwrite existing
// terminal content above the current line, since positioning to row 1 has
// no idea what was there before -- use WriteRows for that instead. This
// was a real regression caught by PTY-testing the treemap example: routing
// its plain (non-watch) print through RawScreen clobbered prior shell
// prompt lines that should have stayed in scrollback.
//
// Concretely, RawScreen guarantees two things no matter what its caller's
// rows contain:
//
//  1. No row ever reaches the terminal wider than the terminal's live
//     column count. A too-wide row is silently clipped (ClipRow), not left
//     to the terminal's own auto-wrap -- auto-wrap is additionally disabled
//     as a second line of defense, but clipping is authoritative. Wrapping
//     is what turns one bad row into a cascading, whole-screen scramble:
//     the wrapped remainder eats the start of the next row's line, and an
//     explicit newline then advances *again*, so every following row lands
//     one line further out of place.
//  2. Every row is positioned with an absolute cursor move rather than
//     drawn via sequential newlines, so a bad or stale row can only corrupt
//     its own line -- the next row still lands exactly where it belongs,
//     and any error self-heals on the very next Draw instead of
//     accumulating.
//
// On a non-terminal out (redirected/piped stdout), Draw writes rows as
// plain newline-separated lines, uncapped -- there is no PTY to auto-wrap
// against, and a piped consumer may legitimately want the untruncated
// content.
type RawScreen struct {
	out    *os.File
	isTerm bool
}

// OpenRawScreen prepares out for a sequence of Draw calls: hides the cursor
// and disables terminal auto-wrap (DECAWM) when out is a real terminal.
// Both are restored by Close. Safe to use on a non-terminal out.
func OpenRawScreen(out *os.File) *RawScreen {
	s := &RawScreen{out: out, isTerm: term.IsTerminal(int(out.Fd()))}
	if s.isTerm {
		fmt.Fprint(out, "\x1b[?25l"+disableAutowrap) //nolint:errcheck
	}
	return s
}

// Draw renders one frame: rows[i] is drawn at terminal row i+1. See the
// RawScreen doc comment for the two safety guarantees this provides on a
// real terminal. Terminal width is re-queried on every call so a resize
// between frames is honored immediately.
func (s *RawScreen) Draw(rows []string) error {
	if !s.isTerm {
		for _, row := range rows {
			if _, err := fmt.Fprintln(s.out, row); err != nil {
				return err
			}
		}
		return nil
	}
	cols, _, err := term.GetSize(int(s.out.Fd()))
	if err != nil || cols < 1 {
		cols = 80
	}
	for i, row := range rows {
		if _, err := fmt.Fprintf(s.out, "\x1b[%d;1H\x1b[K%s", i+1, ClipRow(row, cols)); err != nil {
			return err
		}
	}
	// Clear any rows left over from a taller previous frame.
	_, err = fmt.Fprintf(s.out, "\x1b[%d;1H\x1b[J", len(rows)+1)
	return err
}

// Close restores auto-wrap and cursor visibility. Safe to call multiple
// times, and a no-op on a non-terminal out.
func (s *RawScreen) Close() error {
	if !s.isTerm {
		return nil
	}
	_, err := fmt.Fprint(s.out, enableAutowrap+"\x1b[?25h\n")
	return err
}

// disableAutowrap/enableAutowrap toggle the terminal's DECAWM auto-wrap mode
// (widely supported: xterm, VTE-based terminals, tmux, screen) -- RawScreen's
// second, terminal-level line of defense alongside ClipRow's software clip:
// even a row that somehow reached Draw still too wide (a bug, a race between
// the width query and the write) still can't wrap onto the next physical
// line and cascade into a whole-screen scramble.
const disableAutowrap = "\x1b[?7l"
const enableAutowrap = "\x1b[?7h"

// ClipRow hard-clips row to at most width real terminal columns, regardless
// of what width the caller's renderer (e.g. graph.RenderTreemap) was asked
// to produce -- renderers only know the width they were told, never the
// terminal's actual live width, so this is where that gets reconciled
// against reality.
//
// Escape sequences are copied through untouched (they cost zero screen
// columns) so inline color/style survives clipping -- that is the one thing
// a drawing component is allowed to emit into its rows; see RawScreen's doc
// comment. A trailing reset ("\x1b[0m") is appended whenever any styling was
// seen, so a row clipped mid-style can never bleed its color into whatever
// is drawn next.
func ClipRow(row string, width int) string {
	if width <= 0 {
		return ""
	}
	var b strings.Builder
	used, styled := 0, false
	runes := []rune(row)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == 27 && i+1 < len(runes) && runes[i+1] == '[' { // CSI sequence, e.g. "\x1b[41m"
			start := i
			for i += 2; i < len(runes) && !(runes[i] >= '@' && runes[i] <= '~'); i++ {
			}
			b.WriteString(string(runes[start : i+1]))
			styled = true
			continue
		}
		w := measure.RuneWidth(r)
		if used+w > width {
			break
		}
		b.WriteRune(r)
		used += w
	}
	if styled {
		b.WriteString("\x1b[0m")
	}
	return b.String()
}
