# 211 — Wrappers must not hand-forward optional widget hooks (ticks, invalidation)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Refactor
**Related**: 203, 207, 209

---

## 1. Problem & Motivation
Optional widget interfaces (TickInterval/Tick, invalidation, tick reset, theme) are discovered by type assertion. Any wrapper (e.g. `themedGallery` in `cmd/loom/widgets.go`) silently hides them unless it forwards each one by hand. 207's stopwatch restart bug was exactly this (fixed by forwarding in 6bd3821). That is a caller-side workaround; the rule is to fix the library.

## 2. Technical Specification / Findings
Pick a proven pattern: e.g. an `Unwrap() Widget` interface (like Go's `errors.Unwrap`) that the pane/runtime walks when looking up optional hooks, or container-owned child traversal. Then remove the hand forwarding from `themedGallery` and similar wrappers.

## 3. Implementation & Verification Plan
Test: a minimal wrapper with no forwarding around a ticking widget still ticks and restarts after reset; the existing gallery Stopwatch PTY test stays green.

/goal Optional hooks reach wrapped widgets without per-wrapper forwarding, proven by tests; or stop and report when blocked on a user decision.
