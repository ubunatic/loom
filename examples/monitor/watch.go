// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"text/template"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/collector"
	"gopkg.in/yaml.v3"
)

type watchSpec struct {
	Collect     time.Duration    `yaml:"collect"`
	Redraw      time.Duration    `yaml:"redraw"`
	ClockFormat string           `yaml:"clock_format"`
	Title       string           `yaml:"title"`
	Status      string           `yaml:"status"`
	Command     string           `yaml:"command"`
	Description string           `yaml:"description"`
	WatchHelp   string           `yaml:"watch_help"`
	WidthHelp   string           `yaml:"width_help"`
	Collectors  []collector.Spec `yaml:"collectors"`
	sources     []configuredSource
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
	data, err := documents.ReadFile("spec/monitor.yaml")
	if err != nil {
		return err
	}
	root, cfg, err := loom.BuildWidget(bytes.NewReader(data))
	if err != nil {
		return err
	}
	frame, ok := root.(*loom.Frame)
	if !ok {
		return fmt.Errorf("watch: declared root must be a frame")
	}
	title, err := template.New("title").Option("missingkey=error").Parse(spec.Title)
	if err != nil {
		return err
	}
	baseTitle := frame.Title
	frame.Status = spec.Status
	state := newMonitorState(staticSnapshot, 32)
	runtime := startSources(ctx, spec.sources)
	defer runtime.Close()
	collect := func(now time.Time) error {
		if err := runtime.Err(); err != nil {
			return err
		}
		state.Sample()
		applySnapshot(frame, state.Snapshot())
		var b bytes.Buffer
		if err := title.Execute(&b, struct{ Title, Time string }{baseTitle, now.Format(spec.ClockFormat)}); err != nil {
			return err
		}
		frame.Title = b.String()
		return nil
	}
	// Validate presentation before taking ownership of a terminal.
	if err := collect(time.Now()); err != nil {
		return err
	}
	pane, err := loom.New(cfg.Height(0))
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.MaxCols = cfg.MaxWidth()
	pane.Resizeable = true
	return pane.RunWatch(ctx, frame, loom.Cadence{Collect: spec.Collect, Redraw: spec.Redraw}, collect)
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
