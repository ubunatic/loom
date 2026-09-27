# 134 — Loom output drops ZWJ joiners and pads flag rows 2 cells short

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 096, 132, 048, 133

---

## 1. Problem & Motivation
With ansiviewer's ruler (81d5d6b) the terminal background now shows where each ANSI row ends. The
user's screenshots of docs/progress/096/M2-buttons.ansi (tilix, foot, 2026-09-27) show two Loom
output bugs, which the 096 review had put down to terminal behavior:

1. **ZWJ joiners are dropped.** None of the six 096 frames contains U+200D. The family 👨‍👩‍👧 is
   written as three separate emojis, which is why every terminal drew three. Loom measures the
   cluster as 6 columns (measure.go knows ZWJ) but the output loses the joiner.
2. **Flag rows are 2 cells short.** Loom counts 🇩🇪 as 4 columns, and terminals advance 2. Row
   padding is computed with 4, so each row with a flag ends 2 cells early and the rest of the row
   shows the terminal background. The label order in M2-buttons (sample first, then label) is in the
   file itself and is correct.

## 2. Technical Specification / Findings
- Check: `grep -a -c $'\u200d' docs/progress/096/*.ansi` gives 0 for all frames.
- Loom's ANSI box check (TestAllAnsiAssetsHaveValidBoxes) also counts the flag as 4 (see 048), while
  ansiviewer's replay via ParseANSI places the next text 2 later (132): Loom is inconsistent.
- Where the joiner is lost is unknown: cluster splitting in Canvas.Write/set (textClusters), the
  cell renderer, or the recorder.

## 3. Implementation & Verification Plan
- M1: failing tests: write a ZWJ family and a flag through Canvas and the renderer, and assert that
  the bytes keep U+200D and that a flag row's output is exactly as wide as the row.
- M2: keep the joiner in the output; count a regional-indicator pair as 2 columns everywhere
  (StringWidth, box check, ParseANSI) to match the terminals tested in 096.
- M3: re-record the 096 frames; the user checks them with ansiviewer's ruler (102).
