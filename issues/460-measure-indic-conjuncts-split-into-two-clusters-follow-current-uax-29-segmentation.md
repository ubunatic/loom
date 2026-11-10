# 460 — measure: Indic conjuncts split into two clusters, follow current UAX 29 segmentation

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 433, 458, 461

---

## 1. Problem & Motivation
loom's hand-written cluster scanner in `measure/measure.go` splits the Devanagari conjunct `क्ष`
(KA + VIRAMA + SSA) into `क्` + `ष`, width 2. Unicode 15.1 and later keep it as one grapheme
cluster (rule GB9c). x/ansi v0.11.9 (GraphemeWidth) and tcell v3.6.0, both on Unicode 17 tables
via `clipperhouse/uax29`, report one cluster of width 1 (458 batch 2,
`docs/studies/tuios/report-width.md`). Cursor movement, deletion and selection in TextEntry
then treat half a conjunct as a character. Open ticket 433 (skin-tone modifier splits from its
emoji) is the same kind of gap in the same scanner.

## 2. Technical Specification / Findings
- Segmentation and width are separate questions: one cluster can still be drawn 1 or 2 cells wide
  depending on the terminal. Decide the width per the policy in 461; this ticket fixes boundaries.
- Options: extend the scanner with the missing UAX 29 rules (GB9c, and GB9/GB11 cases from 433),
  or generate tables from the Unicode data files. A new dependency needs an explicit decision
  (`docs/Go.md`: avoid deps). Record which Unicode version the tables follow.
- Keep the documented overrides (`LOOM_ZWJ`, glyph overrides in the emoji spec).

## 3. Implementation & Verification Plan
/goal `measure.Clusters` matches UAX 29 of a stated Unicode version on a conformance fixture
set including `क्ष`, and existing width overrides still pass; stop and report when blocked on the
dependency decision or a denied permission, or when no progress is made (clear the goal and write
a failure report).

- Re-check live code first; 414 shared one boundary scanner between width and clusters.
- Add fixtures from the Unicode `GraphemeBreakTest.txt` for the chosen version (at least the
  Indic conjunct and emoji modifier lines), and the batch 2 samples.
- Check whether 433 is fixed by the same change; close or update it.
