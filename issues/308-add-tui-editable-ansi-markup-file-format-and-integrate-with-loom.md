# 308 — Add .tui editable ANSI markup file format and integrate with loom

**Status**: Open — In progress
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**:

---

## 1. Problem & Motivation
Add a `.tui` ANSI Markup Language file format so ANSI styling can be authored and edited in readable text files, then rendered and integrated with Loom.

## 2. Technical Specification / Findings
The syntax is intentionally undecided. First check whether an existing clean format or library can meet the need. Then discuss markup designs with the user before selecting a format; record findings and decisions as the work progresses.

## 3. Implementation & Verification Plan
**Phase 1:** Research related formats and implementations, identifying whether one can be reused cleanly.

**Phase 2:** Present viable markup designs and discuss them with the user; do not commit to syntax before that discussion.

**Phase 3+:** Implement the agreed `.tui` format and integrate parsing/rendering and authoring workflows with Loom. Verify round-trip and rendered ANSI output.

**/goal** Research reusable approaches, agree on a markup design with the user, then implement and integrate `.tui` with Loom; stop and report if blocked on user input or denied permission.
