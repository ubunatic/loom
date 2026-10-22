# 367 — Media widget: move key bindings, action labels and zoom/pan steps into the spec

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 323 (core widget bindings; use the same spec layout), docs/Spec.md

---

/goal The media widget's default bindings, action IDs and help labels, and its zoom factor, pan step and
loading-bar width come from the embedded spec, use the layout chosen in 323, and no Go table mirrors
them; tests prove a YAML change alters behaviour and help text; or stop and report when blocked on a user
decision.

## Findings (spec audit, 2026-10-22)

- `media/widget.go:111`: binding, action and label maps.
- `media/widget.go:731`: loading-bar width `8`, zoom factor `1.25`, pan step `.12`.

## Notes

- Start after 323 has settled the bindings layout.
