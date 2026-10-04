# 264 — RichTextEdit popover keyboard navigation, box selection and box style submenu

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: `richtextedit.go`, `boxglyph.go`, `BoxBorderStyleSharp/Rounded` (existing border styles), issues 255 (popover, color picker), 257 (C-space), 261 (box drawing), 263 (Home/End)

---

/goal Make the RichTextEdit format popover fully keyboard-driven and box-aware per the milestones below, each with tests and a live gallery check; stop and report when blocked on a user decision.

Re-check live code and recent commits first. Stop for host review after each milestone.

## M1 — C-space on box art selects the box

When the cursor is on a closed box perimeter, `C-space` selects its rectangle and opens the popover. Use reciprocal glyph arms to validate the perimeter under Decisions; off a valid perimeter, keep word selection.

Acceptance: sharp/rounded boxes with styled text; crossing strokes; nested/adjacent/shared-edge boxes and deterministic ties; broken perimeter fallback; wide/combining text and side text outside the rectangle.

**M1 delivered (2026-10-04, 7affb03):** C-space on a closed sharp or rounded box selects the box rectangle (per-row slices) and opens the popover; delete/copy/style/undo act on the rectangle; cursor moves clear it. Host rerun: green except the known 247 flake (passes 5/5 alone).

### M2 Pre-Work / Required Refinements (user live check of M1, 2026-10-04)

M1 confirmed working live. Do these first, each with tests:

1. Esc must not quit the widgets gallery app; only C-q and F10 quit. Esc stays available to widgets (e.g. closing the popover).
2. The widgets gallery app runs full terminal width (no narrower MaxCols cap by default; explicit --width still wins).
3. Box mode paints with the color of the starting glyph: when box drawing starts on an existing box glyph, every glyph drawn or changed along the stroke takes that glyph's foreground (grey start -> grey track, white start -> white track). Starting on non-box text keeps current behavior. Undo stays one step per stroke.
4. Remove the C-S-b box-mode binding (keep F5); add a popover button that starts box drawing mode. Update docs/Widgets.md and docs/TuiInput.md.

## M2 — Tab/Enter in the popover bar

With the popover open, `Tab` / `S-Tab` cycle the bar items with a visible focus highlight, `Enter` applies the focused item, `Esc` closes. Initial focus follows the selection: bold → "B", italic → "I", underline → "U", box → "Box", else the first item.

Acceptance: forward/reverse wraparound and highlight; box focus takes precedence, otherwise first uniformly active B/I/U in that order; mixed/plain selection starts at B; Enter/Esc leave document and host focus unchanged except the requested formatting.

**M2 delivered (2026-10-04, 6a4faef pre-work, e9350a0 bar navigation):** Esc no longer quits the gallery (C-q/F10 only, verified live), box strokes take start glyph color, C-S-b removed, popover box-mode button, Tab/S-Tab/Enter/Space/Esc in the bar. Host rerun green.

### M3 Pre-Work / Required Refinements (host review + user, 2026-10-04)

1. Full width not met: in a 160-col tmux the RichTextEdit demo still draws at most ~70 cols (no content or border beyond col 70). The editor/demo must fill the terminal width. Add a test at a wide size.
2. Stale demo hint: the gallery line still says "F5/C-S-B Box"; drop C-S-B and mention the popover box button.
3. User: C-space on a space character (blank cell, not on a word or box) must still open the popover, with the box-mode button usable, so box drawing can start from empty space.
4. User: cursor Up/Down may move into the "void" (columns past the line end, rows past the last line) without inserting spaces; the virtual position is kept and drawn. Spaces/lines are padded only when the user types or draws there. Moving away without typing leaves the document unchanged. Test both.

5. User: cursor shape follows the mode: bar cursor for normal typing, block cursor while box drawing mode is on. Implement in the library (widget reports desired cursor shape, pane emits DECSCUSR `ESC[n q` only on change and restores the terminal default on exit), not per demo. Test the emitted sequences.

## M3 — Box style dropdown

"Box" opens a dropdown with **plain** (sharp, `┌┐└┘`) and **rounded** (`╭╮╰╯`). Choosing a style wraps the selection, or restyles the selected box (M1). Reuse the existing `BoxBorderStyle*` glyph sets; labels/defaults go in the spec per `docs/Spec.md`.

Acceptance: both styles wrap single/multiple lines and partial-line selections; restyle in place without changing dimensions, interior spans, outside text or crossing arms; one undo/redo step; rounded corner arm masks and sharp T/cross junctions; spec/schema/loader agree.

## M4 — Keyboard navigation in submenus

Flow: `C-space` → bar opens → `Tab`… → `Space` opens the focused item's submenu (FG/BG color picker or box style) → `Tab`/arrows… → `Enter` applies → back to text; `Esc` steps back one level.

Acceptance: complete keyboard flow for FG, BG and Box; reverse cycling and arrow wraparound; submenu Esc restores parent focus, next Esc returns to text; Tab/S-Tab/Space/Enter/Esc all return consumed with Done/Quit false, including under Frame/Split; no inserted whitespace, host traversal or box strokes; mouse and ViewMode regressions.

## Decisions

- **Bounds (M1):** Work in logical document rows and terminal display columns, independent of scroll/viewport. Find closed axis-aligned rectangles whose perimeter contains the cursor glyph: consecutive perimeter cells must have reciprocal arms, corners must contain the required turning arms. Extra arms at `├┤┬┴┼` are allowed; validate the rectangle rather than taking the entire connected component's bounds. Interior text/lines are unrestricted. Choose smallest area, then top row, left column, bottom row, right column; this selects the inner box on its border and resolves shared-edge/adjacent ties. A cursor in interior text selects the word; an open stroke/broken box also falls back to word selection. Heavy/double/ASCII boxes remain out of scope.
- **Rectangle selection:** Retain rectangle coordinates and per-row rune slices for highlighting and selection operations; a single `RichPosition` range would wrongly include side text on intermediate rows. Include the perimeter and interior, exclude outside prefixes/suffixes. Reuse `richLineColumn`, `richLineOffsetAtColumn`, `richLineCell` and cluster widths; wide-rune continuation cells are not independent glyph vertices, and slice boundaries must not split clusters or pills. Revalidate/clear the box target after edits, undo/redo or selection movement. Restyling changes only perimeter glyphs, preserving spans, interior, dimensions and all extra junction arms; do not wrap an already detected box again.
- **Rounded arms (M3):** Extend `boxglyph.go:BoxGlyphArms` so `╭ = Right|Down`, `╮ = Left|Down`, `╰ = Right|Up`, `╯ = Left|Up`, using the existing rounded glyph set from `getBoxBorderGlyphs(BoxBorderStyleRounded)` (`frame.go`, `spec/box.yaml`). In `richtextedit.go:drawBoxStep`/`boxNeighbourArms`, preserve an existing rounded corner if its two-arm mask is unchanged; adding a third/fourth arm produces the existing sharp `BoxGlyph(mask)` T/cross (`├┤┬┴┼`). New stroke corners stay sharp. Restyling to rounded changes only two-arm turning corners; junctions stay sharp. Plain maps to Sharp; default wrapping is plain. Add menu labels/default and focus styling to `spec/defaults.yaml`, its schema and loader; no duplicate Go glyph/label tables.
- **Routing (M2/M4):** Add explicit bar/submenu focus and dismissal state in `richtextedit.go:ConsumeKey`, after the ViewMode guard and before BoxMode or the normal editing switch. Share action dispatch with `applyPopoverAction`, drawing with `drawPopover`/`drawPopoverPalette`, and mouse state with `ConsumeMouse`; do not derive keyboard focus from draw-generated hit rectangles. Bar Tab/S-Tab cycle enabled actions; Enter/Space toggle B/I/U/S or open FG/BG/Box, with applied actions dismissing to text. Link is currently unimplemented: show disabled and skip it. Submenu Tab/S-Tab and Right/Down or Left/Up cycle choices with wraparound; Enter/Space apply and dismiss both levels. Esc closes only the submenu first (retaining bar focus/selection), then dismisses the bar without changing text. Every listed key returns `Handled()` (`EventResult{Consumed:true, Done:false, Quit:false}`), even at boundaries; never `DoneResult()` for formatting. `Frame.ConsumeKey` and `Split.ConsumeKey` already dispatch child first and stop on Consumed, preventing host focus cycling. Other printable typing dismisses the menu and follows normal selection replacement; dismissal must suppress automatic reopening while the selection remains. Clear menu state on focus loss/ViewMode; if geometry cannot show a focused menu, dismiss it and consume that triggering menu key rather than retaining invisible focus.

## Risks

- Rectangular selection must cover existing styling/copy/delete consumers, not just drawing; preserve pill atomicity and cluster widths. Bound candidate search to document glyphs and avoid exponential path enumeration on dense crossing art.
- `applyPopoverColor` currently bypasses `mutate`; make color application one undo step too. `ShowPopover` is a configuration preference, not sufficient dismissal/focus state; clipping and mouse/keyboard transitions need explicit state.
- Issue 263 concurrently changes `ConsumeKey` navigation: re-read live code before implementation, keep its changes, and start event-routing work on `codex:sol:med`. Retain each milestone's host review, live gallery check and host `make test-q1` rerun.
