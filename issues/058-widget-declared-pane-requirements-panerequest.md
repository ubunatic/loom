# 058 — Widget-declared pane requirements (PaneRequest)

**Status**: Closed — completed in lean sprint 2026-09-20
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `widget.go`, `pane.go`, `tabs.go`, `stack.go`, `frame.go`,
`grid.go`, [055](055-add-a-tab-panel-widget-for-pane-hosting.md),
[057](057-hosted-widget-key-contract-child-first-routing-reserved-host-keybinds-and-quit-containment.md),
[062](062-example-widget-factories-newwidget-for-split-and-tabs-hosted-loom-demo-mode-headless-bench-smoke.md)

---

## 1. Problem & Motivation

Second library-side prerequisite of the "examples as just widgets" initiative
(see 055/056). Today every terminal-level capability a widget needs is set by
the *example* on the `Pane` it owns, so a widget handed to a foreign host
silently loses it:

| Knob | Set by | Call site |
|---|---|---|
| `EnableMouseClicks` / `EnableMouse` | filebrowser, tabs example | `pane.go:206-225`; `filebrowser.go:40`, `examples/tabs/tabs/tabs.go:58` |
| `Resizeable` | all five pane-using examples | per-example `Run` |
| `MaxCols` | filebrowser (`0`), monitor (`cfg.MaxWidth()`) | `filebrowser.go:64`, `examples/monitor/monitor/watch.go:136`; default from `spec/defaults.yaml` `pane.max_cols: 50` |
| `DisableDefaultQuit` | filebrowser (`q` is its filter key) | `filebrowser.go:59-61,65` |

Preferred size is already expressible widget-side via `ContentHeighter` /
`WidthHeighter` / `ContentWidther` / `Measurer` (`widget.go:46-70`). The
remaining four knobs have no widget-side expression at all, and mouse mode is
the sharp one: a hosted widget cannot request mouse reporting, and degradation
is silent (`HandleMouse` simply never fires, gated on `p.mouse` at
`pane.go:531`).

## 2. Design — resolved decisions

One optional declaration interface rather than four separate ones, read once
by the pane before the run loop starts:

```go
// PaneRequest declares terminal capabilities a widget needs from its host.
// The zero value means "no special requirements".
type PaneRequest struct {
    Mouse      int  // 0 = none, 1000 = clicks, 1003 = full tracking
    Resizeable bool // widget adapts to terminal resize
    MaxCols    int  // 0 = unbounded; otherwise a cap in terminal cells
    OwnsQuit   bool // widget handles its own exit; suppress the pane fallback
}

// PaneRequester is an optional interface for widgets that need terminal
// capabilities beyond the defaults. Pane reads it once, before the run loop.
type PaneRequester interface {
    Widget
    PaneRequest() PaneRequest
}
```

Merge rules for composites (`Tabs`, `Stack`, `Frame`, `Grid` implement
`PaneRequester` by folding their children's requests):

- `Mouse`: max of all children (most capable wins — mouse mode is a terminal
  mode, it cannot be per-subtree).
- `Resizeable`: logical OR.
- `MaxCols`: `0` (unbounded) wins; otherwise the **maximum** of the non-zero
  values, so no child is starved of width it asked for.
- `OwnsQuit`: logical OR — but note this is a coarse, whole-pane signal.
  Prefer `KeyConsumer` from [057](057-hosted-widget-key-contract-child-first-routing-reserved-host-keybinds-and-quit-containment.md)
  for the per-key case (filebrowser's `q`); `OwnsQuit` exists for widgets that
  genuinely manage their own lifecycle. The explicit `Pane` fields stay as an
  imperative override for direct `Pane` users and win over the request.

Rejected: a narrower `MouseWanter interface{ WantsMouse() int }`. It is
smaller, but the other three knobs are needed by the same examples in the same
release, and four one-method interfaces are worse than one struct.

Rejected: mutating pane state mid-run from a widget. The request is read once
before `run`, keeping the "pane owns the terminal" invariant (`pane.go:67-70`)
intact. Widgets that need to change something later do it through the host
that constructed them.

## 3. Verification & Acceptance

- `PaneRequest`/`PaneRequester` exist in `widget.go` and are documented on the
  `Widget` doc block as the declaration channel for terminal capabilities.
- `Pane.Run` reads the root's request once, before the loop, and applies mouse
  mode, `Resizeable`, `MaxCols` and quit policy from it.
- Explicitly set `Pane` fields still win over the request (test).
- `Tabs`/`Stack`/`Frame`/`Grid` merge children's requests per the rules above,
  with a unit test per rule (including nested composites, and `MaxCols: 0`
  beating a non-zero sibling).
- A widget without `PaneRequester` yields the current defaults byte-for-byte
  (existing pane tests unchanged).
- `go test ./...` and `go vet ./...` pass.
