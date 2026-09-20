# 082 — Adopt new loom framework features across examples and retire obsolete code

**Status**: Closed — modernized all example applications with core Loom primitives across 3 sequential phases, verified by test suites and documented in study
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Refactor
**Related**: `docs/Widgets.md`, `docs/studies/`, `examples/`

---

## 1. Problem & Motivation

Several new high-level framework primitives have landed in Loom (`loom.Split`, dynamic `loom.Tabs`, `loom.MetricStore`, metric-bound `Gauge` / `Sparkline`, `loom.Directory` / `loom.OpenFile`, and `Pane.RunStartup`), but the reference applications in `examples/` still contain ad-hoc custom implementations, duplicated logic, or outdated patterns that predate these features.

## 2. Desired Behavior & Goal

`/goal`: Refactor all example applications in `examples/` to adopt newly introduced Loom framework features, eliminate obsolete custom boilerplate, capture lingering demo app issues, and publish adoption reports in `docs/studies/`.

- **Adoption Scope**:
  - `examples/split`: Adopt first-class `loom.Split` widget and nested focus container.
  - `examples/tabs`: Adopt dynamic `loom.Tabs` methods and declarative `TabsKeys`.
  - `examples/monitor`: Adopt `loom.MetricStore` and bound `Gauge`/`Sparkline` widgets.
  - `examples/filebrowser`: Fully leverage `loom.Directory`, `loom.OpenFile`, and path formatters.
  - `examples/splash`: Leverage `Pane.RunStartup` and `loom.RenderTo`.
  - Remaining examples (`screens`, `treemap`, `winch`, `background`): Update to use unified widget and layout patterns where applicable.
- **Hygiene & Verification**:
  - Remove all retired application-side boilerplate.
  - Maintain 100% green tests (`make test-q1`).
  - Document gaps/issues observed in each demo app.
  - Publish adoption summary studies in `docs/studies/`.

## 3. Implementation Plan

### M1 — Multi-Agent Example Audit & Gap Identification (Completed)
- Completed multi-agent audit across all reference applications in `examples/`.
- Detailed audit findings:
  1. `examples/split`: Missing `pane.EnableMouse()` (blocks divider drag); custom `scrollPane` hacks focus marker; static `"40 / 60"` Box title; lack of nested split and CLI `--help`.
  2. `examples/tabs`: Outdated `SwitchKey = "ctrl-t"` rather than declarative `loom.TabsKeys`; missing dynamic tab manipulation (`Add`, `Insert`, `Remove`); no key hints / status chrome; lacks CLI `--help`.
  3. `examples/splash`: Redundant boilerplate duplicating `RunStartup` defaults in `runWatch`; ignored `out` writer; static 1-line text handover target; repeated fallback arithmetic in `TerminalSize`.
  4. `examples/monitor`: Hand-rolled ring buffer (`boundedAppend`), slice cloning (`cloneSnapshot`), and manual snapshotting duplicating `loom.MetricStore`; manual string table mutations duplicating bound `loom.Gauge` / `loom.Sparkline`.
  5. `examples/filebrowser`: Redundant directory pre-validation; discarding structured `loom.FileEntry` causing redundant `os.Stat` roundtrips on activation; synthetic down-arrow key loop for selection restore; dual-box `Frame` vs first-class `loom.Split`.
  6. `examples/winch`: Low-level `c.Set` bypassing layered compositor; manual breakpoint layout math vs `loom.Split`; ad-hoc pane metric formatting.
  7. `examples/screens`: Duplicate `detected()` vs `p.AutoFullscreenReasons()`; manual boolean toggles vs `p.SetScreenMode()`; manual box borders vs framework styles.
  8. `examples/background`: Orphaned binary `assets/cati.png`; custom `widget` wrapping `Frame`; missing interactive controls (motion/theme toggle); 0% test coverage.

### M2 — Sequential Developer Refactor & Code Retirement
Execute code refactoring sequentially across the example suites:
- **Phase 1: `split`, `tabs`, and `splash`**
  - `examples/split`: Enable mouse support (`pane.EnableMouse()`), remove `scrollPane` hack, dynamic ratio status, nested split demo, add Cobra `--help` (`SupportsHelp: true`).
  - `examples/tabs`: Adopt declarative `loom.TabsKeys`, add interactive tab dynamic actions (`+`/`a` add, `x`/`d` close), add status bar key legend, add Cobra `--help` (`SupportsHelp: true`).
  - `examples/splash`: Rely on `RunStartup` defaults, clean up size fallbacks, enhance destination view to rich widget, add `runWatch` test coverage.
- **Phase 2: `monitor`, `screens`, and `treemap`**
  - `examples/monitor`: Adopt `loom.MetricStore` in `state.go`, delete `boundedAppend`/`cloneSnapshot`, bind `loom.Gauge`/`loom.Sparkline`, use `loom.TerminalSize()`.
  - `examples/screens`: Adopt `p.AutoFullscreenReasons()`, `p.SetScreenMode()`, clean border helpers.
  - `examples/treemap`: Clean up `terminal.go` with `loom.TerminalSize()`.
- **Phase 3: `filebrowser`, `winch`, and `background`**
  - `examples/filebrowser`: Eliminate duplicate directory validation, preserve structured `FileEntry` without redundant `os.Stat`, clean selection restore, adopt `loom.Split`.
  - `examples/winch`: Adopt layered compositor (`PaintForeground`/`PaintSurface`) in `drawBoxBorder`, streamline metrics formatting.
  - `examples/background`: Clarify doc/asset usage, add theme & motion key toggles (`m`/`t`), adopt `loom.Split`, add unit test suite in `background_test.go`.

### M3 — Verification & Study Documentation
- Run full test suite (`make test-q1` / `go test ./...`) and PTY checks.
- Commit author adoption and findings study in `docs/studies/2026-09-20-examples-modernization-and-framework-primitives-audit.md`.
