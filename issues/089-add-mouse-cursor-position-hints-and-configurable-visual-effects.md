# 089 — Add mouse cursor position hints and configurable visual effects

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Enhancement
**Related**: None

---

## Problem & Motivation

Paint cells currently lack a reusable hint about how close they are to the
mouse cursor. Widgets and applications need that spatial information to render
subtle, configurable cursor feedback without each one reimplementing pointer
tracking or distance calculations.

## Scope

When the mouse moves over an app, expose to paint cells the cursor's relative
`dx,dy` position for a configurable radius of `0..n` character cells. The first
MVP uses this information to adjust the background at surrounding positions:

- Theme 1: brighten the existing background, brightest at the cursor and fading
  outward. The radius `n` and brightness level must be specified in `spec/`.
- Theme 2: render a trailing star around the cursor.
- Theme 3: on a button or key press, emit one bright pulse expanding from the
  cursor outward.

Keep cursor-coordinate propagation, distance/radius semantics, theme selection,
and rendering effects separable so future themes can reuse the same hint. Define
behavior at boundaries, during motion, and when no cursor is present.

## Goal

/goal: Provide paint cells with a reliable, spec-driven mouse proximity hint and
implement the three configurable cursor effects—background brightening,
trailing star, and one-shot expanding pulse—without disrupting normal painting
or existing mouse interaction.

## Verification

Add focused coverage for cursor movement, `dx,dy` and radius behavior, edge and
missing-cursor cases, spec configuration, and each theme's timing/intensity
behavior. Verify representative widgets or example apps show the effects while
preserving their existing input behavior.
Extend the PTY tests to drive mouse movement and button/key events and prove the
cursor hints and visual effects work through a real terminal session.

## Implementation Plan (dev-089, reviewed by the host)
- M1: the hint (cursor dx,dy and radius; edges; no cursor) and the brighten theme, with spec/schema
  fields for radius and brightness loaded like theme.go/resize.go. Tests first.
- M2: trailing star. M3: press pulse, one-shot. M4: widget/app coverage and a PTY test.

Pre-work / Required Refinements for M1:
- Motion without a pressed button is only reported in any-event mouse mode (DECSET 1003). Check
  which mode Pane enables; if only 1002, enable 1003 when an effect is on, and reset it on exit.
- Effects are off by default: apps that don't opt in paint exactly as before (test it).
- The effect must not change widget-local mouse coordinates or event dispatch.
- The hint is cleared when the pointer leaves the window or no motion is known yet.

M1 delivered (555ed4f): hint and brighten theme. `Pane.EnableCursorProximity()` opts in (any-motion
mode 1003); spec/cursor.yaml has radius 3 and brightness 0.7; off unless the app opts in.

Pre-Work / Required Refinements for M2:
- Brighten skips claimed cells, so the glow shows only on empty background, not behind text.
  Brighten the background of text cells too, or state in the doc comment why not; test either way.
- Indexed backgrounds get no glow (only RGB works). Convert indexed colors to RGB via the palette if
  Loom has one; the terminal default background may stay unchanged, documented.
- Add an opt-in flag to one example (e.g. loom-demo or paint: `--cursor-fx`) so the user can see the
  effect early. The strength (0.7) is a user decision, checked by eye.

M2 delivered (cc18b91): star trail (spec: glyph ✦, 8 points, 400 ms, 40 ms frames), brighten now
covers text cells and indexed colors, and `ansicanvas_demo --cursor-fx` shows both effects. The
dev's test failure (TestAllAnsiAssetsHaveValidBoxes) came from the host's ruler mockup and is fixed
in 81d5d6b; the host's full run there passed with M2 included.

Pre-Work / Required Refinements for M3:
- The 40 ms trail ticker redraws the whole screen forever once enabled, even with no trail. Redraw
  only while unexpired points exist (test: no frames scheduled when the trail is empty or expired).
  The pulse animation must follow the same rule.
- The trail fades its color toward black, which looks wrong on light backgrounds. Fade toward the
  cell's background instead (RGB when resolvable), with a test.

M3 delivered (f8f2c1c): one-shot press pulse on button and key presses; the frame ticker runs only
while a trail point or pulse is alive; stars fade toward the cell background. `--cursor-fx` in
ansicanvas_demo turns on all three effects.

Pre-Work / Required Refinements for M4:
- Add the manual check to issues/102: `ansicanvas_demo --cursor-fx` in tilix and foot (glow
  strength, trail, pulse, no lag while moving, idle CPU near zero when the mouse rests).
- PTY test: drive motion, a click and a key through a real terminal session and assert the effect
  cells appear and expire, and that the click still reaches the widget at the right cell.
