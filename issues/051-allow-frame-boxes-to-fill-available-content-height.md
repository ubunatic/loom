# 051 — Allow Frame boxes to fill available content height

**Status**: Open
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
