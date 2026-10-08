# 311 — Support embedded Cati images in the .tui format

**Status**: Draft
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**:

---

## 1. Problem & Motivation
Support embedded Cati images in `.tui` documents so image content can participate in ANSI markup alongside text.

## 2. Technical Specification / Findings
Use the image capabilities in the sibling `../cati` project as the integration reference. Decide the embedding representation and runtime boundary during implementation; avoid duplicating Cati image functionality unnecessarily.

## 3. Implementation & Verification Plan
Design and implement `.tui` support for embedded Cati images, then verify image loading and rendering alongside surrounding text and markup.

**/goal** Add and verify embedded Cati image support in `.tui`, or stop and report if blocked on a user decision or denied permission.
