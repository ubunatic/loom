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
