# Topic 2: Bubble Tea v1 and v2 against loom

Ticket 457, searcher output. Evidence file: `features-bubbletea.yaml` (38 rows, all `verified: true`).
`loom` column is left `unknown` for the host.

## Method and limits

- Sources were the GitHub raw files of the release tags, read with `gh api`:
  Bubble Tea **v2.1.1** (module `charm.land/bubbletea/v2`) and **v1.3.10** (module `github.com/charmbracelet/bubbletea`).
  Upgrade guide: `UPGRADE_GUIDE_V2.md` at v2.1.1. Teatest: `charmbracelet/x` at commit `c83711a1`, `exp/teatest/v2`.
- Every `evidence_quote` was checked with `grep -F` against the fetched file. All 38 matched.
- I did **not** run a web search and did **not** read issues, discussions or blog posts. Claims about user
  experience, performance numbers and terminal behaviour are not covered here.
- Version rule: v1 and v2 are separate rows (`version` field). The two APIs differ a lot; do not mix them.

## What Bubble Tea is (for the host)

Bubble Tea is the Elm-style loop used by most Charm apps: a `Model` has `Init`, `Update(msg)` and `View`.
Input, timers and commands become messages that go through one event loop. Loom is not built on it, so the
useful question is which mechanisms it solved that loom might copy or contrast.

## Findings, grouped by ticket priority

### P1 (ticket-level items with a direct Bubble Tea answer)

1. **Backpressure: `Send` blocks.** The message channel is unbuffered (`tea.go#L620`, v2) and `Send` waits
   until the loop takes the message (`tea.go#L1215-L1220`). v1 does the same (`tea.go#L774-L779`). So a
   high-rate producer is throttled by the loop, not dropped or coalesced. Loom's `Pane.Post` is the same
   question; compare how it behaves under a flood (not checked here, needs a canary).
2. **Rendering: frame-rate cap and cell renderer.** The v2 renderer is cell-based on `ultraviolet`
   (`cursed_renderer.go#L13`). The v1 renderer compares the whole frame as a string and skips the flush when
   it is unchanged (`standard_renderer.go#L161-L165`). Default rate is 60 FPS with a 120 cap
   (`tea.go#L496-L499`, v2 wording; v1 `defaultFPS = 60`).
3. **Mode 2026 (synchronized output)**: v2 wraps updates in synchronized output when the terminal supports it,
   but not cursor-only changes (`cursed_renderer.go#L561-L594`). v1 has no 2026 handling (negative grep of
   v1.3.10 for `kitty`, `2026`, `2027` found nothing). Loom's mode 2026 is a ticket P1 item, so this is a
   direct comparison.
4. **Suspend and terminal hand-off.** `tea.Suspend()` returns `SuspendMsg` (`tea.go#L586-L587`). For external
   programs, `ReleaseTerminal` restores the terminal and stops input; `RestoreTerminal` reattaches
   (`tea.go#L1365-L1392`).
5. **Crash restore.** Panics in the loop go through `recoverFromPanic` (`tea.go#L1317`), opt-out via
   `WithoutCatchPanics` (`options.go#L83`).
6. **Resize.** A goroutine handles SIGWINCH and feeds `WindowSizeMsg` into the loop (`tea.go#L709`). In v2 the
   request is `tea.RequestWindowSize` (`commands.go#L173`); v1 `WindowSize` was renamed (upgrade guide).

### P1-P2 (input model)

7. **Kitty keyboard.** v2 always pushes Kitty flag 1 (disambiguation) and sets modifyOtherKeys 2 at start
   (`cursed_renderer.go#L163-L170`, `#L932-L933`). Event types, alternate keys, all-keys-as-escapes and
   associated text are opt-in through `View.KeyboardEnhancements` (`tea.go#L203`). The terminal's answer
   arrives as `KeyboardEnhancementsMsg` with `SupportsEventTypes()` and friends (`keyboard.go#L7-L40`).
   Loom already sends `ESC[>1u`; the gap to check is the opt-in flags and the support message.
8. **Key press and release are separate types** (`key.go#L191`, `#L224`). Release events need Kitty event types.
9. **Canonical key strings.** `KeyPressMsg.Keystroke()` always writes modifiers as ctrl, shift, alt, meta,
   hyper, super; the doc example is `ctrl+shift+alt+a` (`key.go#L211`). That makes key tables plain string
   compares.
10. **Ctrl+C is now a string match.** v2 has no `KeyCtrlC` constant; the guide says to match
    `msg.String() == "ctrl+c"` or `Code` plus `Mod` (`UPGRADE_GUIDE_V2.md#L149`).
11. **Paste is its own message in v2**: `PasteStartMsg`, `PasteMsg`, `PasteEndMsg` (`paste.go#L14-L16`).
    In v1, pasted runes are a `Key` with `Paste bool` (`key.go#L54-L58`), and bubbles key bindings are kept from
    matching shortcuts on paste (`key.go#L73-L78`). Loom's huge-paste robustness is a P1 item; neither version
    shows a size limit in the parts read (not checked further).
12. **Mouse is typed in v2**: `MouseClickMsg`, `MouseReleaseMsg`, `MouseWheelMsg`, `MouseMotionMsg`, all read
    through `Mouse()` (`mouse.go#L83-L128`). v1 has one `MouseMsg` struct with `Ctrl`, `Shift`, `Alt` fields.
13. **Mouse mode per view.** v2: `View.MouseMode` with None, CellMotion, AllMotion (`tea.go#L191`). Motion
    reporting is what lets a drag leave the widget bounds, which is the ticket's pane-capture question.
    v1 uses program options (`options.go#L137`).
14. **Per-view mouse hook.** `View.OnMouse` gets the mouse event with the last rendered content in mind
    (`tea.go#L126`). This is a pattern for hit testing without a global router; loom's claim-cell model is the
    comparison (P2 here).
15. **Focus.** `FocusMsg` and `BlurMsg` (`focus.go#L5-L9`), enabled by `View.ReportFocus` (`tea.go#L184`).
    v1: `WithReportFocus` (`options.go#L248`).
16. **Interrupts.** With a TTY, Ctrl+C is a key; SIGINT or SIGTERM from outside (or non-TTY input) become
    `InterruptMsg` or `QuitMsg` (`tea.go#L670-L671`). `Run` returns `ErrInterrupted` (v1 `tea.go#L35-L37`).

### P2 (desktop, theme, testing, comparisons)

17. **Clipboard**: `SetClipboard`, `ReadClipboard`, `SetPrimaryClipboard`, `ReadPrimaryClipboard`
    (`clipboard.go#L30-L70`). Loom has OSC 52 write; read and primary selection are the gap to check.
18. **Colours**: `View.BackgroundColor` and `ForegroundColor` (`tea.go#L135`), `WithColorProfile` for tests
    (`options.go#L160`). Bubble Tea does not detect dark or light itself; only the view fields were checked.
19. **Testing**:
    - `teatest.NewTestModel`, `WaitFor`, `RequireEqualOutput` (`teatest.go#L129`, `#L291`).
    - `WaitFor` polls with `time.Sleep(wf.CheckInterval)` (`teatest.go#L109`). That breaks loom's no-sleep
      rule in `docs/TestTiers.md`. Use it as a contrast, not as a model.
    - `WithWindowSize` and `WithColorProfile` make golden output deterministic (guide `#L553-L557`).
20. **Key migration**: v1 program options and commands were replaced by `View` fields (guide `#L286-L298`).
    Not a loom issue, but relevant if loom ever reads TUIOS-style code.

### P3

Tick and Every (`commands.go#L154`, `#L102`), Batch and Sequence (`commands.go#L15`), progress bar in `View`
(`tea.go#L147`, wire format not checked), cursor shape (`cursor.go#L12`), `IsRepeat` and `BaseCode` on keys
(guide `#L192`).

## Where Bubble Tea is weak, per the source (for the summary)

- v1 has no Kitty keyboard, no mode 2026 and a whole-frame string diff (rows 2, 3, 7 above).
- The unbuffered `Send` gives no built-in coalescing of floods (row 1). Coalescing was not checked.
- Teatest waits by polling with sleeps (row 19).
- v2 is still moving: 2.1.1 is the latest tag (published 2026-10-28; 2.0.7 to 2.1.1 were all in 2026). The
  API changed a lot between v1 and v2 (guide).

## Open gaps (not checked, need a canary or a second source)

- Real terminal behaviour of mode 2026 and Kitty flags per terminal (foot, kitty, VTE, ghostty, wezterm).
- Whether `Pane.Post` loses or delays messages under a flood, compared with `Send` blocking. A canary is
  needed; no number is proposed here.
- Bubble Tea issues and discussions on backpressure, mouse drag or paste size. Not read.
- `ultraviolet` API details (cell diff, width model). That is topic 5.
- Bubbles and Huh (topic 3) and Lip Gloss (topic 4) are not covered.

## Rows

- Rows written: 38 (`features-bubbletea.yaml`).
- Rows verified: 38 (quote matched against the fetched file).
- Open gaps: see above. The `loom` column is `unknown` in every row.
