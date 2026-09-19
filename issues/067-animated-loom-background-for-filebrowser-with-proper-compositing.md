# 067 — Animated Loom background for filebrowser with proper compositing

**Status**: Closed — compositing contract delivered via examples/background; filebrowser wiring folded into 068 M2
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: `background.go`, `pane.go`, `canvas.go`, `frame.go`, `examples/filebrowser/filebrowser/filebrowser.go`, `docs/AnimatedBackgrounds.md`, `b7a831c` (`feat: add animated background support`), [063](063-convert-filebrowser-example-to-a-hostable-widget.md)

---

## 1. Problem & Motivation

Loom provides an animated Astra/Braille star-field background (`AstraBackground` in `background.go`), but the filebrowser example cannot use it as a visible application background because foreground layout containers (`Frame` and `Box` in `frame.go`) and widgets fill their bounding rectangles with default spaces/styles, completely overwriting the background drawn in the pre-render pass.

Currently:
1. `Pane.run` renders `Pane.Background` onto the canvas *before* invoking `root.Draw`.
2. `Frame.Draw`, `Box.Draw`, `Choice.Draw`, and `View.Draw` fill their rectangular bounds with space cells (`Cell{Text: " ", Style: ...}`), erasing all background glyphs except in unused canvas padding outside the widget tree.
3. If background is rendered after `root.Draw` without awareness of cell claims, background glyphs overwrite foreground text, borders, selections, prompts, scrollbars, and the active terminal cursor.

The filebrowser should showcase a real Loom animated background shining through empty/unclaimed regions (such as empty list rows, blank detail areas, and split-pane gaps) while keeping interactive content, borders, and cursor coordinates completely protected. This requires a formal compositing layer in Loom as specified in `docs/AnimatedBackgrounds.md`.

## 2. Technical Specification & Findings

### 2.1 Post-Render Compositing vs Safe Cell Classification
In accordance with `docs/AnimatedBackgrounds.md`, Loom should treat background effects as a renderer-layer compositing pass:
- **Foreground pass first**: `root.Draw(canvas, bounds)` executes first to establish all content, styled surfaces, and cursor position (`canvas.CursorX`, `canvas.CursorY`).
- **Cell eligibility rules**: A cell at `(x, y)` is safe/eligible for background rendering if and only if:
  1. It is not occupied by printable text (`cell.Text == " "` or `cell.Text == ""`).
  2. It has no opaque/explicit custom styling (e.g. `cell.Style.BG.IsDefault()` or transparent background).
  3. It does not coincide with the active terminal cursor `(x == canvas.CursorX && y == canvas.CursorY)`.
  4. It is not part of an active selection or protected overlay.
- **Cursor protection**: `Canvas` cursor coordinates must remain untouched during background compositing.

### 2.2 Background Contract & Scheduling
- `AnimatedBackground` interface in `pane.go`:
  ```go
  type Background interface {
      DrawBackground(*Canvas, Rect)
  }

  type AnimatedBackground interface {
      Background
      DrawBackgroundAt(*Canvas, Rect, time.Time)
  }
  ```
- **Independent cadence & lifecycle**: Background effects should be able to declare or use specific frame intervals without binding all effects to the global `SpeccedBackground.RedrawInterval`.
- **Teardown & Timer cleanup**: Ensure tickers and redraw channels are cleaned up cleanly on `Pane.Close` or widget teardown with zero leaked goroutines.
- **Accessibility / Reduced-Motion**: Respect user preferences or configuration to disable animated background redraws without breaking static background styling or widget layout.

### 2.3 Filebrowser Integration & Theming
- Enable `AstraBackground` in `examples/filebrowser/filebrowser/filebrowser.go` via `pane.Background = loom.NewAstraBackground()`.
- Verify contrast and aesthetic balance with all supported themes (`mc`, `default`, `julia256`, `plain`) so stars are subtly visible in empty split-pane areas without clashing with file list items or metadata text.
- Ensure file filtering (`q`), navigation (`↑`/`↓`), theme cycling (`F9`), pane resizing, and mouse selection remain completely responsive and unaffected by the background tick loop.

---

## 3. Implementation & Verification Plan

### M1 — Compositing Contract & Safe-Cell Canvas Primitives
- Implement safe-cell eligibility detection on `Canvas` (e.g. `Canvas.IsEligibleBackground(x, y)` or a protected-cell compositor helper `Canvas.ComposeBackground(bg, area, now)`).
- Ensure `Canvas.Set` or compositor pass respects protected cells, preserving text, non-default backgrounds, wide continuation runes, and cursor state.
- Update `Pane.redraw` in `pane.go` to execute foreground rendering first, followed by safe background compositing into unclaimed cells.

*Verification*:
- API and canvas unit tests demonstrating that background rendering writes to blank/unclaimed cells while leaving styled text, box borders, prompt text, and cursor positions intact.

### M2 — Animated Background Scheduling & Lifecycle Safety
- Ensure `Pane.run` schedules background animation ticks smoothly alongside widget events and window resizes (`SIGWINCH`).
- Verify clean teardown on exit, context cancellation, or signal interruption with no leaked background tickers.
- Add support for disabling animation (reduced-motion guardrail) while retaining static background rendering.

**Status: Complete.** `Pane.run` owns the animated-background ticker, derives its
cadence from `BackgroundCadence` when available, stops it on every exit path, and
honors `Pane.ReduceMotion` while preserving normal static redraws. Astra cadence
and quantized frame behavior are covered by deterministic tests.

*Verification*:
- Deterministic canvas tests for frame progression at quantized timestamps.
- Goroutine leak and ticker teardown tests ensuring zero zombies upon pane closure.

### M3 — Filebrowser Example Integration & Theming Polish
- Wire `AstraBackground` into `examples/filebrowser/filebrowser/filebrowser.go`.
- Validate background star visibility and contrast across all themes (`mc`, `julia256`, `plain`, `default`).
- Verify split-pane resizing, file filtering, directory navigation, and mouse click hit-testing.

*Verification*:
- Automated headless/golden canvas tests for `filebrowser` confirming stars render in blank areas of the file list and metadata views.
- PTY smoke test verifying visual fidelity and mouse/keyboard interactivity during continuous animation ticks.

**Status: Complete.** Correction: an earlier review pass of this ticket
incorrectly claimed the filebrowser never opted into `AstraBackground`. It
does: `configurePane` in `examples/filebrowser/filebrowser/filebrowser.go`
sets `pane.Background = loom.NewAstraBackground()` (landed in `9d9b838`,
before this ticket was closed), and `TestBrowserUsesAnimatedBackground` in
`examples/filebrowser/filebrowser/browser_test.go` asserts both the wiring
and that the drawn frame leaves transparent surface for composition. Theme
contrast (`mc`, `default`, `julia256`, `plain`) and PTY interactivity remain
unverified by automated tests since `loom.New`/`Pane.Close` require a real
tty, which is a headless-test-environment limitation tracked under
[068](068-formalize-layered-compositor-semantics-and-background-inheritance.md)
M2, not a missing feature.

### M4 — Documentation, Example Alignment & Final Test Pass

**Status: Deferred to 068.** Unified compositing model, filebrowser
alignment, and doc refresh depend on the explicit layered-compositor API
that issue 068 introduces; tracking them here would duplicate that ticket.

## Closing Summary

- M1 (compositing contract & safe-cell primitives): Complete.
- M2 (animated background scheduling & lifecycle safety): Complete.
- M3 (filebrowser integration & theming polish): Complete; `configurePane`
  wires `AstraBackground` into the filebrowser and
  `TestBrowserUsesAnimatedBackground` covers it. Theme-contrast and PTY
  smoke validation remain manual/untested and are tracked under 068 M2.
- M4 (docs/example alignment/final test pass): Deferred to 068 M3.

Closing 067 because its core goal — a working, tested compositing contract
that lets an animated background shine through unclaimed cells while
protecting text, borders, and cursor state, demonstrated in both
`examples/background` and the filebrowser — is delivered. Remaining
polish (theme contrast validation, PTY smoke tests, unified docs) depends
on 068's explicit layered-compositor API and is tracked there.
