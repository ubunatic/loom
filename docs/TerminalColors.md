# Terminal Colors

Terminal rendering combines independent inputs. Treating them separately makes
themes easier to reason about and exposes which parts an application actually
controls:

- The glyph defines geometry or texture. Unicode defines `░`, `▒`, and `▓` as
  light, medium, and dark shade characters with approximately 25%, 50%, and 75%
  coverage.
- The foreground colors the glyph.
- The background colors the remainder of the terminal cell.
- SGR attributes such as bold and dim modify the terminal's rendition. They are
  not colors.

## Authority Levels

A theme using an explicit foreground, background, and `dim: false` controls the
logical color pair. A shade glyph then spatially mixes those colors without Loom
calculating an intermediate color:

```yaml
scrollbar_track_fg: 68
scrollbar_track_bg: 69
scrollbar_track_dim: false
```

This is authoritative at the terminal-palette level, not necessarily at the RGB
level. Users and terminals may redefine indexed colors. Terminal-default colors
are less constrained still. Truecolor RGB values provide stronger color
authority. Theme specs accept six-digit `#RRGGBB` values in addition to
`default` and palette indices 0–255, though the terminal or display pipeline
may still transform the result.

SGR 2 means faint or decreased intensity, but the standard does not define a
dimming formula or resulting RGB value. A theme that sets `dim: true` therefore
delegates part of its color choice to the terminal.

## Shade Glyphs

Shade glyphs are useful terminal UI primitives because their texture survives
when color escape sequences are removed. They also permit a track or inactive
region to remain distinguishable in monochrome output.

Use the glyph to select density and the color pair to select the palette:

- `░` for a subtle texture, approximately 25% foreground coverage.
- `▒` for a balanced texture, approximately 50% foreground coverage.
- `▓` for a strong texture, approximately 75% foreground coverage.
- ` ` with an explicit background for the most reliable solid cell.
- `█` for a foreground-colored solid glyph when glyph semantics are useful;
  font rasterization can still produce seams.

Exact texture and antialiasing remain font- and renderer-dependent. For a
predictable colored texture, use an explicit foreground and background without
dim. For a predictable solid cell, prefer a space with an explicit background.

## The `mc` Scrollbar Experiment

The `mc` theme uses palette index 69 (`#5f87ff`) as its main background. Its
scrollbar track renders `░` with foreground 69, background 69, and `dim: true`.
Without dim, the identical foreground and background would make the texture
invisible.

Screenshots of the same track produced these foreground pixels:

| Terminal | Dimmed foreground |
|----------|--------------------|
| Tilix | `#567be9` |
| foot | `#577cea` |
| Ptyxis | `#577cea` |

Against the `#5f87ff` background, the 25% shade has an approximate aggregate
color of `#5d84fa`. The samples show that the dimming convention is consistent
across the tested terminals, including both VTE-based terminals and foot, even
though the standard does not guarantee that exact result.

The closest xterm-256 foreground is index 68 (`#5f87d7`), but it is visibly
different: the fixed palette cannot represent the observed bluish-grey dimmed
foreground. Other nearby cube entries shift toward purple or cyan. Using an
explicit value such as `#577cea` is now possible when a fixed RGB request is
preferable; using foreground 69, background 69, and dim remains a deliberate
terminal-adaptive alternative for this track.

## Theme Guidance

- Prefer explicit foreground/background pairs and disable dim in fully specified
  themes.
- Use dim when terminal adaptation is intentional or the indexed palette cannot
  express the desired intermediate color.
- In the `plain` theme, terminal-default colors and dim are appropriate because
  adaptation to the user's terminal is the theme's purpose.
- Keep shade glyphs when monochrome degradation and visible texture matter.
- Do not describe a dimmed indexed color as a fixed RGB value; record it as an
  observation from named terminals.
- Verify tuned themes in more than one terminal family. Screenshots should cover
  the normal background, selection, prompt, frame, scrollbar track, and thumb.

The character coverage values come from the
[Unicode Block Elements chart](https://www.unicode.org/charts/nameslist/n_2580.html).
The definition of SGR 2 comes from
[ECMA-48](https://ecma-international.org/publications-and-standards/standards/ecma-48/).
