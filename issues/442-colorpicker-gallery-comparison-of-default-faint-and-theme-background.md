# 442 — ColorPicker gallery comparison of default faint and theme background

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 350, 347, 361

---

## 1. Problem & Motivation

Owner wants dim foreground/background options beside ∅ but is unsure whether this means ANSI faint or theme-relative colors. Owner agreed to a gallery visual comparison before deciding color-model or save-format changes. Existing ∅ means terminal reset; Style.Dim is a text attribute, not an independent dim-background color.

## 2. Decision / Technical Specification

Keep ONE existing interactive ColorPicker and its FG/BG, reset and navigation behavior. Add static identically worded samples: Default uses reset FG/BG; Faint text uses reset FG/BG plus Dim; Subdued BG (theme) uses reset FG plus current theme.NormalBG.Color(). Explicitly identify NormalBG as a theme-surface illustration, not semantic dim-background support. When NormalBG is reset, sample may look identical to Default; label/docs must not promise a difference or saved semantic color. Do not invent RGB blends, new theme roles/encoding or persistence changes.

Use a gallery-only wrapper around existing picker: theme propagation updates both samples and child without resetting selection; focus/measure/key results delegate. No Unwrap that bypasses ApplyTheme. Picker draws first; samples below, side by side at sufficient width, stacked when narrow. Clip to child rect including nonzero offsets and zero sizes. Samples do not capture keys or mouse; only events within drawn picker child rect are translated once and forwarded. Do not add hardware cursor changes.

## 3. Implementation & Verification Plan

Approved write set: new `gallery/colorpicker.go`, `gallery/gallery.go`, `gallery/gallery_test.go`, `cmd/loom/widgets_colorpicker_pty_test.go` if a new focused CLI test is useful (or existing gallery PTY test file), `docs/Widgets.md`, `docs/GalleryCoverage.md` if useful, this ticket and one bounded sprint study. No library/style/theme schema or save-format changes. Developer terra:med, host reviewer; Searcher and developer read-only plan completed against live code.

Regression-first tests: exact sample FG/BG/Dim/ownership, wide and narrow layout, all samples visible when space permits, zero/tiny/offset bounds, mouse coordinates and noninteractive samples, focus/key/reset behavior, F9 theme refresh without changing selected color and fallback reset NormalBG. Actual CLI PTY checks labels/dim attribute/navigation/theme refresh; preserve existing picker tests. Run focused checks/vet and one logged make test-q1, install before reporting. Last independent host suite failed in untouched session screenshot timeout; do not change that unrelated test or claim a baseline green. Keep issue open until host acceptance. No external edits/push/helpers.

## M2 — Host review: fix child geometry in library

### Pre-Work / Required Refinements

Host reviewed implementation and assertions. Developer reported ColorPicker itself paints outside tiny supplied rectangles, then added SubCanvas/Blit around it only in this demo. This is a library flaw; AGENTS.md requires fixing the library rather than a gallery-only workaround.

- Explicitly expand approved write set to `colorpicker.go` and `colorpicker_test.go` for a narrow bounds fix only; no new color semantics, APIs, theme roles or format changes. First reproduce standalone ColorPicker.Draw on a larger sentinel canvas with tiny/nonzero rect (width/height 0,1,2 and normal case), asserting exact untouched cells outside bounds and no wide glyph crossing edges.
- Fix the originating library Draw/helper geometry so all callers are safe; preserve normal layout, key/reset and child-local mouse contract. Remove the gallery's special SubCanvas/Blit containment if no longer needed; direct-child offset/tiny gallery tests must still pass. Do not weaken assertions.
- Preserve all comparison exact styles, labels and F9 theme behavior. Check the CLI color comparator handles the actual expected theme color representation honestly; do not infer arbitrary indexed/reset colors are RGB.
- One full logged quota run after source edits, focused checks/vet/install, update issue/study and commit. M1 developer final suite passed but no host full verification yet. No unrelated session fixes, pushes or closure. Host independently verifies final source and installed binary.

### Regression-First Evidence

Before library edits, the standalone sentinel-canvas test reproduced `ColorPicker.Draw` clearing cells below supplied offset rectangles of height 1 and 2. A zero-width draw also retained the prior mouse hitbox. The M2 patch bounds the preview-gap and swatch-padding writes to the local height and records empty draw geometry. The gallery's `SubCanvas`/`Blit` workaround has been removed; its direct-child tiny/offset assertions remain strict.

M2 focused tests, vet, the full quota run and installation are pending in the implementation commit. The ANSI PTY comparator checks RGB only when both expected and observed colors expose RGB, and fails rather than treating indexed or reset colors as RGB.

## M3 — Review refinements

- Reproduce and resolve boundary writes when a pre-existing wide glyph straddles either horizontal edge of the picker; preserve outside cells and avoid a gallery-only workaround. Verify the canvas contract and avoid introducing invalid wide-cell state.
- Strengthen the CLI PTY test to navigate the picker and verify selection survives F9 theme refresh. Keep all existing assertions.
- M2 focused tests, vet and install passed; quota suite was blocked before tests by bubblewrap. Full suite and visual acceptance remain owed.

## Goal

/goal Add an honest default/faint/theme-background visual comparison to the ColorPicker gallery, preserving one interactive picker and all existing color semantics, with per-cell/layout/input/theme and live terminal evidence; stop and report when blocked on a user decision or denied permission. On a loop, hard architectural or OS blocker, or no progress, clear the goal and record a failure report explaining why.
