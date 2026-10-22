# 369 — loom session: run a command in a PTY wrapper with hotkeys, first hotkey saves an ANSI screenshot

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `loomtest/vt.go` (test VT emulator), 365 (`loom crop`/`loom fmt`), editor screenshot hotkey

---

/goal `loom session -- <cmd> [args]` runs the command in a PTY like a one-pane tmux, passes all input
and output through unchanged, and grabs a small set of spec-defined hotkeys; the first hotkey writes the
current screen as an `.ansi` file that `loom view`, `loom measure` and `loom eval` accept; PTY tests
cover pass-through, resize, exit code and the screenshot; or stop and report when blocked on a user
decision.

## 1. Problem & Motivation

TUIs like `harnez usage` (boxes, braille bars, 256 colors) have no way to capture what is on screen as
ANSI. A loom wrapper that keeps a VT model of the child's screen can save it on a hotkey, and later
offer more actions on the same screen model (crop, copy, diff).

## 2. Decisions (host, 2026-10-22)

- Command name `loom session`; the child runs on the primary terminal, loom stays invisible until a
  hotkey is used.
- Screen model: promote/reuse `loomtest.VT` (cells, styles, wide runes, alt screen) into a non-test
  package instead of a second emulator. Missing sequences found with real apps get tests.
- Hotkeys and the prefix live in `spec/defaults.yaml` + schema (no Go key tables, see
  docs/KeyDefaults.md). Proposed default: prefix `ctrl-]` then `s` = screenshot; prefix twice sends the
  prefix to the child.
- Screenshot: the visible screen as SGR + text, trailing blanks trimmed, written to
  `./loom-<cmd>-<timestamp>.ansi`; a one-line status flash names the file without corrupting the child's
  screen (or the file name goes to stderr after exit if a flash is not clean).
- Resize (SIGWINCH) is forwarded to the child and the VT; the child's exit code is loom's exit code.

## 3. Verification

- PTY test with a fixture program: output reaches the outer terminal byte-equal, input reaches the
  child, exit code passes through.
- Screenshot of a known fixture screen equals a golden `.ansi`; `loom measure`/`loom eval` accept it.
- Host live check: `loom session -- harnez usage`, take a screenshot, open it with `loom view`.
