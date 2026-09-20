// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package examplesreg

import (
	"reflect"
	"testing"
)

func TestInteractiveExamplesHaveDemoArgs(t *testing.T) {
	want := map[string][]string{
		"monitor": []string{"--watch"},
		"splash":  []string{"--watch"},
		"treemap": []string{"--watch", "--ansi"},
	}
	for _, name := range []string{"monitor", "splash", "treemap"} {
		example, ok := Find(name)
		if !ok {
			t.Fatalf("Find(%q) returned no example", name)
		}
		if !reflect.DeepEqual(example.DemoArgs, want[name]) {
			t.Errorf("Find(%q).DemoArgs = %v, want %v", name, example.DemoArgs, want[name])
		}
	}
}
