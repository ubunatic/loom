# 098 — Document AnimatedBackground ticker initialization pitfall

**Status**: Open
**Priority**: P2
**Severity**: Moderate
**Category**: Documentation
**Related**: `Pane.Run`, `AnimatedBackground`, `BackgroundCadence`

---

## 1. Problem & Motivation

An application can assign an `AnimatedBackground` in response to a key event and
see the background appear without seeing any animation. The pane creates its
background redraw ticker when `Pane.Run` starts, based on the background installed
at that time. Replacing a nil background later does not start that ticker.

This is an easy-to-miss usage pitfall for examples and downstream applications
implementing optional backgrounds or runtime background switching.

## 2. Goal

Document the startup-time ticker requirement and the supported pattern for
runtime background toggles, so an `AnimatedBackground` remains animated when it
is enabled after the pane starts.

## 3. Findings

The background must implement `AnimatedBackground` (including
`DrawBackgroundAt`) and be installed before `Pane.Run`. For a default-off toggle,
keep an animated wrapper installed from startup and make its draw method a no-op
until enabled; alternatively, provide a pane redraw/ticker lifecycle API if
runtime replacement is intended to be supported directly.

## 4. Implementation & Verification Plan

- Add a concise note to the background/pane documentation and relevant API
  comments.
- Include a small example of a default-off animated background wrapper or expose
  an explicit runtime ticker refresh mechanism.
- Add a regression test covering enabling an animated background after startup.
