# 450 — Half-block scrollbar thumbs on a themed track background

**Status**: Closed — half-block thumbs delivered for View and Viewport; TextEntry reuses them (448)
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 448, 449

---

## 1. Problem & Motivation
The scrollbar thumb moves in whole cells, so a bar with little travel (content a few lines longer than the view) shows identical pictures for different offsets (content 8 / 6 visible: offsets 1 and 2 look the same). The track and thumb are drawn as `░` and `▓` characters, which cannot carry a half-cell edge and depend on the terminal's shade glyphs.

## 2. Technical Specification / Findings
Design agreed in chat (decide mode filled the open choices):
- **Track is always a background color.** Use the theme's `ScrollbarTrackBG` and `ScrollbarThumbBG` (`theme.go`, `ScrollbarStyle()`); the spec chars `scrollbar.foreground_char`/`background_char` stay only as the no-color fallback (`Canvas.ColorProfile` without colors): whole-cell `░`/`▓`, rounded to cells as today.
- **Units.** A track of N rows has `2N` half-row units (the unit count per cell is one constant so eighths can follow). Thumb length `L = round(units * visible / content)`, minimum one cell (2 units); start `s = round(offset * (units - L) / maxOffset)`. Offset 0 gives `s = 0`, `maxOffset` gives `s = units - L`.
- **Cell rules.** No thumb: `" "` with the track BG. Thumb fills the cell: `" "` with the thumb BG. Thumb starts inside the cell (lower part covered): `▄`, fg thumb, bg track. Thumb ends inside the cell (upper part covered): `▄` inverted, fg track, bg thumb. Only lower blocks are used so eighths (`▁..▇`) need no new glyph class.
- **Minimum one cell** guarantees no cell holds a middle-only slice.
- **Hit testing.** A cell is "on the thumb" if any part is covered. The pointer reports whole rows, so a drag moves the thumb in 2-unit steps and keeps the grabbed half-row phase (grab offset measured in units: `2*row - s` at the press). Wheel, keys and Esc-restore reach every half-row position.
- Share one renderer and one set of unit helpers in `scrollbar.go` for `View` (`view.go`), `Viewport` (`viewport.go`) and later the TextEntry bar (448). Keep 449's capture and the Viewport press/drag behavior (grab offset, track click centers the thumb) working in units.
- Eighths, and a theme tuning of the default colors, are follow-ups.

## 3. Implementation & Verification Plan
- Replace `scrollbarThumbLength/Start/Offset` with unit-based helpers (tests first), add `scrollbarCells(track, thumbStart, thumbLen)` returning per-row glyph and style kind.
- Tests: the three examples from the design (content 12/6: seven distinct pictures; 8/6: three distinct; 60/6: one-cell thumb that straddles as `▄` over `▀`); per-cell FG/BG assertions with `LOOMCOLOR=truecolor`; no-color fallback unchanged; drag keeps phase; offsets 0 and max reach both ends.
- Docs: `docs/Widgets.md` scrollbar section, spec comment for the fallback chars; check with `loom widgets --show Viewport` and `loom eval`/`loom measure` if an ANSI design changes.
