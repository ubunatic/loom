// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

const (
	minimumWidth  = 160
	minimumHeight = 67
)

const (
	keyPageUp = 256 + iota
	keyPageDown
	keyF10
)

const (
	animationNone = iota
	animationAstra
	animationRainbow
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	options, err := parseOptions(os.Args[1:])
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("repaint demo needs a terminal on stdin and stdout")
	}
	original, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("enable raw input: %w", err)
	}
	rawMode := true
	restoreRaw := func() {
		if rawMode {
			_ = term.Restore(int(os.Stdin.Fd()), original)
			rawMode = false
		}
	}
	defer restoreRaw()

	app := newDemo(options.fps)
	options.apply(app)
	if err := app.resize(); err != nil {
		return err
	}

	_, _ = os.Stdout.WriteString("\x1b[?25l")
	defer os.Stdout.WriteString("\x1b[0m\x1b[?25h\n")
	defer func() {
		if app.alternate {
			_, _ = os.Stdout.WriteString("\x1b[?1049l")
		}
	}()
	if options.altScreen >= 0 {
		app.setAlternate(options.altScreen == 1)
	}

	keys := make(chan int, 16)
	go readKeys(keys)
	resizes := make(chan os.Signal, 1)
	signal.Notify(resizes, syscall.SIGWINCH)
	defer signal.Stop(resizes)
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupts)

	if options.tune.enabled {
		constraints := tuneConstraints{paint: options.paintSet, background: options.backgroundSet, clear: options.clearSet, incremental: options.incrementalSet}
		return runTune(app, options.tune.duration, options.altScreenSet, constraints, keys, resizes, interrupts, restoreRaw, options.outPath)
	}
	ticker := time.NewTicker(time.Second / time.Duration(app.fps))
	defer ticker.Stop()
	for {
		select {
		case key := <-keys:
			oldFPS := app.fps
			if !app.key(key) {
				return nil
			}
			if app.fps != oldFPS {
				ticker.Reset(time.Second / time.Duration(app.fps))
			}
		case <-resizes:
			if err := app.resize(); err != nil {
				return err
			}
		case <-interrupts:
			return nil
		case <-ticker.C:
			if !app.paused {
				if err := app.paint(); err != nil {
					return err
				}
			}
		}
	}
}

func readKeys(keys chan<- int) {
	var buffer [1]byte
	for {
		if _, err := os.Stdin.Read(buffer[:]); err != nil {
			return
		}
		key := int(buffer[0])
		if key == 0x1b {
			var introducer [1]byte
			if _, err := io.ReadFull(os.Stdin, introducer[:]); err != nil {
				return
			}
			if introducer[0] == '[' {
				var sequence strings.Builder
				for i := 0; i < 3; i++ {
					var next [1]byte
					if _, err := io.ReadFull(os.Stdin, next[:]); err != nil {
						return
					}
					if next[0] == '~' {
						switch sequence.String() {
						case "5":
							key = keyPageUp
						case "6":
							key = keyPageDown
						case "21":
							key = keyF10
						}
						break
					}
					if next[0] < '0' || next[0] > '9' {
						key = 0x1b
						break
					}
					sequence.WriteByte(next[0])
				}
			}
		}
		select {
		case keys <- key:
		default:
		}
	}
}

type demo struct {
	width         int
	height        int
	areaWidth     int
	areaHeight    int
	fps           int
	paintMode     int
	background    int
	clearEach     bool
	clearOnResize bool
	incremental   bool
	lastRows      [][]byte
	updatedRows   int
	alternate     bool
	altForced     bool
	paused        bool
	animation     int
	frames        uint64
	lastWrite     time.Duration
	lastCompose   time.Duration
	maxWrite      time.Duration
	totalWrite    time.Duration
	totalBuild    time.Duration
	lastBytes     int
	samples       []frameTiming
	started       time.Time
}

type renderSettings struct {
	fps         int
	areaWidth   int
	areaHeight  int
	paintMode   int
	background  int
	clearEach   bool
	alternate   bool
	paused      bool
	animation   int
	incremental bool
}

type frameTiming struct {
	at         time.Time
	compose    time.Duration
	write      time.Duration
	rowPercent float64
}

func newDemo(fps int) *demo {
	return &demo{fps: fps, areaWidth: minimumWidth, areaHeight: minimumHeight, incremental: true, animation: animationAstra, started: time.Now()}
}

func (d *demo) resize() error {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return fmt.Errorf("read terminal size: %w", err)
	}
	if width != d.width || height != d.height {
		d.width, d.height = width, height
		d.clearOnResize = true
		d.resetMetrics()
	}
	return nil
}

func (d *demo) key(key int) bool {
	before := d.renderSettings()
	switch key {
	case 'q', 'Q', 3, keyF10:
		return false
	case ' ':
		d.paused = !d.paused
	case '+', '=':
		d.areaWidth = min(d.width, d.areaWidth+16)
		d.areaHeight = min(d.height, d.areaHeight+6)
	case '-':
		d.areaWidth = max(16, d.areaWidth-16)
		d.areaHeight = max(8, d.areaHeight-6)
	case keyPageUp:
		d.fps = min(120, d.fps+5)
	case keyPageDown:
		d.fps = max(1, d.fps-5)
	case 'w':
		d.areaWidth = min(d.width, d.areaWidth+16)
	case 'W':
		d.areaWidth = max(16, d.areaWidth-16)
	case 'h':
		d.areaHeight = min(d.height, d.areaHeight+6)
	case 'H':
		d.areaHeight = max(8, d.areaHeight-6)
	case 'p', 'P':
		d.paintMode = (d.paintMode + 1) % paintModeCount
	case 'b', 'B':
		d.background = (d.background + 1) % backgroundModeCount
	case 'c', 'C':
		d.clearEach = !d.clearEach
	case 'i', 'I':
		d.incremental = !d.incremental
	case 'a', 'A':
		d.animation = (d.animation + 1) % len(animationModes)
	case 't', 'T':
		if !d.altForced {
			d.setAlternate(!d.alternate)
		}
	}
	if d.renderSettings() != before {
		d.resetMetrics()
	}
	return true
}

func (d *demo) setAlternate(enabled bool) {
	if d.alternate == enabled {
		return
	}
	d.alternate = enabled
	if enabled {
		_, _ = os.Stdout.WriteString("\x1b[?1049h")
	} else {
		_, _ = os.Stdout.WriteString("\x1b[?1049l")
	}
}

func (d *demo) renderSettings() renderSettings {
	return renderSettings{
		fps: d.fps, areaWidth: d.areaWidth, areaHeight: d.areaHeight,
		paintMode: d.paintMode, background: d.background, clearEach: d.clearEach,
		alternate: d.alternate, paused: d.paused, animation: d.animation, incremental: d.incremental,
	}
}

func (d *demo) resetMetrics() {
	d.frames = 0
	d.lastWrite = 0
	d.lastCompose = 0
	d.maxWrite = 0
	d.totalWrite = 0
	d.totalBuild = 0
	d.lastBytes = 0
	d.samples = nil
	d.lastRows = nil
	d.updatedRows = d.height
	d.started = time.Now()
}

func (d *demo) paint() error {
	composeStart := time.Now()
	frame := d.frame()
	d.lastBytes = len(frame)
	d.lastCompose = time.Since(composeStart)
	writeStart := time.Now()
	if d.incremental {
		if err := d.writeIncremental(frame); err != nil {
			return err
		}
	} else {
		for len(frame) > 0 {
			n, err := os.Stdout.Write(frame)
			if err != nil {
				return fmt.Errorf("write frame: %w", err)
			}
			frame = frame[n:]
		}
	}
	d.clearOnResize = false
	d.lastWrite = time.Since(writeStart)
	d.maxWrite = max(d.maxWrite, d.lastWrite)
	d.totalWrite += d.lastWrite
	d.totalBuild += d.lastCompose
	d.frames++
	rowPercent := 100.0
	if d.height > 0 {
		rowPercent = float64(d.updatedRows) * 100 / float64(d.height)
	}
	d.samples = append(d.samples, frameTiming{at: time.Now(), compose: d.lastCompose, write: d.lastWrite, rowPercent: rowPercent})
	return nil
}

func (d *demo) writeIncremental(frame []byte) error {
	prefix, rows, ok := splitFrameRows(frame, d.height)
	if !ok {
		d.updatedRows = d.height
		return writeAll(os.Stdout, frame)
	}
	full := d.lastRows == nil || bytes.Contains(prefix, []byte("\x1b[2J"))
	if err := writeAll(os.Stdout, prefix); err != nil {
		return err
	}
	updated := 0
	for index, row := range rows {
		if len(row) == 0 || !full && bytes.Equal(row, d.lastRows[index]) {
			continue
		}
		updated++
		if err := writeAll(os.Stdout, row); err != nil {
			return err
		}
	}
	if bytes.Contains(prefix, []byte("\x1b[2J")) {
		updated = d.height
	}
	d.updatedRows = updated
	d.lastRows = rows
	return nil
}

func splitFrameRows(frame []byte, height int) ([]byte, [][]byte, bool) {
	rows := make([][]byte, height)
	firstCursor := -1
	segmentStart := 0
	activeRow := -1
	for scan := 0; scan < len(frame); {
		marker := bytes.Index(frame[scan:], []byte("\x1b["))
		if marker < 0 {
			break
		}
		marker += scan
		row, end, ok := parseCursorRow(frame, marker)
		if !ok {
			scan = marker + 2
			continue
		}
		if firstCursor < 0 {
			firstCursor = marker
		} else if activeRow >= 0 && activeRow < len(rows) {
			rows[activeRow] = append(rows[activeRow], frame[segmentStart:marker]...)
		}
		activeRow = row - 1
		segmentStart = marker
		scan = end
	}
	if firstCursor < 0 {
		return nil, nil, false
	}
	if activeRow >= 0 && activeRow < len(rows) {
		rows[activeRow] = append(rows[activeRow], frame[segmentStart:]...)
	}
	return frame[:firstCursor], rows, true
}

func parseCursorRow(frame []byte, start int) (int, int, bool) {
	index := start + 2
	rowStart := index
	for index < len(frame) && frame[index] >= '0' && frame[index] <= '9' {
		index++
	}
	if index == rowStart || index >= len(frame) || frame[index] != ';' {
		return 0, 0, false
	}
	index++
	columnStart := index
	for index < len(frame) && frame[index] >= '0' && frame[index] <= '9' {
		index++
	}
	if index == columnStart || index >= len(frame) || frame[index] != 'H' {
		return 0, 0, false
	}
	row, err := strconv.Atoi(string(frame[rowStart : columnStart-1]))
	if err != nil {
		return 0, 0, false
	}
	return row, index + 1, true
}

func writeAll(out *os.File, content []byte) error {
	for len(content) > 0 {
		written, err := out.Write(content)
		if err != nil {
			return fmt.Errorf("write frame: %w", err)
		}
		content = content[written:]
	}
	return nil
}

func (d *demo) movingAverage(now time.Time) (time.Duration, time.Duration) {
	cutoff := now.Add(-time.Second)
	first := 0
	for first < len(d.samples) && d.samples[first].at.Before(cutoff) {
		first++
	}
	if first > 0 {
		copy(d.samples, d.samples[first:])
		d.samples = d.samples[:len(d.samples)-first]
	}
	if len(d.samples) == 0 {
		return 0, 0
	}
	var compose, write time.Duration
	for _, sample := range d.samples {
		compose += sample.compose
		write += sample.write
	}
	count := time.Duration(len(d.samples))
	return compose / count, write / count
}

func (d *demo) rowUpdateAverage(now time.Time) float64 {
	cutoff := now.Add(-time.Second)
	var total float64
	var count int
	for _, sample := range d.samples {
		if !sample.at.Before(cutoff) {
			total += sample.rowPercent
			count++
		}
	}
	if count == 0 {
		return 100
	}
	return total / float64(count)
}

const (
	paintModeCount      = 3
	backgroundModeCount = 2
)

func paintModeName(mode int) string {
	switch mode {
	case 1:
		return "spans"
	case 2:
		return "cells"
	default:
		return "rows"
	}
}

func backgroundModeName(mode int) string {
	if mode == 1 {
		return "background + clear"
	}
	return "row fill"
}

func (d *demo) frame() []byte {
	var out strings.Builder
	if d.clearEach || d.clearOnResize {
		out.WriteString("\x1b[48;2;0;0;139m\x1b[2J")
	}
	areaWidth := min(d.areaWidth, d.width)
	areaHeight := min(d.areaHeight, d.height)
	left := (d.width - areaWidth) / 2
	top := (d.height - areaHeight) / 2
	for y := 0; y < d.height; y++ {
		insideRow := y >= top && y < top+areaHeight
		if d.background == 1 && !insideRow {
			continue
		}
		if d.paintMode == 2 {
			start, end := 0, d.width
			if d.background == 1 {
				start, end = left, left+areaWidth
			}
			for x := start; x < end; x++ {
				inside := insideRow && x >= left && x < left+areaWidth
				out.WriteString(cursor(y+1, x+1))
				out.WriteString(cellColor(inside))
				out.WriteByte(' ')
			}
			continue
		}
		if d.paintMode == 1 {
			if d.background == 0 {
				writeSpan(&out, y, 0, left, false)
			}
			writeSpan(&out, y, left, left+areaWidth, insideRow)
			if d.background == 0 {
				writeSpan(&out, y, left+areaWidth, d.width, false)
			}
			continue
		}
		if d.background == 1 {
			if insideRow {
				out.WriteString(cursor(y+1, left+1))
				out.WriteString(cellColor(true))
				out.WriteString(strings.Repeat(" ", areaWidth))
			}
		} else {
			out.WriteString(cursor(y+1, 1))
			writeCells(&out, d.width, left, left+areaWidth, insideRow)
		}
	}
	now := time.Now()
	windowCompose, windowWrite := d.movingAverage(now)
	windowRows := d.rowUpdateAverage(now)
	if mode := animationMode(d.animation); mode.paint != nil {
		mode.paint(&out, left, top, areaWidth, areaHeight, now)
	}
	if d.width >= 60 && d.height >= 8 {
		elapsed := time.Since(d.started)
		measuredFPS := float64(d.frames) / max(elapsed.Seconds(), 0.001)
		averageWrite := time.Duration(0)
		averageBuild := time.Duration(0)
		clearStatus := "yes"
		if !d.clearEach {
			clearStatus = "no; outside retains previous contents"
		}
		incrementalStatus := "off"
		if d.incremental {
			incrementalStatus = fmt.Sprintf("on %d/%d rows (1s avg %.0f%%)", d.updatedRows, d.height, windowRows)
		}
		if d.frames > 0 {
			averageWrite = d.totalWrite / time.Duration(d.frames)
			averageBuild = d.totalBuild / time.Duration(d.frames)
		}
		altHint := "t alt screen"
		if d.altForced {
			altHint = "alt screen forced"
		}
		hud := []string{
			fmt.Sprintf("Terminal repaint probe  |  target %dx%d  |  area %dx%d at %d,%d", d.width, d.height, areaWidth, areaHeight, left, top),
			fmt.Sprintf("FPS target %d  measured %.1f  frames %d  elapsed %s", d.fps, measuredFPS, d.frames, elapsed.Truncate(time.Second)),
			fmt.Sprintf("compose: %s (avg %s)  write %s (avg %s, max %s)  bytes %s", formatMillis(windowCompose), formatMillis(averageBuild), formatMillis(windowWrite), formatMillis(averageWrite), formatMillis(d.maxWrite), formatBytes(d.lastBytes)),
			fmt.Sprintf("paint: %s  background: %s", paintModeName(d.paintMode), backgroundModeName(d.background)),
			fmt.Sprintf("clear each frame: %s", clearStatus),
			fmt.Sprintf("incremental: %s  alt screen: %t  paused: %t  animation: %s", incrementalStatus, d.alternate, d.paused, animationName(d.animation)),
			"q quit | space pause | PgUp/PgDn FPS | +/- area | w/W width | h/H height",
			fmt.Sprintf("p paint | b background | c clear | i incremental | a animation | %s", altHint),
		}
		for row, line := range hud {
			out.WriteString(cursor(top+row+1, left+1))
			for column := 0; column < len(line); column++ {
				bgR, bgG, bgB := animationBackground(d.animation, column, areaWidth, now)
				bgR, bgG, bgB = bgR/2, bgG/2, bgB/2
				fg := uint8(230)
				luminance := (uint32(bgR)*299 + uint32(bgG)*587 + uint32(bgB)*114) / 1000
				if luminance > 145 {
					fg = 20
				}
				out.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm", fg, fg, fg, bgR, bgG, bgB))
				out.WriteByte(line[column])
			}
		}
	}
	out.WriteString("\x1b[0m")
	return []byte(out.String())
}

func cursor(row, column int) string { return fmt.Sprintf("\x1b[%d;%dH", row, column) }

func formatMillis(value time.Duration) string {
	millis := max(float64(value)/float64(time.Millisecond), 0.1)
	return fmt.Sprintf("%.1fms", millis)
}

func formatBytes(value int) string {
	if value < 1024 {
		return fmt.Sprintf("%d", value)
	}
	return fmt.Sprintf("%dk", value/1024)
}

func cellColor(inside bool) string {
	if inside {
		return "\x1b[48;2;64;64;64m"
	}
	return "\x1b[48;2;0;0;139m"
}

func writeSpan(out *strings.Builder, row, start, end int, inside bool) {
	if start >= end {
		return
	}
	out.WriteString(cursor(row+1, start+1))
	out.WriteString(cellColor(inside))
	out.WriteString(strings.Repeat(" ", end-start))
}

func writeCells(out *strings.Builder, width, left, right int, insideRow bool) {
	lastColor := ""
	for x := 0; x < width; x++ {
		inside := insideRow && x >= left && x < right
		color := cellColor(inside)
		if color != lastColor {
			out.WriteString(color)
			lastColor = color
		}
		out.WriteByte(' ')
	}
}
