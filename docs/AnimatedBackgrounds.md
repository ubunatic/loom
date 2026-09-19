---
title: Animated Backgrounds
weight: 42
---

# Animated Backgrounds

This document records the design of the GPT-6 Astra composer effect in the
OpenAI Codex CLI as a reference for Loom custom animated backgrounds.

## Reference implementation

The effect is implemented in the open-source Codex repository:

- [sparkle.rs](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/chat_composer/sparkle.rs)
  owns eligibility, lifecycle state, input handling, and redraw scheduling.
- [sparkle_field.rs](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/chat_composer/sparkle_field.rs)
  paints the deterministic star field.

The renderer uses eight single-cell Braille glyphs:

```rust
["⠁", "⠂", "⠄", "⠈", "⠐", "⠠", "⡀", "⢀"]
```

These are not eight animation frames for one star. Each blank eligible cell
gets one glyph selected from the list by a stable hash of its position. The
result is a distributed field of different single-dot patterns that remains
stable between redraws while still looking textured.

## Rendering model

The effect follows this pipeline:

```text
model/settings/state
        │
        ▼
eligible background cells ──► stable cell hash ──► glyph + color ──► frame buffer
        │                                                            │
        └──────────── protected-content / cursor exclusion ◄─────────┘
```

- Activation is model-gated: the model name is split into alphanumeric parts
  and the effect is enabled when one part equals `astra`, case-insensitively.
- It also requires both animation and whimsy settings to be enabled.
- It is intended for a fresh, untouched composer. Ordinary user input ends the
  opportunity; a slash command can remain eligible until it is accepted or
  dismissed.
- The field is painted only into blank, unstyled cells.
- Placeholder text, typed text, borders, selections, popups, and the terminal
  cursor are protected. Background decoration must never overwrite interactive
  content or alter cursor placement.
- Each cell's position is hashed, then the hash selects one of the eight
  glyphs. The hash must be independent of frame time so the field does not
  visibly shuffle on every redraw.
- Color is derived from the active foreground/theme and blended toward a
  background RGB value. The effect fades by reducing contrast rather than
  adding more glyphs.

The Codex reference schedules a redraw every 150 ms, remains visible for 15
seconds, then fades for one second. Hidden or ineligible frames do not keep a
timer or redraw loop alive.

## Loom extension shape

Loom should treat a background as a renderer-layer effect, not as a widget that
owns input state. A useful interface is conceptually:

```go
type BackgroundEffect interface {
    Draw(area Rect, frame Frame, now time.Time, dst *Canvas)
    Active() bool
}
```

The runtime should own:

- fixed-rate frame scheduling and cancellation;
- the immutable animation snapshot or frame timestamp;
- effect selection and settings;
- redraw requests; and
- the protected-cell mask supplied by the normal layout/render pass.

An effect should only receive a bounded rectangle and write into cells already
classified as safe. It should not inspect or mutate input buffers, move the
cursor, write raw terminal escape sequences, or maintain a package-global
timer. The normal widget render should run first, followed by the background
effect only for cells that remain unclaimed. Loom treats a blank cell with the
default background and no foreground attributes as a transparent foreground
surface: the effect may add its glyph there while preserving the cell's
background color. Text, explicit background colors, foreground attributes,
borders, selections, and the cursor remain protected. Composition never
changes cursor coordinates or selection state.

## Custom-effect contract

Future effects can use the same contract with different glyph and color
strategies:

- **Deterministic texture**: hash `(seed, x, y)` and choose a glyph, as Astra
  does. Good for stars, noise, and static-looking ambience.
- **Animated texture**: add a quantized frame index to the hash only when a
  moving field is wanted. Keep the frame interval explicit and bounded.
- **Particles**: derive positions from a seed and elapsed time, then clip all
  particles to the safe-cell mask.
- **Theme-aware color**: blend within the theme's approved contrast range;
  never assume a dark terminal background.

Background color and foreground decoration are independent. A consumer may
paint a colored surface first and then compose the star field over it; the
colored surface remains protected while transparent blank cells elsewhere can
still show stars. The standalone background demo and filebrowser example use
this same compositor model.

Every effect should define its activation condition, tick interval, lifetime,
fade behavior, and cancellation triggers. The default should be off unless a
consumer explicitly enables it, and a reduced-motion/accessibility setting
must disable it without affecting normal input or rendering.

## Verification requirements

Tests and terminal replay fixtures should verify:

- exact cell width for every glyph, including all Braille patterns;
- no writes over text, placeholder, border, selection, or cursor cells;
- stable glyph placement for the same seed, area, and frame;
- deterministic fade and timeout behavior;
- no redraw requests after cancellation or expiry;
- unchanged cursor position while the effect redraws; and
- graceful behavior at narrow, resized, and color-disabled terminals.

The upstream Codex regression report documents why these boundaries matter:
[Astra composer sparkle animation prevents mouse text selection](https://github.com/openai/codex/issues/44398).
