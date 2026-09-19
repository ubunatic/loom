---
title: Loom Roadmap
weight: 20
---

# Loom Roadmap

Loom is a **library**, not an application. Its value is measured by what a
*host program* can do with it: how little ceremony it takes to construct,
theme, compose and run a widget; how correct the compositor is under nesting;
and how honest the public API surface is. The example apps in `examples/` are
not the product — they are the proof that the library contract holds.

Loom's value is reusable, dependency-light inline terminal UI: applications
such as the Harnez/Voxi reference dashboards should share widgets without
taking over the full terminal. An application should declare its UI structure
and presentation, while Go supplies behavior and changing data. The UI should
be easy to maintain, responsive to terminal size, testable without a live
terminal, and free from application-specific coordinate/layout code.

The primary reference targets are documented in
[`HarnezUsageTarget.md`](HarnezUsageTarget.md) (usage dashboard and Voxi monitor)
and [`HarnezSplashTarget.md`](HarnezSplashTarget.md) (startup loading and splash
screen).

Stage numbers below preserve the original plan and remain as historical
acceptance context. The Now/Next/Later sequence is the live priority signal and
supersedes stage order wherever the two disagree.

## Shipped since the last roadmap pass (2026-09-19)

The compositor and theming track closed out completely this session. This
retires the whole "Later/Park — styling and theming" bucket the previous
roadmap carried, and it changes the *shape* of the remaining hosted-widget
work: overlays and background inheritance are no longer open design questions.

| Issue | What landed | Why it matters downstream |
|---|---|---|
| [067](../issues/067-animated-loom-background-for-filebrowser-with-proper-compositing.md) | Animated background compositing contract, demonstrated by `examples/background` (see [AnimatedBackgrounds.md](AnimatedBackgrounds.md)) | Proved a widget can paint *behind* the tree without owning it |
| [068](../issues/068-formalize-layered-compositor-semantics-and-background-inheritance.md) | Explicit layered compositor API (`PaintSurface` / `PaintForeground` / `PaintDecoration`); nested-canvas surface propagation fixed; **all** widgets migrated off background-inference heuristics; PTY smoke validated | The layering contract 061 (theme propagation) and 065 (ANSI rows) both build on. No widget now guesses its background |
| [066](../issues/066-nested-help-pane-opens-a-second-pane-racing-the-outer-pane-tty-reader.md) | Single-pane ownership guard; default help path no longer opens a second `Pane`/`/dev/tty` | Removes the tty-reader race that would have made *any* hosted-widget work unsafe; the same guard is what 065's treemap conversion must satisfy |
| [069](../issues/069-render-help-as-a-root-level-modal-overlay.md) | Root-level `:help` modal via the `paneHelpRequest` hook pattern — documented as a reusable mechanism in [RootOverlays.md](RootOverlays.md) | Gives Loom a *general* answer for anything that must draw above a clipped widget tree (confirm dialogs, global notices). Future modal needs reuse this hook rather than inventing plumbing |
| [070](../issues/070-prevent-text-overflow-in-framework-help-modals.md) | Framework help modal text display-width truncated to content bounds; Unicode/narrow-width regression coverage | Filed and closed inside the same session as a direct follow-up to 069. Its Unicode-truncation test pattern is the template 038 still needs |
| [050](../issues/050-support-truecolor-rgb-values-in-theme-specs.md) | Truecolor RGB values in theme specs | Unblocks 014's palette work whenever a real consumer asks for it; removes the "themes are palette-limited" objection from 061 |
| [053](../issues/053-pane-run-key-decoding-drops-multi-key-reads-and-splits-escape-sequences-across-tty-reads.md) | Coalesced multi-key reads and split escape sequences handled in `Pane.run` | Was the previous roadmap's top "Now" item. The core input path is now sound enough to build 057's routing contract on top of |
| [047](../issues/047-remove-deprecated-uzu-brand-and-default-global-commands-from-cmdbar.md) | uzu branding and `:home` removed from default globals | Surface hygiene; `cmdBar` defaults are now loom's own |

Everything the previous roadmap listed under "Now" is closed. The sequence
below is therefore a fresh reading of the remaining 25 open tickets, not a
shuffle of the old one.

## Themes in the open backlog

The 25 open tickets fall into six groups that emerged from their content, not
from a fixed taxonomy:

1. **Hosted-widget library contract** (057, 058, 059, 060, 061) — the missing
   API that lets a widget be run by a host it did not construct.
2. **Hosted-widget example conversion** (062, 063, 064, 065) — proving the
   contract by making every `examples/*` app "just a widget".
3. **Rendering and measurement correctness** (038, 048, 039) — geometry bugs
   reachable by real apps, plus two that root-cause *outside* Loom.
4. **Layout capability gaps** (051) — an expressiveness hole a concrete
   example already hits.
5. **API ergonomics and surface honesty** (034, 036, 037, 042, 052) — small,
   independently shippable work that reduces caller boilerplate or removes
   misleading public API.
6. **Declarative data sources and simulated targets** (014–019) — the original
   dashboard track, still without a concrete blocking consumer.

Plus one explicitly aspirational ticket (056) that is not scheduled.

## Now — unblock hosting, and close the cheap correctness debt

**Rationale.** Hosting is Loom's live value axis: a widget library whose
widgets can only run in the app that built them is a collection of examples,
not a library. Stage 13's contract tickets gate everything in theme 2, so they
are the critical path. 059 earns its place independently — it is a *live bug*
any app hitting a clickable widget inside a non-origin `Frame` can reproduce
today, with no hosting involved. 036 and 038 are in Now because they are hours
of work each, not days, and both directly de-risk the bigger items beside them.

- [059](../issues/059-fix-mouse-coordinate-convention-mismatch-between-frame-and-tabs-stack-grid.md)
  (P2, Bug): two incompatible mouse-coordinate conventions coexist —
  `Frame.HandleMouse` re-bases into child-local coordinates while
  `Tabs`/`Stack`/`Grid` forward unchanged, so a `Choice` inside a `Frame`
  subtracts `lastRect` twice and mis-hits rows. `Stack`/`Grid` also route to
  the *focused* child rather than the clicked one. Independent of all other
  work; pick one convention and pin it before more composites are written
  against the ambiguity.
- [057](../issues/057-hosted-widget-key-contract-child-first-routing-reserved-host-keybinds-and-quit-containment.md)
  (P2): child-first key routing, a `KeyConsumer` notion of *consumed* (today
  `HandleKey` reports only *quit*), and quit containment so a hosted app's `q`
  closes its tab instead of the host. Gates 062, 063 and the per-app
  `DisableDefaultQuit` knob.
- [058](../issues/058-widget-declared-pane-requirements-panerequest.md)
  (P2): `PaneRequest` so a widget declares mouse mode, `Resizeable`, `MaxCols`
  and quit policy instead of the example setting them on a pane it owns.
  Currently these degrade *silently* when hosted — a hosted widget's
  `HandleMouse` simply never fires. Co-gates 062 with 057.
- [036](../issues/036-keyevent-ergonomic-helpers-e-name-and-e-is-for-unified-key-text-matching.md)
  (P2, small): `KeyEvent.Name()` / `Is()` / `Rune()`. Today every widget
  hand-writes `key := e.Key; if key == "" { key = e.Text }` or silently fails
  to match printable keys. Pull this in *before* 057 rewrites routing, so the
  new contract is written against the ergonomic form once.
- [038](../issues/038-fix-multi-byte-utf-8-string-truncation-in-popup-title-and-borders.md)
  (P1, Bug — **likely a close**): the code fix already landed
  (`popup.go` now uses `textClusters`/`StringWidth`, commit `6c5dc05`), but the
  ticket's acceptance tests — multi-byte titles at truncating widths, a
  multi-byte title that fits, an ASCII regression — do not exist; `loom_test.go`
  covers only `TestPopupDraw`/`TestPopupEscCloses`. Backfill using 070's
  Unicode/narrow-width test pattern and close. Cheapest P1 on the board.

## Next — finish the contract, then convert the easy examples

**Rationale.** These are gated by Now, or are small capability gaps whose
concrete caller is one of the conversions. 062 is deliberately the *first*
conversion because split/tabs have zero flags, zero theming and no live data,
and because both currently have **no smoke coverage at all** — converting them
closes a real test hole as a side effect.

- [060](../issues/060-periodic-redraw-without-pane-ownership-ticker-interface-and-pane-invalidate.md)
  (P2): a `Ticker` interface plus `Pane.Invalidate`. Today `RunWatch` supports
  exactly **one** cadence and one collect callback per pane, so `monitor` and
  `splash` structurally cannot be two live tabs. Gates 064.
- [061](../issues/061-themeable-host-provided-theme-propagation-through-composite-widgets.md)
  (P3 but promoted): `Themeable` propagation. Nothing auto-themes today — every
  widget's style is baked in at construction, so a hosted filebrowser's F9
  theme cycle would retheme only its own subtree. Gates 063. Now cheaper than
  when filed: 068 removed background inference and 050 widened the value space.
- [062](../issues/062-example-widget-factories-newwidget-for-split-and-tabs-hosted-loom-demo-mode-headless-bench-smoke.md)
  (P2): `NewWidget` factories for `split` and `tabs`, hosted `loom-demo` mode,
  headless bench smoke. The end-to-end proof of 057+058; `tabs`-inside-`Tabs`
  is the meta-test for child-first routing.
- [051](../issues/051-allow-frame-boxes-to-fill-available-content-height.md)
  (P2): let a box or row of boxes consume the frame's content height instead of
  each app re-deriving "rows minus title minus status". filebrowser hard-codes
  `Height: 18` today. Sequence next to 063 so the conversion validates the new
  fill semantics rather than preserving the magic number.
- [042](../issues/042-docs-tuiinput-md-referenced-by-5-code-comments-but-does-not-exist.md)
  (P3, Documentation): five code comments in `pane.go`/`event.go`/`loom_test.go`
  cite `docs/TuiInput.md` §1–§3; the file has never existed. It should document
  `/dev/tty` constraints, the poll-then-read/EINTR pattern, and the two-prefix
  key decode. Write it *while* 057/036 are fresh — that is the same subsystem,
  and 066's single-pane ownership guard now belongs in it too.
- [034](../issues/034-ansi-sgr-escape-sequence-parsing-and-writeansi-canvas-helper.md)
  (P2): `ParseANSI`/`Canvas.WriteANSI` for 3/4-bit, 256-color, 24-bit and
  attribute SGR. Currently `Canvas.Write` and `View.Draw` strip ANSI and apply
  one uniform `Style`. This is a **hard prerequisite for 065** (treemap keeps
  `RawScreen` solely because `View` drops its inline ANSI) and the thing any
  external syntax highlighter or diagram engine needs. Pull it forward out of
  "ergonomics" on that dependency alone.

## Later — the remaining conversions and the surface cleanup

**Rationale.** Each of these is either gated by two or more Next items, or is
genuine but non-urgent surface work with no caller waiting.

- [063](../issues/063-convert-filebrowser-example-to-a-hostable-widget.md)
  (P2): the conversion that exercises 057, 058, 059 and 061 simultaneously —
  `q`-as-filter-key vs default quit, `--theme` and F9 cycling, mouse clicks on a
  `Frame`-based layout, `MaxCols: 0`. Deliberately after 062 so it validates the
  API rather than designing it.
- [064](../issues/064-convert-splash-and-monitor-examples-to-hostable-widgets.md)
  (P2): the live-data pair. Blocked on 060, and additionally needs options
  structs — both apps' cobra `RunE` closure currently *is* the app, and neither
  produces a `Widget` at all in show-once mode.
- [065](../issues/065-ansi-styled-rows-widget-and-treemap-conversion.md)
  (P3): last and hardest. Needs 034's ANSI cell parsing, plus unwinding
  treemap's own `/dev/tty` raw-mode reader, its own signal handler, its own
  ticker and its stderr writes — all of which would fight a host pane. 066's
  ownership guard now makes the current design fail loudly, which is the right
  pressure.
- [037](../issues/037-canvas-drawborder-and-drawbox-primitives-with-configurable-boxstyles.md)
  (P2, Refactor): `Canvas.DrawBorder`/`DrawBox` with configurable `BoxStyle`,
  consolidating the three independent border implementations (`Box.Draw`,
  `Popup.Draw`, and the termaid viewer's hand-rolled rounded modal). Real
  duplication, but the sharpest symptom (038's byte truncation) is already
  fixed, so the urgency dropped. Do it when 052's decision needs it.
- [052](../issues/052-resolve-unused-choicestyle-border-contract.md)
  (P2, Refactor): `ChoiceStyle.Border` is public, theme-mapped and **never
  read** by `Choice.Draw`; the `focused` field's documented border dimming does
  not exist either. Honest-surface work: either delete the field or implement
  the border. Coordinate with 037 — if borders become a container primitive,
  deletion is the answer.

## Close / deprioritize candidates

Flagged for an explicit decision rather than indefinite carry.

- [048](../issues/048-emoji-rune-width-discrepancy-causes-horizontal-border-drift.md)
  (P2, Bug) — **likely reclassify to Documentation.** The ticket's own root
  cause is that `🖼` (U+1F5BC) has Neutral East Asian Width and `wcwidth` 1,
  while emoji-presentation tables report 2, and terminals disagree with each
  other. Its own proposed resolution is "use guaranteed 1-column glyphs
  (`▣ ▧ ▦ ◈ • *`) in borders and titles" — that is guidance, not a fix. Worth
  doing: audit `measure.RuneWidth` against East Asian Width + Emoji
  Presentation (including VS-16) and *document* the guarantee Loom can actually
  make. Not worth doing: chasing per-terminal parity.
- [039](../issues/039-graph-renderbar-subchar-boundary-glyph-shows-a-visible-seam-without-ansi-background-styling.md)
  (P3, Bug/Documentation) — **close as won't-fix-in-code.** Already
  root-caused to terminal/font behavior: VTE-class terminals procedurally
  fill `█ ▓ ▒ ░` but render the eighth-block glyphs `▏▎▍▌▋▊▉` from font
  outlines with their own bearings, producing the seam. Glyph selection math
  was verified correct. The actionable residue is a documented recommendation
  (use `ANSI`/`BackgroundANSI`, or drop `SubChar`, when a seamless bar matters)
  — fold it into the graph docs and close.
- [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md)
  (P2) — **milestone ticket with no consumer.** It gates on 014 and 015, which
  are themselves demand-less. Consider converting it from a blocking milestone
  into a checklist that a future real consumer can pick up, so it stops
  appearing as scheduled work.
- [056](../issues/056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md)
  (P3, self-labelled "aspirational, not scheduled") — direction 1 (in-process
  widget factories) has been fully absorbed by 057–065 and is being delivered.
  Direction 2 (PTY-hosted arbitrary programs: `creack/pty`, a VT100 parser, an
  in-memory screen grid, `TIOCSWINSZ` resize propagation) is a separate product
  in its own right. Either split direction 2 into its own ticket and close 056,
  or keep 056 explicitly parked — but it should not sit in the hosted-widget
  stage table as if it were sequenced.
- [019](../issues/019-evaluate-declarative-source-and-action-wiring.md)
  (P3) — remains parked behind 017/018, unchanged from the prior roadmap. Its
  narrow file/cadence declaration is already proven by 027; keep it open only
  for source mappings, named handlers and the adopt/narrow/reject decision, and
  do not let a second competing declaration syntax grow beside 027's.

## Parked — declarative data sources and simulated targets (014–019)

Unchanged in substance from the prior pass, restated because the tickets are
still open and the reasoning still holds: **no concrete downstream app blocks
on any of them.**

- **Expressive target UIs.**
  [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md) →
  [015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md) →
  [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md):
  declared colors/glyphs, bounded transcript and daemon panels, then the
  two-target matrix. 050's truecolor support removed one technical objection to
  014, but not the demand objection. Revisit when a real consumer asks for
  configurable palettes or Voxi panels.
- **External boundaries.**
  [017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md)
  → [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md):
  separate-process file/socket fixtures and lifecycle canaries, then bounded
  Linux/daemon probes. Reuse 027's file collector and add
  replacement/partial-write, disconnect/reconnect and stale-data evidence
  rather than rebuilding it. Real CPU-percentage parsing, GPU paths/units and
  daemon/transcript formats still need source-specific evidence under 018.
- **Wider declarative wiring.** 019 — see close/deprioritize above.

Deferring this whole track preserves attention for reusable library value while
host-specific integration remains uncertain. Each stage should still leave a
runnable example and deterministic checks; keep examples in the root module so
root test discovery includes them.

## Dependency map

```text
036 ──► 057 ─┬──► 062 ──► 063 ──► 064 ──► 065
058 ─────────┘        ▲        ▲          ▲
061 ──────────────────┘        │          │
060 ───────────────────────────┘          │
034 ──────────────────────────────────────┘
051 ──► (validated by 063)
059, 038, 042, 052 — independent
037 ──► (informs 052)
```

057 and 058 gate 062; 061 gates 063; 060 gates 064; 034 gates 065; 065 is last.
059, 038, 042 and 052 are independent and can run in parallel with any of it.

## Stage table

| Stage | Tickets |
|---|---|
| 1–5 — Shipped Foundation | [006](../issues/006-static-declarative-monitor-shell-with-a-minimal-validated-contract.md), [007](../issues/007-clock-watch-mode-with-independent-collection-and-redraw.md), [008](../issues/008-responsive-declared-box-layout.md), [009](../issues/009-declarative-box-visibility-controls.md), [010](../issues/010-geometry-and-visual-evidence-milestone-before-rich-content.md) |
| Hygiene & Fixes (Shipped) | [021](../issues/021-cmd-help-tty-blocking-in-tests.md), [022](../issues/022-decouple-static-shell-golden-from-evolving-monitor-spec.md), [023](../issues/023-rows-schema-validation-and-negative-controls.md), [026](../issues/026-investigate-monitor-pty-smoke-test-idle-redraw-regression.md), [047](../issues/047-remove-deprecated-uzu-brand-and-default-global-commands-from-cmdbar.md) |
| 6 — Rows & Graphs (Shipped) | [011](../issues/011-aligned-dashboard-rows-and-ansi-safe-truncation.md), [020](../issues/020-review-ticket-011-rows-and-target-state-alignment.md), [024](../issues/024-port-harnez-rograph-primitives-with-provenance.md), [025](../issues/025-integrate-graph-renderers-into-declarative-monitor.md), [012](../issues/012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md) |
| 7 — Simulated live data (Shipped) | [013](../issues/013-deterministic-live-snapshots-and-independent-rolling-histories.md) |
| Reusable measurement and dynamic layout (Shipped) | [028](../issues/028-add-reusable-measurement-and-dynamic-box-layout-primitives.md) |
| Typed fixed-rate file prototype (Shipped) | [027](../issues/027-introduce-first-spec-driven-collector-prototype.md) |
| 8 — Color and glyphs (Parked) | [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md) |
| 9 — Complete simulated targets (Parked) | [015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md), [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md) |
| 10 — Splash & startup screen (Shipped) | [029](../issues/029-declarative-centered-layout-and-viewport-alignment-primitives.md), [030](../issues/030-braille-activity-spinner-and-bracketed-progress-bar-primitives.md), [031](../issues/031-provider-status-pill-cluster-and-lifecycle-state-presentation.md), [032](../issues/032-splash-lifecycle-controller-async-provider-coordination-and-key-dismissal.md), [033](../issues/033-harnez-target-splash-screen-integration-and-golden-tests.md) |
| 11 — External sources (Parked) | [017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md), [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md) |
| 12 — Wider declarative source/action evaluation (Parked) | [019](../issues/019-evaluate-declarative-source-and-action-wiring.md) |
| 13 — Hosted widgets (contract) | [057](../issues/057-hosted-widget-key-contract-child-first-routing-reserved-host-keybinds-and-quit-containment.md), [058](../issues/058-widget-declared-pane-requirements-panerequest.md), [059](../issues/059-fix-mouse-coordinate-convention-mismatch-between-frame-and-tabs-stack-grid.md), [060](../issues/060-periodic-redraw-without-pane-ownership-ticker-interface-and-pane-invalidate.md), [061](../issues/061-themeable-host-provided-theme-propagation-through-composite-widgets.md) |
| 14 — Hosted widgets (example conversion) | [062](../issues/062-example-widget-factories-newwidget-for-split-and-tabs-hosted-loom-demo-mode-headless-bench-smoke.md), [063](../issues/063-convert-filebrowser-example-to-a-hostable-widget.md), [064](../issues/064-convert-splash-and-monitor-examples-to-hostable-widgets.md), [065](../issues/065-ansi-styled-rows-widget-and-treemap-conversion.md) |
| 15 — Compositor & overlays (Shipped) | [066](../issues/066-nested-help-pane-opens-a-second-pane-racing-the-outer-pane-tty-reader.md), [067](../issues/067-animated-loom-background-for-filebrowser-with-proper-compositing.md), [068](../issues/068-formalize-layered-compositor-semantics-and-background-inheritance.md), [069](../issues/069-render-help-as-a-root-level-modal-overlay.md), [070](../issues/070-prevent-text-overflow-in-framework-help-modals.md), [050](../issues/050-support-truecolor-rgb-values-in-theme-specs.md) |

## Stages 13–14 — Hosted widgets

Every `examples/*` app should become "just a widget": a bare `loom.Widget` any
host can construct, theme and run, instead of each example owning and driving
its own `Pane`. This builds on
[055](../issues/055-add-a-tab-panel-widget-for-pane-hosting.md) (the `Tabs`
widget that gave loom a host) and takes direction 1 of
[056](../issues/056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md);
the PTY-hosted variant in 056 stays aspirational and unscheduled.

Stage 13 adds the missing library contract: child-first key routing and quit
containment, with no reserved switch hotkey — `tab`/arrow focus navigation
stays the primary host-level primitive (057); widget-declared terminal
requirements replacing per-example pane knobs (058); one pinned mouse
coordinate convention, which fixes a real pre-existing `Frame`-vs-
`Tabs`/`Stack`/`Grid` bug (059); a single pane-driven tick plus
`Pane.Invalidate` so a widget no longer needs to own the render loop (060); and
host-provided theme propagation (061). Stage 14 converts the examples in
dependency order — split/tabs first as the end-to-end demo and the first smoke
coverage those two ever had, then filebrowser, then the live-data pair
splash/monitor, and finally treemap, which needs 034's ANSI cell parsing and a
new ANSI-preserving rows widget before it can leave `RawScreen` behind.

The compositor prerequisites this stage used to carry implicitly are now
explicit and shipped: see [RootOverlays.md](RootOverlays.md) for the
`paneHelpRequest` hook pattern any host-level modal should reuse, and 068's
`PaintSurface`/`PaintForeground`/`PaintDecoration` layering for how a hosted
widget's background is inherited rather than inferred.

## Verification gates

The shipped 028 measurement/layout foundation is covered by known-column
canaries, content-height, min/preferred/max allocation, fixed sizing,
visibility, wrapping, tiny-terminal and ANSI geometry tests. 013/027 require
deterministic sample-count/timestamp matrices with slower and faster redraw,
changing file fixtures, stale/error cases, bounded retention and joined
shutdown. Retain `make test`, relevant race checks and `make watch-pty` for
implementation delivery. Hosted-widget work additionally needs headless bench
smoke coverage for every converted example (062 onward) and a PTY smoke run,
as 068 established.

The stage descriptions below retain the original scope as historical acceptance
context. Their order is superseded by the Now/Next/Later sequence above.

## Stage 1 — Static monitor shell

Implemented in ticket 006. Run `GOWORK=off go run ./examples/monitor`;
see the [example and contract](../examples/monitor/README.md). The shell uses
declared dimensions, printable ASCII chrome/text, and spec-defined borders.
Schema validation, existing Go tests, baseline geometry tests, and a no-TTY
smoke run pass. This does not complete the broader Stage 5 visual gate.

Build the smallest useful Loom application in show-once mode.

It should render:

- a title bar;
- two empty titled boxes;
- a bottom status bar.

At this initial slice there was no `--watch` mode, live data, graph, or button
behavior. The goal is to establish the application frame, generic box rendering, basic spacing,
and a small declarative UI example with a minimal Go entrypoint.

Include only the shell's declarative vocabulary in an embedded YAML spec with
a companion JSON Schema and validation/loader-fidelity tests. Titles, spacing,
and other declared values must be consumed rather than duplicated in Go. Grow
schema coverage with each stage; a comprehensive framework is not a prerequisite.

Start geometry coverage here: fixed terminal-cell fixtures for chrome, borders,
padding, and child clipping, including tiny bounds. Check terminal positions
independently of the renderer's own width helpers. Expand this baseline through
responsive layout and controls before the full Stage 5 milestone.

## Stage 2 — Basic watch mode and clock

Add a simple watch mode using the current time as the first data source.

The clock is collected independently from rendering. The UI redraws at an
explicit refresh rate, while the time source and redraw loop remain separate
concerns. Show-once mode must continue to work.

This stage establishes the fundamental live-data contract before controls or
domain-specific content are introduced.

## Stage 3 — Responsive box layout

Shipped in 008 as fixed preferred boxes with breakpoint stacking. Content-driven
measurement and min/preferred/max allocation are the new 028 follow-up.

Make the two boxes responsive to terminal width.

- At sufficient width, place the boxes next to each other.
- Below the layout breakpoint, place them in a vertical stack.
- Preserve titles, borders, spacing, and usable inner dimensions in both modes.

The same declared layout should produce both arrangements. It should not
require separate wide and slim UI implementations.

## Stage 4 — Box visibility controls

Add keyboard controls for hiding and showing the boxes, with visible hints in
the title or status bar.

The control model should be declarable, but the first implementation can keep
the action wiring small and explicit. The layout must recompute correctly when
a box is hidden or restored.

## Stage 5 — Visual layout contract and testing milestone

Before adding substantial content, establish a stable way to test geometry.

Promote the early baseline to a canonical terminal-cell fixture for wide, narrow,
and tiny terminals, resize, and hidden/restored boxes. Reference screenshots
vary with data, time, and font, and clipboard copies may be imperfect; fix inputs
and document intentional deviations rather than treating those images as exact
goldens. Independently check terminal positions, including ANSI, combining and
wide characters, child clipping, and border integrity; ANSI replay alone is not
the oracle.

The test system should detect when content damages box borders, causes width
drift, or misaligns rows. The implementation may use a combination of:

- robust ANSI-aware display-width measurement;
- explicit border/rectangle invariants;
- deterministic canvas rendering;
- golden or screenshot-style output comparisons.

This is a milestone requiring visual confirmation when a user is available.
If no user is available, automated geometry and screenshot tests should provide
the gate to the next stage.

## Stage 6 — Static dashboard content and graph primitives

Add the first realistic content using simulated, fixed snapshots.

The target is the Harnez-style dashboard:

- aligned label/value rows;
- determinate bars;
- fixed-width rolling sparklines;
- multiple timelines in one box;
- a split VRAM/GTT row;
- truncation for values that do not fit.

Reuse or copy Harnez's dependency-free `rograph` implementation initially.
Keep collection and application-specific metric names outside the Loom graph
renderer.

## Stage 7 — Simulated live data

Partially implemented in 013; the independent producer-rate matrix remains open.

Connect deterministic simulated data to the dashboard.

The simulation should update snapshots and histories independently from the
render loop. Tests should cover changing values, rolling histories, startup
with short histories, stable widths, and repeated redraws without accidental
history advancement.

This proves changing monochrome content in watch mode. Complete target
acceptance follows color and the Voxi content integration in Stage 9.

## Stage 8 — Color and visual semantics

Add configurable colors after the monochrome geometry is stable.

- Map bar levels through a color range such as green to red.
- Apply equivalent level-aware color to sparkline/timeline output.
- Support Braille and other configured glyph presentations.
- Ensure ANSI styling never changes measured geometry.
- Keep visual values configurable rather than embedding application-specific
  palettes in widgets.

## Stage 9 — Complete simulated target milestone

Demonstrate both Harnez and Voxi targets in show-once and watch modes after
color. Include Harnez's independent VRAM/GTT timelines and Voxi's transcript
feed, active daemon/health content, voice/speed values, history summary,
title/status chrome, visibility hints, bold text, and ANSI-safe ellipsis.

Use deterministic snapshots/events for changing feed and daemon states as well
as metrics. Verify wide/narrow layouts, hidden/restored panes, stable colored
geometry, short histories, and redraws that do not advance sampling. Compare
against the canonical fixtures and document intentional target deviations;
seek visual confirmation when available, with automated geometry/golden checks
as the gate otherwise. This remains the complete simulated UI gate before
broader external adapters, with the scoped 027 file prototype exception above.

## Stage 10 — Splash & startup loading target

Demonstrate the Harnez splash and startup loading screen target
([`HarnezSplashTarget.md`](HarnezSplashTarget.md)) in the example app suite.

- Implement declarative vertical and horizontal viewport centering (029).
- Add braille animation spinners and bracketed dot-matrix progress bars (030).
- Render multi-state provider status pill clusters (`● mic`, `✳ claude`, `֍ codex`, `Λ agy`) with stable spacing and ANSI styling (031).
- Orchestrate asynchronous provider startup tasks, step messaging (`fetching claude...` → `agy done`), and interactive `Esc` skip (032).
- Assemble the integrated example with deterministic golden frame tests (033).

This proves centered viewport composition, interactive early exit, and
smooth transition lifecycles from startup screens into main dashboard views.

## Stage 11 — Real external data sources

027 has pulled forward raw fixed-rate file reads and live retention. The
separate-process file/socket and source-specific work in 017/018 remains later.

Replace simulated producers with small real data-source adapters.

Use Harnez and Voxi as reference implementations for querying Linux metrics,
daemon state, transcript feeds, files, and sockets. Begin with test files and
test sockets filled by a separate process so the source boundary behaves like
the real environment.

The UI must continue to consume snapshots/events rather than knowing how the
data was collected.

## Stage 12 — Declarative data-source and action prototype

027 already consumes embedded typed file/cadence declarations. The remaining
019 evaluation covers wider source mappings and named actions after 017/018;
it must build on that seam rather than introduce a second file declaration.

Evaluate how far data wiring can be expressed in the UI specification.

Possible first experiments include declaring:

- a file watched at a configured frequency;
- a socket or line-oriented source;
- a mapping from source values to a bar or timeline;
- an event source connected to a UI action handler.

This is a feasibility prototype, not a commitment to a large framework. The
goal is to learn which data-source and action concepts belong in Loom's
specification and which should remain Go application code.

## Long-term outcome

The completed target is a declarative Loom application whose document defines
the title/status chrome, responsive boxes, rows, bars, timelines, colors,
visibility controls, and data bindings. Go defines collectors, external
integration, state transitions, and action behavior. The renderer remains
independent of collection frequency and application-specific coordinates.
