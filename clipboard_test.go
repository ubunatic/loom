// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"bytes"
	"testing"

	"ubunatic.com/loom"
)

func TestWriteOSC52ExactBytes(t *testing.T) {
	var got bytes.Buffer
	if err := loom.WriteOSC52(&got, "hi ☃"); err != nil {
		t.Fatal(err)
	}
	const want = "\x1b]52;c;aGkg4piD\x07"
	if got.String() != want {
		t.Fatalf("OSC 52 = %q, want %q", got.String(), want)
	}
}

func TestTextInputCopySelectionAndWholeValue(t *testing.T) {
	input := loom.NewTextInput("hello")
	input.Keys = loom.NewKeyMap(map[string][]string{"copy": {"ctrl-y"}})
	var copied string
	input.SetClipboardWriter(func(s string) { copied = s })
	input.SetSelection(1, 4)
	if !input.ConsumeKey(loom.KeyEvent{Key: "ctrl-y"}).Consumed || copied != "ell" {
		t.Fatalf("selected copy = %q", copied)
	}
	input.SetSelection(2, 2)
	input.ConsumeKey(loom.KeyEvent{Key: "ctrl-y"})
	if copied != "hello" {
		t.Fatalf("whole copy = %q, want hello", copied)
	}
}

func TestMaskedTextInputRefusesCopy(t *testing.T) {
	input := loom.NewTextInput("secret")
	input.Mask = '•'
	input.Keys = loom.NewKeyMap(map[string][]string{"copy": {"ctrl-y"}})
	writes := 0
	input.SetClipboardWriter(func(string) { writes++ })
	if input.ConsumeKey(loom.KeyEvent{Key: "ctrl-y"}).Consumed {
		t.Fatal("masked input consumed copy action")
	}
	if writes != 0 {
		t.Fatalf("masked input wrote clipboard %d times, want 0", writes)
	}
}

func TestTextAreaCopySelectionAndWholeValue(t *testing.T) {
	area := loom.NewTextArea("one\ntwo")
	area.Keys = loom.NewKeyMap(map[string][]string{"copy": {"ctrl-y"}})
	var copied string
	area.SetClipboardWriter(func(s string) { copied = s })
	area.SetSelection(2, 6)
	area.ConsumeKey(loom.KeyEvent{Key: "ctrl-y"})
	if copied != "e\ntw" {
		t.Fatalf("selected copy = %q, want %q", copied, "e\ntw")
	}
	area.SetSelection(0, 0)
	area.ConsumeKey(loom.KeyEvent{Key: "ctrl-y"})
	if copied != "one\ntwo" {
		t.Fatalf("whole copy = %q", copied)
	}
}
