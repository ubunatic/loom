# 150 — Register ansiviewer as a Linux desktop viewing tool

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature / Packaging
**Related**: [148](148-open-ansiviewer-directly-on-a-file-argument.md)

---

## 1. Problem & Motivation

Linux desktop file managers do not currently know that `ansiviewer` can view ANSI files, so users cannot select it as a viewing tool from the desktop.

## 2. Technical Specification / Findings

Add a `.desktop` entry that launches `ansiviewer` with the selected file, and install the entry as part of `make install`. Coordinate the file argument with [148](148-open-ansiviewer-directly-on-a-file-argument.md).

## 3. Implementation & Verification Plan

- **/goal**: Install a Linux desktop entry that lets desktop applications launch ansiviewer to view a selected ANSI file, or stop and report when blocked on a user decision or denied permission.
