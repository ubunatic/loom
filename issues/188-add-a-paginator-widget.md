# 188 — Add a Paginator widget

**Status**: Closed
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: choice.go, table.go, view.go, issues/177, issues/180

---

## 1. Problem & Motivation
Long lists and tables only scroll; there is no page indicator or page control like Bubbles' Paginator (issue 177).

## 2. Technical Specification / Findings
`Paginator` widget: `Page`, `Pages`, dot style (`● ○ ○`) and numeric style (`2/7`); PgUp/PgDn and click to change page; an `OnChange` callback so a host can drive a `Choice`/`Table` offset. No change to Choice or Table themselves.

## 3. Implementation & Verification Plan
/goal Ship `Paginator` with key, mouse (0-based child-local), and render tests and a docs row; stop and report if driving Choice/Table requires changing their public API.
