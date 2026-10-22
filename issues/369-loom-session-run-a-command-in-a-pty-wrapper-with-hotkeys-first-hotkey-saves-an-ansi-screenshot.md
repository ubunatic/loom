# 369 — loom session: run a command in a PTY wrapper with hotkeys, first hotkey saves an ANSI screenshot

**Status**: Closed — loom session -- <cmd> wraps a PTY with spec hotkeys: ctrl-p screenshot (shared with loom edit via YAML anchor), ctrl-q quit, prefix ctrl-] s/q/send. Commits 6784784 2067152 cecf3f7 5e2400a 3075d15. Host q1 green twice, validate-spec ok; live: harnez usage -w --minimal, ctrl-p saved a 10-line colored .ansi that loom view renders, ctrl-q exit 0, clean shell line after exit
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
  docs/KeyDefaults.md). User decision: screenshot is `ctrl-p`, the same key as in `loom edit` (reuse
  that binding's spec value or name, do not copy it); `ctrl-q` quits (kills the child, loom exits).
  Direct hotkeys are swallowed and never reach the child. Plain `q` would break typing in the child, so
  it quits only after the prefix `ctrl-]` (prefix then `q` = quit, prefix then `s` = screenshot,
  prefix twice sends `ctrl-]` to the child).
- Screenshot: the visible screen as SGR + text, trailing blanks trimmed, written to
  `./loom-<cmd>-<timestamp>.ansi`; a one-line status flash names the file without corrupting the child's
  screen (or the file name goes to stderr after exit if a flash is not clean).
- Resize (SIGWINCH) is forwarded to the child and the VT; the child's exit code is loom's exit code.

## 3. Verification

- PTY test with a fixture program: output reaches the outer terminal byte-equal, input reaches the
  child, exit code passes through.
- Screenshot of a known fixture screen equals a golden `.ansi`; `loom measure`/`loom eval` accept it.
- Host live check: `loom session -- harnez usage`, take a screenshot, open it with `loom view`.

## Approved plan (host, 2026-10-22)

d369 plan approved: promote `loomtest.VT` to package `vt` (loomtest keeps aliases, no API break), add
`VT.ANSI()`, decode `ctrl-]` (0x1d), `cmd/loom/session.go` with raw mode, PTY, SIGWINCH, exit code,
four milestone commits (vt, spec, event, session). Pre-work:

- Spec keys follow the existing layout: `session.bindings.{screenshot,quit}` and
  `session.prefix_bindings.{screenshot,quit,send_prefix}` as key lists, `session.prefix: ctrl-]`; reuse
  `ctrl-p` via a YAML anchor on `editor.hotkey_screenshot_binding`. Match with `KeyMap`/`KeyEvent.Is`.
- Hotkey matching must not split multi-byte input: an escape sequence or pasted text containing 0x10,
  0x11 or 0x1d inside it is forwarded unchanged; test it.
- The screenshot also works while the child uses the alternate screen (test with `?1049h`).

## M1–M4 delivered (host review, 2026-10-22)

Commits 6784784 (vt package), 2067152 (spec), cecf3f7 (ctrl-] decode), 5e2400a (session). Host q1 green
(0 `--- FAIL`), validate-spec ok. Live: `loom session -- harnez usage -w --minimal`, ctrl-p, ctrl-q:
screenshot saved with colors, exit 0.

### M5 pre-work (required)

- "Saved screenshot: ..." is printed while the outer tty is still raw: the shell output after exit
  staircases (EXIT=0 at column 52). Restore the terminal before printing; test that stderr lines end in
  a clean line start.
- Screenshot keeps the trailing empty rows (30 lines for 9 lines of content): trim trailing blank rows
  as well as trailing blank cells; golden test.
- A lone Esc (0x1b with no following byte in the same read) must reach the child at once, not wait for
  more input; test it.
- A failed screenshot write must not be silent: report it on stderr after exit.
