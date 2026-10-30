# 391 — agy vim mode UI breaks in loom session (stray letters, repeated lines, wrong popup offset)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: 373 (session bar DECSTBM / child rows), 382 (VT streaming continuity)

---

/goal Investigate and fix terminal emulation and input/output handling in `loom session` so running `agy` in vim mode renders cleanly without stray characters, duplicated lines, or mispositioned inline popups. Verify via PTY integration test or live probe. Stop and report when blocked on a user decision or external tool details; on a loop or no progress, clear the goal and write a failure report saying why.

## 1. Problem & Motivation

When running the `agy` (Antigravity CLI) interactive prompt with vim mode enabled inside a `loom session`, the terminal UI breaks:
- Stray characters / letters appear on the prompt line or input area.
- Screen lines repeat or fail to clear cleanly during redraws and scrolling.
- Inline popups and completions render at incorrect row/column offsets.

In a normal terminal outside `loom session`, `agy` vim mode renders cleanly.

## 2. Technical Specification / Findings

Key areas to investigate:
- **Terminal Queries & CPR (Cursor Position Report)**: Interactive full-screen / prompt applications (like prompt-toolkit or readline in `agy`) frequently query cursor position via `\x1b[6n` or device attributes via `\x1b[c` / `\x1b[>c` to position inline popups. In `internal/session/session.go`, unhandled responses from the outer terminal (e.g. `\x1b[<row>;<col>R`) can be misidentified by `loom.ScanKey` or split, leaking stray letters (`R`, `;`, numbers) into child input. Additionally, the outer terminal coordinates are offset from the child's PTY viewport by the tab/status bar.
- **DECSTBM Margins & Row Offsets**: When the session bar is active, the child PTY runs at `rows - 1` and DECSTBM (`\x1b[1;%dr`) is set on the outer terminal. Any mismatch between terminal window reports, cursor address commands (`\x1b[H`, `\x1b[<row>;<col>H`), and scroll boundaries can produce line repetition and popup misplacement.
- **VT Pass-Through and Mode Sync**: Escape sequences used by vim mode (cursor shape change, bracketed paste, mode changes) may need proper passthrough or state tracking.

## 3. Implementation & Verification Plan

- Reproduce the failure pattern with an integration test in `internal/session` using `ptytest`.
- Ensure terminal queries/responses (CPR, DA) are correctly handled or scoped to child coordinates.
- Verify margin and cursor offset arithmetic when the session bar is displayed.
- Run `make test-q1` to confirm no regressions.
