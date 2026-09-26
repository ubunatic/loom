// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitor

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/collector"
	"codeberg.org/ubunatic/loom/graph"
	"codeberg.org/ubunatic/loom/internal/ptytest"
)

type blockingCollector struct {
	mu      sync.Mutex
	started int
	stopped chan struct{}
}

func (*blockingCollector) Type() collector.Type { return collector.TypeFile }
func (c *blockingCollector) Collect(ctx context.Context, at time.Time) (collector.Record, error) {
	c.mu.Lock()
	c.started++
	c.mu.Unlock()
	<-ctx.Done()
	select {
	case c.stopped <- struct{}{}:
	default:
	}
	return collector.Record{}, ctx.Err()
}

func TestMonitorCollectorsFollowFocusAndClose(t *testing.T) {
	w, err := NewWidget([]string{"--watch"})
	if err != nil {
		t.Fatal(err)
	}
	m := w.(*monitorWidget)
	fake := &blockingCollector{stopped: make(chan struct{}, 2)}
	history, err := collector.NewHistory(time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	m.spec.sources = []configuredSource{{collector: fake, interval: time.Hour, history: history}}
	m.SetFocus(false)
	if m.runtime != nil {
		t.Fatal("inactive monitor started collectors")
	}
	m.SetFocus(true)
	waitCollectorStart(t, fake, 1)
	m.SetFocus(false)
	select {
	case <-fake.stopped:
	case <-time.After(time.Second):
		t.Fatal("collector leaked after deactivation")
	}
	m.SetFocus(true)
	waitCollectorStart(t, fake, 2)
	m.Close()
	select {
	case <-fake.stopped:
	case <-time.After(time.Second):
		t.Fatal("collector leaked after Close")
	}
}

func waitCollectorStart(t *testing.T, c *blockingCollector, count int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		got := c.started
		c.mu.Unlock()
		if got >= count {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("collector starts = %d, want %d", c.started, count)
}

type markerCollector struct{ marker string }

func (*markerCollector) Type() collector.Type { return collector.TypeFile }
func (c *markerCollector) Collect(ctx context.Context, at time.Time) (collector.Record, error) {
	<-ctx.Done()
	if err := os.WriteFile(c.marker, []byte("stopped"), 0600); err != nil {
		return collector.Record{}, err
	}
	return collector.Record{}, ctx.Err()
}

func TestRunWatchShutdownHelper(t *testing.T) {
	marker := os.Getenv("LOOM_MONITOR_STOP_MARKER")
	if marker == "" {
		return
	}
	spec, err := loadWatch()
	if err != nil {
		t.Fatal(err)
	}
	history, err := collector.NewHistory(time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	spec.sources = []configuredSource{{collector: &markerCollector{marker: marker}, interval: time.Hour, history: history}}
	if err := runWatch(context.Background(), spec); err != nil {
		t.Fatal(err)
	}
}

func TestRunWatchClosesCollectorsOnF10(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "collector-stopped")
	old, hadOld := os.LookupEnv("LOOM_MONITOR_STOP_MARKER")
	if err := os.Setenv("LOOM_MONITOR_STOP_MARKER", marker); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if hadOld {
			_ = os.Setenv("LOOM_MONITOR_STOP_MARKER", old)
		} else {
			_ = os.Unsetenv("LOOM_MONITOR_STOP_MARKER")
		}
	}()
	session := ptytest.Start(t, 100, 30, os.Args[0], "-test.run=^TestRunWatchShutdownHelper$")
	session.WaitFor("Loom monitor", 3*time.Second)
	session.Send("\x1b[21~")
	if err := session.Wait(3 * time.Second); err != nil {
		t.Fatalf("monitor did not exit after F10: %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("collector was not stopped during runWatch shutdown: %v", err)
	}
}

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
	if !strings.Contains(rows[0], "Loom monitor (") || !strings.Contains(rows[0], "effective: 80)") {
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
			err := Execute(context.Background(), tc.args, &out)
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
