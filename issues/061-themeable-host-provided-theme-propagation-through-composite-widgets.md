# 061 — Themeable: host-provided theme propagation through composite widgets

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: `widget.go`, `theme.go`, `tabs.go`, `stack.go`, `frame.go`,
`grid.go`, `spec/themes.yaml`,
`examples/filebrowser/filebrowser/browser.go`,
[055](055-add-a-tab-panel-widget-for-pane-hosting.md),
[063](063-convert-filebrowser-example-to-a-hostable-widget.md)

---

## 1. Problem & Motivation

Themes come from `spec/themes.yaml` via `SpeccedThemes`/`Theme(name)`
(`theme.go:146-166`), and — as 055 §2 records deliberately — **nothing
auto-themes**: every widget's style is baked in at construction.

That is fine while each app owns its pane, but it breaks the hosted model:

- filebrowser resolves `--theme` (`filebrowser.go:18-25,44-51`), applies it via
  `browser.applyTheme` (`browser.go:72-84`) and cycles themes at runtime with
  F9 (`browser.go:85-95`, `HandleKey` at `:246-253`). Hosted, F9 would retheme
  only its own subtree while the host's tab bar stays `plain`.
- The `tabs` example is themed implicitly: `NewTabs` → `DefaultTabsStyle()` →
  `Theme("plain")` (`tabs.go:27-29,74`).
- `split`, `monitor`, `splash` use no `ThemeColors` at all; `treemap` carries
  its own ANSI palette (`examples/treemap/treemap/style.go`).

A host has no way to impose a consistent look on widgets it did not construct.

## 2. Design — resolved decisions

Optional interface, forwarded by composites — the same shape as `Focusable`
and the other optional widget interfaces:

```go
// Themeable is an optional interface for widgets that can restyle themselves
// from a ThemeColors set. Composite widgets forward ApplyTheme to children.
type Themeable interface {
    Widget
    ApplyTheme(ThemeColors)
}
```

- `Tabs`/`Stack`/`Frame`/`Grid` implement it by restyling their own chrome and
  forwarding to every child that implements it.
- The host (or `Pane`, for a root that implements it) calls `ApplyTheme` once
  after construction, and again on any runtime theme change.
- filebrowser adopts it nearly for free — `applyTheme(ThemeColors)` already
  exists with the right signature shape (`browser.go:72`); it becomes the
  exported `ApplyTheme` and F9 in a host retherms the *whole* UI including the
  tab bar, which is a good demonstration of the pattern.
- Standalone `Run` wrappers keep `--theme`: they resolve name → `ThemeColors`
  and call `ApplyTheme` on the widget they just built. The hosted path simply
  inherits the host's theme.
- This does **not** introduce auto-theming: nothing themes itself, a caller
  always drives it. 055's three-layer spec/Go boundary is preserved —
  `spec/themes.yaml` stays the source of color roles, Go stays the wiring.

Rejected: passing `ThemeColors` as a constructor parameter to every
`NewWidget(...)`. Simpler on paper, but it loses runtime retheming (filebrowser
already ships F9) and forces the host to thread a theme into six constructors
with no way to change it later.

Also in scope (small, related debt surfaced by 055's review):
`DefaultChoiceStyle()`/`DefaultTableStyle()` hardcode the "plain" theme, pinned
by `theme_test.go`. With `Themeable` in place these can default to `plain` and
be overridden by an `ApplyTheme` call rather than needing to be "removed".

## 3. Verification & Acceptance

- `Themeable` in `widget.go`; `Tabs`/`Stack`/`Frame`/`Grid` forward it.
- Applying a theme to a nested tree (`Tabs{Frame{Choice}}`) restyles the tab
  bar, the frame borders and the choice — asserted by rendering with two
  different themes from `spec/themes.yaml` and diffing the output.
- Widgets that do not implement `Themeable` are skipped without error.
- filebrowser exposes `ApplyTheme` and its F9 cycle goes through the host when
  hosted (covered in [063](063-convert-filebrowser-example-to-a-hostable-widget.md)).
- Standalone `--theme` behavior is unchanged.
- `go test ./...` and `go vet ./...` pass.
