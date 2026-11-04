# 420 — syntax: bash/sh lexer for fenced code in chat replies

**Status**: Closed — shell lexer and CaptureVariable in 1b9a78d; q1 green
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 416; taia#101 (fenced code highlighting by language tag)

---

## 1. Problem & Motivation
taia highlights fenced code blocks by language tag. `syntax.NewLexicalEngine` (`syntax/lexical.go:25`) knows only
go, markdown, json and yaml; shell is the most common missing tag in agent replies. Later: python, diff, toml.

## 2. Specification
- Language names `bash`, `sh`, `shell`, `zsh` map to one shell lexer, in `NewLexicalEngine` and `HighlightString`.
- Captures (existing constants where they exist):
  - `#` comment to end of line, only at line start or after whitespace (not inside `${#x}`, `$#` or words).
  - Strings: `'...'` literal, `"..."` with `\` escapes; unterminated strings run to end of line.
  - Variables: `$name`, `${...}`, `$1`, `$@`, `$?`, `$#` inside and outside double quotes → `CaptureVariable`
    (add the constant and theme entries if missing).
  - Keywords: if then else elif fi for while until do done case esac in function select return local export
    → keyword.
  - Command names: the first word of a command (line start, after `|`, `&&`, `||`, `;`, `(`, `then`, `do`) →
    `CaptureFunction` (or the existing function capture).
  - Numbers as in other lexers; flags like `-x`/`--long` stay plain.
- Line-based, never panics on truncated input, like the JSON lexer (416).

## 3. Verification
Table tests on a sample script (shebang, pipes, `$(...)`, heredoc start line, `${var:-x}`, comments after code,
`#` inside words), truncation at every byte without panic, `HighlightString("sh", s)` equals the engine path.
Update `syntax/README.md`; API snapshot if constants change. Release as v0.8.0 (additive).
