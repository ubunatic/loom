# 153 — Hide the filebrowser search UI until search is active

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature

---

## 1. Problem & Motivation

The file browser uses `/` to activate search by default, but always shows the filter prompt and cursor. This makes the UI look ready for typing even while printable keys go to application shortcuts.

## 2. Technical Specification / Findings

In slash-gated mode, hide the editable filter row and cursor while idle, and show a subtle `/ search` hint. Show the input when `/` starts search; Enter keeps the filter applied without leaving an idle cursor, and Esc clears it. Keep the input visible in `TypeToSearch` mode, where typing always filters. Example states are in [filebrowser-slash-search.ansi](../docs/data/filebrowser-slash-search.ansi) and [filebrowser-type-to-search.ansi](../docs/data/filebrowser-type-to-search.ansi).

## 3. Implementation & Verification Plan

- **/goal**: Make the filebrowser search UI reflect whether typing is routed to search while preserving slash-gated and type-to-search behavior, or stop and report when blocked on a user decision or denied permission.
