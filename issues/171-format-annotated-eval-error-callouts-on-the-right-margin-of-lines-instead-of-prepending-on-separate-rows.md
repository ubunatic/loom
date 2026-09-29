# 171 — Format annotated eval error callouts on the right margin of lines instead of prepending on separate rows

**Status**: Closed — Aligned annotated evaluation margin callout markers with padded right margin
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Bug / UX
**Related**: #166

---

## 1. Problem & Motivation

In `writeAnnotated` (`cmd/loom/main.go`), error callout indicators (`<-- 1`) are currently printed on their own preceding lines via `fmt.Fprintf(out, "<-- %d\n", index)` rather than on the right margin of the offending row.

This breaks the rendered visual layout of the ANSI art / box grid by inserting blank-indented extra rows into the output text.

## 2. Technical Specification

- For each row in `writeAnnotated`:
  - Print the line content.
  - If the row has one or more diagnostic annotations, append the callout indicators (`<-- 1`, `<-- 1, 2`) on the right margin of that row.
  - If aligning right margins across the grid, pad rows to the maximum grid line width so the arrows line up cleanly (e.g. `│ line content │  <-- 1`).
  - Do not emit standalone `<-- N` rows before the grid lines.
- Update tests in `cmd/loom/main_test.go` (`TestEvalAnnotatedOutput` and `TestCheckBoxAnnotatedOutput`) to assert that markers appear inline / on the row's right margin rather than on separate lines.

## 3. Implementation & Verification Plan

- **Milestone 1**: Update `writeAnnotated` in `cmd/loom/main.go` to append callout markers on the right margin of the row, padded neatly.
- Update unit tests in `cmd/loom/main_test.go`.
- Verify with `make test-q1` and `make install`.
