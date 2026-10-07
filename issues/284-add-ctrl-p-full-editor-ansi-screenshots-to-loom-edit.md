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

## 3. Implementation & Verification Plan
/goal ^P saves a faithful ANSI screenshot of the full `loom edit` viewport as `~/Pictures/Screenshots/<num>-loom-edit-<details>.ansi`; verify capture, filenames and failure handling, or stop and report when blocked on a user decision or denied permission.

Before implementation, check live code and recent commits. Acceptance: repeated captures get distinct numeric names; a missing directory is created; permission/write errors are visible; Unicode and styles survive; panels and overlays appear when open; capture preserves editing state. Verify captured files with `loom eval`, `loom measure`, and `loom check-box` for box borders, and inspect with `loom view`. Add focused tests using a temporary home and a PTY keybinding check; run `make test-q1` and `make install` before closure.
