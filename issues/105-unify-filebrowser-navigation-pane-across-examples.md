# 105 — Unify filebrowser navigation pane across examples

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `examples/ansiviewer/ansiviewer/viewer.go`,
`examples/filebrowser/filebrowser/browser.go`,
`examples/filebrowser/filebrowser/filebrowser.go`,
[063](063-convert-filebrowser-example-to-a-hostable-widget.md)

## Goal

Provide one shared filebrowser navigation-pane implementation and use it in
all file-browser examples, so navigation behavior and interaction contracts do
not diverge between `ansiviewer`, `filebrowser`, and future consumers.

## Required behavior

- ESC goes back one directory; at the starting/filesystem root it retains the
  established quit behavior.
- Returning to a parent lands with the directory just left selected.
- `/` enters file filtering/search mode and filters the visible files.
- Mouse-clicking an item selects it and updates the preview/metadata.
- Keyboard navigation, directory opening, preview handling, and quit behavior
  remain available through the shared pane API.

## Acceptance

- The examples no longer maintain separate copies of the filebrowser
  navigation logic.
- Shared tests cover ESC back, selection restoration, slash filtering, and
  mouse selection, including the hosted/pane key-consumption boundary.
- Existing standalone and hosted example tests pass, and a manual smoke pass
  confirms the same interactions in each example.

When work starts, verify the live implementations and recent history first;
the current examples may have evolved beyond the paths listed above.
