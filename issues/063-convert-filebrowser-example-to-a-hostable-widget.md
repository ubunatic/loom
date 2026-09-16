# 063 — Convert filebrowser example to a hostable widget

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `examples/filebrowser/filebrowser/filebrowser.go`,
`examples/filebrowser/filebrowser/browser.go`,
`internal/examplesreg/registry.go`, `cmd/loom-demo`,
[055](055-add-a-tab-panel-widget-for-pane-hosting.md),
[057](057-hosted-widget-key-contract-child-first-routing-reserved-host-keybinds-and-quit-containment.md),
[058](058-widget-declared-pane-requirements-panerequest.md),
[059](059-fix-mouse-coordinate-convention-mismatch-between-frame-and-tabs-stack-grid.md),
[061](061-themeable-host-provided-theme-propagation-through-composite-widgets.md),
[062](062-example-widget-factories-newwidget-for-split-and-tabs-hosted-loom-demo-mode-headless-bench-smoke.md)

---

## 1. Problem & Motivation

filebrowser is the second conversion wave of the "examples as just widgets"
initiative, and the one that actually exercises every library addition made in
057/058/059/061 at once:

- **Factory already exists**: `newBrowser(dir, themeName, theme) (*browser, error)`
  (`browser.go:36`), and `Run` is already *parse args → build widget → open
  pane → configure pane → run* (`filebrowser.go:16-42`). The mechanical part is
  nearly free.
- **Quit/`q` conflict**: it sets `DisableDefaultQuit = true` because `q` is its
  filter key (`filebrowser.go:59-65`). Under a shared host pane this is
  all-or-nothing, which is exactly what 057's `KeyConsumer` retires.
- **Theming**: it owns `--theme` (`filebrowser.go:18-25,44-51`), applies it via
  `applyTheme` (`browser.go:72-84`) and cycles at runtime with F9
  (`browser.go:85-95,246-253`) — the concrete consumer for 061's `Themeable`.
- **Mouse**: it calls `EnableMouseClicks` on its own pane (`filebrowser.go:40`)
  and is `Frame`-based, so it is also the real-world check on 059's coordinate
  fix.
- **Pane knobs**: `MaxCols = 0` to defeat the 50-column default
  (`filebrowser.go:64`), `Resizeable`, and an initial height — all moving into
  058's `PaneRequest`.

It is deliberately converted *after* split/tabs so the API is already in place
and this ticket validates it rather than designing it.

## 2. Design — resolved decisions

- Export `NewWidget(args []string) (loom.Widget, error)` wrapping the existing
  `newBrowser`, keeping the `flag.FlagSet` parsing (`filebrowser.go:17-29`) in
  the standalone `Run` path and accepting the same args in the factory.
- `browser` implements `KeyConsumer`: `q` (and its other text-entry keys while
  filtering) report `consumed = true`, `ctrl-q`/`F10` report `quit = true`.
  Delete the `DisableDefaultQuit` pane call — the widget now declares its own
  policy and a host can contain the quit via `Tabs.OnChildQuit`.
- `browser` implements `PaneRequester`: `Mouse: 1000`, `Resizeable: true`,
  `MaxCols: 0`. Delete the corresponding imperative pane calls from `Run`.
- Rename `applyTheme` to the exported `ApplyTheme(loom.ThemeColors)`
  (`Themeable`). Standalone `--theme` resolves a name and calls it; F9 keeps
  cycling, but when hosted it retherms the host chrome too (061).
- Register `NewWidget` in `examplesreg` and add filebrowser to loom-demo's
  hosted tab set.

## 3. Verification & Acceptance

- `filebrowser.NewWidget(args)` returns a widget usable by any host; `Run` is a
  thin wrapper with no `DisableDefaultQuit`/`EnableMouseClicks`/`MaxCols`
  calls left in it.
- Standalone behavior is byte-for-byte unchanged: `q` filters, `F9` cycles
  themes, `F10`/`ctrl-q` quits, mouse clicks select — verified by the existing
  filebrowser tests plus a manual/PTY pass.
- Hosted in loom-demo: `q` does **not** quit the host, `ctrl-q` closes only the
  filebrowser tab, mouse clicks land on the correct row inside its `Frame`
  boxes (the 059 regression case), and F9 rethemes the tab bar as well.
- `loom-bench` renders it headlessly via `loom.Render` at 80x24 and tiny sizes.
- `go test ./...` and `go vet ./...` pass; `make install` run afterwards.
