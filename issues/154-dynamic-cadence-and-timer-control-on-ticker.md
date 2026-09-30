# 154 — Dynamic Cadence and Timer Control on Ticker

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [loom-games](../loom-games)

---

## 1. Problem & Motivation

When developing real-time interactive applications (such as terminal games or dynamic dashboards) with `loom.Ticker`, the current implementation evaluates `shortestTickInterval(root)` at the top of each loop iteration in `Pane.run`. If an application dynamically changes speed (e.g. difficulty changes) or pauses (returning interval `0`), `tickTimer` is adjusted only after waking up from an event.

Furthermore, there is no direct mechanism for a widget to notify the host `Pane` to immediately reset/reschedule its tick countdown (e.g., when a user takes a manual action that should immediately postpone or trigger the next tick).

## 2. Technical Specification / Findings

- `loom.Ticker` defines `TickInterval() time.Duration` and `Tick(now time.Time)`.
- When pausing (interval returning `0`), `tickC` becomes `nil`. When unpausing, the widget relies on an external event or `Invalidate()` to wake the pane.
- Proposed enhancements:
  - Provide an explicit `TickController` interface or enhance `InvalidationAware` to signal ticker interval updates directly.
  - Support instant ticker reset / reschedule methods so games and dynamic widgets can reset step cadence on user interaction.

## 3. Implementation & Verification Plan

### Goal
Provide ergonomic and responsive dynamic cadence control for `loom.Ticker` widgets in `loom.Pane`.

### Acceptance Criteria
- [ ] Widgets can dynamically adjust tick rates or pause/resume with immediate pane responsiveness.
- [ ] Unit tests verifying that interval changes take effect without needing synthetic input events.
- [ ] Clean integration test verifying pause/resume and timer resetting in `pane_test.go`.

## Sprint goal (roadmap 180)
/goal Let a Ticker change or pause its interval and reset its countdown with immediate effect in Pane, with pane_test.go coverage; stop and report if this needs a breaking change to the Ticker interface.
