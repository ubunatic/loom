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

## Boxes

A box block wraps lines in a border. The opening and closing tags each stand
alone on their own line (surrounding spaces are allowed):

```
[box style=rounded title="Status" border="bold cyan" pad=1]
[bold]Disk:[/] 93% full
[/box]
```

```
╭─ Status ───────╮
│ Disk: 93% full │
╰────────────────╯
```

| Option | Values | Default |
|---|---|---|
| `style` | `sharp` (alias `normal`), `rounded`, `double`, `ascii` | `sharp` |
| `title` | inline markup shown in the top border | none |
| `border` | attributes and colors for the border, as in a tag | default style |
| `pad` | spaces between border and text on the left and right | `1` |

Options are `key=value`; quote values that contain spaces. A quoted value
cannot contain `"`. Lines inside a box may use inline tags and nested boxes.

The inner width is the widest line plus padding on both sides, measured in
terminal columns, so wide characters align. A title widens the box to fit it
with one space and one border glyph on each side. Glyphs come from
`spec/box.yaml`.

`ParseMarkup` draws the box into plain lines, so a `RichDocument` holds no box
structure. `ToMarkup` writes the drawn border characters, not the `[box]`
tag; the result parses to the same lines.

Box errors: an unclosed `[box]` at the end of the file, `[/box]` without an
open box, unknown options or styles, a bad `pad`, and text after the tag on
the same line.

## Tables

A table block holds pipe-delimited rows. Like box tags, the `[table]` and
`[/table]` tags each stand alone on their line:

```
[table style=rounded border=cyan]
| Mount | Size   |   Used |
|:------|:------:|-------:|
| /     | 50 GiB | [red]93%[/] |
| /home | 2 TiB  |    12% |
[/table]
```

```
╭───────┬────────┬──────╮
│ Mount │  Size  │ Used │
├───────┼────────┼──────┤
│ /     │ 50 GiB │  93% │
│ /home │ 2 TiB  │  12% │
╰───────┴────────┴──────╯
```

| Option | Values | Default |
|---|---|---|
| `style` | `sharp` (alias `normal`), `rounded`, `double`, `ascii` | `sharp` |
| `border` | attributes and colors for the lines, as in a tag | default style |
| `header` | attributes for the header row; `none` or `""` for no extra style | `bold` |
| `pad` | spaces on each side of every cell | `1` |

Rows:

- Every row starts with `|`; the trailing `|` is optional. Spaces around a
  cell are trimmed. Blank lines inside the table are skipped.
- Each cell is inline markup on its own, so a tag must close inside its cell.
  The header style is the base style of header cells; cell tags apply on top.
- `\|` writes a literal pipe; the other escapes work as in any line.
- A separator row such as `|---|---|` directly after the first row makes that
  row the header and draws a line under it. Without it the table has no
  header. Its cells set the column alignment: `:--` or `---` left, `:-:`
  center, `--:` right.
- A row with fewer cells than the widest row gets empty cells.

Each column is as wide as its widest cell plus `pad` on both sides, measured
in terminal columns without the markup, so wide characters align. Tables never
shrink or wrap to fit the terminal. Junction glyphs (`┬ ┴ ├ ┤ ┼` and their
double and ascii forms) come from `spec/box.yaml`. A table may sit inside a
box; boxes and tables inside a table are errors. As with boxes, `ParseMarkup`
draws plain lines and `ToMarkup` writes the drawn characters.

Table errors: an unclosed `[table]`, `[/table]` without an open table, a
block tag inside a table, a row that does not start with `|`, a separator row
anywhere but second, unclosed or unknown tags in a cell, and the option errors
of boxes. See `examples/markup/table.tui`.

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
