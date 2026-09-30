// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package monitor prints the embedded static shell once, without opening a TTY.
package monitor

import (
	"bytes"
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/graph"
	"github.com/spf13/cobra"
)

//go:embed spec/*.yaml
var documents embed.FS

func run(out io.Writer) error {
	return runWidth(out, 0)
}

func runWidth(out io.Writer, width int) error {
	document, err := documents.Open("spec/monitor.yaml")
	if err != nil {
		return err
	}
	defer document.Close()
	root, cfg, err := loom.BuildWidget(document)
	if err != nil {
		return err
	}
	observed := terminalCols()
	if width == 0 {
		width = outputWidth(out, cfg.MaxWidth())
	} else {
		width = min(width, cfg.MaxWidth())
	}
	if frame, ok := root.(*loom.Frame); ok {
		observedText := "n/a"
		if observed > 0 {
			observedText = fmt.Sprintf("%d", observed)
		}
		frame.Title = fmt.Sprintf("%s (observed: %s, effective: %d)", frame.Title, observedText, width)
	}
	applySnapshot(root, staticSnapshot)
	height := cfg.Height(0)
	if responsive, ok := root.(loom.WidthHeighter); ok {
		height = responsive.HeightForWidth(width)
	}
	cols := terminalCols()
	for _, row := range loom.Render(root, width, height) {
		// This monochrome shell has no styles; omit Render's row reset so
		// redirected output is plain terminal text, with no cursor controls.
		row = strings.TrimSuffix(row, "\x1b[0m")
		// Belt-and-suspenders against a stale/wrong terminal-width detection
		// (see loom.RawScreen's doc comment for the failure mode this
		// guards): row should already be exactly `width` columns since width
		// was computed from the same terminal query above, but this is the
		// one place that safety belongs -- not in loom.Render/the widget
		// tree itself. Plain sequential output, no cursor control: this is a
		// one-shot "print once" path, not a redraw loop.
		if cols > 0 {
			row = loom.ClipRow(row, cols)
		}
		if _, err := fmt.Fprintln(out, row); err != nil {
			return err
		}
	}
	return nil
}

func outputWidth(out io.Writer, fallback int) int {
	if cols := terminalCols(); cols > 0 {
		return min(cols, fallback)
	}
	return fallback
}

func terminalCols() int {
	cols, _, err := loom.TerminalSize()
	if err != nil || cols < 1 {
		return 0
	}
	return cols
}

// monitorSnapshot is application data: the declaration owns the row layout,
// while the example owns the numerical values that populate graph columns.
type monitorSnapshot struct {
	timestamp time.Time
	usage     []float64
	usage2    []float64
	load      map[string][]float64
	vram      []float64
	gtt       []float64
}

// Options configures the standalone command and hosted monitor widget.
type Options struct {
	Watch bool
	Width int
}

type monitorWidget struct {
	frame        *loom.Frame
	spec         watchSpec
	baseTitle    string
	title        *template.Template
	state        *monitorState
	active       bool
	focusManaged bool
	runtime      *sourceRuntime
}

func (m *monitorWidget) Draw(c *loom.Canvas, r loom.Rect) {
	if !m.focusManaged {
		m.SetFocus(true)
	}
	m.frame.Draw(c, r)
}
func (m *monitorWidget) ConsumeKey(e loom.KeyEvent) loom.EventResult { return m.frame.ConsumeKey(e) }
func (m *monitorWidget) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	return m.frame.ConsumeMouse(e)
}
func (m *monitorWidget) TickInterval() time.Duration {
	if !m.active || !m.spec.watchEnabled {
		return 0
	}
	return m.spec.Redraw
}
func (m *monitorWidget) Tick(now time.Time) {
	if !m.active || !m.spec.watchEnabled {
		return
	}
	m.state.SampleAt(now)
	if m.runtime != nil {
		if err := m.runtime.Err(); err == nil {
			for _, src := range m.spec.sources {
				records := src.history.Snapshot()
				if len(records) >= 2 {
					a, ea := parseProcStat(records[len(records)-2].Data)
					b, eb := parseProcStat(records[len(records)-1].Data)
					if ea == nil && eb == nil {
						m.state.metrics.Publish("cpu (16c)", cpuPercentage(a, b), now)
					}
				}
			}
		}
	}
	applySnapshot(m.frame, m.state.Snapshot())
	var b bytes.Buffer
	if m.title.Execute(&b, struct{ Title, Time string }{m.baseTitle, now.Format(m.spec.ClockFormat)}) == nil {
		m.frame.Title = b.String()
	}
}
func (m *monitorWidget) Focused() bool { return m.active }
func (m *monitorWidget) SetFocus(focused bool) {
	m.focusManaged = true
	if m.active == focused {
		return
	}
	m.active = focused
	if focused && m.spec.watchEnabled {
		m.runtime = startSources(context.Background(), m.spec.sources)
	}
	if !focused && m.runtime != nil {
		m.runtime.Close()
		m.runtime = nil
	}
}
func (m *monitorWidget) Close() {
	m.active = false
	if m.runtime != nil {
		m.runtime.Close()
		m.runtime = nil
	}
}
func (m *monitorWidget) PaneRequest() loom.PaneRequest {
	return loom.PaneRequest{Resizeable: true, MaxCols: 80}
}

// NewWidget parses monitor flags and constructs the corresponding root widget.
func NewWidget(args []string) (loom.Widget, error) {
	var opts Options
	flags := flag.NewFlagSet("monitor", flag.ContinueOnError)
	flags.BoolVar(&opts.Watch, "watch", false, "collect live monitor data")
	flags.IntVar(&opts.Width, "width", 0, "explicit terminal width for show-once")
	flags.IntVar(&opts.Width, "w", 0, "explicit terminal width for show-once")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	if flags.NArg() > 0 {
		return nil, fmt.Errorf("monitor: unexpected arguments: %v", flags.Args())
	}
	if opts.Width < 0 {
		return nil, fmt.Errorf("width must be positive")
	}
	return NewWidgetFromOptions(opts)
}

// NewWidgetFromOptions constructs the declaration-backed monitor widget.
func NewWidgetFromOptions(opts Options) (loom.Widget, error) {
	spec, err := loadWatch()
	if err != nil {
		return nil, err
	}
	spec.watchEnabled = opts.Watch
	return newMonitorWidget(spec)
}

func newMonitorWidget(spec watchSpec) (*monitorWidget, error) {
	data, err := documents.ReadFile("spec/monitor.yaml")
	if err != nil {
		return nil, err
	}
	root, _, err := loom.BuildWidget(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	frame, ok := root.(*loom.Frame)
	if !ok {
		return nil, fmt.Errorf("monitor: declared root must be a frame")
	}
	title, err := template.New("title").Option("missingkey=error").Parse(spec.Title)
	if err != nil {
		return nil, err
	}
	baseTitle := frame.Title
	frame.Status = spec.Status
	if len(spec.sources) > 0 {
		for i := range frame.Boxes {
			if frame.Boxes[i].ID == "load" {
				frame.Boxes[i].Footer = "(real collector data)"
			}
		}
	}
	m := &monitorWidget{frame: frame, spec: spec, baseTitle: baseTitle, title: title, state: newMonitorState(staticSnapshot, 32)}
	applySnapshot(frame, m.state.Snapshot())
	var initial bytes.Buffer
	if err := title.Execute(&initial, struct{ Title, Time string }{baseTitle, time.Now().Format(spec.ClockFormat)}); err != nil {
		return nil, err
	}
	frame.Title = initial.String()
	return m, nil
}

var staticSnapshot = monitorSnapshot{
	usage:  []float64{60, 95, 99, 41},
	usage2: []float64{3, 0, 0, 1},
	load: map[string][]float64{
		"cpu (16c)": {1, 4, 8, 12, 9, 14, 11, 7, 5, 3, 1, 2, 4, 6, 8, 10, 12, 10, 8, 6},
		"ram (45G)": {32, 34, 35, 36, 36, 37, 36, 36, 35, 36, 36, 37, 36, 36, 36, 36, 36, 36, 36, 36},
		"gpu (Phx)": {0, 3, 7, 4, 2, 0, 1, 5, 8, 4, 2, 0, 1, 3, 5, 2, 0, 1, 4, 2},
	},
	vram: []float64{4, 5, 6, 7, 6, 6, 7, 8},
	gtt:  []float64{1, 1, 2, 2, 2, 3, 2, 2},
}

func applySnapshot(root loom.Widget, snapshot monitorSnapshot) {
	frame, ok := root.(*loom.Frame)
	if !ok {
		return
	}
	for i := range frame.Boxes {
		box := &frame.Boxes[i]
		switch box.ID {
		case "usage":
			values := box.Rows.GetValues()
			for row := range values {
				if row < len(snapshot.usage) && len(values[row]) > 1 {
					values[row][1] = graph.RenderBar(snapshot.usage[row], graph.BarOptions{Width: 4, SubChar: true})
				}
				if row < len(snapshot.usage2) && len(values[row]) > 3 {
					values[row][3] = graph.RenderBar(snapshot.usage2[row], graph.BarOptions{Width: 4, SubChar: true})
				}
			}
			box.SetRowsValues(values)
		case "load":
			values := box.Rows.GetValues()
			for row := range values {
				if len(values[row]) < 2 {
					continue
				}
				name := values[row][0]
				history, ok := snapshot.load[name]
				if !ok {
					if name == "vram/gtt" {
						values[row][1] = splitTimeline(snapshot.vram, snapshot.gtt)
					}
					continue
				}
				values[row][1] = timeline(history, 10)
			}
			box.SetRowsValues(values)
		}
	}
}

func timeline(values []float64, width int) string {
	return "[" + graph.RenderSparkline(values, graph.SparklineOptions{
		Width: width, FixedRange: true, Min: 0, Max: 100,
	}) + "]"
}

func splitTimeline(left, right []float64) string {
	return timeline(left, 4) + timeline(right, 4)
}

// Run runs the monitor example with the given command-line args, writing to
// stdout, matching the Run(args []string) error signature shared by the
// other examples for loom-demo/loom-bench registration.
func Run(args []string) error {
	return Execute(context.Background(), args, os.Stdout)
}

// Execute runs the monitor example's cobra command against ctx/args/out, for
// callers that need explicit context and output control (e.g. tests).
func Execute(ctx context.Context, args []string, out io.Writer) error {
	spec, err := loadWatch()
	if err != nil {
		return err
	}
	var watch bool
	var width int
	cmd := &cobra.Command{Use: spec.Command, Short: spec.Description, Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Flags().Changed("width") && width <= 0 {
				return fmt.Errorf("width must be positive")
			}
			if watch && cmd.Flags().Changed("width") {
				return fmt.Errorf("--width is for show-once; watch uses terminal width")
			}
			if watch {
				return runWatch(cmd.Context(), spec)
			}
			return runWidth(out, width)
		},
	}
	cmd.Flags().BoolVar(&watch, "watch", false, spec.WatchHelp)
	cmd.Flags().IntVarP(&width, "width", "w", 0, spec.WidthHelp)
	cmd.SetArgs(args)
	cmd.SetOut(out)
	cmd.SetErr(out)
	return cmd.ExecuteContext(ctx)
}
