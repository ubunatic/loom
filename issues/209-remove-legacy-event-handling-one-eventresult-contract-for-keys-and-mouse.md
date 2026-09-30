# 209 — Remove legacy event handling: one EventResult contract for keys and mouse

**Status**: Open
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
- M1 Plan: list every type and call site per style; propose the mechanical mapping (`HandleKey` true → `QuitResult()`, false → `Handled()`/`Ignored()` by whether the widget acted). Host reviews before code.
- M2 Library (root package, `media/`, `gallery/`, `cmd/`): convert, delete old interfaces, update `docs/Widgets.md` §Event Handling and the Widget doc comment. All library tests green.
- M3 `examples/`: convert; where a conversion is not cheap, disable the feature, note it in the example, and file one follow-up ticket per disabled feature.
- Proof: a compile-level guarantee (old methods no longer exist) plus existing gallery PTY tests (200, 203) still passing; `rg 'HandleKey\(|HandleMouse\(|quit, consumed bool|ConsumeKeyEvent|ConsumeMouseEvent'` returns nothing outside history docs.

/goal One event contract remains, the suite is green, and every disabled example feature has a follow-up ticket; or stop and report when blocked on a user decision.
