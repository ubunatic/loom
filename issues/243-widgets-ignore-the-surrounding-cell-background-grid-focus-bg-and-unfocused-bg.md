# 243 — Widgets ignore the surrounding cell background (Grid focus BG and unfocused BG)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 222 (KeyHelp/Viewport theme BG), 239 (ProgressBar colors), 244 (astra in focused cell)

---

/goal Every widget in the All tab shows the background of its Grid cell (focused or unfocused) wherever it
does not draw an explicit own color, fixed once in the library with tests; or stop and report when blocked
on a user decision.

## 1. Problem & Motivation
User check of `loom widgets` → All. Correct: "Click Me" (button) takes the cell's background, the focused
cell BG when the cell is selected and the unfocused BG otherwise. Also correct: Toggle, Checkbox,
NumberInput, Badge, Pill, Sparkline, Spinner, Timer, StopWatch, Paginator, TextInput, Calendar, MenuBar,
Chart, Dialog, Popup, TextArea.

Wrong (they paint the unselected/default BG over the cell):
- ProgressBar: bar, decoration and label.
- Choice: non-selected items and the whitespace.
- KeyHelp: entirely.
- Table: nearly entirely; only the header row continuation shows the focused cell BG.
- Tree: non-selected items; the remainder is correct.
- Viewport: entirely.

## 2. Technical Specification / Findings
- `grid.go:136` fills the focused cell with `c.Fill(cr, Cell{... BG: g.FocusBG})`, a foreground write,
  before the child draws. Widgets that set an explicit theme BG on every cell overwrite it.
- The compositor already has the right model (docs/AnimatedBackgrounds.md, "Loom compositor contract"):
  `PaintSurface` sets a background without claiming cells, and cells with no explicit BG inherit it.
  Likely fix: Grid paints the cell BG as a surface, and the listed widgets leave their default-BG cells
  without an explicit BG (or use a surface-aware style) instead of painting the theme's normal BG.
- Fix at library level for all widgets; no per-widget or gallery workarounds. Keep ticket 222's intent
  (themed widgets still show the theme BG when standalone).

## 3. Implementation & Verification Plan
- Developer: `agy:flash38` (user choice), not luna.
- Test: render each listed widget in a Grid cell, focused and unfocused, and assert the cell BG on its
  non-highlight cells; a standalone render keeps the theme BG.
- `make test-q1`, `make install`, PTY check in julia256 and plain.
