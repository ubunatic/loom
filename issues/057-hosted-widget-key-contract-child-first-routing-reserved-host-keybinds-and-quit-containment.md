# 057 — Hosted widget key contract: child-first routing and quit containment
(filename retains original title incl. "reserved host keybinds" — see §2.2, dropped on revision)

**Status**: Closed — completed in lean sprint 2026-09-20
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Feature
**Related**: `widget.go`, `pane.go`, `tabs.go`, `stack.go`, `grid.go`,
`frame.go`,
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

### 2.2 No reserved switch hotkey — revised 2026-09-16

An earlier draft of this ticket proposed reserving `ctrl-t` as a host-level
"switch tab" key, modeled on browser tab shortcuts. **Rejected on review**:
in every real browser `Ctrl+T` means *open a new tab*, not *next tab* — there
is no universal "next tab" chord (Firefox/Chrome use `ctrl-tab`/`ctrl-pgdn`,
not `ctrl-t`), so the analogy the proposal leaned on doesn't actually hold,
and shipping it would train users on a binding that means something else
everywhere else they use a terminal or browser.

Rather than pick a replacement chord, this ticket now starts **without any
dedicated switch hotkey**. `Tabs.SwitchKey` remains exactly what it already
is post-055: an optional, app-configured convenience (`tabs.go:62-64`,
`examples/tabs` sets it to `ctrl-t` for its own standalone demo) — not a
host-reserved, unrebindable binding. Nothing in this ticket makes any key
reserved. If cross-app switching later needs its own hotkey, that is a
separate, smaller follow-up ticket once real hosted apps expose whether they
have keybind pressure — not a decision to front-load speculatively here.
`event.go`'s `DecodeKey` is unchanged by this ticket (no `ctrl-t` decode gap
to close, no `spec/defaults.yaml` `host_keys:` block).

### 2.3 Tab/shift-tab focus cycling is the primary host-level navigation primitive

Per the owner's explicit instruction, `tab`/`shift-tab` (focus traversal,
already used by `Frame` — `frame.go:429-472` — and swallowed but not yet
cycled by `Stack` — `stack.go:98-101`) and the four arrow keys must stay
available to whichever widget currently owns focus; this is the mechanism
hosted switching should build on instead of a bespoke hotkey. Concretely for
this ticket:

- `Tabs` stops binding bare `left`/`right` unconditionally. Add
  `Tabs.ArrowSwitch bool` (default `true`, preserving 055's standalone
  behavior — `examples/tabs` keeps working unchanged); hosts embedding full
  apps set it `false` so arrows reach the active child via `KeyConsumer`
  routing (§2.1) instead of switching tabs out from under it.
- `Tabs` does **not** yet implement `Focusable` itself, so it is not yet a
  node in `tab`/`shift-tab` traversal the way `Frame`'s boxes are — hosted
  switching purely via `tab`-cycling (tab bar as a focus stop, then into the
  active child) is real follow-on work, not solved by this ticket. Track it
  as an explicit gap for 062 (the first ticket that actually hosts multiple
  full apps under one `Tabs`) rather than inventing it speculatively here.
- Losing an app-level keybind to a future host convenience remains acceptable
  (the owner's position: most goals are reachable more than one way); losing
  `tab`/arrow navigation is not, and this ticket makes no change that would.

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
- `Tabs.ArrowSwitch` defaults to `true` (055/standalone behavior unchanged);
  setting it `false` routes `left`/`right` to the active child via
  `KeyConsumer` instead of switching tabs.
- New tests: nested `Tabs{Frame{Choice}}` delivering `left`/`right` to the
  `Choice` when `ArrowSwitch == false`; `q` consumed by a `KeyConsumer` child
  does **not** trigger the pane quit fallback; `tab`/`shift-tab` still reach
  the focus-owning widget unchanged.
- `Tabs.OnChildQuit` returning false keeps `Pane.Run` alive; returning true or
  being nil quits, covered by a test.
- `Tabs` clears focus on inactive children (test asserts `Focused() == false`).
- `go test ./...` and `go vet ./...` pass.
