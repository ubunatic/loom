// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestRunPaneCompletionExitsAndCancellationAborts(t *testing.T) {
	choice := NewChoice([]Item{{Name: "one"}})
	runner := paneableRunner{Paneable: choice}
	if result := runner.ConsumeKey(KeyEvent{Key: "enter"}); !result.Quit || result.Done {
		t.Fatalf("standalone Choice confirmation = %+v, want runner exit", result)
	}
	if item, ok := choice.Selected(); !ok || item.Name != "one" {
		t.Fatalf("confirmed selection = %+v, %t; want one", item, ok)
	}

	choice = NewChoice([]Item{{Name: "one"}})
	runner = paneableRunner{Paneable: choice}
	if result := runner.ConsumeKey(KeyEvent{Key: "esc"}); !result.Quit || result.Done {
		t.Fatalf("standalone Choice cancellation = %+v, want abort exit", result)
	}
	if _, ok := choice.Selected(); ok || !choice.Aborted() {
		t.Fatal("cancelled Choice selected an item or did not abort")
	}
}
