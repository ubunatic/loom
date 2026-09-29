# 167 — Add ProgressBar widget with determinate, indeterminate, and custom glyph styling

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: graph/bar.go, graph/spinner.go, widget.go

---

## 1. Problem & Motivation
Loom provides `graph.Bar` and `Gauge` for metric visualization, as well as `graph.Spinner` for ongoing background activity. However, applications frequently require a standard, standalone `ProgressBar` widget for long-running operations (file downloads, task completion, progress dialogs, batch jobs).
Currently, consumers must hand-roll progress displays or configure low-level graph bars with custom labels, lacking:
- Clean percentage / fraction formatting (`[████████░░░░░░░░] 50% (12/24)`).
- Indeterminate / marquee mode (bouncing pulse or striped animation when total work is unknown).
- Standard themeable fill / empty / head glyphs.
- Automatic integration with `Pane.Invalidate()` or `Ticker` for animated indeterminate states.

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
