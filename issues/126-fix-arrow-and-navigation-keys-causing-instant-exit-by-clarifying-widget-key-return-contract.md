# 126 — Fix arrow and navigation keys causing instant exit by clarifying widget key return contract

**Status**: Open
**Priority**: P0 (Urgent)
**Severity**: Critical
**Category**: Architecture / Bug
**Related**: `pane.go`, `widget.go`, `docs/Widgets.md`, `examples/ansiedit/`, `examples/textedit/`, `examples/filebrowser/`

---

## Goal

`/goal`: Fix the recurring architectural bug where arrow keys and navigation inputs cause applications and example widgets (such as `ansiedit`, `textedit`, etc.) to terminate immediately on key press. Implement framework-level prevention and clean contract ergonomics so widget authors cannot accidentally confuse event consumption (`consumed = true`) with application exit (`quit = true`).

## 1. Problem & Root Cause

1. **Boolean Ambiguity in `Widget.HandleKey(KeyEvent) bool`**:
   - In Loom's standard `Widget` interface:
     ```go
     type Widget interface {
         Draw(*Canvas, Rect)
         HandleKey(KeyEvent) bool
         HandleMouse(MouseEvent) bool
     }
     ```
   - For a root widget run directly via `pane.Run(root)`, returning `true` from `HandleKey` signals **"Quit the application event loop"**, whereas returning `false` signals **"Keep the event loop running"**.
   - Almost every developer and widget author intuitively assumes `return true` means **"I handled/consumed this event"** (e.g. "I processed the Up arrow key, don't pass it to fallbacks").
   - When an author implements arrow key navigation, text insertion, or panel switching returning `true`, pressing any arrow key or handled hotkey immediately exits the whole program.

2. **Contrast with `KeyConsumer` interface**:
   - Issue #057 / #063 introduced `KeyConsumer`:
     ```go
     type KeyConsumer interface {
         ConsumeKey(KeyEvent) (quit bool, consumed bool)
     }
     ```
   - `pane.dispatchKey` checks `KeyConsumer` first (`if quit, consumed := c.ConsumeKey(ke); consumed { return quit }`), but falls back to `root.HandleKey(ke)` where the boolean meaning is inverted between consumption and quitting.

3. **Silent Fallback to Default Quit Keys**:
   - When `HandleKey` returns `false` (meaning "event consumed, do not quit"), `pane.dispatchKey` executes:
     ```go
     return root.HandleKey(ke) || p.handleKeyFallback(ke)
     ```
   - If an unhandled key is passed, `handleKeyFallback` checks `defaultQuitKeyMap` (`q`, `esc`, `ctrl-c`, etc.), which can unintentionally terminate editing applications unless `OwnsQuit: true` / `DisableDefaultQuit: true` is explicitly configured.

## 2. Solution & Architectural Fixes

1. **Framework-Level Contract / Widget Wrapper**:
   - Clarify and document the root event dispatch contract across `docs/Widgets.md` and SDK docstrings.
   - Prefer or encourage `KeyConsumer` / explicit `App` wrappers at the library level so `(quit bool, consumed bool)` is unambiguous for root applications.
   - Ensure `ansiedit`, `textedit`, and any other interactive examples implement `KeyConsumer` or correct `HandleKey` return semantics (`return false` on normal handled keystrokes, `return true` only on explicit quit requests).

2. **Library-Level Safety Guards**:
   - Audit `pane.dispatchKey` to ensure standard navigation keys (`up`, `down`, `left`, `right`, `home`, `end`, `pgup`, `pgdn`) are never inadvertently treated as default quit triggers by fallback handlers.
   - Verify `OwnsQuit` / `DisableDefaultQuit` interactions when `PaneRequest` is declared by widgets.

3. **Automated Regression Tests**:
   - Add unit tests in `pane_test.go` and example tests (e.g., `ansiedit_test.go`) simulating arrow key presses, ensuring arrow navigation updates state and does NOT return `quit=true`.
