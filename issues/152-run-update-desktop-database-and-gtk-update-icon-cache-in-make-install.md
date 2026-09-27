# 152 — Run update-desktop-database and gtk-update-icon-cache in make install

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature / Packaging
**Related**: [150](150-register-ansiviewer-as-a-linux-desktop-viewing-tool.md), [151](151-add-and-install-desktop-icon-for-ansiviewer.md)

---

## 1. Problem & Motivation

After installing `ansiviewer.desktop` and `ansiviewer.svg` via `make install`, Linux desktop environments (GNOME, KDE, XFCE, etc.) and file managers may not recognize new MIME associations or display updated application icons immediately without refreshing the user desktop application database and icon caches.

## 2. Technical Specification / Findings

In `Makefile`, add guarded update commands to the `install` target:
- `update-desktop-database "$(ANSIVIEWER_DESKTOP_DIR)" 2>/dev/null || true` (if command exists)
- `gtk-update-icon-cache -q -t "$(HOME)/.local/share/icons/hicolor" 2>/dev/null || true` (if command exists)
Ensure the commands fail silently and cleanly in headless or non-GUI environments.

## 3. Implementation & Verification Plan

- **/goal**: Automatically update desktop database and icon cache during `make install` when the tools are present, or stop and report when blocked on a user decision or denied permission.
