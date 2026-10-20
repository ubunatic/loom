# 346 — loom: ColorPicker widget for FG/BG; replace #FG/#BG popover items with one 🎨 item

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 255 (current FG/BG palette in the popover), 345 (lean popover, hamburger), 342 (Ctrl+Space toggle)

---

/goal Ship a reusable library `ColorPicker` widget that picks FG, BG or both, and use it from one
`🎨` item in the RichTextEdit popover in place of `#FG` and `#BG`; spec, test, document and show it in
the `loom widgets` gallery; run `make test-q1` and `make install`, or stop and report when blocked on
a user decision or denied permission.

## Request

- Popover bar: replace `#FG` and `#BG` (spec/defaults.yaml `popover_labels`, ~l.306) with a single
  `🎨` item. Fits the lean-bar decision from 345.
- `🎨` opens a FG/BG color picker built as a general loom widget (not RichTextEdit-private code; the
  palette from 255 is the starting point to extract or replace).
- One picker session can set one or two colors:
  - Click or Space on a swatch → pick FG.
  - Shift+click or Shift+Space on a swatch → pick BG.
  - Show the current FG/BG choice (preview) so a user can set both before closing.
- Apply to the selection (or typing style when nothing is selected), same as today's #FG/#BG.

## Constraints and open points

- Keys, labels, glyph and palette in the spec; no hardcoded values in Go (docs/Spec.md).
- Mouse/event routing change: start on `codex:sol:med`. Mouse coords are 0-based child-local.
- Uncertain: many terminals keep Shift+click for their own text selection, and SGR mouse reports
  carry the Shift bit only when the terminal forwards it. Probe first (docs/Canary.md) in the user's
  terminals; Shift+Space then needs `shift-space` decoding (kitty CSI-u only). Record results; if
  either is unavailable, ask the user before adding a fallback (e.g. a FG/BG mode toggle).
- How the picker closes (Enter/Esc/click outside, Ctrl+Space per 342) and whether it allows "no
  color"/reset: follow 255 behaviour unless the user decides otherwise.
- `🎨` is double width; validate the bar and picker with `loom eval`, `loom measure` and
  `loom check-box`.

## Decision: input fallbacks (user, 2026-10-20)

All of these are supported; Shift variants remain primary where the terminal forwards them:

- Mouse: click → FG, Shift+click → BG, right-click → BG.
- Keyboard on a focused swatch: Space → active target, Shift+Space → BG, `f` → FG, `b` → BG.
- Visible `FG │ BG` target toggle at the top of the picker (clickable); Space and plain click pick
  for the active target; `x` (and Left/Right at the swatch edge) flips the target. Not Tab (Tab
  belongs to containers).
- Preview shows both colors; Enter and Esc close; Ctrl+Space closes all per 342.
- Rejected: split swatch halves; Enter/Shift+Enter picking.

## Verification

Widget unit tests (FG, BG, both, keyboard and mouse, modifier decoding), popover test that `🎨`
replaces both items, gallery entry, docs/Widgets.md and F1 help updated.
