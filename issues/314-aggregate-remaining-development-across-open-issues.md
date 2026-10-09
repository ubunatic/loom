# 314 — Aggregate remaining development across open issues

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [313 — Manual acceptance checklist for open issue follow-ups], [014], [090], [095], [116], [120], [133], [139], [147], [157], [241], [252], [259], [265], [266], [274]–[290]

---

## 1. Problem & Motivation
A read-only source review of the open backlog found implementation gaps spread across older library/tooling tickets and the newer editor/widget cluster. Some tickets appear implemented or nearly implemented, while others need substantial development or a deeper audit. This ticket aggregates the development work so it can be planned without treating every open ticket as equally unfinished.

**Goal**: Reconcile the provisional work list below against the live source and linked tickets, then ensure each genuine implementation task is completed in its source ticket or explicitly ruled out as no longer needed. Stop and report when scope, compatibility, or a human decision blocks progress; do not close source tickets solely by completing this aggregation.

## 2. Technical Specification / Findings
The list is based on a targeted, read-only scout assessment; related symbols were not treated as proof of completion, and no tests or acceptance probes were run. Re-check the live issue and code before starting. Keep detailed requirements and implementation in the source tickets; this ticket is an index, not a replacement specification.

### Development work indicated by the review
- [ ] **014 — Graph presentation:** Implement/verify configurable colors and glyph semantics across graph types, including style-leak prevention and geometry/sample invariance.
- [ ] **090 — Image backgrounds:** Complete any remaining loading dimming, animated-cell average-color continuity, theme switching, focused tests, and PTY coverage.
- [ ] **095 — Observable PTY flow:** Implement the paced human-observer mode and feedback/title/editor/draft workflow; existing PTY infrastructure alone is insufficient.
- [ ] **116 — ANSI subpackage:** Audit the current package/API first; if still missing, extract the requested zero-allocation styling/parsing APIs and row optimization.
- [ ] **120 — Tree-Sitter engine:** Implement the requested wazero-backed engine, embedded grammar/query assets, and integration on top of the existing syntax foundation.
- [ ] **133 — Pane debug ruler:** Verify current naming/implementation, then implement the debug mode and ruler overlay if absent.
- [ ] **139 — `loom play`:** Implement the requested interactive TUI command and ANSI capture workflow, reusing existing child-TUI recording where appropriate.
- [ ] **147 — Box validation:** Verify existing validation against the acceptance criteria; add unclosed-box detection and error prioritization if still missing.
- [ ] **157 — Golden comparator:** Implement the ANSI golden-mockup visual comparator and mismatch tests if absent.
- [ ] **241 — Grab package:** Implement the public grab package, SDL helper, demo, tests, and required platform acceptance in the source issue.
- [ ] **252 — Spec-default audit:** Record accepted/rejected candidates and reasons, migrate accepted defaults, and verify spec integrity; locate any existing audit record before duplicating work.
- [ ] **259 — OSC 8:** Implement OSC 8 emission through rendering/cell diff; coordinate required real-terminal click verification via [313].
- [ ] **265 — RichTextEdit context menu:** Implement the requested right-click/context and clear-style/popover behavior not already present.
- [ ] **266 — Box push mode:** Implement the separate S-F5 box-push behavior while preserving existing characters per the issue criteria.
- [ ] **274–277 — Widget architecture:** Audit the actual general-purpose overlay API, draw/focus coupling, container-owned mouse translation/clipping, and click-only focus invariants; implement only the gaps confirmed against each source ticket.
- [ ] **282–284, 287–290 — Editor follow-ups:** Reconcile partial implementations with current acceptance; finish file-browser/search/screenshot/panel behavior and tests, search selection/mouse isolation, and deterministic ANSI-design test/evidence workflow as needed.

### Likely acceptance-first rather than new development
Issues **137, 161, 279, 280, and 281** appeared substantially implemented in the scout review. Use [313] to perform remaining CLI, visual, PTY, and configuration checks; avoid duplicating implementation unless those checks expose a concrete gap.

## 3. Implementation & Verification Plan
For each unchecked entry, inspect current code, recent history, and its source issue before acting. Record the actual remaining work and verification in that issue; implement there or split out a focused ticket if scope warrants. Run relevant project tests and issue-specific checks after implementation, and coordinate human/terminal checks through [313]. Close this aggregation only when every item is completed or explicitly resolved as unnecessary, with the disposition recorded in its source ticket.
