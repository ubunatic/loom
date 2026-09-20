# 086 — Fix post-refactor demo bugs in background split and treemap

**Status**: Closed — restored background panel mc-dark surfaces, wired split ratio key toggle, and added --ansi to treemap DemoArgs with tests
**Priority**: P1 (High)
**Severity**: Moderate
**Category**: Bug
**Related**: `examples/background/`, `examples/split/`, `internal/examplesreg/`, `cmd/loom-demo/`

---

## 1. Problem & Motivation

Following the example application modernizations in Issue 082, interactive testing via `loom-demo` uncovered three visual and behavioral bugs:

1. **`examples/background`**: The split panes lost their dark blue/purple tinted background surfaces, preventing the Astra starfield from demonstrating its surface contrast and color adaptation.
2. **`examples/split`**: Pressing the advertised ratio key `/` does nothing in the interactive split application.
3. **`treemap` in `loom-demo`**: The registry does not include `--ansi` in `DemoArgs` for `treemap`, resulting in uncolored output during interactive demo launches.

## 2. Desired Behavior & Goal

`/goal`: Fix the three post-refactor regressions across `background`, `split`, and `loom-demo` registry configurations.

- **`examples/background`**: Restore the tinted dark blue/purple background styling on split pane surfaces using Loom surface/styling features (`c.PaintSurface` or theme styling) so the Astra starfield contrast scaling is clearly visible.
- **`examples/split`**: Wire up `/` (and advertised ratio adjusting keys) to cycle or adjust the split ratio as intended.
- **`treemap` in `loom-demo`**: Add `--ansi` to `treemap`'s `DemoArgs` in `internal/examplesreg/registry.go` so `loom-demo` launches the treemap with rich ANSI colors.
