# 207 — Richer gallery demos: animation, seconds, variants, labels

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 195, 197-202 gallery

---

## 1. Problem & Motivation
User feedback on demo content: ProgressBar and Spinner: animate or show more states. Stopwatch: show seconds, animate, and add optional controls (start/stop/reset). Timer: use seconds. NumberInput: width grows with the number ("◂ 3.5 ▸" vs "◂ 4 ▸"); show variants, e.g. fixed width. TextInput: the demo masks text like a password; show a plain input by default plus variants (placeholder, masked). Toggle: has no label. Table: non-interactive is fine unless cheap to add a cursor. Pill cluster and others are fine.

## 2. Technical Specification / Findings
Demos live in `gallery/gallery.go`; animated ones need the widget's `TickInterval`/`Tick` forwarded through Tabs and `themedGallery`. Optional Stopwatch controls and a fixed-width NumberInput may need small non-breaking library options.

## 3. Implementation & Verification Plan
Proof: each item gets a regression test that fails before the fix, driven through the real `loom widgets --show <Name>` binary in a PTY (`gallery/gallery_test.go`, `internal/ptytest`), plus a `docs/progress/<Widget>.ansi` capture where the look changes. For animated demos, assert the render changes over time.

/goal Each listed demo shows the requested states or behaviour, proven by tests and .ansi captures; or stop and report when blocked on a user decision.
