// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

// monitorState owns simulated producer data. Sampling is the only operation
// that advances histories; Snapshot returns independent data for rendering.
type monitorState struct {
	current  monitorSnapshot
	capacity int
	tick     int
}

func newMonitorState(initial monitorSnapshot, capacity int) *monitorState {
	state := &monitorState{capacity: capacity}
	state.current = cloneSnapshot(initial)
	return state
}

// Sample appends one deterministic observation to every series.
func (s *monitorState) Sample() {
	s.tick++
	phase := float64(s.tick % 8)
	s.current.usage = []float64{60 + phase, 95 - phase/2, 99 - phase/3, 41 + phase/2}
	s.append("cpu (16c)", 3+phase*2)
	s.append("ram (45G)", 35+phase/2)
	s.append("gpu (Phx)", phase*2)
	s.appendValues(&s.current.vram, 4+phase/2)
	s.appendValues(&s.current.gtt, 1+phase/4)
}

func (s *monitorState) append(name string, value float64) {
	values := s.current.load[name]
	s.current.load[name] = boundedAppend(values, value, s.capacity)
}

func (s *monitorState) appendValues(values *[]float64, value float64) {
	*values = boundedAppend(*values, value, s.capacity)
}

// Snapshot copies all slices so rendering cannot mutate producer history.
func (s *monitorState) Snapshot() monitorSnapshot {
	return cloneSnapshot(s.current)
}

func cloneSnapshot(source monitorSnapshot) monitorSnapshot {
	target := monitorSnapshot{
		usage: append([]float64(nil), source.usage...),
		load:  make(map[string][]float64, len(source.load)),
		vram:  append([]float64(nil), source.vram...),
		gtt:   append([]float64(nil), source.gtt...),
	}
	for name, values := range source.load {
		target.load[name] = append([]float64(nil), values...)
	}
	return target
}

func boundedAppend(values []float64, value float64, capacity int) []float64 {
	values = append(append([]float64(nil), values...), value)
	if capacity > 0 && len(values) > capacity {
		values = values[len(values)-capacity:]
	}
	return values
}
