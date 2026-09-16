// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitor

import (
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom"
)

func TestMonitorIndependentSlowFastSampling(t *testing.T) {
	state := newMonitorState(monitorSnapshot{load: map[string][]float64{}}, 16)
	t0 := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)

	// Simulate 10 fast usage ticks for every 1 slow hardware tick.
	for i := 1; i <= 10; i++ {
		tFast := t0.Add(time.Duration(i*100) * time.Millisecond)
		state.SampleUsage(tFast)
	}
	snapUsageOnly := state.Snapshot()
	if len(snapUsageOnly.load["cpu (16c)"]) != 0 {
		t.Fatalf("expected 0 cpu samples, got %d", len(snapUsageOnly.load["cpu (16c)"]))
	}
	if snapUsageOnly.timestamp != t0.Add(1000*time.Millisecond) {
		t.Fatalf("unexpected snapshot timestamp: %v", snapUsageOnly.timestamp)
	}

	// Now sample hardware at 1s mark.
	state.SampleHardware(t0.Add(time.Second))
	snapWithHW := state.Snapshot()
	if len(snapWithHW.load["cpu (16c)"]) != 1 {
		t.Fatalf("expected 1 cpu sample, got %d", len(snapWithHW.load["cpu (16c)"]))
	}
	if len(snapWithHW.vram) != 1 || len(snapWithHW.gtt) != 1 {
		t.Fatalf("expected 1 vram/gtt sample, got vram=%d gtt=%d", len(snapWithHW.vram), len(snapWithHW.gtt))
	}
	// VRAM and GTT should have independent distinct values.
	if snapWithHW.vram[0] == snapWithHW.gtt[0] {
		t.Fatalf("vram (%v) and gtt (%v) should have independent values", snapWithHW.vram[0], snapWithHW.gtt[0])
	}
}

func TestIndependentProducerAndRedrawMatrix(t *testing.T) {
	// Matrix of collection cadence vs redraw cadence.
	// 1. Slow collect (1s), fast redraw (50ms / 20Hz)
	// 2. Fast collect (100ms), slow redraw (1s)
	t0 := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)

	t.Run("SlowCollectFastRedraw", func(t *testing.T) {
		state := newMonitorState(monitorSnapshot{load: map[string][]float64{}}, 32)
		var redrawCount int

		for sec := 1; sec <= 3; sec++ {
			state.SampleAt(t0.Add(time.Duration(sec) * time.Second))
			// Redraw 20 times between each collection tick.
			for frame := 1; frame <= 20; frame++ {
				snap := state.Snapshot()
				redrawCount++
				if len(snap.load["cpu (16c)"]) != sec {
					t.Fatalf("frame %d in sec %d: expected %d samples, got %d", frame, sec, sec, len(snap.load["cpu (16c)"]))
				}
				if snap.timestamp != t0.Add(time.Duration(sec)*time.Second) {
					t.Fatalf("frame %d in sec %d: snapshot timestamp changed unexpectedly: %v", frame, sec, snap.timestamp)
				}
			}
		}
		if redrawCount != 60 {
			t.Fatalf("expected 60 redraws, got %d", redrawCount)
		}
	})

	t.Run("FastCollectSlowRedraw", func(t *testing.T) {
		state := newMonitorState(monitorSnapshot{load: map[string][]float64{}}, 32)
		for tick := 1; tick <= 10; tick++ {
			state.SampleAt(t0.Add(time.Duration(tick*100) * time.Millisecond))
		}
		// Redraw once after 10 ticks.
		snap := state.Snapshot()
		if len(snap.load["cpu (16c)"]) != 10 {
			t.Fatalf("expected 10 samples, got %d", len(snap.load["cpu (16c)"]))
		}
		if snap.timestamp != t0.Add(time.Second) {
			t.Fatalf("expected timestamp %v, got %v", t0.Add(time.Second), snap.timestamp)
		}
	})
}

func TestSnapshotRenderingIdempotency(t *testing.T) {
	// Rendering the same snapshot repeatedly must produce byte-identical output
	// and must not advance producer history.
	state := newMonitorState(staticSnapshot, 16)
	state.Sample()
	snap := state.Snapshot()

	document, err := documents.Open("spec/monitor.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer document.Close()
	root, _, err := loom.BuildWidget(document)
	if err != nil {
		t.Fatal(err)
	}

	applySnapshot(root, snap)
	firstRender := loom.Render(root, 80, 12)

	for i := 0; i < 10; i++ {
		// Re-apply same snapshot and re-render.
		applySnapshot(root, snap)
		repeatedRender := loom.Render(root, 80, 12)
		if strings.Join(firstRender, "\n") != strings.Join(repeatedRender, "\n") {
			t.Fatalf("render %d differed from first render", i)
		}
	}

	// Ensure state did not advance.
	snapAfter := state.Snapshot()
	if len(snapAfter.load["cpu (16c)"]) != len(snap.load["cpu (16c)"]) {
		t.Fatalf("state history mutated during repeated render: before=%d, after=%d",
			len(snap.load["cpu (16c)"]), len(snapAfter.load["cpu (16c)"]))
	}
}
