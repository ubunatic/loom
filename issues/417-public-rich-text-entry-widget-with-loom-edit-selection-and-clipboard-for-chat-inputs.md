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

## Decision (decide mode, 2026-11-04)

New thin public widget `loom.TextEntry` (option B, argon:med architect). Not a chat mode in `RichTextEdit`, whose
document keymap (^O ^W ^S F1 ^B …) and box modes conflict with a host's keys. Plain text only (taia confirmed).
Reuse RichTextEdit's cluster navigation, selection and undo helpers where practical; do not change its behaviour.

```go
type TextEntryOptions struct {
	Placeholder string
	Prompt      string   // first-row prefix, e.g. "> "
	MaxHeight   int      // visual rows before scrolling; 0 = unbounded
	SubmitKeys  []string // default ["enter"]
	NewlineKeys []string // default ["shift-enter", "ctrl-j", "alt-enter"]
	OnChange    func(value string)
	OnSubmit    func(value string) EventResult
}
func NewTextEntry(opts TextEntryOptions) *TextEntry
func (e *TextEntry) Value() string
func (e *TextEntry) SetValue(s string)               // cursor to end, clears selection, one undo step
func (e *TextEntry) Cursor() (row, col int)          // visual cell of the cursor, relative to the widget, after wrap
func (e *TextEntry) Measure(width int) measure.Size  // soft-wrapped height, 1..MaxHeight
// Widget, Focusable, PasteConsumer (bracketed paste = one undo step), clipboard-aware (OSC 52 on copy/cut)
```

- Host-first keys: the host intercepts before `ConsumeKey`. TextEntry returns `Ignored()` for every key that is
  not an editing key (Tab, Shift-Tab, Ctrl+O/S/G/F, F-keys …), so containers and hosts get them.
- Editing keys: arrows, Home/End, word moves, shift-selection, ctrl+a/c/x/v, ctrl+z/y, backspace/delete, mouse
  click/drag, double/triple click. Soft wrap at cluster boundaries; wide runes correct.
- Tests: selection, cut/copy/paste, undo/redo, wrap and Measure growth, Up/Down across wrapped rows, Cursor()
  after wrap, Submit vs Newline keys, Tab and Ctrl+O return Ignored. Gallery entry, spec/widgets.yaml, docs/Widgets.md.
- Release as v0.7.0 (additive).
