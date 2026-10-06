# 236 — Gallery status bar truncates the F10 Quit hint at 100 columns

**Status**: Closed — fixed by 278 HintBar; F10 Quit visible at 100 cols with astra (on redraw)
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 233 (astra on-redraw mode)

---

/goal Keep the gallery's quit hint visible at 100 columns with every F8 background label, with a test,
or stop and report when blocked on a user decision (e.g. which hint to shorten).

## 1. Problem & Motivation
`loom widgets --show <Demo>` at 100 columns: with F8 set to "astra (on redraw)" the status bar is
too long and the trailing "F10 Quit" hint is cut off (seen in the 233 host PTY check). Users lose
the quit hint exactly in the mode they just picked.

## 2. Technical Specification / Findings
Status text is built in `cmd/loom/widgets.go` (gallery status bar, `galleryBackgrounds` labels).
Options: shorter label (e.g. "astra/redraw"), drop lower-priority hints first, or keep the quit
hint right-aligned and truncate the middle. Prefer a rule over a shorter string.

## 3. Implementation & Verification Plan
- Fix the status layout; test that "F10 Quit" is present at 100 columns for all three F8 states.
- `make test-q1`, `make install`, PTY check of `loom widgets --show Dialog` at 100 columns.
