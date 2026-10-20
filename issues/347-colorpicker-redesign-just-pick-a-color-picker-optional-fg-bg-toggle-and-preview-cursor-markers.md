# 347 — ColorPicker redesign: just-pick-a-color picker without FG|BG switcher, optional preview, cursor markers

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

Decisions (user, 2026-10-21, supersede items 1 and 5 above):
- `FG │ BG` switcher: remove it entirely (not an optional setting). Drop its spec keys, code,
  `x` binding and tests.
- Right-click depends on the design: include at least one design where right-click is the BG
  action. Which picker design is used depends on whether mousegrab is on (342 default on; 303
  temporary grab when off): propose which design applies in each mode, and verify in a live
  terminal that right-click reaches loom in both modes before relying on it.
- Markers take no extra columns: they reuse existing swatch spacing; layout width is unchanged.

Chosen design (user, 2026-10-21, after gapless proposals f114015): one picker for both mousegrab
modes. One row of 16 contiguous 2-cell `Aa` swatches (32 columns, no spacers).
- FG mode (default on open): swatches show each color as FG on the current BG (design B FG row);
  the focused swatch is marked by underlining its text. This underline also signals FG mode.
- BG mode: swatches show the current FG on each color (design B BG row); the focused swatch is
  marked by a line on the row below the cell (`▔▔`). This marker is hidden in FG mode.
- Mode switch (no color change): Down, Tab, `b`, and right-click. Down acts like Tab here; Tab is
  this widget's documented action (document it under the Tab policy in docs/Widgets.md).
- Click or Space picks the focused/clicked color for the current mode.
- Up switches back to FG mode (counterpart of Down).
- Dropped: two-row design A, `●` cursor, right-click as BG pick, and the 346 shortcuts
  Shift+click, Shift+Space and `f` (user "go" on the recommendation, 2026-10-21).

## M2 — Implement chosen design

Keep the 346 widget API where possible minus the target switcher; spec keys
`color_picker.show_preview: true`, marker glyphs, and the mousegrab-dependent design choice;
right-click per the chosen design; update tests,
gallery, docs/Widgets.md, F1 help, `docs/progress/ColorPicker.ansi`. Mouse/event change: start on
`codex:sol:med`. Live-check the installed `loom edit`.
