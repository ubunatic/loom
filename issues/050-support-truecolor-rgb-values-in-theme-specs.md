# 050 — Support truecolor RGB values in theme specs

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [Terminal colors](../docs/TerminalColors.md),
[theme spec](../spec/themes.yaml),
[theme schema](../spec/schemas/themes.schema.json),
[#043](043-widget-color-themes-plain-and-mc-midnight-commander.md)

---

## 1. Problem & Motivation

Theme colors currently accept only `default` or an xterm-256 palette index.
This is sufficient for portable indexed themes, but the fixed color cube is too
coarse for some deliberately tuned colors.

The `mc` scrollbar experiment exposed a concrete example. Dimming palette index
69 produces approximately `#567be9` in Tilix and `#577cea` in foot and Ptyxis,
which gives the desired bluish-grey track against the index 69 background. The
closest indexed foreground, index 68 (`#5f87d7`), is visibly different. Keeping
dim delegates the derived color to the terminal; specifying the observed RGB is
not currently possible in `spec/themes.yaml`.

## 2. Proposed Contract

- Extend theme color values to accept a six-digit RGB form such as `"#577cea"`
  alongside `default` and integer indices 0–255.
- Preserve the distinction between terminal default and real palette index 0.
- Extend `ThemeColor` with an RGB representation that converts to `ColorRGB`.
- Define the accepted RGB syntax in `spec/schemas/themes.schema.json`; reject
  malformed, abbreviated, and out-of-range values during validation/loading.
- Keep every concrete palette choice in `spec/themes.yaml`; do not duplicate RGB
  values in Go.
- Document that truecolor requests are more precise than indexed colors but can
  still be transformed by the terminal or display pipeline.

Whether the initial implementation changes the adaptive `mc` track should remain
a separate visual decision. Adding RGB support must not silently replace the
currently accepted rendering.

## 3. Verification & Acceptance

- YAML loading and schema validation accept `default`, indices including 0, and
  six-digit RGB values.
- `ThemeColor.Color()` emits the existing 24-bit `ColorRGB` representation for
  RGB values without changing indexed-color behavior.
- Tests cover valid lowercase and uppercase RGB, palette index 0, terminal
  default, malformed strings, and schema negative controls.
- At least one opt-in theme or fixture exercises RGB end to end.
- `go test ./...`, `go vet ./...`, and `make validate-spec` pass.
