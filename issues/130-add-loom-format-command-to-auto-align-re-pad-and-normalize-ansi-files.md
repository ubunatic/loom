# 130 — Add loom format command to auto-align, re-pad, and normalize ANSI files

**Status**: Open
**Priority**: P1 (High)
**Severity**: Normal
**Category**: Feature / CLI Tooling
**Related**: `cmd/loom/`, `ansibuffer.go`, `measure/`

---

## Goal

`/goal`: Implement `loom format [file.ansi]` (with `--width N`, `--write`, `--trim-trailing`, and `--align-box` flags) in the `loom` CLI, enabling automated re-padding, visual line width equalization, trailing escape normalization, and in-place fixing of ragged ANSI assets without manual scripting.

## 1. Context & Motivation

- While `loom measure` and `loom check-box` detect ragged rows and broken boxes, developers currently have to write ad-hoc Python/shell scripts to compute per-line visual shortfall and inject spaces before closing borders.
- Providing `loom format` closes the gap between detection and automated repair.

## 2. Command Specification

- **`loom format <file.ansi>`**:
  - `--width <N>`: Sets target visual column width for all rows (pads shorter rows with whitespace, optionally truncates longer rows).
  - `--align-box`: Detects boxed lines (`│...│`) and pads inner content so closing right borders align at the target column.
  - `--write` / `-w`: Modifies file in-place (defaults to writing formatted ANSI to stdout).
  - `--trim-trailing`: Strips redundant trailing whitespace before newline or EOF.
