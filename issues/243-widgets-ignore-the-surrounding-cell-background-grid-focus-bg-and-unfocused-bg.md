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

## 4. Milestones
- Plan (dev-243, agy:flash38): root causes confirmed. Grid `grid.go:136` uses `c.Fill` (claims cells);
  Choice, Table, Tree, KeyHelp, Viewport `PaintSurface` the theme NormalBG over their whole rect;
  ProgressBar sets an explicit BG on every cell. The first write turn stalled 45 min without edits and was
  stopped by the host.
- M1 Pre-Work / Required Refinements (host review, user approved):
  - Grid paints the focused cell with `PaintSurface`, not `Fill` (also fixes 244).
  - Rejected: a canvas rule that skips any `PaintSurface` over an existing surface. It would also drop
    intentional surfaces (Dialog, Popup inside a colored pane).
  - Instead add a weak "default surface" (additive API, e.g. `PaintDefaultSurface` or a Cell flag):
    a parent's surface replaces it at merge; standalone it shows the theme NormalBG (keeps 222).
    Explicit surfaces stay authoritative. The six widgets paint their theme normal BG as default
    surface; ProgressBar leaves BG unset where it would only repeat the normal/track BG, keeping 239's
    fill vs selection distinction.
  - Tests: each widget in a focused and an unfocused Grid cell; standalone theme BG (222 tests unchanged);
    Dialog/Popup surface still wins inside a colored parent; astra decoration in the focused cell (244).
  - Document the default-surface rule in docs/AnimatedBackgrounds.md (compositor contract) and docs/Widgets.md.
- M1 delivered (dev-243-sonnet, agy:sonnet): weak default surface `PaintDefaultSurface` and Grid focused
  cell as `PaintSurface` (c3b3f76); the six widgets switched to it (2c39017). The agent stopped on the agy
  quota and left uncommitted work in `canvas.go`, `grid_surface_test.go` and `docs/AnimatedBackgrounds.md`
  (narrower skip rule; the astra test checks only Braille glyphs). Targeted tests green; full `make test-q1`
  not yet run.
- M2 Pre-Work / Required Refinements (host review + user gallery check, All tab):
  - Review and commit or rework the uncommitted M1 leftovers first.
  - Choice, Table: user-confirmed OK. Table's full-row selection is ticket 240, not this one.
  - ProgressBar still wrong: the bracket, track, fill glyphs and the `90%` label keep a darker explicit BG
    instead of the Grid cell's BG. Only a fill color may set BG; every other cell must leave BG unset so it
    inherits the surface. Add a test: ProgressBar in an unfocused and a focused Grid cell, every cell
    outside the filled part has the cell's BG; 239's fill vs selection colors unchanged.
  - Check Tree, KeyHelp, Viewport the same way in the gallery before closing.
