# 167 — Add ProgressBar widget with determinate, indeterminate, and custom glyph styling

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: graph/bar.go, graph/spinner.go, widget.go

---

## 1. Problem & Motivation
Loom provides low-level rendering helpers (`graph.Bar`, `graph.RenderBracketedBar`), a metric `Gauge`, and a specialized startup splash screen widget (`SplashView`, used in `examples/splash`).
However:
- In `SplashView` (`examples/splash`), the progress display is hard-coded into the splash screen's specific 8-row layout (centered title, spinner, pills, step text, and footer), directly calling the low-level string formatter `graph.RenderBracketedBar(...)`. It is not an independent, reusable widget that arbitrary containers (`Frame`, `Stack`, `Split`, `Modal`) can host.
- Applications frequently require a standard, standalone `ProgressBar` widget for long-running operations (file downloads, task completion, progress dialogs, batch jobs).
- Currently, consumers outside `SplashView` must hand-roll progress displays or configure low-level string formatters with custom labels, lacking:
  - Standard `loom.Widget` lifecycle (`Draw`, `Measure`).
  - Clean percentage / fraction formatting (`[████████░░░░░░░░] 50% (12/24)`).
  - Indeterminate / marquee mode (bouncing pulse or striped animation when total work is unknown).
  - Standard themeable fill / empty / head glyphs.
  - Automatic integration with `Pane.Invalidate()` or `Ticker` for animated indeterminate states.
- Once implemented, `SplashView` can also optionally adopt `ProgressBar` internally or share its configuration.

## 2. Technical Specification / Findings
Introduce `loom.ProgressBar` (implementing `loom.Widget` and `loom.Themeable`):
- **Fields**:
  - `Value float64`, `Total float64` (determinate 0..Total).
  - `Indeterminate bool` (marquee pulse mode).
  - `Width int` (fixed or fill available width).
  - `ShowPercent bool`, `ShowCount bool`, `Unit string`.
  - `FillRune rune`, `EmptyRune rune`, `HeadRune rune`.
  - `StyleFill loom.Style`, `StyleEmpty loom.Style`.
- **Indeterminate Animation**:
  - Internal tick / frame counter for moving marquee pulse.
  - Implements `loom.PaneRequester` or `Ticker` listener to advance frame smoothly.
- **Rendering**:
  - Layout bar and text label compactly according to available `image.Rectangle` bounds.
  - Safe bounds clamping (0 <= Value <= Total).

## 3. Implementation & Verification Plan
- Create `progressbar.go` and `progressbar_test.go`.
- Unit tests:
  - Value clamping at 0, Total, and beyond.
  - Custom formatting with percentage and count labels.
  - Indeterminate pulse position calculation across frames.
  - Canvas rendering output checks for filled vs empty runes and colors.
- Document in `docs/Widgets.md`.

## 4. Outcome, 2026-09-29 (milestone 1: determinate bar)
Delivered `progressbar.go` (`NewProgressBar`, `Set`, `Done`, `Reset`, `Value`, `IsDone`, `Style`, `Align`), `graph.BracketedBarStep`, and the `progress_bar` section in `spec/defaults.yaml` (width 24, done pattern `:`).
- Throttle: `Set` compares the visible step (filled half-cells) and calls the pane invalidate only when it changes; the pane's invalidate channel coalesces bursts.
- Done: one-frame swap to the done pattern, same as `SplashView` (user chose this over a drain animation).
- Tests in `progressbar_test.go`: invalidation count, clamping incl. NaN/Inf, done/reset, clipping, step-vs-render agreement.

Remaining (still open): indeterminate pulse mode, percent/count labels, separate fill/empty styles, `Themeable`, migrating `SplashView` onto `ProgressBar`.
