# 095 — Add human-observable PTY test view mode and feedback flow

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Enhancement
**Related**: Issues 088–094

---

## Problem & Motivation

PTY tests prove terminal behavior, but a fully automated run is difficult for a
human to observe and assess. Test runs need a deliberate, human-speed viewing
mode with an explicit handoff between scenarios and a low-friction way to turn
observations into tracker drafts.

## Scope

- Add a PTY test view mode that renders the tested TUI at regular human speed,
  making each interaction observable.
- Between tests, pause and let the observer enter feedback or confirm starting
  the next test. Keep the test sequence and current result visible.
- When feedback is entered, automatically create a draft via `harnez issue new`
  using the raw comments as the draft body and a proposed short, generic title.
- Let the user accept, edit, or cancel the proposed title and draft. When the
  user chooses to write feedback or change the title, support opening `$EDITOR`
  with familiar git-style behavior before filing.
- Preserve a non-interactive/automated mode for CI and unattended regression
  runs, with no observer prompts.

Define interruption, timeout, empty-feedback, cancellation, and failed-test
behavior, and ensure raw observer comments are not lost.

## Goal

/goal: Make PTY test suites human-observable at normal interaction speed, pause
between scenarios for confirmation or feedback, and turn observer feedback into
editable Harnez issue drafts without weakening unattended test execution.

## Verification

Add PTY-level coverage for view-mode pacing, visible test boundaries,
continue/feedback prompts, title editing, `$EDITOR` integration, draft creation,
cancellation, and non-interactive mode. Run a representative Loom TUI suite in
view mode and verify a human can follow each scenario and file feedback safely.
