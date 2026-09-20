# Tabs Example Feature-Gap Analysis

Scope: `examples/tabs/tabs/tabs.go` compared with the current `loom` APIs.

## Findings

- **Tab header bar:** already a framework feature. `loom.NewTabs` and `loom.Tabs.Draw` in `tabs.go` render active/inactive titles, the separator rule, sizing, and mouse hit regions. The example only supplies `loom.Tab` values at `tabs.go:45`.
- **Tab switching:** already centralized for left/right arrows, mouse clicks, and the optional `SwitchKey` in `Tabs.HandleKey` / `Tabs.HandleMouse` (`tabs.go:157`, `tabs.go:184`). The example still chooses and assigns the shortcut manually at `tabs.go:50`.
- **Keyboard policy:** shortcut defaults are incomplete. `root.SwitchKey = "ctrl-t"` is app policy, while `Tabs` has no built-in previous/next key map, numeric tab selection, or configurable key bindings.
- **Dynamic tabs:** hard-coded in the example and unavailable in the framework. `NewTabs` receives a fixed variadic slice; `Tabs` documents dynamic add/remove as out of scope (`tabs.go:55`).
- **Pane/container setup:** the example manually creates and configures the runtime container (`loom.New(18)`, `Resizeable = true`, `EnableMouseClicks`, `Run`) at `tabs.go:53-60`. This is normal application startup, but a reusable app shell could combine pane creation, mouse setup, and a root widget.
- **Child layout:** `Tabs.Draw` already reserves the header/rule rows and delegates the remaining rectangle to the active child. No example-side pane layout duplication was found.

## Proposed framework opportunities

1. Extend `Tabs` with lifecycle operations that preserve focus safely:

   ```go
   func (t *Tabs) Add(tab Tab) int
   func (t *Tabs) Insert(index int, tab Tab) error
   func (t *Tabs) Remove(index int) error
   func (t *Tabs) SetTabs(tabs ...Tab)
   func (t *Tabs) Select(index int) bool
   ```

2. Make navigation policy declarative, with sensible defaults and optional numeric selection:

   ```go
   type TabsKeys struct {
       Previous, Next, Cycle string
       Select []string
   }
   func (t *Tabs) SetKeys(keys TabsKeys)
   ```

   `SwitchKey` could remain for compatibility or become `Keys.Cycle`.

3. Add a small pane/application builder for common examples:

   ```go
   func RunWidget(root Widget, opts ...PaneOption) error
   func WithHeight(height int) PaneOption
   func WithMouseClicks(enabled bool) PaneOption
   func WithResizeable(enabled bool) PaneOption
   ```

   This should compose existing `New`, `Resizeable`, `EnableMouseClicks`, and `Run`; it should not hide custom pane use cases.

## Priority

High: dynamic tab lifecycle and declarative key bindings. Medium: pane runner convenience. Low: header rendering and pane layout, already covered by `loom.Tabs`.
