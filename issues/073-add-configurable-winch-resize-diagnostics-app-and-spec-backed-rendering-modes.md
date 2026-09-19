# 073 — Add configurable Winch resize diagnostics app and spec-backed rendering modes

**Status**: Open
**Priority**: P2
**Severity**: Moderate
**Category**: Feature
**Related**: [Issue 072](072-ensure-stable-responsive-frame-layout-during-terminal-resize.md)

---

## 1. Problem & Motivation

Resize rendering has several interacting mitigations, including event coalescing, buffered output, stale-row clearing, synchronized output, and auto-wrap handling. Current examples do not provide a controlled way to enable or disable these behaviors independently, making terminal-specific flicker and stale-content reports difficult to reproduce.

Add an `examples/winch` diagnostic application that exposes the resize modes as runtime switches. Put the switches and their defaults in `spec/` so the framework can validate and evolve them before selecting stable defaults.

## 2. Technical Specification / Findings

The app should exercise real Loom pane and resize behavior, display active modes, and support repeated wide → narrow → wide resize bursts. Candidate switches are:

1. Resize-event coalescing.
2. Atomic buffered frame flushes.
3. Per-row `CSI K` clearing for stale content.
4. Synchronized output mode (`CSI ?2026h` / `CSI ?2026l`).
5. Terminal auto-wrap handling.
6. Out-of-band resize clearing, retained only as a diagnostic comparison mode if safely exposed.

The current fixes are candidate defaults. The spec must be the single source of truth for switch names, descriptions, and defaults; Go code must not duplicate those values.

## 3. Implementation & Verification Plan

### M1 — Define spec-backed resize modes

- Add the mode schema and defaults under `spec/`.
- Extend spec validation and Go loading/access as needed.
- Add positive and negative validation coverage.

Verification: `go run ./cmd/validate-spec` and focused spec tests.

### M2 — Add framework configuration switches

- Route each supported mode through the pane/rendering path.
- Preserve current behavior as the default configuration.
- Keep unsafe or obsolete behavior explicitly diagnostic-only.

Verification: unit tests prove each switch changes the intended output/path without changing unrelated layout behavior.

### M3 — Build `examples/winch`

- Add a runnable resize stress surface and active-mode display.
- Provide keyboard controls to toggle modes and restore defaults.
- Keep the UI usable at narrow and wide terminal sizes.

Verification: the example builds and runs; headless tests construct `&loom.Pane{}` for configuration checks and avoid `/dev/tty` dependencies.

### M4 — Resize regression coverage

- Add PTY coverage for repeated wide → narrow → wide bursts.
- Verify reduced motion, resize, theme, nested-pane, and stale-row behavior under the default mode set.
- Record manual terminal verification notes for at least one modern terminal.

Verification: `make test-q1`, focused PTY tests, and manual resize checks.

### M5 — Select stable defaults

- Compare modes using the diagnostic app across supported terminal environments.
- Update spec defaults only after behavior is verified.
- Document selected defaults and known terminal limitations.
