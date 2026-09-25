# 122 — Add AST-driven breadcrumbs, symbol outline, and code folding to textedit

**Status**: Closed — implemented in b23e606
**Priority**: P3
**Severity**: Minor
**Category**: Feature / Ergonomics
**Related**: `#117`, `#120`, `#121`

---

## Goal

`/goal`: Leverage the Tree-Sitter concrete syntax tree to provide scope breadcrumbs in the header bar, AST-aware code folding, and a symbol outline panel in `examples/textedit`.

## 1. Context & Motivation

Beyond colorization, Tree-Sitter maintains a full syntactic tree of the active document. Exposing this AST unlocks editor features previously found only in heavy IDEs: scope breadcrumbs showing current enclosing function/class, structural code folding, and a symbol explorer in the sidebar.

## 2. Technical Specification

1. **Scope Breadcrumbs**:
   - Query the AST node at the current caret line/column.
   - Walk up the parent chain to identify enclosing package, type/class, and function/method names.
   - Display breadcrumbs in the top bar (e.g. `📁 demo.go > 🔧 HandleKey > 🔀 switch`).
2. **Symbol Outline View**:
   - Add a toggleable tab in the left sidebar (`[Files | Outline]`).
   - Query top-level declaration nodes (`function_declaration`, `method_declaration`, `type_spec`, `const_spec`).
   - Render a navigable tree of symbols; selecting a symbol moves the editor caret directly to its definition.
3. **AST-Aware Code Folding**:
   - Compute fold ranges from block nodes (`block`, `function_declaration`, `struct_type`).
   - Display `▾` / `▸` in the line number gutter next to foldable blocks.
   - Fold action (e.g. `F2` or gutter click) collapses the block to a single summary line.

## 3. Acceptance Criteria

- Caret movement updates the scope breadcrumb path in the top bar.
- Sidebar outline tab lists all functions/types and jumps to their definition upon selection.
- Code folding collapses and expands blocks cleanly without corrupting line numbers or text buffer.
- Unit and PTY smoke tests verify navigation and fold integrity.
