// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

// ── command-mode activation ───────────────────────────────────────────────────

func TestCmdColonActivates(t *testing.T) {
	c := loom.NewChoice([]loom.Item{{Name: "git"}, {Name: "make"}})
	quit := c.HandleKey(loom.KeyEvent{Text: ":"})
	if quit {
		t.Error("':' should not quit")
	}
	// Filter must not be applied — both items still visible.
	if c.FilteredItem(1).Name != "make" {
		t.Error("':' should activate command mode, not filter items")
	}
}

func TestCmdSlashActivates(t *testing.T) {
	c := loom.NewChoice([]loom.Item{{Name: "git"}, {Name: "make"}})
	c.HandleKey(loom.KeyEvent{Text: "/"})
	if c.FilteredItem(1).Name != "make" {
		t.Error("'/' should activate command mode, not filter items")
	}
}

// ── Esc and backspace deactivation ───────────────────────────────────────────

func TestCmdEscDeactivatesWithoutQuitting(t *testing.T) {
	c := loom.NewChoice([]loom.Item{{Name: "x"}})
	c.HandleKey(loom.KeyEvent{Text: ":"})    // activate
	quit := c.HandleKey(loom.KeyEvent{Key: "esc"}) // deactivate
	if quit {
		t.Error("Esc in command mode should deactivate, not quit")
	}
	// Second Esc (now in normal mode) should quit.
	quit = c.HandleKey(loom.KeyEvent{Key: "esc"})
	if !quit {
		t.Error("Esc in normal mode should quit")
	}
}

func TestCmdBackspaceToEmptyDeactivates(t *testing.T) {
	c := loom.NewChoice([]loom.Item{{Name: "x"}})
	c.HandleKey(loom.KeyEvent{Text: ":"})
	c.HandleKey(loom.KeyEvent{Text: "h"})                 // query = "h"
	c.HandleKey(loom.KeyEvent{Key: "backspace"})           // query = ""
	quit := c.HandleKey(loom.KeyEvent{Key: "backspace"})   // deactivates
	if quit {
		t.Error("backspace to empty should deactivate command mode, not quit")
	}
	// Normal typing should filter now, not accumulate command query.
	c.HandleKey(loom.KeyEvent{Text: "x"})
	if c.FilteredItem(0).Name != "x" {
		t.Error("after deactivation, typing should filter normally")
	}
}

// ── :back ─────────────────────────────────────────────────────────────────────

func TestCmdBackAbortsWidget(t *testing.T) {
	c := loom.NewChoice([]loom.Item{{Name: "x"}})
	typeCmd(c, "back")
	quit := c.HandleKey(loom.KeyEvent{Key: "enter"})
	if !quit {
		t.Error(":back enter should quit")
	}
	if !c.Aborted() {
		t.Error(":back should abort the widget")
	}
	if c.Nav() != loom.NavNone {
		t.Error(":back should not set NavHome")
	}
}

func TestCmdBackTabCompletes(t *testing.T) {
	c := loom.NewChoice([]loom.Item{{Name: "x"}})
	c.HandleKey(loom.KeyEvent{Text: ":"})
	c.HandleKey(loom.KeyEvent{Text: "b"})                // "b" matches "back"
	quit := c.HandleKey(loom.KeyEvent{Key: "tab"})       // complete + execute
	if !quit {
		t.Error(":b tab should complete to :back and quit")
	}
	if !c.Aborted() {
		t.Error(":b tab should abort the widget")
	}
}

// ── :home ─────────────────────────────────────────────────────────────────────

func TestCmdHomeSetsNavHome(t *testing.T) {
	c := loom.NewChoice([]loom.Item{{Name: "x"}})
	typeCmd(c, "home")
	quit := c.HandleKey(loom.KeyEvent{Key: "enter"})
	if !quit {
		t.Error(":home enter should quit")
	}
	if c.Nav() != loom.NavHome {
		t.Errorf("Nav() = %v, want NavHome", c.Nav())
	}
}

// ── :help ─────────────────────────────────────────────────────────────────────

func TestCmdHelpDoesNotQuitInTest(t *testing.T) {
	// In a test environment there is no /dev/tty, so showHelp() silently fails
	// and returns without quitting. Nav stays NavNone.
	c := loom.NewChoice([]loom.Item{{Name: "x"}})
	typeCmd(c, "help")
	quit := c.HandleKey(loom.KeyEvent{Key: "enter"})
	if quit {
		t.Error(":help should not quit the parent widget")
	}
	if c.Nav() != loom.NavNone {
		t.Errorf(":help Nav() = %v, want NavNone", c.Nav())
	}
}

// ── prefix matching ───────────────────────────────────────────────────────────

func TestCmdPrefixMatchEnter(t *testing.T) {
	c := loom.NewChoice([]loom.Item{{Name: "x"}})
	c.HandleKey(loom.KeyEvent{Text: ":"})
	c.HandleKey(loom.KeyEvent{Text: "b"}) // matches "back"
	c.HandleKey(loom.KeyEvent{Text: "a"})
	c.HandleKey(loom.KeyEvent{Text: "c"})
	c.HandleKey(loom.KeyEvent{Text: "k"}) // full "back"
	quit := c.HandleKey(loom.KeyEvent{Key: "enter"})
	if !quit || !c.Aborted() {
		t.Error(":back full match enter should abort")
	}
}

func TestCmdNoMatchEnterDeactivates(t *testing.T) {
	c := loom.NewChoice([]loom.Item{{Name: "git"}})
	c.HandleKey(loom.KeyEvent{Text: ":"})
	c.HandleKey(loom.KeyEvent{Text: "z"}) // no command starts with "z"
	quit := c.HandleKey(loom.KeyEvent{Key: "enter"})
	if quit {
		t.Error("unmatched enter should deactivate but not quit")
	}
	// After deactivation, normal filter should work.
	c.HandleKey(loom.KeyEvent{Text: "g"})
	if c.FilteredItem(0).Name != "git" {
		t.Error("after unmatched cmd deactivation, filter should work")
	}
}

// ── filter unchanged during command mode ──────────────────────────────────────

func TestCmdModeDoesNotFilter(t *testing.T) {
	items := []loom.Item{{Name: "git"}, {Name: "make"}}
	c := loom.NewChoice(items)
	c.HandleKey(loom.KeyEvent{Text: ":"})
	c.HandleKey(loom.KeyEvent{Text: "g"}) // "g" is cmd query, not filter
	if c.FilteredItem(0).Name != "git" || c.FilteredItem(1).Name != "make" {
		t.Error("in command mode, typing should not filter the item list")
	}
}

// ── Table command mode ────────────────────────────────────────────────────────

func TestTableCmdHomeViaColon(t *testing.T) {
	tbl, _, _ := psTestTable()
	typeTableCmd(tbl, "home")
	quit := tbl.HandleKey(loom.KeyEvent{Key: "enter"})
	if !quit {
		t.Error(":home enter should quit the table")
	}
	if tbl.Nav() != loom.NavHome {
		t.Errorf("tbl.Nav() = %v, want NavHome", tbl.Nav())
	}
}

func TestTableCmdModeBlocksSortTab(t *testing.T) {
	tbl, _, _ := psTestTable()
	tbl.HandleKey(loom.KeyEvent{Text: ":"}) // activate command mode
	// Tab in command mode should try to complete a command, not cycle sort.
	sortColBefore := tbl.SortCol
	tbl.HandleKey(loom.KeyEvent{Key: "tab"}) // completes "help" (no tty → no-op)
	if tbl.SortCol != sortColBefore {
		t.Error("Tab in command mode should not cycle the sort column")
	}
}

// ── AddCmd custom command ─────────────────────────────────────────────────────

func TestCmdAddLocalCommand(t *testing.T) {
	called := false
	c := loom.NewChoice([]loom.Item{{Name: "x"}})
	c.AddCmd(loom.Cmd{
		Name:  "refresh",
		Title: "reload data",
		Fn:    func() error { called = true; return nil },
	})
	typeCmd(c, "refresh")
	c.HandleKey(loom.KeyEvent{Key: "enter"})
	if !called {
		t.Error("custom :refresh command Fn should be called")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// typeCmd activates command mode and types name into c.
func typeCmd(c *loom.Choice, name string) {
	c.HandleKey(loom.KeyEvent{Text: ":"})
	for _, ch := range name {
		c.HandleKey(loom.KeyEvent{Text: string(ch)})
	}
}

// typeTableCmd activates command mode and types name into t.
func typeTableCmd(t *loom.Table, name string) {
	t.HandleKey(loom.KeyEvent{Text: ":"})
	for _, ch := range name {
		t.HandleKey(loom.KeyEvent{Text: string(ch)})
	}
}
