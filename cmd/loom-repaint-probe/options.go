// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type cliOptions struct {
	fps            int
	areaWidth      int
	areaHeight     int
	tune           tuneOption
	animation      int
	paintMode      int
	background     int
	clearEach      int
	altScreen      int
	altScreenSet   bool
	outPath        string
	paintSet       bool
	backgroundSet  bool
	clearSet       bool
	incremental    bool
	incrementalSet bool
}

type tuneOption struct {
	enabled  bool
	duration time.Duration
}

func (o *tuneOption) String() string {
	if !o.enabled {
		return ""
	}
	return o.duration.String()
}

func (o *tuneOption) Set(value string) error {
	if value == "true" {
		o.enabled = true
		o.duration = 10 * time.Second
		return nil
	}
	if value == "false" {
		o.enabled = false
		return nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return fmt.Errorf("tune duration must be a positive Go duration, got %q", value)
	}
	o.enabled = true
	o.duration = duration
	return nil
}

func (*tuneOption) IsBoolFlag() bool { return true }

func parseOptions(args []string) (cliOptions, error) {
	options := cliOptions{fps: 30, areaWidth: minimumWidth, areaHeight: minimumHeight, animation: animationAstra, paintMode: -1, background: -1, clearEach: -1, altScreen: 1, incremental: true}
	var animationName, paintName, backgroundName, clearName string
	set := flag.NewFlagSet("loom-repaint-probe", flag.ContinueOnError)
	set.Usage = func() {
		fmt.Fprintf(set.Output(), "Usage: %s [options]\n", set.Name())
		set.PrintDefaults()
	}
	set.IntVar(&options.fps, "fps", options.fps, "target frame rate (1–120)")
	set.IntVar(&options.areaWidth, "width", options.areaWidth, "starting rectangle width in columns")
	set.IntVar(&options.areaWidth, "W", options.areaWidth, "alias for --width")
	set.IntVar(&options.areaHeight, "height", options.areaHeight, "starting rectangle height in rows")
	set.IntVar(&options.areaHeight, "H", options.areaHeight, "alias for --height")
	set.Var(&options.tune, "tune", "benchmark each mode for this duration (default when omitted: 10s)")
	set.StringVar(&animationName, "anim", "astra", "initial animation: astra, rainbow, or none")
	set.StringVar(&animationName, "A", "astra", "alias for --anim")
	set.IntVar(&options.altScreen, "altscreen", options.altScreen, "alternate screen: 0 or 1")
	set.IntVar(&options.altScreen, "t", options.altScreen, "alias for --altscreen")
	set.StringVar(&options.outPath, "out", "", "write --tune results to this file")
	set.StringVar(&options.outPath, "o", "", "alias for --out")
	set.BoolVar(&options.incremental, "incremental", true, "only write terminal rows that changed")
	set.BoolVar(&options.incremental, "i", true, "alias for --incremental")
	set.StringVar(&paintName, "paint", "", "initial paint mode: rows, spans, or cells (default rows)")
	set.StringVar(&paintName, "p", "", "alias for --paint")
	set.StringVar(&backgroundName, "background", "", "initial background mode: fill or clear")
	set.StringVar(&backgroundName, "b", "", "alias for --background")
	set.StringVar(&clearName, "clear", "", "initial clear mode: 0 or 1")
	set.StringVar(&clearName, "c", "", "alias for --clear")
	if err := set.Parse(normalizeTuneArgs(args)); err != nil {
		return options, err
	}
	set.Visit(func(item *flag.Flag) {
		switch item.Name {
		case "paint", "p":
			options.paintSet = true
		case "background", "b":
			options.backgroundSet = true
		case "clear", "c":
			options.clearSet = true
		case "incremental", "i":
			options.incrementalSet = true
		case "altscreen", "t":
			options.altScreenSet = true
		}
	})
	if len(set.Args()) != 0 {
		return options, fmt.Errorf("unexpected arguments: %s", strings.Join(set.Args(), " "))
	}
	if options.fps < 1 || options.fps > 120 {
		return options, errors.New("fps must be between 1 and 120")
	}
	if options.areaWidth < 1 || options.areaHeight < 1 {
		return options, errors.New("rectangle width and height must be positive")
	}
	if options.altScreen < -1 || options.altScreen > 1 {
		return options, errors.New("altscreen must be 0 or 1")
	}
	if options.outPath != "" && !options.tune.enabled {
		return options, errors.New("--out requires --tune")
	}
	var err error
	options.animation, err = parseAnimationMode(animationName)
	if err != nil {
		return options, fmt.Errorf("invalid animation: %w", err)
	}
	if paintName != "" {
		options.paintMode, err = parsePaintMode(paintName)
		if err != nil {
			return options, fmt.Errorf("invalid paint mode: %w", err)
		}
	}
	if backgroundName != "" {
		options.background, err = parseBackgroundMode(backgroundName)
		if err != nil {
			return options, fmt.Errorf("invalid background mode: %w", err)
		}
	}
	if clearName != "" {
		options.clearEach, err = parseClearMode(clearName)
		if err != nil {
			return options, fmt.Errorf("invalid clear mode: %w", err)
		}
	}
	return options, nil
}

func normalizeTuneArgs(args []string) []string {
	normalized := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--tune" && i+1 < len(args) {
			if _, err := time.ParseDuration(args[i+1]); err == nil {
				normalized = append(normalized, "--tune="+args[i+1])
				i++
				continue
			}
		}
		normalized = append(normalized, args[i])
	}
	return normalized
}

func parseAnimationMode(value string) (int, error) {
	value = strings.ToLower(value)
	for index, mode := range animationModes {
		if value == strings.ToLower(mode.name) {
			return index, nil
		}
		for _, alias := range mode.aliases {
			if value == strings.ToLower(alias) {
				return index, nil
			}
		}
	}
	if index, err := strconv.Atoi(value); err == nil && index >= 0 && index < len(animationModes) {
		return index, nil
	}
	names := make([]string, 0, len(animationModes))
	for _, mode := range animationModes {
		names = append(names, strings.ToLower(mode.name))
	}
	return 0, fmt.Errorf("want one of %s", strings.Join(names, ", "))
}

func parsePaintMode(value string) (int, error) {
	switch strings.ToLower(value) {
	case "rows", "row", "0":
		return 0, nil
	case "spans", "span", "1":
		return 1, nil
	case "cells", "cell", "2":
		return 2, nil
	default:
		return 0, fmt.Errorf("want rows, spans, or cells")
	}
}

func parseBackgroundMode(value string) (int, error) {
	switch strings.ToLower(value) {
	case "fill", "row-fill", "0":
		return 0, nil
	case "clear", "background-clear", "1":
		return 1, nil
	default:
		return 0, fmt.Errorf("want fill or clear")
	}
}

func parseClearMode(value string) (int, error) {
	switch strings.ToLower(value) {
	case "0", "off", "false", "no":
		return 0, nil
	case "1", "on", "true", "yes":
		return 1, nil
	default:
		if mode, err := strconv.Atoi(value); err == nil && (mode == 0 || mode == 1) {
			return mode, nil
		}
		return 0, fmt.Errorf("want 0 or 1")
	}
}

func (o cliOptions) apply(d *demo) {
	d.fps = o.fps
	d.areaWidth = o.areaWidth
	d.areaHeight = o.areaHeight
	d.animation = o.animation
	d.altForced = o.altScreenSet
	d.incremental = o.incremental
	if o.paintMode >= 0 {
		d.paintMode = o.paintMode
	}
	if o.background >= 0 {
		d.background = o.background
	}
	if o.clearEach >= 0 {
		d.clearEach = o.clearEach == 1
	}
}
