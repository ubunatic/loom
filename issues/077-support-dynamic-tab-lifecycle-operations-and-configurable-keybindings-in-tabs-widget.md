# 077 — Support dynamic tab lifecycle operations and configurable keybindings in Tabs widget

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `docs/studies/2026-09-20-feature-gap-analysis-tabs.md`, `tabs.go`, `examples/tabs/tabs/tabs.go`

---

## 1. Problem & Motivation

`loom.Tabs` is currently configured with a static slice of tabs at instantiation. It lacks runtime mutation APIs (adding, removing, or reordering tabs) and only supports a single hardcoded switch key rather than declarative key maps (such as Next/Previous arrows or 1-9 direct index jumps).

## 2. Desired Behavior & Goal

`/goal`: Add dynamic lifecycle mutation methods (`Add`, `Remove`, `Insert`, `SetTabs`) and declarative navigation key configuration (`TabsKeys`) to `loom.Tabs`.

- Allow adding and removing tabs at runtime while safely clamping and preserving active tab index and focus.
- Provide declarative `TabsKeys` supporting customizable Previous/Next shortcuts and direct numerical navigation (e.g. `Alt-1`..`Alt-9` or `1`..`9`).
- Maintain full backwards compatibility with existing `SwitchKey` usage.

## 3. Implementation Plan

1. Add lifecycle methods (`Add(Tab)`, `Insert(int, Tab)`, `Remove(int)`, `SetTabs(...Tab)`, `Select(int) bool`) to `Tabs`.
2. Add `TabsKeys` struct for declarative key navigation and wire into `Tabs.HandleKey`.
3. Add unit tests covering dynamic addition/removal, boundary index handling, and key navigation.
