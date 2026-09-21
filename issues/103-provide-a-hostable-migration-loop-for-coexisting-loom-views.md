# 103 — Provide a hostable migration loop for coexisting Loom views

**Status**: Open

---

Reserved placeholder ticket.
# 103 — Port minimal colored `harnez usage --compact --watch`

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature

## Goal

Port the complete `../harnez` `harnez usage --compact --watch` execution path
into a minimal, standalone Loom example. It must be runnable and fully colored,
with the two All Usage and Load boxes, asynchronous data collection, an
independent redraw loop, terminal resize/input handling, and realistic local
data for Load plus fake All Usage data.

The port should preserve the useful Harnez behavior rather than replacing it
with a placeholder renderer. Once the colored clone is working and covered by
tests/ANSI evidence, use it as the host application for the Loom migration:
introduce Loom views incrementally, gate old/new views behind CLI flags, allow
both views to coexist, and provide an interactive switch between them.

The SDK changes required by the migration should be kept small and generic.
File separate follow-up issues for Loom capabilities that are discovered but
are not necessary to complete this port.
