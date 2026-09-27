# 151 — Add and install desktop icon for ansiviewer

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature / Packaging
**Related**: [150](150-register-ansiviewer-as-a-linux-desktop-viewing-tool.md)

---

## 1. Problem & Motivation

The `ansiviewer.desktop` entry created in issue 150 does not have an associated desktop icon, so it shows a generic default system icon in Linux desktop application menus and file managers.

## 2. Technical Specification / Findings

1. Create/add an icon asset for `ansiviewer` (SVG or PNG, e.g. `examples/ansiviewer/assets/ansiviewer.svg` or `examples/ansiviewer/ansiviewer.png`).
2. Update `examples/ansiviewer/ansiviewer.desktop` to specify `Icon=ansiviewer`.
3. Update `Makefile` `install` target to install the icon into `~/.local/share/icons/hicolor/scalable/apps/ansiviewer.svg` (or `~/.local/share/icons/hicolor/128x128/apps/ansiviewer.png` / `~/.local/share/pixmaps/ansiviewer.svg`).
4. Validate with `desktop-file-validate` and verify `make install` copies both desktop entry and icon.

## 3. Implementation & Verification Plan

- **/goal**: Add an ansiviewer icon, declare it in ansiviewer.desktop, and install it via `make install`, or stop and report when blocked on a user decision or denied permission.
