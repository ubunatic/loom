# Topic 5: Charm low-level libraries vs loom (ticket 457)

Scope: ultraviolet, x/ansi, x/term, x/input, x/vt, colorprofile, cellbuf. Rows are in `features-charmx.yaml` (21 rows: 17 verified, 4 not verified). Every row has `loom: unknown`; the host fills that in. Fetched on 2026-11-10.

## How to read the verification flags

- `verified: true` means a web-fetch subagent fetched the GitHub or pkg.go.dev page and copied a quote from it. I did not re-fetch the pages myself, so the quote text is as the subagent reported it. The host should spot-check the quotes that matter for decisions.
- Line numbers: `ultraviolet` and `x/vt` line numbers were lost in the fetch, so most rows say `line n/a`. Only `term.go` (exact), `ansi/kitty.go` and `ansi/width.go` (approximate) and `colorprofile/env.go` (approximate) have line numbers.
- `verified: false` means I could not confirm it from a fetched page. These rows must not be cited as facts.

## What these libraries are, in one paragraph each

- **ultraviolet** (`github.com/charmbracelet/ultraviolet`): the low-level primitive layer. It has a flat cell grid (`Buffer`), a child-window model (`Window`), and a diffing renderer (`TerminalRenderer`, `TerminalScreen`). Its README says: "Ultraviolet is a set of primitives for building terminal user interfaces in Go." It is the layer under Bubble Tea v2 and Lip Gloss v2: "Ultraviolet powers Bubble Tea v2 and Lip Gloss v2." Its README also says: "Note: Ultraviolet is in active development. The API may change." This is a pre-release dependency risk. The pkg.go.dev page shows a pseudo-version, not a release tag.
- **x/ansi**: escape-sequence constants and helpers, plus string width. Width is the number of terminal cells a string occupies.
- **x/term**: small wrappers for raw mode, terminal state and window size on a file descriptor.
- **x/input**: a parser that turns raw input bytes into key and mouse events. It is the input side of the Charm stack.
- **x/vt**: a virtual terminal emulator, a Go program that interprets escape sequences and keeps a screen model. It is the closest peer to loom's internal `vt.VT` used by `loom session`.
- **colorprofile**: detects how many colors the output can show (none, ASCII-only, 16, 256, or 24-bit true color), honouring `NO_COLOR` and `CLICOLOR_FORCE`.
- **cellbuf** (`github.com/charmbracelet/x/cellbuf`): a lower-level cell screen with its own `Render` and `Flush` steps.

## Findings that matter for loom

### 1. Two-step render: build the diff, then write

Both uv and cellbuf split drawing into two steps. `Render` computes what changed and puts the bytes in an internal buffer. `Flush` is the step that writes to the terminal. uv's `TerminalScreen` has `Render()` and `Flush()`, and `TerminalRenderer` has `Render(newbuf)` and `Flush()` (pkg.go.dev index, verified). The diff is per line: a row is hashed and only rows whose hash changed are rewritten.

```go
// uv: per-line diff; rows marked drift-prone are always repainted.
// Quote (terminal_renderer.go): "A drift-prone row, or any row when force is set, is repainted;
// every other row is diffed."
```

"Drift" is the library's name for rows whose on-screen state may differ from its model. That is the case that makes a diff unsafe, such as after a width mismatch. Loom's "minimal diff" idea (P1 in ticket 457) can be compared with this rule. The uv code reads more deeply than the README, so the host should read `transformLine` before copying the idea.

### 2. Synchronized output (mode 2026): only uv does it

- uv's `TerminalScreen` has `SetSynchronizedUpdates(enabled bool)`. The setter only stores a flag. In `Flush()`, uv writes `ansi.SetModeSynchronizedOutput` before the buffer and the reset after it. The default is off (verified by reading code, not a doc quote).
- cellbuf's `Screen.Flush` writes its buffer straight to the output. A search of the file for "2026" found nothing (verified).
- x/vt and x/ansi were not checked for mode 2026 beyond the above.

Loom lists mode 2026 as a P1 composition item. The host should check loom's own flush path against uv's rule: wrap only when the buffer is non-empty.

### 3. Width: two rules, grapheme by default

`x/ansi` has a `Method` type with two values, `WcWidth` and `GraphemeWidth` (`ansi/method.go`, verified). `StringWidth` uses `GraphemeWidth`; `StringWidthWc` uses `WcWidth` (`ansi/width.go`, about L64-L80, verified). So the default API counts grapheme clusters (for example, a ZWJ emoji counts as one cell pair, not the sum of its parts), and wcwidth is opt-in.

The vt emulator's `WidthMethod` returns `GraphemeWidth` when the Unicode core mode (2027) is set. This comes from code reading only (`verified: false`). It is the one place where the terminal mode and the width rule meet.

Loom's measure package has `DetectZWJMode`, which probes each terminal for ZWJ (zero-width joiner) behaviour. The host should compare: does loom choose the width rule per terminal, the same way uv chooses between the two methods, or does it use one fixed rule?

### 4. Kitty keyboard flags are named constants

`ansi/kitty.go` names the progressive-enhancement flags: `KittyDisambiguateEscapeCodes` (1), `KittyReportEventTypes` (2), `KittyReportAlternateKeys` (4), `KittyReportAllKeysAsEscapeCodes` (8), `KittyReportAssociatedKeys` (16), and `KittyAllFlags` (the OR of all five). It also has `PushKittyKeyboard`, `PopKittyKeyboard`, and `DisableKittyKeyboard` (`ESC [ > u`) (verified, approximate lines L8-L90).

Loom sends `ESC[>1u`. The value 1 matches `KittyDisambiguateEscapeCodes`. The host should confirm that loom's push and pop match the library's pairs, and that loom pops on exit.

uv's `Key` has `ShiftedCode`, `BaseCode` and `IsRepeat`. Per the doc comments, `ShiftedCode` and `IsRepeat` exist only under the kitty protocol. `BaseCode` is "the key pressed according to the standard PC-101 key layout" (verified). A likely use is layout-independent shortcuts such as Ctrl+C on non-US keyboards; this is my inference, not stated in the source.

### 5. Input parsing and split escapes (needs a canary)

`x/input`'s `parseSequence(buf []byte) (n int, Event Event)` returns the number of bytes it consumed, so unused bytes stay in the buffer (verified). Per the subagent's reading of the code (not stated in the doc), a buffer holding a single ESC byte is returned at once as `KeyEscape`. So an escape sequence split across two reads can be decoded as a bare Esc key, with the rest arriving as plain characters. This is a hypothesis.

Canary to confirm it: feed `ESC` and `[A` in two separate `Write` calls to the input parser, and check whether one `KeyEscape` plus a `[`-string comes out. Run the same bytes through loom's own parser. Only then decide whether it is a bug.

### 6. x/vt is the peer of loom's VT

`vt.Emulator` has `NewEmulator`, `Write`, `Resize`, `Render`, `Paste`, `SendKeys`, `SetCallbacks`, and a scrollback API (verified for the first set). Alternate screen with save cursor (1047/1048) is handled (verified, `csi_mode.go`). Its package doc says it is "a virtual terminal emulator that can be used to emulate a" (verified).

Two things are not verified. The scrollback doc comment and the code disagree about the alternate screen, per a subagent's reading (`verified: false`). And a search snippet says upstream lacks insert/replace mode (IRM); that claim comes from a fork's README, not from charmbracelet, so it is also unverified.

### 7. Terminal size and raw mode

`term.go` (exact lines): `IsTerminal` L11, `MakeRaw` L18, `GetState` L24, `SetState` L29, `Restore` L35, `GetSize` L42. `GetSize` is documented as "returns the visible dimensions of the given terminal." Resize handling (SIGWINCH) is not in this file, so it lives in a platform file the fetch did not cover. This is a gap.

### 8. Color detection

`colorprofile.Detect(output io.Writer, env []string) Profile` (`env.go` L32). `NO_COLOR` lowers the profile to ASCII only when the output is a TTY (L85, `if envNoColor(env) && isatty`). `CLICOLOR_FORCE` raises it to at least ANSI, even without a TTY (L92). The README: "If output is not a TTY ANSI will be dropped entirely." (verified). Text decoration (bold, italic) survives `NO_COLOR`, per the package doc (seen in a search snippet; the verbatim line was not fetched).

Profile constant names are unverified because `profile.go` was not fetched. The README spells `colorprofile.Ascii` while `env.go` spells `ASCII`.

### 9. Ultraviolet stability and dependency

- uv's status: "in active development. The API may change." (README, verified).
- Its pseudo-version on pkg.go.dev is `v0.0.0-20261023060817-a45b7c86333f`, and the page says it is "not in the latest version of its module".
- Advice for loom: do not import uv directly until it has a tagged release. Copy ideas (line hash diff, 2026 wrapping on non-empty flush) rather than depend on it.

## Gaps and what would close them

1. **x/vt maturity.** A mirror calls it "experimental", but the charmbracelet pages read did not say so. Check the `x/vt` directory and any release notes.
2. **Mode 2027 detection in uv.** A lipgloss release note (search snippet) says uv asks the terminal at startup whether it supports mode 2027. Not fetched. Check `terminal_renderer.go` or `terminal_screen.go` for the query.
3. **Line numbers** for uv, x/vt, x/cellbuf and x/input (most rows say `line n/a`). A local clone with `grep -n` would fix this in minutes.
4. **Golden and VT testing**: not researched. The ticket asks for it; this needs a separate pass.
5. **SIGWINCH**: the platform file for `x/term` was not fetched.
6. **Fork and reuse signals**: `tuios.dev/vt` is a copy of x/vt with a ghostty backend and needs Go 1.26 (search snippet). This belongs to the TUIOS topic (topic 1) and is not verified here.
7. **Version pins**: TUIOS's own pins were not checked; the `x/*` rows say `main@unpinned`.
8. **Canary for split ESC** (section 5), not yet run.

## Proposed next steps for the host

1. Spot-check the quotes in `features-charmx.yaml` marked `P1` (about 10 rows).
2. Fill `loom` for each row from `docs/`, `spec/` and code, starting with `ansi.width-method`, `uv.sync-output-2026` and `ansi.kitty-keyboard-flags`.
3. Run the split-ESC canary (section 5) against loom's input parser before filing anything.
4. Close gaps 1-3 with a local clone; gaps 4-6 need another searcher pass.

No tickets filed. No loom code or other files touched. Only the two files under `docs/studies/tuios/` were written, and nothing was committed.

## Rows written, verified, open gaps

- Rows written: 21 (`features-charmx.yaml`).
- Rows verified: 17 (quote from a fetched page). Not verified: 4 (`vt.widthmethod-mode-2027`, `vt.scrollback-api`, `vt.insert-mode-irm`, `colorprofile.profile-constants`).
- Open gaps: x/vt maturity; uv mode 2027 startup query; line numbers for 4 libraries; golden and VT testing; SIGWINCH; fork signals (tuios.dev/vt); version pins; split-ESC canary.

## Sources

- ultraviolet README: https://github.com/charmbracelet/ultraviolet
- ultraviolet `terminal_renderer.go`: https://github.com/charmbracelet/ultraviolet/blob/main/terminal_renderer.go
- ultraviolet `terminal_screen.go`: https://github.com/charmbracelet/ultraviolet/blob/main/terminal_screen.go
- ultraviolet `key.go`: https://github.com/charmbracelet/ultraviolet/blob/main/key.go
- ultraviolet pkg.go.dev: https://pkg.go.dev/github.com/charmbracelet/ultraviolet
- x/ansi `method.go`: https://github.com/charmbracelet/x/blob/main/ansi/method.go
- x/ansi `width.go`: https://github.com/charmbracelet/x/blob/main/ansi/width.go
- x/ansi `kitty.go`: https://github.com/charmbracelet/x/blob/main/ansi/kitty.go
- x/vt directory: https://github.com/charmbracelet/x/tree/main/vt
- x/vt `emulator.go`: https://github.com/charmbracelet/x/blob/main/vt/emulator.go
- x/vt `csi_mode.go`: https://github.com/charmbracelet/x/blob/main/vt/csi_mode.go
- x/cellbuf `screen.go`: https://github.com/charmbracelet/x/blob/main/cellbuf/screen.go
- x/term `term.go`: https://github.com/charmbracelet/x/blob/main/term/term.go
- x/input `parse.go`: https://github.com/charmbracelet/x/blob/main/input/parse.go
- colorprofile README: https://github.com/charmbracelet/colorprofile
- colorprofile `env.go`: https://github.com/charmbracelet/colorprofile/blob/main/env.go
- Search-only (not fetched): unixshells/vt-go https://github.com/unixshells/vt-go; tuios.dev/vt https://pkg.go.dev/tuios.dev/vt
