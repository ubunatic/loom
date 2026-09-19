# 068 — Formalize layered compositor semantics and background inheritance

**Status**: Closed — M1-M3 delivered: explicit layered compositor API, nested-canvas surface propagation fixed, all widgets migrated off inference heuristics, PTY smoke validated
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Architecture
**Related**: `canvas.go`, `frame.go`, `background.go`,
`examples/filebrowser/filebrowser/filebrowser.go`,
`docs/AnimatedBackgrounds.md`, [067](067-animated-loom-background-for-filebrowser-with-proper-compositing.md) (Closed —
supplied the safe-cell compositing contract and the filebrowser's
`AstraBackground` wiring that this ticket now formalizes and extends)

## Problem

The current compositor infers transparency, ownership, and background
inheritance from `Cell.Text` and `Style`. This enables the filebrowser and
background demo to combine colored pane surfaces with animated star glyphs,
but it leaves policy embedded in `Canvas.Set` and makes nested canvas behavior
hard to reason about.

In particular, a default background can mean either “inherit the parent
surface” or “reset to the terminal default.” Child canvases also need to carry
surface state when their cells are merged into a parent. These meanings should
be explicit rather than inferred from blank text and reset colors.

## Goal

Define a first-class layered compositor contract for Loom:

```text
surface/background color
        ↓
foreground content and claims
        ↓
background decoration, such as stars
        ↓
cursor and selection overlays
```

The contract must support colored surfaces with animated foreground decoration,
while protecting text, borders, selections, and cursor state.

## Milestones

### M1 — Explicit cell ownership and inheritance

**Status: Complete.** `Cell.Claim` gives widgets explicit foreground
ownership independent of text/style inference, and `Canvas.Set` inherits a
parent surface's background color into a blank child cell instead of
resetting it. Wide-rune, cursor, and selection protection were already
covered by pre-existing tests.

- Define explicit foreground claim/opacity semantics instead of relying only on
  blank text and style heuristics.
- Define how default backgrounds inherit through nested canvases and how an
  explicitly colored cell background overrides inheritance.
- Preserve wide-rune continuation, cursor, and selection protection.

**Verification:** unit tests cover blank, text, colored-surface, explicitly
colored-cell, child-canvas, wide-rune, and cursor cases.

### M2 — Layered compositor API

**Status: Complete, with visual tuning deferred.** `PaintSurface` /
`PaintForeground` / `PaintDecoration` are the new explicit Canvas ops.
`AstraBackground`, `ImageBackground`, `Box`, `Frame`, `Choice`, `View`,
`Table`, and `Notif` are migrated onto them. Two real bugs were found and
fixed along the way, not just theoretical gaps:
1. The `Surface` marker didn't survive being merged up through nested
   `paintClipped` canvases (`Frame` → `Box` → child widget), so an inherited
   background color looked like claimed foreground one level up.
2. `Choice`/`View`/`Table`/`Notif` painted their background rows with raw
   `Fill(Cell{Text:" ", Style: themed})` instead of `PaintSurface`, so any
   theme with a real (non-reset) background color claimed the whole content
   area and blocked decoration — this is why stars only ever appeared in the
   `plain` theme and only in the Frame's outer chrome, never inside list/detail
   panes.

Manually verified across `mc`, `default`, `julia256`, and `plain`: stars now
render inside painted panes in all themes — confirmed working with
caveats on visual balance (contrast/density tuning) left for later polish,
not a compositing-contract gap. The headless-safe filebrowser test (no
`/dev/tty` dependency) landed as part of this milestone.

- Add focused Canvas/compositor operations for painting surfaces, foreground
  content, and background decoration.
- Keep `Background` effects independent from widget input and raw terminal
  output.
- Migrate `AstraBackground`, `Box`, and the `examples/background` and
  `examples/filebrowser` demos (both already wire
  `pane.Background = loom.NewAstraBackground()`) to the explicit API.
- Validate star contrast across the `mc`, `default`, `julia256`, and `plain`
  themes, and cover filebrowser/`Pane` lifecycle paths that today require a
  real tty (`loom.New`/`Pane.Close`) with a headless-safe smoke path so CI
  can assert animated-background behavior without `/dev/tty`.

**Verification:** deterministic canvas tests prove that each layer composes in
  order and that decoration inherits the active surface color; headless
  filebrowser coverage confirms stars render in blank list/detail areas
  without disturbing selection, filtering, or navigation; manual PTY pass
  across all four themes confirmed stars visible inside painted panes.

### M3 — Documentation and compatibility cleanup

**Status: Complete.** M1/M2 landed, and M3 documents and narrows the
contract. PTY smoke validation: ran the `background` demo through a real
Linux PTY (headless harness), 2s of live animation produced ~144KB of
redraw output with a clean exit and restored terminal state — no hang or
crash from removing the blank/reset-color inference heuristic.

- Document the layering and inheritance contract in `docs/AnimatedBackgrounds.md`
  and relevant geometry documentation.
- Remove or narrow inference helpers once all consumers use the explicit API.
- Confirm reduced-motion, resize, theme, and nested-pane behavior.
- Visual tuning pass: star contrast/density balance per theme was confirmed
  working end-to-end; Astra's low-end contrast was lifted without changing
  density or compositor semantics.

**Verification:** `make test-q1`, geometry replay, and headless example coverage;
manual PTY smoke validation for colored panes and animated stars.

## Acceptance Criteria

- Colored pane backgrounds remain visible through nested child widgets.
- Star glyphs render as foreground decoration over any pane background color.
- Explicit cell background colors remain authoritative.
- Text, borders, selections, wide glyphs, and cursor coordinates remain intact.
- The public compositor contract no longer depends on undocumented `Cell.Set`
  heuristics.
