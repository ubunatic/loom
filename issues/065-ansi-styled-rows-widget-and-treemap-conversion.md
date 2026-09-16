# 065 — ANSI-styled rows widget and treemap conversion

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: `canvas.go`, `view.go`, `rawscreen.go`, `widget.go`,
`examples/treemap/treemap/treemap.go`,
`examples/treemap/treemap/input.go`, `examples/treemap/treemap/style.go`,
`graph/treemap.go`,
[055](055-add-a-tab-panel-widget-for-pane-hosting.md),
[056](056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md),
[060](060-periodic-redraw-without-pane-ownership-ticker-interface-and-pane-invalidate.md),
[064](064-convert-splash-and-monitor-examples-to-hostable-widgets.md)

---

## 1. Problem & Motivation

Last and hardest conversion of the "examples as just widgets" initiative.
`treemap` never uses `Widget`/`Pane` at all — and deliberately so: it renders
`[]string` carrying inline ANSI and paints them via
`loom.OpenRawScreen(os.Stdout)` (`treemap.go:98-123`), because `loom.View`
strips inline ANSI in favor of one uniform `Style` (`view.go:71`,
`stripANSI` at `view.go:163-164`), which would silently drop `--ansi` coloring.
The rationale is documented at `treemap.go:80-88`.

It is also the one example that fights the pane for the terminal:

- `watchForQuitKey` opens its **own** `/dev/tty`, calls `term.MakeRaw`, and
  runs its own poll/read loop decoding with `loom.DecodeKey`
  (`examples/treemap/treemap/input.go:27-79`). Two concurrent readers on one
  tty fd steal each other's keystrokes — exactly the failure `pane.go:67-70`
  exists to prevent.
- It installs its own `signal.NotifyContext` (`treemap.go:155`), owns its own
  `time.Ticker` (`treemap.go:112-123`), and `resolveDimensions` writes warnings
  straight to `os.Stderr` (`terminal.go:121-124`) — all of which would scribble
  over a host pane's region.
- Its `Run` threads 12 flags through 12 positional parameters into
  `renderOnce` (`treemap.go:48,129-175`), so an options struct is a
  prerequisite here as much as in [064](064-convert-splash-and-monitor-examples-to-hostable-widgets.md).

## 2. Design — resolved decisions

### 2.1 "Give the app a region to paint freely" — mostly already true

The owner's ask was to hand a hosted application a region it may paint freely.
Checked against the code: **the existing `Draw(c *Canvas, r Rect)` contract
already is that region.** `r` is guaranteed inside `c.Bounds()`, `Draw` must not
write outside `r` (`widget.go:20-22`), `Canvas.Set`/`Fill`/`Write` clip
(`canvas.go:73,122,133`), and every `Cell` carries its own `Style`
(`canvas.go:28`) which `Canvas.Row` emits as ANSI (`canvas.go:148-161`). So a
widget can already paint arbitrary styled content, cell by cell, anywhere in
its rect. No new "free region" API is needed, and none should be added.

The **real remaining gap is narrower**: there is no way to get a *pre-rendered
ANSI string* into that region with its styling intact. `View` is the only rows
widget and it strips ANSI by design. That, not region ownership, is what blocks
treemap.

### 2.2 New widget: ANSI-styled rows

Add a rows widget (working name `StyledRows`) that parses inline SGR sequences
from each input line into per-cell `Style` values and writes them through
`Canvas.Set`, clipping to its `Rect` by display width (reusing `measure/` for
cluster/width handling, cf. 028, and the existing ANSI-aware truncation from
011). It is the Canvas-native counterpart of `RawScreen`'s clip-and-position
guarantees:

- over-wide rows are truncated at the rect edge (never wrap);
- unterminated sequences are reset at the rect edge, so a bad row cannot leak
  style into the rest of the canvas;
- unsupported sequences are dropped rather than passed through as literal text.

`View` keeps its current uniform-style semantics; `StyledRows` is a sibling,
not a replacement.

### 2.3 treemap conversion

- Options struct + `NewWidget(args []string) (loom.Widget, error)` returning a
  `StyledRows`-backed widget fed by the existing `renderOnce` output.
- Delete the private tty reader (`input.go`) and the private
  `signal.NotifyContext` from the hosted path — key handling comes from
  `HandleKey`, signals from the pane (`pane.go:663-671`).
- Periodic refresh via `Ticker` (060) instead of the private `time.Ticker`.
- `resolveDimensions` warnings must not go to `os.Stderr` while hosted: return
  them, or render them into the widget's own rect.
- Standalone `Run` keeps the `RawScreen` path if it still renders better for
  the full-screen `watch`-style use; if `StyledRows` proves equivalent, collapse
  to one path. Decide on evidence (a visual comparison), not up front.

## 3. Verification & Acceptance

- `StyledRows` exists, satisfies `Widget`, and round-trips colored input:
  a line with SGR color renders to a canvas whose `Row(y)` output carries the
  same visible text and equivalent styling; `stripANSI` of both matches.
- Clipping tests: over-wide rows, wide/emoji clusters (cf. 048), and a row with
  an unterminated sequence — style must not leak past the rect.
- `treemap.NewWidget` exists, is registered in `examplesreg`, renders headlessly
  under `loom.Render` in `loom-bench`, and runs as a tab in loom-demo.
- The hosted treemap opens **no** second `/dev/tty`, installs no signal handler
  and writes nothing to `os.Stdout`/`os.Stderr` (asserted by a test that runs
  the widget's `Draw`/`Tick` with stdout/stderr captured and expects them empty).
- Colored treemap output is visually verified by the user before this is
  called done (AgenticLoop Invariant 7); standalone `--ansi` output is
  unchanged.
- `go test ./...`, `go test -race ./...` and `go vet ./...` pass.
