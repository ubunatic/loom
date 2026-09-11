# 033 — Harnez Target Splash Screen Integration and Golden Tests

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [docs/HarnezSplashTarget.md](../docs/HarnezSplashTarget.md), [issues/029-declarative-centered-layout-and-viewport-alignment-primitives.md](029-declarative-centered-layout-and-viewport-alignment-primitives.md), [issues/030-braille-activity-spinner-and-bracketed-progress-bar-primitives.md](030-braille-activity-spinner-and-bracketed-progress-bar-primitives.md), [issues/031-provider-status-pill-cluster-and-lifecycle-state-presentation.md](031-provider-status-pill-cluster-and-lifecycle-state-presentation.md), [issues/032-splash-lifecycle-controller-async-provider-coordination-and-key-dismissal.md](032-splash-lifecycle-controller-async-provider-coordination-and-key-dismissal.md)

---

## 1. Problem & Motivation

With centering layout, spinner/bar widgets, status pills, and the lifecycle controller in place, we need an integrated executable demo in `examples/` (or within the Harnez target example app) along with deterministic golden output tests matching the visual target in `docs/HarnezSplashTarget.md`.

## 2. Technical Specification

1. **Declarative Spec / Component Assembly**:
   - Construct the full splash screen layout (header, bar, step text, provider pills, footer).
   - Integrate with the example application driver to demonstrate the startup sequence before launching the dashboard.
2. **Deterministic Test Suite**:
   - Render frames at specific lifecycle stages (0%, 50%, done) against golden text snapshots.
   - Test layout fidelity under standard (80x24), narrow (40x20), and wide (120x40) terminal geometries.
3. **Interactive Manual Demo**:
   - Executable sub-command / example flag (`go run ./examples/...`) allowing live testing of animation and `Esc` key skip.

## 3. Implementation & Verification Plan

- [x] Assemble splash screen view in example application.
- [x] Golden output unit tests for initial, mid-flight, and finished splash frames.
- [x] Viewport resize tests ensuring center alignment is maintained across dimensions.
- [x] Document execution in example README.

