# 237 — Gallery uses julia256 as the default theme

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 044 (julia256 theme), 231 (gallery F9 theme switch)

---

/goal `loom widgets` starts in the julia256 theme when `--theme` is not given, with a test,
or stop and report when blocked on a user decision.

## 1. Problem & Motivation
User request: make julia256 the gallery's default theme. Today `loom widgets` defaults to `plain`.

## 2. Technical Specification / Findings
- Default is the `--theme` flag default in `cmd/loom/widgets.go` (`"plain"`). F9 cycling keeps working
  from the new start index.
- Check gallery tests and PTY tests that assume `plain` as the start theme; pass `--theme plain`
  explicitly where a test needs it rather than loosening assertions.

## 3. Implementation & Verification Plan
- Change the default, update help/docs mentioning the default, test the default theme name.
- `make test-q1`, `make install`, PTY check of `loom widgets`.

## 4. Milestones
- M1 delivered (a43a19d): `--theme` default is julia256; cmd/loom test covers the default, F9 from it and
  explicit `--theme plain`.
- M2 Pre-Work / Required Refinements: host `HTO=0 make test-q1` fails in `gallery/gallery_test.go`:
  TestGalleryChoiceQuitContractPTY, TestGalleryThemeFooterSurfacePTY, TestGalleryMouseRoutingPTY and
  TestWidgetKeyRoutingPTY start the gallery without `--theme` and wait for "Theme: plain". Pass
  `--theme plain` explicitly where a test needs plain (keep assertions). The developer's run was cut by
  the harnez 60 s timeout; run `HTO=0 make test-q1`.
