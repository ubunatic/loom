# Study: Examples Modernization & Framework Primitives Adoption

**Date**: 2026-09-20  
**Status**: Complete  
**Related Tickets**: Issue 082, Issue 076, Issue 077, Issue 078, Issue 079, Issue 080

---

## 1. Context & Motivation

As Loom evolved from early static layouts to a mature TUI toolkit featuring reactive metric stores, dynamic tabs, layered compositing, and nested split view containers, reference applications in `examples/` retained varying amounts of legacy boilerplate and pre-feature workarounds.

Issue 082 established a structured audit and sequential refactoring effort to modernize all example applications against current core framework primitives.

---

## 2. Refactored Examples & Adopted Primitives

| Example Application | Prior State / Workaround | Adopted Modern Primitives | Impact / Benefits |
|---|---|---|---|
| `examples/split` | Custom `scrollPane` hack for focus marker; mouse disabled; static box title; flat 2-pane view. | `pane.EnableMouse()`, `loom.View.FocusStyle`, nested `loom.Split` (horizontal outer + vertical inner), dynamic ratio calculation, Cobra `--help`. | Full divider dragging, multi-branch `FocusContainer` `Tab`/`Shift-Tab` navigation, clean non-interactive `--help`. |
| `examples/tabs` | Deprecated `SwitchKey = "ctrl-t"`; static tabs with no dynamic lifecycle; exit swallowed by child tables. | Declarative `loom.TabsKeys` (directional, cycle, and numeric `1`..`9` jumps), dynamic `Add`/`Remove` key actions, status footer chrome, Cobra `--help`. | Full interactive tab mutation demonstration and clean top-level quit bindings. |
| `examples/splash` | Manual `SplashView` & `Cadence` re-implementation in `runWatch`; 1-line text handover; complex fallback math. | `Pane.RunStartup` built-in defaults, `resolveTerminalDimensions` helper, rich interactive `Frame`/`Choice` handover target, cancellation tests. | Concise startup orchestration and verified context-safe cancellation. |
| `examples/monitor` | Custom circular buffers (`boundedAppend`), slice copying (`cloneSnapshot`), and manual snapshotting. | `loom.MetricStore` with `loom.Retention(32)`, bound `loom.Gauge` and `loom.Sparkline` widgets, `loom.TerminalSize()`. | Removed redundant buffer management; unified metric publication and decoupled render pipeline. |
| `examples/screens` | Hand-rolled fullscreen reason detectors (`App.detected()`), manual mode flags, and custom ASCII borders. | `pane.AutoFullscreenReasons()`, `pane.SetScreenMode(...)`, framework `BoxBorder`/`BoxStyle` / `c.PaintForeground`, reusable `sizeStepper`. | Eliminated duplicated detection algorithms; clean state synchronization with `loom.Pane`. |
| `examples/treemap` | Ad-hoc dimension clamping and raw terminal queries. | `loom.TerminalSize()`, streamlined dimension clamping. | Standardized terminal dimension detection across execution modes. |
| `examples/filebrowser` | Duplicate directory validation; string map conversion discarding `loom.FileEntry`; repeated `os.Stat` on select. | Retained structured `loom.FileEntry`, single-pass symlink statting, adopted `loom.Split`, simplified cursor restore. | Reduced filesystem syscalls; unified dual-pane split container. |
| `examples/winch` | Low-level `c.Set` calls bypassing the layered compositor for borders. | `c.PaintForeground` for border glyphs, `c.PaintSurface` for panel surfaces, streamlined metrics formatting. | Full compliance with layered compositor and Astra starfield inheritance. |
| `examples/background` | Static un-interactive frame; 0% test coverage; conflicting doc comments. | Interactive `m`/`M` motion toggle, `t`/`T` theme cycling, `loom.Split` divider drag over Astra starfields, comprehensive unit test suite in `background_test.go`. | Validated starfield rendering under active theme scaling and ratio changes with automated regression coverage. |

---

## 3. Verification Matrix

- **Unit & Example Tests**: All example test suites (`go test ./examples/...`) pass cleanly.
- **PTY Regression Suites**: Screen mode transitions and resize diagnostics pass (`examples/screens` PTY tests).
- **Code Coverage**: Added automated tests for previously untested packages (`examples/background/background_test.go`, `examples/split/split_test.go`, `examples/tabs/tabs_test.go`, `examples/splash/splash_test.go`).
