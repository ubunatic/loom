---
title: Splash Screen Architecture and Runtime Safeguards
---

# Splash Screen Architecture and Runtime Safeguards

## Header & Context

Date: 2026-09-11. Scope: Implement Stage 10 splash screen and startup lifecycle (Tickets 029–033) matching [`docs/HarnezSplashTarget.md`](../HarnezSplashTarget.md), diagnose raw terminal exit edge-cases, establish library-level fallback safeguards, and move all runtime defaults into formal embedded specifications.

## Executive Summary

Stage 10 deliverables were completed end-to-end, establishing centered viewport alignment, braille rotation spinners, bracketed progress bars with sub-character precision, provider status pill clusters, asynchronous task coordination, and deterministic golden test verification:

1. **Alignment Primitives (029)**: Introduced `layout.Align` (`AlignStart`, `AlignCenter`, `AlignEnd`) and `layout.AlignOffset`, plus `loom.AlignBox` / `loom.NewCenter` with zero-allocation dimension clamping.
2. **Braille Spinner & Bracketed Bar (030)**: Added `graph.SpinnerGlyph` for 10-frame braille rotation and `graph.RenderBracketedBar` for determinate braille fill (`[⣿⣿⣿⣿⡇...]`) and dot-matrix patterns (`[::::...]`).
3. **Provider Status Pills (031)**: Built `loom.ProviderPill` and `loom.PillCluster` with ANSI lifecycle state styling (`Pending` dim, `Fetching` yellow, `Done` green, `Failed` red).
4. **Splash Controller (032)**: Implemented `loom.SplashController` coordinating background worker probes, timer ticks, immutable snapshot publications, and keyboard skip hooks.
5. **Integrated Example & Goldens (033)**: Created `loom.SplashView` widget and `examples/splash/` with live `--watch` and show-once CLI, validated by golden text frames across 80x24, 40x20, and 120x40 geometries.

During interactive testing, an entrapment issue was uncovered: newly created widgets with incomplete `HandleKey` methods left terminal users unable to quit via `Esc` or `q`. This led to two foundational architectural upgrades:
- **Global Driver Safeguard**: Added `Pane.handleKeyFallback` to ensure unhandled keys fallback to safe exit (`Esc`, `Ctrl-C`, `Ctrl-Q`, `Ctrl-D`, `q`) unless explicitly disabled.
- **Specced Defaults**: Moved all fallback keys and runtime constants into [`spec/defaults.yaml`](../../spec/defaults.yaml) and [`spec/schemas/defaults.schema.json`](../../spec/schemas/defaults.schema.json), embedded into `loom.SpeccedDefaults` as single source of truth.

## Key Learnings & Architecture Decisions

### 1. The Terminal Entrapment Trap
In raw terminal mode, standard signals like SIGINT (Ctrl-C) may be captured as raw byte streams (`\x03`) rather than triggering OS-level termination. When a widget returns `false` from `HandleKey` without handling exit keys, the user is trapped.
- **Solution**: The `Pane` event loop now acts as the safety net of last resort. If `root.HandleKey(ke)` returns `false` (unconsumed), `Pane` consults `SpeccedDefaults.FallbackQuitKeys`. If the user presses an exit key, `Pane` restores terminal state and exits cleanly.

### 2. Specced Library Defaults as Single Source of Truth
Following Loom's specification principles in [`docs/Spec.md`](../Spec.md), hardcoded Go fallback tables were eliminated:
- `spec/defaults.yaml` declares fallback keys, pane max columns, and splash timing/dimensions.
- `cmd/validate-spec` validates `defaults.yaml` against its JSON Schema on every build.
- `defaults.go` embeds and unmarshals this spec into `loom.SpeccedDefaults`.

## Verification Evidence

- `make test`: Schema validation, geometry replay canary, vet, and all package tests pass.
- `go test -race ./...`: Zero data races across concurrent splash controller routines and redraw loops.
- `make watch-pty`: PTY watch test confirms clean redraw, resizing, toggle actions, and terminal restoration.
- `harnez status`: All 33 tracker tickets verified and consistent.
