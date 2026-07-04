// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestConfirmDefaultYesEnter(t *testing.T) {
	c := loom.NewConfirm("Commit?")
	if c.Answered() {
		t.Fatal("fresh Confirm should not be answered")
	}
	if quit := c.HandleKey(loom.KeyEvent{Key: "enter"}); !quit {
		t.Fatal("Enter should quit")
	}
	if !c.Answered() || !c.Confirmed() {
		t.Errorf("default Enter: answered=%v confirmed=%v, want true/true",
			c.Answered(), c.Confirmed())
	}
}

func TestConfirmDirectNoKey(t *testing.T) {
	c := loom.NewConfirm("Commit?")
	if quit := c.HandleKey(loom.KeyEvent{Text: "n"}); !quit {
		t.Fatal("'n' should answer and quit")
	}
	if !c.Answered() || c.Confirmed() {
		t.Errorf("'n': answered=%v confirmed=%v, want true/false",
			c.Answered(), c.Confirmed())
	}
}

func TestConfirmToggleThenEnter(t *testing.T) {
	c := loom.NewConfirm("Commit?")          // starts on Yes
	c.HandleKey(loom.KeyEvent{Key: "right"}) // move to No
	c.HandleKey(loom.KeyEvent{Key: "enter"})
	if !c.Answered() || c.Confirmed() {
		t.Errorf("toggle→Enter: answered=%v confirmed=%v, want true/false",
			c.Answered(), c.Confirmed())
	}
}

func TestConfirmEscIsNo(t *testing.T) {
	c := loom.NewConfirm("Commit?")
	if quit := c.HandleKey(loom.KeyEvent{Key: "esc"}); !quit {
		t.Fatal("Esc should quit")
	}
	if !c.Answered() || c.Confirmed() {
		t.Errorf("Esc: answered=%v confirmed=%v, want true/false",
			c.Answered(), c.Confirmed())
	}
}

func TestConfirmDefaultNo(t *testing.T) {
	c := loom.NewConfirm("Delete?").DefaultNo()
	c.HandleKey(loom.KeyEvent{Key: "enter"}) // confirms the highlighted (No)
	if c.Confirmed() {
		t.Error("DefaultNo + Enter should not confirm")
	}
}
