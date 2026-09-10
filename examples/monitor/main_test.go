// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom/graph"
)

func TestShowOnce(t *testing.T) {
	var out bytes.Buffer
	if err := run(&out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "\x1b") {
		t.Fatal("plain output contains terminal controls")
	}
	rows := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(rows) != 10 {
		t.Fatalf("got %d rows, want 10", len(rows))
	}
	for i, row := range rows {
		if len([]rune(row)) != 64 {
			t.Errorf("row %d is not 64 cells: %q", i, row)
		}
	}
	if !strings.Contains(rows[1], "All Usage") || !strings.Contains(rows[1], "Load") {
		t.Fatal("embedded declaration not rendered")
	}
	content := out.String()
	for _, expected := range []string{
		"Claude",
		graph.RenderBar(60, graph.BarOptions{Width: 4, SubChar: true}),
		"Gemini",
		"cpu (16c)",
		splitTimeline(staticSnapshot.vram, staticSnapshot.gtt),
	} {
		if !strings.Contains(content, expected) {
			t.Errorf("monitor output missing %q:\n%s", expected, content)
		}
	}
}

func TestMonitorGraphWidths(t *testing.T) {
	if got := timeline(nil, 10); got != "[▁▁▁▁▁▁▁▁▁▁]" {
		t.Fatalf("empty timeline = %q", got)
	}
	if got := timeline([]float64{50}, 10); len([]rune(got)) != 12 {
		t.Fatalf("short timeline width = %d, want 12", len([]rune(got)))
	}
	if got := splitTimeline(nil, nil); len([]rune(got)) != 12 {
		t.Fatalf("empty split timeline width = %d, want 12", len([]rune(got)))
	}
}

func TestMonitorStateSamplesIndependentlyFromSnapshot(t *testing.T) {
	state := newMonitorState(monitorSnapshot{load: map[string][]float64{}}, 2)
	state.Sample()
	first := state.Snapshot()
	state.Sample()
	second := state.Snapshot()
	if len(first.load["cpu (16c)"]) != 1 || len(second.load["cpu (16c)"]) != 2 {
		t.Fatalf("history lengths = %d, %d", len(first.load["cpu (16c)"]), len(second.load["cpu (16c)"]))
	}
	first.load["cpu (16c)"][0] = 999
	if second.load["cpu (16c)"][0] == 999 {
		t.Fatal("snapshot shares mutable history storage")
	}
	state.Sample()
	if got := len(state.Snapshot().load["cpu (16c)"]); got != 2 {
		t.Fatalf("retained history length = %d, want 2", got)
	}
}

func TestCommand(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
		fail bool
	}{
		{"help", []string{"-h"}, "--watch", false},
		{"once", nil, "All Usage", false},
		{"unknown", []string{"--watc"}, "", true},
		{"argument", []string{"extra"}, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			err := execute(context.Background(), tc.args, &out)
			if (err != nil) != tc.fail {
				t.Fatalf("got %v", err)
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Fatal(out.String())
			}
		})
	}
}

type brokenWriter struct{ err error }

func (w brokenWriter) Write([]byte) (int, error) { return 0, w.err }

func TestOutputFailure(t *testing.T) {
	want := errors.New("closed output")
	if err := run(brokenWriter{want}); !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}
