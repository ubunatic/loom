# 250 — Extract reusable SearchBar component and improve search bar styling

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 245

---

/goal Extract a reusable SearchBar component, refactor Choice, Table, and other search-capable widgets to use it, improve search bar styling across themes (especially julia256), and verify with tests; or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation
Widgets such as `Choice`, `Table`, and filesystem pickers currently implement search bar and filter input handling individually. This duplicates query tracking, placeholder rendering, command-mode prompt logic, and styling.
Additionally, current theme styling in themes like `julia256` produces disjointed visuals (e.g. solid cyan block on prompt prefix `>` next to query/placeholder text on normal background).

## 2. Technical Specification / Findings
- **Reusable Component**: Extract a standalone `SearchBar` component responsible for prompt rendering, input query handling, cursor display, placeholder hints, and command-mode integration (`:` / `/`).
- **Widget Integration**: Refactor `Choice`, `Table`, and any related widgets to delegate search bar drawing and input handling to the shared component, enabling future widgets to easily adopt search capabilities.
- **Styling Refinements**: Implement the contained input field design (see `docs/data/searchbar-design-002-field-input.ansi`) with cohesive background container styling across prompt, query, placeholder, and match counter, harmonizing across `julia256`, `mc`, and other themes.

## 3. Implementation & Verification Plan
- Create `SearchBar` component with full unit tests (rendering, placeholder truncation, command mode, input handling).
- Integrate `SearchBar` into `Choice`, `Table`, and verify backwards compatibility.
- Refine theme styles and verify appearance in the gallery All tab and individual widget demos under `julia256` and `plain` themes.
- Run `make test-q1`.
