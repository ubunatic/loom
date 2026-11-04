# 418 — TextEntry: caret offset, ReplaceRange and highlight hook

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 417; taia#84 (completion and /command colouring in the chat input)

---

## 1. Problem & Motivation
taia's completion replaces the word at the caret and colours /commands in the input. `loom.TextEntry` (v0.7.0)
exposes only the visual `Cursor()` and has no styling hook.

## 2. Specification (additive)

```go
// Offsets are byte offsets into Value(). Values inside a grapheme cluster snap to its start;
// out-of-range values clamp to [0, len(Value())].
func (e *TextEntry) CursorOffset() int
func (e *TextEntry) SetCursorOffset(off int)              // clears selection
func (e *TextEntry) ReplaceRange(start, end int, s string) // one undo step; caret to end of s; no OnChange

type TextEntrySpan struct {
	Start, End int // byte offsets into the value, end exclusive
	Style      Style
}
// TextEntryOptions gains:
//	Highlight func(value string) []TextEntrySpan // called on draw; nil = plain
```

- Highlight styles are drawn under the selection: selected cells always use the selection style.
- Overlapping or out-of-range spans: clamp, later span wins; never panic.

## 3. Verification
Tests: offset round trip with wide runes and multi-byte clusters; snapping; ReplaceRange plus one undo restores;
highlight per-cell FG with selection on top; bad spans do not panic. API snapshot, docs/Widgets.md TextEntry section.
Release as v0.7.1.
