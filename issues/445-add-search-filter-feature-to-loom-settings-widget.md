# 445 — Add search/filter feature to loom.Settings widget

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: issues/407, issues/408, issues/164

---

/goal Add search/filter capability to `loom.Settings`, matching the patterns established by `loom.Choice` and `loom.SearchBar`, verified with unit and PTY tests, or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation
`loom.Settings` currently provides a navigable list of configurable options (KindBool, KindString, KindChoice, KindNumber). However, in applications with many configuration options or complex settings pages, finding a specific option requires scrolling through the entire list.

Other widgets in loom, such as `loom.Choice` and `loom.View`, support search/filter using `loom.SearchBar` (triggered by `/` or `ctrl-f` per key policy in issue 407/408). Adding search/filter to `loom.Settings` improves UX and consistency across compound widgets.

## 2. Technical Specification / Findings
- **Key Binding**: Follow key conventions (e.g. `/` or `ctrl-f` to open search prompt, Esc to close/cancel search). When filtering is active, only matching items are shown and navigable.
- **Components**: Integrate or reuse `loom.SearchBar` (or `loom.TextInput`) similar to how `loom.Choice` embeds search/filtering.
- **Selection & Filtering**:
  - Filter items by `Label` (case-insensitive substring or fuzzy match matching loom standards).
  - Track mapping between filtered indices and the original `Items` slice so activating/editing an option operates on the true underlying `Setting`.
  - Esc or clearing search restores full view.
- **Drawing**: Render search bar / filter prompt either in the header or prompt area of `Settings.Draw`.

## 3. Implementation & Verification Plan
- Add search activation key binding in `spec/defaults.yaml` under settings (if appropriate) or follow `loom.Choice` defaults.
- Implement search filtering and search bar embedding in `settings.go`.
- Add unit tests in `settings_test.go` covering:
  - Activating search, typing filter query, navigating filtered results, editing/toggling an item, and cancelling/clearing search.
- Verification:
  - `make test-q1` passes.
  - `go test -race ./...` passes.
