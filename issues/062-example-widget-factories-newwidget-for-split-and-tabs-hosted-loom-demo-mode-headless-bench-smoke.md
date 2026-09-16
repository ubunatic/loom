# 062 — Example widget factories: NewWidget for split and tabs, hosted loom-demo mode, headless bench smoke

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `internal/examplesreg/registry.go`, `cmd/loom-demo`,
`cmd/loom-bench`, `screenshot.go` (`loom.Render`),
`examples/split/split/split.go`, `examples/tabs/tabs/tabs.go`,
[054](054-add-installable-loom-bench-and-loom-demo-apps-covering-all-examples.md),
[055](055-add-a-tab-panel-widget-for-pane-hosting.md),
[056](056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md),
[057](057-hosted-widget-key-contract-child-first-routing-reserved-host-keybinds-and-quit-containment.md),
[058](058-widget-declared-pane-requirements-panerequest.md)

---

## 1. Problem & Motivation

This is the first end-to-end demonstration of the "examples as just widgets"
pattern (direction 1 of 056): the minimal credible slice that proves the
contract from [057](057-hosted-widget-key-contract-child-first-routing-reserved-host-keybinds-and-quit-containment.md)
and [058](058-widget-declared-pane-requirements-panerequest.md) actually works,
using the two easiest examples.

Today each example's `Run(args)` bundles arg parsing, pane creation and
`pane.Run(widget)` into one function, so there is no bare `loom.Widget` to hand
to a host. `split` and `tabs` are the cheapest conversions: zero flags, zero
theme, no live data, root built inline in ~10 lines
(`examples/split/split/split.go:101-110`, `examples/tabs/tabs/tabs.go:43-60`).

Converting them also fixes a real coverage hole: `examplesreg.Example` has a
`SupportsHelp` flag (`internal/examplesreg/registry.go:33-37`) and `split` and
`tabs` are `false`, so `cmd/loom-bench` skips them entirely — they have **no**
smoke coverage today (cf. 054).

`tabs`-inside-`Tabs` is also the meta-test for 057's key routing: the inner
`Tabs` must still get its switch key while the outer host keeps `ctrl-t`.

## 2. Design — resolved decisions

### 2.1 Factory shape

Each converted example exposes, in its library package:

```go
// NewWidget builds the example's root widget from its command-line args,
// without creating or running a Pane.
func NewWidget(args []string) (loom.Widget, error)
```

`Run(args)` becomes the thin standalone wrapper: *parse args → NewWidget →
open Pane → apply PaneRequest → pane.Run*. `filebrowser` already has this shape
(`newBrowser` at `examples/filebrowser/filebrowser/browser.go:36`, driven by
`filebrowser.go:16-42`) and is the template; it is converted separately in
[063](063-convert-filebrowser-example-to-a-hostable-widget.md) because it is
the forcing function for theming and mouse.

Terminal requirements that the example used to set on its own pane
(`tabs` calls `EnableMouseClicks`, `examples/tabs/tabs/tabs.go:58`) move into
the widget's `PaneRequest` per 058, so a host picks them up automatically.

### 2.2 Registry

Add one field to `examplesreg.Example`:

```go
// NewWidget builds the example's root widget for in-process hosting and
// headless rendering. Nil for examples not yet converted.
NewWidget func(args []string) (loom.Widget, error)
```

Nil means "not converted yet" — the registry stays usable while the remaining
examples are converted in 063/064/065.

### 2.3 loom-demo hosted mode

Add a `Tabs`-based hosted mode alongside today's sequential `Run` launcher:
one tab per converted example, host reserved keys per 057
(`ctrl-t` switch, `ctrl-w` close), and `Tabs.OnChildQuit` containing a hosted
app's quit to its own tab. Keep the existing sequential mode; it remains the
only way to launch unconverted examples.

### 2.4 loom-bench headless smoke

Replace the `Run(["--help"])` probe (`cmd/loom-bench/main.go:77-103`) for
converted examples with a headless `loom.Render(w, 80, 24)` (`screenshot.go:11`)
plus a tiny-terminal pass (e.g. 20x5), asserting non-empty output, stable line
count and no panic. That is a strictly better harness than a help-text exit and
it works for examples with no flag parsing at all. `SupportsHelp` stays for the
unconverted remainder and is removed once every example has `NewWidget`.

## 3. Verification & Acceptance

- `split` and `tabs` expose `NewWidget(args) (loom.Widget, error)`; their `Run`
  is a thin wrapper with no layout or key-binding logic of its own.
- `examplesreg.Example` has the `NewWidget` field, populated for `split` and
  `tabs`.
- `loom-bench` renders both headlessly at 80x24 and at a tiny size without
  panicking, covering the two examples that had no smoke test before.
- `loom-demo` has a hosted `Tabs` mode listing the converted examples;
  `ctrl-t` switches tabs even while a hosted `tabs` example is focused
  (the nested-`Tabs` meta-test), and `tab`/arrows still reach the focused
  child (057 §2.3).
- Closing a hosted app returns to the host instead of terminating loom-demo.
- Standalone `go run ./examples/split` and `./examples/tabs` behave as before.
- `go test ./...` and `go vet ./...` pass; `make install` run afterwards.
