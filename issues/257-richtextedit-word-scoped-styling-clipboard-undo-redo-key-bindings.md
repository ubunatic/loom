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
- Undo granularity (per keystroke vs. coalesced typing runs) is open; pick a sensible coalescing and document it.
