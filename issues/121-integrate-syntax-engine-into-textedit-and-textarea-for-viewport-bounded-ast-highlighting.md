# 121 — Integrate syntax engine into textedit and TextArea for viewport-bounded AST highlighting

**Status**: Open
**Priority**: P2
**Severity**: Minor
**Category**: Feature / UI Integration
**Related**: `#117`, `#119`, `#120`, `#122`

---

## Goal

`/goal`: Wire the `syntax.Engine` into `loom.TextArea` and `examples/textedit`, replacing ad-hoc string scanning with viewport-bounded AST highlighting that automatically selects grammars by file extension.

## 1. Context & Motivation

With the `syntax.Engine` and Tree-Sitter provider available, `loom.TextArea` and `examples/textedit` should query the engine for syntax spans within the visible screen area (`[scrollY, scrollY + height)`), automatically applying rich token colors based on the active file type.

## 2. Technical Specification

1. **TextArea Integration**:
   - Add `SetHighlighter(h syntax.Engine)` to `loom.TextArea`.
   - During `TextArea.Draw`, query `h.HighlightViewport(...)` for visible rows and overlay styles on rendered text runs.
   - On text edits (`HandleKey`), emit incremental `syntax.Edit` events to the attached engine.
2. **TextEdit App Integration**:
   - Detect file language from path extension (`.go`, `.md`, `.yaml`, `.json`).
   - Instantiate or select the corresponding grammar engine in `OpenFile`.
   - Automatically fall back to the lightweight lexical engine if no grammar matches.
3. **Viewport & Performance Guarantee**:
   - Ensure highlighting only computes spans for visible rows, maintaining smooth 60fps scrolling and instant responsiveness on large files.

## 3. Acceptance Criteria

- Opening `.go`, `.md`, `.json`, and `.yaml` files in `textedit` renders full Tree-Sitter syntax highlighting.
- Editing text updates highlights immediately without lag or visual corruption.
- Unit and PTY smoke tests in `examples/textedit` and `examplesreg` pass cleanly.
- `make test-q1` and `make install` succeed.
