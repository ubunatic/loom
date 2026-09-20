# 079 — Extract reusable Directory model FileBrowser primitives and platform file opener

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `docs/studies/2026-09-20-feature-gap-analysis-filebrowser.md`, `examples/filebrowser/filebrowser/browser.go`

---

## 1. Problem & Motivation

The filebrowser example contains valuable, generic filesystem browsing logic—such as directory entry mapping, unprintable path character escaping, metadata inspection, and detached platform file opening (e.g. `xdg-open`)—that is currently locked inside example code.

## 2. Desired Behavior & Goal

`/goal`: Extract reusable `Directory` models, terminal-safe path formatters, and an injectable platform `OpenFile` helper into reusable Loom primitives.

- Provide `loom.ReadDirectory(path, opts)` returning structured entries with parent/child navigation support.
- Provide `loom.DisplayPath` / `loom.QuoteUnprintable` for terminal-safe width and control-character escaping.
- Provide `loom.OpenFile(path)` with platform detection and mockable execution for testing.
- Enable file browser apps to be composed with minimal boilerplate.

## 3. Implementation Plan

1. Implement filesystem helpers in `fs_model.go` (or `fileutil.go`).
2. Add unit tests for unprintable escaping, symlink handling, and path resolution.
3. Update `examples/filebrowser` to use the extracted framework helpers.
