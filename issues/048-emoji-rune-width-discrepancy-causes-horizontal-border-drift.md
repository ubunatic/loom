# 048 — Discrepancy between Unicode / Loom width calculation and terminal rendering for emoji in box titles

**Status**: Closed — VS16 is removed before terminal rendering; foot visual check and make test-q1 pass
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug / Rendering
**Related**: `measure/measure.go`, `canvas.go`, `frame.go`, `rawscreen.go`

---

## 1. Problem & Motivation

When emojis or pictorial symbols such as `🖼` (U+1F5BC `FRAMED PICTURE`) are used in a `loom.Box` title, terminal emulators (e.g. `tilix`, `foot`, VTE-based terminals) render the glyph across **1 column** when in text/monochrome presentation (East Asian Width Neutral, single column in `wcwidth`), whereas Loom's `measure.RuneWidth(r)` categorizes the range `0x1f000`..`0x1faff` unconditionally as width **2**.

Because Loom assumes the glyph takes 2 cells, it advances by 2 columns when building the row and skips one trailing border dash. When the terminal emulator renders the glyph in 1 column, all subsequent characters in that row shift left by 1 column, causing:
1. The right border of the first box to drift left.
2. The gap between boxes and the second box's top-left corner to be visibly misaligned.

---

## 2. Technical Evidence

1. In `measure/measure.go`:
   ```go
   if r >= 0x1f000 && r <= 0x1faff {
       return 2
   }
   ```
2. Character `🖼` (U+1F5BC):
   - East Asian Width: `N` (Neutral).
   - Standard POSIX `wcwidth(0x1F5BC)`: `1`.
   - Without Variation Selector-16 (`\uFE0F`), text terminals render it as a single-column monochrome box icon.
3. Because `Canvas.Row()` outputs the rune and advances Loom's grid by 2, the terminal advances its cursor by only 1, offsetting the entire rest of the line.

---

## 3. Proposed Resolution

1. Audit and refine `measure.RuneWidth` against standard Unicode East Asian Width / `wcwidth` tables and Emoji Presentation specifications (distinguishing default emoji presentation vs text presentation / VS-16).
2. For applications requiring deterministic 1-column icons in borders and titles, standard 1-column geometric/symbol glyphs (like `▣`, `▧`, `▦`, `◈`, `•`, `*`) or explicit bracket tags should be used to guarantee single-cell alignment across all terminals.

## 4. Verification after initial fix (2026-09-17)

The default-presentation width table fixes bare `🖼` in the observed terminal.
However, the two terminals used for visual checks disagree about `🖼️` (the
same base rune followed by VS16): the first capture advances one cell, while
foot advances two. Counting VS16 as an extra cell aligns foot but leaves a
one-cell gap in the first terminal; counting it as zero does the reverse. The
foot screenshot shows the second box's top-right corner one cell beyond its
side border. Automated canvas tests and `make test-q1` pass under the
current one-cell policy, so they do not establish cross-terminal alignment.

Acceptance still needed: choose and verify a portable presentation policy for
border/title content, or expose a terminal-specific width policy with a real
terminal canary. Keep an actual foot render in the verification matrix.

## 5. Resolution (2026-09-17)

Loom now removes VS16 from text before measuring and rendering it. The bare
`🖼` therefore reaches both terminals and occupies one cell in each. A box
regression asserts that both `🖼` and `🖼️` input render the same aligned top
border; a measurement regression asserts that VS16 does not reach rendered
text. `make test-q1` passed, and the user visually confirmed aligned borders
in foot after the change. Other terminal-specific glyph widths remain outside
this ticket's verified matrix.

Follow-up: `ClipRow` now applies the same text normalization before writing
RawScreen rows, while preserving ANSI styling and combining marks. Declared
scrollbar glyphs reject VS16 and ZWJ before their one-cell check; the box
border schema already restricts its glyphs to fixed choices and ASCII title
padding. A pre-fix probe confirmed VS16 leakage from `ClipRow`; the added
regressions and `make test-q1` pass after the fix.
