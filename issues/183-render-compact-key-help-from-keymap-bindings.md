# 183 — Render compact key help from KeyMap bindings

**Status**: Closed
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: issues/158, cmd.go, issues/177, issues/180

---

## 1. Problem & Motivation
Help text for keys is written by hand and drifts from the actual bindings. Bubbles generates a compact help line from its key descriptors (issue 177).

## 2. Technical Specification / Findings
Once `KeyMap` (158) carries a help label per action, add a `KeyHelp` widget that renders `key label · key label …` truncated to width, with an optional expanded multi-column form. Use it in one example.

## 3. Implementation & Verification Plan
/goal Ship `KeyHelp` rendering from a `KeyMap` with width-truncation tests and use it in one example; stop and report if 158's KeyMap has no help labels to render.
