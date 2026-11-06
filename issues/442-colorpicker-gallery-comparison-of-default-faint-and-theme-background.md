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

## Goal

/goal Add an honest default/faint/theme-background visual comparison to the ColorPicker gallery, preserving one interactive picker and all existing color semantics, with per-cell/layout/input/theme and live terminal evidence; stop and report when blocked on a user decision or denied permission. On a loop, hard architectural or OS blocker, or no progress, clear the goal and record a failure report explaining why.
