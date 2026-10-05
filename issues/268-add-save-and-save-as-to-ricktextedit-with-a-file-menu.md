# 268 — Add Save and Save as to RickTextEdit with a File menu

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**:

---

## 1. Problem & Motivation
RichTextEdit currently has no Save or Save as action. Add a File menu to its bottom bar so users can save the current document or choose a destination, and clean up the bar's layout while adding it. Keep the existing text-selection formatting popover (opened with Ctrl+Space) unchanged; it owns text actions, while the bottom bar owns document actions, document status, and keyboard hints.

## 2. Technical Specification / Findings
The widget is named `RichTextEdit` in the library. Exact save format and destination-picker behavior should follow the existing document serialization and file-picker APIs; confirm against live code before implementation.

## 3. Implementation & Verification Plan
/goal Add Save and Save as to RichTextEdit through a File menu in its bottom bar, clean up the bar, and verify the behavior with tests; stop and report if blocked on a user decision or denied permission.

Acceptance: Save writes the current document to its associated path; Save as lets the user choose a path and then saves there; the bottom bar includes the File menu, document status, and keyboard hints without repeating formatting actions from the popover. Verify with `make test-q1`.

### Milestones

- **M1 delivered — document persistence and save picker:** Added `RichTextEdit.Save` / `SaveAs`, an additive save mode in `FilePicker`, serializer customization, and tests for successful saves, association updates, cancellation, and serialization/write failures. Commit: `bc8c2d7`. `make test-q1` passed; no `--- FAIL` lines.
- **M2 pre-work / required refinement — bottom File bar:** Implement the File menu, document status, and keyboard hints in the bottom bar, keeping text formatting in the Ctrl+Space popover. Start the Save as picker with filename input focused even when the document has no associated path, so typing a new name works immediately; add a test for direct typing and keep directory navigation available through an explicit focus switch. Preserve the unrelated pre-existing `docs/README.md` change; do not stage it.
