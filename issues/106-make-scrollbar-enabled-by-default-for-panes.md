# 106 — Make automatic scrollbars the default for panes

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `examples/ansiviewer/ansiviewer/viewer.go`,
`pane.go`, [092](092-add-mouse-drag-support-for-scrollbars.md)

## Goal

Make `scrollbar: auto` the default for panes/widgets that expose scrollable
content: show and use a scrollbar when content overflows, without reserving or
rendering one when it does not. Applications should receive this behavior
without repeating local configuration. In particular, `ansiviewer` must
automatically expose a scrollbar for long previews and file lists.

## Acceptance

- The pane/widget default is explicitly represented in the relevant defaults or
  spec source of truth as `auto`.
- `ansiviewer` displays and uses a scrollbar when its file list or preview
  exceeds the available viewport.
- Existing applications that explicitly set `true` or `false` remain
  unchanged.
- Tests cover the default, explicit opt-out, rendering, and interaction paths;
  existing scrollbar and filebrowser tests continue to pass.

## Plan (dev-106, reviewed by host)

Finding: View and Choice already show scrollbars on overflow; ansiviewer preview has none.
- M1: tests + spec `auto` default with explicit true/false override (value lives in spec/defaults.yaml only).
- M2: widget tests: overflow reserves/renders a bar, fit reserves nothing, opt-out unchanged.
- M3: ansiviewer preview scrollbar: long/short tests, track click and thumb drag move `b.offset`; implement.
Pre-Work: conventional commits `(issue 106 Mn)`; `make test-q1` after the last edit.

## M1/M2 delivered: scrollbar mode (ea66279)

`scrollbar.mode: auto|always|never` in spec; per-widget `ScrollbarMode` override on View and Choice. Host reran targeted tests: pass. Note: with `always` and fitting content the bar draws but track clicks are ignored (no offset to move) — correct.

Pre-Work for M3: ansiviewer preview uses the same `scrollbarVisible`/thumb helpers, not a copy.
