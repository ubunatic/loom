# 257 — RichTextEdit word-scoped styling, clipboard, undo/redo key bindings

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: `richtextedit.go`, `richtextedit_test.go`, `spec/defaults.yaml`, issues 254, 255, 256

---

/goal Add the editing behaviors and key bindings below to RichTextEdit with tests, verify via `make test-q1`, or stop and report when blocked on a user decision or denied permission.

Re-check live code and recent commits first; 254–256 shaped the current widget.

## Requested behavior

- **Word-scoped styling**: `C-b`, `C-i`, `C-u` with no selection but the cursor on a word apply/toggle the style on that word.
- **Style menu**: `C-space` selects the current word and opens the format popover (from 255).
- **Copy**: `C-c` copies the selection, or the current word when nothing is selected.
- **Paste**: `C-v` pastes, keeping styled spans.
- **Clipboard is internal** for now (no `wl-copy`/OSC 52/system clipboard).
- **Undo**: `C-z`, `C-y`.
- **Redo**: `C-r`, `C-S-y`, `C-S-z`.
- **Mouse**: double-click selects a word, triple-click selects the line.
- **Common aliases**: `S-Insert` paste, `C-Insert` copy, `S-Delete` cut; add cut (`C-x`) if it fits.

## Notes / uncertainties

- `C-i` equals Tab in legacy terminals; distinguishing it needs kitty/CSI-u keyboard protocol. Verify what the key decoder delivers and record the fallback.
- `C-y` as undo is unusual (often redo); implemented as requested.
- Key bindings belong in the spec if the project specs widget keymaps (see `docs/Spec.md`).
- Undo granularity: word level, simple best-effort for the first version (a typing run up to a word boundary is one undo step; style/paste/cut each one step).

## Implementation notes

Decoder finding (`event.go` `DecodeKey`): byte 0x09 is `tab`, so `ctrl-i` never
arrives on legacy terminals. Added minimal CSI-u decoding (`\x1b[105;5u` ->
`ctrl-i`, `ctrl-shift-y/z`, `ctrl-space`); plain 0x09 stays Tab. Also decoded:
byte 0x00 -> `ctrl-space`; modified tilde keys `ctrl-insert`, `shift-insert`,
`shift-delete` (previously the modifier was dropped). Key bindings are not in the
spec (no widget keymaps are specced), so no spec changes.
