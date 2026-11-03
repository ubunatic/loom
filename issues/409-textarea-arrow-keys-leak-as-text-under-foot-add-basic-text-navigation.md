# 409 — TextArea: arrow keys leak as text under foot; add basic text navigation

**Status**: Closed — Fixed in v0.5.9 (PR #55); owner verified text navigation live in foot via taia (taia#14 closed)
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: 389; codeberg.org/harnez/taia#14

---

/goal In loom's `TextArea`, the arrow keys, Home and End work under foot, and basic text navigation works. Done when the key decoder handles the byte sequences foot actually sends and the keys below act as listed, with decoder tests on those exact bytes and a live check in foot and tmux, and taia picks up the fix by bumping loom (taia#14). Exit: stop and report on an owner decision or a blocker. On a loop or no progress, write a failure report here.

## 1. Problem

The owner saw this in taia (loom v0.5.6, `loom.TextArea` as the chat input) on 2026-11-02. Under foot, these keys insert text instead of acting:

| Key | Inserted |
|---|---|
| Left | `9D` |
| Right | `9C` |
| Up | `9A` |
| Down | `9B` |
| Home | `9H` |
| End | `9F` |

The final letters are the normal CSI/SS3 finals, so the escape prefix is being swallowed and the rest leaks into the text. The `9` may be a modifier parameter (`CSI 1;9 X`) or a split sequence. This is unverified. Record the raw bytes first.

The owner says this belongs in loom, not in each app.

## 2. Scope

| Key | Action |
|---|---|
| Left / Right / Up / Down | Move the cursor. Up and Down keep their current TextArea meaning. |
| Home / End | Start or end of the line |
| Ctrl-Left / Ctrl-Right | Word left or right |
| Ctrl-Home / Ctrl-End | Start or end of the whole text |
| ^A / ^E | Start or end of the current line |
| ^U | Delete from the cursor to the start of the line |

Some of these may already exist in `richtextedit_navigation.go` or `key_navigation_test.go`. Reuse them rather than duplicate. Flag any clash with existing default keys (`key_defaults.go`) before coding.

## 3. Verification

- Decoder tests on the exact foot byte sequences, plus xterm and tmux variants.
- TextArea tests for each action.
- A live check in foot and in tmux.

## Host findings and fix (2026-11-03)

- Cause: `DecodeKey` mapped `CSI 1;<mod><final>` only for single-digit mods 2..8. `CSI 1;9D` (and `1;1D`, two-digit mods) decoded to no key; the pre-869c304 scanner (v0.5.6, v0.5.8) consumed `ESC [ 1 ;` and leaked `9D` as text. On main after 869c304 the key was dropped silently instead.
- Fix: any modifier value decodes through `csiModifierPrefix` (`csiCursorModifier` in `event.go`); `CSI 1;<col>R` keeps the old 2..8 range so row-1 cursor reports never become F3.
- TextArea: added Ctrl-Home/End, Ctrl-Left/Right (word), ^A/^E, ^U. Arrows, Home and End already existed.
- Tests: `key_409_test.go`; decoder and TextArea tests fail without the fix. `make test-q1` green.
- Open: raw foot bytes were never captured. If foot sends `CSI 1;9D` for a plain Left, the key now arrives as `super-left`, which TextArea does not treat as Left. Live check in foot still pending.
