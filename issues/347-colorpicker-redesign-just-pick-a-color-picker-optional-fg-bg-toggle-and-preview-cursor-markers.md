# 347 — ColorPicker redesign: just-pick-a-color picker, optional FG|BG toggle and preview, cursor markers

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Design
**Related**: 346 (ColorPicker widget, 🎨 popover item; commits 61e9c25, cd39f81)

---

/goal Redesign the ColorPicker into a "just pick a color" picker per the user feedback below: first
propose `.ansi` designs and get the user's choice (M1), then implement the chosen design (M2); spec,
test, document; run `make test-q1` and `make install`. Stop and report when blocked on the user's
design feedback, a user decision, or a denied permission.

## User feedback on the 346 picker (2026-10-21)

1. Right-click must not pick BG: with the BG target active, left- and right-click do the same thing.
2. The `FG │ BG` row above the swatches feels odd: it is unclear where the cursor is.
3. Default should be a plain color picker without the `FG │ BG` row.
4. Idea: swatches show the effect of the pick using current colors: a FG pick shows each color as
   FG on the current BG; a BG pick shows the current FG on each color.
5. `FG │ BG` toggle: optional, off by default (spec).
6. Preview: optional, on by default (spec).
7. Cursor markers around the focused swatch, e.g. `>■<█` (first focused) or `■>█<` (second focused).

## M1 — Design proposals (no code)

Write 3–4 alternatives as `.ansi` files under `docs/data/` (e.g. `colorpicker-redesign-a.ansi`),
each showing FG-pick and BG-pick state, cursor markers, with and without preview. Validate each with
`loom eval` and `loom measure` (plus `loom check-box` if bordered). Present them to the user and record
the chosen design and decisions here. Do not start M2 before that.

Open questions for the user in M1:
- What does right-click do instead: nothing, or another action?
- With the toggle off, how is a BG pick started? Candidates: Shift+click/Shift+Space and `b` stay
  (from 346); or the 🎨 bar offers two entry points; or a key flips the swatch rendering (item 4).
- Do markers take extra columns (layout width) or replace swatch padding?

## M2 — Implement chosen design

Keep the 346 widget API where possible; spec keys `color_picker.show_target_toggle: false`,
`color_picker.show_preview: true`, marker glyphs; remove the right-click → BG mapping; update tests,
gallery, docs/Widgets.md, F1 help, `docs/progress/ColorPicker.ansi`. Mouse/event change: start on
`codex:sol:med`. Live-check the installed `loom edit`.
