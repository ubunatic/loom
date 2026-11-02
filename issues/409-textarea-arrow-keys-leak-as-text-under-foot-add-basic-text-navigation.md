# 409 — TextArea: arrow keys leak as text under foot; add basic text navigation

**Status**: Open
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
