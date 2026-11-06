# 392 — loom session mouse modes: off|full|basic (basic suppresses drag to preserve text selection)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 372 (session mouse restore), 373 (session bar mouse status and toggle)

---

/goal In `loom session`, implement a 3-way mouse mode (`off` | `basic` | `full`) where `basic` mode passes mouse click and scroll events to the child process while suppressing mouse drag events so text selection remains at the session / outer terminal level. The mouse toggle (`prefix m` and clicking the bar mouse icon) cycles through the three modes, the status bar reflects the active mode, and `session.mouse_mode` in the spec configures the default. Stop and report when blocked on a user decision; on a loop or no progress, clear the goal and write a failure report saying why.

## 1. Problem & Motivation

Many full-screen terminal applications (e.g., editors, file managers, interactive TUIs) enable button-event tracking (`?1002h`) or any-event tracking (`?1003h`). When active, dragging with the mouse emits tracking escape sequences instead of allowing the user to select and copy text.

Users often want mouse clicking (focus, buttons, cursor placement) and mouse wheel scrolling to work in child apps, but do not want child apps to steal drag-to-select text selection. Today's session mouse toggle is a binary on/off switch (`mouseOverride int`: -1 disabled, 0 follows child, 1 enabled), offering no middle ground.

## 2. Technical Specification / Findings

- **Mouse Modes**:
  - `off`: Mouse tracking is disabled on the outer terminal (`\x1b[?9l\x1b[?1000l\x1b[?1002l\x1b[?1003l...`). All clicks, scrolls, and drags remain native to the outer terminal.
  - `basic`: Passes mouse scroll (wheel up/down) and discrete clicks (button press and release) through to the child, but does not allow drag events (button-down motion / drag sequences) to reach the child. The outer terminal tracking is clamped to normal tracking (`?1000h` + SGR `?1006h`) rather than motion tracking (`?1002h`/`?1003h`), or drag motion sequences received from the outer terminal are swallowed by `loom session`. Dragging retains text selection at the session / outer terminal level.
  - `full`: Unrestricted pass-through honoring the child's requested mode, including drag tracking (`?1002h`) and any-event motion (`?1003h`).
- **Mode Toggle & Bar Indicator**:
  - `prefix m` and clicking the mouse icon on the session bar cycle through the three modes: `basic` -> `full` -> `off` -> `basic` (or configured cycle order).
  - The session bar mouse icon and style distinguish the three states (e.g., dim/inactive for `off`, standard for `basic`, highlighted/distinct glyph for `full` as specced in `defaults.yaml`).
- **Configuration & Defaults**:
  - Specced in `spec/defaults.yaml` under `session.mouse_mode` (with schema in `defaults.go` / `spec/`), allowing `basic` as the preferred default.

## 3. Implementation & Verification Plan

- Replace binary `mouseOverride` in `sessionTab` with a 3-way enum (`MouseModeOff`, `MouseModeBasic`, `MouseModeFull`, plus follow-child default).
- Update `syncSessionMouse` and `applyTabModes` to clamp outer terminal mouse tracking modes when in `basic` mode (e.g. requesting `?1000h` instead of `?1002h`/`?1003h`).
- Update the input parser in `internal/session/session.go` to filter motion/drag events in `basic` mode.
- Update `sessionBarContent` to render distinct indicator representations for `off`, `basic`, and `full`.
- Add PTY integration tests in `internal/session`:
  - Verify clicks and scroll events reach child in `basic` mode.
  - Verify drag / motion sequences are filtered out in `basic` mode.
  - Verify full mode passes drag events.
  - Verify off mode passes no mouse events.
- Validate with `make test-q1`.

## Jules ideas note (2026-11-03)

The session prefix toggle (`internal/session/session.go:1113-1118`) flips `tab.mouseOverride` between -1 and 1, so once toggled a tab can never return to 0 (follow the child's own mouse mode). Consider this when adding modes here. From Jules ideas run session-vt (GitHub #76).

## Jules runs (wave 2, 429)

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| implement | jules-implement-392 | 2798506688453647188 | GH #116 | rejected: weakened TestSessionTabSwitchPrefixNAndP (WaitForUpdate to WaitFor) to dodge 430; drag test cannot fail (looks for `^[<32;6;6M`, `cat -v` prints `^[[<32;6;6M`) and sleeps 1s; renames `bar_labels.mouse_icon` (breaking); no off-mode test; 10 stray scripts; replies: 2 |

Lead from PR 116: spec `session.mouse_mode: "basic"` plus `mouse_icon_basic|full|off`; drag filter keys on SGR button bit 32 (`\x1b[<32;x;yM`); PTY fixture `printf '\033[?1003h\033[?1006hREADY\n'; cat -v` echoes what reaches the child. Decide how `mouse_mode` relates to the existing `session.mouse: "safe"` first. Event routing: codex:sol:med.
