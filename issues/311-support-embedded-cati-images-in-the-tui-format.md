# 311 — Support embedded Cati images in the .tui format

**Status**: Draft
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 308 (.tui markup), 309 (box blocks), 310 (tables)

---

## 1. Problem & Motivation
`.tui` documents can hold styled text and boxes but no pictures. Authors of
banners, help screens and design mockups want to place a logo or screenshot
next to markup text and have it drawn with the same Unicode-block renderers
that `cati` uses. Images must also work inside `[box]` blocks and pass the
usual `loom eval` / `loom measure` / `loom check-box` validation.

## 2. Technical Specification / Findings

### 2.1 Current state in loom
- `ParseMarkup(src string) (*RichDocument, error)` (`markup.go:74`) works line
  by line. Block constructs are lines whose trimmed text is a tag
  (`markupBoxTag`, `markup_box.go:33`); everything else goes through
  `parseMarkupLine`.
- Boxes are **rendered at parse time** into plain `RichLine`s
  (`markupBox.render`). `RichDocument` has no block structure, and
  `ToMarkup` writes the drawn glyphs, not the `[box]` tag. Images follow the
  same model (see 2.4).
- `RichSpan.Style` carries `Color`, which supports true color
  (`ColorRGB`, `style.go:81`), so a rendered image cell maps onto a span
  with one glyph, an RGB foreground and an RGB background.
- `ParseMarkup` takes a string only: it has no file path, so relative image
  paths cannot be resolved today. Callers: `cmd/loom/markup.go:38` and
  `richtextedit_file.go:95`; both know the file path.
- Root package `loom` does **not** import cati. Only `loom/media`
  (`media/widget.go`) does, using `cati/v1/{core,halfblock,quadblock,sextant}`
  and the `media.Mode` names `halfblock`, `quadblock`, `sextant`.

### 2.2 Current state in cati
- Loading: `halfblock.LoadImage(path)` / `LoadImageWithTarget(path, w, h)`
  decode PNG and JPEG and rasterize SVG (via an external tool,
  `v1/halfblock/svg.go`).
- Rendering: `halfblock|quadblock|sextant.RenderToGrid(img, cols, opts)`
  scale the image (`opts.Rows` caps the height) and return a
  `*core.Grid` of `core.Cell{Ch, Fg, Bg, HasFg, HasBg, Transparent}`.
  This is the only API we need; no ANSI string parsing.
- Module graph: loom requires `ubunatic.com/cati v0.2.6`, and cati requires
  `ubunatic.com/loom v0.2.18`. Package-level there is no cycle (cati `v1/`
  imports no loom package; only cati examples and `cmd` do), but importing
  cati from the root `loom` package would push image decoding, SVG
  subprocesses and `golang.org/x/image` onto every loom user.

### 2.3 Syntax
A self-closing block tag that stands alone on its line, like `[box]`:

```
[box style=rounded title="Build"]
[image src="logo.png" width=24 mode=half alt="cati logo"]
[green]passed[/] 312 tests
[/box]
```

| Option | Values | Default |
|---|---|---|
| `src` | path, relative to the `.tui` file's directory; required | — |
| `width` | columns, 1..`spec` max | image's native size in cells, capped by `max_width` |
| `height` | rows; with `width`, fits inside both while keeping aspect | derived from aspect |
| `mode` | `half` (`halfblock`), `quad` (`quadblock`), `six` (`sextant`) | `half` |
| `alt` | plain text used when no image renderer is set or loading fails in lenient mode | `[image: <src>]` |

- Option parsing reuses the box tag's `key=value` / quoted-value lexer
  (factor it out of `parseMarkupBoxTag` into a shared helper).
- Inline images inside a text line are **out of scope**: an image is several
  rows tall and a line holds one row. `[image ...]` mid-line is a
  `MarkupError` ("image tag must stand alone on its line").
- No `[/image]` closer. Unknown options, missing `src`, a bad `width`,
  `height` or `mode`, and trailing text are `MarkupError`s with line/col.
- Mode names, defaults and limits (`max_width`, e.g. 200) go into
  `spec/markup.yaml`, not Go constants (docs/Spec.md).

### 2.4 Rendering model
- Images are rasterized **at parse time** into `RichLine`s, exactly like
  boxes: one line per grid row, one span per run of equal style. Each
  `core.Cell` becomes `Style{Fg: ColorRGB(Fg), Bg: ColorRGB(Bg)}`;
  `HasFg=false`/`HasBg=false` or `Transparent` map to the default color so
  the terminal background shows through.
- Every row has exactly `grid.Width` columns (pad short rows with spaces) so
  enclosing boxes measure a straight edge. All glyphs cati emits for these
  three modes are single-width; a test asserts this.
- Consequence (accepted, same as boxes): `ToMarkup` writes the cells as
  `[#rrggbb on #rrggbb]▀[/]` runs, not the `[image]` tag. Saving a loaded
  `.tui` from `RichTextEdit` therefore bakes the image in. Preserving block
  source is a separate, later concern for boxes, tables and images together;
  file a follow-up ticket instead of solving it here.
- The ANSI output from `loom markup` depends on the color profile the caller
  writes with; true color is expected. Lower profiles degrade through the
  existing `Color` downsampling.

### 2.5 Runtime boundary and packages
- Root `loom` stays free of cati. It defines the hook:

  ```go
  // MarkupImage is a parsed [image] tag.
  type MarkupImage struct {
      Src, Alt      string
      Width, Height int    // 0 = unset
      Mode          string // spec name: half, quad, six
      Line          int
  }

  // MarkupOptions configures ParseMarkupWith.
  type MarkupOptions struct {
      BaseDir string // resolves relative src; "" = current directory
      // Image renders an [image] tag into lines. nil: the tag renders
      // its alt text as one dim line.
      Image func(img MarkupImage, baseDir string) ([]RichLine, error)
  }

  func ParseMarkupWith(src string, opts MarkupOptions) (*RichDocument, error)
  ```

  `ParseMarkup(src)` becomes `ParseMarkupWith(src, MarkupOptions{})`
  (no breaking change; no `docs/Upgrading.md` row).
  An error from `Image` is wrapped in `*MarkupError` with the tag's line and
  column 1 (strict, like the rest of the parser).
- `loom/media` provides the cati implementation:

  ```go
  // MarkupImage renders .tui [image] tags with cati.
  func MarkupImage(img loom.MarkupImage, baseDir string) ([]loom.RichLine, error)
  // MarkupOptions returns options with BaseDir and Image set for a .tui file.
  func MarkupOptions(path string) loom.MarkupOptions
  ```

  It loads with `halfblock.LoadImageWithTarget`, converts `Mode` to
  `media.Mode`, calls the matching `RenderToGrid`, and converts the
  `core.Grid` with a helper shared with `media.Widget` where that avoids
  duplicate cell-to-style code.
- Wiring: `cmd/loom/markup.go` and `loom edit` use
  `media.MarkupOptions(path)`. `RichTextEdit` (root package) gets an
  optional `MarkupOptions func(path string) loom.MarkupOptions` field that
  `richtextedit_file.go` uses when set; `loom edit` sets it. Without it,
  `RichTextEdit` shows alt text, so library users without cati still load
  `.tui` files.
- No new module dependency: cati is already required. Check that the
  required cati version (`v0.2.6`) exports what is used; bump only if needed.

### 2.6 Security and limits
- `src` is a local path only; no URLs. Absolute paths and `..` are allowed
  (the author controls the file), but documented.
- Decoding runs synchronously during parse; cap the requested `width` by the
  spec limit so a typo cannot render a 10000-column image.

## 3. Implementation & Verification Plan

**M1 — Syntax and hook (root package, no cati)**
1. Factor the option lexer out of `markup_box.go` into a shared helper;
   box tests must stay green unchanged.
2. Add `markup_image.go`: tag detection (`[image` + space or `]`), option
   parsing, `MarkupImage`, `MarkupOptions`, `ParseMarkupWith`; images
   inside boxes append to the box's document.
3. Add `image` settings (modes, aliases, `max_width`) to
   `spec/markup.yaml` and its schema.
4. Tests in `markup_test.go`: every option, each error with line/col,
   mid-line tag error, alt fallback without a hook, hook error wrapping,
   fake hook output inside a box keeps straight borders.

**M2 — cati renderer (`loom/media`)**
1. `media/markup.go`: `MarkupImage`, `MarkupOptions`, grid-to-lines
   conversion (span merging, transparent and missing colors, row padding).
2. Tests with a generated in-memory PNG written to `t.TempDir()`: dimensions
   per mode, width/height fitting, a known 2×2 color image gives the
   expected glyphs and RGB styles, missing file gives an error.
3. Add `examples/markup/image.tui` with a small checked-in PNG (REUSE
   license entry), alone and inside a box.

**M3 — Wiring and docs**
1. `loom markup` and `loom edit` use `media.MarkupOptions(path)`;
   `RichTextEdit` gains the optional options hook.
2. `docs/TuiMarkup.md`: new "Images" section (syntax, options table, model,
   bake-in on save, alt fallback).
3. File the follow-up ticket for block-source preservation on save.

**Verification**
- `make test-q1` passes (host reruns it before closing).
- `loom markup examples/markup/image.tui > /tmp/image.ansi`, then
  `loom eval`, `loom measure` and `loom check-box` on it show no geometry
  errors; check visually in a true-color terminal.
- `loom edit examples/markup/image.tui` opens and shows the image.
- `make install`.

**/goal** Add and verify embedded Cati image support in `.tui`, or stop and report if blocked on a user decision or denied permission.
