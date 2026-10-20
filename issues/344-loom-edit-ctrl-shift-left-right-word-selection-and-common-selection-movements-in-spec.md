# 344 — loom edit: Ctrl+Shift+Left/Right word selection and common selection movements in spec

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 254 (RichTextEdit selection)

---

/goal Make RichTextEdit support the common text-selection movements, each bound in spec/defaults.yaml, tested and documented; run `make test-q1` and `make install`, or stop and report when blocked on a user decision or denied permission.

## Problem

In `loom edit`, Ctrl+Shift+Left/Right moves nothing into the selection. The RichTextEdit keymap in
spec/defaults.yaml (~l.236–265) has `word_left: ["ctrl-left"]` / `word_right: ["ctrl-right"]` but no
`select_word_left/right`. Unverified: whether the key decoder (event.go, ~l.17 and l.234) emits
`ctrl-shift-left/right` for `ESC[1;6D` / `ESC[1;6C` at all.

## Scope: common selection movements

Each needs a spec binding, decoder support, and a test (select, then extend/shrink across the anchor):

| Movement | Keys |
|---|---|
| char | Shift+Left/Right (present) |
| line up/down | Shift+Up/Down (present) |
| word | Ctrl+Shift+Left/Right (missing) |
| line start/end | Shift+Home/End (present) |
| document start/end | Ctrl+Shift+Home/End (check) |
| page | Shift+PgUp/PgDn (check) |
| select all | `ctrl-alt-a`, `alt-a`, `super-a` (decided 2026-10-20); Ctrl+A stays `line_start` |

Select all: Ctrl+A is kept as line-start navigation and must not be the default. Bind
`select_all: ["ctrl-alt-a", "alt-a", "super-a"]`. Verify the decoder emits each: `ESC a` /
`ESC ^A` (legacy alt), and kitty CSI-u forms incl. super (`super-a` only arrives where the terminal
forwards Cmd/Super, e.g. kitty, WezTerm, Ghostty). macOS: Option must act as Meta for Alt keys; note
this in docs/TuiInput.md.

Also add the new keys to `view_navigation` where read-only selection applies, and to F1 help.
Event routing ticket: start on `codex:sol:med`.

## Verification

Unit tests per movement, plus a decoder test for the modifier sequences (CSI `1;2`, `1;5`, `1;6`
and kitty CSI-u forms). Docs: docs/TuiInput.md, docs/Widgets.md (RichTextEdit keys).
