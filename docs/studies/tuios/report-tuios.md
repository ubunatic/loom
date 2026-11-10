# TUIOS study (ticket 457, topic 1)

Scope: what TUIOS is, which Charm libraries it uses and how, what it does that loom does not,
and what loom could learn. Feature rows with sources are in `features-tuios.yaml` (rows marked
`verified: true` were read from the source at one commit).

Sources and method:
- Web pages, fetched on 2026-11-10: https://tuios.dev/ (index), https://tuios.dev/loom (the
  loom package page). Both are summaries; quotes from them are paraphrased in this report.
- Source: https://github.com/Gaurav-Gosain/tuios at commit
  `6fe2c0df9922b2eec28f392fee4e196efd86be36` (main, pushed 2026-11-10). Shallow clone, read with
  grep and sed. Line numbers in this report refer to that commit.
- Latest release: v0.12.1 (published 2026-10-31, via `gh api`). License: MIT (repo metadata).
- Not done: running TUIOS, measuring its performance, reading its tests in detail, the
  e2e/clienttests suites, the web and SSH servers in depth.

## 1. What TUIOS is

TUIOS (Terminal UI Operating System) is a terminal window manager written in Go. It runs
several programs in panes, tiles or floats them, and keeps the sessions alive when the user
detaches. The root of the repo is an application (`cmd/tuios`), plus a web server
(`cmd/tuios-web`) and a set of Go modules that are also published as libraries:
`tuios.dev/vt` (terminal emulator), `tuios.dev/loom` (TML document renderer), `tuios.dev/tml`,
`tuios.dev/plot`, `tuios.dev/sketch`, `tuios.dev/shot`, `tuios.dev/swatch`,
`tuios.dev/progstatus`. Size: about 60 `internal/` packages (for example `internal/app` has
about 1000 files including tests, `internal/session` about 720).

The root `AGENTS.md` describes it as "built in Go using the Charm stack (Bubble Tea v2, Lipgloss
v2)". The site says it also has a daemon, an inbox for coding agents, remote hosts, a plugin
system, a control protocol over a Unix socket and a browser client. Those site claims are
not checked against code in this report.

## 2. Charm stack: what TUIOS uses

Versions are from the root `go.mod` (lines 8-18 and 52).

| Library | Version in go.mod | Use in TUIOS | Notes |
|---|---|---|---|
| `charm.land/bubbletea/v2` | v2.0.10 | Event loop, every client (`tea.NewProgram`) | v2 only. No v1 import. About 430 import sites in the clone. |
| `charm.land/lipgloss/v2` | v2.0.6 | Styling; `lipgloss/v2/table` for tables | About 193 import sites (tests included). |
| `github.com/charmbracelet/ultraviolet` | pseudo-version 2026-09-06 | Cell buffer and style diff in the renderer | About 169 import sites. Own renderer over it. |
| `github.com/charmbracelet/x/ansi` | v0.11.8 | ANSI parsing and writing | About 242 import sites. |
| `github.com/charmbracelet/colorprofile` | v0.4.3 | Colour depth detection | 28 sites. |
| `github.com/charmbracelet/x/xpty`, `x/term`, `x/windows` | various | PTY and terminal plumbing | |
| `charm.land/ssh`, `charm.land/wish/v2` | v0.4.3, v2.0.5 | SSH server mode | Also `github.com/Gaurav-Gosain/sip` for serving Bubble Tea over SSH. |
| `charm.land/tone` | v0.1.0 | Not checked | 8 import sites. |
| Bubbles, Huh | not in go.mod | Not used | No import found. |
| Glamour | not in go.mod | Not used | Only mentioned in a comment; see 2.3. |

### 2.1 Bubble Tea v2 (what they rely on and where they work around it)

- **Options.** Every `tea.NewProgram` gets one shared list, `app.ProgramOptions()`, in
  `internal/app/program_options.go#L44-L61`. It sets the frame rate, turns off Bubble Tea's
  signal handler (`tea.WithoutSignalHandler`) and sets one filter (`tea.WithFilter`). A test
  (`internal/app/program_options_test.go#L30`) reads the source and fails if any `tea.NewProgram`
  does not take the list. Worth copying if loom has more than one place that builds a Bubble Tea
  program (host to check).
- **Filter is one slot.** `tea.WithFilter` keeps one function, so the last one set wins
  (`program_options.go#L36-L38`). TUIOS puts its own filter last so that the filters from `wish`
  and `sip` (the SSH libraries) do not replace it. That filter is also a whitelist for mouse
  motion (`FilterMouseMotion`, `#L164-L166`): a pointer move reaches `Update` only when a
  clause asks for it. This keeps an idle pointer from costing frames.
- **Frame ticker.** Bubble Tea fixes the frame ticker when the program is created and clamps it
  to 120 frames per second. TUIOS calls `BindProgram` (`internal/app/frame_rate.go#L99-L104`)
  to set the ticker past that clamp and to change it when the user changes `max_fps`. The
  comment at `program_options.go#L50-L57` says a 120 fps ticker woke the idle process 590 times
  per second (1.0% of a core) and 60 fps woke it 370 times. These are their numbers, not
  measured here.
- **A gap in Bubble Tea.** `cmd/tuios-web/main.go#L517-L520` says: if `Run` fails before
  `Init` runs, `Cleanup` never runs, and closing that "needs Run's channel to exist before Run,
  which only bubbletea can do." They point to "the note on Wait in the upstream report"; that
  report is not fetched here. This matters for loom's crash-restore item.
- **No forks.** No `replace` directive in the root, `vt`, `loom` or `tml` `go.mod` files (grep).

### 2.2 Lip Gloss v2 and ultraviolet (rendering)

- The renderer diffs cells by style. `internal/app/frame_render.go#L52` calls `uv.StyleDiff`
  and caches the diff for each pair of styles (`diffs map[[2]uv.Style]string`, `#L34`). The
  result is meant to be byte-for-byte the same as `uv.TrimSpace(uv.Lines(lines).Render())`
  (`#L57`). So ultraviolet is used for the data model and the diff, while TUIOS writes its own
  frame loop.
- Lip Gloss is used for styles and tables. No Bubbles or Huh components are used, so TUIOS's
  widgets (pane borders, dock, menus, palette) are hand-written on top of Lip Gloss.

### 2.3 Glamour (Markdown): avoided on purpose

`internal/markdown/markdown.go#L12-L15` says goldmark with its GFM extensions adds about 480 KB
to the binary and Glamour (which links every chroma lexer) adds about 8.8 MB, so the package
parses the Markdown subset that agent replies use, in about 94 KB. Code blocks are highlighted by
`internal/diffview`. Loom already has Markdown in TextEntry, so the comparison is about size
and subset, not features.

### 2.4 Charm x/* packages

- `x/ansi` is used in about 240 files. Their own VT emulator (`tuios.dev/vt`) sits on it.
- Their emulator package is `tuios.dev/vt` (`vt/go.mod` requires `x/ansi`, ultraviolet and
  `go.mitchellh.com/libghostty`). Its pure Go backend is the default. `-tags ghostty` swaps in
  libghostty-vt; `docs/ARCHITECTURE.md#L25` says exactly one backs a binary. This is a second
  implementation loom can test against. Which backend the default build uses for width is not
  checked here.

## 3. The name clash: `tuios.dev/loom` is not loom

`tuios.dev/loom` is a Go module inside the TUIOS repo (`loom/go.mod`, module `tuios.dev/loom`).
Its README (`loom/README.md#L5-L8`) says it "draws TML documents into terminal cells", and it
gives a cell buffer that renders to ANSI or plain text. Its selling points on the site are flex
layout, theme roles at each colour depth, and redraws of only what changed. It imports `tml`,
`plot`, ultraviolet and `x/ansi`, and no TUIOS package. It is a document renderer, not a widget
SDK.

Our loom is `ubunatic.com/loom`. The two share a name only. A search for "loom" in Go
docs will land on theirs. The host should decide whether the study refers to them as
"TUIOS loom" and never as "loom" alone.

Two points are relevant to our design:
- Theme roles: "Every colour is a role, such as `success`, `mute` or `series2`" (site wording).
  Loom has a palette; check whether it has roles.
- Patch-only redraw and flex layout are the same ideas as loom's cell canvas and layout work.

## 4. Architecture: where TUIOS differs from loom

1. **Daemon sessions.** A daemon (`internal/session`) owns the sessions, the PTYs and one VT
   emulator per pane (`docs/ARCHITECTURE.md#L23`). A pane keeps a 64 KB ring with stream
   positions (`#L75`). A client attaches with a snapshot plus the stream from the snapshot's
   position (`docs/REHYDRATION.md`, not read). The daemon can also serve panes on other
   machines through SSH ("federation"). This is the core of "persistent sessions". Loom has
   `loom session` with an internal `vt.VT` emulator; the comparison is about attach and
   rehydration, which is not covered by the rows here.
2. **Modal input.** The leader key is `ctrl+b` by default (`internal/config/userconfig.go#L1333`).
   Keys go to one of three modes: Terminal (to the pane), Window Management (commands), or Copy
   Mode (`docs/ARCHITECTURE.md#L159` and `#L225-L229`). Copy Mode is vim-style with 50+ motions
   (`#L661`). This is the opposite of loom's desktop model (Ctrl+C/V/X/Z without a leader).
   The ticket says loom is not to be vim-like, so this is a contrast, not a target.
3. **Layouts.** Four modes: BSP (default), master-stack, scrolling (niri-style strip of
   columns) and tiling off (`docs/LAYOUT_MODES.md#L24-L29`). The layout code is in
   `internal/layout` (`bsp.go`, `master.go`, `scrolling.go`). Each of the nine workspaces keeps its
   own layout data (`LAYOUT_MODES.md#L34-L35`).
4. **Clipboard.** OSC 52 writes from a pane are gated by a setting (`off`, `ask`, `focused`,
   `on`; default `focused`) (`internal/app/clipboard_osc52.go#L13-L29`). A query returns empty
   content (`#L13-L14`). This is a safety policy loom does not have (host to check its OSC 52
   write path).
5. **Drags.** While a drag is held outside the content area, a 50 ms tick keeps scrolling
   (`internal/app/update.go#L715-L721`). Whether a drag is captured by the pane when the
   pointer leaves it is not shown in the code read here.
6. **Wheel.** A 250 ms gap ends a wheel gesture, and the axis with the majority wins, so
   sideways drift is dropped (`internal/app/wheel_axis.go#L5-L10`). There is no acceleration
   curve in that file; loom's curve is a different approach.
7. **Input and paste.** Kitty keyboard flags are a stack of up to 8 entries, separate for the
   main and the alternate screen (`vt/kitty_keyboard.go#L14-L24`). Paste is sanitised, then
   wrapped in bracketed-paste markers only if the application turned them on
   (`internal/terminal/paste.go#L14-L23`).
8. **Width.** The emulator forces DEC mode 2027 (grapheme clustering) on and reports it as
   permanently set (`vt/mode.go#L51`). Loom's measure package probes the terminal instead. A
   comparison needs both behaviours on the same terminals (not done).
9. **Synchronized output.** Mode 2026 wraps only text-sizing passthrough output
   (`internal/app/text_sizing_passthrough.go#L292-L295`). The main frame is not wrapped in
   2026 in the code read (grep of `internal/app`, `internal/terminal`, `internal/session`,
   `internal/served`). Loom's frame-level 2026 is a real difference, and it may be a gap for
   TUIOS. Needs a canary before it is claimed.

## 5. What TUIOS does not do (as far as read)

- No Bubbles or Huh widgets. Their UI is hand-built on Lip Gloss.
- No Glamour. Markdown is their own package.
- No replace directives or forks in `go.mod`.
- Bubble Tea v2 only. No v1 path to compare.

## 6. Quality points (for the summary)

- Tests: `internal/lint` and `internal/app` contain source-scanning tests (for example the
  shared-options test). Their `vt` package has fuzz tests and differential tests against
  libghostty (`vt/ghostty_differential_test.go`, `vt/vtgen/`). Loom has `loomtest.VT` and
  `CompareWidget`; the two approaches are different and not yet compared.
- Docs: `AGENTS.md`, `docs/ARCHITECTURE.md`, `docs/LAYOUT_MODES.md`, `docs/KEYBINDINGS.md` (the
  site holds the full keybinding list). Docs point to code files.
- Release: v0.12.1 on 2026-10-31; main was last pushed 2026-11-10. Cadence over months was not
  measured. The library modules in the root `go.mod` are pinned to pseudo-versions from
  2026-10 and 2026-11.

## 7. Website-only claims (not checked in code)

These come from tuios.dev and are not in the rows: agent inbox ("one Inbox for what coding
agents in your panes are waiting on"), remote hosts, hints ("a short label on each URL, path,
hash"), themes and glyph sets, `tuios-web` (browser client via xterm.js: `docs/ARCHITECTURE.md`
says so), plugins pinned in `plugins.lock`, control protocol (line-delimited JSON on a socket).
Each is a candidate for a row after a read of its file.

## 8. Open gaps

- Not read: `internal/input` (key routing), `internal/session` (rehydration, attach), `vt/snapshot`,
  `clipboard_paste.go` (host clipboard read), `docs/REHYDRATION.md`, `docs/KEYBINDINGS.md` on
  the site, the e2e suites (`clienttests/`, `.github/workflows/e2e-*`).
- Not measured: idle CPU, frame cost, width probes, paste limits.
- The upstream Bubble Tea report about `Run` before `Init` is not fetched (number unknown).
- Whether TUIOS detects terminals without mode 2027 or 2026 support is not checked.
- `loom` column in the feature rows is `unknown` for every row; the host fills it from loom's
  docs, spec and code.

## 9. Suggested next steps for the host (not filed)

1. Fill the `loom` column of the 23 rows from loom's code (`docs/`, `spec/widgets.yaml`,
   `internal/` modules). Start with the P1 rows: input model, paste, drag, OSC 52 policy,
   width 2027, frame-level 2026, session rehydration.
2. Canary: run a TUIOS build in a PTY and check whether its main frame is wrapped in 2026
   (one capture, one grep). Only then decide whether 2026 is a gap for them or a non-issue.
3. Re-read `internal/session` attach to compare with `loom session`. This is the largest
   difference in the study.
4. Decide what "TUIOS loom" means in the summary, to avoid the name clash.

Counts: 23 feature rows written; 23 verified (all rows read from the commit above).
Open gaps: section 8.
