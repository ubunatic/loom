# 313 — Manual acceptance checklist for open issue follow-ups

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Documentation
**Related**: [314 — Aggregate remaining development across open issues], [090], [102], [137], [161], [259], [279]–[290]

---

## 1. Problem & Motivation
A source-only scout review found several tickets with human, terminal, visual, or end-to-end acceptance still outstanding. Keep one actionable checklist of checks for the project owner to perform; do not close related tickets based only on code presence or this checklist's creation.

**Goal**: Complete and record every applicable check below, or leave unchecked items with a brief reason/blocker. Stop and report if a check requires unavailable hardware, permissions, or a user decision.

## 2. Technical Specification / Findings
This list is based on a recent read-only issue status review. Confirm each referenced ticket's live acceptance criteria before testing; the review did not run tests or PTY sessions. Mark each item only after performing it, and record evidence (environment, command, screenshot/log path, or concise result) in the corresponding ticket or below.

### Human and real-terminal checks
- [ ] **102 — Manual review collection:** Complete each outstanding checklist item in [102], including evidence-frame checks, scrollbar/file-browser mouse behavior, ticker, keys, boxes, treemap, ansiviewer, and process cleanup. Record results in 102.
- [ ] **259 — OSC 8 hyperlinks:** Verify actual terminal hyperlink emission and clickability in supported real terminals; record terminal/version and outcome in 259. Source styling alone is not acceptance.
- [ ] **090 — Background/theme behavior:** Verify loading dimming, average-color continuity during animation, theme toggle behavior, and focused PTY/real-terminal behavior; record evidence in 090.
- [ ] **161 — Media redesign:** Compare the example against `docs/data/media-demo-redesign-005.ansi` at the target size and a narrow terminal size; record visual results in 161.
- [ ] **137 — CLI asset workflow:** Re-run the current CLI asset workflow/checks and verify the feedback disposition before deciding whether 137 can close.

### Editor end-to-end / PTY acceptance
- [ ] **279 — Editor command:** Verify installed-binary launch, help/completion/man behavior, PTY behavior, and terminal restoration as specified in 279.
- [ ] **280 — Mouse capture:** Verify the explicit CLI override/config behavior and terminal mouse-mode restoration in a PTY/terminal.
- [ ] **281 — Settings:** Run the specified schema/config tests and verify invalid/unknown settings are rejected as expected.
- [ ] **282/290 — File browser:** Verify pointer activation, cancel/reopen behavior, and all required Save/Discard/Cancel outcomes; record any remaining failure under the correct ticket.
- [ ] **283/287 — Search:** Verify focus and mouse routing, selection isolation, stale-match handling, and documented controls using the issue acceptance criteria.
- [ ] **284 — Screenshots:** Verify styled/Unicode/editor-state screenshot cases and the specified PTY/ANSI behavior.
- [ ] **288 — Responsive panels:** Verify clipping, narrow layouts, standard widget composition, and persistent hint-bar hit regions.
- [ ] **289 — ANSI design safety:** Verify ordinary tests cannot overwrite tracked ANSI designs and document/run any required deterministic evidence-generation path.

### Closure discipline
- [ ] For each passing check, add the evidence/result to its source ticket; do not use this aggregation ticket as a substitute for updating acceptance notes.
- [ ] Re-read live ticket criteria before closing any ticket; close only when all required implementation and manual checks pass.

## 3. Implementation & Verification Plan
No code change is required. Work through the checklist, capture evidence alongside the referenced tickets, and leave blocked or unperformed items unchecked with an explanation. Verify all acceptance rows are either evidenced as complete or explicitly recorded as blocked before closing this checklist.
