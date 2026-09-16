// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package treemap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"codeberg.org/ubunatic/loom/graph"
)

// Process collection is demo data plumbing. The graph package only receives
// TreemapNodes and has no dependency on ps or the local process table.
type procEntry struct {
	pid, ppid int
	pcpu      float64
	name      string
}

// readProcTree shells out to ps and builds a graph.TreemapNode tree. When
// excludeSelf is set, the current process and its descendants are omitted.
// A direct `go run ./examples/treemap` launcher is also omitted.
func readProcTree(ctx context.Context, excludeSelf bool) (graph.TreemapNode, error) {
	excludePID := 0
	if excludeSelf {
		excludePID = os.Getpid()
		parentPID := os.Getppid()
		cmdline, err := exec.CommandContext(ctx, "ps", "-ww", "-p", strconv.Itoa(parentPID), "-o", "args=").Output()
		if err == nil && isTreemapGoRun(string(cmdline)) {
			excludePID = parentPID
		}
	}
	cmd := exec.CommandContext(ctx, "ps", "-eo", "pid,ppid,pcpu,comm", "--no-headers")
	out, err := cmd.Output()
	if err != nil {
		return graph.TreemapNode{}, fmt.Errorf("treemap: running ps: %w", err)
	}
	return buildProcTree(string(out), excludePID), nil
}

func isTreemapGoRun(commandLine string) bool {
	args := strings.Fields(commandLine)
	if len(args) < 3 || filepath.Base(args[0]) != "go" || args[1] != "run" {
		return false
	}
	for _, arg := range args[2:] {
		if arg == "./examples/treemap" || strings.HasPrefix(arg, "./examples/treemap/") || strings.HasSuffix(arg, "/examples/treemap") {
			return true
		}
	}
	return false
}

func buildProcTree(psOutput string, excludePID int) graph.TreemapNode {
	var procs []procEntry
	for _, line := range strings.Split(strings.TrimSpace(psOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		pcpu, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			continue
		}
		procs = append(procs, procEntry{pid: pid, ppid: ppid, pcpu: pcpu, name: strings.Join(fields[3:], " ")})
	}

	byPID := make(map[int]procEntry, len(procs))
	byParent := map[int][]int{}
	for _, p := range procs {
		byPID[p.pid] = p
		byParent[p.ppid] = append(byParent[p.ppid], p.pid)
	}
	var build func(pid int) graph.TreemapNode
	build = func(pid int) graph.TreemapNode {
		p := byPID[pid]
		n := graph.TreemapNode{Name: p.name, Value: p.pcpu}
		for _, childPID := range byParent[pid] {
			if childPID == pid || childPID == excludePID {
				continue
			}
			n.Children = append(n.Children, build(childPID))
		}
		return n
	}
	root := graph.TreemapNode{Name: "root"}
	for _, p := range procs {
		if p.pid == excludePID {
			continue
		}
		if _, hasParent := byPID[p.ppid]; !hasParent || p.pid == p.ppid {
			root.Children = append(root.Children, build(p.pid))
		}
	}
	return root
}
