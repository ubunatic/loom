# 092 — Add mouse-drag support for scrollbars

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Enhancement
**Related**: None

---

## Problem & Motivation

Scrollbars need direct manipulation so users can quickly navigate long content
with a mouse instead of relying only on wheel events, keys, or paging actions.

## Scope

- Make a scrollbar thumb draggable with the primary mouse button.
- Convert pointer movement into proportional scroll-offset updates for both
  vertical and horizontal scrollbars where supported.
- Keep the thumb attached to the pointer throughout the drag, including when
  the pointer leaves the thumb's original cell, and end the drag on release or
  cancellation.
- Define sensible behavior for track clicks, minimum thumb sizes, content that
  does not overflow, viewport edges, and disabled/read-only scrollbars.
- Preserve existing wheel, keyboard, paging, and scrollbar activation behavior.

## Goal

/goal: Let users smoothly drag scrollbar thumbs with the mouse to control
scroll position, with consistent vertical/horizontal behavior and no regression
to existing scrolling interactions.

## Verification

Add focused coverage for drag start/move/release/cancel, pointer-to-offset
mapping, bounds and non-overflow cases, both orientations, and interaction with
existing wheel and keyboard scrolling.
Extend the PTY tests to drag scrollbar thumbs and prove scrolling changes the
visible content through a real terminal session.

---

## Resolved Design (host decisions, binding for the sprint)

- Find the existing scrollbar type and its mouse handling first; extend it, add no second scrollbar.
- One pure mapping function shared by both orientations: thumb geometry (track length, thumb length,
  content length, viewport length) to offset and back, integers, rounding to nearest, clamped to
  [0, max]. Table-tested, including minimum thumb size 1, thumb equal to track (no overflow), content
  smaller than viewport, and length-1 tracks.
- Drag state machine: press on the thumb with the primary button starts a drag and stores the grab
  offset (pointer minus thumb start) so the thumb stays attached to the pointer; motion events update the
  offset even when the pointer leaves the thumb, the scrollbar or the widget (capture until release);
  release ends the drag; a cancel (Esc, or losing focus/capture) ends the drag and restores the offset
  from the drag start. Non-primary buttons never start a drag.
- Track click (not on the thumb) keeps its existing behavior; if none exists, it pages one viewport
  toward the click. Content that does not overflow: no thumb, no drag. Disabled or read-only: ignore mouse.
- Wheel, keyboard and paging behavior stay byte-identical (existing tests unchanged).
- Vertical and horizontal go through the same code path with the axis as a parameter.

## Milestones (lean sprint, developer: luna:low)

Host reviews only diffs, test output and evidence frames; this ticket is the only channel. Root-package tests
needing `/dev/tty` fail before this work; ignore them. Commit each milestone (message ends '(issue 092 MX)'),
stage only your files, never docs/README.md, no stray binaries in the repo root; if `.git/index.lock` blocks the
commit, stage your files and say so (the host commits). Evidence: frames produced by code, gated on env
`LOOM_EVIDENCE=1`, written to repo-root `docs/progress/092/` (find root by walking up to `go.mod`); run only this
ticket's evidence test; view frames ANSI-stripped before finishing; labels on their own rows. gofmt. Do not leave
dead or commented-out code.

### M1 - Mapping and drag state machine
- Pure mapping function plus drag state machine, table and sequence tests (press, move outside, release,
  cancel, non-primary button, no-overflow, disabled).

### M2 - Wire into the scrollbar widget, both orientations
- Extend the real scrollbar and a scrollable widget that uses it; synthetic mouse events drive drags; existing
  wheel/key/paging tests unchanged and green.
- Evidence: `M2-vertical-before.ansi`, `M2-vertical-mid.ansi`, `M2-vertical-after.ansi`,
  `M2-horizontal-before.ansi`, `M2-horizontal-mid.ansi`, `M2-horizontal-after.ansi` (thumb and visible content change).

### M3 - PTY drag
- Extend the PTY tests (existing helper in `internal/ptytest`) to drag a thumb in a real terminal session and
  assert the visible content changes; evidence `M3-pty-drag.ansi` (final screen).

### M1-M2 Review (host)
Code committed by the host (index.lock and quota_1.state blocked the developer). `go test .` and vet
green. Accepted: shared axis-neutral mapping and drag state, wiring in `View` and `Choice`, cancel paths.
Missing: all evidence, the PTY test, and proof of the Esc-cancel path. No widget has a horizontal
scrollbar, so horizontal frames are dropped from the deliverables (mapping-level tests stay).

### M3 - Pre-Work and Evidence
1. Tests: Esc during a drag restores the drag-start offset for `View` and `Choice`; a drag that
   leaves the widget bounds keeps tracking; a non-primary button press on the thumb starts nothing.
   Add whichever of these is missing.
2. Evidence test (gated on `LOOM_EVIDENCE=1`, repo-root `docs/progress/092/`, run only this test):
   `M2-vertical-before.ansi`, `M2-vertical-mid.ansi`, `M2-vertical-after.ansi` from a `View` with long content
   (thumb position and visible lines change; label rows above the frame; ANSI-stripped check before finishing);
   the same trio for a `Choice` as `M2-choice-before/mid/after.ansi`.
3. PTY test with `internal/ptytest` (see existing PTY tests): drag the thumb of a scrollable example in a real
   terminal session and assert the visible content changes; evidence `M3-pty-drag.ansi`.
4. Commit '(issue 092 M3)'; if git is blocked say so and leave files staged.

### M3 Review (host)
Committed by the host. Accepted: the three cancel/capture/button tests and all six `M2-*` frames (thumb and
content visibly move). Still missing: the PTY drag test and `M3-pty-drag.ansi`.

### M4 - PTY drag (architecture given by the host)
Use an existing example whose scrolling pane is a `View` or `Choice` with a scrollbar; the ansiviewer preview
pane or the filebrowser list are the candidates (pick one, say which). Launch it in a real PTY with
`internal/ptytest` on a long temp file or directory (at least 3x the viewport). Send the drag as raw SGR mouse
bytes on the PTY: press `ESC[<0;X;YM`, motion `ESC[<32;X;YM` with growing Y, release `ESC[<0;X;Ym`, where X is
the scrollbar column and Y walks down the track (find X and Y from the drawn frame, do not hardcode a guess:
locate the thumb glyph in the screen the PTY returns). Make sure the example enabled mouse tracking; if it
did not, that is the bug to report. Assert that the visible text after the drag differs from before and that
the first visible line moved forward. Write `M3-pty-drag.ansi` from the final screen (ANSI-stripped check).
If the ptytest helper lacks a way to write raw bytes or read the screen, add the smallest helper and test it.
Commit '(issue 092 M4)'; if git is blocked say so and leave files staged.

### M4 Review (host)
Not committed: the developer added `ptytest.Session.SendRaw` (uncommitted, `internal/ptytest/session.go`) and
`examples/split/split/split_pty_test.go` (uncommitted), but the PTY test is RED: raw SGR input is accepted and
the visible `View` content does not move. Multi-layer problem (PTY, mouse decode, routing through the hosted
widgets, View drag), so the step moves up the escalation ladder (`codex:sol:low`).

### M5 - Diagnose and finish the PTY drag (host guidance)
Bisect by layers, cheapest first, and keep each layer's test:
1. Headless: drive the same widget tree the split example builds (`NewWidget`) with synthetic `MouseEvent`
   press/drag/release at the thumb cell and check the View offset. If it moves, the widget is fine.
2. Decode: check that the raw bytes the test sends decode to the same `MouseEvent` kinds (`MousePress`, `MouseDrag`,
   `MouseRelease`, button, coordinates 1-based vs 0-based) as the synthetic events; note that mode 1003 vs
   1002 reporting and the motion bit 32 matter.
3. Routing: the split (divider drag), the frame or the pane may consume drag events before the View sees
   them; find where drag events are captured and whether the View's scrollbar column is hit-tested with the
   right offsets (pane origin, frame border, title row).
Fix a genuine bug in library code if it is small and covered by a failing-then-passing test in the layer
where it lives; if the split example is the wrong vehicle (its divider owns drags), switch the PTY test to
the ansiviewer preview or the filebrowser list and say why. Final: the PTY test is green, `M3-pty-drag.ansi`
written from the final screen (ANSI-stripped check), `go vet ./...` and touched packages green, and every
bug found is listed in your report. Commit '(issue 092 M5)'; if git is blocked say so and leave files staged.

### M5 Review (host)
`sol:low` found two real bugs (uncommitted in the working tree): `Split` dropped drag/release events outside
the child bounds (routing layer, capture added in `split.go`), and `View` scrollbar hit-testing used inconsistent
edge coordinates (`view.go`). Decode layer is correct. The split PTY test is STILL RED and no evidence frame exists.
Escalation ladder: next is native Sonnet.

### M6 - Finish (native Sonnet)
Start from the working tree as it is (review the `split.go` and `view.go` changes, keep what is right, each with a
regression test in its layer). Find why the split example's View scrollbar still does not move in the real PTY:
compare the coordinates of the thumb you locate on the PTY screen with the View's own hit-test rect (pane origin,
frame border, title row, 1-based SGR vs 0-based cells), and check for a second consumer of the drag (split divider
hit area, focus handling: does a View need focus before it handles a press?). If the split example is the wrong
vehicle, use ansiviewer or the filebrowser list. Done when: PTY test green, `M3-pty-drag.ansi` written from the final
screen, go vet and all touched packages green, commit '(issue 092 M6)' with only your files. Report every bug.
