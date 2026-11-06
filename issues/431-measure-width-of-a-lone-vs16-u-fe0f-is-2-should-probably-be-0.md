# 431 — measure: width of a lone VS16 (U+FE0F) is 2, should probably be 0

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 414 (shared cluster scanner, 8397a50)

---

/goal Decide and test the width of a lone VS16 and of VS16 after a non-emoji base (`a️`), checked against real terminals (foot, kitty), so `StringWidth` and `ClustersNew` agree on the right value; stop and report when blocked on a user decision.

## 1. Problem
Before 414, `StringWidthNew("️")` was 0 and `ClusterWidth` of the same cluster was 2. 414 unified both on 2 (Jules PR 109). Host probe on 8397a50: `"️"` = 2, `"a️"` = 2, `"#️"` = 2, `"❤️"` = 2. Most terminals likely show a lone VS16 as width 0 and `a️` as width 1 (unverified).

## 2. Next step
Probe foot and kitty with `loom measure` against what the cursor does, then fix `ClusterWidth` and extend `measure/measure_parity_test.go` with the measured values.
