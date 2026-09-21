# 093 — Restrict filebrowser mouse interaction to item content

**Status**: Closed — delivered via lean sprint (M1-M5), luna:low
**Priority**: P1
**Severity**: Major
**Category**: Bug
**Related**: None

---

## Problem & Motivation

In the filebrowser, clicking empty space in an item's row currently can select
or activate that item. This makes hit behavior surprising, especially in wide
rows where the pointer is far from the visible item.

## Scope

Restrict mouse interaction to the item's actual interactive region: item text
and any visible frame or decoration belonging to that item. Clicking whitespace
elsewhere in the row must not select, activate, or change the current item.

Preserve normal keyboard navigation, row layout, selection state, and intended
interactions for item decorations. Define behavior for truncated text, icons,
multi-column metadata, and rows with no visible interactive content.

## Goal

/goal: Make filebrowser mouse hit-testing precise so only an item's text, frame,
or decoration responds to clicks, while empty row space is inert and cannot
change selection or activation state.

## Verification

Add focused coverage for clicks on text, whitespace before/after text, frames,
icons/decorations, truncated labels, and neighboring rows. Verify selection and
activation remain correct for mouse and keyboard interaction.
Extend the PTY tests to click item content and row whitespace and prove that
only the intended hit regions change selection or activation.

---

## Resolved Design (host decisions, binding for the sprint)

- First reproduce: a failing test that clicks empty row space in the filebrowser list and changes the selection.
- Hit region of an item row: one contiguous span from the first cell of the item's visible content (marker,
  icon or decoration included) to the last non-blank cell of the item's rendered text, measured with display
  width (runes, wide runes, combining marks). Leading indentation and trailing padding are inert; gaps between
  columns inside that span count as inside. A truncated label ends at its ellipsis. A row with no visible content
  is inert. Only rows inside the list's bounds respond; neighbouring rows are never affected.
- Fix it in the layer that owns the behavior (probably `Choice.HandleMouse` in the library, used by the
  filebrowser list). Do not change other Choice users: make it opt-in through a field (name it
  `MouseTextOnly`), unless every existing test still passes with it as the default, in which case make it the
  default and say so. The filebrowser enables it.
- Keyboard navigation, wheel, scrollbar drag (ticket 092) and activation by Enter stay unchanged.
- Coordinate pitfalls found in ticket 092: mouse events reaching widgets are 0-based; PTY SGR sequences are
  1-based; use rune counts or display widths, never byte offsets, when locating text on a screen.

## Milestones (lean sprint, developer: luna:low)

Host reviews only diffs, test output and evidence frames; this ticket is the only channel. Root-package tests
needing `/dev/tty` fail before this work; ignore them. Commit each milestone (message ends '(issue 093 MX)'),
stage only your files, never docs/README.md, no stray binaries in the repo root; if `.git/index.lock` blocks the
commit, stage your files and say so (the host commits). Evidence: frames produced by code, gated on env
`LOOM_EVIDENCE=1`, written to repo-root `docs/progress/093/` (find the root by walking up to `go.mod`); run only
this ticket's evidence test; do not run evidence generation across `./...`; view frames ANSI-stripped before
finishing; labels on their own rows. gofmt. No dead code.

### M1 - Reproduction and hit region
- Failing reproduction test first, then the hit-region function (pure, display-width aware) and the opt-in
  field, table-tested: text, whitespace before and after, icon or marker, truncated label, CJK label, neighbour row,
  empty row.

### M2 - Wire into Choice and the filebrowser
- `Choice.HandleMouse` uses the hit region when the field is set; the filebrowser enables it; existing tests green.
- Evidence: `M2-hit-overlay.ansi` (the filebrowser list with the hit cells of every row marked, for example with a
  reverse-video or colored background, plus a label row explaining the marking) and `M2-click-before.ansi` /
  `M2-click-after-text.ansi` / `M2-click-after-whitespace.ansi` (selection changes only for the text click).

### M3 - PTY test
- With `internal/ptytest` and its `SendRaw` helper: launch the filebrowser on a temp dir, click on item text
  (selection or activation changes), click on row whitespace (nothing changes). Locate the screen position from
  the returned screen by rune or display width. Evidence: `M3-pty-click.ansi` (final screen).

### M1-M3 Review (host)
Code and PTY test committed by the host (index.lock). Tests and vet green. Accepted: `MouseTextOnly`, display-width
hit region, filebrowser wiring, PTY click test. Notes:
- `Choice.Draw` now truncates with `measure.Truncate(line, w, "…")` for ALL Choice users (before: a rune cut).
  Correct for wide runes, but it is a visible behavior change outside this ticket; keep it only with a test
  that pins the new truncation (wide rune and ellipsis) and mention it in your report.
- The hit test rebuilds the row text (marker, name, desc) a second time inside `HandleMouse`. Refactor so `Draw`
  and `HandleMouse` share one function that builds the displayed row string (no duplicated formatting).
- Missing: the M2 evidence frames (`M2-hit-overlay.ansi`, `M2-click-before.ansi`, `M2-click-after-text.ansi`,
  `M2-click-after-whitespace.ansi`). `M3-pty-click.ansi` is full of animated-background braille noise; switch the
  background animation off for the evidence run if the browser allows it (see the existing background test),
  otherwise render it with a fixed frame, so the list and metadata are readable.

### M4 - Pre-Work
1. Shared row-text function used by `Draw` and `HandleMouse` (test that both agree).
2. Pin the new truncation with a test.
3. The four M2 evidence frames from the real filebrowser, headless, gated as usual; overlay marks hit cells.
4. Regenerate `M3-pty-click.ansi` without background noise; also assert in the PTY test that the whitespace click
   leaves the selection unchanged and the text click changes it (state read from the screen).
5. Commit '(issue 093 M4)'; if git is blocked say so and leave files staged.

### M4 Review (host)
Committed by the host. Accepted: shared row text, truncation pin, overlay frame, clean PTY frame. One evidence flaw:
`M2-click-after-whitespace.ansi` clicks the whitespace of the row that is ALREADY selected from the text click, so the
unchanged selection proves nothing (a buggy row-wide hit would look the same).

### M5 - Pre-Work
Change the evidence sequence: `before` (selection on row 0) -> whitespace click on the trailing space of ANOTHER
row (row 2, far right of its text): selection must stay on row 0 (`M2-click-after-whitespace.ansi`, assert it in the
test too) -> text click on row 2: selection moves (`M2-click-after-text.ansi`). Update the frame labels. Regenerate only
these two frames. Commit '(issue 093 M5)'; if git is blocked say so and leave files staged.
