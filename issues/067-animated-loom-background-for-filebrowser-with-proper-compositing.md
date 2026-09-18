# 067 — Animated Loom background for filebrowser with proper compositing

**Status**: Open
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

**Status: In Progress.** The filebrowser now opts into the Astra background, and
foreground blank cells with default backgrounds are transparent to compositing;
explicitly colored surfaces remain protected. Headless coverage verifies the
integration and transparent/opaque surface split. Theme and PTY smoke validation
remain manual follow-up work.

### M4 — Documentation, Example Alignment & Final Test Pass
- Update `docs/AnimatedBackgrounds.md` if compositing APIs or interfaces were refined.
- Ensure standalone `examples/background` and `examples/filebrowser` consistently use the unified compositing model.
- Execute single-test boundary verification (`make test-q1` or `harnez exec --quota-1 -- go test ./...`).
