# Geometry Gate

Ticket 010 gates richer content on terminal-cell geometry, not byte or rune counts.

## Supported text policy

Canvas supports printable base characters plus combining marks; tested fixtures
include ASCII, `é`, CJK `中`/`界`, `🔍`, block and Braille glyphs. Ambiguous-width
characters use one cell. CJK and the supported emoji ranges use two. This is a
bounded terminal policy, not complete Unicode grapheme conformance: emoji ZWJ,
flag sequences, variation-selector presentation and terminal-specific ambiguous
width modes are not guaranteed. These require a later explicit extension.

Leading combining marks are dropped. Wide glyphs crossing either boundary are
omitted as whole glyphs. Replacing either half erases the old wide glyph cleanly.
`Set` accepts one base-plus-marks cluster, never an arbitrary string. `Write`
accepts strings. Raw terminal sequences/control characters are removed; styles
must be supplied as `Style`. Input ANSI SGR is stripped, not interpreted as style.
`StringWidth` ignores terminal instructions. Styled output contains only the
SGR produced by Loom. Frame declaration labels remain printable ASCII for now.

Boxes draw children on isolated canvases: even out-of-bounds Fill/Write/Set
cannot touch padding, borders, siblings or chrome. Bounds below a complete 2x2
border omit the box. Earlier boxes take priority when terminal height is short.

## Independent checks

`geometry_test.go` interprets final emitted SGR rows using a deliberately bounded
fixture glyph table. It never calls production width or write helpers. Unknown
glyphs and non-SGR terminal instructions fail the oracle rather than silently
assuming a width. Tests assert display widths, explicit border rectangles and
padding, and repeat resize/visibility cycles. Negative controls corrupt a border,
insert an extra column, and inject a clear-screen instruction; all are rejected.
The checker is independent of the implementation, but not a full terminal emulator.

Saved artifacts under `testdata/geometry/` are deterministic ANSI rows in JSON,
plus Bash replay scripts. They are **not raster screenshots**. Tests compare the
current renderer against these exact rows/scripts. The replay check probes shell
escape semantics first, then compares actual replay bytes with saved ANSI rows.

```sh
make test
go test -race ./...
python3 scripts/check-geometry-replay.py
bash testdata/geometry/wide.sh
bash testdata/geometry/slim.sh
```

Wide: 64x9, boxes `(0,1,31,7)` and `(33,1,31,7)`.
Slim: 40x18, boxes `(0,1,31,7)` and `(0,10,31,7)`.
Coordinates are zero-based outer rectangles. Each box has one padding cell;
top/bottom chrome occupy the first/last rows. Green bold and underlined test
content intentionally attempts overflow. Dots make the child area visible.

Evidence date: 2026-09-10. Automated/headless results: passed. Human visual
confirmation: requested, not yet recorded. No desktop/raster capture claimed.
Unattended continuation is permitted after the automated gate passes; a visual
report of misalignment reopens this gate before further dependent work.
