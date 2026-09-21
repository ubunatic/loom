# 106 — Make scrollbar enabled by default for panes

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `examples/ansiviewer/ansiviewer/viewer.go`,
`pane.go`, [092](092-add-mouse-drag-support-for-scrollbars.md)

## Goal

Make `scrollbar: true` the default for panes/widgets that expose scrollable
content, so applications receive usable scrolling without repeating local
configuration. In particular, `ansiviewer` must visibly expose a scrollbar for
long previews and file lists.

## Acceptance

- The pane/widget default is explicitly represented in the relevant defaults or
  spec source of truth as enabled.
- `ansiviewer` displays and uses a scrollbar when its file list or preview
  exceeds the available viewport.
- Existing applications that explicitly disable scrollbars remain unchanged.
- Tests cover the default, explicit opt-out, rendering, and interaction paths;
  existing scrollbar and filebrowser tests continue to pass.
