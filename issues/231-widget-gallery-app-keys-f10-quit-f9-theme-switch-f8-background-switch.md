# 231 — Widget gallery app keys: F10 quit, F9 theme switch, F8 background switch

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `cmd/loom/widgets.go`, `gallery/gallery.go`, 201, 203, 205

---

## 1. Problem & Motivation

Standardize application-level keybindings and status hints in the `loom widgets` gallery:
- `F10`: Quit the gallery application.
- `F9`: Cycle color themes (migrating/updating from `F2`).
- `F8`: Cycle application background (e.g., none/default, `astra`, and any other available backgrounds).
- `q` and `Esc`: Remain available as hidden quit keys (unless consumed/needed by the active focused widget, such as text inputs or dialogs).
- Update the gallery bottom status bar to display the active theme, active background, and key legend (`F8 Background | F9 Theme | F10 Quit`).

## 2. Technical Specification / Findings

- In `cmd/loom/widgets.go` (`themedGallery`), update key dispatch:
  - `F10` returns `loom.QuitResult()`.
  - `F9` cycles through `loom.ThemeNames()`.
  - `F8` cycles through available backgrounds (e.g. `nil`, `loom.NewAstraBackground()`, etc.).
  - `q` and unconsumed `Esc` remain fallback quit keys without intercepting keys needed by child widgets.
- Update `themedGallery.Draw` to show active background and theme status along with key hints.

## 3. Implementation & Verification Plan

- Update `themedGallery` key routing, background state, and status bar rendering in `cmd/loom/widgets.go`.
- Add/update tests in `cmd/loom/main_test.go` and `gallery/gallery_test.go` for key handling and background switching.
- Verify with `make test-q1`.

/goal Implement gallery key routing for F10, F9, F8, and hidden q/Esc, update status bar, and verify with tests; or stop and report when blocked on a user decision or denied permission.
