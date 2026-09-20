# 088 — Extend Loom key capture coverage and sane action defaults

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Enhancement
**Related**: None

---

## Problem & Motivation

Loom's key-capture system should be dependable for the full range of keys users
expect in terminal applications, while widgets and example apps should share
clear, predictable defaults for common actions. Existing choices are useful
starting points, but coverage and consistency need to be made explicit before
the input model grows further.

## Scope

Support capture and matching for letters (`a-z`, `A-Z`), digits, punctuation,
German umlauts, and modified keys using Shift, Ctrl, and Alt. Review and extend
common action defaults across Loom widgets and example apps, including:

- quit/close: `q`, `Q`, `F10`, `Ctrl-C`
- activate/toggle/select: `Space`, `Insert`
- undo/redo: `u`, `U`, `Ctrl-Z`, `Ctrl-R`
- edit/view: `F4`/`e`, `F3`/`v`
- navigation: arrows, Page Up, Page Down, with WASD only as an opt-in

Preserve intentional application-specific bindings, document conflicts and
platform/terminal limitations, and keep the resulting defaults sane and
consistent.

## Goal

/goal: Make Loom reliably capture and distinguish the specified printable,
international, function, navigation, and modified keys, and establish tested,
documented default bindings for common actions across widgets and example apps
without breaking existing intentional choices.

## Verification

Add focused coverage for key decoding/matching and representative widget/app
bindings, including modifier combinations, umlauts, punctuation, and conflict
handling. Verify the examples remain usable with their documented defaults.
