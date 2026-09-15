# Themes

Loom themes are named sets of semantic color roles. They are application code,
not runtime configuration: `spec/themes.yaml` is validated, embedded into the Go
package, and loaded when the package initializes.

```text
spec/themes.yaml
        │
        ├── validated by spec/schemas/themes.schema.json
        │
        └── embedded and decoded into SpeccedThemes
                              │
                              └── ThemeColors style adapters
                                         │
                                         └── widget Style fields
```

The spec is the single source of truth for named palettes. Go code maps roles to
widgets but must not duplicate palette values.

## Color Values

`ThemeColor` distinguishes two forms:

- `default` resets to the terminal's default foreground or background.
- An integer from 0 through 255 selects that real palette index. Index 0 is not
  a reset sentinel.

The zero value of `ThemeColor` is terminal default, while
`ThemeColorIndex(0)` explicitly selects palette index 0. This distinction lets
fixed themes request black without changing the behavior of the adaptive
`plain` theme.

Theme specs do not yet accept RGB values. See
[issue 050](../issues/050-support-truecolor-rgb-values-in-theme-specs.md) and
[Terminal Colors](TerminalColors.md) for the authority levels and the measured
scrollbar-dimming case that exposed this gap.

## Semantic Roles

| Theme role | Primary consumers |
|------------|-------------------|
| Normal | `Choice`, `Table`, `Frame` background, `Box` background/footer |
| Selected | `Choice` and `Table` selected rows |
| Header | `Table` header and sort header |
| Prompt | `Choice` and `Table` prompts |
| Placeholder | Empty `Choice` filter value |
| Scrollbar track/thumb | `Choice` and `View` scrollbars |
| Status | `Frame` status row |
| Border | `Box` border/title and `Frame` title |
| Focus background | `Grid` focus highlighting |

Roles carry the attributes needed for their semantics, such as selected/header
bold, placeholder/track/status dim, and scrollbar-thumb bold. Applications
should consume the adapters instead of reconstructing styles field by field:

```go
theme := loom.Theme("mc")

choice.Style = theme.ChoiceStyle()
view.Style = theme.ChoiceStyle().Normal
view.Scrollbar = theme.ScrollbarStyle()
frame.Style = theme.FrameStyle()
box.Style = theme.BoxStyle()
table.Style = theme.TableStyle()
grid.FocusBG = theme.FocusBGColor()
```

## Selection and Runtime Switching

`Theme(name)` returns the named theme and falls back to `plain` for an unknown
name. This is useful for tolerant library callers. User-facing selectors should
instead validate membership in `SpeccedThemes` so a typo produces an actionable
error and a list of available names.

Runtime theme switching is application orchestration. The application must:

1. Store both the active theme name and `ThemeColors` value.
2. Apply every relevant adapter to existing widgets.
3. Apply the stored theme when recreating widgets, such as a file list rebuilt
   after entering another directory.
4. Display the active name when the interface offers a cycle action.

The filebrowser example implements this pattern with `--theme`, an F9 cycle in
sorted spec-name order, and an `applyTheme` method. F10 and Ctrl-Q quit while a
plain `q` remains available to the filter.

## Adding a Theme

1. Add the complete named role set to `spec/themes.yaml`.
2. Use `default` only when adaptation to the user's terminal is intentional.
3. Define explicit indices for fixed themes, including index 16 when fixed black
   is preferable to a configurable 16-color entry.
4. Add representative palette assertions without duplicating the whole spec in
   tests.
5. Run `make validate-spec`, `go test ./...`, and `go vet ./...`.
6. Inspect the theme in the filebrowser and more than one terminal family.

The open [Julia256 theme ticket](../issues/044-add-julia256-theme.md) is a worked
example of mapping an upstream skin into these roles.

## Known Boundaries

- Indexed colors are palette-authoritative, not physically RGB-authoritative.
- SGR dimming is terminal-defined. Use it deliberately; do not treat its output
  as a fixed color.
- Shade glyph appearance depends partly on the terminal font and renderer.
- `ChoiceStyle.Border` is currently populated but not rendered; [issue 052](../issues/052-resolve-unused-choicestyle-border-contract.md)
  tracks removal or an explicit border contract.
- Widget geometry is separate from theme styling. The filebrowser currently
  hard-codes its box height to fill the frame; [issue 051](../issues/051-allow-frame-boxes-to-fill-available-content-height.md)
  tracks an explicit layout contract.
