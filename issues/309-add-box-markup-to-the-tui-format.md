# 309 — Add box markup to the .tui format

**Status**: Closed — Delivered [box] block markup in .tui format with rounded/sharp/double/ascii borders, padding, titles, and tests
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**:

---

## 1. Problem & Motivation
Add `.tui` syntax for wrapping text blocks in normal or rounded boxes, so box layout can be authored as readable markup.

## 2. Technical Specification / Findings
This is a future feature for the `.tui` format currently under development. Syntax and rendering details should fit the format design as it evolves.

## 3. Implementation & Verification Plan
Define and implement box block markup for normal and rounded borders. Verify both styles render with correct borders and preserve the enclosed text.

**/goal** Add and verify `.tui` markup for normal and rounded text boxes, or stop and report if blocked on a user decision or denied permission.
