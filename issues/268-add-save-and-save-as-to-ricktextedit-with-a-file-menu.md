# 268 — Add Save and Save as to RickTextEdit with a File menu

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**:

---

## 1. Problem & Motivation
RichTextEdit currently has no Save or Save as action. Add a File menu to its bottom bar so users can save the current document or choose a destination, and clean up the bar's layout while adding it.

## 2. Technical Specification / Findings
The widget is named `RichTextEdit` in the library. Exact save format and destination-picker behavior should follow the existing document serialization and file-picker APIs; confirm against live code before implementation.

## 3. Implementation & Verification Plan
/goal Add Save and Save as to RichTextEdit through a File menu in its bottom bar, clean up the bar, and verify the behavior with tests; stop and report if blocked on a user decision or denied permission.

Acceptance: Save writes the current document to its associated path; Save as lets the user choose a path and then saves there; the bottom bar includes the File menu and has a clearer layout. Verify with `make test-q1`.
