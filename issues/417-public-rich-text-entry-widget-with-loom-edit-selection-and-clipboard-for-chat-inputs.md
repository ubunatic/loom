# 417 — Public rich-text entry widget with loom edit selection and clipboard, for chat inputs

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 254, 257, 194, 265; downstream taia (Codeberg harnez/taia) chat input

---

/goal Ship a public, embeddable rich-text entry widget in loom that gives a host app's input line the same text selection and cut/copy/paste behaviour as `loom edit`, with tests and a gallery or demo entry, or stop and report when blocked on a user decision or denied permission. On a loop, a hard blocker or no progress, clear the goal and write a failure report here.

## 1. Problem & Motivation

The owner wants taia's chat input to support `loom edit`-style text selection and cut/copy/paste. `loom edit` (`cmd/loom/edit.go`) builds on `internal/editor`, which other modules cannot import. taia today renders its own input line with only `loom`, `loom/measure` and `loom/syntax`. loom should capture what `loom edit` does in a reusable rich-text-entry widget, so that apps like taia can embed it instead of copying it.

## 2. Technical Specification / Findings

- Check first what the public `RichTextEdit` (254, 257, 264) already covers versus `internal/editor`. The fix may be to extend or document `RichTextEdit`, not to add a new widget.
- Needed for a chat entry: keyboard and mouse selection (shift+arrows, word and line, multi-click, drag), cut, copy and paste (internal clipboard plus OSC 52 from 194, and bracketed paste), undo and redo, single-line growing to multi-line with soft wrap, and a host hook for Enter (submit) versus newline.
- The host must keep its own keys: the widget must let the host intercept keys before it (e.g. taia's Ctrl+O prefix, Ctrl+S, Ctrl+G, completion popups).
- Plain-text mode is enough for chat. Styling is optional.

## 3. Implementation & Verification Plan

- Expose or extend the widget in a public package, with a short usage example in the docs.
- Tests: selection, cut, copy, paste, undo, wrap and the key-intercept hook.
- `loom edit` may move onto the shared widget where practical, so both stay in step.
- Release in a tagged loom version that taia can adopt (a separate taia ticket).
