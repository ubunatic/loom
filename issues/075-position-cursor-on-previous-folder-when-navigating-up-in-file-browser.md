# 075 — Position cursor on previous folder when navigating up in file browser

**Status**: Closed — implemented and verified with full test suite
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `examples/filebrowser/filebrowser/browser.go`, `examples/filebrowser/filebrowser/browser_test.go`

---

## 1. Problem & Motivation

When navigating up to the parent directory (via `..` or directory up) in the file browser widget, the cursor currently resets to the top of the list (index 0, usually `..`). When traversing directory hierarchies, this loses context and requires the user to manually scroll down to locate where they just were.

## 2. Desired Behavior & Goal

`/goal`: When leaving a subdirectory (navigating up to the parent directory), the file browser cursor automatically positions on the directory entry that was just exited rather than resetting to the top of the parent directory list.

- When navigating from `/path/to/child` up to `/path/to`, the selection in the parent directory list should focus on `child/` (or `child`).
- If the exited directory entry cannot be found in the parent list (e.g. renamed or deleted), fallback gracefully to selecting the first item.
- Regular navigation down into a child directory should continue to select the top item (or `..`).

## 3. Implementation & Verification Plan

1. In `examples/filebrowser/filebrowser/browser.go`, update navigation handling when opening parent directory (`..`) to record the previous directory basename.
2. In `loom.Choice` or browser state update, select the item corresponding to the previous directory when populating the listing.
3. Add automated tests in `browser_test.go` verifying cursor positioning when navigating up into parent directories.
