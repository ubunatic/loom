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

## Jules runs

| wave | kind | session name | Jules ID | PR | state |
|---|---|---|---|---|---|
| 1 | implement | jules-render-diff | pending | none yet | running (started 2026-11-10 20:03, only session started; owner asked for one at a time) |
