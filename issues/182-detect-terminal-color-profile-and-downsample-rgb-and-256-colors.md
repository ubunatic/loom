# 182 — Detect terminal color profile and downsample RGB and 256 colors

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: style.go, issues/177, issues/178, issues/180

---

## 1. Problem & Motivation
`Style.ANSI` always emits the color as given (RGB or 256). On terminals without true color, RGB themes render wrong. Bubble Tea/Lip Gloss detect the color profile and downsample; ncurses negotiates via terminfo (issues 177, 178).

## 2. Technical Specification / Findings
Detect a profile once at the output boundary from `NO_COLOR`, `COLORTERM`, and `TERM` (no terminfo database): TrueColor, 256, 16, None. Downsample RGB→256→16 by nearest palette color when emitting SGR; widgets and themes stay unchanged. Allow an override (`Pane` option or `LOOMCOLOR` env). No new dependency.

## 3. Implementation & Verification Plan
/goal Add profile detection and SGR downsampling with table-driven tests for detection and color mapping, plus a docs note; stop and report if downsampling must reach code outside the SGR emit path.
