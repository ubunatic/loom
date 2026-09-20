# 100 — Investigate full-width ANSI top-bar background in ansiviewer

**Status**: Open
**Priority**: P3
**Severity**: Minor
**Category**: Bug
**Related**: `examples/ansiviewer`, `docs/data/mc-julia256.ansi`, `docs/data/mc-mc46.ansi`

## Problem

The ANSI viewer reproduces the recorded mc panes and lower status rows, but the
top menu row does not always extend its recorded background through the full
terminal width. The new mc fixtures provide contrasting backgrounds for
investigation.

## Goal

Determine how the ANSI input establishes the top-bar background and make
ansiviewer reproduce that background across the complete top row, independently
of the Julia256 viewer theme.

## Verification

Add regression coverage using both mc ANSI fixtures and verify that the top row
background reaches the terminal's right edge without changing pane or status-row
colors.
