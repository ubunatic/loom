# 190 — Add a Viewport widget that scrolls arbitrary child content

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: issues/156, view.go, issues/177, issues/178, issues/180

---

## 1. Problem & Motivation
`View`, `TextArea`, `Choice`, and `Table` each scroll their own content, but there is no way to scroll an arbitrary child that is taller or wider than its slot — the ncurses pad and the Bubbles viewport (issues 177, 178).

## 2. Technical Specification / Findings
`Viewport` wraps a child: draws it into an off-screen canvas of the child's measured size (uses 156's `SubCanvas`/`Blit`), shows the visible window, scrolls with arrows/PgUp/PgDn/Home/End and mouse wheel, and translates mouse events to the child's coordinates (0-based child-local, docs/Widgets.md). Optional scrollbar.

## 3. Implementation & Verification Plan
/goal Ship `Viewport` with scroll, clamp, and mouse-translation tests and a docs row; stop and report if 156 is not done or a child cannot measure its full size.
