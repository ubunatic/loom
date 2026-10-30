package editor

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
)

// Issue 287 M3: a search highlight is display-only and must not move the
// editing cursor.

// Expected red before M3: the search highlight assigned edit.Cursor.
func Test287SearchHighlightKeepsEditingCursor(t *testing.T) {
	view, _ := new287View(t, "alpha beta alpha\n")
	view.edit.Cursor = loom.RichPosition{Offset: 4}
	before := snap287(view)

	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	for _, r := range "alpha" {
		view.ConsumeKey(loom.KeyEvent{Text: string(r)})
		requireSame287(t, fmt.Sprintf("typing %q into the query", string(r)), before, view, true, true)
	}
	if len(view.searchMatches) != 2 || len(view.edit.Highlights) != 1 {
		t.Fatalf("setup: matches=%d highlights=%d", len(view.searchMatches), len(view.edit.Highlights))
	}
	for _, key := range []string{"enter", "enter", "shift-enter"} {
		view.ConsumeKey(loom.KeyEvent{Key: key})
		requireSame287(t, key+" in search", before, view, true, true)
	}
}

// Expected red before M3: the match is reachable only through the cursor, so
// the cursor assertion fails today. After M3 it also guards the reveal.
func Test287SearchRevealsOffscreenMatchWithoutMovingCursor(t *testing.T) {
	var doc strings.Builder
	for i := 0; i < 40; i++ {
		if i == 35 {
			doc.WriteString("the needle is here\n")
			continue
		}
		fmt.Fprintf(&doc, "line %d\n", i)
	}
	view, _ := new287View(t, doc.String())
	draw287(view, 100, 24)
	before := snap287(view)

	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	type287(view, "needle")
	draw287(view, 100, 24)

	if len(view.searchMatches) != 1 || view.searchMatches[0].line != 35 {
		t.Fatalf("setup: matches = %+v", view.searchMatches)
	}
	if got := view.edit.Cursor; got != before.cursor {
		t.Fatalf("search moved the editing cursor: %+v, want %+v", got, before.cursor)
	}
	if line, top, rows := 35, view.edit.ScrollY, view.editorRect.H; line < top || line >= top+rows {
		t.Fatalf("match line %d not visible: ScrollY=%d rows=%d", line, top, rows)
	}
	requireSame287(t, "searching", before, view, true, false)
}

// Expected red before M3: the status line follows the match through the cursor.
func Test287EditPTYSearchHighlightKeepsCursorStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s, _ := startMouseEditPTY(t, "alpha beta alpha\n")
	timeout := 3 * time.Second * ptytest.RaceFactor
	cursorPattern := regexp.MustCompile(`Ln [0-9]+, Col [0-9]+`)
	status := func(rows []string) string { return cursorPattern.FindString(strings.Join(rows, "\n")) }

	// The status row is repainted after the query row, so a snapshot taken as
	// soon as the query shows can still hold last frame's cursor. A help round
	// trip repaints everything first; wait for the status to be back.
	settledStatus := func() string {
		syncHelpPTY(t, s)
		var got string
		s.WaitForState(timeout, func(rows []string) bool { got = status(rows); return got != "" })
		return got
	}

	s.Send("\x06alpha")
	s.WaitForUpdateState(timeout, func(rows []string) bool {
		text := strings.Join(rows, "\n")
		return strings.Contains(text, "Find: alpha") && strings.Contains(text, "1 / 2 matches")
	})
	if got := settledStatus(); got != "Ln 1, Col 1" {
		t.Fatalf("typing the query moved the cursor: %q", got)
	}
	s.Send("\r")
	s.WaitForUpdateState(timeout, func(rows []string) bool {
		return strings.Contains(strings.Join(rows, "\n"), "2 / 2 matches")
	})
	if got := settledStatus(); got != "Ln 1, Col 1" {
		t.Fatalf("next match moved the cursor: %q", got)
	}
}
