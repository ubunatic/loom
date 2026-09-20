// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitor

import (
	"time"

	"codeberg.org/ubunatic/loom"
)

// monitorState owns simulated producer data. Sampling is the only operation
// that advances histories; Snapshot returns independent data for rendering.
type monitorState struct {
	current      monitorSnapshot
	metrics      *loom.MetricStore
	hardwareTick int
	usageTick    int
}

func newMonitorState(initial monitorSnapshot, capacity int) *monitorState {
	state := &monitorState{metrics: loom.NewMetricStore(loom.Retention(capacity))}
	state.publishSnapshot(initial, time.Time{})
	return state
}

// SampleHardware appends one deterministic observation to hardware load series.
func (s *monitorState) SampleHardware(at time.Time) {
	s.hardwareTick++
	s.current.timestamp = at
	phase := float64(s.hardwareTick % 8)
	s.metrics.Publish("cpu (16c)", 3+phase*2, at)
	s.metrics.Publish("ram (45G)", 35+phase/2, at)
	s.metrics.Publish("gpu (Phx)", phase*2, at)
	s.metrics.Publish("vram", 4+phase/2, at)
	s.metrics.Publish("gtt", 1+phase/4, at)
}

// SampleUsage updates the usage meters at their own rate.
func (s *monitorState) SampleUsage(at time.Time) {
	s.usageTick++
	s.current.timestamp = at
	phase := float64(s.usageTick % 8)
	s.current.usage = []float64{60 + phase, 95 - phase/2, 99 - phase/3, 41 + phase/2}
	s.current.usage2 = []float64{3 + phase/4, phase / 4, phase / 4, 1 + phase/4}
}

// SampleAt appends deterministic observations to all series for the given timestamp.
func (s *monitorState) SampleAt(at time.Time) {
	s.SampleHardware(at)
	s.SampleUsage(at)
}

// Sample appends one deterministic observation to every series.
func (s *monitorState) Sample() {
	s.SampleAt(time.Now())
}

// Snapshot copies all slices so rendering cannot mutate producer history.
func (s *monitorState) Snapshot() monitorSnapshot {
	snapshot := s.metrics.Snapshot()
	result := s.current
	result.load = make(map[string][]float64, len(snapshot.Series))
	for _, name := range []string{"cpu (16c)", "ram (45G)", "gpu (Phx)"} {
		result.load[name] = snapshot.Values(name)
	}
	result.vram = snapshot.Values("vram")
	result.gtt = snapshot.Values("gtt")
	return result
}

func (s *monitorState) publishSnapshot(snapshot monitorSnapshot, at time.Time) {
	s.current = snapshot
	for name, values := range snapshot.load {
		for _, value := range values {
			s.metrics.Publish(name, value, at)
		}
	}
	for _, value := range snapshot.vram {
		s.metrics.Publish("vram", value, at)
	}
	for _, value := range snapshot.gtt {
		s.metrics.Publish("gtt", value, at)
	}
}
