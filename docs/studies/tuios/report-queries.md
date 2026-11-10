# Batch 3: terminal queries and negotiation

The principal gap is reply routing: Pane sends even classified terminal replies to widgets as key events; an incomplete reply that outlives the escape timeout can produce text keys in both Pane and ultraviolet. Session has a separate filter. A complete-buffer parser result alone does not establish safe negotiation.

## Pins and scope

Resolved 2026-11-10 against latest releases, or current HEAD where no release exists. All links below use immutable commits. Loom is the requested origin revision `777d170c4b8f99ae2cd48bb36ddd91dd59572b7f`; Bubble Tea v2.1.1 `67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7`; ultraviolet `a45b7c86333f9dfedb95f81064e80259b02937f0`; tcell v3.6.0 `4bed18057fce88934c9b9a576b5f83632196de11`; x/ansi v0.11.9 `faa4adf95555331dfb439b716cc83767ecdadeb4`; x/input v0.3.7 `15a1484e6848922233d6180a57d4734cd59573ee`; TUIOS v0.12.1 `850335c0c08bb8b802406fcc7f254355c89af04a`.

TUIOS pins older Charm dependencies in its own go.mod, so its observed application code must not be attributed to the separately audited latest Bubble Tea and UV revisions. The standalone Charm columns describe those latest versions; TUIOS's column describes its own calls and guest behavior. The Loom VT is a guest emulator, not the real terminal. Absence claims below are bounded to the named input, startup and cleanup paths, not every conceivable caller.

Protocol sources: [foot 1.28.0 ctlseqs](https://codeberg.org/dnkl/foot/src/commit/ab33c9a19d626f8d0ac0bd1adfdba21946ada948/doc/foot-ctlseqs.7.scd#L292-L351); [Contour synchronized output](https://github.com/contour-terminal/vt-extensions/blob/05050a11e793c8f4362bf4e34a59ed3f7e5105fe/synchronized-output.md#L25-L47); [in-band resize specification](https://gist.github.com/rockorager/e695fb2924d36b2bcf1fff4a3704bd83/a1e61ea1782326e975b6e4cfe0c538bac54c1f42#file-in-band-resize-notifications-md-L22-L56); [kitty keyboard protocol](https://github.com/kovidgoyal/kitty/blob/0273bc12c3dcc835f4da4a0ea88233ace073706e/docs/keyboard-protocol.rst#L311-L336) and [detection](https://github.com/kovidgoyal/kitty/blob/0273bc12c3dcc835f4da4a0ea88233ace073706e/docs/keyboard-protocol.rst#L440-L451). Contour's pinned extensions repository contains synchronized-output.md but no 2048 document; the independently pinned resize specification supplies the 2048 protocol. Foot's “TODO” is the attribution field, not proof that the protocol has no specification. Mode reports 0–4 distinguish unrecognized, set, reset, permanently set and permanently reset. Kitty `CSI ? u` reports *current* flags; push/pop is a distinct operation and a current-flags reply is not a supported-features mask.

## Capability ledger

P = proved for the named boundary, p = partial, D = disproved for the named full capability. “Full capability” means an application-level query/response policy, except in the explicitly parser/encoder-only columns and guest-emulator column. A p in a parser/encoder column means it supplies a relevant building block, not a runtime negotiation. Individual policies and immutable evidence are in V1–V9 below; linked column headings and feature links identify the relevant implementations.

| Feature | [Pane](#v1-pane-input-and-output) | [loom-probe](#v2-probe) | [session / vt](#v3-session-and-vt) | [Bubble Tea](#v4-bubble-tea) | [ultraviolet](#v5-ultraviolet) | [tcell](#v6-tcell) | [x/ansi](#v7-xansi) | [x/input](#v8-xinput) | [TUIOS](#v9-tuios) |
|---|---|---|---|---|---|---|---|---|---|
| 2026 DECRQM and use | p: wraps, no query | D: keys probe only | p: snapshots; reports 0 | p: gated query, handles reset | p: decode and configurable output | D: no startup query policy | p: encode | p: decode | p: host renderer dependency; guest hold |
| 2027 DECRQM | D: no query | D | D: reports 0 | P: set/reset/always-set | P: same, unless override | D: no startup negotiation | p: encode | p: decode | p: guest width; host via dependency |
| 2048 resize | D: report becomes empty key | D | D: unrecognized | p: UV cell-size event; no enable policy shown | p: cell and pixel decoder | P: query, enable, resize, disable | p: encode mode/report | p: generic window-op event | p: guest resize and cell-size consumers; no host 2048 negotiation found |
| kitty query vs push | p: opt-in push/pop, no query | P: query + DA1, then push/pop | p: query answer; single flag slot, no stack | p: push then query, flags event | p: typed reply decoder | p: query then `CSI = flags u`, reset via `=0u` | p: query/push/pop encoders | p: flags decoder | p: keyboard flags through Tea, guest keyboard handling |
| DA1 end sentinel | D | P: DA1 or 1 s | p: VT answers DA1; not a querying sentinel | D: event passed through, not sentinel | p: typed DA1 decoder | P: DA1 or 1 s | p: query encoder | p: DA1 decoder | P: startup DA1 or 300 ms; late DA1 recovery |
| XTVERSION | D | D | P: VT answers identity | p: explicit request and typed message | p: typed DCS reply | P: startup request and identity event | p: query encoder | p: DCS decoder | P: query before DA1, identity and fallback |
| focus 1004 | D: no focus routing | D | P: VT tracks mode and session mirrors it | P: view requests and resets, typed events | p: focus/blur decoding | P: opt-in set/reset, focus event | p: mode encoders | p: focus decoder | P: view opts in and forwards focus |
| 2031 / OSC 11 | D: no host query path | D | D: 2031 unrecognized | p: OSC 11 request/message | p: OSC 11 and scheme event decoding | D: no corresponding startup policy found | p: encoders | p: OSC 11 decoder | P: local 2031 and OSC 11, theme requery and cleanup |

The statuses do not assert that every unsupported terminal responds, that a parser initiates a query, or that a mode's previous state is preserved. Bubble Tea's `ModeReportMsg` and UV's event are visible to applications, but remain typed reports rather than key events. TUIOS mode-2027 guest support is a separate emulator property, not evidence that its host negotiated 2027; see [width report](report-width.md).

## Synthetic reply matrix

Run `go run .` in `/tmp/claude-1000/-home-uwe-projects-loom/9e4d0913-fa4b-4cb2-aec0-0c51a6bd4a87/scratchpad/b3-canary`. The same bytes are fed to a real Pane.Run in a disposable controlling PTY and UV TerminalReader.StreamEvents over an in-process pipe. Each side/check has a five-second context; no visible windows, sleeps, or real terminal. The first printable `a` is acknowledged before a split continuation is written. The ordinary split waits on a 10 ms timer after that acknowledgement; the late split waits on a 130 ms timer, beyond both 50 ms escape-completion timers. `z` is a harness termination marker and omitted below. All runs returned nil errors. `K(text,report)` denotes a key delivered to the Pane widget, including empty keys.

| Synthetic input | Pane widget events | UV reader events |
|---|---|---|
| `CSI ? 2026 ; N $ y`, N = 0,1,2,3,4 | `K("",true)` for each | `ModeReportEvent{2026,N}` for each |
| `CSI ? 2027 ; N $ y`, N = 0,1,2,3,4 | `K("",true)` for each | `ModeReportEvent{2027,N}` for each |
| silence for 130 ms | none | none |
| `a`, complete 2026/1 reply, `b` | `K("a",false), K("",true), K("b",false)` | key a, mode report 2026/1, key b |
| split `a CSI ? 2026 ;` then `1 $ y` | `K("a",false), K("",true)` | key a, mode report 2026/1 |
| same split, continuation after 130 ms | `K("a",false), K("",false), K("1",false), K("$",false), K("y",false)` | key a, `UnknownEvent("\x1b[?2026;")`, keys 1, $, y |
| `CSI 48;24;80;480;800 t` | `K("",false)` | `WindowSizeEvent{80,24}`, `PixelSizeEvent{800,480}` |
| `CSI ? 1 u` | `K("",false)` | `KeyboardEnhancementsEvent{1}` |
| `CSI ? 62;22 c` | `K("",true)` | `PrimaryDeviceAttributesEvent[62,22]` |

The split writes exercise the actual readers; the acknowledged `a` and 10 ms interval give the first fragment time to enter the reader, but do not instrument the exact instant the remainder entered Pane's pending buffer. The delayed case additionally proves timeout expiration before the second write through the resulting events. No query was issued during the silence case: it establishes reader behavior, not a negotiation timeout. A five-second test bound is distinct from the 50 ms escape timer and from the protocol deadlines below.

## Loom implementation

### V1 Pane input and output

[Scanner and classifier](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/event.go#L689-L725):

```go
	ev := DecodeKey(b[:n])
	if ev == (KeyEvent{}) && isTerminalReply(b[:n]) {
		ev.Report = true
	}
	return ev, n, true
```

[Pane input loop](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/pane_run.go#L688-L708) buffers an incomplete escape for the configured 50 ms and dispatches decoded events; [dispatcher](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/event.go#L800-L821) forwards to the root widget. It does not start a DECRQM transaction, so there is no negotiation deadline or silence fallback. [Canvas](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/canvas.go#L576-L610) wraps output in 2026 without querying; this is output framing, not proof of synchronization. [Kitty](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/pane_modes.go#L143-L165) is opt-in push flag 1, with [pop on close](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/pane_terminal.go#L220-L240); screen transitions also manage it. Focus, color, version and resize negotiation are not implemented in these Pane input/startup paths.

### V2 Probe

[Kitty/DA1 probe](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/cmd/loom-probe/keys.go#L82-L105):

```go
	deadline := time.Now().Add(time.Second)
	support := kittySupport{}
	for !support.answered && time.Now().Before(deadline) {
		chunk, err := readChunk(inFD, time.Until(deadline))
		if err != nil {
			return fmt.Errorf("read kitty query reply: %w", err)
		}
```

It writes kitty query followed by DA1; accumulates the raw chunk, scans for flags and DA1, exits on DA1 or one second. Silence means no DA1, and it skips the kitty phase if flags were not seen. Interleaved user bytes read during probing are held in that probe buffer rather than replayed to the later key phase; bytes arriving after the probe may enter that phase. Push is conditional on a reply and [pop/tty restoration](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/cmd/loom-probe/keys.go#L108-L125) follows; this is a diagnostic, not Pane's negotiation. The other listed modes are outside this keys probe.

### V3 Session and VT

[VT private modes](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/vt/vt.go#L735-L765):

```go
	case 2004:
		on = v.BracketedPaste
	default:
		return 0
	}
	if on {
		return 1
	}
```

The explicit list includes 1004 but not 2026, 2027, 2031 or 2048. VT [answers DECRQM, DA1, XTVERSION and kitty](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/vt/vt.go#L798-L854); kitty pop clears a single stored flag value, not a nested/per-screen stack. [2026 reset captures frames](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/vt/vt.go#L894-L911), although DECRQM replies 0. Session [drops classified reports](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/internal/session/session.go#L1080-L1099) before prefix/search and [mirrors focus](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/internal/session/session.go#L473-L482). It is a responder/filter, not a host query transaction with a deadline; no restoration claim follows from guest state alone.

## Upstream mechanisms

### V4 Bubble Tea

[Response policy](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/tea.go#L814-L828):

```go
			case ModeReportMsg:
				switch msg.Mode {
				case ansi.ModeSynchronizedOutput:
					if msg.Value == ansi.ModeReset {
						// The terminal supports synchronized output and it's
						// currently disabled, so we can enable it on the renderer.
						p.renderer.setSyncdUpdates(true)
```

Startup [conditionally queries 2026 and 2027](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/tea.go#L1141-L1148); 2027 accepts 1, 2, 3 and selects grapheme width. There is no blocking query deadline in that branch: silence leaves the initial renderer policy. [Input translation](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/input.go#L8-L53) maps typed reports, focus, size, color, version and kitty messages; unlisted UV events pass through. The renderer [sets/resets focus and pushes/pops kitty](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/cursed_renderer.go#L143-L170), and [queries kitty after requesting enhancements](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/cursed_renderer.go#L418-L448). [Close resets 2027](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/cursed_renderer.go#L250-L260), which is not restoration of a previously set state. [OSC 11](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/color.go#L12-L15) and [XTVERSION](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/xterm.go#L17-L23) are explicit commands, not a DA1 sentinel.

### V5 ultraviolet

[2027 policy](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal.go#L320-L336):

```go
func (t *Terminal) handleEvent(ev Event) {
	if mr, ok := ev.(ModeReportEvent); ok {
		if mr.Mode == ansi.ModeUnicodeCore &&
			(mr.Value.IsSet() || mr.Value == ansi.ModeReset) {
			t.scr.enableGraphemeWidth()
		}
	}
```

Start [requests 2027](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal.go#L305-L315); the reply is also forwarded to the application. [Screen override and enable](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_screen.go#L160-L190) write 2027h unless overridden; silence leaves the initial width method, no blocking response deadline. [Reader](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_reader.go#L139-L225) uses a 50 ms escape-completion timeout, not a negotiation deadline. [Decoder](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/decoder.go#L380-L402) types DECRPM, DA1 and kitty; [resize](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/decoder.go#L660-L680), [focus/scheme](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/decoder.go#L426-L446), [OSC 11](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/decoder.go#L830-L842) and [XTVERSION](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/decoder.go#L1050-L1055) are decoded separately. These decoder cases do not themselves send queries or restore modes; do not infer prior-mode restoration from accepting 2027/1.

### V6 tcell

[Startup sentinel](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/tscreen.go#L517-L536):

```go
	expire := time.After(time.Second)

	for {
		select {
		case <-expire:
			t.initted = true
			return
		case ev := <-t.initQ:
```

[Query batch](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/tscreen.go#L1647-L1679) requests 2048, kitty and XTVERSION, then DA1 last; on DA1 or one second it continues with what it learned. [Filter](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/tscreen.go#L573-L590) routes private replies to initQ rather than key events; [mode policy](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/tscreen.go#L545-L565) uses changeable 2048 status. [CSI 48](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/input.go#L1502-L1515) posts resize with pixel dimensions. It [enables 2048](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/tscreen.go#L1717-L1726) and [disables on disengage](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/tscreen.go#L1745-L1753). Kitty uses `CSI =1u` or `=31u`, not `CSI >...u` push; it resets with `=0u`. [Disengage](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/tscreen.go#L1779-L1804) also disables focus and keyboard modes. This resets rather than proves restoration of previous values. The audited startup query batch does not negotiate 2026, 2027 or 2031/OSC 11.

### V7 x/ansi

[Mode constants](https://github.com/charmbracelet/x/blob/faa4adf95555331dfb439b716cc83767ecdadeb4/ansi/mode.go#L620-L637):

```go
	ModeSynchronizedOutput = DECMode(2026)

	SetModeSynchronizedOutput     = "\x1b[?2026h"
	ResetModeSynchronizedOutput   = "\x1b[?2026l"
	RequestModeSynchronizedOutput = "\x1b[?2026$p"
```

The encoder also supplies [1004](https://github.com/charmbracelet/x/blob/faa4adf95555331dfb439b716cc83767ecdadeb4/ansi/mode.go#L501-L505), [2031 and 2048](https://github.com/charmbracelet/x/blob/faa4adf95555331dfb439b716cc83767ecdadeb4/ansi/mode.go#L642-L677), [CSI 48](https://github.com/charmbracelet/x/blob/faa4adf95555331dfb439b716cc83767ecdadeb4/ansi/inband.go#L5-L12), [kitty query/push/pop](https://github.com/charmbracelet/x/blob/faa4adf95555331dfb439b716cc83767ecdadeb4/ansi/kitty.go#L18-L83), [DA1 and XTVERSION](https://github.com/charmbracelet/x/blob/faa4adf95555331dfb439b716cc83767ecdadeb4/ansi/ctrl.go#L8-L86) and [OSC 11](https://github.com/charmbracelet/x/blob/faa4adf95555331dfb439b716cc83767ecdadeb4/ansi/background.go#L143-L147). It does not own a reader, deadline, silence policy or lifecycle; callers must implement these.

### V8 x/input

[Reply parsing](https://github.com/charmbracelet/x/blob/15a1484e6848922233d6180a57d4734cd59573ee/input/parse.go#L274-L287):

```go
		return i, ModeReportEvent{Mode: ansi.DECMode(mode), Value: ansi.ModeSetting(value)}
	case 'c' | '?'<<parser.PrefixShift:
		// Primary Device Attributes
		return i, parsePrimaryDevAttrs(pa)
	case 'u' | '?'<<parser.PrefixShift:
		// Kitty keyboard flags
```

This is a parser/driver, not a mode negotiator. [CSI window operations](https://github.com/charmbracelet/x/blob/15a1484e6848922233d6180a57d4734cd59573ee/input/parse.go#L519-L537) return a generic operation and arguments, not UV's two typed 2048 events. [Focus](https://github.com/charmbracelet/x/blob/15a1484e6848922233d6180a57d4734cd59573ee/input/parse.go#L314-L320), [OSC 11](https://github.com/charmbracelet/x/blob/15a1484e6848922233d6180a57d4734cd59573ee/input/parse.go#L677-L685) and [XTVERSION](https://github.com/charmbracelet/x/blob/15a1484e6848922233d6180a57d4734cd59573ee/input/parse.go#L883-L887) produce distinct event types. Caller owns query deadline, silence and restoration; parser support does not prove buffered-reader behavior.

### V9 TUIOS

[Startup deadline](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/app/capabilities.go#L314-L322):

```go
// probeTimeout is the backstop for the capability probe. It is only ever spent
// in full by a host that answers nothing at all, since the probe stops on the
// DA1 reply rather than on the clock.
const probeTimeout = 300 * time.Millisecond
```

[Probe](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/app/capabilities.go#L336-L395) enters raw mode, writes geometry, graphics, palette, XTVERSION then DA1, reads until DA1 or 300 ms, and restores tty attributes. It consumes the raw probe buffer; this path does not demonstrate preservation of interleaved typed input. Silence uses default cell geometry and records late DA1; [late reply handler](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/app/sixel_probe.go#L18-L56) recovers a late DA1 from the running event stream. [XTVERSION parsing](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/app/kitty_placeholder_caps.go#L44-L70) uses identity as a heuristic, not proof of graphics capabilities.

[Color](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/app/host_colors.go#L230-L277) enables 2031 only for a local client and asks OSC 11; [typed scheme events requery colors](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/app/host_colors.go#L289-L317). There is no blocking color query deadline: missing fields retain fallback, and answers are batched for 15 ms. [Terminal cleanup](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/terminal/cleanup.go#L15-L25) resets 2031 and 1004. [Focus view](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/app/render.go#L710-L716), [focus handling](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/app/update.go#L2283-L2302), [keyboard reply](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/app/update.go#L2327-L2336), [guest 2026 hold](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/terminal/window_geometry.go#L359-L367), and [cell-size sources](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/cellsize/cellsize.go#L1-L12) show the other application boundaries. The local reset is not evidence that every remote/suspend path restores a prior mode.

### Additional feature excerpts

These excerpts connect the rows whose important distinction is not visible in the mode-report snippets above. For 2048, [tcell's report parser](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/input.go#L1504-L1512) posts an actual resize; UV instead emits the two typed events shown in the matrix, while x/input emits a generic window operation:

```go
				ws := WindowSize{Width: P[2], Height: P[1]}
				if len(P) > 4 {
					ws.PixelWidth, ws.PixelHeight = P[4], P[3]
				}
				ip.post(newEventResize(ws))
```

For kitty, [Bubble Tea](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/cursed_renderer.go#L431-L447) pushes before its query, so the answer describes the resulting current flags; [Loom's VT](https://codeberg.org/ubunatic/loom/src/commit/777d170c4b8f99ae2cd48bb36ddd91dd59572b7f/vt/vt.go#L800-L817) answers from a single flag slot. This differs from the probe and tcell query-first policies:

```go
			_, _ = s.scr.WriteString(ansi.PushKittyKeyboard(kittyFlags))
```

For focus and 2031, [UV's decoder](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/decoder.go#L428-L443) makes distinct events instead of keys; these decoding cases alone do not enable either mode. The x/input focus cases and OSC 11 parser, and TUIOS's actual enable/requery/cleanup calls, are linked in V8 and V9.

```go
		case 997: // [ansi.LightDarkReport]
			switch darkLight {
			case 1:
				return i, DarkColorSchemeEvent{}
			case 2:
				return i, LightColorSchemeEvent{}
			}
```

For DA1 and XTVERSION, V2, V3, V6 and V9 identify actual request or response paths; V4, V5, V7 and V8 identify encoders or typed events without a sentinel. TUIOS and tcell bound startup by a timer, while Pane has no DA1 startup transaction. No reply is evidence of a deadline only when the caller actually starts such a transaction.

## Terminal support and owner check

Owner check run 2026-11-10 with `manual-queries.py` (Linux, no multiplexer). DECRQM values:
0 not recognised, 1 set, 2 reset. Raw buffers held only the replies; the owner was typing while
the kitty window opened, and no typed bytes appeared in any reply buffer.

| Query | foot 1.27.0 | kitty 0.47.1 |
|---|---|---|
| DECRQM 2026 | 2 (reset) | 2 (reset) |
| DECRQM 2027 | 1 (set) | **0** (not recognised) |
| DECRQM 2048 | 2 (reset) | 2 (reset) |
| DECRQM 1004 | 2 (reset) | 2 (reset) |
| DECRQM 2031 | 2 (reset) | 2 (reset) |
| kitty `CSI ? u` | `CSI ? 0 u` | `CSI ? 0 u` |
| XTVERSION | `foot(1.27.0)` | `kitty(0.47.1)` |
| DA1 | `CSI ? 62;4;22;28;52 c` | `CSI ? 62;52; c` |
| reply time per query | 0.08–0.16 ms | 3.4–5.1 ms |

Both terminals answer every query, and DA1 last, within 6 ms locally; a DA1-ended batch with a
short backstop costs little. Remote (SSH) latency is not measured. Ghostty is not installed.

The spec pin is foot 1.28.0; it is not the installed foot 1.27.0 and documentation is not a live reply. Run the owner-only `/tmp/claude-1000/-home-uwe-projects-loom/9e4d0913-fa4b-4cb2-aec0-0c51a6bd4a87/scratchpad/manual-queries.py` with `foot` or `kitty` as argv[1] inside that terminal. It sends only DECRQM 2026/2027/2048/1004/2031, kitty flags, XTVERSION and DA1 last; raw mode is restored in finally, two seconds per reply, bytes and elapsed milliseconds saved beside the script. A late reply can be included in a later query's raw record; inspect the complete transcript. Syntax checked, deliberately not run.

## Stage 4 proposals

* In `features.md`, `summary.md` and corresponding `features-*.yaml` rows, split unconditional 2026 frame wrapping from negotiated support; distinguish VT snapshot frames and a DECRQM 0 answer from physical synchronized output. Cite V1, V3 and V4.
* Record 2027 as no Pane negotiation, probe as unrelated and VT as 0; Bubble Tea/UV accept 1–3 with explicit width overrides as applicable. The mode policy, probing and restoration overlap ticket 461; avoid a duplicate 2027 ticket. Width equivalence remains separate.
* Record 2048 separately for query, enable, resize decoding (including pixel dimensions), fallback and exit reset. tcell proves a complete host path; Pane's synthetic report is an empty key. Propose a candidate for bounded in-band resize support only if ordered cell/pixel resize improves Loom; it does not overlap 461.
* Candidate: typed reply routing and buffering shared by host negotiations, preserving user input and testing late/split responses, with lifecycle cleanup. V1, V2 and the matrix demonstrate why; the 2027-specific query portion overlaps 461, while general routing, kitty and other replies do not. The late-fragment ambiguity requires a policy, not an assumption that a timeout identifies a reply.
* Record probe DA1 as bounded with consumed interleaved input; kitty current flags are not capability bits, and VT's single kitty slot is not a stack. Candidate: faithful guest kitty nested/per-screen stack, separate from 461. Record focus and color as distinct features and identify TUIOS's local-only 2031 cleanup.
* Ticket 459 concerns rendering diffs; do not duplicate it. API stability is P3 while Loom is alpha. These are integration proposals, not filed tickets or adopted implementation.

For the eventual YAML integration, update `features-terminals.yaml` rows `term.sync-output-2026`, `term.decrqm-2026-replies`, `term.inband-resize-2048`, `term.kitty-keyboard.flags` and `term.xtversion` with the immutable spec pins and the 2048 detection/report details. Update `features-bubbletea.yaml` rows `bt.sync-output-2026`, `bt.kitty.enable-on-start`, `bt.kitty.keyboard-enhancements-msg`, `bt.focus.msgs` and `bt.theme.background-color` with the policies in V4. Update `features-charmx.yaml` rows `uv.sync-output-2026` and `ansi.kitty-keyboard-flags`, and add separately identified UV reader and mode-2027 rows. Revisit `features-tuios.yaml` rows `tuios.width-mode-2027`, `tuios.sync-2026-scope` and `tuios.kitty-keyboard-stack` at the new pin, keeping host and guest scopes separate. Add rows for DA1 deadlines, 2048 reports, 1004 and OSC 11/2031 where missing; do not replace an older-dependency finding with a latest-dependency claim.

## Flow notes

The real-loop matrix, per-feature source comparison and three requested protocol source families were completed. The immediate split tests a separately written continuation but does not instrument Pane's internal read boundary; late reply and interleaved-key effects are observed at the widget and UV event channels. The agent queried no real terminal; the host ran the owner check afterwards (table above). No tracked implementation, ticket or commit was created.
