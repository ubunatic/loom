# 483 — measure: VS16 after a non-emoji letter, marks inside Indic conjuncts, vt fuzz cursor bound

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 481, 460, 431

---

## 1. Problem & Motivation
The agy:flash38 review of adoption PR A (Codeberg PR 92, 481) left three minor findings. None
affects common text, but each makes a width or a test less exact.

## 2. Technical Specification / Findings
Found by review, not yet confirmed by a test (verify each first):

- `isEmojiPresentationBase` (`measure/measure.go`, near line 900) treats every non-ASCII rune
  other than VS15/VS16 as an emoji base, so `é️` may count 2 columns. Terminals draw a
  letter with a stray VS16 at the letter's width (1). Check foot/kitty/VTE before changing.
- `linkedConsonants` (`measure/measure.go`, near line 285) only matches consonant + virama +
  consonant directly; a nukta or other mark between them (allowed by the cluster boundary code)
  falls back to width 1.
- `vt/vt_fuzz_test.go:26` checks `cursorX > 80` / `cursorY > 24`; the 0-based bound is `>= 80` /
  `>= 24` (check pending-wrap semantics first: x may legally equal Cols while wrap is pending).

## 3. Implementation & Verification Plan
/goal Confirm each finding with a failing test, fix the ones that hold in `measure`/`vt`, and
add the cases to `TestWidthTargets`; stop and report when a terminal check is needed from the
owner, or when no progress is made (clear the goal and write a failure report).

- Re-check live code first; line numbers may have moved.
- Verify: `go test -short ./measure/ ./vt/`, then `make test-q1`.
