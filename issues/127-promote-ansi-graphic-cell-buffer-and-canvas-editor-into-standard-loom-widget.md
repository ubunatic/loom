# 127 — Promote ANSI graphic cell buffer and canvas editor into standard Loom widget

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Architecture / Component
**Related**: `examples/ansiedit/ansiedit/buffer.go`, `canvas.go`, `parse_ansi.go`, `widget.go`, `docs/Widgets.md`

---

## Goal

`/goal`: Extract the 2D spatial ANSI cell buffer model, SGR/CSI parser, ANSI serializer, and 2D overtype graphic editor from `examples/ansiedit` into a first-class standard Loom widget (`loom.AnsiCanvas` / `loom.AnsiEditor`), enabling reusable 2D diagramming, terminal art creation, and interactive ANSI canvas editing across Loom applications.

## 1. Context & Motivation

- Loom provides `loom.TextArea` for 1D stream-based text editing (with line wrapping, tree-sitter syntax highlighting, and gutters) and `loom.Canvas` for immediate-mode drawing.
- Graphic ANSI art editing (`ansiedit`, diagram builders, banner painters) requires a distinct spatial 2D cell grid model:
  - Persistent per-cell styling (FG, BG, bold, underline, invert attributes stored directly on each cell).
  - 2D Overtype & Paint modes with free coordinate cursor navigation without line-reflow.
  - SGR/CSI escape-sequence serialization and parsing roundtrips.
- Rather than leaving this logic isolated inside `examples/ansiedit/ansiedit/buffer.go`, promoting it to a standard Loom SDK widget provides a foundational component for ANSI artwork, diagramming tools, and terminal graphics.

## 2. Scope & Acceptance Criteria

1. **SDK Component (`loom.AnsiBuffer` & `loom.AnsiEditor`)**:
   - Promote the 2D cell grid buffer (`AnsiBuffer`, `BufferCell`) into the Loom package.
   - Standardize `LoadBuffer(path/reader)`, `SaveBuffer(path/writer)`, and `ParseBuffer(data)`.
   - Provide an `AnsiEditor` widget implementing `loom.Widget`, `loom.EventConsumer`, and `loom.MouseConsumer`.

2. **Refactor Examples**:
   - Refactor `examples/ansiedit` to embed and compose the standard `loom.AnsiEditor` component.
   - Ensure compatibility with `examples/ansiviewer` for static or animated ANSI inspection.

3. **Automated Testing & Documentation**:
   - Move unit tests and roundtrip ANSI encoding tests into the core test suite.
   - Document `loom.AnsiEditor` in `docs/Widgets.md`.
