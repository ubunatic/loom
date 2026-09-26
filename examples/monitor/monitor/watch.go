// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitor

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/collector"
	"gopkg.in/yaml.v3"
)

type watchSpec struct {
	Collect      time.Duration    `yaml:"collect"`
	Redraw       time.Duration    `yaml:"redraw"`
	ClockFormat  string           `yaml:"clock_format"`
	Title        string           `yaml:"title"`
	Status       string           `yaml:"status"`
	Command      string           `yaml:"command"`
	Description  string           `yaml:"description"`
	WatchHelp    string           `yaml:"watch_help"`
	WidthHelp    string           `yaml:"width_help"`
	Collectors   []collector.Spec `yaml:"collectors"`
	sources      []configuredSource
	watchEnabled bool
}

type configuredSource struct {
	collector collector.Collector
	interval  time.Duration
	history   *collector.History
}

func loadWatch() (watchSpec, error) {
	var spec watchSpec
	data, err := documents.ReadFile("spec/watch.yaml")
	if err != nil {
		return spec, err
	}
	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)
	if err := d.Decode(&spec); err != nil {
		return spec, err
	}
	if err := (loom.Cadence{Collect: spec.Collect, Redraw: spec.Redraw}).Validate(); err != nil {
		return spec, err
	}
	if spec.ClockFormat == "" || spec.Title == "" || spec.Command == "" || spec.Description == "" || spec.WatchHelp == "" || spec.WidthHelp == "" {
		return spec, fmt.Errorf("watch: missing required declaration")
	}
	for _, declaration := range spec.Collectors {
		c, interval, retention, err := declaration.Build()
		if err != nil {
			return spec, err
		}
		history, err := collector.NewHistory(retention)
		if err != nil {
			return spec, err
		}
		spec.sources = append(spec.sources, configuredSource{collector: c, interval: interval, history: history})
	}
	return spec, nil
}

func runWatch(ctx context.Context, spec watchSpec) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	spec.watchEnabled = true
	w, err := newMonitorWidget(spec)
	if err != nil {
		return err
	}
	pane, err := loom.New(10)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	if request := w.PaneRequest(); request.MaxCols > 0 {
		pane.MaxCols = request.MaxCols
	}
	return pane.Run(w)
}

type cpuStat struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

func parseProcStat(data []byte) (cpuStat, error) {
	var stat cpuStat
	var label string
	n, err := fmt.Sscanf(string(data), "%s %d %d %d %d %d %d %d %d",
		&label, &stat.user, &stat.nice, &stat.system, &stat.idle,
		&stat.iowait, &stat.irq, &stat.softirq, &stat.steal)
	if err != nil && n < 5 {
		return stat, fmt.Errorf("malformed stat record: %w", err)
	}
	if label != "cpu" {
		return stat, fmt.Errorf("unexpected stat prefix: %s", label)
	}
	return stat, nil
}

func cpuPercentage(prev, curr cpuStat) float64 {
	prevIdle := prev.idle + prev.iowait
	currIdle := curr.idle + curr.iowait
	prevTotal := prevIdle + prev.user + prev.nice + prev.system + prev.irq + prev.softirq + prev.steal
	currTotal := currIdle + curr.user + curr.nice + curr.system + curr.irq + curr.softirq + curr.steal
	if currTotal <= prevTotal {
		return 0
	}
	totalDiff := float64(currTotal - prevTotal)
	idleDiff := float64(currIdle - prevIdle)
	pct := (totalDiff - idleDiff) / totalDiff * 100.0
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return pct
}

type sourceRuntime struct {
	cancel context.CancelFunc
	done   sync.WaitGroup
	errs   chan error
}

func startSources(parent context.Context, sources []configuredSource) *sourceRuntime {
	ctx, cancel := context.WithCancel(parent)
	runtime := &sourceRuntime{cancel: cancel, errs: make(chan error, len(sources))}
	for _, source := range sources {
		source := source
		runtime.done.Add(1)
		go func() {
			defer runtime.done.Done()
			err := collector.Run(ctx, source.collector, source.interval, func(record collector.Record) error {
				source.history.Append(record)
				return nil
			})
			if err != nil && ctx.Err() == nil {
				runtime.errs <- err
			}
		}()
	}
	return runtime
}

func (r *sourceRuntime) Err() error {
	select {
	case err := <-r.errs:
		return err
	default:
		return nil
	}
}

func (r *sourceRuntime) Close() {
	r.cancel()
	r.done.Wait()
}
