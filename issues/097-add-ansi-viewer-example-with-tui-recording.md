# 097 — Add ANSI viewer example with TUI recording

**Status**: Open
**Priority**: P2
**Severity**: Moderate
**Category**: Feature
**Related**: `examples/loom-demo`, `docs/data/harnez-usage.ansi`

---

## Goal

Add an `examples/ansiviewer` TUI that browses a directory in a left-hand file
pane and renders the selected file in a right-hand viewer pane without
overflow. Text files are shown as plain text, binary and image-like files show
metadata, and `.ansi` files retain their full color and ANSI styling. Extend
the Loom widgets safely as needed and include the viewer in `loom-demo`.

The viewer must also support `ansiviewer --record <time>`: launch a TUI,
capture one bounded screenshot after the requested delay, and cleanly finish.
Recording must work for long-lived TUIs such as `go run ./examples/splash
--watch` (closing the test app after capture) and for self-exiting TUIs such as
`go run ./examples/splash`.

## Acceptance Criteria

- `ansiviewer <dir>` starts the browser rooted at `<dir>` with usable keyboard
  navigation and a left file browser/right viewer layout.
- Plain-text files, metadata-only binary/image files, and colored `.ansi`
  files are rendered according to the goal; ANSI output remains inside the
  viewer pane and does not overflow its bounds.
- The example is available through `loom-demo` and has focused automated
  coverage for file classification, rendering bounds, and key interactions.
- `ansiviewer --record <time>` captures exactly one post-delay TUI snapshot,
  handles both long-lived and self-exiting subprocesses, and terminates or
  reaps the test app cleanly.
- Existing Loom examples and tests continue to pass.

## Notes

Investigate the current widget and terminal-capture APIs before implementation;
do not duplicate spec values or introduce unsafe process handling. Resolve any
format or platform limitations discovered during implementation in the ticket
or accompanying documentation.
