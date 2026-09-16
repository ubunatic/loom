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
	if len(rows) != 12 {
		t.Fatalf("got %d rows, want 12", len(rows))
	}
	for i, row := range rows {
		if len([]rune(row)) != 80 {
			t.Errorf("row %d is not 80 cells: %q", i, row)
		}
	}
	if !strings.Contains(rows[1], "All Usage") || !strings.Contains(rows[1], "Load") {
		t.Fatal("embedded declaration not rendered")
	}
	if !strings.Contains(rows[0], "Loom monitor (observed: n/a, effective: 80)") {
		t.Fatalf("width title missing: %q", rows[0])
	}
	if !strings.Contains(strings.Join(rows, "\n"), "(simulated data)") {
		t.Fatalf("box provenance hints missing:\n%s", strings.Join(rows, "\n"))
	}
	content := out.String()
	for _, expected := range []string{
		"Claude",
		graph.RenderBar(staticSnapshot.usage[0], graph.BarOptions{Width: 4, SubChar: true}),
		graph.RenderBar(staticSnapshot.usage2[0], graph.BarOptions{Width: 4, SubChar: true}),
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
	state := newMonitorState(monitorSnapshot{
		usage:  []float64{10},
		usage2: []float64{20},
		load:   map[string][]float64{},
	}, 2)
	state.Sample()
	first := state.Snapshot()
	state.Sample()
	second := state.Snapshot()
	if len(first.load["cpu (16c)"]) != 1 || len(second.load["cpu (16c)"]) != 2 {
		t.Fatalf("history lengths = %d, %d", len(first.load["cpu (16c)"]), len(second.load["cpu (16c)"]))
	}
	if len(first.usage2) != 4 || len(second.usage2) != 4 {
		t.Fatalf("usage2 lengths = %d, %d", len(first.usage2), len(second.usage2))
	}
	first.load["cpu (16c)"][0] = 999
	first.usage2[0] = 999
	if second.load["cpu (16c)"][0] == 999 || second.usage2[0] == 999 {
		t.Fatal("snapshot shares mutable history storage")
	}
	state.Sample()
	if got := len(state.Snapshot().load["cpu (16c)"]); got != 2 {
		t.Fatalf("retained history length = %d, want 2", got)
	}
}

func TestTinyTerminalGraphTargetRendering(t *testing.T) {
	for _, width := range []int{10, 15, 20, 25} {
		var out bytes.Buffer
		if err := runWidth(&out, width); err != nil {
			t.Fatalf("runWidth(%d) failed: %v", width, err)
		}
		rows := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
		for i, row := range rows {
			if runes := len([]rune(row)); runes != width {
				t.Fatalf("width %d row %d rendered %d runes: %q", width, i, runes, row)
			}
		}
	}
}

func TestCommand(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []string
		want  string
		width int
		fail  bool
	}{
		{"help", []string{"-h"}, "--watch", 0, false},
		{"once", nil, "All Usage", 0, false},
		{"short width", []string{"-w", "40"}, "All Usage", 40, false},
		{"width capped by spec", []string{"-w", "100"}, "effective: 80", 80, false},
		{"unknown", []string{"--watc"}, "", 0, true},
		{"argument", []string{"extra"}, "", 0, true},
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
			if tc.width > 0 {
				rows := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
				for i, row := range rows {
					if len([]rune(row)) != tc.width {
						t.Fatalf("row %d width=%d, want %d", i, len([]rune(row)), tc.width)
					}
				}
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

func TestParseProcStatAndPercentage(t *testing.T) {
	raw1 := []byte("cpu  1000 0 200 8000 0 0 0 0 0 0\n")
	s1, err := parseProcStat(raw1)
	if err != nil {
		t.Fatalf("parse raw1 failed: %v", err)
	}
	if s1.user != 1000 || s1.system != 200 || s1.idle != 8000 {
		t.Fatalf("unexpected stat values: %+v", s1)
	}

	raw2 := []byte("cpu  1050 0 250 8700 0 0 0 0 0 0\n")
	s2, err := parseProcStat(raw2)
	if err != nil {
		t.Fatalf("parse raw2 failed: %v", err)
	}

	// Total diff: (1050+250+8700) - (1000+200+8000) = 10000 - 9200 = 800
	// Idle diff: 8700 - 8000 = 700
	// Busy diff: 800 - 700 = 100
	// Pct: 100 / 800 = 12.5%
	pct := cpuPercentage(s1, s2)
	if pct != 12.5 {
		t.Fatalf("expected 12.5%%, got %f%%", pct)
	}

	// Test invalid / malformed inputs
	if _, err := parseProcStat([]byte("mem 10 20 30")); err == nil {
		t.Fatal("accepted non-cpu prefix")
	}
	if _, err := parseProcStat([]byte("cpu not-a-number")); err == nil {
		t.Fatal("accepted malformed numbers")
	}
	if _, err := parseProcStat(nil); err == nil {
		t.Fatal("accepted nil data")
	}

	// Edge case: equal totals
	if p := cpuPercentage(s1, s1); p != 0 {
		t.Fatalf("expected 0 for equal stats, got %f", p)
	}
}
