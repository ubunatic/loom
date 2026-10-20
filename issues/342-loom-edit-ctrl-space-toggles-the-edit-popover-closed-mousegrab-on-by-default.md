# 342 — loom edit: Ctrl+Space toggles the edit popover closed; mousegrab on by default

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 303 (temporary mouse grab)

---

/goal Implement both changes, spec and test them, run `make test-q1` and `make install`, or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation

1. In `loom edit`, Ctrl+Space opens the edit popover (format/Box menu). Pressing Ctrl+Space again
   while it is open should close it (toggle). Today this is not specified; docs/Widgets.md (~l.377)
   and docs/TuiInput.md (l.81) only describe opening/reopening.
2. `--mousegrab` should be on by default. Today `spec/defaults.yaml:148` has `mousegrab: false`,
   read via `SpeccedDefaults.Editor.MouseGrab` in `cmd/loom/edit.go:97`.

## 2. Plan & Verification

- Popover: Ctrl+Space while the popover (or its submenu) is open closes it, focus returns to the
  editor, document unchanged. Record the toggle wherever key behaviour is specced (spec YAML if a
  key/menu spec exists, else docs/Widgets.md and docs/TuiInput.md). Fix in the library widget, not
  the caller. Event routing ticket: start on `codex:sol:med`.
- Mousegrab: set `mousegrab: true` in `spec/defaults.yaml`; check that `--mousegrab=false` / `-m=false`
  still disables it and update help/man/docs mentioning the old default.
- Tests: unit test for the Ctrl+Space toggle; default test for mousegrab.

## 3. Searcher findings (2026-10-20)

- `RichTextEdit.consumeKey` (richtextedit_input.go:86) calls `consumePopoverKey` before the
  Ctrl+Space binding (l.134). Unfocused popover or Box submenu: Ctrl+Space falls through and
  reopens/resets the popover; it never closes it. No spec YAML covers Ctrl+Space; document in
  docs/Widgets.md:377 and docs/TuiInput.md:81.
- Toggle test goes beside `TestRichTextEditCtrlSpaceSelectsWordAndOpensPopover` (richtextedit_test.go:964).
- Mousegrab tests: `TestEditorFlagDefaultsComeFromSpec` (cmd/loom/edit_spec_test.go:58),
  `TestLoadEditorConfigDefaults` (editor_config_test.go:15), override test cmd/loom/main_test.go:1048.
  Stale doc: docs/Editor.md:53 ("mouse capture off").
- Write set: richtextedit_popover_input.go, richtextedit_test.go, spec/defaults.yaml, one of the
  default tests, docs/Widgets.md, docs/TuiInput.md, docs/Editor.md.

## M1 — Popover toggle + mousegrab default (single milestone)

M1 delivered (e13dc01): popover toggle and mousegrab default; regression tests failed before fix, pass after.

## M2 — PTY tests for the new mousegrab default

Pre-Work / Required Refinements (from M1 review): `make test-q1` fails in three cmd/loom PTY tests that
assumed mouse capture off by default: `Test302EditPTYUntitledAndMouseOff`, `Test303EditPTYTemporaryMouseGrab`,
`TestEditPTYHeaderHasNoDuplicateFileLabel`. Tests that exercise mouse-off mode must pass `--mousegrab=false`
explicitly (keep their intent); a test that only failed by side effect must be fixed for the new default,
not loosened. Then the full `make test-q1` must be green (ticket 341 flake excepted, reported if hit).
