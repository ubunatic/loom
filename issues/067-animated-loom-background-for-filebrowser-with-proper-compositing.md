# 067 — Animated Loom background for filebrowser with proper compositing

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: `b7a831c` (`feat: add animated background support`), [Animated Backgrounds](../docs/AnimatedBackgrounds.md)

---

## 1. Problem & Motivation

The Loom library now provides an animated Astra/Braille star-field background,
but the filebrowser example cannot use it as a visible application background
without allowing the foreground frame and boxes to overwrite the entire field.
The filebrowser should demonstrate a real Loom background rather than a
filebrowser-specific rendering workaround.

The implementation should also strengthen Loom's general background capability
so future widgets and applications can safely use animated effects without
interfering with text, borders, selections, prompts, scrollbars, mouse
interaction, or the terminal cursor.

## 2. Technical Specification / Findings

- `Pane.Background` currently renders before the root widget.
- `Frame.Draw` and `Box.Draw` fill their areas with opaque styles, so a star
  field assigned directly to the filebrowser pane is mostly hidden.
- Animated backgrounds are driven by the pane event loop at the global
  `SpeccedBackground.RedrawInterval`.
- The intended design in `docs/AnimatedBackgrounds.md` calls for a renderer
  layer that writes only to safe, unclaimed cells.
- The solution must remain compatible with normal static backgrounds, themes,
  terminal resizing, color-disabled terminals, and reduced-motion behavior when
  that setting becomes available.

## 3. Implementation & Verification Plan

### M1 — Define the background composition contract

- Decide whether Loom tracks claimed cells, exposes a protected-cell mask, or
  provides an equivalent post-render composition API.
- Keep background effects bounded to the supplied canvas/rectangle and prevent
  them from moving the cursor or writing raw terminal escapes.
- Allow effect-specific scheduling/configuration where appropriate instead of
  coupling all animated effects to Astra's global interval.

Verification: API-level tests demonstrate that a background cannot overwrite
foreground cells or cursor state.

### M2 — Implement safe animated-background composition

- Render the foreground tree and compose the background into eligible cells.
- Preserve borders, titles, status text, file rows, metadata, filter input,
  selections, scrollbars, and cursor placement.
- Handle narrow layouts, resize events, empty regions, and ANSI/color-disabled
  output.

Verification: deterministic canvas tests cover protected-cell behavior,
resizing, stable star placement, animation ticks, and cancellation/teardown.

### M3 — Integrate the filebrowser example

- Enable the Astra star field as the filebrowser's Loom background through the
  public API.
- Make its colors legible with the filebrowser themes, including `mc`.
- Keep the background independent of filebrowser input and navigation logic.

Verification: filebrowser rendering tests confirm that interactive content is
unchanged while the background animates; a terminal replay/manual smoke test
confirms visible stars behind the split panes.

### M4 — Documentation and regression coverage

- Document the public background/compositing contract and filebrowser usage.
- Add or update the background example to exercise the same composition path.
- Run `make test-q1` once after the implementation changes and include focused
  tests for each milestone before that suite run.
