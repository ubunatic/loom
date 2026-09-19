# 069 — Render help as a root-level modal overlay

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [066](066-nested-help-pane-opens-a-second-pane-racing-the-outer-pane-tty-reader.md), `cmd.go`, `popup.go`, `frame.go`

---

## 1. Problem & Motivation

The current default `:help` implementation creates an inline `Popup` from the
`Choice` or `Table` that owns the command bar. In split layouts such as the
file browser, `paintClipped` correctly confines that child to its pane, so help
appears only inside the left child pane rather than as a modal over the complete
Loom root.

Issue 066 removed the unsafe nested `Pane`/TTY path. The remaining usability
gap is that help is not globally modal across composite layouts.

## 2. Desired Behavior

Help should be rendered and handled at the root-pane level:

- center the help popup within the full active `Pane` canvas;
- capture keyboard and mouse events before the underlying widget tree;
- dismiss on the documented close keys without changing the underlying layout;
- preserve the single-pane ownership guarantee from issue 066;
- keep headless rendering and injected help-runner tests working.

## 3. Implementation & Verification Plan

1. Add a root-level overlay contract or equivalent event-loop plumbing so a
   descendant command bar can request a modal overlay without drawing outside
   its `paintClipped` region.
2. Lift the help popup state to the active root/Pane and route draw, key, and
   mouse handling through that overlay before the normal widget tree.
3. Add headless tests for split/composite roots proving the popup uses the full
   canvas bounds and captures input.
4. Run `go test ./...`, `go test -race ./...`, and `go vet ./...`, plus a PTY
   smoke test in the file browser covering `:help`, dismissal, resize, and
   terminal restoration.
