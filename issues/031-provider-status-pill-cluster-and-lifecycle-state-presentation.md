# 031 — Provider Status Pill Cluster and Lifecycle State Presentation

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [docs/HarnezSplashTarget.md](../docs/HarnezSplashTarget.md), [issues/011-aligned-dashboard-rows-and-ansi-safe-truncation.md](011-aligned-dashboard-rows-and-ansi-safe-truncation.md)

---

## 1. Problem & Motivation

The splash screen displays a horizontal cluster of status pills representing individual provider backends:
`● mic   ✳ claude   ֍ codex   Λ agy`

Each provider item pairs a distinct unicode symbol with a provider name and state-dependent color styling (e.g. gray for pending, yellow/cyan for active fetching, green for done/success, red for failed). Loom needs a reusable inline pill / badge cluster component that handles multi-item spacing and ANSI styling without distorting horizontal width.

## 2. Technical Specification

1. **Status Item / Pill Model**:
   - `Name`: string (e.g. `mic`, `claude`, `codex`, `agy`)
   - `Symbol`: string/rune (e.g. `●`, `✳`, `֍`, `Λ`)
   - `State`: `Pending`, `Fetching`, `Done`, `Failed`, `Skipped`
2. **Styling & Presentation**:
   - Map state to color styling (dim gray, highlighted accent, bright green, warning red).
   - Uniform inter-item spacing (e.g. 2 spaces between pill pairs).
   - Calculate exact aggregate display width taking full Unicode symbols into account.

## 3. Implementation & Verification Plan

- [ ] Implement `StatusPill` / `PillCluster` widget/renderer.
- [ ] Add state-to-style color mapping functions.
- [ ] Tests for width calculation and ANSI sequence safety when states change.
- [ ] Render tests matching target layout spacing.
