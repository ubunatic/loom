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

## Sprint Log

**M1 delivered (44c11c0): StyledRows widget** — `styledrows.go` parses SGR via `ParseANSI` and clips whole clusters at the rect edge; tests cover a per-row style reset, width clipping and height clipping.

### M2 — hosted treemap

**Pre-Work / Required Refinements (from M1 review):**
1. Add the acceptance round-trip test: colored input drawn to a canvas, `Row(y)` carries the same visible text (stripANSI equal) and the equivalent SGR color.
2. Unterminated sequence: assert that cells in the rect *after* the text (e.g. `(3,0)` for `"\x1b[31mred"`) are not red, and that a truncated escape at the end of a line (`"ab\x1b[3"`) draws no literal escape bytes.
3. Unsupported sequences (`"\x1b[2Jab"`, OSC `"\x1b]0;t\x07ab"`) are dropped, not drawn as text.
4. Rows shorter than the rect and rect rows past `len(Lines)` must be cleared (blank, Reset style), so a redraw with fewer or shorter lines leaves no stale cells. Add a test that draws twice on the same canvas.

**M2 scope:** follows the plan — `Options` struct, `NewWidget(args)`, key handling via `HandleKey`, refresh via `Ticker`, no tty, signal or stdout/stderr access while hosted (capture test), `examplesreg` registration.

**M2 pre-work delivered (8f883aa):** StyledRows round-trip test, escape filtering, and clearing of stale cells.
**M2 delivered (f060955): hosted treemap** — `Options`, `NewWidget`, `Ticker` refresh, capture test showing Draw and Tick stay silent, registration in `examplesreg`, and loom-demo passing `DemoArgs`.

### M3 — standalone decision, demo tab, visual check

**Pre-Work / Required Refinements (from M2 review):**
1. **Size from the rect, not the terminal.** The hosted widget renders at `opts.Width/Height` or, when those are 0, at the *terminal* size (`resolveDimensionsWithOutput` probes the tty). While hosted, render at `r.W x r.H` from `Draw`. Re-render whenever the rect size changes, with no tty probe on the hosted path. Test: `NewWidget(nil)` drawn into a 40x10 rect fills exactly 40x10, and a second Draw at 60x15 re-renders to the new size.
2. **One flag definition.** `parseWidgetOptions` duplicates every cobra flag (names, defaults and help text). Define the flags once, for example a `bindFlags(fs *pflag.FlagSet, *Options)` used by both the cobra command and `NewWidget`. Also have the cobra `RunE` call `validateOptions` instead of repeating its inline checks.
3. **Tick must not block the UI.** Tick scans `/proc` synchronously on the pane loop, and errors are silently dropped. Collect in the background like the monitor collectors from 064: Tick triggers or reads the latest result, and the widget provides `Close()`. Draw a collection error inside the rect instead of swallowing it.
4. **Quit-key semantics.** `HandleKey` returns true for `q`, `esc`, `ctrl-c` and `ctrl-q`. Confirm this matches the monitor widget's hosted convention from 064. If "true" only means the key was consumed, a hosted tab would silently swallow `q`/`esc`. Align with monitor and state the convention in a comment.

**M3 scope:** decide, with evidence, whether standalone `--watch` keeps `RawScreen` (compare output for the same data). Make sure treemap shows up as a loom-demo tab, and have `loom-bench` / `loom.Render` produce colored output for the user's visual check (write an `.ansi` snapshot to `/tmp` and report its path).

**Pre-work delivered (0a56571):** Hosted renders size from the Draw rect and refresh on resize without probing terminal dimensions; Cobra and `NewWidget` share `bindFlags`; collection runs asynchronously, exposes `Close`, and displays collection errors; `HandleKey` follows Widget quit semantics.

**M3 delivered:** Kept `RawScreen` for standalone `--watch`. A deterministic same-row comparison verifies that redirected `RawScreen` output and `StyledRows` render equivalent text and per-cell SGR styles. `RawScreen` remains useful for standalone watch because it owns the physical screen, clips to its live width, and suppresses auto-wrap. loom-demo includes the live treemap tab and closes hosted widgets; loom-bench waits for its colored frame. Headless ANSI snapshot: `/tmp/loom-065-m3-treemap.ansi` (80x24). User visual verification of that snapshot remains outstanding.

**M3 pre-work delivered (0a56571):** the widget renders at the rect size and re-collects when the size changes; one shared `bindFlags` for cobra and the widget; collection runs in the background with `Close()`; errors are drawn inside the rect; quit keys follow the Widget child-quit semantics.
**M3 delivered (30a17e4):** standalone `--watch` keeps `RawScreen` (a parity test shows equal text and SGR, and RawScreen still owns full-screen positioning and auto-wrap suppression); `WaitReady` for headless renderers; loom-bench renders treemap; loom-demo closes hosted tabs on exit. Snapshot: `/tmp/loom-065-m3-treemap.ansi`.

### M4 — repaint after background collection

**Pre-Work / Required Refinements (from M3 review):**
1. **No repaint when a collection finishes.** The goroutine updates rows, but nothing asks the pane to redraw. With `--watch`, the first real frame waits for the next tick (up to 2s after "Collecting…"). Without `--watch`, `TickInterval()==0`, so a hosted treemap shows "Collecting process tree…" until some unrelated event happens. Fix this with the existing mechanism (`Pane.Invalidate`, pane.go:625, or whatever widget-to-host redraw path 060 established). If widgets cannot reach it, use the smallest generic hook: for example, a short tick interval while a collection is pending. Test: a non-watch hosted widget under a pane or test host shows real rows without any key event or ticker.
2. Check that the `closeHostedTabs` loop in loom-demo also covers the monitor from 064 (so no double Close with its own `defer w.Close()`), and that `Close` is idempotent. Add a test that calls it twice.
