# Issue 443: ColorPicker selectable faint text and feasible subdued background

## Scope & Problem

Following the gallery comparison in Issue 442, the owner requested selectable faint text in ColorPicker and investigated whether a subdued background can be supported cleanly without hacks. Faint is an ANSI text attribute (`Style.Dim`, SGR 2); the theme-background sample demonstrated `NormalBG`, not a semantic dim-background color.

## Searcher Findings (s443)

- **ColorPicker Architecture:** Defined in `colorpicker.go` (`NewColorPicker`, `Measure`, `Draw`, `ConsumeKey`/`ConsumeMouse`). Palette swatches and reset `∅` (`ColorReset()`). Current callbacks: `OnChange(target, color)` where target is foreground or background, and `OnClose()`. No style/text attribute selection currently exists.
- **Style & Attributes:** `ansi/style.go` defines `Color` and `Style`. `Style.Dim` is an existing boolean emitted as ANSI SGR 2.
- **Subdued Background Feasibility:** Themes (`theme.go`) provide semantic roles (`NormalBG`, `PopupBG`, etc.), but ANSI does not support a background dim operation (SGR 2 is text/glyph faint). Inventing an RGB blend without semantic theme support would be an arbitrary hack and break terminal palettes. Subdued background should be documented as an architectural limitation requiring a dedicated theme role before it can be semantically offered.
- **Consumers & Persistence:** `RichTextEdit` (`richtextedit_style.go`) applies colors to `ActiveStyle` and `RichSpan.Style`. Spans store `Style` including `Dim`.

## Milestones

- **M1 (Plan & Investigation):** Developer inspects live contracts, designs faint-text selection in ColorPicker and its integration with `RichTextEdit` / consumers, documents the subdued background limitation, and outlines regression tests.
