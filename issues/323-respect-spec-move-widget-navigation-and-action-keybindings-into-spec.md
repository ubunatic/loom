# 323 — respect spec: move widget navigation and action keybindings into spec

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: `docs/Spec.md`, `spec/defaults.yaml`, `spec/schemas/defaults.schema.json`, `defaults.go`, `tabs.go`, `view.go`, `split.go`, `key_defaults.go`

---

## 1. Problem & Motivation

Under `docs/Spec.md` and the `/respect` practice, default keybindings and aliases that describe standard application behavior belong in the embedded YAML specification (`spec/defaults.yaml`) rather than hardcoded in Go switch-cases or package helper functions.

Currently, default navigation key sets (such as Tabs default keys, View navigation keys, and Split focus cycling) are either hardcoded in constructors (`DefaultTabsKeys()`) or embedded directly inside widget `ConsumeKey` dispatch switches. `spec/defaults.yaml` provides global `quit_keys`, but lacks structured widget navigation bindings.

## 2. Technical Specification / Candidates

### Candidate Values

| Location (file:line) | Value | Target Spec File and Key |
|---|---|---|
| `tabs.go:40-42` | `TabsKeys{Previous: "left", Next: "right"}` | `spec/defaults.yaml` -> `tabs.keys.previous`, `tabs.keys.next` |
| `split.go:149-162` | Next: `"tab"`, Previous: `"shift-tab"`, Ratio adjust: `"[", "]"` | `spec/defaults.yaml` -> `split.keys.next`, `split.keys.previous`, `split.keys.shrink`, `split.keys.grow` |
| `view.go:194-234` | Scroll keys: `"up", "k"`, `"down", "j"`, `"pgdown", "pgdn", "pagedown", "ctrl-f", " "`, `"pgup", "pageup", "ctrl-b", "b"`, `"home", "g"`, `"end", "G"`, `"esc", "ctrl-c", "q"` | `spec/defaults.yaml` -> `view.keys.*` |
| `key_defaults.go:8-20` | `libraryKeyDefaults` table documenting default key conventions | `spec/defaults.yaml` -> `key_conventions` or generate from spec |

### What Stays in Code and Why

- Low-level key event parsing and ANSI escape code translation (`event.go`) stay in Go because they implement terminal protocol specifications, not application-level configurable defaults.
- Dynamic key dispatch branches and widget-specific state mutation (`ConsumeKey`) stay in Go.
- Opt-in key filters and custom callbacks registered by caller applications remain in Go.

## 3. Acceptance Criteria

- `spec/defaults.yaml` defines default key bindings for standard widgets (`tabs`, `split`, `view`), matching `spec/schemas/defaults.schema.json`.
- Go constructors (such as `DefaultTabsKeys`) and widget dispatch logic read directly from `SpeccedDefaults`.
- No mirrored hardcoded fallback tables exist in Go code.
- Unit tests assert all defined keybinding fields exist, are unmarshaled properly from the embedded YAML, and drive key handling behavior.
