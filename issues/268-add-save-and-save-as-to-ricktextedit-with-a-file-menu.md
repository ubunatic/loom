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
- **M2 delivered — bottom File bar and gallery integration:** Added an opt-in library File bar with Save/Save as, filename/status, keyboard hints, and bottom-anchored menu placement. The gallery now uses the library bar. Commit: `a311b77`. The first `make test-q1` run failed on three M2 assertions; fixes passed focused root/gallery tests, but the full suite has not passed against the final M2 diff.
- **M3 pre-work / required refinements:** Make the F7 hint name the destination mode (`View` while editing, `Edit` while viewing) and assert both states. Run the final `make test-q1` from a clean worktree after the M3 changes; retain its complete output and check for `--- FAIL`. Do not touch or stage the pre-existing `docs/README.md` change.
- **M3 delivered — destination-mode F7 hint:** Updated the bar to show the mode F7 will enter and added library/gallery assertions. Commit: `e9c9b0e`; focused tests passed.
- **M4 pre-work / required refinements:** Update `TestRichTextEditGalleryTypingPTY` and `TestRichTextEditGalleryBoxModeHintAndFallback` to assert the new destination-mode File bar hints and save controls; keep formatting actions out of the bar. Run `make test-q1` from a clean worktree after fixes, record the full log, and resolve all failures before closing. The first final run failed only those two gallery assertions (`/tmp/loom-issue-268-final-test-q1.log`). Host must run the installed `loom widgets --show RichTextEdit` view several times in a PTY and inspect its rows after the tests pass. Preserve the unrelated `docs/README.md` change.
- **M4 delivered — gallery hint assertions:** Updated the PTY and Box Mode tests to check the File bar's destination-mode F7 and save hints, without expecting formatting actions there. Commit: `eaa4bb0`. Focused tests and the final clean-worktree `make test-q1` passed.
- **M5 pre-work / required refinement — compact status visibility:** The installed 80-column PTY showed `Untitled · Unsa…`, truncating the save state, while at 100 columns the full `Untitled · Unsaved` fit. Adjust the concise hint/status allocation so the save state remains fully visible at 80 columns; add a regression test at that width. Keep the useful save and View/Edit shortcuts and do not add text-formatting actions. Re-run `make test-q1` from a clean worktree after the change and check its captured log for `--- FAIL`.
