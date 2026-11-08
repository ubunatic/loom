# 447 — TextEntry soft-wrap at word boundaries (taia#164)

**Status**: Closed — TextEntry wraps rows at word boundaries; tests and docs added, make test-q1 passed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: taia#164

---

## 1. Problem & Motivation
`TextEntry.rows` broke a row as soon as the next grapheme did not fit, so ordinary words were split across rows in apps such as taia (taia#164). Words that fit on a row should move down instead.

## 2. Technical Specification / Findings
- `rows` in `text_entry.go` is the one layout function; `Cursor`, `locate`, `positionAt`, Up/Down, Home/End, Ctrl+U and `Measure` all use it, so changing it keeps them consistent.
- Rule: when a non-space unit does not fit, the row ends after the last space on it. A space that ends a full row hangs into the reserved caret cell (not at the end of a line, where the caret needs it). A word with no earlier space, or wider than a row, breaks at the edge. Value and newlines are unchanged.

## 3. Implementation & Verification Plan
- Tests: `TestTextEntryWrapsAtWordBoundaries`, `TestTextEntryWrapKeepsLongTokenAndNewlines`, `TestTextEntryWordWrapCaretAndNavigation`.
- Docs: "Soft wrap" in `docs/Widgets.md`.
- Taia picks it up by upgrading loom; no taia workaround.
