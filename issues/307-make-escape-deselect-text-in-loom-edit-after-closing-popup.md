# 307 — Make Escape deselect text in loom edit after closing popup

**Status**: Draft
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:

---

## 1. Problem & Motivation
In `loom edit`, after selecting text and opening a popup, the first Escape closes the popup and returns to the editor. A second Escape currently does nothing, leaving the text selected; it should clear the selection.

## 2. Technical Specification / Findings
Keep this behavior scoped to `loom edit`. Do not change or interfere with general Popup Escape handling. The selection should remain intact while navigating the popup and clear only after control returns to the editor and Escape is pressed again.

## 3. Implementation & Verification Plan
Implement the editor-local Escape behavior and verify the described sequence, along with ordinary popup dismissal and Escape handling elsewhere.

**/goal** Make a second Escape clear the editor selection after a popup closes, verify general Popup handling is unaffected, or stop and report if blocked on a user decision or denied permission.
