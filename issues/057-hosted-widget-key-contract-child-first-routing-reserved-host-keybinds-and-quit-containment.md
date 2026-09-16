# 057 — Hosted widget key contract: child-first routing, reserved host keybinds, and quit containment

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Feature
**Related**: `widget.go`, `pane.go`, `tabs.go`, `stack.go`, `grid.go`,
`frame.go`, `spec/defaults.yaml`,
[055](055-add-a-tab-panel-widget-for-pane-hosting.md),
[056](056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md),
[058](058-widget-declared-pane-requirements-panerequest.md),
[062](062-example-widget-factories-newwidget-for-split-and-tabs-hosted-loom-demo-mode-headless-bench-smoke.md)

---

## 1. Problem & Motivation

This is the first library-side ticket of the "examples as just widgets"
initiative (follow-up to 055's `Tabs` widget and direction 1 of 056): every
`examples/*` app should expose a bare `loom.Widget` that any host — e.g.
`cmd/loom-demo` driving a `Tabs` — can construct and run, instead of each
example owning and driving its own `Pane`.

Key routing is the first hard blocker. `Widget.HandleKey(e) (quit bool)`
(`widget.go:25`) reports *quit*, not *consumed*, and composites intercept
before delegating:

- `Tabs.HandleKey` swallows `left`/`right` and `SwitchKey` unconditionally
  (`tabs.go:157-178`), so a hosted `filebrowser`/`split`/`Choice` never sees
  horizontal arrows.
- `Stack` swallows `tab` (`stack.go:98-101`), `Frame` swallows its declared
  `Actions` then `tab`/`shift-tab` (`frame.go:429-472`), `Grid` swallows all
  four arrows.
- Quit is global: a hosted app's own quit returns `true` through `Tabs`
  (`tabs.go:177`) up to `pane.run` (`pane.go:576-578`), so `filebrowser`'s
  `ctrl-q` or `split`'s `q` action (`examples/split/split/split.go:109`)
  would kill the whole host, not close a tab.
- Per-app quit policy is a *pane* field: filebrowser needs
  `DisableDefaultQuit = true` because `q` is its filter key
  (`filebrowser.go:65`). Under one shared pane that is all-or-nothing.
  The fallback (`pane.go:598-607`, keys from `spec/defaults.yaml:2-7`) only
  fires when `HandleKey` returns false — but a widget that *consumed* `q` as
  text still returns false, so "consumed" is unobservable today.

## 2. Design — resolved decisions

### 2.1 Consumption signal (additive optional interface)

Add next to `Focusable` in `widget.go`:

```go
// KeyConsumer is an optional interface for widgets that can report whether
// they used a key, so hosting composites and Pane can route child-first.
type KeyConsumer interface {
    Widget
    ConsumeKey(e KeyEvent) (quit, consumed bool)
}
```

Composites (`Tabs`, `Stack`, `Frame`, `Grid`) offer the key to the focused
child **first** via type assertion, and only apply their own binding when the
child did not implement `KeyConsumer` or returned `consumed == false`. Widgets
that do not implement it keep exactly today's behavior — zero migration.
`Pane.handleKeyFallback` consults it too, which retires `DisableDefaultQuit`
as an app-level requirement: filebrowser declares `q` consumed rather than the
pane globally disabling quit.

Rejected: changing `HandleKey`'s signature (breaks every widget and every
downstream consumer for no gain over the optional interface).

### 2.2 Reserved host keybinds (owner's Ctrl+T-inspired proposal)

Child-first routing alone would let a greedy child starve the host of its
switch key. So a small, explicitly reserved set is checked **before** child
delegation and may not be rebound by hosted apps:

| Key | Meaning | Grounding |
|---|---|---|
| `ctrl-t` | next tab / next hosted app | `Tabs.SwitchKey` already exists and documents `ctrl-t` as the intended value (`tabs.go:62-64`); `examples/tabs` sets it today (`examples/tabs/tabs/tabs.go:50`) |
| `ctrl-w` | close / leave the current hosted app | already decoded (`event.go:73`, byte 23) |
| `esc`, `ctrl-c` | host-level quit fallback | already in `fallback_quit_keys` (`spec/defaults.yaml:2-7`) |

**Decode gap to close in this ticket**: `ctrl-t` (byte 20) is *not* decoded by
`DecodeKey` today — the decoded control keys are `ctrl-b/c/d/f/q/u/w`
(`event.go:56-75`). Adding one `case b[0] == 20` is the only new decode logic
required; no new escape-sequence parsing. Do not invent alt-arrow bindings —
the `alt-` prefix currently only reaches `alt-up`/`alt-down` (`event.go:88`,
doc comment `event.go:11`), so alt-left/right is not a free default.

The reserved set lives in `spec/defaults.yaml` (new `host_keys:` block,
alongside `fallback_quit_keys`) so it is spec-sourced, not hardcoded in Go,
per `docs/Spec.md`.

### 2.3 Protected navigation primitives

Per the owner's explicit instruction, the reserved set must **not** capture the
base navigation primitives. `tab`/`shift-tab` (focus traversal, `stack.go:98`,
`frame.go:429-472`) and the four arrow keys stay available to whichever widget
currently owns focus. Concretely:

- `Tabs` stops binding bare `left`/`right`. Add `Tabs.ArrowSwitch bool`
  (default `true` to preserve 055 behavior for standalone use; hosts that
  embed full apps set it `false`). Tab switching under hosting is `ctrl-t`.
- Losing one or two app-level keybinds to the reserved set is acceptable (the
  owner's position: most goals are reachable more than one way). Losing
  `tab`/arrow navigation is not.

### 2.4 Quit containment

Add to `tabs.go`:

```go
// OnChildQuit is called when the active child's HandleKey/HandleMouse
// returns quit=true. Returning true propagates the quit to the Pane;
// returning false keeps the host running (e.g. return focus to the tab bar).
OnChildQuit func(i int) (quitHost bool)
```

Default (nil) keeps today's propagate-everything behavior. `loom-demo` sets it
to contain a hosted app's quit to its own tab.

### 2.5 Drive-by fix

`Tabs.Draw` calls `SetFocus(true)` on the active child (`tabs.go:144-147`) but
never `SetFocus(false)` on the others — unlike `Stack.Draw` (`stack.go:43-45`).
Fix here; it is one loop and it matters as soon as two focusable apps coexist.

## 3. Verification & Acceptance

- `KeyConsumer` exists in `widget.go`; `Tabs`/`Stack`/`Frame`/`Grid` route
  child-first and fall back to their own binding only on `consumed == false`.
- Widgets that do not implement `KeyConsumer` behave exactly as before —
  existing `tabs_test.go`/`stack_test.go`/`frame_test.go` pass unchanged.
- `DecodeKey` returns `ctrl-t` for byte 20, with a unit test in `event_test.go`.
- Reserved host keys are read from `spec/defaults.yaml`, not literals in Go
  (grep: no `"ctrl-t"` string literal outside spec loading and tests).
- New tests: nested `Tabs{Frame{Choice}}` delivering `left`/`right` to the
  `Choice`; `q` consumed by a `KeyConsumer` child does **not** trigger the
  pane quit fallback; `ctrl-t` reaches `Tabs` even when the child consumes
  everything; `tab`/`shift-tab` still reach the focus-owning widget.
- `Tabs.OnChildQuit` returning false keeps `Pane.Run` alive; returning true or
  being nil quits, covered by a test.
- `Tabs` clears focus on inactive children (test asserts `Focused() == false`).
- `go test ./...` and `go vet ./...` pass.
