# 456 — TextEntry passes through keys that have no effect

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 438, 448, 454

---

## 1. Problem & Motivation
Owner request: a `TextEntry` that cannot act on a key should return `Ignored()`, so the parent (a global shortcut, the app) can catch it. Today `Ctrl+B` is always captured when bold is enabled, even with an empty entry and no selection. Examples from the owner: `^B` with nothing to make bold or plain, `^C` and `^X` with nothing to copy or cut.

## 2. Technical Specification / Findings
Audit of `ConsumeKey` (`text_entry.go`, checked against HEAD cc0a3fb1):

- **Already pass when ineffective**: `Ctrl+C` / `Ctrl+X` without a selection, `Ctrl+V` with an empty clipboard, `Ctrl+A` on empty text, `Ctrl+Z` / `Ctrl+Y` with empty history, `Backspace` / `Delete` at the text edges, `Ctrl+U` at the start, `Up` / `Down` / `PgUp` / `PgDn` on the first or last row, the wheel at the scroll limits, Enter / `Ctrl+Enter` without `OnSubmit`.
- **Always consumed, though they can be no-ops**:
  - `Ctrl+B` (bold enabled): no selection toggles the typing style, so it is never ignored. Open question: with an empty entry and no selection it still changes what the next typed text looks like. The owner's wording says to pass it; decide whether the typing-style toggle stays when the text is empty.
  - `Left` / `Right` (and `Shift`) at the start or end of the text.
  - `Ctrl+Left` / `Ctrl+Right` at the edges; `Home` / `End` already at the row edge; `Ctrl+Home` / `Ctrl+End` already at the text edge (and the `Shift` forms when the selection does not change).
  - `Ctrl+A` when everything is already selected.
  - `Ctrl+V` when the paste would not change the text (empty paste).
  - `Tab` / `End` / `Ctrl+E` placeholder accept: already guarded by `AcceptPlaceholder`, fine.
- **Typed text and newline/submit keys** stay captured: an `Enter` that submits is the entry's job even with an empty value (the app's `OnSubmit` decides).
- Which are worth the change: the ones a parent plausibly binds. `Ctrl+B`, `Left` / `Right` at the edges (a parent could use them to leave the entry, as `Up` / `Down` already do) and `Home` / `End` / `Ctrl+Home` / `Ctrl+End`. Word jumps and select-all are low value. Unverified: how `Pane` / containers route an `Ignored` key (check docs/Widgets.md key routing before changing semantics); this is a behavior change for callers that rely on arrow keys never leaving the entry, so it may need a row in `docs/Upgrading.md`.

## 3. Implementation & Verification Plan
- Decide the set (owner), then make each chosen key return `Ignored()` when it would change neither text, caret, selection, scroll nor typing style. Make a shared "did anything change" check, not one-off guards.
- Tests per key: ineffective case is `Ignored`, effective case is `Consumed`; the gallery demo parent shows a pass-through (for example a status line).
- Docs: `docs/Widgets.md` TextEntry key table gets an "ignored when" column; `docs/Upgrading.md` if arrow keys change.

/goal Make `TextEntry` return `Ignored()` for the agreed keys when they have no effect, with tests and docs, then stop; or stop and report when blocked on the key set decision or a denied permission.
