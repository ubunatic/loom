// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command treemap renders the live process CPU-usage tree (via `ps`) as a
// graph.RenderTreemap box layout, sized to the terminal by default. It is a
// thin, standalone demo of codeberg.org/ubunatic/loom/graph's
// AggregateTreemap + RenderTreemap: all process-tree reading lives here in
// the example, not in the dependency-free graph package.
//
// --watch keeps redrawing in place on an interval (like `watch ps`) until
// interrupted, instead of loom's declarative Pane/View widgets, so --ansi
// coloring survives (a loom.View strips inline ANSI from its lines). It
// still quits on loom's own default keys (q, Esc, Ctrl-C, Ctrl-Q, Ctrl-D)
// via a small raw-mode /dev/tty reader -- see watchForQuitKey.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/graph"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// quitKeys mirrors loom.Pane's own default fallback quit keys (q, Esc,
// Ctrl-C, Ctrl-Q, Ctrl-D; see pane.go's handleKeyFallback and
// spec/defaults.yaml's fallback_quit_keys) so --watch quits the same way
// any other loom-based pane does, even though it isn't built on a
// loom.Pane itself.
var quitKeys = func() map[string]bool {
	m := make(map[string]bool, len(loom.SpeccedDefaults.FallbackQuitKeys))
	for _, k := range loom.SpeccedDefaults.FallbackQuitKeys {
		m[k] = true
	}
	return m
}()

// watchForQuitKey opens /dev/tty in raw mode and calls cancel as soon as a
// quitKeys key is pressed. Uses a short poll timeout (rather than a bare
// blocking Read) so the returned cleanup can always stop the goroutine and
// restore the terminal, mirroring loom.Pane's own input-reader goroutine
// (see pane.go's run method) for the same reason: Close cannot interrupt a
// blocking Read on a detached tty fd.
//
// Returns a no-op cleanup when no controlling terminal is available (e.g.
// stdin/stdout redirected) so --watch degrades to signal-only quitting
// (Ctrl-C via the process's normal SIGINT handling) instead of failing.
func watchForQuitKey(cancel context.CancelFunc) (cleanup func()) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return func() {}
	}
	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		_ = tty.Close()
		return func() {}
	}
	done := make(chan struct{})
	go func() {
		pfd := []unix.PollFd{{Fd: int32(tty.Fd()), Events: unix.POLLIN}}
		buf := make([]byte, 64)
		for {
			select {
			case <-done:
				return
			default:
			}
			n, perr := unix.Poll(pfd, 200)
			if perr != nil {
				if perr == unix.EINTR {
					continue
				}
				return
			}
			if n == 0 {
				continue // timeout: re-check done
			}
			m, rerr := tty.Read(buf)
			if m > 0 {
				ke := loom.DecodeKey(buf[:m])
				key := ke.Key
				if key == "" {
					key = ke.Text
				}
				if quitKeys[key] {
					cancel()
					return
				}
			}
			if rerr != nil {
				return
			}
		}
	}()
	return func() {
		close(done)
		_ = term.Restore(int(tty.Fd()), oldState)
		_ = tty.Close()
	}
}

type procEntry struct {
	pid, ppid int
	pcpu      float64
	name      string
}

// readProcTree shells out to `ps -eo pid,ppid,pcpu,comm` and builds a
// graph.TreemapNode tree rooted at a synthetic "root" node.
func readProcTree(ctx context.Context) (graph.TreemapNode, error) {
	cmd := exec.CommandContext(ctx, "ps", "-eo", "pid,ppid,pcpu,comm", "--no-headers")
	out, err := cmd.Output()
	if err != nil {
		return graph.TreemapNode{}, fmt.Errorf("treemap: running ps: %w", err)
	}
	return buildProcTree(string(out)), nil
}

func buildProcTree(psOutput string) graph.TreemapNode {
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
			if childPID == pid {
				continue
			}
			n.Children = append(n.Children, build(childPID))
		}
		return n
	}
	root := graph.TreemapNode{Name: "root"}
	for _, p := range procs {
		if _, hasParent := byPID[p.ppid]; !hasParent || p.pid == p.ppid {
			root.Children = append(root.Children, build(p.pid))
		}
	}
	return root
}

func terminalSize(out *os.File) (width, height int, ok bool) {
	if !term.IsTerminal(int(out.Fd())) {
		return 0, 0, false
	}
	cols, rows, err := term.GetSize(int(out.Fd()))
	if err != nil || cols < 1 || rows < 1 {
		return 0, 0, false
	}
	return cols, rows, true
}

// resolveDimensions fills in any unset (<=0) width/height from the current
// terminal size, falling back to a fixed 80x24 when stdout isn't a
// terminal (e.g. redirected output).
//
// It also CLAMPS an explicit --width/--height down to the real terminal
// size when stdout is a terminal, warning on stderr when it does. This
// matters more than it looks: graph.RenderTreemap always produces exact
// Width-column rows, but if that Width exceeds the terminal's actual
// column count, the terminal's own auto-wrap (not a bug in RenderTreemap)
// wraps the overflow onto the next physical line -- and this program's own
// explicit newline then advances *again*, so every subsequent row lands
// one line lower than intended. The result looks exactly like scrambled,
// wrongly-offset boxes, cascading worse with every row: a real repro (see
// issue tracker) confirmed this exact failure with an over-wide --width on
// an 80-column PTY. A single-row bar can usually reflow harmlessly, but a
// multi-row grid like this one cannot -- each row's screen position is
// load-bearing, so this is the one thing that must never be allowed to
// silently overflow. When output isn't a terminal at all (redirected/
// piped), there's no PTY to auto-wrap against, so an explicit width is
// honored as-is.
func resolveDimensions(width, height int) (int, int) {
	cols, rows, ok := terminalSize(os.Stdout)
	if !ok {
		if width <= 0 {
			width = 80
		}
		if height <= 0 {
			height = 24
		}
		return width, height
	}
	w, h, widthClamped, heightClamped := clampDimensions(width, height, cols, rows)
	if widthClamped {
		fmt.Fprintf(os.Stderr, "treemap: --width %d exceeds the terminal's %d columns; clamping to avoid line-wrap corruption\n", width, cols)
	}
	if heightClamped {
		fmt.Fprintf(os.Stderr, "treemap: --height %d exceeds the usable terminal height %d; clamping\n", height, h)
	}
	return w, h
}

// clampDimensions applies resolveDimensions' terminal-size clamp given an
// already-known terminal size (cols, rows), separated out from the actual
// terminal query for testability. Reports whether an explicit width/height
// had to be clamped down, so the caller can warn.
func clampDimensions(width, height, cols, rows int) (w, h int, widthClamped, heightClamped bool) {
	switch {
	case width <= 0:
		w = cols
	case width > cols:
		w = cols
		widthClamped = true
	default:
		w = width
	}
	maxHeight := rows - 1 // always leave the cursor's own line free
	if maxHeight < 1 {
		maxHeight = 1
	}
	switch {
	case height <= 0:
		h = rows - 2 // leave room for the shell prompt
		if h < 1 {
			h = 1
		}
	case height > maxHeight:
		h = maxHeight
		heightClamped = true
	default:
		h = height
	}
	return w, h, widthClamped, heightClamped
}

// treemapTheme maps the user-facing --theme flag (1 or 2) to
// graph.TreemapTheme; validated by parseTheme before reaching here.
func treemapTheme(theme int) graph.TreemapTheme {
	switch theme {
	case 2:
		return graph.TreemapThemeBlocks
	case 3:
		return graph.TreemapThemeNumbered
	case 4:
		return graph.TreemapThemeNumberedFilled
	default:
		return graph.TreemapThemeClassic
	}
}

// parseTheme validates the --theme flag, matching graph's four themes: 1
// (TreemapThemeClassic, the default box-drawing style), 2
// (TreemapThemeBlocks, half-block boundary glyphs), 3 (TreemapThemeNumbered,
// seven-eighths-block boundary glyphs plus a corner number on every box,
// floating over the ambient background), and 4 (TreemapThemeNumberedFilled,
// same as 3 but the corner number is filled solid like ordinary label text)
// -- see issue 040. Themes 2-4 all require --ansi, since none of them have
// any color to render their edge glyphs with otherwise.
func parseTheme(theme int, ansi bool) error {
	if theme < 1 || theme > 4 {
		return fmt.Errorf("treemap: --theme must be 1, 2, 3, or 4, got %d", theme)
	}
	if theme != 1 && !ansi {
		return fmt.Errorf("treemap: --theme %d requires --ansi (it has no color to render its edges with)", theme)
	}
	return nil
}

func buildOptions(width, height, theme int, ansi, showValues bool) graph.TreemapOptions {
	opts := graph.TreemapOptions{
		Width: width, Height: height,
		Theme:      treemapTheme(theme),
		ShowValues: showValues, ValuePrecision: 0, ValueSuffix: "%",
		ANSI: ansi,
	}
	if ansi {
		opts.BackgroundANSI = []string{"41", "42", "43", "44", "45", "46", "100", "47"}
		opts.ForegroundANSI = []string{"97"}
	}
	return opts
}

// renderOnce reads a fresh process tree and renders it. Width/height are
// re-resolved against the current terminal size on every call, so watch
// mode picks up terminal resizes between redraws.
func renderOnce(ctx context.Context, width, height, maxNodes, theme int, ansi, showValues bool) ([]string, error) {
	root, err := readProcTree(ctx)
	if err != nil {
		return nil, err
	}
	w, h := resolveDimensions(width, height)
	segments := graph.AggregateTreemap(root, maxNodes)
	return graph.RenderTreemap(segments, buildOptions(w, h, theme, ansi, showValues)), nil
}

// run prints the current process tree once and exits. It uses loom.WriteRows
// (plain, sequential, cursor-position-agnostic output), not loom.RawScreen:
// this is a one-shot "print and exit" command and must never touch cursor
// position or clear the screen -- that would clobber whatever the shell
// already has on screen above it. See loom.RawScreen's doc comment for why
// that distinction matters (it was a real regression here, caught by PTY
// testing).
func run(width, height, maxNodes, theme int, ansi, showValues bool) error {
	rows, err := renderOnce(context.Background(), width, height, maxNodes, theme, ansi, showValues)
	if err != nil {
		return err
	}
	return loom.WriteRows(os.Stdout, rows)
}

// runWatch redraws in place on interval until ctx is cancelled (SIGINT/
// SIGTERM), like a colored `watch ps`. It draws via loom.RawScreen rather
// than a loom.Pane/View, because loom.View strips inline ANSI from its
// lines in favor of a single uniform Style -- which would silently drop
// --ansi coloring. RawScreen is loom's "final render loop" for exactly this
// case: it clips every row to the terminal's live width and disables
// auto-wrap so an over-wide or stale-sized row can never wrap and cascade
// into a whole-screen scramble, and it positions each row absolutely so a
// bad row can only ever corrupt its own line. See rawscreen.go.
func runWatch(ctx context.Context, width, height, maxNodes, theme int, ansi, showValues bool, interval time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("treemap: --interval must be positive")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	cleanupQuitKey := watchForQuitKey(cancel)
	defer cleanupQuitKey()

	screen := loom.OpenRawScreen(os.Stdout)
	defer screen.Close()

	draw := func() error {
		rows, err := renderOnce(ctx, width, height, maxNodes, theme, ansi, showValues)
		if err != nil {
			return err
		}
		return screen.Draw(rows)
	}
	if err := draw(); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := draw(); err != nil {
				return err
			}
		}
	}
}

func main() {
	var width, height, maxNodes, theme int
	var ansi, showValues, watch bool
	var interval time.Duration
	cmd := &cobra.Command{
		Use:           "treemap",
		Short:         "Render the live process CPU-usage tree as a graph.RenderTreemap box layout",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := parseTheme(theme, ansi); err != nil {
				return err
			}
			if !watch {
				return run(width, height, maxNodes, theme, ansi, showValues)
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runWatch(ctx, width, height, maxNodes, theme, ansi, showValues, interval)
		},
	}
	cmd.Flags().IntVarP(&width, "width", "w", 0, "grid width in columns (default: terminal width)")
	cmd.Flags().IntVarP(&height, "height", "H", 0, "grid height in rows (default: terminal height - 2)")
	cmd.Flags().IntVar(&maxNodes, "max-nodes", graph.MaxTreemapNodes, "node budget passed to AggregateTreemap")
	cmd.Flags().BoolVar(&ansi, "ansi", false, "color each box with a cycling ANSI background")
	cmd.Flags().BoolVar(&showValues, "values", true, "append each segment's %CPU to its label")
	cmd.Flags().BoolVar(&watch, "watch", false, "keep redrawing in place on an interval until interrupted (Ctrl-C)")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "redraw interval in --watch mode")
	cmd.Flags().IntVar(&theme, "theme", 1, "visual style: 1 (bordered boxes), 2 (half-block edges), 3 (thin edges + corner numbers on ambient bg), or 4 (thin edges + corner numbers on node bg) -- 2-4 require --ansi")
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
