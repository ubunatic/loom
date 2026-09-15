# 049 — Add PgUp / PgDn and Home / End support to loom.Choice

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: `choice.go`, `choice_test.go`, `event.go`

---

## 1. Problem & Motivation

`loom.Choice.HandleKey` currently only handles single-step `up` and `down` navigation:
- `pgup` and `pgdown` (or `pgdn`) keys are ignored and fall through to printable text matching or no-op.
- `home` and `end` keys are also not handled in `Choice.HandleKey`.

When viewing large directories or long lists of choices, users expect standard pager navigation:
- `pgup` / `pageup`: jump up by visible page size (or `itemRows`).
- `pgdown` / `pgdn` / `pagedown`: jump down by visible page size.
- `home`: jump to top of list (first item).
- `end`: jump to bottom of list (last item).

In addition, `loom.DecodeKey` emits `pgdown` for CSI `\x1b[6~`, while widgets and downstream code often test against `pgdn` or `pgdown`.

---

## 2. Technical Implementation

1. **`choice.go`**:
   - In `HandleKey`:
     - `case "pgup", "pageup":` step `sel` backwards by `max(1, c.itemRows-1)` (clamped to 0).
     - `case "pgdown", "pgdn", "pagedown":` step `sel` forwards by `max(1, c.itemRows-1)` (clamped to `len(c.filtered)-1`).
     - `case "home":` `c.sel = 0`.
     - `case "end":` `c.sel = max(0, len(c.filtered)-1)`.
2. **`choice_test.go`**:
   - Add unit tests verifying `pgup`, `pgdown`, `home`, and `end` jump `sel` correctly.

---

## 3. Verification & Resolution

- Run `go test ./...` in `loom`.
- Verify in `imgbrowser` that PgUp / PgDn jumps by a page in the files list and preview pane.
