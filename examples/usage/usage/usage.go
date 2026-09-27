// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package usage is a small Loom rebuild of the compact Harnez usage watch.
package usage

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/graph"
)

const (
	defaultCollectInterval = time.Second
	defaultRedrawInterval  = 100 * time.Millisecond
)

// Options configures collection and redraw cadence for a hosted widget.
type Options struct {
	Source          DataSource
	CollectInterval time.Duration
	RedrawInterval  time.Duration
}

type usageWidget struct {
	mu           sync.RWMutex
	snapshotData Snapshot
	source       DataSource
	collect      time.Duration
	redraw       time.Duration
	cancel       context.CancelFunc
	done         chan struct{}
	frame        *loom.Frame
	rows         [2]*loom.StyledRows
	invalidateMu sync.RWMutex
	invalidate   func()
	closeOnce    sync.Once
}

var _ loom.Widget = (*usageWidget)(nil)
var _ loom.Ticker = (*usageWidget)(nil)
var _ loom.InvalidationAware = (*usageWidget)(nil)

// NewWidget parses the standalone command's options and builds its root widget.
func NewWidget(args []string) (loom.Widget, error) {
	flags := flag.NewFlagSet("usage", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	collect := flags.Duration("collect", defaultCollectInterval, "local data collection cadence")
	redraw := flags.Duration("redraw", defaultRedrawInterval, "independent pane redraw cadence")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	if flags.NArg() != 0 {
		return nil, fmt.Errorf("usage: unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	return NewWidgetFromOptions(Options{Source: &localSource{}, CollectInterval: *collect, RedrawInterval: *redraw})
}

// NewWidgetFromOptions creates a hostable widget with an injectable source.
func NewWidgetFromOptions(opts Options) (loom.Widget, error) {
	if opts.CollectInterval <= 0 {
		return nil, fmt.Errorf("usage: collection interval must be positive")
	}
	if opts.RedrawInterval <= 0 {
		return nil, fmt.Errorf("usage: redraw interval must be positive")
	}
	if opts.Source == nil {
		opts.Source = &localSource{}
	}
	u := &usageWidget{source: opts.Source, collect: opts.CollectInterval, redraw: opts.RedrawInterval, done: make(chan struct{})}
	u.frame, u.rows = makeFrame()
	ctx, cancel := context.WithCancel(context.Background())
	u.cancel = cancel
	go u.collectLoop(ctx)
	return u, nil
}

func makeFrame() (*loom.Frame, [2]*loom.StyledRows) {
	usageRows := loom.NewStyledRows("Collecting usage…")
	loadRows := loom.NewStyledRows("Reading local CPU and memory…")
	border := loom.BoxBorder{TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘", Horizontal: "─", Vertical: "│", TitlePrefix: " ", TitleSuffix: " "}
	cyan := loom.Style{FG: loom.ColorRGB(80, 190, 230)}
	frame := &loom.Frame{
		Title:      "Loom Usage",
		Status:     "Live local Load · deterministic All Usage · q quit",
		Gap:        2,
		Breakpoint: 96,
		Style:      loom.FrameStyle{Title: loom.Style{FG: loom.ColorRGB(100, 210, 245), Bold: true}, Status: loom.Style{FG: loom.ColorRGB(145, 155, 170)}},
		Boxes: []loom.Box{
			{ID: "all-usage", Title: "All Usage", Width: 48, Height: 7, MinWidth: 26, Padding: 1, Border: border, Style: loom.BoxStyle{Border: cyan, Title: loom.Style{FG: loom.ColorRGB(185, 125, 255), Bold: true}}, Child: usageRows},
			{ID: "load", Title: "Load", Width: 48, Height: 7, MinWidth: 26, Padding: 1, Border: border, Style: loom.BoxStyle{Border: cyan, Title: loom.Style{FG: loom.ColorRGB(80, 220, 150), Bold: true}}, Child: loadRows},
		},
	}
	return frame, [2]*loom.StyledRows{usageRows, loadRows}
}

func (u *usageWidget) collectLoop(ctx context.Context) {
	defer close(u.done)
	ticker := time.NewTicker(u.collect)
	defer ticker.Stop()
	for {
		if snapshot, err := u.source.Collect(ctx, time.Now()); err == nil {
			snapshot.CapturedAt = time.Now()
			u.mu.Lock()
			u.snapshotData = cloneSnapshot(snapshot)
			u.mu.Unlock()
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func cloneSnapshot(snapshot Snapshot) Snapshot {
	snapshot.Quotas = append([]Quota(nil), snapshot.Quotas...)
	snapshot.CPUHistory = append([]float64(nil), snapshot.CPUHistory...)
	return snapshot
}

func (u *usageWidget) snapshot() Snapshot {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return cloneSnapshot(u.snapshotData)
}

func (u *usageWidget) Draw(c *loom.Canvas, r loom.Rect) {
	snapshot := u.snapshot()
	u.rows[0].Lines = usageLines(snapshot)
	u.rows[1].Lines = loadLines(snapshot)
	u.frame.Draw(c, r)
}

func (u *usageWidget) HandleKey(e loom.KeyEvent) bool {
	if e.Rune() == 'q' || e.Rune() == 'Q' || e.Is("ctrl-c", "ctrl-q", "esc") {
		return true
	}
	return false
}

func (*usageWidget) HandleMouse(loom.MouseEvent) bool { return false }
func (u *usageWidget) TickInterval() time.Duration    { return u.redraw }
func (*usageWidget) Tick(time.Time)                   {}

// SetInvalidate accepts the pane's redraw callback for asynchronous updates.
func (u *usageWidget) SetInvalidate(invalidate func()) {
	u.invalidateMu.Lock()
	u.invalidate = invalidate
	u.invalidateMu.Unlock()
}

func (u *usageWidget) Close() {
	u.closeOnce.Do(func() {
		u.cancel()
		<-u.done
	})
}

func (u *usageWidget) PaneRequest() loom.PaneRequest {
	return loom.PaneRequest{Resizeable: true}
}

func usageLines(snapshot Snapshot) []string {
	if len(snapshot.Quotas) == 0 {
		return []string{dim("no quota windows available")}
	}
	lines := make([]string, 0, len(snapshot.Quotas))
	for _, quota := range snapshot.Quotas {
		bar := graph.RenderProgressBar(quota.Used, 8)
		lines = append(lines, fmt.Sprintf("%-8s %s %s %s", quota.Name, heat(bar, quota.Used), heat(fmt.Sprintf("%3.0f%%", quota.Used), quota.Used), dim("reset "+quota.ResetIn)))
	}
	return lines
}

func loadLines(snapshot Snapshot) []string {
	if !snapshot.CapturedAt.IsZero() && !snapshot.MemoryOK && !snapshot.CPUOK {
		return []string{dim("local load unavailable")}
	}
	cpu := "n/a"
	if snapshot.CPUOK {
		cpu = fmt.Sprintf("%2.0f%%", snapshot.CPUPercent)
	}
	lines := []string{fmt.Sprintf("CPU %s %s · load %.2f · %d cores", heat(graph.RenderProgressBar(snapshot.CPUPercent, 8), snapshot.CPUPercent), heat(cpu, snapshot.CPUPercent), snapshot.Load1, snapshot.NumCPU)}
	if snapshot.MemoryOK {
		pct := 100 * snapshot.MemoryUsed / snapshot.MemoryTotal
		lines = append(lines, fmt.Sprintf("RAM %s %s/%s GiB %s", heat(graph.RenderProgressBar(pct, 8), pct), humanGiB(snapshot.MemoryUsed), humanGiB(snapshot.MemoryTotal), heat(fmt.Sprintf("%.0f%%", pct), pct)))
	} else {
		lines = append(lines, dim("RAM unavailable"))
	}
	return lines
}

func humanGiB(value float64) string { return fmt.Sprintf("%.1f", value) }

func heat(value string, percent float64) string {
	code := 34
	switch {
	case percent > 75:
		code = 31
	case percent > 50:
		code = 33
	case percent > 25:
		code = 32
	}
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", code, value)
}

func dim(value string) string { return "\x1b[90m" + value + "\x1b[0m" }

func Run(args []string) error {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		_, err := fmt.Fprintln(os.Stdout, "usage: loom-usage [--collect duration] [--redraw duration]")
		return err
	}
	widget, err := NewWidget(args)
	if err != nil {
		return err
	}
	defer widget.(*usageWidget).Close()
	pane, err := loom.New(12)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	return pane.Run(widget)
}
