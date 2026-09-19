---
title: Compositor Layering and Root Modal Overlays
---
# Compositor Layering and Root Modal Overlays

## Header & Context

Date: 2026-09-19. Spans issues [067](../../issues/067-explicit-compositing-contract-for-widgets-over-animated-backgrounds.md),
[068](../../issues/068-migrate-widgets-to-the-explicit-layered-compositor-api.md),
[066](../../issues/066-nested-help-pane-opens-a-second-pane-racing-the-outer-pane-tty-reader.md),
[069](../../issues/069-render-help-as-a-root-level-modal-overlay.md), and files
[070](../../issues/070-prevent-text-overflow-in-framework-help-modals.md) as a
follow-up. All four issues are closed; 070 remains open.

## Executive Summary

A multi-session arc that (1) gave Loom an explicit four-layer compositor API
(`PaintSurface`/`PaintForeground`/`PaintDecoration`, `Cell.Claim`/`Cell.Surface`)
replacing an inference heuristic for background decoration eligibility, (2)
found and fixed two independent, symptom-identical bugs that made the Astra
star-field decoration invisible on themed panes, (3) tuned Astra's color model
to scale with the actual surface color instead of assuming a dark terminal,
and (4) fixed a tty-ownership race in `:help` and then its follow-on scoping
gap by introducing a reusable root-overlay hook pattern. See
[RootOverlays.md](../RootOverlays.md) and [AnimatedBackgrounds.md](../AnimatedBackgrounds.md)
for the resulting design docs.

## What Changed

- `canvas.go`: `Cell.Claim`, `Cell.Surface` fields; `Set`/`PaintSurface`/
  `PaintForeground`/`PaintDecoration` replace the removed `transparentForeground()`
  heuristic. Fixed a surface-color/flag propagation-order bug across nested
  `paintClipped` merges (two-level regression test added).
- `choice.go`, `table.go`, `view.go`, `notif.go`, `cmd.go`, `settings.go`,
  `tabs.go`, `textarea.go`, `textinput.go`, `confirm.go`: raw `Fill` background
  rows migrated to `PaintSurface` so themed panes don't opaquely claim cells
  decoration should still reach. `popup.go`/`grid.go` deliberately left as
  raw `Fill` (they are meant to occlude decoration).
- `style.go`: `Color.RGB()` + `xterm256RGB()` — resolves any `Color` (RGB or
  indexed) to approximate 24-bit RGB, reused by Astra's color math.
- `background.go`: `AstraBackground` dim phase now interpolates from the
  actual surface color (not a fixed gray); bright phase now uses
  `peakChannel(bg, floor, margin)` per channel so peak brightness scales
  with the surface instead of only flooring above it.
- `pane.go`, `cmd.go`: `paneHelpRequest` root-overlay hook, `Pane.help`,
  single-pane `paneOwnership` guard. See [RootOverlays.md](../RootOverlays.md).

## Process Notes (what worked, what didn't)

**Two bugs, one symptom.** "Stars aren't visible on themed panes" had two
independent root causes — a compositor merge-order bug and several widgets
bypassing the new `PaintSurface` API with raw `Fill`. Fixing the first alone
looked plausible in isolation but the user's re-test ("still same as before")
caught that it wasn't the whole story. Lesson: a single plausible root cause
matching the reported symptom is not confirmation — re-test after each fix
before considering an issue closed, especially when two code paths could
independently produce the same visible failure.

**Manual PTY confirmation beat scripted PTY confirmation.** A custom
`pty.openpty()`-based harness for the filebrowser's `:help` flow was
inconclusive (synthetic keystrokes appeared to land in the wrong context in
captured output), so the issue was correctly left "in progress" rather than
falsely marked done. The user's own direct terminal run (`:h<CR>` → help
shown → `<CR>` → dismissed) then confirmed the fix in seconds. Lesson: for
interactive-state UI verification (show/dismiss cycles, focus behavior),
prefer a short manual check over investing further in synthetic PTY-keystroke
timing, unless the PTY harness is purpose-built for that exact protocol (as
`scripts/check-watch-pty.py` is for `monitor --watch`).

**A ticket was closed before its code was committed.** The dev agent's
closing commit for issue 069 (`a242b23`) landed before `cmd.go`/`pane.go`/
`pane_internal_test.go` were committed — caught only by `git status --short`
showing those files still modified despite the "Closed" ticket. The
implementation was committed separately (`e4eefc1`) with a note explaining the
gap. Lesson: when reviewing a dev-agent report claiming a ticket is done,
check `git log -- <files-the-ticket-should-have-touched>` against the ticket's
own closing commit, not just that *a* closing commit exists.

## Follow-up

Issue 070 (open): the shared popup/help text layout does not wrap or clip
long lines at narrow canvas widths, so filebrowser help text can visibly cross
the modal border. Root-overlay ownership and input-capture semantics from 069
are unaffected — this is scoped as a text-layout fix only.
