# 127 — Promote ANSI graphic cell buffer and canvas editor into standard Loom widget

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Architecture / Component
**Related**: `examples/ansiedit/ansiedit/buffer.go`, `examples/ansicanvas_demo/`, `canvas.go`, `parse_ansi.go`, `widget.go`, `docs/Widgets.md`

---

## Goal

`/goal`: Extract the 2D spatial ANSI cell buffer model, SGR/CSI parser, ANSI serializer, and 2D overtype graphic editor from `examples/ansiedit` into a first-class standard Loom SDK component (`loom.AnsiBuffer` and `loom.AnsiEditor` in `ansibuffer.go` / `ansieditor.go`), and create a super simple standalone example app (`examples/ansicanvas_demo/` or `examples/ansicanvas/`) to test and demonstrate its standalone widget embedding.

## 1. Context & Motivation

- Loom provides `loom.TextArea` for 1D stream-based text editing (with line wrapping, tree-sitter syntax highlighting, and gutters) and `loom.Canvas` for immediate-mode drawing.
- Graphic ANSI art editing (`ansiedit`, diagram builders, banner painters) requires a distinct spatial 2D cell grid model:
  - Persistent per-cell styling (FG, BG, bold, underline, invert attributes stored directly on each cell).
  - 2D Overtype & Paint modes with free coordinate cursor navigation without line-reflow.
  - SGR/CSI escape-sequence serialization and parsing roundtrips.
- Promoting it to a standard Loom SDK widget provides a foundational component for ANSI artwork, diagramming tools, and terminal graphics.

## 2. Milestones

- **M1 (Core SDK `AnsiBuffer` & `AnsiEditor`)**:
  - Implement `AnsiBuffer` and `AnsiEditor` in root package (`ansibuffer.go` / `ansieditor.go` or `ansicanvas.go`).
  - Support `LoadAnsiBuffer`, `SaveAnsiBuffer`, `ParseAnsiBuffer`, and full 2D editing operations (put, cut, copy, paste, delete, backspace, overtype/insert).
  - Implement `loom.Widget`, `loom.EventConsumer`, and `loom.MouseConsumer` on `AnsiEditor`.
  - Add core unit tests in `ansibuffer_test.go` and `ansieditor_test.go`.

- **M2 (Super Simple Example App)**:
  - Create a lightweight example in `examples/ansicanvas_demo/main.go` demonstrating embedding `loom.AnsiEditor` in a minimal Loom frame with simple navigation and quit keys.
  - Refactor `examples/ansiedit` to reuse the core `loom.AnsiBuffer` / `loom.AnsiEditor` components.

- **M3 (Verification, PTY Tests & make install)**:
  - Add PTY tests for the new example and verify entire suite with `go test ./...` and `make install`.
