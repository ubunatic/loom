# 032 — Splash Lifecycle Controller, Async Provider Coordination, and Key Dismissal

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [docs/HarnezSplashTarget.md](../docs/HarnezSplashTarget.md), [cadence.go](../cadence.go), [driver.go](../driver.go)

---

## 1. Problem & Motivation

The splash screen orchestrates asynchronous background tasks (fetching status from multiple providers) while driving animation frames and processing user input (`Esc to skip`). We need a clean lifecycle controller that coordinates these tasks, updates immutable snapshot states, allows user cancellation/skip, and triggers a clean view transition once complete.

## 2. Technical Specification

1. **State Machine**:
   - States: `Initializing`, `RunningTasks`, `Completed`, `Dismissed`.
   - Dispatches step status updates (e.g. `fetching claude...` -> `agy done`).
2. **Asynchronous Task Aggregation**:
   - Run provider probes/initializations concurrently or sequentially.
   - Post state changes into the snapshot channel without blocking redraw cadence.
3. **Input Handling**:
   - Listen for `Esc` (and optionally `Enter`/`q`) to immediately transition or dismiss the splash screen.
4. **Transition Hook**:
   - Signal when the app is ready to hand over the terminal viewport to the primary dashboard view.

## 3. Implementation & Verification Plan

- [x] Implement splash lifecycle controller with channel-based event handling.
- [x] Connect input listener for `Esc` key skip.
- [x] Concurrency tests: verify mock provider completions trigger smooth state progression.
- [x] Ensure redraw loop remains responsive during heavy task simulation.

