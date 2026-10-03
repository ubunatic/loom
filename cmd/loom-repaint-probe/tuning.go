// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type tuneConfig struct {
	paintMode   int
	background  int
	clearEach   bool
	animation   int
	alternate   bool
	incremental bool
}

type tuneConstraints struct {
	paint       bool
	background  bool
	clear       bool
	incremental bool
}

type tuneResult struct {
	config      tuneConfig
	frames      uint64
	elapsed     time.Duration
	measuredFPS float64
	compose1s   time.Duration
	write1s     time.Duration
	composeAvg  time.Duration
	writeAvg    time.Duration
	writeMax    time.Duration
	rowsLast    int
	rowsTotal   int
	rowsAvg     float64
	lastBytes   int
	valid       bool
}

type tuneMetadata struct {
	terminalSizes  []string
	initial        tuneConfig
	width          int
	height         int
	terminalWidth  int
	terminalHeight int
	fps            int
	duration       time.Duration
	forceAlt       bool
	constraints    tuneConstraints
	configCount    int
}

func runTune(app *demo, duration time.Duration, forceAlt bool, constraints tuneConstraints, keys <-chan int, resizes <-chan os.Signal, interrupts <-chan os.Signal, restoreRaw func(), outPath string) error {
	base := tuneConfig{
		paintMode: app.paintMode, background: app.background, clearEach: app.clearEach,
		animation: app.animation, alternate: app.alternate, incremental: app.incremental,
	}
	configs := tuneConfigs(base, forceAlt, constraints)
	metadata := tuneMetadata{
		terminalSizes: []string{fmt.Sprintf("%dx%d", app.width, app.height)},
		initial:       base, width: app.areaWidth, height: app.areaHeight,
		terminalWidth: app.width, terminalHeight: app.height,
		fps: app.fps, duration: duration, forceAlt: forceAlt,
		constraints: constraints, configCount: len(configs),
	}
	results := make([]tuneResult, 0, len(configs))
	interval := time.Second / time.Duration(app.fps)
	var runErr error
	interrupted := false
	for _, config := range configs {
		app.applyTuneConfig(config)
		ticker := time.NewTicker(interval)
		timer := time.NewTimer(duration)
		active := true
		for active {
			select {
			case <-ticker.C:
				if err := app.paint(); err != nil {
					runErr = err
					active = false
					interrupted = true
				}
			case <-timer.C:
				active = false
			case <-resizes:
				if err := app.resize(); err != nil {
					runErr = err
					active = false
					interrupted = true
				} else {
					size := fmt.Sprintf("%dx%d", app.width, app.height)
					if metadata.terminalSizes[len(metadata.terminalSizes)-1] != size {
						metadata.terminalSizes = append(metadata.terminalSizes, size)
					}
				}
			case <-interrupts:
				active = false
				interrupted = true
			case key := <-keys:
				if key == 'q' || key == 'Q' || key == 3 || key == keyF10 {
					active = false
					interrupted = true
				}
			}
		}
		ticker.Stop()
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		results = append(results, app.tuneResult(config, time.Now()))
		if interrupted {
			break
		}
	}
	if app.alternate {
		app.setAlternate(false)
	}
	restoreRaw()
	_, _ = fmt.Fprint(os.Stdout, "\x1b[0m\x1b[?25h\r\n")
	summary := formatTuneSummary(results, metadata, interrupted)
	_, _ = fmt.Fprint(os.Stdout, summary)
	if outPath != "" {
		target, err := resolveTuneOutputPath(outPath, metadata)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create tune results directory %q: %w", filepath.Dir(target), err)
		}
		if err := os.WriteFile(target, []byte(summary), 0o644); err != nil {
			return fmt.Errorf("write tune results to %q: %w", target, err)
		}
		_, _ = fmt.Fprintf(os.Stdout, "Saved tune report to %s\n", target)
	}
	return runErr
}

func tuneConfigs(base tuneConfig, forceAlt bool, constraints tuneConstraints) []tuneConfig {
	configs := []tuneConfig{base}
	if !constraints.paint {
		for mode := 0; mode < paintModeCount; mode++ {
			if mode == base.paintMode {
				continue
			}
			candidate := base
			candidate.paintMode = mode
			configs = append(configs, candidate)
		}
	}
	if !constraints.background {
		for mode := 0; mode < backgroundModeCount; mode++ {
			if mode == base.background {
				continue
			}
			candidate := base
			candidate.background = mode
			configs = append(configs, candidate)
		}
	}
	if !constraints.clear {
		for _, clear := range []bool{false, true} {
			if clear == base.clearEach {
				continue
			}
			candidate := base
			candidate.clearEach = clear
			configs = append(configs, candidate)
		}
	}
	if !constraints.incremental {
		candidate := base
		candidate.incremental = !base.incremental
		configs = append(configs, candidate)
	}
	if !forceAlt {
		candidate := base
		candidate.alternate = !base.alternate
		configs = append(configs, candidate)
	}
	return configs
}

func (d *demo) applyTuneConfig(config tuneConfig) {
	d.paintMode = config.paintMode
	d.background = config.background
	d.clearEach = config.clearEach
	d.animation = config.animation
	d.incremental = config.incremental
	d.setAlternate(config.alternate)
	d.resetMetrics()
}

func (d *demo) tuneResult(config tuneConfig, now time.Time) tuneResult {
	compose1s, write1s := d.movingAverage(now)
	result := tuneResult{
		config: config, frames: d.frames, elapsed: now.Sub(d.started), compose1s: compose1s, write1s: write1s,
		writeMax: d.maxWrite, rowsLast: d.updatedRows, rowsTotal: d.height, rowsAvg: d.rowUpdateAverage(now),
		lastBytes: d.lastBytes, valid: d.frames > 0,
	}
	if d.frames > 0 {
		result.measuredFPS = float64(d.frames) / max(result.elapsed.Seconds(), 0.001)
		result.composeAvg = d.totalBuild / time.Duration(d.frames)
		result.writeAvg = d.totalWrite / time.Duration(d.frames)
	}
	return result
}

func resolveTuneOutputPath(path string, metadata tuneMetadata) (string, error) {
	info, statErr := os.Stat(path)
	if statErr != nil && !os.IsNotExist(statErr) {
		return "", fmt.Errorf("inspect tune output path %q: %w", path, statErr)
	}
	directory := statErr == nil && info.IsDir() || strings.HasSuffix(path, string(filepath.Separator)) || filepath.Ext(path) == ""
	if directory {
		filename := fmt.Sprintf("tune-terminal-%dx%d-rect-%dx%d-%dfps.txt",
			metadata.terminalWidth, metadata.terminalHeight, metadata.width, metadata.height, metadata.fps)
		return filepath.Join(path, filename), nil
	}
	return path, nil
}

func formatTuneSummary(results []tuneResult, metadata tuneMetadata, interrupted bool) string {
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].valid != results[j].valid {
			return results[i].valid
		}
		return results[i].compose1s+results[i].write1s < results[j].compose1s+results[j].write1s
	})
	var summary strings.Builder
	fmt.Fprintf(&summary, "Terminal repaint benchmark\n")
	fmt.Fprintf(&summary, "Terminal sizes: %s\n", strings.Join(metadata.terminalSizes, " -> "))
	fmt.Fprintf(&summary, "Requested rectangle: %dx%d (clamped to terminal size)\n", metadata.width, metadata.height)
	fmt.Fprintf(&summary, "Target FPS: %d\n", metadata.fps)
	fmt.Fprintf(&summary, "Duration per configuration: %s\n", metadata.duration)
	fmt.Fprintf(&summary, "Animation: %s\n", animationName(metadata.initial.animation))
	fmt.Fprintf(&summary, "Initial modes: %s\n", metadata.initial.String())
	fmt.Fprintf(&summary, "Pinned options: %s\n", pinnedTuneOptions(metadata))
	fmt.Fprintf(&summary, "Configurations: %d\n", metadata.configCount)
	fmt.Fprintf(&summary, "Ranked tune results (1s compose + write average):\n")
	for index, result := range results {
		if !result.valid {
			fmt.Fprintf(&summary, "%2d. %s | no frames recorded\n", index+1, result.config.String())
			continue
		}
		fmt.Fprintf(&summary, "%2d. %s | measured %.1f FPS  elapsed %s  compose %s (avg %s)  write %s (avg %s, max %s)  rows %d/%d (1s avg %.0f%%)  frames %d  bytes %s\n",
			index+1, result.config.String(), result.measuredFPS, result.elapsed.Truncate(time.Millisecond),
			formatMillis(result.compose1s), formatMillis(result.composeAvg),
			formatMillis(result.write1s), formatMillis(result.writeAvg), formatMillis(result.writeMax),
			result.rowsLast, result.rowsTotal, result.rowsAvg,
			result.frames, formatBytes(result.lastBytes))
	}
	if interrupted {
		fmt.Fprintln(&summary, "Tune stopped early.")
	}
	return summary.String()
}

func pinnedTuneOptions(metadata tuneMetadata) string {
	var pinned []string
	if metadata.constraints.paint {
		pinned = append(pinned, "paint")
	}
	if metadata.constraints.background {
		pinned = append(pinned, "background")
	}
	if metadata.constraints.clear {
		pinned = append(pinned, "clear")
	}
	if metadata.constraints.incremental {
		pinned = append(pinned, "incremental")
	}
	if metadata.forceAlt {
		pinned = append(pinned, "altscreen")
	}
	if len(pinned) == 0 {
		return "none"
	}
	return strings.Join(pinned, ", ")
}

func (config tuneConfig) String() string {
	clear := 0
	if config.clearEach {
		clear = 1
	}
	alt := 0
	if config.alternate {
		alt = 1
	}
	return fmt.Sprintf("paint=%s background=%s clear=%d anim=%s altscreen=%d incremental=%t",
		paintModeName(config.paintMode), backgroundModeName(config.background), clear,
		animationName(config.animation), alt, config.incremental)
}
