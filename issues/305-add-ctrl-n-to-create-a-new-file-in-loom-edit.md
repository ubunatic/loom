# 305 — Add Ctrl+N to create a new file in loom edit

**Status**: Draft
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:

---

## 1. Problem & Motivation
`loom edit` has no documented `Ctrl+N` action for creating a new file. Add this shortcut so users can start a new file directly from the editor.

## 2. Technical Specification / Findings
The exact creation flow and filename prompt should follow the editor's existing conventions; implementation details remain for the developer to determine.

## 3. Implementation & Verification Plan
Implement `Ctrl+N` in `loom edit`, document the shortcut, and verify the editor can create and edit a new file through it.

**/goal** Implement and verify `Ctrl+N` file creation in `loom edit`, or stop and report if blocked on a user decision or denied permission.
