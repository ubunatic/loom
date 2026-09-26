# 129 — Build loom CLI tool for TUI asset validation, measurement, and interactive viewing

**Status**: Open
**Priority**: P1 (High)
**Severity**: Normal
**Category**: Feature / CLI Tooling
**Related**: `cmd/loom/`, `examples/ansiviewer/`, `ansibuffer.go`, `measure/`, `docs/data/`

---

## Goal

`/goal`: Create a first-class `loom` CLI utility in `cmd/loom/main.go` that serves as the official developer tool for processing, evaluating, measuring, and interacting with TUI and ANSI assets during development. Include `loom view <file>` (with `.ansi` as the first supported file type) and `loom <tui-task>` subcommands (`loom measure`, `loom check-box`, `loom eval`).

## 1. Context & Motivation

- During TUI and ANSI asset development (mockups in `docs/data/*.ansi`, logos, splash banners, and layout files), developers frequently need to:
  1. Inspect and view ANSI art files interactively (`loom view <file.ansi>`).
  2. Measure visual character column widths and line geometries (`loom measure <file.ansi>`).
  3. Validate box frame alignment and assert unbroken rectangular borders without writing ad-hoc python/bash scripts (`loom check-box <file.ansi>`).
- Providing an official `loom` CLI built directly on the Loom SDK unifies these development workflows and gives developers a single binary to inspect and validate TUI assets.

## 2. Command Specifications

1. **`loom view <file>`**:
   - Interactively views `.ansi` files (embedding `ansiviewer` / `loom.AnsiCanvas` viewer).
   - Supports keyboard scrolling, inspection, and `q` / `F10` quit.

2. **`loom measure <file>` / `loom eval <file>`**:
   - Outputs per-line visual display column widths, detecting ragged rows, trailing whitespace, and maximum bounding box dimensions.

3. **`loom check-box <file...>`**:
   - Evaluates box border characters (`┌`, `┐`, `└`, `┘`, `│`, `─`, `╔`, `╗`, `╚`, `╝`, `║`, `═`) and verifies that framed boxes are geometrically aligned and unbroken.
   - Exits with non-zero status if box borders are misaligned, suitable for CI/pre-commit verification.

4. **Installation & Testing**:
   - Add `cmd/loom` to the repository `Makefile` `install` and `build` targets.
   - Comprehensive unit and PTY tests for all subcommands.
