# 408 — loom view and loom.View: add find on ctrl-f

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 407 (key policy: ctrl-f is find), 368

---

/goal Add find to `loom.View` and `loom view`, bound to ctrl-f in the spec, or stop and report when blocked on a user decision or denied permission.

## 1. Problem
Since 407, ctrl-f is reserved for find and `loom.View` leaves it unconsumed, but neither `loom.View` nor `loom view` (`cmd/loom` `ansiView`) has find. Split out of 407 step 2.

## 2. Plan
1. Add a `find` binding (`ctrl-f`) under `view.bindings` in `spec/defaults.yaml`.
2. Reuse `loom.SearchBar` (docs/Widgets.md section 15) for the prompt; jump to and highlight matches, next/previous match.
3. Use it in `loom view` and document it in docs/Widgets.md section 11.

Open question: inside `loom session`, ctrl-f opens the session's scrollback search first (`spec/defaults.yaml` session search binding), so view find there needs the prefix or another route. Decide with the owner.

Event routing ticket: start on `codex:sol:med` (AGENTS.md). Re-check current code before starting.
