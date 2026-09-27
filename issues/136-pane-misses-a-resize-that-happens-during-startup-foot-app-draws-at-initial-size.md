# 136 — Pane misses a resize that happens during startup (foot -- app draws at initial size)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: pane.go (`New`, `installSignalHandler`, `applyWinch`, `run`), 134 M2d (ZWJ probe at startup)

---

## 1. Problem & Motivation
User report (2026-09-27): `foot -- ansiviewer docs/progress/096` draws the app in a smaller area than
the window (about 80x24) and never grows. Starting `foot` first and then `ansiviewer` inside it uses the
full width. foot opens the pty at its initial size and resizes it right after the child starts, so the
resize lands while the Pane is starting up.

## 2. Technical Specification / Findings
- The Pane reads the size before `signal.Notify(SIGWINCH)` is installed, and `run` only re-queries the
  size in the full-screen, non-alt path. A SIGWINCH delivered before Notify is lost; nothing re-reads
  the size afterwards.
- The 134 M2d ZWJ probe (up to 200 ms at startup) widens this window but is not the root cause.

## 3. Implementation & Verification Plan
- Test first: PTY test that resizes the pty between process start and the first frame (or before
  Notify via a test hook) and asserts the first stable frame uses the new size.
- Fix: after the SIGWINCH handler is installed (and after the ZWJ probe), always re-query the size and
  reflow if it changed, on every screen mode.

## Plan (dev-136, Sonnet; reviewed by host)

Race confirmed: `New()` reads the size (termSize, then a DSR round-trip) before `installSignalHandler`; a WINCH in that window is dropped. `run()` re-queries only when `full && !alt`, so inline and alt panes keep the stale size.
- M1: red test via a same-package hook in `New()` between size read and `installSignalHandler`; the test resizes the pty inside the window.
- M2: after the handler is installed and the ZWJ probe ran, re-query the size and reflow before the first canvas in every screen mode.

Pre-Work (host):
- No timing-dependent tests. Drop the "best-effort" ptytest resize-after-Start test unless it is deterministic (e.g. waits for the first frame, then asserts the frame after a known resize).
- Inline mode: re-querying must not move `inlineStart` or clear scrollback; add a test that an inline pane with unchanged size produces no extra reflow.
- The hook is test-only (unexported var, nil in production).
- Commits: `test: ... (issue 136 M1)`, `fix: ... (issue 136 M2)`; `make test-q1` once after the last edit.

## Handoff (2026-09-27)

No code yet. The Sonnet dev (dev-136b) failed at start (claude exit 1, likely the usage limit).
Resume point: dispatch a fresh developer for M1 (red test) then M2 (fix) from the plan and pre-work above.
