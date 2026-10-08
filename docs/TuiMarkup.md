---
title: .tui Markup
weight: 76
---

# .tui Markup

`.tui` files hold styled text as readable bracket markup in the style of
Rich and Spectre.Console. Loom parses them into a `RichDocument` with
`loom.ParseMarkup` and writes them with `RichDocument.ToMarkup`.

```
[bold red on #202020]Warning:[/] disk [italic]almost[/] full
[cyan]a[underline]b[not underline on blue]c[/]d[/]e[/]
```

## Tags

| Form | Meaning |
|---|---|
| `[attrs]` | push a style: the current style plus `attrs` |
| `[/]` | pop the innermost style |
| `[/attrs]` | pop the innermost style; `attrs` must match its opening tag |

Attributes, separated by spaces and case-insensitive:

| Attribute | Effect |
|---|---|
| `bold`, `b` | bold |
| `dim` | dim |
| `italic`, `i` | italic |
| `underline`, `u` | underline |
| `strike`, `s` | strikethrough |
| `invert`, `reverse` | swap foreground and background |
| `not <attr>` | clear an inherited attribute |
| `<color>` | foreground color |
| `on <color>` | background color |

## Colors

- `default`: the terminal's default color.
- A name from `spec/markup.yaml`: `black`, `red`, `green`, `yellow`, `blue`,
  `magenta`, `cyan`, `white`, their `bright_*` variants, and `grey`/`gray`
  (palette indexes 0-15).
- `color(N)`: 256-color palette index, `N` from 0 to 255.
- `#rgb` or `#rrggbb`: true color.

## Escapes

| Write | For |
|---|---|
| `\[` or `[[` | `[` |
| `\]` | `]` |
| `\\` | `\` |

A lone `]` and a backslash before any other character are kept literally.

## Rules

- Parsing is strict. Unknown attributes or colors, empty or unterminated tags,
  mismatched `[/attrs]` and tags still open at the end of a line are errors
  of type `*loom.MarkupError` with a 1-based line and column (in characters).
- Tags close on the line where they open, so every line parses on its own.
- CRLF and CR become LF. One trailing newline is dropped on read and always
  written on save.
- The serializer writes one `[attrs]text[/]` group per styled span and leaves
  default-styled text bare, with attributes in the order of the table above.
- Pill, link and code span metadata are not part of v1; their visible text is
  kept.

## Tools

- `loom edit file.tui` opens and saves markup. `RichTextEdit` picks the format
  from the file extension on every load and save, so Save As from `.tui` to
  `.ansi` (or back) converts the file. A custom `SerializeDocument` hook still
  takes precedence.
- `loom markup file.tui` prints ANSI; `loom markup file.ansi` prints markup.
  Validate a design with `loom markup art.tui > art.ansi`, then `loom eval`,
  `loom measure` and `loom check-box` on `art.ansi`.
