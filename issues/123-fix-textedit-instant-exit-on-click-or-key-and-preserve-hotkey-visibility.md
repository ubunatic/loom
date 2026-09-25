# 123 — Fix textedit instant exit on click or key and preserve hotkey visibility

**Status**: Closed — Fixed in `04192a0`: top-level handlers return `false` on consumed events, `F10`/`Ctrl-Q` quit preserved globally, hotkey hints remain visible after saving
**Priority**: P1 (High)
**Severity**: Critical
**Category**: Bug
**Related**: `examples/textedit/textedit/textedit.go`, `examples/textedit/textedit/textedit_test.go`, `docs/Widgets.md`

---

## Goal

`/goal`: Resolve the bug in `examples/textedit` where clicking anywhere or pressing any key caused the application to exit immediately, ensure `F10` and `Ctrl-Q` always quit even when the editor is focused, and keep hotkey hints visible in the status bar after saving files.

## 1. Problem & Root Cause

1. **Instant Exit on Input**:
   In Loom's application event loop contract, returning `true` from a root widget's `HandleKey` or `HandleMouse` method requests quitting the application event loop, rather than signaling that the event was consumed.
   In `examples/textedit/textedit/textedit.go`, `HandleKey` and `HandleMouse` returned `true` whenever an event was handled or forwarded from child widgets. As a result, clicking anywhere or pressing any key immediately terminated the program.

2. **Quit Key Precedence & Editor Input**:
   When typing into the editor, pressing `q` was initially intercepted as the frame quit action, preventing typing `q` as normal text. Conversely, global quit shortcuts (`F10`, `Ctrl-Q`) needed to reliably quit even when text entry was active in the focused editor.

3. **Status Bar Hotkeys Disappearing on Save**:
   `app.statusMsg` originally contained the initial hotkey string and was completely overwritten by `"Saved " + path` or `"Opened " + path`, removing the shortcut hints from view.

## 2. Fix Implementation

1. **Event Return Values**:
   Updated `HandleKey` and `HandleMouse` to return `false` when consuming inputs and actions, keeping the event loop alive.

2. **Global Quit Handling**:
   - Intercepted `F10` and `Ctrl-Q` at the top of `HandleKey` to return `true` immediately regardless of active focus area.
   - Allowed `q` to pass through to the active editor as ordinary character input.
   - Configured `app.frame.Actions` with `quit_f10` and `quit_ctrl_q`.

3. **Persistent Hotkey Status Hints**:
   - Integrated Loom's native `loom.FrameAction` hints and `ControlSeparator` on `app.frame`.
   - `app.frame.Status` now carries dynamic messages (e.g. `Saved demo.go`), while the action hint list (`[C-s] Save`, `[Tab] Focus`, `[F10] Quit`) remains persistently attached to the frame status bar.

4. **Regression Tests**:
   Added regression test cases in `examples/textedit/textedit/textedit_test.go`:
   - `TestInputDoesNotTreatConsumedEventsAsQuit`
   - `TestF10AndCtrlQAlwaysQuitWhenEditorFocused`
   - `TestHotkeysStayVisibleAfterSave`

## 3. Verification

- `make test-q1` passed (all unit tests, spec validation, and PTY smoke suites).
- `make install` and `git diff --check` passed cleanly.
- Verified in live interactive runs and headless PTY recordings.
