# 052 — Resolve unused ChoiceStyle Border contract

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Refactor
**Related**: `choice.go`, `theme.go`,
[Themes](../docs/Themes.md),
[#037](037-canvas-drawborder-and-drawbox-primitives-with-configurable-boxstyles.md),
[#043](043-widget-color-themes-plain-and-mc-midnight-commander.md)

---

## 1. Problem & Motivation

`ChoiceStyle` publicly exposes a `Border Style`, `DefaultChoiceStyle` initializes
it, and `ThemeColors.ChoiceStyle()` maps the theme's border role into it. However,
`Choice.Draw` never reads `c.Style.Border` and draws no border. The private
`focused` field is likewise documented as dimming the border when unfocused, but
focus currently controls only cursor placement.

This creates a misleading API: callers can configure a style that has no visible
effect, and theme documentation can easily claim a consumer that does not exist.

## 2. Decision Required

Choose and document one coherent contract:

- Remove `ChoiceStyle.Border` and correct the focus comment because framing is
  owned by `Box`; or
- Make `Choice` own an optional border, define whether it consumes rows/columns,
  and implement focused/unfocused border styling.

Avoid silently adding always-on chrome to `Choice`, which would change existing
geometry and reduce its item viewport. If borders belong to reusable container
primitives, coordinate the decision with issue 037.

## 3. Verification & Acceptance

- Every public `ChoiceStyle` field is consumed or intentionally removed.
- Focus documentation matches visible behavior.
- Existing unframed choices retain their geometry unless an explicit border is
  requested.
- Theme role documentation names only actual consumers.
- Tests demonstrate the selected contract and `go test ./...` passes.
