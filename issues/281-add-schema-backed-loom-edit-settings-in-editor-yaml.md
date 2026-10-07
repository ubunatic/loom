# 281 — Add schema-backed loom edit settings in editor.yaml

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [280](280-add-mouse-capture-flag-to-loom-edit.md), [settings design](../docs/design/loom-edit-05-settings.ansi)

---

## 1. Problem & Motivation
`loom edit` needs persistent preferences in `~/.config/loom/editor.yaml`. The first settings are `theme`, `mousegrab` on/off, and `altscreen` on/off, with a schema for the YAML.

## 2. Technical Specification / Findings
Implementation constraint: the [approved design guidance](../docs/design/loom-edit-design-notes.md) allows the final app to differ with standard Loom widget behavior. New UI elements must use standard Loom widgets; small tweaks are acceptable, heavy hacks are not. Always file or link a library issue for larger observed Loom gaps instead of adding app workarounds.

Load editor preferences from the requested path (honor XDG config location conventions). Provide a shipped JSON Schema and a YAML schema association/example. Use existing theme names and YAML booleans for the toggles. Explicit CLI flags override file settings; a missing file uses documented spec defaults. Validate malformed YAML, unknown settings, wrong types and invalid themes with actionable diagnostics. Keep defaults in the spec. The design's adjacent `editor.schema.json` association is illustrative; document the actual installed schema location.

## 3. Implementation & Verification Plan
/goal `loom edit` loads schema-backed theme, mousegrab and altscreen preferences, with documented defaults and CLI precedence; verify loading and terminal behavior, or stop and report when blocked on a user decision or denied permission.

Before implementation, check live code and recent commits. Acceptance: valid settings apply; absent config works; invalid settings fail clearly; schema and runtime validation agree; explicit mouse flag wins, including false; both toggles work and terminal state is restored on exit. Verify with focused config/schema tests and PTY checks, then `make test-q1` and `make install`. Event-routing changes start on `codex:sol:med`.
