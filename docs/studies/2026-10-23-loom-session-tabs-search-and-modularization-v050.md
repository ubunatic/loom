# loom session: tabs, search and modularization (v0.5.0)

## 1. Header & Context

- **Date:** 2026-10-23, one host session (Claude Opus 5.5 as Host Orchestrator, lean sprints).
- **Start:** loom v0.4.0, no way to wrap a running TUI. The user's ask: run `harnez usage -w --minimal`
  inside a tmux-like `loom session` and grab hotkeys to work with its ANSI output, starting with a
  screenshot.
- **Goals (grew during the day):** screenshot and quit; tabs; mouse-safe terminal restore; bar from
  the user's design; help; mouse toggle; scrollback search; then modularization and a release.

## 2. Executive Summary

Tickets 369–376 and 378 were delivered and closed, and v0.5.0 was released. `loom session -- <cmd>`
now gives tabs (`ctrl-n`, `^B` prefix, `alt-1..9`), a bar with hints and a mouse icon, `^B ?` help,
`^B m` mouse toggle, `ctrl-f` scrollback search, and screenshots (`ctrl-p`). A new `terminal` package
restores every mode on all exit paths. Session and editor moved to `internal/session` and
`internal/editor`. Three harnez tickets (872, 876, 877) and one loom test roadmap (377) carry the
follow-ups. 56 commits; v0.4.0..v0.5.0: 91 files, +7,844/−1,981.

## 3. What Worked Well

- **Plan-first developer turns:** every sol plan came back in 1–2 minutes and needed only small host
  edits (key direction, smart case, load limits) before writes were granted.
- **Live tmux checks after green tests:** each found something tests missed (stale bar after quit,
  alt screen left on with one tab, SIGTERM exit 0, doubled key names, swapped search direction, help
  cut off at 12 rows). Each went into the ticket as pre-work and came back fixed in one round.
- **Spec as source of truth:** every key, label and glyph lives in `spec/defaults.yaml`; changing the
  prefix (370) and adding `ctrl-f` (376) needed no Go key tables.
- **codex:sol:med** as developer: 373 and 374 in ~18 minutes each, 378 in ~9 minutes, ratings 4–5.

## 4. Honest Post-Mortem

- **Host overloaded the user's machine.** The host reran `make test-q1` (unbounded `go test ./...` with
  PTY tests) after every ticket, right after the developer's own full run. One run maxed CPU and swap
  (7.8/8 GiB). Caught by the user interrupting. Fix: targeted package runs, full runs only with the
  user's go (377, harnez 876).
- **Opaque chained commands.** Host review commands chained diff, q1, install and a tmux script; the UI
  showed one abbreviated line. The user asked for one step per command or a shown script.
- **A flash agent judged stuck while working.** `agy:gemini-3.8-flash` (d373) showed only `start` in
  its job log for 3 minutes and was replaced by sol. Its own log later showed 42 normal steps. agy
  prints no progress to harnez. Filed harnez 872 (opt-in liveness, no noise).
- **Host proposal passed off as a decision.** Prefix `f` for search (avoiding `ctrl-f`) was a host
  idea in a list the user accepted; the user later preferred `ctrl-f` (376).
- **Skipped full run by a developer (370):** `defaults_test.go` still expected `ctrl-]`; host q1 caught
  it.
- **Cold-cache probe (378):** the "only 2 packages re-run" check re-ran 23, likely because other
  packages had not run since the move. Parked in 377; the full post-move run is also still open.

## 5. Quality & Invariants Audit

| Area | Assessment |
|---|---|
| Architecture & modules | Session/editor out of `cmd/loom` into `internal/`; new public `terminal` and `vt` packages. Root package (29k lines) still a monolith that makes most packages re-run (377). |
| Idempotency | `terminal.Restorer` runs once (sync.Once); restore is safe on repeated or nested exit paths. |
| Backward compatibility | Additive APIs only; `loomtest` keeps aliases for moved VT types; no `docs/Upgrading.md` row needed. |
| Tests & verification | ~42 session PTY tests, VT scrollback unit tests; host q1 green on 369–375; 376 developer q1 + user live check; 378 package tests only, full run pending (377). About 100 `time.Sleep` calls in tests break the TestTiers rule (377). |

## 6. Efficiency & Velocity

- 9 loom tickets in one day, each one plan turn, one write turn, at most one fix round.
- sol plan turns: ~1–2 min; write turns: 5–18 min. The flash turn produced only a handoff.
- Host cost: repeated full test runs were the main waste, in machine load and wall time.
- `cmd/loom` test time grew 30 s → 53 s with the session tests; only 11 of 39 packages were cached on
  a typical run.

## 7. Key Learnings & Evergreen Upstream

- Host reruns only touched packages after a developer's green full run (harnez 876).
- New features go into their own package, not `cmd/<app>` or the root (harnez 877).
- Read a silent provider's own log before declaring an agent stuck (harnez 872).
- Flag host proposals as choices with their trade-off when asking the user.
- Keep live tmux checks; they found a bug in most tickets despite green tests.
- Recorded in `docs/Session.md` (new) and `docs/LeanSprints.md`.

## 8. File & Diff Summary

- New: `terminal/`, `vt/` (from `loomtest`), `internal/session/`, `internal/editor/`,
  `internal/clitest/`, `docs/Session.md`, `docs/data/loom-session-bar-design-doc.ansi` (user design).
- Changed: `spec/defaults.yaml` + schema + `defaults.go` (session section), `keycap.go`, `event.go`,
  `cmd/loom` (thin commands), `pane_terminal.go` (restore), probes, `testdata/api`.
- Tickets: 369–376, 378 closed; 377 open; harnez 872, 876, 877 open.
- Release: v0.5.0 (cebe30b), signed Linux x86_64/aarch64 archives on Codeberg.
