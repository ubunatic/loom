# 108 — Fix loomoji hover highlight one row above pointer (dy=-1)

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [107](107-pty-mouse-test-system-detect-loomoji-hover-vs-highlight-offset-via-rendered-fg-bg-cells.md),
`examples/loomoji/loomoji/loomoji.go`, `examples/loomoji/loomoji_107_pty_test.go`,
`docs/HoverTesting.md`

## /goal

Hovering a loomoji grid cell highlights the item under the pointer.
`TestLoomoji107HoverProbe` passes (dominant offset (0,0), no `OFFSET_BUG`),
and the rest of the suite stays green.

## Evidence

The 107 probe measured (dx=+0, dy=-1) across 10/10 measured probes: the
highlight lands one grid row above the hovered cell. Likely suspects are the
mouse→grid mapping around the top padding row and search bar (95926f9,
b71f8c9, 30f558e). Unverified; find the root cause first.

## Constraints

- Fix the mapping in loomoji. Don't change the probe test to make it pass.
  Only adjust its expectations if a rendering change legitimately moves items.
- Clicks and drag use the same mapping; they must land on the same item.
