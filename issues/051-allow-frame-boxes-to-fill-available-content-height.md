# 051 — Allow Frame boxes to fill available content height

**Status**: Closed — delivered via lean sprint (M1-M3), haiku dev
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `frame.go`, `frame_test.go`,
`examples/filebrowser/browser.go`, `288230f`

---

## 1. Problem & Motivation

`Frame.Layout` sizes horizontally arranged boxes from each box's preferred
`Height`. It clamps that height to the available content rows but does not offer
a way to request all available rows. Applications must know that the frame
reserves one title row and one status row, then duplicate that arithmetic.

The filebrowser initially used `Height: 16` inside a 20-row frame, leaving two
unused rows between its boxes and status bar. Removing the visual gap required
changing both boxes to the application-specific constant `Height: 18`. That
works for the example's current dimensions but does not express the actual
intent: fill the frame content area and remain flush with the status bar.

## 2. Design Questions

Add an explicit, spec-compatible way for a box or horizontal row of boxes to
consume the available frame content height. The contract must answer:

- Whether fill is a `Box` property, a height constraint mode, or a frame-level
  cross-axis alignment.
- How preferred, minimum, and maximum heights constrain a filling box.
- Whether all boxes in a horizontal row fill equally or may retain independent
  heights and alignment.
- How fill behaves when the frame switches to its stacked layout, where height
  is the primary allocation axis rather than the cross axis.
- How `ContentHeight` and `HeightForWidth` report a preferred height for a frame
  containing fill semantics without depending on the most recent draw size.

Do not assign special meaning to zero without checking compatibility with
existing validation and programmatic callers. Prefer a named intent that is
clear in both Go and YAML.

## 3. Verification & Acceptance

- Filebrowser boxes reach the status row without hard-coding the frame's
  content height.
- Resizing the available frame height grows and shrinks filling boxes while
  respecting minimum and maximum constraints.
- Horizontal and stacked layouts have documented, tested behavior.
- Hidden boxes and breakpoint transitions do not leave stale gaps.
- Existing fixed-height frames preserve their layout.
- `go test ./...` and `go vet ./...` pass.

---

## Resolved Design (host decisions, binding for the sprint)

- **Property**: a boolean `FillHeight` on `Box` (YAML `fill_height: true`), following the
  existing spec system in docs/Spec.md (spec YAML is source of truth; validate-spec must accept it).
  Zero-height semantics are unchanged; there is no special meaning for 0.
- **Horizontal row**: a filling box takes all available content rows (frame height minus the title
  and status rows), clamped by the box's min/max height constraints if the Box already has them.
  Non-filling boxes in the same row keep today's preferred-height behavior, so a row may mix both.
- **Stacked layout**: filling boxes share the height left after fixed boxes, equally, remainder to the
  earlier boxes; if nothing is left they get their minimum (or 1).
- **Preferred size**: `ContentHeight` and `HeightForWidth` treat a filling box as its preferred
  `Height` if greater than 0, else 1, never as the most recent draw size.
- **Hidden boxes** are excluded from the fill computation; breakpoint transitions recompute it.
- Existing fixed-height frames render byte-identical (assert with a golden test before changing code).

## Milestones (lean sprint, dev agent: haiku)

Host reviews only diffs and test output; this ticket is the only channel. Root-package tests
needing `/dev/tty` fail before this work; ignore them. Commit each milestone (message ends
'(issue 051 MX)'), stage only your files, never docs/README.md.

Evidence rules: frames produced by code, gated on env `LOOM_EVIDENCE=1`, written to repo-root
`docs/progress/051/` (find root by walking up to `go.mod`). Run with `LOOM_EVIDENCE=1` and confirm
`git status` shows only intended files. Run gofmt on touched files.

### M1 - FillHeight in horizontal layout
- Golden test that pins existing fixed-height layout first, then the field, spec entry, validation,
  and horizontal fill with min/max clamping and mixed rows.
- Evidence: `M1-fill-h12.ansi`, `M1-fill-h20.ansi`, `M1-fill-h30.ansi` (same frame at three terminal heights).

### M2 - Stacked, hidden, breakpoint, filebrowser
- Stacked behavior, hidden boxes, breakpoint transition, `ContentHeight`/`HeightForWidth`, all tested.
- filebrowser boxes use `FillHeight` instead of the hard-coded 18 and reach the status row; existing
  filebrowser tests stay green.
- Evidence: `M2-stacked.ansi`, `M2-breakpoint.ansi`, `M2-filebrowser-fill.ansi`.

### M1-M2 Review (host)
Delivered: 85dcc1c, e6ea79f. Tests and vet green. `FillHeight` already existed on `Box` and
the horizontal case already worked apart from min/max clamping (fixed in M1, accepted).
Defect: the stacked behavior of the Resolved Design was NOT implemented. `M2-stacked.ansi` shows
two filling boxes at 4 rows each with 10 empty rows below, so fill does nothing when stacked;
`TestFrameLayoutStackedWidthBehavior` tests widths, not fill. Also `Height: 4` in the
filebrowser is a hidden hack: in stacked mode it makes the boxes tiny.

### M3 - Pre-Work / Required Refinements
1. Test first: a stacked frame (width below Breakpoint) with two `FillHeight` boxes at heights
   12, 20, 30 must have the boxes' rects cover all content rows (height minus title and status
   rows, minus the Gap), sharing equally, remainder to the earlier box; a fixed box next to a
   filling box keeps its height and the fill box gets the rest; if nothing is left, a filling
   box gets its minimum (or 1). Then implement it in the stacked branch of `Frame.Layout`
   (frame.go around the `layout.Plan`/`balancedStackedAllocations` calls). Existing fixed-height
   and dynamic stacked behavior stays byte-identical (the golden test must stay green).
2. Regenerate `M2-stacked.ansi` so the boxes visibly fill down to the status row; add
   `M3-stacked-h12.ansi` and `M3-stacked-h30.ansi`.
3. filebrowser: remove the `Height: 4` hack (use a small sane preferred height only if
   validation requires one, and explain in your report), and prove with tests that both the wide
   and the narrow (stacked) layouts reach the status row. Evidence: `M3-filebrowser-wide.ansi` and
   `M3-filebrowser-narrow.ansi`.
4. Commit '(issue 051 M3)'; gofmt, go test ./..., go vet ./... green.
