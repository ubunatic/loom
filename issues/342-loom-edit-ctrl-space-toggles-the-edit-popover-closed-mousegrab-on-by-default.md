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
