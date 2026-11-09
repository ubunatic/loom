# 455 — TextEntry SetPrompt

**Status**: Closed — SetPrompt and Prompt added; taia asked for it (taia#104)
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 454

---

## 1. Problem & Motivation
`Prompt` was fixed at construction. taia changes the indicator ("❯", "‼︎", shell-mode prompt) while the entry is in use, so it drew the indicator itself and padded the other rows.

## 2. Technical Specification / Findings
`SetPrompt(string)` replaces the prefix; continuation rows already indent by its display width, and the text re-wraps at the next `Measure` or `Draw`. `Prompt()` reads it. No `OnChange`.

## 3. Implementation & Verification Plan
Test `TestTextEntrySetPromptReindentsAndRewraps`; `docs/Widgets.md` options table; API snapshot.
