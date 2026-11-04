# 415 — Add HTML screenshot output: offline monospace grid with colours and bold/italic/underline

**Status**: Closed — WriteHTML, RenderHTML, WriteHTMLRows in b38bb5c; q1 green after 416
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: taia#55 (Ctrl+P screenshot of the current view), taia#56 (/export of a whole chat as html); requested by the owner via taia, 2026-11-04

---

## 1. Problem & Motivation
`screenshot.go` exports a frame only as ANSI rows (`Render`, `RenderTo`) or a bash replay script
(`ScreenshotScript`). Hosts want a file a person can open in a browser, offline: taia saves the current view
on Ctrl+P and exports a whole chat as HTML.

## 2. Technical Specification / Findings
Additive API in package `loom` (`screenshot_html.go`):

```go
// HTMLOptions controls HTML screenshot output. The zero value is valid.
type HTMLOptions struct {
	Title string // <title>; empty gives "loom screenshot"
}

// WriteHTML writes canvas c as one self-contained HTML document.
func WriteHTML(w io.Writer, c *Canvas, opts HTMLOptions) error

// RenderHTML draws root off-screen at cols x rows and writes it with WriteHTML.
func RenderHTML(w io.Writer, root Widget, cols, rows int, opts HTMLOptions) error

// WriteHTMLRows writes ANSI rows (as from Render, or any SGR-styled lines of any count) as HTML.
func WriteHTMLRows(w io.Writer, rows []string, opts HTMLOptions) error
```

- Output: one `<pre>` monospace grid, inline `<style>` only, no external assets, no scripts.
- Per cell run: glyph, FG, BG (truecolor, 256 and 16-colour resolved to hex), bold, italic, underline;
  adjacent cells with equal style merge into one `<span>`. Default FG/BG come from the canvas default
  (dark background when unset).
- HTML-escape `<`, `>`, `&`, `"`. Wide runes occupy their cell plus the continuation cell, which emits nothing.
- `WriteHTMLRows` parses SGR via the existing ANSI buffer/parser; non-SGR escapes are dropped.

## 3. Implementation & Verification Plan
- Golden test: a small canvas with a red bold `A`, italic, underline, BG run, wide rune and `<&>`; assert
  exact HTML output.
- `WriteHTMLRows(Render(...))` equals `RenderHTML` for the same widget.
- No `http`, `src=` or `<script` in the output.
- API snapshot refresh; document in `docs/Widgets.md` next to the screenshot helpers.
