# 128 — Add box frame geometry and line-width validator for ANSI assets

**Status**: Open
**Priority**: P1 (High)
**Severity**: Normal
**Category**: Tooling / Quality Gate
**Related**: `ansibuffer.go`, `measure/`, `docs/data/`, `cmd/validate-spec/`

---

## Goal

`/goal`: Add an automated box-frame geometry and visual line-width validator for ANSI assets and mocked TUI designs, ensuring any malformed, misaligned, or ragged boxed ANSI artwork causes immediate test/validation failure during `go test ./...` and `make test`.

## 1. Context & Motivation

- Boxed ANSI art, splash banners, and mockups rely on visual character display width (considering ANSI SGR sequences, multi-byte UTF-8, and double-width runes).
- When an asset's box borders (`┌───┐`, `│...│`, `└───┘`) have uneven visual widths (e.g. 48 cols on top but 53 cols in the body), the right border breaks and shifts.
- To prevent regressions, Loom should provide a box measurement check in `loom.ValidateAnsiBox` (or test validator) that:
  1. Measures visual display column widths of all lines in boxed ANSI strings/files.
  2. Asserts all rows in a framed box have uniform visual width.
  3. Validates all tracked ANSI files in `docs/data/*.ansi`, `*.ansi`, and embedded assets in tests.

## 2. Acceptance Criteria

1. **`loom.ValidateAnsiBox` / `measure` Helper**:
   - Computes display width per line (stripping ANSI escapes, accounting for wide runes).
   - Validates that outer box boundaries and framed lines align to identical visual column widths.

2. **Automated Test Suite Integration**:
   - Add a test in `ansibuffer_test.go` / `measure_test.go` iterating through all `.ansi` files in the repository and asserting uniform line widths and box alignment.
   - Any broken box in an asset immediately fails `go test ./...` and `make test`.

## Delivered

- Added tracked-asset coverage with a source-walk fallback, table-driven geometry cases, path-specific errors, and aligned the two progress mockups flagged by the gate.
