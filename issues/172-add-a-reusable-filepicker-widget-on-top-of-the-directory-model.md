# 172 — Add a reusable FilePicker widget on top of the Directory model

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: fs.go, docs/Widgets.md §4 (`NavigationPane` in examples/filebrowser, a starting point), examples/filebrowser/filebrowser/browser.go, issues/079-*.md

---

## 1. Problem & Motivation
Consumers (e.g. ~/projects/settings wizard) need to let users pick a file or directory. loom has the `Directory` model (`fs.go`, issue 079) but the only picker UI lives inside `examples/filebrowser` and `examples/ansiviewer`, so it cannot be reused.

## 2. Technical Specification / Findings
A `FilePicker` widget: navigate directories, filter by extension/glob, pick file or directory mode, return the chosen path via a callback; Esc cancels. Reuse `ReadDirectory` and the example's navigation (cursor on previous folder, issue 075). Themeable.

## 3. Implementation & Verification Plan
/goal Ship `loom.FilePicker` with tests and a docs/Widgets.md row, and switch the filebrowser example to it where it fits; stop and report if the example's behavior would change in a way the user must decide.
