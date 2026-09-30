# 209 — Remove legacy event handling: one EventResult contract for keys and mouse

**Status**: Closed
**Priority**: P1 (High)
**Severity**: Major
**Category**: Refactor
**Related**: supersedes 196; 112, 199, 200, 203; docs/Widgets.md §Event Handling, §8

---

## 1. Problem & Motivation
loom has four parallel event styles, and mixing them silently drops input:
- `Widget.HandleKey(e) (quit bool)` / `HandleMouse(e) (quit bool)` (true = quit the app; ~124 / ~92 files),
- tuple `ConsumeKey(e) (quit, consumed bool)` (`KeyConsumer`, 13 sites, incl. `media.Widget.ConsumeMouse`),
- `EventResult` `ConsumeKey`/`ConsumeMouse` (`EventConsumer`, `MouseConsumer`),
- adapters `ConsumeKeyEvent`/`ConsumeMouseEvent`.
Bugs this caused: media clicks never delivered (112), gallery wrapper turning every unhandled key into quit (203). The `Widget` doc also still says mouse coordinates are canvas-absolute, contradicting the child-local rule.

User decision (2026-09-30): remove the old ways. Breaking the public API and `examples/` is acceptable; example features may be disabled, each with a follow-up ticket.

## 2. Technical Specification / Findings
Target contract: `Widget` = `Draw` + `ConsumeKey(e KeyEvent) EventResult` + `ConsumeMouse(e MouseEvent) EventResult`, mouse coordinates 0-based and child-local. Delete `HandleKey`/`HandleMouse`, `KeyConsumer`, tuple forms, and the `*Event` adapters; `DispatchKeyEvent`/`DispatchMouseEvent` shrink to one path plus the fallback quit keys. Keep the precedence rule that only unconsumed events reach fallback quit.

## 3. Implementation & Verification Plan
- M1 Removal plan (2026-09-30):
  - Remove `Widget.HandleKey` / `Widget.HandleMouse` and their quit-only boolean contract from `widget.go`; update the lifecycle and mouse-coordinate comments there.
  - Remove `KeyConsumer` (tuple `ConsumeKey`), the dispatch-only `ConsumeKeyEvent` / `ConsumeMouseEvent` adapters, and `media.Widget.ConsumeMouseEvent`. Keep `EventConsumer` and `MouseConsumer` as optional interfaces with the unified `EventResult` signatures; the final widget contract is `ConsumeKey(KeyEvent) EventResult` and `ConsumeMouse(MouseEvent) EventResult`.
  - Remove the compatibility branches in `DispatchKeyEvent` / `DispatchMouseEvent` (`event.go`), and migrate every caller/delegator in `pane.go`, `startup.go`, `popup.go`, `frame.go`, `viewport.go`, `split.go`, `grid.go`, `stack.go`, `tabs.go`, and `form.go` to propagate `EventResult` directly. Preserve the rule that fallback quit keys apply only to unconsumed keys.
  - Convert the root widgets still exposing only or alongside legacy handlers (including `Choice`, `TextArea`, `TextInput`, `View`, `Viewport`, `Dialog`, `Popup`, `Tree`, `FilePicker`, `AnsiEditor`, `Table`, `Settings`, `Confirm`, and leaf widgets with no-op handlers), plus the mixed-contract wrappers in `gallery/`, `cmd/`, `startup.go`, and `media/`. Internal command helpers such as `cmdBar.HandleKey` are not `Widget` APIs; retain them only if they remain private implementation details rather than compatibility dispatch paths.
  - Every standalone example listed below currently implements or calls the legacy widget methods and will need conversion before it compiles against the new `Widget` contract: `ansicanvas_demo`, `ansiedit`, `ansiviewer`, `background`, `filebrowser`, `loomoji`, `media`, `monitor`, `screens`, `splash`, `split`, `tabs`, `textedit`, `textrender`, `treemap`, `usage`, and `winch`.
  - Defer and disable the standalone `examples/media` demo in M3; its demo currently relies on the tuple key API and `ConsumeMouseEvent` adapter. File follow-up issue “Restore media example after EventResult migration” for restoring media input and interaction. All other listed examples are planned for direct conversion; if review finds another conversion is not cheap, disable only that example feature and file one narrowly scoped restoration issue before landing M3.
- M2 Library (root package, `media/`, `gallery/`, `cmd/`): implement the conversions above, delete obsolete interfaces/adapters, update `docs/Widgets.md` §Event Handling and the Widget doc comment. Compile the library and retain existing event/gallery behavior.
- M3 `examples/`: convert the listed examples except a deferred example explicitly called out above; record any further disabled feature in its example and file one follow-up ticket for each.
- Proof: a compile-level guarantee (old methods no longer exist) plus existing gallery PTY tests (200, 203) still passing; `rg 'HandleKey\(|HandleMouse\(|quit, consumed bool|ConsumeKeyEvent|ConsumeMouseEvent'` returns nothing outside history docs.

/goal One event contract remains, the suite is green, and every disabled example feature has a follow-up ticket; or stop and report when blocked on a user decision.
