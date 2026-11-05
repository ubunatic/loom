# 423 — Widget gallery coverage audit and human test runbook

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Docs / Testing
**Related**: 195, 207 (gallery), 102, 313 (manual check lists), 417, 418, 422 (TextEntry, maybe not demoed)

---

/goal Every public loom widget and feature has a live demo in the widget gallery (`loom widgets --show`), and
the owner has a runbook to test each demo and send feedback; or stop and report when blocked on an owner
decision or denied permission. On a loop or no progress, clear the goal and write a failure report here.

## 1. Problem & Motivation
The gallery (`gallery/`) was built in 195/207, and widgets added since then (for example `loom.TextEntry`,
its caret API, highlight hook and runtime placeholder from 417/418/422) may have no demo. The owner wants one
place to try every widget by hand and give structured feedback.

## 2. Scope
- **M1 Audit:** list every exported widget and user-visible feature of the root `loom` package (and sub-packages
  with visual output, such as `graph`, `syntax`) against the gallery tabs. Output a coverage table in this ticket:
  feature, gallery tab or `missing`, notes. Check the live code, not this ticket.
- **M2 Demos:** add a gallery demo for each `missing` item, following the existing gallery patterns and tests
  (`gallery/all_parity_test.go`). Features that cannot be shown in a gallery tab are listed with the reason.
- **M3 Runbook:** `docs/GalleryRunbook.md` (evergreen): how to start the gallery, then one short section per tab
  with numbered steps, the expected result, and a feedback line (`ok` / `bug: …` / `idea: …`). The owner fills it
  in and hands it back; the host turns each `bug`/`idea` into tickets via `/issue`.
- Start the runbook as a manual test in foot per CLAUDE.md, with the full steps in chat.

## 3. Verification
Gallery tests green with the new demos (`make test-q1`); `loom eval`/`loom measure` for any new ANSI design;
coverage table shows no unexplained `missing`; runbook covers every gallery tab.
