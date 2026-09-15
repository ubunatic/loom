# 043 — Widget Color Themes: `plain` and `mc` (Midnight Commander)

**Status**: Closed — implemented and verified
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Themes](../docs/Themes.md),
[Terminal colors](../docs/TerminalColors.md),
[theme spec](../spec/themes.yaml),
[theme schema](../spec/schemas/themes.schema.json),
[#037](037-canvas-drawborder-and-drawbox-primitives-with-configurable-boxstyles.md),
[#044](044-add-julia256-theme.md),
[#050](050-support-truecolor-rgb-values-in-theme-specs.md),
`f5f37fc`, `0a5db81`, `288230f`

---

## 1. Problem & Motivation

Loom widgets previously carried disconnected styles with no named palette that
an application could apply consistently across a widget tree. The initial goal
was to provide an adaptive terminal-default theme and a Midnight Commander-style
blue-panel theme without duplicating palette values in Go.

## 2. Implemented Contract

- `spec/themes.yaml` is the single source of truth and is validated by
  `spec/schemas/themes.schema.json`.
- The embedded spec is decoded into `SpeccedThemes`; `Theme(name)` returns a
  named entry and falls back to `plain` when absent.
- `ThemeColor` distinguishes the terminal-default `default` token from every
  xterm-256 index, including the real palette index 0.
- Semantic roles cover normal, selected, header, prompt, placeholder, scrollbar
  track/thumb, status, border, and focus colors plus their relevant attributes.
- `ThemeColors` exposes adapters for `ChoiceStyle`, `TableStyle`,
  `ScrollbarStyle`, `FrameStyle`, `BoxStyle`, and `Grid.FocusBG`.
- `plain`, `mc`, `mc-classic`, and `mc-dark` are shipped palettes. The first is
  terminal-adaptive; the MC variants use explicit indexed colors.
- The filebrowser validates `--theme`, defaults to `mc`, applies the theme to its
  entire widget tree, and cycles spec-defined themes with F9.

The `mc` scrollbar intentionally uses a shade glyph with identical indexed
foreground/background colors plus dim. Cross-terminal measurements and the
reason for this exception are recorded in `docs/TerminalColors.md`.

## 3. Deferred Work

- Add the Julia256 palette in issue 044.
- Add truecolor values to the theme spec in issue 050.
- Resolve the unused `ChoiceStyle.Border` contract in issue 052.
- Keep border glyph variants and reusable drawing primitives in issue 037.
- Keep fill-height layout behavior separate in issue 051.

## 4. Verification

- Tests distinguish terminal default from palette index 0.
- Theme tests cover representative roles and verify distinct MC variants.
- Choice, View, Frame, and Box tests verify style propagation.
- Filebrowser tests cover valid/invalid theme selection, runtime cycling,
  persistence across directory changes, themed frames/boxes/scrollbars, and F10
  quit behavior.
- `go test ./...`, `go vet ./...`, and `make validate-spec` passed before the
  implementation commit.
