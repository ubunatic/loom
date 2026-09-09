---
title: From Monitor Screenshots to a Declarative UI Backlog
---

# From Monitor Screenshots to a Declarative UI Backlog

## Header & Context

Date: 2026-09-09. Scope: assess Loom, preserve concrete UI targets, turn a
dictated roadmap into implementation tickets, and review the resulting handoff.

Loom started as a released Go inline-widget library with a canvas, terminal
runtime, basic layouts, and a limited YAML factory. The user's goal was broader:
declare the UI and its presentation, while Go supplies behavior and changing
data. Harnez usage/load screenshots and a Voxi monitor supplied the target.
The initial deliverable should be small: show once with a title bar, two empty
boxes, and a status bar. Implementation was outside this session's scope.

## Executive Summary

The session produced an assessed [roadmap](../Roadmap.md), a
[context attachment](../RoadmapContext.md), [visual targets](../HarnezUsageTarget.md)
with three saved screenshots, and 14 open tickets, 006–019, in the
[issue index](../../issues/README.md). The sequence progresses through watch,
responsive layout, visibility controls, geometry checks, graphs, simulated
dashboards, external sources, and a declarative-source experiment.

The main architectural decision was to keep sampling and history advancement
independent of redraw. The main planning correction was to prove the declaration
and geometry contracts early, then require both simulated targets after color
support before connecting external producers. No library implementation shipped.

## What Worked Well

Concrete screenshots exposed requirements that a generic widget wishlist would
miss: fixed graph widths during startup, stable columns at 99% versus 100%,
independent VRAM/GTT histories, ellipsis, keyboard hints, and pane reflow.

Code inspection identified reusable foundations without treating their names as
proof of the desired behavior. Render already supports headless ANSI output;
Stack and Grid allocate space but do not provide responsive dashboard layout.
Harnez has graph rendering code worth copying, while Voxi provides useful text
and monitor behavior to study.

Two subagents had distinct write scopes: one assessed the roadmap, the other
filed issues. The parent reviewed code and ticket bodies, corrected findings,
and linked the completed ticket set into the roadmap. No further agents were
started after the user requested restraint. Scope boundaries avoided competing
edits to the tracker and roadmap.

## Honest Post-Mortem (Failures, Bugs & Near-Misses)

- **An unnecessary permission gate entered ticket 012.** The draft treated
  missing source license metadata as a reason to stop copying Harnez code,
  despite the user's explicit copy instruction. Parent review caught this.
  Commit `1ba5ca1` records existing authorization and requires accurate provenance
  and preservation of notices without asking for the same permission again.
- **The original milestone claimed completeness too early.** Harnez live data
  appeared before color, and Voxi's transcript/daemon content lacked explicit
  delivery scope. Roadmap review moved complete target acceptance after color;
  tickets 015 and 016 cover Voxi and the combined milestone. Stale wording in
  ticket 015 was also corrected in `1ba5ca1`.
- **The first shell needed earlier geometry coverage.** The initial ticket
  draft left too much to the later visual gate. Review added baseline border,
  padding, clipping, and tiny-bounds checks to 006 in `1ba5ca1`.
- **Existing rendering helpers could give false confidence.** Source inspection
  found coarse rune widths, ANSI counted as visible characters, and clipping to
  canvas rather than child bounds. ANSI replay is not a raster screenshot or an
  independent terminal oracle. Ticket 010 requires independent final-position
  checks and negative controls. These defects were documented, not repaired.
- **Waiting was noisier than necessary.** The parent made several short follow-up
  waits on a running tool cell. The user asked not to start too many agents and
  to wait for them. Only two ran, both were awaited and closed, but future waits
  should use fewer bounded calls and avoid updates that merely restate waiting.
- **The context attachment is a summary, not the requested full transcript.**
  Inspection of the saved artifact shows a curated account. It should not be
  represented as a verbatim record of everything said or seen; that fidelity
  gap remains in the earlier handoff.

No data loss or runtime regression was observed. Earlier documents and images
were initially untracked; the final planning commit preserved them with the
roadmap so the ticket references would survive beyond the working directory.

## Quality & Invariants Audit

| Area | Evidence and result | Limit |
|---|---|---|
| Architecture & module separation | Tickets keep declarations, Go producers, histories, and rendering distinct; sibling code stays untouched. | Proposed architecture, not an implemented contract. |
| Idempotency | Duplicate/backlog review preceded filing; the final working tree was clean. | Filing was not rerun; clean status does not prove idempotency. |
| Backward compatibility | Changes were documentation, issue files, and PNG assets only. | Future YAML/root corrections still need compatibility tests. |
| Test coverage & verification | Filing agent reported tracker lint/index/link checks and isolated Go vet/tests passing; parent reviewed ticket bodies and whitespace. | Existing tests do not validate the unbuilt monitor or visual targets. |
| Target fidelity | Three reference images and Harnez terminal copies were saved; both target families have milestone criteria. | No rendered Loom reproduction or human visual acceptance occurred. |
| Repository hygiene | Three local planning commits; no push; clean status before this study. | The context attachment is condensed, as noted above. |

The isolated module commands use `GOWORK=off`: the earlier assessment found
that the parent go.work excluded Loom. This was an environment workaround,
not a change to the library. No additional runtime tests were needed for this
documentation-only retrospective.

## Efficiency & Velocity Assessment

The planning result spans 21 files, with 1,173 insertions and 6 deletions from
`73806bf` to `73c9f53`. It contains 14 new tickets, three planning/reference
documents, three PNGs, and an updated issue index. There were two subagents and
one explicit correction commit touching three tickets.

The three planning commits were recorded between 23:12:55 and 23:14:01 CEST.
That 66-second commit window is not the session duration: inspection, drafting,
and waiting preceded it. Token usage, total elapsed time, and a manual-work
baseline were not measured. No defensible percentage for development or
maintenance savings can be claimed yet.

The reusable output is a dependency-ordered handoff with concrete evidence and
acceptance checks. Delegation separated responsibilities, but the filing pass
remained the bottleneck and parent review was necessary. Ticket count and line
count measure output volume, not delivered UI capability.

## Key Learnings & Evergreen Upstream

Candidate rules for future evergreen updates; this session did not change
AGENTS.md or managed conventions:

- Carry prior user authorization into delegated work; distinguish attribution
  work from a new approval requirement.
- Prove the smallest consumed declaration schema in the first visible example.
- Test final terminal positions independently of production width helpers;
  include a known-broken case that the test must reject.
- Advance histories on samples, never on frames. Verify mismatched rates with
  fake time and counters before adding real I/O.
- Give each reference application explicit coverage before calling a milestone
  complete. A graph demo does not establish transcript or daemon behavior.
- Cap concurrent agents to useful independent scopes, respect user steering,
  and wait with bounded intervals instead of repeated short checks.
- Label summaries as summaries. Preserve a transcript separately when the user
  asks for an exact session record.

## File & Diff Summary

- `docs/Roadmap.md`: assessed stages, current-code findings, and ticket mapping.
- `docs/RoadmapContext.md`: condensed product and session context.
- `docs/HarnezUsageTarget.md` and `docs/assets/*.png`: visual references.
- `issues/006-*.md` through `issues/019-*.md`: implementation and exploratory
  tickets; `issues/README.md` indexes them.
- This study and `docs/README.md`: retrospective and its documentation index.

Planning commits:

| Commit | Change |
|---|---|
| `eae99a1` | Filed tickets 006–019 and updated the issue index. |
| `1ba5ca1` | Corrected early geometry coverage, copy authorization, and Voxi wording. |
| `73c9f53` | Saved the assessed roadmap, context, target documentation, and images. |

No source files were changed or deleted. The study and index are recorded in
a subsequent documentation commit.
