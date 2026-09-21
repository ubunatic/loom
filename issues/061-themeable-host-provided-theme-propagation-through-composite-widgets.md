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

---

## Milestones (lean sprint, dev agent: haiku)

The host reviews only diffs and test output; this ticket is the only channel. Root-package
tests needing `/dev/tty` fail before this work; ignore them. Commit each milestone
(message ends '(issue 061 MX)'), staging only your own files, never docs/README.md.

Evidence rules (learned in 034): frames are produced by code, never hand-written; the
evidence test writes only when env `LOOM_EVIDENCE=1` is set; it finds the repo root by
walking up to `go.mod` and writes to repo-root `docs/progress/061/`. Run it with
`LOOM_EVIDENCE=1` and check `git status` shows only the intended files.

### M1 - Themeable interface and composite forwarding
- Add `Themeable` (section 2) to `widget.go`; `Tabs`, `Stack`, `Frame`, `Grid` implement it,
  restyle their own chrome and forward to children that implement it; other children are skipped.
- Tests: nested `Tabs{Frame{Choice}}` rendered under two themes from `spec/themes.yaml`
  differs in tab bar, border and choice; a non-Themeable child does not error.
  Read the theme list from the spec API, never hardcode color values in Go.
- Evidence: `M1-nested-themes.ansi`, the same tree rendered once per theme, with a label
  line before each.

### M2 - Choice/Table defaults and filebrowser adoption
- `Choice` and `Table` implement `Themeable`; `DefaultChoiceStyle()`/`DefaultTableStyle()` stay
  `plain` and are overridable through `ApplyTheme`. Update `theme_test.go` accordingly.
- filebrowser: export `ApplyTheme(loom.ThemeColors)` (replacing `applyTheme`); standalone
  `--theme` and F9 behave exactly as before (existing tests stay green).
- Evidence: `M2-filebrowser-themes.ansi`, filebrowser rendered headless under two themes.

### M1-M2 Review (host)
Delivered: 69443ea (Themeable and composites), e9c0370 (Choice/Table, filebrowser). Tests and vet
green, evidence gated and rooted correctly. Gaps against the ticket:
- `M2-filebrowser-themes.ansi` shows a filebrowser-like Frame+Choice, not the real filebrowser.
- Nothing tests `browser.ApplyTheme` itself. It resolves the name by comparing ThemeColors
  in a loop and falls back to a fake name "custom"; that name leaks into F9 cycling.

### M3 - Pre-Work / Required Refinements
1. Add a filebrowser test (package filebrowser, `browser_test.go`): build the real browser on a
   temp dir with a few files, render it headless under two themes via `ApplyTheme`, assert
   the output differs and that a following F9 keypress still cycles to a valid named theme
   (never "custom"). If F9 would produce "custom", make `ApplyTheme` keep the current
   name when no exact match is found instead of inventing one.
2. Regenerate `M2-filebrowser-themes.ansi` (LOOM_EVIDENCE=1) from that real browser,
   one frame per theme with a label line.
3. Fix comments: 'restyled' should read 'restyles' in tabs.go and frame.go.
4. Commit with '(issue 061 M3)'; tests and vet green.
