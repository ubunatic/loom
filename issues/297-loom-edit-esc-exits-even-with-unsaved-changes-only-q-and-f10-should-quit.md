# 297 — loom edit: Esc exits even with unsaved changes, only ^Q and F10 should quit

**Status**: Closed — implemented and verified in c09abb7
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [294](294-loom-edit-add-o-to-open-file-pane-and-w-to-close-file.md), [290](290-keep-the-loom-edit-file-browser-usable-after-cancel-and-mouse-activation.md), [291](291-loom-edit-s-saves-file-but-saved-modified-indicator-is-not-updated.md)

---

## 1. Problem & Motivation
`Esc` at the top level of `loom edit` exits immediately, even with unsaved changes. The same framework default makes `Esc` (and `q`) quit every Loom app and demo when the widget does not consume it.

Framework-wide defaults (required):
- **Esc** closes the topmost dialog, overlay, popup or menu, or clears the selection or mode. It **never** closes the app unless the app explicitly opts in.
- **^Q and F10** **always** close the app unless the app explicitly opts out.
- **Close interception**: stateful apps (e.g. `loom edit`) intercept every close request (`^Q`, `F10`, SIGHUP / terminal close) to ask Save / Discard / Cancel when there is unsaved work.

## 2. Technical Specification / Findings
- `spec/defaults.yaml:2-7` `fallback_quit_keys: [esc, ctrl-c, ctrl-q, ctrl-d, q]`. `pane.go:1508` `handleKeyFallback` quits on any of these when the root returns not-consumed, unless `DisableDefaultQuit` is set.
- `cmd/loom/edit.go:96` sets only `DisableGlobalF10Quit`, not `DisableDefaultQuit`. So a top-level `Esc` that `RichTextEdit` ignores reaches the Pane fallback and quits **without** going through `handleQuit()` (`edit.go:237`). This is the bug.
- `edit.go` scatters quit checks (`:533`, `:577` `ctrl-q/ctrl-c/ctrl-d/f10`, `:611`, `:623`, `:637`, `:772`), each calling `handleQuit` and setting `shouldQuit`. That is a per-app workaround for a missing Pane contract.
- `pane.go:1465` handles global F10 before dispatch; `pane.go:891` `OwnsQuit` disables both defaults wholesale.
- Library changes:
  1. Spec: replace `fallback_quit_keys` with `quit_keys: [ctrl-q, f10]` (always-on, opt-out) and `escape_quits: false` (opt-in). `ctrl-c` stays an interrupt that also goes through interception; `ctrl-d` and `q` are removed from defaults (apps like the gallery may opt in).
  2. Pane: `Esc` is dispatched to the topmost overlay, then the focused widget. If nobody consumes it, it is a no-op.
  3. Pane: add `OnCloseRequest func(reason CloseReason) CloseDecision` (Allow / Veto). The quit keys, `ctrl-c` and terminal hangup all call it. A veto keeps the loop running so the app can show a dialog and call `pane.Quit()` later.
  4. Shared `UnsavedChangesGuard` helper (Dialog Save / Discard / Cancel) used by `loom edit` close, `^W` (294) and open-with-dirty-buffer. Labels come from the spec.
- `loom edit` then drops its manual quit-key branches and `shouldQuit` plumbing.

## 3. Implementation & Verification Plan
- Spec, schema and Go accessors; Pane routing and `OnCloseRequest`; the guard helper; migrate `loom edit`; audit the demos and the gallery that relied on `esc`/`q` quitting, and set the opt-in where intended.
- Add a `docs/Upgrading.md` breaking-change row (Esc/q no longer quit by default; `fallback_quit_keys` renamed) and update docs/Widgets.md.
- Tests: Pane unit (Esc unhandled leaves the app running; `^Q`/F10 quit; veto keeps running; opt-out and opt-in flags). PTY in `loom edit`: type then Esc keeps running with text intact; Esc with an open dialog or panel closes only that; `^Q`/F10 dirty shows the dialog; Discard quits, Cancel returns, Save writes then quits.
- `make test-q1`, `make install`.

/goal Make Esc never quit by default and ^Q/F10 always quit by default (spec-configurable), add a Pane close-request veto hook with a shared Save/Discard/Cancel guard, migrate loom edit and the demos to it, and verify with Pane unit and loom edit PTY tests, or stop and report when blocked on a user decision or denied permission.
