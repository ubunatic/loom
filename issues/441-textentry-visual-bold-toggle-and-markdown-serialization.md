# 441 — TextEntry visual bold toggle and Markdown serialization

**Status**: Closed — done: bold toggle + MarkdownValue delivered, gallery verified by owner, q1 green
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

#### Pre-Work / Required Refinements from M1 host review

- Preserve non-opted-in rendering: M1 changed selection color behavior unconditionally and updated two older assertions. Restore those original non-enabled selection assertions and behavior; only bold-enabled entries may compose retained highlight colors/bold/inversion. Keep API comments explicit about this opt-in difference and document any opted-in semantics.
- Repair Ctrl+D state matrix: adding Ctrl+B then Ctrl+Z now undoes formatting rather than typing, so cases named empty/placeholder still contain text. Rebuild actual empty/text/selected/placeholder states with explicit value/selection assertions before probing the key; preserve the original non-enabled matrix and add enabled formatting-state cases. Compare current bold flags, typing mode and enable state as well as history snapshots after direct and Frame routing.
- Complete core edit-path coverage with exact Markdown plus plain text, caret/selection and undo expectations for Backspace/Delete, cut, internal paste, bracketed paste in bold mode, selection replacement, SetValue reset and format redo clearing. Current tests often check only alignment lengths; that does not prove correct styles survived. Add Markdown cases for all-bold whitespace padding, whitespace-only/newline-only selection, neighboring runs, literal punctuation/backslashes, multiline Unicode and serializer purity (no state/history/callback mutation).
- No generic Markdown parser or unrelated feature. Keep plain callback/clipboard contracts exact; gallery submission must explicitly call MarkdownValue. Host has reviewed M1 assertions and requires these compatibility refinements before final acceptance. M1 developer q1 passed; no host full run yet (host reruns on final M2 state).

After M1 host review, approved candidate files: `gallery/coverage.go`, `gallery/coverage_test.go`, `docs/GalleryCoverage.md`, `cmd/loom/widgets_paste_pty_test.go` (or a focused new bold PTY test there), this ticket/study. Gallery enables bold, advertises Ctrl+B, and demonstrates its submission callback receiving the serialized Markdown. Test selected and typing bold, undo/redo, actual submitted Markdown plus plain Value, and live PTY bold cell styling/caret/retained controls. Test host interception and Ctrl+D propagation, preserve existing Ctrl+C quit and Ctrl+U/placeholder functionality. No ANSI sketch rewrite required; if one is changed validate with loom eval/measure/check-box as applicable.

Host is sole reviewer; developer terra:med. No known baseline failures (last host q1 green). Preserve unrelated changes. No pushes; leaf agents only; do not close until host independently verifies final source and installed binary. Report any unsupported Markdown representation honestly, do not claim a full Markdown round-trip parser.

## Host review and verification — delivered, full-suite blocker

Host reviewed M1/M2 source and assertion diffs and required preserving non-opted-in selection styling, repairing true Ctrl+D test states and strengthening mutation/serialization coverage. Those refinements and gallery Markdown submission are committed. Independent host `make test-q1` failed only at untouched `internal/session/TestSessionDetachPrefixDAndReattach`: retained output appeared but tabs did not repaint (log `/tmp/loom-issue441-host-test-q1.log`). This also failed in the developer M2 run and intermittently before this feature; no rerun or unrelated test weakening performed. Other logged packages passed. Full green verification remains owed, so leave this issue open.

Installed-binary PTY smoke passed: F2 cleared entry, Ctrl+B then `sent` displayed plain text without literal Markdown markers at offset 4; Enter displayed `Submitted Markdown: **sent**`; F10 exited successfully. Developer PTY tests assert selected/typing bold cell attributes, inversion, caret and undo/redo. Final developer install passed. Changes remain local, not pushed. Resume point: investigate the session reattach failure separately with owner authorization, then perform permitted full verification; no further bold implementation is currently required by host review.

## M3 — Faithful escaped gallery submission preview

Owner's screenshot showed three input lines but gallery `Submitted Markdown: 123`; owner asks to double-check code and gallery and show full characters including newlines and others. Read-only developer verified MarkdownValue preserves raw newlines, but Canvas.Write renders the unquoted submission on one row and flattens controls. This is a real gallery display bug, not evidence that serialization lost line breaks.

### Approved pre-work and write set

- `gallery/coverage.go`, `gallery/coverage_test.go`, `cmd/loom/widgets_paste_pty_test.go`, `docs/GalleryCoverage.md` if needed, this ticket and the existing study. No core serializer/API change needed; stop for approval if a true serializer defect is found.
- Quote/escape display only (prefer standard strconv.Quote semantics): visible `"1\\n2\\n3"`, tabs, carriage returns, controls, quotes/backslashes, readable Unicode and literal Markdown delimiters. Actual submitted/plain values must not change. Distinguish input normalization from serialized value; do not falsely assert original clipboard control bytes survive if TextEntry intentionally normalizes them.
- Wrap escaped output by display width within the supplied rectangle, reserving editable content and controls. Show full escaped value when space permits, and explicit overflow when it does not; no silent truncation presented as full. Preserve tiny-height safety, selection/bold/caret and controls, no painting outside child bounds or stale preview after redraw.
- Exact regression tests for raw `1\n2\n3` plus quoted visible backslash-n output, control/quote/backslash/Unicode and bold Markdown cases, wrapping/overflow/tiny sizes. Actual CLI PTY test must submit multiline/control content and assert escaped preview while retaining existing bold attributes/caret and other key tests. Update existing single-line expectations to quoted output without weakening them to vague substring assertions.
- Last host q1 failed in known untouched session reattach test; no unrelated fixes authorized. Developer runs one logged quota suite after edits plus focused checks/vet/install; host independently reviews/verifies. No push or closure without host acceptance.

## M3 implementation evidence (2026-11-07)

- Gallery preview now quotes serialized Markdown with `strconv.Quote`, wraps it to the supplied rectangle, and displays `[overflow]` when the available preview rows cannot show the full value. Tiny-height layouts retain the editor and controls where space permits; preview and controls are clipped to the child width.
- Serialization and plain callback data are unchanged. Exact tests cover raw multiline `1\n2\n3`, quoted visible `\\n`, Unicode, tabs/CR/control escaping, quotes, backslashes, Markdown delimiters, width wrapping, overflow, tiny geometry and no painting outside the rectangle. CLI PTY coverage submits multiline bold content containing tab/CR/control bytes and checks the normalized quoted preview while retaining the existing styling/caret assertions.
- Focused gallery/PTY tests passed; `go vet ./...` passed; one logged `make test-q1` passed all packages, including `internal/session`; `make install` passed. Log: `/tmp/loom-issue441-m3-test-q1.log`. Implementation commit: `fe83d8f`. Host review and independent installed-binary verification remain required; do not close or push.

## M3 host verification

Host reviewed preview and assertion diffs. Installed-gallery PTY verified exactly `Submitted Markdown: "1\\n2\\n3"` for raw three-line input. A second submission verified existing normalization (tab expands to spaces, CR becomes LF, other control byte is dropped) and quoted `"A\\nB    C\\nD"` preview. Source serialization unchanged; unit tests prove raw submissions, quoting, Unicode/control escaping, bounded wrapping and overflow.

Developer M3 full suite passed. Independent host `make test-q1` failed in a different untouched session test: `TestSessionFailedScreenshotReportedOnStderr`, timeout at session_pty_test.go:836; library/gallery/CLI packages passed. Log `/tmp/loom-issue441-m3-host-test-q1.log`. No retry or unrelated test changes performed. M3 preview is installed and verified, but #441 remains open for full-suite signoff. Earlier session reattach failure did not reproduce in this host run; do not attribute this new screenshot timeout to the bold/preview change without evidence.

## Goal

/goal Deliver opt-in visual Ctrl+B bold toggling in TextEntry with reliable editing/history, bold-only Markdown serialization and actual gallery Markdown submission, while preserving existing plain APIs and key contracts; prove it in unit/per-cell/terminal tests and install. Stop and report when blocked on a user decision or denied permission. On a loop, hard architectural or OS blocker, or no progress, clear the goal and record a failure report explaining why.

## M1 implementation evidence (2026-11-07)

- Added opt-in `EnableBold()` and plain-text-preserving `MarkdownValue()` APIs; Ctrl+B remains ignored until opted in.
- Added grapheme-aligned formatting and typing style to edits, replacement, placeholder insertion, undo/redo and drawing. Format-only edits do not notify `OnChange`; clipboard stays plain.
- Added tests for mixed and selected bold, typing style, Unicode ranges, Markdown escaping and whitespace, multiline values, Ctrl+U, placeholder acceptance, callback history and highlighted selection rendering.
- Checks passed: focused bold/Markdown tests, full `TestTextEntry` tests, `go vet ./...`, refreshed public API snapshot, and the single `make test-q1` run (no `--- FAIL` lines). `make install` passed. M1 implementation commit: `22ec2fc`.
- M1 study: `docs/studies/2026-11-07-issue-441-textentry-bold-m1.md`. M2 gallery submission and terminal proof have not started; host review is required before M2.

## M2 implementation evidence (2026-11-07)

- Restored legacy non-opt-in selection styling and documented the opt-in distinction. Rebuilt Ctrl+D coverage with actual empty/text/selected/placeholder states plus enabled formatting states, comparing bold state and deep history snapshots.
- Added exact edit-path and serializer assertions. Gallery TextEntry now opts into bold, advertises Ctrl+B and displays the Markdown produced by its submission callback; widget Value and callback input remain plain.
- Focused core, gallery and PTY checks pass, including selected/typing cell styles, caret, format undo/redo, retained controls, placeholder interactions and `**sent**` submission. `go vet ./...`, the public API snapshot check and `make install` passed. The single `make test-q1` run failed only at unrelated `internal/session/TestSessionDetachPrefixDAndReattach`; its PTY output showed `BASE_SCREEN` and `DETACHED_OUTPUT` instead of the retained screen and tabs. No q1 rerun was made.

## Final signoff (2026-11-09)

Host `make test-q1` on HEAD bc3e8228 passed (exit 0, no `--- FAIL`); log `/tmp/loom-issue441-signoff-test-q1.log`. Owner confirmed bold and Markdown submission in the gallery demo. Closing.
