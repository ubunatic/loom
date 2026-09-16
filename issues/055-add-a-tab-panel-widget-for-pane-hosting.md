# 055 — Add a Tab panel widget for Pane hosting

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: `widget.go`, `pane.go`, `stack.go`, `frame.go`,
`examples/filebrowser`, `examples/split`

---

## 1. Problem & Motivation

Loom composes UIs from `Widget` implementations (`Choice`, `Table`,
`TextInput`, `Frame`, `Stack`, etc.) that a `Pane` hosts and drives via
`Pane.Run`/`loom.RunPane`. There is currently no built-in widget that
switches between multiple child widgets under a row of tabs — an
application that wants a tabbed layout (e.g. several `examples/filebrowser`-
style panels, or grouping unrelated views) must hand-roll tab-bar rendering,
key handling for switching tabs, and delegation to the active child's
`Draw`/`HandleKey`/`HandleMouse`.

Add a `Tab` panel widget so applications can host multiple child widgets
under one `Pane`, switch between them (e.g. left/right arrow, a configured
switch key, or direct number/name selection), and render a tab bar showing
titles and the active tab, without each application reimplementing this.

## 2. Design Questions

- Where does this live: a new `tabs.go` at the package root, following the
  existing `stack.go`/`frame.go` pattern (spec-driven where applicable, per
  `docs/Spec.md`), or is it purely programmatic (like `Stack`) with no YAML
  spec surface?
- API shape: a `NewTabs(children ...Widget, titles ...string)`-style
  constructor analogous to `NewStack`, or a `Tab` struct type holding
  `{Title string; Widget Widget}` entries?
- Key handling: what switches tabs (`Tab`/`Shift+Tab`, left/right arrows, a
  configurable key set), and how does that interact with a child widget that
  itself consumes those same keys (focus delegation — see `Focusable` in
  `widget.go`)?
- Layout: how much vertical/horizontal space does the tab bar itself
  reserve, and does `Tab` implement `HeightForWidth`/`ContentHeight`-style
  sizing hooks the way `Frame` boxes do (see #051, fill-height design)?
- Mouse support: should clicking a tab title switch tabs, consistent with
  `HandleMouse` on other widgets?
- Styling: does `Tab` need its own theme role(s) in `spec/themes.yaml`
  (active/inactive tab colors), or does it reuse existing header/border
  roles?
- Does closing/hiding individual tabs matter for v1, or is a fixed, static
  set of tabs (set at construction) sufficient, deferring dynamic add/remove
  to a follow-up?

## 3. Verification & Acceptance

- A new `Tab` (or similarly named) widget implements the `Widget` interface
  and can be passed directly to `Pane.Run`/`loom.RunPane`, hosting two or
  more child widgets.
- Switching tabs via keyboard (and mouse, if implemented) changes which
  child widget receives `Draw`/`HandleKey`/`HandleMouse` and is visibly
  reflected in the rendered tab bar.
- Unit tests cover: tab switching, delegation of key/mouse events to the
  active child only, rendering of the tab bar (active vs. inactive titles),
  and any edge case around a single-tab or zero-tab panel.
- At least one example (new or extended, e.g. `examples/filebrowser`)
  demonstrates the `Tab` widget hosting multiple child widgets.
- `go test ./...` and `go vet ./...` pass.
