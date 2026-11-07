# 443 — ColorPicker selectable faint text and feasible subdued background

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: issues/442, issues/346, issues/347

---

## Problem

After reviewing the comparison in issue 442, the owner wants faint text to be selectable in ColorPicker, and a subdued background if it can be supported without hacks. The existing comparison is otherwise approved. Faint is a text attribute; the theme-background sample is not currently a semantic dim-background color.

## Goal

/goal Recheck the live ColorPicker, style and persistence contracts; provide a selectable faint-text option that is applied and preserved correctly. Investigate whether a subdued background can be offered cleanly without hacks; implement it only if a sound representation and behavior are available, otherwise document the limitation and report it. Verify selection, rendering, persistence and existing color behavior with appropriate tests. Stop and report when blocked on a user decision or denied permission; on a loop, hard architectural or OS blocker, or no progress, clear the goal and record why in this ticket.

## Searcher Findings (s443)

- **ColorPicker Structure**: Defined in `colorpicker.go`. Uses `OnChange(target, color)` where target is foreground or background; currently has no attribute or style callback.
- **Style & Dim**: `ansi/style.go` provides `Style.Dim` (ANSI SGR 2).
- **Subdued Background**: ANSI has no background dim equivalent. Theme surfaces (`theme.go`) provide `NormalBG`, but there is no semantic dim-background role or mechanism without inventing arbitrary RGB blends. Per goal, this limitation should be cleanly documented.
- **Consumers & Persistence**: `RichTextEdit` (`richtextedit_style.go`) handles span styles; spans store `Style` including `Dim`.

## Milestones

- **M1 — Developer Plan & Feasibility Analysis**:
  - Live code inspection completed by `dev443` (`terra:med`).
  - ColorPicker faint text design: Add `Dim bool` and `OnDimChange func(bool)` to `ColorPicker`, retaining `OnChange` and constructor signature. Display a labeled toggle control after the FG palette row. Right from last FG swatch navigates to Faint, Left returns; Space or click toggles, Enter toggles and closes. Preview reflects candidate `Dim`. BG row does not offer faint.
  - Subdued background: Confirm ANSI/Loom limitation (`Style.Dim` is text SGR 2; `NormalBG` is a surface, not a dimmed background). Document in `docs/Widgets.md` and keep theme sample as comparison only.
  - RichTextEdit: Connect `OnDimChange` to `ActiveStyle` / selection spans, preserving other style attributes.
  - Test strategy: TDD regression tests for keyboard/mouse toggling, preview styling, reset preservation, RichTextEdit undo/serialization, and gallery PTY verification.

- **M2 — Implementation & Verification (Pre-Work / Scope)**:
  - Implement regression-first tests in `colorpicker_test.go` and `richtextedit_test.go` that demonstrate failure before the fix.
  - Implement `Dim`, `OnDimChange`, navigation, rendering, and mouse handling in `colorpicker.go`.
  - Wire `OnDimChange` in `richtextedit_style.go` with undo/span preservation.
  - Update `docs/Widgets.md` documenting faint text selection and the architectural limitation of subdued background.
  - Update `gallery/colorpicker.go`, `docs/progress/ColorPicker.ansi`, and validate ANSI design (`loom eval`, `loom measure`).
  - Refresh public API test snapshot (`go test ./internal/apitest -run '^TestPublicAPI$' -update`) if exported API changes.
  - Run `go vet ./...`, package tests, and logged `make test-q1`. Run `make install`.


