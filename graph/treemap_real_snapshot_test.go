// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

// realProcSnapshot is a trimmed `ps -eo pid,ppid,pcpu,comm --no-headers`
// capture from a real developer machine (2026-09-14), kept as a regression
// fixture for AggregateTreemap on realistic process-tree shapes: several
// same-name "Isolated Web Co" processes (Firefox's per-tab content
// processes, i.e. browser tabs) scattered under one forkserver, three
// unrelated "claude" processes under different shells, and a mix of
// zero-usage housekeeping processes that should disappear entirely from
// the aggregated result rather than clutter it.
const realProcSnapshot = `
1 0 0.0 systemd
2 2 0.0 kthreadd
2286 2166 1.4 gnome-shell
7056 2166 4.2 tilix
19297 7056 0.0 zsh
56643 2166 2.2 firefox
56769 56643 0.0 forkserver
56911 56769 0.0 Socket Process
56941 56769 0.0 Privileged Cont
56950 56769 0.6 RDD Process
57747 56769 0.0 WebExtensions
57928 56769 0.0 Utility Process
113527 7056 0.0 zsh
119403 7056 0.0 zsh
130685 2166 0.0 lmcoder
130692 130685 0.2 llama-server
138665 56769 0.0 Isolated Web Co
226848 56769 0.0 Isolated Web Co
231773 56769 0.0 Isolated Web Co
236753 2166 0.0 voxi
236761 236753 0.7 voxi
262857 19297 10.8 agy
309645 56769 0.2 Isolated Web Co
314572 56769 0.2 Isolated Servic
314648 56769 2.0 Isolated Web Co
314812 56769 4.9 Isolated Web Co
315383 56769 0.0 Isolated Web Co
345355 56769 0.0 Isolated Web Co
351637 56769 1.8 Isolated Web Co
365553 7056 0.0 zsh
366912 56769 0.0 Isolated Web Co
370002 113527 8.6 harnez
371180 119403 3.8 claude
375313 56769 0.0 Isolated Web Co
382040 56769 0.0 Web Content
383068 365553 3.3 claude
400172 7056 0.0 zsh
401335 400172 3.7 claude
402297 401627 10.4 macos
401627 401625 0.0 bash
401625 401623 0.0 tini
401623 2166 0.0 conmon
408186 56769 0.0 Web Content
411860 56769 0.0 Web Content
415543 401335 0.0 zsh
415545 415543 0.0 gitui
415552 415545 0.0 bash
`

// buildProcTree parses a `ps -eo pid,ppid,pcpu,comm` snapshot into a
// TreemapNode tree rooted at a synthetic "root" whose children are every
// process with no known parent in the snapshot.
func buildProcTree(t *testing.T, snapshot string) TreemapNode {
	t.Helper()
	type proc struct {
		pid, ppid int
		pcpu      float64
		name      string
	}
	var procs []proc
	for _, line := range strings.Split(strings.TrimSpace(snapshot), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			t.Fatalf("bad pid in line %q: %v", line, err)
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			t.Fatalf("bad ppid in line %q: %v", line, err)
		}
		pcpu, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			t.Fatalf("bad pcpu in line %q: %v", line, err)
		}
		procs = append(procs, proc{pid: pid, ppid: ppid, pcpu: pcpu, name: strings.Join(fields[3:], " ")})
	}

	byPID := make(map[int]proc, len(procs))
	byParent := map[int][]int{}
	for _, p := range procs {
		byPID[p.pid] = p
		byParent[p.ppid] = append(byParent[p.ppid], p.pid)
	}
	var build func(pid int) TreemapNode
	build = func(pid int) TreemapNode {
		p := byPID[pid]
		n := TreemapNode{Name: p.name, Value: p.pcpu}
		for _, childPID := range byParent[pid] {
			if childPID == pid {
				continue
			}
			n.Children = append(n.Children, build(childPID))
		}
		return n
	}
	root := TreemapNode{Name: "root"}
	for _, p := range procs {
		if _, hasParent := byPID[p.ppid]; !hasParent || p.pid == p.ppid {
			root.Children = append(root.Children, build(p.pid))
		}
	}
	return root
}

func TestAggregateTreemapRealProcSnapshot(t *testing.T) {
	root := buildProcTree(t, realProcSnapshot)
	segments := AggregateTreemap(root, MaxTreemapNodes)

	want := map[string]float64{
		"claude":          10.8, // 3.8 + 3.7 + 3.3 under three different shells
		"Isolated Web Co": 8.9,  // a dozen Firefox content (tab) processes merged
		"agy":             10.8,
		"macos":           10.4,
		"harnez":          8.6,
		"tilix":           4.2,
		"firefox":         2.2,
		"gnome-shell":     1.4,
		"RDD Process":     0.6,
		"voxi":            0.7,
		"llama-server":    0.2,
		"Isolated Servic": 0.2,
	}
	if len(segments) != len(want) {
		t.Errorf("len(segments) = %d, want %d (segments=%+v)", len(segments), len(want), segments)
	}
	var total float64
	for _, s := range segments {
		total += s.Value
		wantValue, ok := want[s.Name]
		if !ok {
			t.Errorf("unexpected segment %q (value %.1f) -- zero-usage housekeeping processes should have been dropped", s.Name, s.Value)
			continue
		}
		if math.Abs(s.Value-wantValue) > 0.05 {
			t.Errorf("segment %q = %.2f, want %.2f", s.Name, s.Value, wantValue)
		}
	}
	const wantTotal = 59.0
	if math.Abs(total-wantTotal) > 0.05 {
		t.Errorf("total = %.2f, want %.2f", total, wantTotal)
	}

	values := make([]float64, len(segments))
	for i, s := range segments {
		values[i] = s.Value
	}
	bar := RenderStackedBar(values, StackedBarOptions{Width: 60})
	if got := len([]rune(bar)) - 2; got != 60 { // -2 for the "[" "]" wrapper
		t.Errorf("RenderStackedBar width = %d, want 60 (bar=%q)", got, bar)
	}
	t.Logf("stacked bar: %s", bar)
}
