# 284 — Add Ctrl-P full-editor ANSI screenshots to loom edit

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [283](283-add-f3-and-ctrl-f-search-panel-to-loom-edit.md)

---

## 1. Problem & Motivation
Add ^P (Ctrl-P) in `loom edit` to screenshot the full editor and save an ANSI file in `~/Pictures/Screenshots/` named `<num>-loom-edit-<details>.ansi`.

## 2. Technical Specification / Findings
Capture the full rendered editor viewport, including chrome, hints, visible side/search panels and overlays, preserving colors, styling and terminal-cell geometry. Create the destination directory when missing, allocate a numeric prefix without overwriting existing screenshots, and sanitize filename details. The exact numeric padding and contents of `<details>` remain implementation choices to document.

Store the binding and applicable defaults in the spec. Use supported Loom canvas/ANSI serialization; standard widgets for any feedback, small tweaks only, and file or link issues for larger observed library gaps. Start key-routing work on `codex:sol:med`. Report the saved path or write error without changing document content, dirty state or focus.

Review (2026-10-07, PR #16 head 9f41045, unmerged): ^P now creates a full 100×24 ANSI render; a PTY capture passed loom eval/measure/check-box. Allocation uses len(directory entries)+1 and os.WriteFile: capturing twice, deleting the first file, opening search and capturing again overwrote the second screenshot. Use collision-safe exclusive creation. The binding is hardcoded outside the spec, filename details are unsanitized, and capture errors are discarded without feedback. The unsaved-dialog input branch precedes ^P, so that overlay cannot be captured. The new test checks only that a directory entry exists. Keep open for collision, failure, overlay and editing-state acceptance coverage.

## 3. Implementation & Verification Plan
/goal ^P saves a faithful ANSI screenshot of the full `loom edit` viewport as `~/Pictures/Screenshots/<num>-loom-edit-<details>.ansi`; verify capture, filenames and failure handling, or stop and report when blocked on a user decision or denied permission.

Before implementation, check live code and recent commits. Acceptance: repeated captures get distinct numeric names; a missing directory is created; permission/write errors are visible; Unicode and styles survive; panels and overlays appear when open; capture preserves editing state. Verify captured files with `loom eval`, `loom measure`, and `loom check-box` for box borders, and inspect with `loom view`. Add focused tests using a temporary home and a PTY keybinding check; run `make test-q1` and `make install` before closure.

## Integration verification (2026-10-07)

Host integration fixes numeric allocation with exclusive creation, sanitizes basename details, moves the binding into the spec, reports status/errors and captures the unsaved dialog. Collision and overlay regression tests and a real PTY probe pass; captures pass Loom ANSI validation. Keep open for broader styled/Unicode and editing-state coverage.
