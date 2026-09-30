# 202 — Media zoom crops instead of scaling

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 195 gallery

---

## 1. Problem & Motivation
In the media widget (gallery and `examples/media`), zoom in/out only shrinks or grows the inner image area; it does not scale the zoomed region back up to the panel size. Zoom is effectively a crop.

## 2. Technical Specification / Findings
Zooming should select a source region of 1/zoom of the image (around the pan centre) and resample it to fill the full widget area. See `media/widget.go` (112 controls).

## 3. Implementation & Verification Plan
Test: at 2× the rendered area stays the same size as at 1× and shows a magnified region (distinct pixels widen); update `docs/progress/Media.ansi`.

/goal Zoom magnifies the image to fill the panel, verified by test; or stop and report when blocked on a user decision.
