# 445 — Add search/filter feature to loom.Settings widget

**Status**: Closed — M1 delivered search filtering in loom.Settings with case-insensitive label matching, index mapping, and unit tests (commit 3e800d7)
**Closed-By**: session 25f2a8e0-914e-47e2-9d37-740f360d3ca2
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
- **Key Binding**: Follow key conventions (`/` or `ctrl-f` to open search prompt, Esc to close/cancel search). While searching or when filter query is active, typing updates query; Enter retains filtered list or confirms; Esc clears/closes search and restores full view.
- **Components**: Embed `loom.SearchBar` (or integrate consistent search prompt) in `Settings`.
- **Selection & Filtering**:
  - Filter items by `Label` (case-insensitive substring match).
  - Track mapping between filtered indices and the original `Items` slice so activating/editing an option operates on the true underlying `Setting`.
  - Esc or clearing search restores full view.
- **Drawing**: Render search bar cleanly (e.g. above items or below prompt), accounting for dimensions in `ContentHeight` and `Draw`.

## 3. Implementation & Verification Plan

### M1: Core Search Filtering and Bindings in Settings
- Update `spec/defaults.yaml` under `settings.bindings` with `search: ["/", "ctrl-f"]` (run `make generate` if schema/spec code requires it).
- In `settings.go`:
  - Integrate `SearchBar` into `Settings` with filtered index mapping.
  - Implement search mode activation, typing/backspace handling in `ConsumeKey`.
  - Maintain correct selection mapping between filtered indices and underlying `Items`.
  - Render search bar in `Draw` when active/configured.
- Add unit tests in `settings_test.go` verifying search filtering, navigation across filtered items, inline editing of filtered items, clearing search, and empty match handling.
- Verify with `make test-q1` and `go test -race ./...`.
