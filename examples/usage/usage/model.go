// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package usage

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Quota is a fake usage window used by this example's All Usage panel.
type Quota struct {
	Name    string
	Used    float64
	ResetIn string
}

// Snapshot is the immutable data consumed by both example views.
type Snapshot struct {
	CapturedAt  time.Time
	Quotas      []Quota
	CPUPercent  float64
	CPUOK       bool
	CPUHistory  []float64
	NumCPU      int
	Load1       float64
	MemoryUsed  float64 // GiB
	MemoryTotal float64 // GiB
	MemoryOK    bool
	Err         string
}

// DataSource collects one coherent example snapshot.
type DataSource interface {
	Collect(context.Context, time.Time) (Snapshot, error)
}

// DataSourceFunc adapts a function into a DataSource.
type DataSourceFunc func(context.Context, time.Time) (Snapshot, error)

// Collect implements DataSource.
func (f DataSourceFunc) Collect(ctx context.Context, at time.Time) (Snapshot, error) {
	return f(ctx, at)
}

func sampleSnapshot(at time.Time) Snapshot {
	return Snapshot{
		CapturedAt: at,
		Quotas: []Quota{
			{Name: "Claude", Used: 64, ResetIn: "3h 20m"},
			{Name: "Codex", Used: 38, ResetIn: "1d 6h"},
			{Name: "Gemini", Used: 81, ResetIn: "5h 10m"},
		},
		CPUPercent:  27,
		CPUOK:       true,
		CPUHistory:  []float64{10, 15, 22, 18, 27},
		NumCPU:      runtime.NumCPU(),
		Load1:       0,
		MemoryUsed:  0,
		MemoryTotal: 0,
		MemoryOK:    false,
	}
}

type localSource struct {
	mu          sync.Mutex
	previous    cpuCounters
	haveCPU     bool
	cpuHistory  []float64
	quotaSample uint64
}

func (s *localSource) Collect(ctx context.Context, at time.Time) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := sampleSnapshot(at)
	s.quotaSample++
	for i := range result.Quotas {
		result.Quotas[i].Used += float64((s.quotaSample / uint64(i+2)) % 3)
	}
	result.NumCPU = runtime.NumCPU()
	if avg, err := readLoad1(); err == nil {
		result.Load1 = avg
	}
	if current, err := readCPU(); err == nil {
		if s.haveCPU {
			result.CPUPercent, result.CPUOK = cpuDelta(s.previous, current)
			if result.CPUOK {
				s.cpuHistory = append(s.cpuHistory, result.CPUPercent)
				if len(s.cpuHistory) > 20 {
					s.cpuHistory = append([]float64(nil), s.cpuHistory[len(s.cpuHistory)-20:]...)
				}
			}
		}
		s.previous, s.haveCPU = current, true
	}
	result.CPUHistory = append([]float64(nil), s.cpuHistory...)
	if used, total, err := readMemory(); err == nil {
		result.MemoryUsed, result.MemoryTotal, result.MemoryOK = used, total, true
	}
	return result, nil
}

type cpuCounters struct{ total, idle uint64 }

func readCPU() (cpuCounters, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuCounters{}, err
	}
	var label string
	var values [10]uint64
	n, scanErr := fmt.Sscan(string(data), &label, &values[0], &values[1], &values[2], &values[3], &values[4], &values[5], &values[6], &values[7], &values[8], &values[9])
	if scanErr != nil || n < 5 || label != "cpu" {
		return cpuCounters{}, fmt.Errorf("usage: malformed /proc/stat CPU row")
	}
	var total uint64
	for _, value := range values[:n-1] {
		total += value
	}
	return cpuCounters{total: total, idle: values[3] + values[4]}, nil
}

func cpuDelta(previous, current cpuCounters) (float64, bool) {
	if current.total <= previous.total || current.idle < previous.idle {
		return 0, false
	}
	total := float64(current.total - previous.total)
	idle := float64(current.idle - previous.idle)
	return max(0, min(100, (total-idle)/total*100)), true
}

func readLoad1() (float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0, fmt.Errorf("usage: malformed /proc/loadavg")
	}
	return strconv.ParseFloat(fields[0], 64)
}

func readMemory() (usedGiB, totalGiB float64, err error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	var total, available float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, parseErr := strconv.ParseFloat(fields[1], 64)
		if parseErr != nil {
			continue
		}
		switch strings.TrimSuffix(fields[0], ":") {
		case "MemTotal":
			total = value
		case "MemAvailable":
			available = value
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}
	if total <= 0 {
		return 0, 0, fmt.Errorf("usage: MemTotal unavailable")
	}
	return (total - available) / 1024 / 1024, total / 1024 / 1024, nil
}
