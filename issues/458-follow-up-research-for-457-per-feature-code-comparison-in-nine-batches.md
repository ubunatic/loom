# 458 — Follow-up research for 457: per-feature code comparison in nine batches

**Status**: In Progress — owner: run batch 1 (render) alone to check the flow
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Research
**Related**: 457

---

## 1. Problem & Motivation
Ticket 457 produced web research on TUIOS and the Charm stack (`docs/studies/tuios/`). A review by
sol:high found that several conclusions are weak or wrong and that "verified" rows only mean a quote
was found, not that the claim was checked against code. Owner request: a follow-up that
1. looks at loom's code per feature, 2. looks at the other libraries' code per feature,
3. captures their key implementations, 4. updates the finding docs and writes more reports where needed.

## 2. Technical Specification / Findings
The full plan, with review findings, source pins and per-batch steps, is in
`docs/studies/tuios/followup-plan.md`.

Main review corrections to carry into the docs:
- loom has no frame diff (`canvas.go` `FlushWithConfig` writes every row) but redraws only when dirty; no measured slowness.
- TUIOS `frameRenderer.diffs` is SGR caching, not a cell diff.
- `measure/zwj.go` measures ZWJ advance by cursor report; the missing piece is the 2027 query.
- `Pane.Post` has an unbounded queue, batches run together, and Close can drop callbacks.
- Drag capture covers scrollbar and pan only, not selection outside the entry.
- Table draws visible rows only but filters and sizes over all data; TextEntry paste already removes control chars.
- `vt/vt.go` keeps one kitty flag value (no stack) and replies 0 for mode 2026.

Batches (each writes only its own `report-<batch>.md`):
1. render and diff, 2. width and graphemes, 3. terminal queries, 4. input, kitty and paste,
5. mouse and drag, 6. widgets and editing, 7. session and VT, 8. theming, composition and desktop,
9. testing, lifecycle and producer load.

Models: haiku:med fetch and quote; terra:med code comparison (required for 4 and 5, event routing);
astra:high once for conflicting protocol evidence. Live terminal and clipboard canaries run one at a
time; one integrator merges results into `features.md`, the `features-*.yaml` loom column and `summary.md`.

## 3. Implementation & Verification Plan
- M0: owner picks the batches (all nine, or a subset such as 1, 3, 4).
- M1: run the chosen batches by the plan's verification rule V and stop rule S.
- M2: integrator merges, recounts rows, rejects non-verbatim quotes and moving links; every loom status
  carries revision, file and symbol evidence; unrun canaries stay explicit.
- Research only: no loom code changes, no tickets until the owner reviews the results.

/goal Deliver the chosen batch reports and the merged, evidence-backed feature list and summary in
`docs/studies/tuios/`, then stop; or stop and report when blocked on owner batch choice or a denied
permission. On a loop or no progress, clear the goal and write a failure report here.
