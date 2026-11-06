# 441 — TextEntry visual bold toggle and Markdown serialization

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 417, 422, 438, 439, 440

---

## 1. Problem & Motivation

Owner: "Also allow ^B to bold text. and submit it as markdown (only bold for now), other features maybe later"; clarified "^B toggles visual boldness". The editor must show bold glyph styling, not literal Markdown delimiters. Limit this delivery to bold; no generic rich-text editor, other formatting, or external app edits.

## 2. Decision / Technical Specification

Read-only Searcher and terra:med developer plan inspected the plain TextEntry and RichTextEdit models. Adopt a minimal grapheme-aligned bold flag model, not RichTextEdit's rune/span document. Public byte offsets remain offsets in plain Value.

Compatibility decision: additive `EnableBold()` opts into Ctrl+B (otherwise key remains ignored) and `MarkdownValue() string` serializes current text/bold. Existing Value, OnChange, OnSubmit and clipboard remain plain text. The gallery enables bold and calls MarkdownValue from its existing submission callback, demonstrating actual Markdown submission without silently breaking other hosts. Additive methods only; avoid adding fields to TextEntryOptions (unkeyed callers).

- Store private bold flags aligned with units plus typingBold mode; clone/restore both in undo snapshots. Use existing Style.Bold in Draw. Apply highlights then stored bold, preserve selection inversion/colors. No visible delimiters in the editor.
- With selection, Ctrl+B bolds all selected non-newline graphemes unless all are already bold, then unbolds them; preserve selection. With no selection, toggle typingBold. Both actions consume, end typed-run grouping, and are undoable/clear redo. Empty/newline-only selection is harmless and documented. Cursor/selection moves do not implicitly change typing mode.
- All unit mutation paths splice flags: insert/replace, selection delete, Backspace/Delete, Ctrl+U, newline, paste, ReplaceRange and accepted placeholder. Inserted units inherit typingBold. SetValue resets formatting/mode. Copy/cut/OSC52 stay plain; paste uses destination mode. No clipboard changes from format toggles.
- Format-only edits do not fire text-only OnChange; actual text edits retain one notification/undo step. Existing callbacks and methods keep their contracts.
- MarkdownValue emits bold runs as `**text**`, escapes literal ASCII Markdown punctuation/backslashes so stars cannot create accidental formatting, places leading/trailing whitespace outside delimiters, closes runs at each hard newline and preserves original newline characters (do not introduce unrelated hard-break syntax). Whitespace-only runs have no semantic bold in Markdown; document this limitation. Serialize without mutating text, state, history or callbacks. Test exact expected strings for plain text, stars/backslashes, mixed styles, multiline/blank lines, Unicode, whitespace and adjacent spans.

## 3. Milestones and Verification

### M1 — Core state, editing, rendering and serialization

Approved write set: `text_entry.go`, new `text_entry_bold.go`/`text_entry_bold_test.go` if useful, `text_entry_test.go`, `text_entry_paste_test.go`, `testdata/api/loom.txt`, `spec/widgets.yaml`, `docs/Widgets.md`, `docs/Upgrading.md`, this ticket, one bounded sprint study. Implement complete core contract above, regression first, with exact per-cell bold/FG/BG/selection assertions, mixed selection, Unicode units/byte ranges, all edit paths/placeholder, format-mode undo/redo, notifications, plain clipboard and opt-in parent routing. Preserve Ctrl+C/U/D and placeholder key assertions, including deep snapshots in Ctrl+D tests. Run focused checks/vet and one logged make test-q1, install and commit. Do not start M2 before host review.

### M2 — Gallery Markdown submission and terminal proof

After M1 host review, approved candidate files: `gallery/coverage.go`, `gallery/coverage_test.go`, `docs/GalleryCoverage.md`, `cmd/loom/widgets_paste_pty_test.go` (or a focused new bold PTY test there), this ticket/study. Gallery enables bold, advertises Ctrl+B, and demonstrates its submission callback receiving the serialized Markdown. Test selected and typing bold, undo/redo, actual submitted Markdown plus plain Value, and live PTY bold cell styling/caret/retained controls. Test host interception and Ctrl+D propagation, preserve existing Ctrl+C quit and Ctrl+U/placeholder functionality. No ANSI sketch rewrite required; if one is changed validate with loom eval/measure/check-box as applicable.

Host is sole reviewer; developer terra:med. No known baseline failures (last host q1 green). Preserve unrelated changes. No pushes; leaf agents only; do not close until host independently verifies final source and installed binary. Report any unsupported Markdown representation honestly, do not claim a full Markdown round-trip parser.

## Goal

/goal Deliver opt-in visual Ctrl+B bold toggling in TextEntry with reliable editing/history, bold-only Markdown serialization and actual gallery Markdown submission, while preserving existing plain APIs and key contracts; prove it in unit/per-cell/terminal tests and install. Stop and report when blocked on a user decision or denied permission. On a loop, hard architectural or OS blocker, or no progress, clear the goal and record a failure report explaining why.

## M1 implementation evidence (2026-11-07)

- Added opt-in `EnableBold()` and plain-text-preserving `MarkdownValue()` APIs; Ctrl+B remains ignored until opted in.
- Added grapheme-aligned formatting and typing style to edits, replacement, placeholder insertion, undo/redo and drawing. Format-only edits do not notify `OnChange`; clipboard stays plain.
- Added tests for mixed and selected bold, typing style, Unicode ranges, Markdown escaping and whitespace, multiline values, Ctrl+U, placeholder acceptance, callback history and highlighted selection rendering.
- Checks passed: focused bold/Markdown tests, full `TestTextEntry` tests, `go vet ./...`, and refreshed public API snapshot. The locked `make test-q1`, install and commit remain pending.
- M1 study: `docs/studies/2026-11-07-issue-441-textentry-bold-m1.md`. M2 gallery submission and terminal proof have not started; host review is required before M2.
