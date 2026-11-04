# 416 — syntax: JSON keys, standalone HighlightString and theme roles for chat output

**Status**: Closed — CaptureProperty, CaptureConstant, HighlightString in e55f66d
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: taia#59 (syntax colours in chat output); requested by the owner via taia, 2026-11-04

---

## 1. Problem & Motivation
taia wants coloured tool lines such as `background_task_result {"id":5,"offset":0,"limit":8000}`. Split agreed
with taia: taia types each chat section (tool name, args JSON, result, diff, error, notice, markdown) and maps
types to theme roles; loom owns lexing in `syntax`.

## 2. Technical Specification / Findings
`NewLexicalEngine("json")` already exists (`syntax/lexical.go`), is line-based and tolerates unterminated
strings, but: object keys get `string`, not a key capture; `-5` and exponents (`1e3`) are not one number;
`true/false/null` get `keyword`; highlighting needs an Engine and `Parse`.

Additive API:

```go
const (
	CaptureProperty Capture = "property"  // JSON object key
	CaptureConstant Capture = "constant"  // true, false, null
)

// HighlightString lexes s as lang without an editor buffer and returns its spans
// (byte and rune offsets into s; multi-line s gives per-line spans). Unknown lang gives nil.
func HighlightString(lang, s string) []Span
```

- JSON: a string followed (after spaces) by `:` is `property`; numbers include sign, fraction, exponent;
  `true/false/null` become `constant`. Partial/truncated input (cut mid-string, mid-number, unbalanced
  brackets) never panics and still colours every complete token.
- `DefaultThemeMap` and the default SGR map gain `property` and `constant` entries.
- Second step (separate ticket, not here): `diff` and fuller `markdown` lexers.

## 3. Implementation & Verification Plan
- Table tests on the taia example, nested objects, arrays, negative/exponent numbers, escaped quotes, and
  truncations at every byte of a sample (no panic; spans within bounds; no span over a cut token boundary).
- `HighlightString` equals `NewLexicalEngine(lang)` + `Parse` + `HighlightLine` per line.
- API snapshot refresh; update `syntax/README.md`.
