# 398 — DecodeKey: ESC CR decodes as alt-\r instead of alt-enter; modifyOtherKeys alt-enter decodes empty

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 321 (kitty keyboard: shift-enter, ctrl-enter), harnez/taia#1 and taia#2 (chat input: Enter sends, Alt+Enter newline)

---

/goal Alt+Enter decodes to one stable key name (`alt-enter`) for the legacy `ESC CR` encoding and the
xterm modifyOtherKeys form, proven by `DecodeKey` table tests; or stop and report when blocked on a user decision.

## Findings (taia#1 review, 2026-11-01, probe of `DecodeKey`)

- Legacy terminals send Alt+Enter as `ESC CR`; `DecodeKey` returns `alt-\r`, not `alt-enter`. Bindings
  written as `alt-enter` never match.
- xterm modifyOtherKeys `\x1b[27;3;13~` (Alt+Enter) decodes to `""`.

## Motivation

`taia` (harnez org) uses loom's `TextArea` for chat input with Enter = send, Alt+Enter = newline. taia
works around this in its own key routing for now; the fix belongs in loom.

## Notes

Check live code first; with kitty keyboard (321) on, Alt+Enter may already decode correctly.
