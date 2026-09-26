# 124 — Make F10 standard global quit key across loom applications

**Status**: Closed — pane-level global F10 quit with opt-out, focus-state tests, hint audit (4a1b668)
**Priority**: P1 (High)
**Severity**: Normal
**Category**: Architecture / Usability
**Related**: `pkg/ui/frame/`, `pkg/ui/widget/`, `examples/`, `docs/Widgets.md`

---

## Goal

`/goal`: Standardize `F10` as the universal, top-priority global quit key across the Loom framework and all example applications/widgets, ensuring that pressing `F10` exits the application immediately from any state, modal, editor, or nested focused widget without conflict.

## 1. Context & Motivation

- Modern terminal standards (Midnight Commander, DOS/Norton Commander, Turbo Vision, standard TUI convention) use `F10` as the standard function key to exit the entire program.
- In text inputs or graphic editing modes, single-letter shortcuts like `q` or `ESC` cannot reliably act as global quit keys because they collide with character input or modal dismissals.
- While individual applications (such as `examples/textedit` and `examples/ansiviewer`) have added local handlers for `F10`, Loom should provide framework-level support / default handling ensuring `F10` is universally respected across any hosted widget or application hierarchy.

## 2. Scope & Acceptance Criteria

1. **Framework-Level Global Quit**:
   - Ensure `loom.Frame` / root event dispatching provides built-in or easily enabled top-level interception for `F10` (alongside `Ctrl-Q` / `Ctrl-C` conventions where appropriate).
   - Ensure nested or child widgets with input focus (editors, canvases, palettes, modals) cannot trap or swallow `F10` unless explicitly opted out by design.

2. **Application Consistency**:
   - Verify all example applications (`ansiviewer`, `ansiedit`, `textedit`, `filebrowser`, `paint`, etc.) reflect `F10 Quit` in their status bar / keycap footers and terminate cleanly upon pressing `F10`.

3. **Automated Testing**:
   - Add unit/regression tests verifying `F10` key events request termination across diverse focus states.

## Delivered

- Added pre-dispatch Pane F10 quit handling with an opt-out, focus-state regressions, and `F10 Quit` hints across the example apps.
