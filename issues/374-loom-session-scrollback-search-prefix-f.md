# 374 — loom session: scrollback search (prefix f)

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 371 (tabs), 373 (bar search hint), `vt` package, `SearchBar` widget

---

/goal Prefix `f` in `loom session` opens a search over the active tab's screen and scrollback, shows
matches highlighted, and lets the user step through them and close the search with the child's screen
restored unchanged; the key and labels come from the spec; PTY tests cover find, next/previous, no
match and close; or stop and report when blocked on a user decision.

## Decision (user, 2026-10-23)

Search is scrollback search, on prefix `f` (not `ctrl-f`, which shells and `less` use).

## Open questions

- `vt.VT` has no scrollback yet: lines that scroll off the top must be kept (size limit in spec).
- Alt-screen children (mc, htop) have no scrollback; search only their visible screen?
- UI: reuse the loom `SearchBar` in the session bar row; scrolled view while searching (like tmux copy
  mode); copy of the match is a later step.

## Decisions (host, 2026-10-23, user: "decide yourself")

- `vt.VT` keeps scrollback for the main screen: lines that scroll off the top are kept, limit from the
  spec (`session.scrollback_lines`, default 5000), ring buffer, no growth past the limit.
- Alt-screen children (mc, htop) have no scrollback; search covers their visible screen only.
- UI: prefix `f` turns the bar row into a search prompt (reuse `SearchBar` look and keys); the view
  scrolls to the match like tmux copy mode; next/previous and close keys from the spec; Esc closes and
  repaints the child's current screen unchanged. Child output during search updates its VT only.
- Copy of the match is out of scope. The bar's search hint (373, hidden) becomes visible.

## Approved plan (host, 2026-10-23)

d374 plan approved (opt-in VT ring, snapshot on entry, SearchBar in the bar row, local routing during
search, milestones VT / spec+UI+routing / PTY+visual). Pre-work:

- Previous match: `shift-enter` is not distinguishable in most terminals; use `up` (older) and `down`
  (newer) plus `enter` = older; keep keys in the spec.
- Smart case: case-insensitive unless the query has an uppercase letter.
- Global keys (`ctrl-q` quit) still work during search; `ctrl-n`, `ctrl-p`, alt digits and the prefix are
  swallowed while searching.
- When done with 374, also fix issue 375 (help cut off on short terminals) in its own commit
  `(issue 375)` with a PTY test at 12 rows; the footer stays visible.

## M1–M3 delivered (host review, 2026-10-23)

df7c630, 6571f76, 5f0d0e5, 3eb6afa; 375 in 896787b. Host q1 green. Live tmux 80x12: prefix f, "line7"
scrolls to the match, Esc restores, 12-row help fits in two columns.

### M4 pre-work (required)

- Search hint renders raw names next to keycaps: `enter/down Enter/↓ older  up ↑ newer  esc Esc close`.
  Show keycaps only (`Enter/↓ older  ↑ newer  Esc close` style, via KeyCap); PTY test asserts the bar text.
- Direction is swapped against the approved plan: `up` must go to older matches (upwards in history),
  `down` to newer; `enter` = older. Fix spec and test.
