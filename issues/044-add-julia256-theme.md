# 044 — Add `julia256` Theme

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: [#043](043-widget-color-themes-plain-and-mc-midnight-commander.md),
[spec/themes.yaml](../spec/themes.yaml),
[upstream Julia256 skin](https://github.com/MidnightCommander/mc/blob/master/misc/skins/julia256.ini)

---

## 1. Problem & Motivation

Loom currently ships `plain` and classic-blue `mc` themes. Midnight Commander's
upstream `julia256` skin offers another useful MC-derived palette: a dark,
calmer panel using `color237`, cyan selection and input rows, yellow headers,
and light-gray text and frames. Add it as a named Loom theme so applications
such as `examples/filebrowser` can select it without defining application-local
styles.

## 2. Technical Specification & Open Questions

Add a `julia256` entry to `spec/themes.yaml`, mapped from the upstream skin's
`[core]` roles:

- `_default_ = lightgray;color237` maps to normal content.
- `selected = black;cyan` maps to selected content.
- `input = black;cyan` maps to the prompt.
- `header = yellow;color237` maps to table headers.
- `frame = lightgray;color237` maps to borders.

The upstream skin mixes explicit 256-color index 237 with named terminal
colors. Choose and document deterministic palette indices for those named
colors rather than assuming a user's configurable 16-color palette. In
particular, Loom currently reserves index `0` as the terminal-default sentinel,
so the implementation must decide whether black maps to a non-sentinel 256-color
black (for example index 16) or whether the theme color representation should
first distinguish palette index 0 from terminal default.

Choose a `focus_bg` that remains distinguishable from both the normal
`color237` background and cyan selection. Record the choice in the spec rather
than duplicating palette values in Go.

## 3. Verification & Acceptance

- `spec/themes.yaml` contains a schema-valid theme named `julia256` whose roles
  trace back to the upstream Julia256 `[core]` palette.
- `loom.SpeccedThemes` and `loom.Theme("julia256")` expose the theme without a
  hard-coded Go-side duplicate.
- Unit tests cover representative normal, selected, prompt, header, border,
  and focus colors.
- `go test ./...` passes.
- `go run ./examples/filebrowser --theme julia256 .` accepts and renders the
  new theme; visually verify its contrast before treating the palette as final.
