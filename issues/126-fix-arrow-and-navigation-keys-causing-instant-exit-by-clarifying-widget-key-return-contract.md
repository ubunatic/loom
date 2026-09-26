# 126 — Fix arrow and navigation keys causing instant exit by clarifying widget key return contract

**Status**: Closed
**Resolution**: Implemented `EventResult` value struct and `EventConsumer`/`MouseConsumer` interfaces in `event.go` and `widget.go`, integrated with `pane.go` (`dispatchKey`, `dispatchMouse`) and fallback safety. Updated `ansiedit` and composite container widgets to consume navigation keys cleanly (`Handled()`), and verified through comprehensive unit and PTY tests.
**Commits**: `de2b802` (M1), `ec8848a` (M2)
**Related**: `pane.go`, `widget.go`, `event.go`, `docs/Widgets.md`, `examples/ansiedit/`, `examples/textedit/`, `examples/filebrowser/`

---

## Goal

`/goal`: Fix the recurring architectural bug where arrow keys and navigation inputs cause applications and example widgets (such as `ansiedit`, `textedit`, etc.) to terminate immediately on key press. Replace ambiguous bool returns with a small, value-type struct (not a pointer) for event results (e.g. `EventResult` / `KeyResult` with `Consumed` and `Quit` flags), integrate with `KeyConsumer` / `Widget` event dispatch in `pane.go`, and update examples (`ansiedit`, etc.) so navigation keys work properly.

## 1. Problem & Root Cause

1. **Boolean Ambiguity in `Widget.HandleKey(KeyEvent) bool`**:
   - Returning `bool` creates constant bugs because developers assume `true` = "consumed/handled", while `pane.Run` treated `true` = "quit application".
   - When an author implements arrow key navigation or text editing returning `true`, pressing any arrow key or handled hotkey immediately exits the whole program.

2. **Contrast with `KeyConsumer` interface**:
   - `KeyConsumer` returns `(quit bool, consumed bool)`. We should formalize event results as a small struct passed by value (not pointer):
     ```go
     type EventResult struct {
         Consumed bool
         Quit     bool
     }
     ```
     With helper constructors/constants, e.g.:
     - `Handled()` / `Consumed()` -> `EventResult{Consumed: true, Quit: false}`
     - `Ignored()` / `Unhandled()` -> `EventResult{Consumed: false, Quit: false}`
     - `Quit()` -> `EventResult{Consumed: true, Quit: true}`

3. **Fallback Safety**:
   - `pane.dispatchKey` must cleanly interpret `EventResult` or `KeyConsumer`, and never treat navigation keys (`up`, `down`, `left`, `right`, `home`, `end`, `pgup`, `pgdn`) as default quit triggers.

## 2. Milestones

- **M1 (EventResult Value Struct & Dispatch Contract)**:
  - Define `EventResult` value-type struct (non-pointer) with `Consumed` and `Quit` booleans and standard helpers (`Handled()`, `Ignored()`, `Quit()`).
  - Support `EventResult` / `KeyConsumer` in `pane.go` `dispatchKey` and `dispatchMouse`.
  - Update `docs/Widgets.md` and SDK documentation to specify the contract clearly.

- **M2 (Ansiedit & Examples Event Handler Update)**:
  - Update `examples/ansiedit/ansiedit/app.go` to use `KeyConsumer` / `EventResult`, ensuring all navigation keys (`up`, `down`, `left`, `right`, `ctrl-arrows`, `tab`, `del`, `backspace`, etc.) consume the event (`Consumed: true, Quit: false`) without quitting.
  - Verify `F10` and `Ctrl-Q` explicitly request quit (`Quit: true`).

- **M3 (Tests & Verification)**:
  - Add unit and PTY tests in `examples/ansiedit/` and `pane_test.go` asserting that arrow keys navigate and do not quit.
  - Run full suite (`go test ./...`, `make test-q1`) and verify `make install`.
