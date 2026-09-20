# 084 — Enforce explicit luna agent model selection without falling back to host provider

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Agentic Ergonomics
**Related**: `AGENTS.md`, `@docs/AgenticLoop.md`

---

## 1. Problem & Motivation

When the user or orchestration loop specifies `luna` (e.g. `codex:luna:low` or low-cost worker tiers) for developer tasks, agents sometimes substitute or fall back to their own host models (e.g., native subagents) instead of dispatching the requested external model via Harnez. This violates explicit cost and tier boundaries and increases token expenditure.

## 2. Desired Behavior & Goal

`/goal`: Guarantee that whenever instructions specify `luna` (or any explicit model/provider tier), agents must strictly dispatch the requested worker via `harnez agent start <provider:model:tier>` without falling back to host default models or native subagents.

- **Strict Model Compliance**: Explicit worker specs (e.g. `luna`, `codex:luna:low`) must directly map to `harnez agent start codex:luna:low` (or specified provider:model) without silent fallback.
- **Fail-Fast Policy**: If the specified external model/provider is unavailable or unconfigured, the agent must report the error or prompt for guidance rather than silently substituting an expensive host model.
