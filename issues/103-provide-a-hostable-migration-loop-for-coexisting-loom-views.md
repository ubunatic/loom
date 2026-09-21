# 103 — Provide a hostable migration loop for coexisting Loom views

**Status**: Open

---

Reserved placeholder ticket.
# 103 — Provide a hostable migration loop for coexisting Loom views

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature

## Goal

Expose a supported host API that lets an inline or altscreen Go TUI mount
multiple Loom widgets, share one event/redraw loop, and switch active views
without each child owning `/dev/tty` or starting its own `Pane.RunWatch`.

The monitor migration example currently proves the need with a local adapter;
promote the lifecycle and view-switching contract into the SDK and cover
coexistence, resize, input routing, and cancellation with deterministic tests.
