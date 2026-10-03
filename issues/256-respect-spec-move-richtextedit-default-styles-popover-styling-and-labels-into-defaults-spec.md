# 256 — respect spec: move richtextedit default styles, popover styling and labels into defaults spec

**Status**: Closed — moved RichTextEdit defaults to spec/defaults.yaml
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: `docs/Spec.md`, `spec/defaults.yaml`, `spec/schemas/defaults.schema.json`, `defaults.go`, `richtextedit.go`, `richtextedit_test.go`

---

## 1. Problem & Motivation

In accordance with `docs/Spec.md` and the `/respect` skill, all hardcoded widget runtime defaults, labels, delimiters, and default presentation colors must reside in the single-source-of-truth `spec/defaults.yaml` and its JSON schema `spec/schemas/defaults.schema.json`, rather than being hardcoded in Go code.

Currently, `richtextedit.go` contains several hardcoded defaults:
- Selection background color: `ColorIndex(24)`
- Popover toolbar styling: `FG: ColorIndex(15), BG: ColorIndex(239)`
- Separator glyph and style: `│`, `FG: ColorIndex(8), BG: ColorIndex(239)`
- Anchor pointer glyph and style: `▲`/`▼`, `FG: ColorIndex(8)`

## 2. Technical Specification

### A. Spec Definition (`spec/defaults.yaml` & `spec/schemas/defaults.schema.json`)
Add a `rich_text_edit` section to `spec/defaults.yaml`:
```yaml
rich_text_edit:
  selection_bg: 24
  toolbar_fg: 15
  toolbar_bg: 239
  separator_glyph: "│"
  separator_fg: 8
  separator_bg: 239
  pointer_fg: 8
```
Update `spec/schemas/defaults.schema.json` with matching required properties and validation types.

### B. Go Spec Embeddings & Consumption (`defaults.go` & `richtextedit.go`)
- Add `RichTextEditDefaults` struct to `defaults.go` mapping to `LibDefaults.RichTextEdit`.
- In `richtextedit.go`, consume `SpeccedDefaults.RichTextEdit.*` for default selection styling, popover toolbar colors, separators, and pointers.

## 3. Implementation & Verification Plan

- **M1 (Spec & Defaults Definition)**:
  - Add `rich_text_edit` to `spec/defaults.yaml` and update `spec/schemas/defaults.schema.json`.
  - Add `RichTextEditDefaults` struct and field in `defaults.go`.
  - Verify schema validation via `go run ./cmd/validate-spec`.
- **M2 (RichTextEdit Consumption & Verification)**:
  - Update `richtextedit.go` to consume `SpeccedDefaults.RichTextEdit.*`.
  - Add/update unit tests in `defaults_test.go` and `richtextedit_test.go` verifying spec values are loaded and used.
  - Run `make test-q1`, run `make install`.
  - Close issue 256.
