# 055 — Add a Tab panel widget for Pane hosting

**Status**: Closed — implemented in `tabs.go` (`Tabs`/`NewTabs`, `ThemeColors.TabsStyle()`), tested in `tabs_test.go`, demoed in `examples/tabs`.
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

## 2. Design Questions — resolved (Opus assessment, 2026-09-16)

An architectural review of `frame.go`, `stack.go`, `rows.go`, `yaml.go`,
`theme.go`, `spec/` established loom's actual spec/Go boundary is a
three-layer rule, not a binary "spec-driven or not":

1. `spec/*.yaml` (`box.yaml`, `themes.yaml`, `defaults.yaml`) holds only
   library-internal constants (glyphs, color roles, keys, timings) — never
   app composition.
2. YAML struct tags on a widget are reserved for **inert, non-interactive
   chrome** (`Frame`/`Box`/`Rows` — `HandleMouse`/`HandleKey` return false on
   these). Any widget with focus, behavior, or arbitrary children
   (`Stack`, `Grid`, `Popup`, `Choice`) is Go-only construction, optionally
   reachable from user layout YAML via a `type:` DTO in `yaml.go`
   (`compileWidget`) — a format deliberately separate from `spec/`, with no
   JSON Schema.
3. Theming applies universally but only via Go-wired
   `ThemeColors.*Style()` methods (e.g. `ChoiceStyle()`, `TableStyle()`) —
   nothing auto-themes.

`Tab` is interactive (key/mouse switching), focus-bearing, and holds
arbitrary `Widget` children — the same category as `Stack`/`Grid`/`Choice`.
Design questions resolved accordingly:

- **Location/spec surface**: `tabs.go` at the package root, **Stack-style
  pure Go composition** — no YAML struct tags on `Tab`. Copying `Frame`'s
  tagged-struct approach would reproduce its awkward half-declarative state
  for a widget whose entire value *is* its children.
- **API shape**: `NewTabs(...)` constructor over `Tab{Title, Widget}`
  entries, analogous to `NewStack`.
- **Key handling**: standard tab-switch keys with delegation to the focused
  child, per `Focusable` in `widget.go` — same focus-delegation pattern
  `Stack`/`Grid` already use.
- **Layout**: reserve tab-bar height via `ContentHeight`/`HeightForWidth`
  hooks (as `Stack` does), not a YAML `height` field.
- **Mouse support**: click-to-switch, consistent with `HandleMouse` on other
  widgets.
- **Styling**: add `ThemeColors.TabsStyle()` in `theme.go`, **reusing
  existing `header_*` (active tab), `normal_*` (inactive), and `border_*`
  (bar rule) roles** — no new `spec/themes.yaml` keys or schema churn across
  themes unless a real contrast failure is found. If a tab-bar separator is
  drawn, source its glyph from `spec/box.yaml` rather than hardcoding (see
  the `popup.go` shadowing exception noted below).
- **Dynamic add/remove**: out of scope for v1 — fixed, static tab set at
  construction, matching `Stack`'s fixed `Children` model.
- **Future declarative use**: leave the door open via `type: tabs` in
  `compileWidget`/`YamlElement` (additive, non-breaking) — not needed for
  v1, since `Stack` and other composite widgets follow the same deferred
  path.

**Exceptions/shadowing found during the review** (tracked separately, not
blocking this ticket): `grid.go`'s `NewGrid` hardcoded `ColorIndex(238)`
duplicating `themes.yaml`'s `plain.focus_bg` — **fixed** in this pass by
sourcing `Theme("plain").FocusBGColor()` instead. Still open: `popup.go`
hardcodes box-drawing glyphs instead of using `spec/box.yaml`;
`DefaultChoiceStyle()`/`DefaultTableStyle()` hardcode the "plain" theme,
pinned by `theme_test.go` rather than removed.

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
