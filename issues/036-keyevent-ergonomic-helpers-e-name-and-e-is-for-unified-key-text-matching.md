# 036 — KeyEvent Ergonomic Helpers (e.Name and e.Is) for Unified Key/Text Matching

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [event.go](file:///home/uwe/projects/loom/event.go), [widget.go](file:///home/uwe/projects/loom/widget.go)

---

## 1. Problem & Motivation
`loom.KeyEvent` separates decoded input into two distinct fields:
- `Key string`: Set for special/control keys (`"up"`, `"down"`, `"enter"`, `"pgup"`, `"esc"`, `"ctrl-c"`, etc.), but empty `""` for regular printable characters.
- `Text string`: Set for printable keystrokes (`"j"`, `"k"`, `"?"`, `"t"`, `"q"`, `" "`, etc.), while `Key` is empty.

In practice, widget developers frequently implement `HandleKey` using `switch e.Key { case "up": ... case "j": ... case "q": ... }`. Because printable characters have `e.Key == ""`, cases matching `"j"`, `"q"`, `"?"`, or `"t"` are silently bypassed unless the developer explicitly writes boilerplate:
```go
key := e.Key
if key == "" {
    key = e.Text
}
```

## 2. Proposed Solution
Add ergonomic helper methods on `KeyEvent`:
1. `func (e KeyEvent) Name() string`:
   Returns `e.Key` if non-empty, otherwise returns `e.Text`.
2. `func (e KeyEvent) Is(keys ...string) bool`:
   Returns `true` if `e.Name()` matches any of the provided key identifiers (e.g. `if e.Is("q", "esc", "ctrl-c") { return true }`).
3. `func (e KeyEvent) Rune() rune`:
   Returns the first rune of `e.Text`, or `0` if not a printable character.

## 3. Verification & Acceptance
- Unit tests in `event_test.go` verifying `e.Name()` and `e.Is()` across both special keys (`"up"`, `"pgup"`) and printable characters (`"?"`, `"t"`, `"j"`).
