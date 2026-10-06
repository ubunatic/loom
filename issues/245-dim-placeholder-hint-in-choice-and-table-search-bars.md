# 245 — Dim placeholder hint in Choice and Table search bars

**Status**: Closed — fixed in 297f222
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 243, 240

---

/goal Choice and Table show a dim placeholder hint in their empty search bar, verified by tests and in the
gallery All tab, or stop and report when blocked on a user decision.

## 1. Problem & Motivation
In the gallery All tab, Choice and Table show an empty search bar as a bare `>` prompt. Nothing tells the
user that typing filters the list (user feedback during the 243 gallery review).

## 2. Technical Specification / Findings
- Show a dim hint (e.g. "type to filter") after the prompt while the query is empty; hide it on the first
  typed character. Reuse TextInput's placeholder style if it has one, otherwise a theme dim style.
- The hint is foreground content: it must not hide the cell background or astra on the rest of the row.
- Make the text configurable on the widget, with a short default.

## 3. Implementation & Verification Plan
- Tests: empty query renders the hint dimmed; one typed rune hides it; the hint is clipped at narrow widths.
- Check both widgets in the gallery All tab, focused and unfocused.
