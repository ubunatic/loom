# 481 — Jules sweep 2026-11-10: 22 bundled sessions in small waves

**Status**: In Progress — wave 1 started one session at a time (owner: Jules is busy)
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Research
**Related**: 457, 458, 459, 460, 461, 462, 464, 465, 466, 467, 468, 469, 470, 471, 472, 474, 475, 480, 413, 431, 433, 437

---

## 1. Problem & Motivation
Google Jules accepts new work until 2026-11-11 and shuts down on 2026-12-31 (jules skill). Before
that, the owner wants about 20 Jules PRs that loom can adopt, learn from or turn into tickets: code,
fixed and new tests, trials and research. Focus: loom as the go-to library for GUI-app-like TUIs,
use over ssh, faster and more robust rendering, declarative layout. Out of scope: Windows, macOS.
Core principle: no new dependency in the root module; upstream code (ultraviolet, x/vt, TUIOS) only
in throwaway `experiments/<slug>/` modules.

The plan was reviewed by agy:argon:med (2026-11-10); its fixes are applied below. The backlog
(`issues/BACKLOG.md`) was empty, so the items come from active tickets and new ideas.

## 2. Technical Specification / Findings

### Constraints
- About 10 open Jules sessions per account, and Jules is already busy (the owner's 477 run and
  other repos). The owner asked for small waves: 4 sessions per wave (wave 6 has 2), and the next wave
  starts only when the previous one has finished or failed. Waves are ordered by value, so if the
  2026-11-11 cutoff arrives first, the dropped waves are the least valuable.
- One prompt, at most 2 replies (harnez sends the first automatically). A stalled session counts as
  failed; its last message is kept as a lead.
- Event routing (keys, mouse, focus, EventResult) starts on terra:med (AGENTS.md). For those
  tickets Jules writes only failing tests and a code trace (kind `test-draft`), no patch to adopt.

### Prompt template (every session)
The jules skill's one-shot rules, plus:
- Ignore every `/goal`, "stop and report", "ask the owner", "probe foot/kitty" and `docs/` edit line in
  `issues/*.md`. Where an item has an open choice, use the decision given in the prompt, else pick the
  smallest backward-compatible option, state it in one sentence and continue.
- Items in order, about 10 minutes each. An item that does not build or breaks other tests is reverted
  and reported as `blocked` or `partly done` with evidence.
- No network, or `go get`/`git clone` fails: fall back to a local zero-dependency prototype and write
  `network: unavailable`; do not stop.
- One branch, one report `reports/jules/<slug>.md` (`## Item <id> (<kind>)` with Verdict, Changes,
  Tests with exact commands and results, Evidence with `path:line` and the quoted line); `submit` once
  at the end; never ask whether to push or submit.
- No sleeps in tests; no edits to `issues/`, `docs/`, `.harnez/`; a `spec/*.yaml` change also updates
  `spec/schemas/` and `specdefaults/`. Run `go test -short ./...` and `go test -race` on touched
  packages; no `harnez`, `make test-q1`, `make install`.

Kinds: implement (code and tests, each new test shown failing without the change), test-draft (tests
only, bug-exposing tests stay failing), experiment (any code, never merged, numbers with commands),
research (pinned-SHA permalinks, report only).

### Waves

| wave | # | slug | items (kind) | files |
|---|---|---|---|---|
| 1 | 1 | render-diff | 459 (implement): flush only changed cells; bytes and ns per frame vs full repaint on 3 examples | canvas.go, pane_screen.go |
| 1 | 2 | unicode-width | 431, 433, 460 (implement): lone U+FE0F width 0, `a`+U+FE0F width 1; 👍🏽 one cluster width 2; क्ष one cluster (GB9c, own table) width 2 (owner: VTE/Tilix shows 👍 and क्ष as 2 columns; width 1 only under mode 2027, 461) | measure/ |
| 1 | 3 | fuzz | new (test-draft): fuzz ScanKey/DecodeMouse, vt.Write, ParseAnsiBuffer, measure (only `FuzzMarkupParse` exists, `markup_test.go:203`), 15 s each; crashers as failing seeds | new *_fuzz_test.go |
| 1 | 4 | ansibuffer | 413 (implement; lead in the ticket) | ansibuffer.go |
| 2 | 5 | vt-diff | new (test-draft): loom vt vs charmbracelet/x vt, split-every-byte streams, separate module | experiments/vtdiff |
| 2 | 6 | test-hygiene | 480 (implement): timing sleeps in pane_kitty/suspend/pty/post tests to condition waits; `-race -count=3` | 4 pane test files |
| 2 | 7 | routing-tests-a | 462, 468 (test-draft + trace) | new test files |
| 2 | 8 | routing-tests-b | 469, 475, 467 (test-draft + trace; port the tickets' canaries) | new test files |
| 3 | 9 | ssh | new (experiment, local): throttled PTY (ptytest): bytes per frame, key-to-paint latency, SIGWINCH, TERM/COLORTERM loss, mode 2026; zero-dep proposals | experiments/ssh |
| 3 | 10 | tml-declarative | new (research + experiment): TUIOS `tml`/`loom` vs loom spec, `.tui` blocks, layout; zero-dep declarative layout prototype for 2 examples | experiments/declarative |
| 3 | 11 | perf | Table filter incremental index (implement; 6.7 ms per key at 100k rows) + `Benchmark*` for Draw, Flush, measure in new files | table.go, new bench files |
| 3 | 12 | uv-trial | new (experiment): Canvas flush on ultraviolet; bytes, ns, allocs vs loom and vs session 1 | experiments/uv |
| 4 | 13 | session-backpressure | 474 (test-draft; implement a client write deadline if reproduced) | internal/session writer path |
| 4 | 14 | vt-kitty | 472 (implement) | vt.go kitty functions, session.go applyTabModes |
| 4 | 15 | width-2027 | 461 (implement): vt mode 2027, one width policy | vt/, measure policy switch |
| 4 | 16 | ptytest-view | 437 (implement M1.1-M1.3, already canaried) | internal/ptytest/view |
| 5 | 17 | paste | 464 (implement: one paste sanitiser, RichTextEdit accepts paste); 471 (test-draft) | text widget paste code |
| 5 | 18 | clipboard | 465 M1 + 466 non-key parts (implement): injectable clipboard reader/writer, OSC 52 write, `[Image #n]` token model; key bindings as test-draft | clipboard code, text_entry token code |
| 5 | 19 | coverage | new (test-draft): 3 lowest-covered public packages, tests in new `*_cover_test.go` | new test files |
| 5 | 20 | examples | new (implement): `Example*` tests for the 10 most-used public types | new example_test.go files |
| 6 | 21 | sel-core | 470 option A (experiment; kept as last wave, decided 2026-11-10): shared selection/undo core under TextInput, TextArea, TextEntry | text widgets |
| 6 | 22 | sel-wrap | 470 option B (experiment; evidence only, does not decide 470): wrappers reusing TextEntry; same tests as 21 | text widgets |

Dropped after review: three audit sweeps of old tickets (most already audited in `reports/jules/`,
others are epics or manual checklists), widget-gap research (done in 458; MenuBar, Tree, Tabs, Dialog,
Split exist), repo-wide godoc pass (churn). 479 is done by the host.
Integration order for known overlaps: 2 before 15 (measure), 13 before 14 (session.go, different
functions), 17 before 18 (text_entry.go); 21/22 are evidence only.

### Mechanics
- Before each start, `git ls-remote` origin and github `main` are equal (harnez syncs at start).
- `HTO=0 harnez agent start --model jules --role developer --name jules-<slug> -- "<title>\n\n<prompt>"`
  as a background job; exit 3 -> `wait`; a question -> second reply "Submit what you have as a draft PR
  now."; stall -> failed.

### Checking and adoption (later sessions)
1. Gate script in the scratchpad (not the repo): fetch every `pull/<n>/head`, check changed paths
   against the session's files, verify each quoted `path:line` within 10 lines at the base commit, run
   `go test -race` on touched packages, and run each new test on the base (it must fail there).
2. agy:flash38 review only for PRs that pass the gate.
3. Reports in one commit on `main`; experiment, research and test-draft PRs closed after capture.
4. Implement code adopted in three themed AGit PRs: A measure + vt + ansibuffer (2, 4, 14, 15, plus
   tests from 3 and 5), B render + session + perf (1, 11, 13, 16), C widgets + tests (6, 17, 18, 19,
   20). `make test-q1` and `make install` once per stack; fixes are the host's, never Jules's.
5. test-draft output for routing tickets goes into those tickets for the terra:med developer.
6. Each checked finding becomes a ticket via /issue or a quick fix with a test.

## 3. Implementation & Verification Plan
/goal After owner sign-off, run the waves in order and record each session below; then check every PR
with the gate and capture its findings, adopting code through the three themed AGit PRs; finish with
`docs/studies/jules-sweep-2026-11.md` (what each session gave, adopted vs dropped, lessons). Stop and
report when blocked on an owner decision, a denied permission or the Jules cutoff; on a loop or no
progress, clear the goal and write a failure report here.

## Findings (checked by the host)

- F1 (#118, real bug): `ParseAnsiBuffer("\x1b[7m0")` keeps `Invert`, but `Serialize` writes
  `\x1b[0m0`: `AnsiCell.Style()` swaps FG and BG instead of setting `ansi.Style.Invert`
  (`ansibuffer.go:70`), so reverse video on default colours is lost. Fix after the ansibuffer PR
  (same file): Serialize emits SGR 7 for `Invert`; round-trip seed `"\x1b[007m0"`.
- F2 (#118, real bug, goes to 477): `vt.Write("0\xc2")` leaves the cursor at 0,0, while
  `Write("0")` then `Write("\xc2")` leaves it at 1,0: a trailing incomplete UTF-8 byte holds back
  the cluster before it. The report quotes the wrong input (`"\x1b[31ma"`); the seed is `"0\xc2"`.
- F3 (#118, not a bug): `measure.Clusters("\b")` is empty because Clusters drops control
  characters by design; the fuzz invariant must skip them.
- F4 (#120, deviation from the owner's decision): the PR makes `क्ष` one cluster but width 1;
  the target is width 2 (sum of the consonant widths; main gives 2 today, as VTE/Tilix and foot
  draw it). Fix at adoption in the conjunct branch of the width code (PR head `measure.go:152-191`).
  Host probe on the PR head: lone U+FE0F 0, `a`+U+FE0F 1, `#`/`❤`+U+FE0F 2, `👍🏽` 1 cluster width 2,
  `a👍🏽b` 3 clusters; `🤷🏽‍♂️` 1 cluster width 4 and the family emoji width 6 follow the default
  `LOOM_ZWJ` split mode (main gives 6 for both), so not a regression.
- Adoption PR A (Codeberg PR 92, merged; host fixes 3af53556): #120, #118 and #121 are on main;
  F1, F2 and F4 are fixed; F3 is handled in the fuzz invariants. New at adoption: a ZWJ that
  joins nothing added 2 columns in join mode (`"0‍"`, also on main before): fixed. Fuzz
  crashers `"👩‍\x7f0"`, a lone combining mark and the ansibuffer trailing blank were
  invariant problems, not bugs. `TestWidthTargets` pins the widths. GH #118, #120, #121 closed.
  Review follow-ups (minor) went to ticket 483.
- Experiment PRs #124, #126 and #127 are closed; their reports are on main under `reports/jules/`.

## Jules runs

| wave | kind | session name | Jules ID | PR | state |
|---|---|---|---|---|---|
| 1 | implement | jules-render-diff | 13272737396249956342 | none | failed: asked back after ~75 min, no reaction to the first reply, COMPLETED after "submit a draft" without a PR (replies: 2). Lead: it added `lastFrame *Canvas` to `Pane`, a `FlushDirty` method in `canvas.go`, and `lastFrame = nil` in `Invalidate()` and `Resize()`; suspend/resume, alt-screen and the pane loop were not done |
| 1 | implement | jules-unicode-width | 14705779270409174332 | GH #120 | checked: scope clean (measure/, tests, report); `go test -short ./...` and `-race ./measure/ .` pass on the PR head; targets hold except F4; report shows no before-run, but every target fails on main (host probe); adopt in PR A with F4 fixed |
| 1 | implement | jules-ansibuffer | 13101811809704567128 | GH #121 | checked: delivered after the second reply (replies: 2); files ansibuffer.go, new ansibuffer_413_test.go, report, plus a needed cmd/loom/crop.go change (wide continuation cells now hold rune 0); `go test -short ./...` and `-race .` pass on the head, the old erase test unchanged and passing; `Test413_ParseAnsiBuffer` fails on the base; adopt in PR A with F1
| 1 | test-draft | jules-fuzz | 1079445786546642272 | GH #118 | checked: scope clean (4 fuzz tests, 3 seeds, report); 3 failing seeds reproduced; findings F1-F3 below; tests to adopt in PR A with the measure invariant fixed |
| 2 | test-draft | jules-routing-tests-a | 7091831739430289177 | GH #122 | checked: scope clean (2 test files, report); both bugs reproduced on the head (462: replies reach widgets as keys; 468: horizontal wheel decodes as MousePress); trace 8 of 12 references hold; test notes written into 462 and 468; PR to close after the terra:med developer picks the tests |
| 2 | test-draft | jules-routing-tests-b | 10199914946274802065 | GH #123 | checked: scope clean (3 test files, report); all 3 bugs reproduced on the head; 15 of 15 trace references hold; 467 auto-scroll skipped; notes written into 469, 475, 467; PR to close after the terra:med developer picks the tests |
| 2 | implement | jules-test-hygiene | 14168310180638339112 | GH #125 | checked: scope clean (3 pane test files, report); `go test -race -count=3 -run TestPane .` passes on the head (host: 75 s), `go test -short ./...` passes; 20 sleeps replaced by condition waits, 13 kept with a comment; to fix at adoption: gofmt (`pane_kitty_test.go`, `pane_pty_test.go`), poll loops end silently at the deadline, `pane_suspend_test.go` "let the capture see the last writes" kept although it is a timing wait; adopted in PR C (Codeberg PR 93, merged): `waitDims` fails on timeout and exposed two waits for a width the pane never had (2 s lost each); GH #125 closed |
| 2 | test-draft | jules-vt-diff | 5267456316068289451 | GH #124 | checked: scope clean (experiments/vtdiff module, report); 8 of 25 streams differ, host rerun agrees; the report calls loom's missing ICH/DCH/IL/DL and DECSTBM upstream bugs, wrongly: filed as 482 (with BS from pending wrap); VS16/skin tone are measure policy (#120), ZWJ is LOOM_ZWJ policy, incomplete UTF-8 is F2; experiment PR, close after report capture |
| 3 | experiment | jules-ssh | 1348299604964268629 | none yet | running (started 22:50 after ansibuffer and routing-tests-a finished) |
| 3 | research + experiment | jules-tml-declarative | 17660407404275388988 | GH #126 | checked: scope clean (experiments/declarative, report); prototype tests pass; TUIOS permalinks (pinned 6fe2c0d: `tml/xml.go`, `loom/surface.go`, layout.tml) and loom refs (`split.go:43`, `tabs.go:109`) hold; `TestPublicAPI` fails only because the prototype sits in the root module without an API snapshot (prompt allowed it); finding: TUIOS `tml` is an XML dialect (~5k lines) feeding its own `loom` package; recommends a Go builder API, not a text format; lead for the wrap-up study; experiment PR, close after capture |
| 4 | implement | jules-vt-editing | 8746886200815324221 | none | failed: Jules reported "unable to complete the task" (FAILED state), no PR, no message; retried smaller as jules-vt-rowedit |
| 3 | implement | jules-perf | 6799813322793390744 | GH #128 | checked: delivered after the auto-reply; benchmarks rerun by host; they exposed `ParseAnsiBuffer` sizing a cursor-addressed 200x60 screen to 12000 columns (47 ms, 15.6 MB), fixed by host to 1.3 ms, 382 kB; benchmarks adopted in PR B (Codeberg PR 94, merged); incremental Table filter not adopted (its benchmark matched every row, no gain shown; misses direct `Table.Rows` writes; equality test compared only heights); stray `plan_review_request.md` dropped; GH #128 closed |
| 3 | experiment | jules-uv-trial | 13571252018964085731 | GH #127 | checked: scope clean (experiments/uv module, report); benchmarks rerun by host, numbers agree within machine speed; full loom flush 12.9 kB per frame vs ultraviolet 3 B (two cells), 403 B (selection), 244 B (scroll); ultraviolet 10x slower on first paint; finding added to 459; experiment PR, close after report capture |
| 4 | test-draft + implement | jules-session-backpressure | 16549792241857724638 | none | failed: stalled after 3 replies, no PR. Lead: reproduced the stall (`writeOutput` holds `writeMu` while writing to a client that stopped reading, `internal/session/server.go:88`); a write deadline plus `go detach` broke the deadlock but made `TestSessionCtrlRightBracketPassesThrough` and `TestSessionFailedScreenshotReportedOnStderr` fail; retried as test-draft only (jules-backpressure-test) |
| 4 | test-draft | jules-backpressure-test | 9706802655975301250 | none yet | running (474 failing test only, no fix) |
| 4 | implement | jules-ptytest-view | 749146208954348203 | GH #129 | adopted in PR D (Codeberg PR 95, merged): scope clean; host fixed 3 bugs (wide continuation cells rendered as spaces; `Relay.Render` dropped child bytes; attach test never checked the restore and put the test terminal into raw mode); flash38 found the last two; q1 pass; open points written into 437; GH #129 closed |
| 4 | implement | jules-vt-kitty | 9888375810308224771 | none | failed: "unable to complete the task", no PR, no message; retried smaller as jules-vt-kittystack (vt only) |
| 4 | implement | jules-vt-kittystack | 6727396211403640995 | none yet | running (472 part 1: vt flag stack; session restore left for the host) |
| 4 | implement | jules-vt-rowedit | 13475704617068467011 | none yet | running (482 part 1: ICH, DCH, ECH, BS after pending wrap; IL/DL/DECSTBM left for later); auto-reply sent |
| 5 | implement + test-draft | jules-paste | 5850673660620747468 | none | failed: "unable to complete the task", no PR; not retried, 464 and 471 go to a host developer |
| 5 | implement | jules-examples | 7667649542158170067 | none yet | running (`Example*` tests for the 10 most-used types) |
