# 343 — loom edit: hotkey to toggle mousegrab

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 342 (mousegrab default on), 303 (temporary mouse grab)

---

/goal Add a specced hotkey that toggles mousegrab in `loom edit`, test and document it, run `make test-q1` and `make install`, or stop and report when blocked on a user decision (key choice) or denied permission.

## Problem

Mousegrab can only be switched with `--mousegrab` at startup or by clicking the `m`/`M` status
indicator (`cmd/loom/edit.go:495`, `toggleMouseGrab` at l.968). Unverified: with grab off the
terminal sends no clicks, so the `m` indicator likely cannot turn grab back on; users who turn it
off to use native text selection must restart to get it back.

## Plan

- Verify the unverified claim above first (PTY test: grab off, click indicator).
- Add a binding in the `editor` section of spec/defaults.yaml, following the existing
  `hotkey_<name>_binding` / `_key` / `_label` pattern (l.150ff). Key not chosen yet; ask the user
  (candidates: a free F-key or Ctrl-key; F2/F3 are taken by Files/Search).
- Bind it to `toggleMouseGrab`; list it in F1 help (and hotkey bar if space allows).
- Tests: key toggles on→off→on and updates the indicator; works while grab is off.
- Docs: docs/Editor.md (Mouse Capture, Bottom Chrome).
