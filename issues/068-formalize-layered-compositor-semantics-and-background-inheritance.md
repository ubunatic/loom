# 068 — Formalize layered compositor semantics and background inheritance

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Architecture
**Related**: `canvas.go`, `frame.go`, `background.go`, `docs/AnimatedBackgrounds.md`, [067](067-animated-loom-background-for-filebrowser-with-proper-compositing.md)

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

- Define explicit foreground claim/opacity semantics instead of relying only on
  blank text and style heuristics.
- Define how default backgrounds inherit through nested canvases and how an
  explicitly colored cell background overrides inheritance.
- Preserve wide-rune continuation, cursor, and selection protection.

**Verification:** unit tests cover blank, text, colored-surface, explicitly
colored-cell, child-canvas, wide-rune, and cursor cases.

### M2 — Layered compositor API

- Add focused Canvas/compositor operations for painting surfaces, foreground
  content, and background decoration.
- Keep `Background` effects independent from widget input and raw terminal
  output.
- Migrate `AstraBackground`, `Box`, and the filebrowser/background examples to
  the explicit API.

**Verification:** deterministic canvas tests prove that each layer composes in
  order and that decoration inherits the active surface color.

### M3 — Documentation and compatibility cleanup

- Document the layering and inheritance contract in `docs/AnimatedBackgrounds.md`
  and relevant geometry documentation.
- Remove or narrow inference helpers once all consumers use the explicit API.
- Confirm reduced-motion, resize, theme, and nested-pane behavior.

**Verification:** `make test-q1`, geometry replay, and headless example coverage;
manual PTY smoke validation for colored panes and animated stars.

## Acceptance Criteria

- Colored pane backgrounds remain visible through nested child widgets.
- Star glyphs render as foreground decoration over any pane background color.
- Explicit cell background colors remain authoritative.
- Text, borders, selections, wide glyphs, and cursor coordinates remain intact.
- The public compositor contract no longer depends on undocumented `Cell.Set`
  heuristics.
