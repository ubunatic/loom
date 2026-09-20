# 080 — Add declarative Startup transition runner with deterministic completion lifecycle

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `docs/studies/2026-09-20-feature-gap-analysis-splash.md`, `splash.go`, `examples/splash/splash/splash.go`

---

## 1. Problem & Motivation

The splash screen example hard-codes application orchestration: manual `time.Sleep` calls to allow the final frame to render before exiting, manual snapshot copying into views, and custom cancellation coordination. There is no standard framework runner to display a startup/intro layer and deterministically transition to a destination root widget upon task completion or user dismissal.

## 2. Desired Behavior & Goal

`/goal`: Provide a `Pane.RunStartup` or declarative transition layer that orchestrates splash/startup completion, final frame hold, and handover to the main application widget without ad-hoc sleep loops.

- Support `StartupConfig` specifying the splash controller, destination widget, and completion transition.
- Ensure clean deterministic teardown of splash routines and smooth handover to subsequent UI widgets.
- Provide a non-interactive one-shot render helper (`RenderTo` / `TerminalSize`) for CLI splash previews.

## 3. Implementation Plan

1. Define `StartupConfig` and transition runner in `startup.go` (or `pane_startup.go`).
2. Replace ad-hoc `time.Sleep` completion handling with deterministic state transitions.
3. Add automated tests verifying startup lifecycle and handover.
