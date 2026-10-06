# 242 — Gallery All tab: Dialog demo leaks left/right arrow keys to the grid

**Status**: Closed — Dialog/Table consume keys they act on; test-q1 green
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: 209 (one EventResult contract), 232 (All tab widgets)

---

/goal In the All tab, left/right arrows that move the "Save?" Dialog's button selection are consumed and do
not move grid focus, with key and mouse path tests; or stop and report when blocked on a user decision.

## 1. Problem & Motivation
User report: in `loom widgets` → All, the "Save?" Dialog demo leaks arrow keys. Right arrow selects "Yes"
but also moves grid focus to the next cell; left arrow selects "No" and moves focus to the previous cell.

## 2. Technical Specification / Findings
- A key the Dialog uses must return `Handled()`, so `Grid` (`grid.go`) does not also navigate.
  Either Dialog returns `Ignored()` for arrows it acted on, or Grid navigates before/regardless of the
  child result; find which.
- Fix the library (Dialog or Grid routing), not the gallery demo. Event routing ticket: start on
  `codex:sol:med` (AGENTS.md).

## 3. Implementation & Verification Plan
- Reproduction test first: Grid with a Dialog child, right/left arrow changes the Dialog selection and
  keeps grid focus. Also check that an arrow the Dialog does not use still moves grid focus.
- Mouse path test per AGENTS.md; `make test-q1`, `make install`, PTY check in the All tab.

---

## Delivered

- Dialog returns `Handled()` for keys it acts on (left/right selection change, Enter, Esc); Esc/Enter close only the dialog, never the host. Tab is left to the container (library focus key; the gallery uses it to switch demos). Table consumes filter, sort (Tab, `!`), paging, help, command and selection keys; no-op arrows still bubble.
- Commits: `c3fe9a6` (dev242, codex:sol:med); host fixes `f6afcaf` (Dialog leaves Tab, Grid test layout, racy PTY check, box validator accepts a nested box in a grid cell) and `68f9ce5` (pane reader closed a reassigned channel: panic surfaced by the suite).
- Host verification: `make test-q1` 0 FAIL; `make install`; installed `loom widgets --show All`: right arrow toggles No/Yes, Esc closes the dialog and the app keeps running.
