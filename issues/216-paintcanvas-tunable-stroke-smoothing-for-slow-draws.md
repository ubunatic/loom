# 216 — PaintCanvas: tunable stroke smoothing for slow draws

**Status**: Closed — fixed with tests, suite green
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Enhancement
**Related**: 208

---

## 1. Problem & Motivation
User feedback: slow drawing produces jagged, stair-step lines because only the latest pointer delta is used. Example slow stroke:
```
   ⡏
  ⡏⠁
  ⡇
  ⠁
```
should roughly become:
```
  ⢠⠋
 ⢠⠋
 ⡇
 ⠁
```

## 2. Technical Specification / Findings
Keep the last stroke segment(s)/curve in memory a bit longer and interpolate or smooth through them, so slow strokes come out smooth. Do not overshoot: small rectangles drawn in one stroke must keep their corners. Make the smoothing a widget option (e.g. window length / strength). The demo gets a "tune" button that cycles or adjusts the option and shows the current value.

## 3. Implementation & Verification Plan
Unit tests: a slow diagonal point sequence renders without stair steps; a small rectangle stroke keeps its corners; option 0 means today's behavior. PTY test for the tune button.

/goal PaintCanvas has a tunable smoothing option with a tune button in the demo, proven by tests; or stop and report when blocked on a user decision or denied permission.
