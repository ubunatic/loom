---
title: Loom CLI
weight: 75
---

# Loom CLI

`loom info` prints the loom version, terminal dimensions when available,
terminal environment variables, detected colour profile, a Unicode width
sample, and the detected graphics rendering path. Pixel dimensions are shown
as unknown when the terminal does not report them.

The mouse tracking line describes Loom's SGR 1003 mode request; it does not
claim the current terminal supports every mouse reporting mode.

`loom info --watch` opens a full-screen view, enables SGR any-motion mouse
tracking, and refreshes terminal and environment values once per second and
after resizes. The pointer line
shows the widget's resolved 0-based cell coordinates and the original 1-based
SGR report coordinates. A `+` marks the resolved cell in the view. Press `q` or
`Esc` to exit.

The command reuses `loom.DetectColorProfile` and `measure.DetectRenderPath` so
reported capabilities follow the same detection rules as Loom rendering.

`loom markup <file>` converts a `.tui` markup file to ANSI, or any other file
(read as ANSI) to `.tui` markup, and prints the result. See
[.tui markup](TuiMarkup.md).

`loom diff <want.ansi> <got.ansi>` compares every terminal cell, including the
glyph and foreground/background colours. It accepts ANSI sketches and output
from `tmux capture-pane -pe`. Differences are printed in row-major order with
1-based row and display-column coordinates; reset, indexed, and truecolour
values are shown distinctly. Different screen sizes compare as though the
smaller screen were padded with default blank cells. Exit status is 0 when the
screens match, 1 when cells differ, and 2 when an input or option is invalid.

Use `--region row,col,width,height` to compare a 1-based cell rectangle in
both screens. `--ignore-glyph` and `--ignore-colour` omit those fields from the
comparison.

`loom crop <file.ansi> --region row,col,width,height` writes a selected ANSI
rectangle to stdout; add `--output <file>` to save it. `loom diff` also accepts
`--got-region` to crop only the capture, or `--find <text>` to locate a unique
anchor in the capture and compare a reference-sized area from that point.

`loom fmt <file.ansi>` rewrites equivalent SGR sequences into Loom's canonical
ANSI encoding without changing the displayed cells. Add `--write` to replace
the file. This is separate from `loom format`, which changes row width or box
alignment.

For widget references, crop a state from a sketch, run `loom fmt`, and keep the
asset under `testdata/` beside its test. Compare each meaningful widget state
with `loomtest.CompareWidget`; use `-update` only for deliberate design changes.

`loom man` prints the generated `loom(1)` roff page. Pipe it to `man -l -` to
view it or redirect it to a file. `loom man --install` generates and installs
the full Cobra command page tree in the standard man directory: under
`$XDG_DATA_HOME/man/man1` (or `~/.local/share/man/man1`) for regular users and
`/usr/local/share/man/man1` for root. `loom man --dir <path>` installs to a
custom directory.
