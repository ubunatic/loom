# 083 — Configure splash and treemap DemoArgs with watch mode in registry to prevent instant exit in loom-demo

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: `internal/examplesreg/registry.go`, `cmd/loom-demo/main.go`, `examples/splash/splash/splash.go`, `examples/treemap/treemap/treemap.go`

---

## 1. Problem & Motivation

When selecting `splash` or `treemap` in the `loom-demo` interactive launcher, the UI renders for a split second and immediately kicks the user back to the `loom-demo` main menu.

Both example applications implement interactive watch modes (`--watch`), but `internal/examplesreg/registry.go` only supplies `DemoArgs: []string{"--watch"}` for the `monitor` example. As a result, `splash` and `treemap` execute their default non-interactive single-shot render paths (`runShowOnce` / `renderOnce`) and terminate immediately upon printing.

## 2. Desired Behavior & Goal

`/goal`: Ensure `splash` and `treemap` run in interactive watch/lifecycle mode when launched from `loom-demo`, remaining on screen until the user explicitly dismisses them.

- Set `DemoArgs: []string{"--watch"}` on `splash` and `treemap` in `internal/examplesreg/registry.go`.
- Ensure `splash`'s `runWatch` displays its full startup progression and waits for user dismissal (`q`/`Esc`) before returning to the demo launcher.
- Ensure `treemap`'s `runWatch` continuously polls CPU process trees and handles `q`/`Esc`/`Ctrl-C` to exit back to `loom-demo`.
- Verify in `loom-bench` that headless smoke execution remains non-blocking and safe.

## 3. Implementation Plan

1. In `internal/examplesreg/registry.go`, add `DemoArgs: []string{"--watch"}` to `splash` and `treemap` registry entries.
2. Verify `splash` and `treemap` interactive execution in `loom-demo` and automated regression tests in `internal/examplesreg/` or PTY smoke suites.
