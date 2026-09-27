# 091 — Add double-click interaction for filebrowser and path trees

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Enhancement
**Related**: None

---

## Problem & Motivation

Loom needs a reusable double-click gesture so common hierarchical interfaces
behave naturally. In particular, selecting an item in the filebrowser or any
generic path-tree widget should make opening or entering that item immediate.

## Scope

- Add double-click recognition with clear timing, movement, and target rules.
- Expose the gesture through the generic Loom input/widget model so it can be
  reused by filebrowser and path-tree implementations.
- On a filebrowser or path-tree item, double-click should enter/open it: open a
  file, enter a directory, or perform the appropriate item activation action.
- Preserve single-click selection and existing keyboard activation behavior;
  avoid duplicate activation when a double-click follows the first click.
- Define behavior for empty space, disabled/non-openable items, and clicks on
  different targets.

## Goal

/goal: Provide reliable reusable double-click input and make double-clicking a
filebrowser or generic path-tree item open files, enter directories, or invoke
the item's normal activation behavior without regressing selection or keyboard
navigation.

## Verification

Add focused coverage for click timing/movement thresholds, target identity,
single-versus-double-click dispatch, and file/directory/disabled-item behavior.
Verify the filebrowser and any path-tree widget or example with mouse and
keyboard interaction.
Extend the PTY tests to send double-click input and prove files open and
directories are entered through a real terminal session.

## Plan (dev-091, reviewed by host)

- M1 recognition tests: pair clicks by same button and target, movement tolerance and timing threshold; movement, timeout, other target or button break the pair.
- M2 spec and recognizer: threshold in `spec/` plus schema; reusable recognizer with injectable clock (SGR has no click count).
- M3 Choice dispatch: first click selects without activating; recognized second click activates once. Keep single-click confirm default, Enter, `SelectOnlyOnClick`.
- M4 consumers: NavigationPane/filebrowser file open and dir entry; textedit outline jump.
- M5 PTY: timed press/release pairs in filebrowser PTY test.

Pre-Work: tests first per milestone; no double activation when Choice single-click confirm is enabled; threshold value must live only in spec (no Go duplicate).
