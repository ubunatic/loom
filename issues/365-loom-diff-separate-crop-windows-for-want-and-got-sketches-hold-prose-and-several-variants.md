# 365 — Check widgets against design references: loom crop, a loomtest compare helper, live crop for loom diff

**Status**: Closed — loom crop + loom fmt (encoding-only, idempotent, screen unchanged), loomtest.CompareWidget with -update, ColorPicker split into three testdata assets verified by user; 25b0180, 224c13d, 56920d8; host q1 green
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 362 (`loom diff`), 351 (live-check script), 359–361 (ColorPicker look bugs)

---

/goal A widget's look is checked cell by cell against a reference `.ansi` file in normal Go tests, the
reference can be cut from a design sketch and deliberately updated when the design changes, and a live
capture can be compared against the same reference; shown for the ColorPicker; or stop and report when
blocked on a user decision.

## Flow (agreed with user, 2026-10-22)

1. **Sketch:** free-form design doc (prose, variants), e.g. docs/data/colorpicker-ideas-001.ansi. It stays
   as history and is not the final truth.
2. **Reference:** `loom crop <file> --region row,col,width,height` cuts the chosen variant into its own
   file (e.g. `docs/data/ColorPicker-b2.ansi`). This is the first reference.
3. **Test:** a `loomtest` helper renders the widget alone at a fixed size and state and compares it with
   the reference cell by cell (same report format as `loom diff`). A `-update` flag rewrites the reference
   from the current render.
4. **Deliberate changes:** when the design changes on purpose (e.g. 360: picker BG changed to the Ctrl+Space
   bar colour), update the reference with `-update` in the same commit and say why in the ticket. The
   reference diff in that commit is what the reviewer checks. Unintended changes fail the test.
5. **Live (when needed):** `loom diff --got-region …` or `--find <text>` locates the widget in a tmux
   capture and compares that part with the reference.

## Revised flow (user, 2026-10-22; replaces the flow above where they differ)

1. **Sketch** in `docs/data/` (comments, variants allowed).
2. **Extract:** `loom crop` cuts one widget state; `loom fmt` rewrites it in loom's own encoding
   (SGR order, redundant resets, equivalent colour numbers, trailing spaces). `fmt` never changes what is
   shown; design fixes are hand edits.
3. **Verify:** the user checks the asset with `cat`; `loom eval/measure/check-box` pass.
4. **Use:** the asset lives in `testdata/` next to the test, one file per state
   (e.g. `testdata/ColorPicker-fg-active.ansi`); no `go embed`. `CompareWidget` checks it; the test sets
   the theme explicitly.
5. **Change on purpose:** `-update` with a written reason; the user checks the asset again.

### M2 pre-work (after 25b0180, 224c13d)

- Add `loom fmt` (encoding-only, idempotent; a test proves `cat` output is unchanged).
- Move the ColorPicker reference to `testdata/`, split into states: FG row active, BG row active, BG pick
  inside the preview rug. Keep `docs/data/ColorPicker-b2.ansi` only if it was cut from the sketch.
- Docs (CLI.md, Testing.md) describe the revised flow.

## Notes

- `--ignore-colour` covers sketches whose colours are not final yet (layout-only check).
- Out of scope: invisible part markers in sketches and part names emitted by the running app.
