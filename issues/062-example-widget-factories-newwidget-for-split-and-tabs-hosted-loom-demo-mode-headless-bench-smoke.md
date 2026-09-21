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

`tabs`-inside-`Tabs` is also the meta-test for 057's key routing: with
`ArrowSwitch = false` on the outer host, `left`/`right` must reach the inner
`Tabs` (via `KeyConsumer`) rather than switching the outer host's tab, and the
inner example's own `SwitchKey` (`ctrl-t`, its own app-level convenience per
057 §2.2 — not a host-reserved binding) must still work.

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
one tab per converted example, `ArrowSwitch = false` (057 §2.3) so `left`/
`right` reach the focused hosted app instead of switching tabs — tab
switching in this mode is by mouse click on the tab bar (`tabs.go:184-201`)
or `Tabs.SetFocusIndex`/`Focus()` wired to whatever selection UI loom-demo
adds (e.g. number keys, or `tab`/`shift-tab` once `Tabs` joins the focus
chain — an open gap noted in 057 §2.3, not solved here) — and
`Tabs.OnChildQuit` containing a hosted app's quit to its own tab. Keep the
existing sequential mode; it remains the only way to launch unconverted
examples.

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
- `loom-demo` has a hosted `Tabs` mode listing the converted examples with
  `ArrowSwitch = false`; the hosted `tabs` example's own `left`/`right`/
  `ctrl-t` still work for switching *its* inner tabs while focused (the
  nested-`Tabs` meta-test), without switching the outer host's tab.
- Closing a hosted app returns to the host instead of terminating loom-demo.
- Standalone `go run ./examples/split` and `./examples/tabs` behave as before.
- `go test ./...` and `go vet ./...` pass; `make install` run afterwards.

---

## Milestones (lean sprint, dev agent: haiku)

Host reviews only diffs and test output; this ticket is the only channel. Root-package tests
needing `/dev/tty` fail before this work; ignore them. Commit each milestone (message ends
'(issue 062 MX)'), staging only your own files, never docs/README.md.

Evidence rules: frames are produced by code, gated on env `LOOM_EVIDENCE=1`, written to
repo-root `docs/progress/062/` (find the root by walking up to `go.mod`; a relative
`../..` path is a known mistake). Run with `LOOM_EVIDENCE=1`, confirm `git status`
shows only intended files.

### M1 - NewWidget factories and registry field
- `NewWidget(args []string) (loom.Widget, error)` for `split` and `tabs`; `Run` becomes a thin
  wrapper; tabs' `EnableMouseClicks` moves into the widget's PaneRequest (see 058 and
  the existing PaneRequest API; do not invent a new one).
- `examplesreg.Example.NewWidget` field, set for split and tabs; registry test covers it.
- Evidence: `M1-standalone-split.ansi` and `M1-standalone-tabs.ansi` via `loom.Render(w, 80, 24)`.

### M2 - loom-bench headless smoke
- For examples with `NewWidget`, `loom-bench` renders at 80x24 and 20x5, asserting non-empty
  output, stable line count, no panic. `SupportsHelp` stays for unconverted examples.
- Test the new bench path with a table test in the bench package.
- Evidence: `M2-bench-tiny-split.ansi` and `M2-bench-tiny-tabs.ansi` (the 20x5 frames).

### M3 - loom-demo hosted Tabs mode
- Hosted mode: one tab per converted example, `ArrowSwitch = false`, `OnChildQuit` contains a
  hosted app's quit; sequential mode stays. Nested-Tabs meta-test: with the hosted `tabs` example
  focused, left/right/ctrl-t switch the inner tabs and never the outer tab; a hosted quit does
  not terminate the host.
- Tests drive it headless with synthetic key events.
- Evidence: `M3-hosted-split.ansi`, `M3-hosted-tabs-inner-switched.ansi` (after a right key).
- Run `make install` at the end of M3 and say so in the report.

### M1-M3 Review (host)
Delivered: 325ca9a, 1eef59f, 4d06668. Tests and vet green. Defects:
- Stray untracked binaries `loom-bench` and `loom-demo` in the repo root (from a build or
  `go install` misuse). Delete them and make sure the build cannot recreate them there.
- `TestNestedTabsMeta` is vacuous: the host has ONE tab, so "focus stays 0" always holds, and
  the inner tabs are never shown to switch. `TestHostedModeSetup` does not cover quit containment.
- The M3 evidence `M3-hosted-tabs-inner-switched.ansi` must actually show the inner tabs switched.

### M4 - Pre-Work / Required Refinements
1. Rewrite the meta-test with a host of at least TWO tabs (split, tabs) focused on the tabs example:
   after a `right` key the inner Tabs' active index changed (assert it) and the outer host's index
   did not; the same for `left`; `ctrl-t` switches the inner tabs and not the outer host.
   The hosted `tabs` widget must be reachable through the same code path `loom-demo --hosted`
   uses (extract a helper `newHostedTabs()` from `runHosted` and test that, not a hand-built host).
2. Test quit containment through that helper: a hosted app's quit key does not return quit from the host.
3. Regenerate `M3-hosted-tabs-inner-switched.ansi` from the helper after a `right` key; it must
   visibly differ from the unswitched frame.
4. In the split example, confirm the PaneRequest mouse mode equals what `Run` enabled before
   this ticket (git show 325ca9a^:examples/split/split/split.go); fix if you changed it.
5. Run gofmt on all files you touched (the loom-demo test has trailing whitespace).
6. Commit '(issue 062 M4)'; go test ./... and go vet ./... green; `make install` again.
