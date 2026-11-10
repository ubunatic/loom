# 462 — Pane consumes terminal replies centrally instead of dispatching them as key events

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: 458, 461, 405

---

## 1. Problem & Motivation
Pane hands terminal replies (DECRPM, DA, CPR) to the widget tree as `KeyEvent{Report: true}`,
and each widget must drop them itself (`event.go` `KeyEvent.Report` doc). Only TextEntry checks
(`text_entry.go:460`); `loom session` filters in its own loop. Any other widget, and every
third-party widget, sees an empty key per reply. Worse, some replies are not classified at all:
kitty flags `CSI ? 1 u` and the 2048 resize report `CSI 48;… t` arrive as `KeyEvent{}` with
`Report == false`. And a reply split past the 50 ms escape timeout reaches widgets as text keys
(`1`, `$`, `y`), which a TextEntry would insert. ultraviolet decodes these as typed events
(`ModeReportEvent`, `KeyboardEnhancementsEvent`, `WindowSizeEvent`) and never as keys; it shares
the late-fragment problem. Evidence: 458 batch 3, `docs/studies/tuios/report-queries.md`
(synthetic reply matrix, V1).

This blocks safe capability queries (461 and query-before-enable for 2026 and kitty).

458 batch 8 (`docs/studies/tuios/report-desktop.md`, canaries C, C2, `TestScanReplies`): the
scanner has no OSC state. A complete OSC 11 reply (BEL or ST) becomes `alt-]` plus text keys, and
a TextEntry ends up with `11;rgb:1e1e/1e1e/1e1e`. Complete `CSI I`, `CSI O` and `CSI ? 997 ; 1 n`
become empty, unclassified keys. Unsolicited replies count too: a reply loom never asked for
(e.g. to a query from an earlier program) must not reach widgets either.

## 2. Technical Specification / Findings
- Fix the library, not the caller: Pane (the input loop in `pane_run.go`) consumes replies and
  never forwards them to widgets. A typed channel or callback can deliver them to whoever asked
  (startup negotiation, `Pane` API), not via the widget key path.
- Classify all reply forms loom may ask for: DECRPM, DA1/DA2, CPR, DSR, kitty `CSI ? flags u`,
  XTVERSION `DCS > | … ST`, OSC 10/11 colour replies, `CSI 48 ; … t`, focus `CSI I/O` when enabled.
- Late fragments: while a query is outstanding, extend the reply window (tcell and TUIOS end the
  startup batch on the DA1 reply, with a 1 s / 300 ms backstop) instead of the 50 ms key timeout.
  Define what happens to user keys typed during that window (keep, in order). tcell also keeps
  a bare Esc at 50 ms but waits 1 s for an incomplete sequence (458 batch 4, report-input.md);
  consider the same split for Pane's escape timeout.
- Event routing ticket: start on terra:med (project rule).

## 3. Implementation & Verification Plan
/goal No terminal reply reaches a widget as a key event, in complete, split and late cases while
a query is outstanding, and user keys typed around replies arrive unchanged and in order; stop
and report when blocked on an API decision or denied permission, or when no progress is made
(clear the goal and write a failure report).

- Re-check live code first (`event.go` `ScanKey`/`isTerminalReply`, `pane_run.go` input loop,
  `internal/session` filter, `text_entry.go` Report check).
- Port the batch 3 matrix (`scratchpad/b3-canary`, described in the report) into Pane tests with a
  recording widget: replies 0–4, split, late, interleaved `a`/`b`, kitty, 2048, DA1.
- Remove the per-widget Report checks once Pane filters, and update `docs/Widgets.md` /
  `docs/TuiInput.md`.
