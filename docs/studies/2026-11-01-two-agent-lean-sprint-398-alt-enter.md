---
kind: sprint
---
# Lean sprint 398 — Alt+Enter decoding (2026-11-01)

- **Models**: searcher haiku:low (33s), developer haiku:med (plan turn + write turn), reviewer opus:low.
- **Searcher**: located the generic alt fallback (`event.go:217`) and the a..z-only `decodeModifyOtherKeys`; accurate, no host digging needed.
- **M1 (decode both forms as alt-enter)**: `86c0181`. ESC CR → `alt-enter`; modifyOtherKeys code 13 → `alt-/shift-/ctrl-enter`. Repro test failed on the 4 predicted rows first. Upgrading.md v0.5.7 breaking row.
- **Verification**: root package green; `make test-q1` red only on `TestRenderEmojiGrid`, which the host showed fails on the parent too → ticket 402.
- **Review**: approve; nits — untested `ctrl-shift-enter`/LF-code near-misses, dead `b[1] != 13` guard, lone-ESC split caveat (pre-existing).
- **Flow**: plan-first worked; haiku:med plan was precise enough to approve unchanged.
