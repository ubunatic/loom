# Lean sprints 2026-10-22 (second half): reference crops, PTY flake, spec shadows, key bindings, v0.4.0

## 1. Header & Context

- **Date:** 2026-10-22, after v0.3.11.
- **Scope:** tickets 365 (`loom crop`/`loom fmt`, `loomtest.CompareWidget`), 352 (edit help PTY flake),
  a `/respect` spec audit, 366 (spec shadows), 323 (core widget bindings), 367 (media bindings), release
  v0.4.0. Earlier tickets of the day (353–364) are in the 2026-10-21/22 sprint studies and evergreen
  notes 00bca37, d8861d5.
- **Setup:** Claude Opus host orchestrator, no host code; developers via `harnez agent start`
  (codex:sol:med for 365, 352, 366; flash38:med for 323 and 367, chosen by the user). The host
  reviews diffs, reruns `make test-q1`, checks live, closes, pushes, rates and deletes each agent.

## 2. Executive Summary

Six tickets closed, two filed from the audit (366, 367), 323 extended instead of duplicated, one
filed from a live check (368), and v0.4.0 released to Codeberg with signed Linux artifacts. Every
ticket ended with host q1 green and a regression test shown failing before the fix. flash38:med did
two spec migrations at sol quality on clear tickets.

## 3. What Worked Well

- **Audit before tickets.** One read-only advisor (2.5 min) produced a ranked table; the host split
  real shadows (366, init panic when YAML changes) from new spec content (323/367) and asked the user
  once.
- **Plan-only first turn for 323.** The plan named loader, validation, test probe and milestones; the
  host recorded it with three pre-work rules in the ticket, then granted writes. Four clean milestone
  commits followed with no rework round.
- **"Fails before the fix" held.** All three spec tickets showed their YAML-change probe failing on
  unfixed code (366: init panic; 323: tabs ignored YAML; 367: media). 352 proved the race with a VT
  replay when stress could not reproduce it.
- **Old-vs-new binary comparison.** For a doubtful live result (View paging), building b7d2fb7 in a
  scratch worktree settled within a minute that 323 was not the cause.

## 4. Honest Post-Mortem

- **Stale "Unreleased" rows.** At release time 20 rows in `docs/Upgrading.md` still said
  "Unreleased", most shipped in v0.3.1–v0.3.11. Caught while preparing v0.4.0; fixed with `git blame`
  plus `git tag --contains` per row (475b180). Releases did not relabel rows before.
- **Commits without issue numbers.** `issues close` warned three times (366, 323, 367) that no code
  commit references the ticket; the developer prompt only asked for a conventional commit. Fixed by a
  prompt line requiring "(issue N)" in the subject.
- **Live check started wrong.** The first tmux check of `loom view` used a `.txt` file; view accepts
  only `.ansi` and exited at once, which looked like an instant quit. Seen with `remain`-style
  `echo EXIT`; redone with `.ansi`.
- **Pre-existing bug found, not caused.** Space and Ctrl+F do not page in `loom view` on both binaries;
  filed as 368 rather than folded into 323.
- **Small code smell accepted.** 323's Table keeps `t.keys == nil` fallback branches that cannot run
  (the constructor always sets a KeyMap) and a `page-up` alias. Rated 4/5 instead of fixing in a
  micro-round.
- **Sweep hazard (harnez 859)** from earlier in the day remains open: `issues open --commit` commits
  every staged file. Nothing was swept in this half; the host's own ticket commits were made with no
  developer between `add` and `commit`.

## 5. Quality & Invariants Audit

| Aspect | Result |
|---|---|
| Architecture | Bindings live in `spec/defaults.yaml` + schema, loaded once into `SpeccedDefaults` with validation; widgets use `KeyMap`/`Is` on specced slices; no mirrored Go tables left in scope |
| Idempotency | `loom fmt` is encoding-only and idempotent (test proves `cat` output unchanged); validate-spec clean |
| Backward compatibility | Default keys and values unchanged (live j/q and per-widget tests); 366 turned three exported constants into vars, recorded in Upgrading.md; v0.4.0 is a minor bump for breaking changes |
| Tests | Each ticket: regression failing first, package tests, host `make test-q1` green (0 `--- FAIL`); 352 also `-race -count=20` on 2 CPUs |

## 6. Efficiency & Velocity

- Spec audit 2.5 min; 366 dev 4.3 min; 323 plan + 4 milestones in two turns; 367 one turn. Host q1
  reruns are the longest step per ticket.
- No rework rounds on 366, 323, 367. All four agents rated 4/5.
- flash38:med matched codex:sol:med on well-specified spec migrations at lower cost.

## 7. Key Learnings & Evergreen Upstream

- Developer commits carry "(issue N)" in the subject (done: `.harnez/prompts/developer.md`).
- Release step: relabel `Unreleased` rows in `docs/Upgrading.md` with the new version before
  `harnez release` (candidate for docs/GoRelease.md or a release check that fails on stale rows).
- When a live check looks wrong, compare against the binary from before the change before blaming it.
- A spec audit should separate shadows (bug, fix now) from new spec content (design, own ticket).

## 8. File & Diff Summary

| Ticket | Commits | Key files |
|---|---|---|
| 365 | 25b0180, 224c13d, 56920d8 | `loom crop`, `loom fmt`, `loomtest.CompareWidget`, `testdata/ColorPicker-*.ansi` |
| 352 | b3eef58 | `cmd/loom/edit_302_test.go` |
| 366 | 985e7ee | collector, graph, measure, `spec/catalog.go`, `internal/specdefaults` |
| 323 | 1e48bcb, 9550a25, 0758526, 333e0be | `spec/defaults.yaml`, schema, `defaults.go`, nine widgets, `key_defaults.go` |
| 367 | 5c5adc9 | `media/widget.go`, `spec/defaults.yaml`, schema |
| release | 475b180, f2448c2 (tag v0.4.0) | `docs/Upgrading.md`, `version.yaml`, `version.go` |
| filed | 366, 367, 368 | `issues/` |
