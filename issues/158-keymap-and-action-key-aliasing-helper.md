# 158 — KeyMap and Action Key Aliasing Helper

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [loom-games](../loom-games)

---

## 1. Problem & Motivation

Interactive widgets often need to bind multiple key aliases to a single logical action (e.g. moving up via `"up"`, `"w"`, `"k"`; quitting via `"q"`, `"esc"`, `"ctrl-c"`). Checking raw key names with repetitive `e.Is("up", "w", "W", "k", "K")` creates boilerplate and is prone to missed aliases.

Furthermore, spec-driven architectures declare key bindings in YAML specs (e.g. `spec/actions.yaml`). Having a native `KeyMap` structure makes binding spec actions directly to event handlers seamless.

## 2. Technical Specification / Findings

- Add `KeyMap` helper:
  ```go
  type KeyMap struct {
      actions map[string][]string // action -> list of key triggers
  }

  func NewKeyMap(bindings map[string][]string) *KeyMap
  func (km *KeyMap) Action(e loom.KeyEvent) string
  func (km *KeyMap) Matches(e loom.KeyEvent, action string) bool
  ```
- Integrates case-insensitive key comparison and special key names.

## 3. Implementation & Verification Plan

### Goal
Provide a clean key-mapping and action-dispatching helper in `loom`.

### Acceptance Criteria
- [ ] Match multi-key aliases to logical action names.
- [ ] Case-insensitive ASCII matching for character keys.
- [ ] Unit tests covering single keys, aliases, modifiers, and unmatched events in `event_test.go`.

## Sprint goal (roadmap 180)
/goal Ship `KeyMap` with aliases and a short help label per action (used by 183), with event_test.go coverage; stop and report if existing widget key defaults (key_defaults.go) would have to change behavior.
