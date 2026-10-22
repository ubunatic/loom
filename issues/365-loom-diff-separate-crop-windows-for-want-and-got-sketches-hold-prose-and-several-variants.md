# 365 — loom diff: compare named parts of a design sketch with named widget parts of a running app

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 362 (`loom diff`), 351 (live-check `diff` step), 133 (pane debug ruler overlay)

---

/goal A TUI designed in an `.ansi` sketch can be checked part by part against the running app:
`loom diff` takes a named part of the sketch and a named widget part of a capture (or explicit crop windows
for either side) and compares them cell by cell; the ColorPicker variant b2 in
docs/data/colorpicker-ideas-001.ansi passes against a live `loom edit` capture; tests and docs/CLI.md; or stop
and report when blocked on a user decision.

## Request (user, 2026-10-22)

The user will design TUIs in `.ansi` files (as with the ColorPicker). `loom diff` must extract parts of the
design doc and compare them with parts of an app, either by crop geometry or by tagging diffable parts in
docs and marking widget parts in the tested apps.

## Findings (host, loom at db59497)

- Already robust: trailing spaces and empty rows compare equal; `--ignore-colour` is the plain-text pass,
  `--ignore-glyph` the colour-only pass.
- Gap: `--region` applies one window to both files; sketches hold prose and several variants, and the
  widget sits elsewhere in the capture.

## Decisions (host)

- **M1, crop geometry:** `--want-region` and `--got-region` (same width/height); `--region` stays as the
  shorthand for both. This is the fallback that always works.
- **M2, parts in the sketch:** mark a part with invisible start/end markers (a private OSC sequence, e.g.
  `ESC ] loom;part=b2 BEL` … `ESC ] loom;part-end BEL`). Terminals ignore unknown OSC, so `cat` and
  `loom eval` still show the sketch as drawn, and the marker moves with the content when the sketch is
  edited, unlike row/column numbers in a sidecar file. `loom diff --want-part b2`; `loom parts <file>` lists
  parts with their windows. Verify that loom's parser and RichTextEdit keep (not strip) these markers.
- **M3, parts in the app:** widgets get an optional part name; with an env var (e.g. `LOOMPARTS=<file>`)
  the running app writes the current screen rect of each named widget, so `--got-part ColorPicker`
  crops the capture to it. The live-check script (351) passes the file through.
- No comment syntax or visible frontmatter in `.ansi` files.

## Open questions for the plan turn

- Whether a sketch part can be smaller than the widget (e.g. only the marker rows), and how to anchor it.
- Name collisions when the same widget appears twice on screen.
