# Batch 4: input, kitty and paste — findings

The canaries found that Pane loses kitty event type, alternate-key and associated-text information; TextInput and TextArea retain controls in pasted text, and RichTextEdit does not consume bracketed paste. These findings are bounded by the tests below. The comparative ledger remains incomplete; the source evidence and new PTY checks are recorded below.

## Pins and scope

Loom `e26942a8a935d923e3bc2684ff0085d3e73d4241` (the requested origin revision). On 2026-11-10 `go list -m ...@latest` resolved Bubble Tea v2.1.1 `67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7`, ultraviolet `a45b7c86333f9dfedb95f81064e80259b02937f0`, x/input v0.3.7 `15a1484e6848922233d6180a57d4734cd59573ee`, x/ansi v0.11.9 `faa4adf95555331dfb439b716cc83767ecdadeb4`, tcell v3.6.0 `4bed18057fce88934c9b9a576b5f83632196de11`, tview v0.42.0 `5ce6a2b588145610060000a4f75d7e2af081a794`, and TUIOS v0.12.1 `850335c0c08bb8b802406fcc7f254355c89af04a`. The TUIOS v0.12.1 dependency `tuios.dev/vt` resolves to `f3383d2005bdb2ddb11b7a5c41947da994e5d04d`; it is its deployed dependency, not the independently current x/ansi. Release tags were independently dereferenced with `git ls-remote`; UV HEAD was also re-resolved.

## Capability ledger (partial)

P = proved at the named boundary; p = partial; D = disproved; ? = not established. No status below implies all input forms work.

| Feature | Pane | Bubble Tea | UV reader | x/input | x/ansi | tcell v3 | tview | TUIOS |
|---|---|---|---|---|---|---|---|---|
| split CSI, SS3, UTF-8; ESC timeout | p, V1/matrix | ? | p, matrix | ? | ? | ? | ? | ? |
| CSI-u flags 1/2/4/8/16 representation | D, V2 | p, V4 | p, matrix | ? | ? | p, V6 | ? | ? |
| push/pop, alternate screen, Suspend | p, V3, source only | p, V4 | ? | ? | p, encoder only | ? | ? | ? |
| legacy fallback | p, up arrow matrix | ? | p, up arrow matrix | ? | ? | ? | ? | ? |
| paste frame and split end | P, matrix | p, V4 | P, matrix | ? | ? | ? | ? | ? |
| oversized paste then key | P for matrix size | ? | P for matrix size, unbounded policy not established | ? | ? | ? | ? | ? |
| sanitize paste | D at Pane boundary | ? | D for controls in tested payload | ? | ? | ? | ? | p, V5 guest boundary |
| paste cannot invoke shortcuts | p: dispatched separately at root | ? | p: distinct typed events | ? | ? | ? | ? | ? |

Unknown cells are unfinished research, not assertions of missing support. Additional traces V7–V10 below narrow several of these cells but do not complete the ledger. Pane's widget statuses are separate below.

## Shared byte matrix

Harness: `/tmp/claude-1000/-home-uwe-projects-loom/9e4d0913-fa4b-4cb2-aec0-0c51a6bd4a87/scratchpad/b4-canary`, `go run .` and `go run . oversize`. Real Pane.Run runs under a disposable PTY with a recording widget; UV uses its actual TerminalReader on a pipe. Each side has a five-second context; no window. Split inputs prefix `a`; continuation is sent when `a` is observed, with no artificial pause. The original two writes could coalesce; the subsequent acknowledged one-byte read canary below crosses every split in the listed representative sequences. Only the late ESC case uses a 130 ms timer to cross the nominal 50 ms deadline. All shown cases completed without error. `K(x)` means Pane KeyEvent, `U(x)` UV KeyPressEvent; `a` prefix and the final harness `z` sentinel are omitted.

| Bytes, split positions | Pane | UV |
|---|---|---|
| `ESC[1;5A`, each of 1–5 | K(ctrl-up) | U(ctrl-up) |
| `ESCOA`, 1–2 | K(up) | U(up) |
| UTF-8 `é`, `€`, `😀`, every byte split | K(text=corresponding rune) | U(text=corresponding rune) |
| `ESC[97;5u`, 1–6 | K(ctrl-a) | U(ctrl-a) |
| `ESC[97;1:1u` press | K(text=a) | U(text=a, repeat=false) |
| `ESC[97;1:2u` repeat | K(text=a) | U(text=a, repeat=true) |
| `ESC[97;1:3u` release | K(text=a) | KeyRelease(text=a) |
| `ESC[97:65:113;2u` | K(shift-a) | U(code=97, shift=113, base=113, text=q, shift modifier); actual returned fields, not a correctness claim |
| `ESC[97;1;120:121u` | K(text=a) | U(code=97, text=xy) |
| legacy `ESC[A` | K(up) | U(up) |
| ESC then `[A` immediately | K(up) | U(up) |
| ESC then `b` after 130 ms | K(esc), K(text=b) | U(esc), U(text=b) |
| paste `A\x03\r\x1b\r\n\u0085B` | Paste(9 bytes, same payload) | PasteStart, Paste(same payload), PasteEnd |
| paste `AB`, end marker split at `ESC[20` / `1~` | Paste(AB) | PasteStart, Paste(AB), PasteEnd |
| 1,052,672 `x` bytes, terminator then `b` | Paste(first 1,048,576 bytes), K(b) | PasteStart, Paste(all 1,052,672 bytes), PasteEnd, U(b) |

The CSI-u byte forms exercise syntax corresponding to enhancement flags, not a terminal negotiating those flags. No synthetic flags were enabled in Loom; actual terminal behavior and a timed check immediately before 50 ms remain unverified; the synthetic flag forms and mode restoration canaries are below. The oversized case's marker was not split; the smaller split-end case tests that condition separately. UV's displayed `ShiftedCode` on the alternate form is an observation needing a dedicated correctness check.

## Per-widget paste

A direct `DispatchPasteEvent` canary used the same control-bearing payload. The later real-Pane wrapper canary uses an expanded payload and counts key and submit callbacks; neither exercises container focus routing. TextEntry's OnSubmit counter remained zero; no other shortcut callback was instrumented. RichTextEdit initially had a nil-document fixture panic, corrected for the reported result.

| Widget | Resulting value | Result / shortcut evidence |
|---|---|---|
| TextEntry | `A\n\nB` | consumed; OnSubmit 0 |
| TextInput | `A\x03 \x1b \u0085B` | consumed; shortcut not instrumented |
| TextArea | `A\x03\r\x1b\r\n\u0085B` | consumed; shortcut not instrumented |
| RichTextEdit | empty document | ignored; no paste handler; shortcut not instrumented |

## Loom implementation

### V1 Buffering and paste

[Pane input loop](https://codeberg.org/ubunatic/loom/src/commit/e26942a8a935d923e3bc2684ff0085d3e73d4241/pane_run.go#L688-L692):

```go
ke, used, ok := scanKey(raw)
if !ok {
    pending = append([]byte(nil), raw...)
    pendingC = time.After(escKeyTimeout)
    pendingDeadline = time.Now().Add(escKeyTimeout)
```

[Scanner](https://codeberg.org/ubunatic/loom/src/commit/e26942a8a935d923e3bc2684ff0085d3e73d4241/event.go#L608-L657) waits on incomplete UTF-8, CSI and SS3. [Pane paste loop](https://codeberg.org/ubunatic/loom/src/commit/e26942a8a935d923e3bc2684ff0085d3e73d4241/pane_run.go#L629-L686) directs completed frames to `DispatchPasteEvent`; it bounds payload at 1 MiB and discards overflow through the terminator, retaining later bytes. This is separation from key routing, not sanitation of its payload.

### V2 Kitty representation

[CSI-u decoder](https://codeberg.org/ubunatic/loom/src/commit/e26942a8a935d923e3bc2684ff0085d3e73d4241/event.go#L376-L392):

```go
fields := strings.SplitN(string(b[2:len(b)-1]), ";", 2)
codeText := fields[0]
if i := strings.IndexByte(codeText, ':'); i >= 0 {
    codeText = codeText[:i]
}
```

The alternate codes are discarded. `KeyEvent` has only Key, Text and Report; no press/repeat/release field. Matrix shows release being delivered as a press-like `a`, and associated text ignored. Pane opts into flag 1 only; this failure is observable for injected sequences and does not assert Loom requests flag 2.

### V3 Mode and widgets

[Mode constants and calls](https://codeberg.org/ubunatic/loom/src/commit/e26942a8a935d923e3bc2684ff0085d3e73d4241/pane_modes.go#L143-L165) push `CSI >1u` and pop `CSI <u`. [Screen transitions](https://codeberg.org/ubunatic/loom/src/commit/e26942a8a935d923e3bc2684ff0085d3e73d4241/pane_screen.go#L235-L245) pop before switch and push after; [Suspend](https://codeberg.org/ubunatic/loom/src/commit/e26942a8a935d923e3bc2684ff0085d3e73d4241/pane_suspend.go#L180-L215) leaves terminal and [resume](https://codeberg.org/ubunatic/loom/src/commit/e26942a8a935d923e3bc2684ff0085d3e73d4241/pane_suspend.go#L234-L256) re-enables recorded state. This is source evidence, not a nested-stack runtime canary.

[TextEntry normalization](https://codeberg.org/ubunatic/loom/src/commit/e26942a8a935d923e3bc2684ff0085d3e73d4241/text_entry.go#L1101-L1125) normalizes CR/CRLF and removes controls. [TextInput](https://codeberg.org/ubunatic/loom/src/commit/e26942a8a935d923e3bc2684ff0085d3e73d4241/textinput.go#L54-L61) only replaces line breaks, then calls ConsumeKey with text; [TextArea](https://codeberg.org/ubunatic/loom/src/commit/e26942a8a935d923e3bc2684ff0085d3e73d4241/textarea.go#L81-L96) inserts event.Text directly. RichTextEdit has no ConsumePaste in the searched `.go` files; scope is the current checkout.

## Upstream mechanisms (partial)

### V4 Bubble Tea and UV

[Bubble Tea event translation](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/input.go#L20-L39) maps UV press/release and paste to distinct messages. [Enhancement flags](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/cursed_renderer.go#L932-L945) are constructed from view settings; [renderer](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/cursed_renderer.go#L422-L447) pushes/sets and queries. [UV reader paste](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_reader.go#L289-L318) buffers paste; exact tested deliveries are above. This does not prove Bubble Tea's deployed UV dependency has every property of the separately pinned latest UV.

### V5 TUIOS guest paste

[TUIOS PastePayload](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/terminal/paste.go#L11-L21) calls its VT dependency's sanitizer before optional framing. [Actual sanitizer](https://github.com/Gaurav-Gosain/tuios/blob/f3383d2005bdb2ddb11b7a5c41947da994e5d04d/vt/paste.go#L33-L45):

```go
case r == utf8.RuneError && size == 1:
    continue
case r == '\t' || r == '\n' || r == '\r':
case r < 0x20, r >= 0x7f && r <= 0x9f:
    continue
```

It drops invalid UTF-8 and C0/C1 other than tab, LF and CR. This is guest-bound payload sanitation, not proof that the application never routes a shortcut from its own host input.

### V6 tcell

[tcell CSI-u path](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/input.go#L1454-L1476) uses associated text when supplied; this alone does not prove its repeat/release representation or screen restoration. Complete source traces of tcell, tview, x/input and x/ansi remain outstanding.

## Owner check

`/tmp/claude-1000/-home-uwe-projects-loom/9e4d0913-fa4b-4cb2-aec0-0c51a6bd4a87/scratchpad/manual-input.py` is a self-contained raw `/dev/tty` check. In foot and kitty, run `python3 -I <script> foot` or `... kitty`, then follow prompts: press/release `a`, hold/release `a`, Shift+a, Ctrl+i, type `a`, paste two-line `hello`/`world`. Each step records 10 seconds and raw bytes to `manual-input-<name>.txt`. It pushes each selected flag and pops it, enables/disables bracketed paste, and restores tty attributes in `finally`. This adds evidence about actual terminal release and text output; it was syntax-checked with `python3 -I -m py_compile`, not run. No claim is made that the reset restores a pre-existing bracketed-paste setting.

## Stage 4 proposals

- In `features.md`, `summary.md`, `features-terminals.yaml` and `features-charmx.yaml`, split request, event representation, stream buffering, per-screen restoration and paste sanitation rows; retain incomplete cells until the remaining traces are done. Name exact YAML row IDs after checking the existing inventories; this report has not done that mapping.
- Candidate: a coherent per-widget bracketed-paste policy and RichTextEdit support, with control and shortcut tests. The TextEntry no-effect pass-through issue 456 is distinct; this concerns pasted data and event delivery.
- Candidate: represent kitty repeat/release, alternate keys and associated text before considering additional requested flags; cover fallback and screen/suspend. Query reply routing overlaps 462, and the kitty query portion of startup negotiation overlaps 461; neither issue covers all representation and paste behavior.
- API stability is P3 while Loom is alpha. No tickets filed or implementation selected.

## Flow notes

Only the named harness and this report were written, plus the owner script; no Loom code, tickets or commits were changed. The new canaries close the acknowledged-read, synthetic per-flag, PTY lifecycle, and root widget paste checks below. Outstanding: a complete per-feature, per-library ledger with V excerpts; exhaustive flag combinations; runtime verification of actual terminal behavior; and paste through container focus routing rather than the recording wrapper. The matrix tests Pane and UV, not every compared library. These gaps prevent treating this batch as complete.

Host decision (2026-11-10): the batch is accepted with the gaps above. The agent stopped early
twice and was not resumed again. Host canary: `Canvas.Write` drops control bytes and escape
sequences from cell text (`a\x1b[31mb\x03c\x07d` flushes as `abcd`), so pasted controls stay in
widget values but never reach the terminal. Filed: 464 (shared paste sanitiser, RichTextEdit
paste). Added to 462: tcell's 1 s timeout for incomplete sequences. Not ticketed: kitty event
representation, unreachable while Pane pushes flag 1 only. The owner script was not run.


## Additional read-boundary and protocol canaries

`go build -overlay=overlay.json -o probe .` builds the real Pane loop with a disposable copy of `pane_suspend.go` instrumented immediately after `tty.Read`; the only extra operation writes the returned bytes to an acknowledgement pipe. The driver writes the next byte only after verifying that acknowledgement. No checkout file is altered. `./probe boundary` confirmed each byte separately read for `ESC[1;5A`, `ESCOA`, UTF-8 é/€/😀, `ESC[97;5u`, and `ESC[200~AB ESC[201~` including each end-marker byte. UV receives the same streams as one-byte `io.Pipe` writes, each of which returns only when its reader has accepted the byte. All checks passed and produced the existing matrix events. Each adjacent split is thereby crossed in the same run. The test does not prove behavior when an inter-byte delay reaches the 50 ms deadline.

`./probe flags` returned these additional results; none of these writes requests a flag from the terminal:

| Flag-specific input | Pane | UV |
|---|---|---|
| 1: `ESC[27u` | esc | press code 27 |
| 2: `ESC[97;1:1u`, `:2u`, `:3u` | a, a, a | press a, repeat a, release a |
| 4: `ESC[97:65:113;2u` | shift-a | code 97, shifted 113, base 113, text q |
| 8: `ESC[97u`, `ESC[13u` | a, enter | a, enter |
| 16: `ESC[97;1;120:121u` | a | text xy |

This samples the forms, not every combination of flags. The 113 value in both UV alternate fields is the actual event, not a claim of correct protocol interpretation.

`./probe lifecycle` recorded these PTY output sequences in order, with `a` switching inline→alt, `i` switching alt→inline, `s` invoking Suspend and returning, and `z` quitting. Draw acknowledgements separate operations; no sleeps or visible window:

```text
startup:      CSI ?2004h  CSI >1u
enter alt:    CSI <u      CSI ?1049h  CSI >1u
leave alt:    CSI <u      CSI ?1049l  CSI >1u
suspend:      CSI ?2004l  CSI <u
resume:       CSI ?2004h  CSI >1u
exit:         CSI <u      CSI ?2004l  CSI ?2004l
```

The fixture observes one screen switch each way and a successful Suspend from inline; the final duplicate paste reset is observed. It does not establish preservation of a prior external paste-mode state or Suspend while already in alt mode.

`./probe realwidgets` sent `ESC[200~A\x03\r\x1b\tB\r\n\u0085C ESC[201~` through a live Pane. A root wrapper counts all delivered keys and TextEntry submit callbacks and forwards paste directly to the target widget; the `z` sentinel is excluded. Results:

| Target | Resulting value | Keys / submit callbacks |
|---|---|---|
| TextEntry | `A\n    B\nC` | 0 / 0 |
| TextInput | `A\x03 \x1b\tB \u0085C` | 0 / n/a |
| TextArea | `A\x03\r\x1b\tB\r\n\u0085C` | 0 / n/a |
| RichTextEdit | `[{Spans:[]}]` | 0 / n/a; paste ignored |

This establishes isolation for this Pane→wrapper→widget route and payload; it does not exercise a focus container or all widget-configured actions.

## Additional source traces

### V7 x/input: independent reader is not UV

[x/input readEvents](https://github.com/charmbracelet/x/blob/15a1484e6848922233d6180a57d4734cd59573ee/input/driver.go#L101-L161) performs one read and parses that read's buffer, retaining a paste buffer but no general incomplete-sequence buffer. In particular its paste end detection parses the current buffer before saving individual paste bytes, so an end marker split across reads cannot be assumed to close the frame. Excerpt:

```go
nb, err := d.rd.Read(d.buf[:])
buf := d.buf[:nb]
nb, ev := d.parser.parseSequence(buf[i:])
```

[x/input kitty parser](https://github.com/charmbracelet/x/blob/15a1484e6848922233d6180a57d4734cd59573ee/input/kitty.go#L214-L305) retains repeat/release, alternate fields and associated text. Its paste reader decodes runes and drops `utf8.RuneError`, with no visible size bound in that path; it does not sanitize valid control characters. [Tests](https://github.com/charmbracelet/x/blob/15a1484e6848922233d6180a57d4734cd59573ee/input/key_test.go#L633-L660) explicitly expect Ctrl+C inside paste. Output mode lifecycle is the host's responsibility.

### V8 x/ansi: encoding primitives, not an input loop

[x/ansi kitty functions](https://github.com/charmbracelet/x/blob/faa4adf95555331dfb439b716cc83767ecdadeb4/ansi/kitty.go#L43-L84) encode set/push/pop, and [paste constants](https://github.com/charmbracelet/x/blob/faa4adf95555331dfb439b716cc83767ecdadeb4/ansi/paste.go#L3-L7) define delimiters. Excerpt:

```go
func PushKittyKeyboard(flags int) string
func PopKittyKeyboard(n int) string
```

The encoder by itself neither runs a reader nor owns screen/suspend lifecycle; treating those functions as full application support would overstate the boundary.

### V9 tcell and tview

[tcell timeouts](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/input.go#L68-L82) distinguish 50 ms bare Escape from one second for an incomplete sequence. [CSI-u parsing](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/input.go#L1349-L1397) exposes pressed/repeat and a physical key; [associated text](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/input.go#L1016-L1033) filters C0/C1. [Screen](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/tscreen.go#L194-L197) offers basic and advanced flags; it uses `CSI =31u` and `CSI =0u`, so this is not evidence of a kitty push/pop stack. Paste is [start/end events](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/paste.go#L21-L27), with payload delivered as key events; host routing matters.

[tview's application loop](https://github.com/rivo/tview/blob/5ce6a2b588145610060000a4f75d7e2af081a794/application.go#L399-L467) collects only rune, Enter and Tab while paste mode is active before key capture or Ctrl+C handling, then dispatches a paste handler. Excerpt:

```go
if pasting {
    switch event.Key() {
    case tcell.KeyRune:
        pasteBuffer.WriteRune(event.Rune())
```

This tview release actually depends on tcell **v2.8.1**, not independently pinned v3; do not transfer v3 capabilities to it. Its `strings.Builder` has no explicit cap in the cited loop.

### V10 deployed dependency boundaries and proposal rows

[Bubble Tea v2.1.1 uses UV 6b8d4baf91b4](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/go.mod#L9), not separately tested latest UV a45b7c86333f. [TUIOS v0.12.1 uses UV 0277a179edd9](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/go.mod#L18); its guest-side sanitizer V5 cannot establish host input behavior. Map proposals to existing rows `term.kitty-keyboard.flags`, `term.kitty-keyboard.legacy-exceptions`, `ansi.kitty-keyboard-flags`, `input.parsesequence-consumed`, `uv.key-kitty-fields`, `bt.paste.separate-msgs`, `bt.key.press-release`, `bt.key.isrepeat-basecode`, `tview.application.paste-no-navigation`, `tview.application.paste-keys-not-captured`, `tcell.screen.paste`, `tcell.kitty-keyboard.enable`, and `tuios.bracketed-paste-sanitize`. Add separate rows for read boundaries, size policy, per-widget sanitation and lifecycle rather than implying those existing rows cover them.
