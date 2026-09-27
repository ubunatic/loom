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
