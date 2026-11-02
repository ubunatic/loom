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

## Jules experiment (2026-11-01): failed run, one lead host-verified

The run ended asking a question (no PR, not answered). Its lead holds at 5c3c603: `scanEscapeKey` (`event.go:612`) does not recognise CPR or DA replies. Host probe with `scanKey`: `ESC[24;80R` consumes 7 bytes and leaves a stray `R`; `ESC[>1;10;0c` and `ESC[?62;22c` consume only 3 bytes and leave the rest as text. That matches the "stray letters" symptom if these replies pass through the key decoder on the session path (not yet checked). Second open question from the run: CPR rows may need rewriting by the status-bar offset before reaching the child.

## Searcher findings (2026-11-02, HEAD 5bf20a3)
- The session server decodes stdin with `loom.ScanKey` (`internal/session/session.go:1061` → `event.go:579-612`) and forwards each unit to the child PTY raw (`session.go:1305`). Plain forwarding is byte-exact; search (`~1140`) and prefix (`~1214`) see the split pieces (stray `R`/`c` as query text or prefix commands).
- `scanEscapeKey` (`event.go:612-660`) treats `R`/`c` as finals only when `b[2]` is a letter; `<`, `=`, `>`, `?` private parameter bytes fall back to 3 bytes.
- The session VT (`vt/vt.go:702-760`) has no `n` (DSR/CPR) or `c` (DA) case, and child output is redrawn from cells, so a child's `ESC[6n`/`ESC[c` may never reach the outer terminal and stay unanswered (hypothesis: matches the misplaced popups; not verified live).
- No session-path tests for CPR/DA.

## Milestones
- **M1 (full CSI scan, library):** `scanEscapeKey` consumes a whole CSI per ECMA-48 (params 0x30-0x3F incl. `<=>?`, intermediates 0x20-0x2F, final 0x40-0x7E); a CPR/DA reply is one unit with no stray bytes. Unit tests for `ESC[24;80R`, `ESC[>1;10;0c`, `ESC[?62;22c`, plus split-across-reads cases; existing key names (arrows, F-keys, SGR mouse, modifyOtherKeys, CSI u) unchanged.
- **M2 (child queries):** a session PTY test where the child sends `ESC[6n` and `ESC[c`; decide from evidence whether the session answers them from the child VT (cursor in child coordinates) or forwards them. Implement the smaller correct option; test that the child gets a well-formed reply and that search/prefix modes never see reply bytes.
- Close only when M1 and M2 tests are green and `make test-q1` passes.

## Plan review (host, 2026-11-02)
Plan from d391 (haiku:med) approved: full-CSI fallback after the existing fast paths, `KeyEvent.Report` for replies, session drops reports before search/prefix, M2 answers DSR 6n / DA from the child VT. Pre-work for M1:
- Before adding `Report`, check every other consumer of CPR in the library (Pane path, `pane_pty_test.go:777`, any code waiting for `ESC[...R` from the outer terminal) so the new flag does not swallow a reply loom itself asked for.
- Golden table of current key names lands in the same commit, before the scanner edit is visible in the diff order.
- M2: confirm whether raw child bytes reach the outer terminal (`session.go:741-775`); if they do, strip the answered queries so the child never gets two replies.
