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
