# 090 — Support image-backed app backgrounds and background theme switching

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Enhancement
**Related**: None

---

## Problem & Motivation

Loom apps can paint Astra backgrounds, but should also support rich image
backdrops while retaining readable foreground content and animated-background
continuity. The background example should make both modes easy to compare.

## Scope

- Add an app background image mode using the current `ubunatic.com/cati`
  TUI-image renderer Go library.
- Render the full-resolution foreground image characters using the same cell
  painting approach as Astra foreground characters.
- Dim the image during loading and while it is being prepared for display.
- For cells where animated background effects cannot paint foreground image
  characters, compute and use the average color of the corresponding dimmed
  image region so the image remains visually continuous beneath app content.
- Extend `examples/background` with a switch between the Astra theme and the
  supplied cat image theme.

Resolve the renderer version/API and the exact image-to-cell sampling behavior
against the live dependency and current Astra implementation before coding.

## Goal

/goal: Let Loom applications load and display a dimmed full-resolution image as
their background, preserve best-possible image continuity beneath foreground
and animated cells, and let `examples/background` switch reliably between Astra
and image themes.

## Verification

Add focused coverage for image loading/dimming, cell rendering and sampling,
average-color fallback, loading/error behavior, and theme switching. Verify the
background example with both themes and representative foreground content.
Extend the PTY tests to launch the background example, switch themes, and prove
the image-backed mode renders correctly through a real terminal session.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| audit | jules-audit-090 | 9a08c30d-a08b-4ac1-8163-ec20e27de642 | GH #66 | discarded: refs ~2/5; host check, note below |

Audit note (host check of Jules lead, 2026-11-03 at 2179487): partly done. `ImageBackground` exists (`background.go:131`, Braille sampling via `c.PaintDecoration`) and does not use `ubunatic.com/cati`. Left: average-colour fill under animated cells (no averaging in `background*.go`), and the Astra/cat switch in `examples/background` (it sets only `loom.NewAstraBackground()` at `examples/background/background/background.go:128`; `assets/cati.png` is already present).
