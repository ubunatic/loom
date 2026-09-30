# 192 — Add an axes-based Chart widget for line and grouped bar series

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: graph/, widget_graph.go, issues/179, issues/180

---

## 1. Problem & Motivation
Loom has Gauge, ProgressBar, Sparkline, and graph bar renderers, but no chart with axes and multiple series. Ratatui ships Chart and BarChart (issue 179).

## 2. Technical Specification / Findings
`Chart` widget: X/Y axes with tick labels, multiple series, line mode (braille or half-block) and grouped bar mode, legend, series colors from theme. Reuse graph/ renderers where they fit. Needs human review of the look (axis labels, glyph choice).

## 3. Implementation & Verification Plan
/goal Ship `Chart` (line and grouped bars) with scaling and render tests, a docs row, and an example for human review; stop and report when the glyph or axis layout needs a user decision.
