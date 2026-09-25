# 119 — Define core syntax package with highlighter engine interface and theme capture mapping

**Status**: Open
**Priority**: P2
**Severity**: Minor
**Category**: Architecture / API Design
**Related**: `#118`, `#120`, `#121`, `codeberg.org/ubunatic/loom/syntax`

---

## Goal

`/goal`: Create a clean, decoupled `codeberg.org/ubunatic/loom/syntax` subpackage providing the syntax engine interface, span representations, standard capture token taxonomies, and Loom theme mappings.

## 1. Context & Motivation

To avoid hard-coupling Loom's UI widgets (`TextArea`, `textedit`) to a specific parser implementation (Tree-Sitter, Chroma, or fallback regex lexers), Loom needs a well-defined syntax highlighting API.

The syntax package should define lightweight data structures for styled spans, incremental edit notifications, and declarative capture-to-theme mappings.

## 2. Technical Specification

1. **Package Location**: `codeberg.org/ubunatic/loom/syntax`.
2. **Core Types**:
   - `Point{Row, Column int}`: Position in a source document.
   - `Edit{StartByte, OldEndByte, NewEndByte int, StartPoint, OldEndPoint, NewEndPoint Point}`: Document edit descriptor for incremental parsers.
   - `Span{StartCol, EndCol int, Style loom.Style, Capture string}`: Line-local styled token span.
   - `Engine interface`:
     - `Language() string`
     - `NotifyEdit(edit Edit)`
     - `Parse(source []byte) error`
     - `HighlightLine(source []byte, line int) []Span`
     - `HighlightViewport(source []byte, startLine, endLine int) map[int][]Span`
3. **Capture Mapping & Theming**:
   - Standard Tree-Sitter capture taxonomy (`@keyword`, `@function`, `@function.call`, `@type`, `@string`, `@number`, `@comment`, `@punctuation.bracket`, etc.).
   - `ThemeMap`: Resolves capture identifiers to `loom.Style` based on the active Loom theme (`spec/themes.yaml`).
   - Fallback `NullEngine` and `LexicalEngine` for environments where heavy parsing is disabled.

## 3. Acceptance Criteria

- `syntax` package is created with 100% unit test coverage for `Span`, `ThemeMap`, and default engine mocks.
- Zero cyclic dependencies between `syntax`, `layout`, and `loom` root package.
- Clear documentation in `syntax/README.md` explaining how syntax providers plug in.
