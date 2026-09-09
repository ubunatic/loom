// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"codeberg.org/ubunatic/loom"
	"gopkg.in/yaml.v3"
)

type watchSpec struct {
	Collect     time.Duration `yaml:"collect"`
	Redraw      time.Duration `yaml:"redraw"`
	ClockFormat string        `yaml:"clock_format"`
	Title       string        `yaml:"title"`
	Status      string        `yaml:"status"`
	Command     string        `yaml:"command"`
	Description string        `yaml:"description"`
	WatchHelp   string        `yaml:"watch_help"`
	WidthHelp   string        `yaml:"width_help"`
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
	collect := func(now time.Time) error {
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
