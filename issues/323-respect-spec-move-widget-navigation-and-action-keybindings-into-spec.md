# 323 — respect spec: move widget navigation and action keybindings into spec

**Status**: Closed — Bindings of Tabs, Split, View, Table, Choice, Tree, Menu, Form, Settings and the key-convention table now come from spec/defaults.yaml + schema (1e48bcb, 9550a25, 0758526, 333e0be); YAML-change probe failed before and passes; no default key changed (live View j/q same as pre-change binary); host q1 green, validate-spec ok
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: `docs/Spec.md`, `spec/defaults.yaml`, `spec/schemas/defaults.schema.json`, `defaults.go`, `tabs.go`, `view.go`, `split.go`, `key_defaults.go`

---

/goal Default key bindings of the core widgets come from the embedded spec and drive key handling, with no
mirrored Go table, and the help/key-convention text is generated from the same data; tests prove a YAML
change alters behaviour; or stop and report when blocked on a user decision.

## Update (spec audit, 2026-10-22)

- Also covers `table.go:133` (navigation and paging) and the bindings of Choice, Tree, Menu, Form and
  Settings; check live code for each.
- Decide in a plan-only first turn whether bindings fit `spec/defaults.yaml` or need a new `spec/keys.yaml`
  (per-widget bindings with action IDs); the host reviews that before writes.
- The media widget's bindings are ticket 367.
- Key aliases (`pgdn`, `pagedown`, `pageup`) remain accepted.

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

## Plan (d323, approved by host 2026-10-22)

Bindings go into `spec/defaults.yaml` + schema, loaded into `SpeccedDefaults` with validation (missing
action panics at init); widgets read via `KeyMap` or `key.Is` on specced slices; `libraryKeyDefaults` is
built from spec `key_conventions`, checked against docs/KeyDefaults.md. Milestones: M1 spec+schema,
M2 loader + regression probe, M3 widget migration, M4 API snapshot, docs, q1.

Pre-Work / Required Refinements:
- Do not commit a failing test: show the probe failing on unfixed code, then commit it with M3.
- Existing aliases and current default keys stay identical (no behaviour change); tests cover key and
  mouse paths unaffected.
- Upgrading.md row only if an exported symbol changes; additions just go into the API snapshot.
