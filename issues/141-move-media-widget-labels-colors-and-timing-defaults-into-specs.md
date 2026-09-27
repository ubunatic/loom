# 141 — Move media widget labels, colors, and timing defaults into specs

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [140](140-lazy-load-media-widget-images-and-video-previews.md)

---

## 1. Problem & Motivation

`media/widget.go` hard-codes user-visible labels (`loading`, `render error`), their RGB colors, and the 50 ms loading threshold. These values are not themeable or declared in the spec system, so media styling and timing are hidden from Loom's spec-of-record.

## 2. Technical Specification / Findings

Move applicable media labels and timing defaults into an appropriate YAML spec, and move loading/error colors into `spec/themes.yaml`. Update schemas and consume the values from the media widget, following the existing embedded spec and theme patterns. Audit nearby media literals for other user-facing defaults that belong in specs; leave implementation details and API enum values in Go.

## 3. Implementation & Verification Plan

- **/goal**: Make media widget user-facing labels, colors, and behavior timing defaults spec-driven and theme-aware, validate the specs and tests, or stop and report when blocked on a user decision or denied permission.
