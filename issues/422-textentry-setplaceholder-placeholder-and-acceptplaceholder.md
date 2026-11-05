# 422 — TextEntry: SetPlaceholder, Placeholder and AcceptPlaceholder

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 417, 418; harnez/taia#110 (model-proposed next prompt as dim hint)

---

## 1. Problem & Motivation
`TextEntryOptions.Placeholder` (text_entry.go:20) is set once at construction and drawn dim when the value is
empty (text_entry.go:278). taia wants to change it at runtime to show a model-proposed next prompt.

## 2. Specification (additive)
```go
func (e *TextEntry) SetPlaceholder(s string) // no OnChange; redraw only
func (e *TextEntry) Placeholder() string
func (e *TextEntry) AcceptPlaceholder() bool // empty value + non-empty placeholder: insert it as text,
                                            // caret at end, one undo step, fires OnChange; else false
```
- Placeholder never appears in `Value()`; typing hides it; caret stays at the start while it shows.
- Multi-line placeholder: draw only the first line, truncated to the row width (check what Draw does today and
  make it safe for `\n` and wide runes).

## 3. Verification
Tests: set/get; drawn dim cells after SetPlaceholder; no OnChange on set; multi-line and wide-rune placeholder
clipped to width, no panic; AcceptPlaceholder inserts, fires OnChange once, undo restores empty; returns false
when value is non-empty. API snapshot, docs/Widgets.md TextEntry section. Release as v0.9.0 (additive minor).
