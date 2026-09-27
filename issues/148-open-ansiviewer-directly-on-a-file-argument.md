# 148 — Open ansiviewer directly on a file argument

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [097](097-add-ansi-viewer-example-with-tui-recording.md)

---

## 1. Problem & Motivation

`ansiviewer` currently accepts directories as startup arguments. Users who provide a file path cannot open that file directly in the viewer.

## 2. Technical Specification / Findings

When the argument is a file, start in its containing directory, select that file, and show its preview immediately. Preserve the existing directory-argument behavior.

## 3. Implementation & Verification Plan

- **/goal**: Allow ansiviewer to start with either a directory or a file, showing file arguments immediately in their parent directory, or stop and report when blocked on a user decision or denied permission.
