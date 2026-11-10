# Mouse and drag follow-up

At Loom `7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c`, the headless Pane canary demonstrates correct nested coordinate translation and modal ownership, but TextEntry selection does not capture the pointer and losing focus does not end its active pan. Neither terminal delivery outside a physical window nor comparable runtime behavior in the other projects was tested. This is a research report, not a recommendation to make the SDK modal.

## Pins and scope

Resolved 2026-11-10 against latest releases or, where there is no release, latest repository commit: Loom `7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c` (origin); TUIOS v0.12.1 `850335c0c08bb8b802406fcc7f254355c89af04a`; Bubble Tea v2.1.1 `67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7`; ultraviolet `a45b7c86333f9dfedb95f81064e80259b02937f0` (HEAD); tview v0.42.0 `5ce6a2b588145610060000a4f75d7e2af081a794`; tcell v3.6.0 `4bed18057fce88934c9b9a576b5f83632196de11`; gocui `c81715e95462cf95447bbd619e4dc85f3e4e1006` (HEAD). Source snapshots were fetched by SHA. TUIOS's dependency versions are not substituted for the independent latest pins.

## Capability ledger

`P` means proved in the stated source or canary scope; `D` means the specific behavior was disproved in that scope; `~` means a relevant mechanism exists but the complete feature, all widgets, or terminal delivery is not established. These are not comparable benchmark results: Loom has a runtime probe, other columns are source traces. BT, UV and tcell are lower-level event/terminal libraries; their `~` means they supply event building blocks, not that they promise a widget behavior. TUIOS has application-specific behavior rather than an SDK contract. `U` would mean a blocked source or failed canary; no source pin below was blocked.

| Feature and scope | Loom | TUIOS | BT v2 | UV | tview | tcell v3 | gocui |
|---|---|---|---|---|---|---|---|
| SGR and modifiers, including pixel 1016 | ~ | ~ | ~ | P | ~ | P | ~ |
| 1002 held motion versus 1003 all motion | P | ~ | P | P | ~ | P | ~ |
| Drag ownership following press | ~ | ~ | ~ | ~ | ~ | ~ | ~ |
| Nested child-local, 0-based translation | P | ~ | ~ | ~ | ~ | ~ | ~ |
| Drag beyond widget and OS window | ~ | ~ | ~ | ~ | ~ | ~ | ~ |
| Entry/area selection outside bounds | D | ~ | ~ | ~ | D | ~ | ~ |
| Release and Esc during drag | ~ | ~ | ~ | ~ | ~ | ~ | ~ |
| Focus loss and re-entry during drag | D | ~ | ~ | ~ | ~ | ~ | ~ |
| Modal capture | P | ~ | ~ | ~ | ~ | ~ | ~ |
| Stationary outside selection autoscroll | D | P | ~ | ~ | ~ | ~ | ~ |
| Wheel bursts and curve | P | ~ | ~ | ~ | ~ | ~ | ~ |
| Horizontal wheel and axis lock | D | P | ~ | P | ~ | P | ~ |
| Double and triple click | P | P | ~ | ~ | ~ | ~ | ~ |

Loom's `D` for selection is specifically TextEntry's outside drag and TextArea's mouse handling in the adapter described below, not a claim about all editors. Stationary `D` is for that selection path over 300 ms, supported by the ticker/capture source; it does not rule out unrelated ticker widgets. Loom `P` for 1002/1003 is mode emission plus injected decoding, not terminal-originated behavior. For other columns partial is an evidence-bound result, not a claim that an unexamined widget implements the complete feature. P for UV/tcell pixel describes source execution/API, not a successful pixel-mode negotiation with a particular terminal. Wheel curve P is scoped to Loom’s configured vertical acceleration; TUIOS axis P is scoped to modifier scrolling in its scrolling layout. Double/triple P is scoped to Loom TextEntry and TUIOS copy mode. tview D is scoped to its TextArea handler outside its rectangle, as independently probed below.

## Headless Pane harness

The disposable `scratchpad/b5-canary` has its own module and `replace` to the checkout. `go run .` runs two independent child processes on 80×24 PTYs, with a five-second deadline each, no visible windows, and a 300 ms timer solely to measure the stationary interval. The child writes observations to a separate pipe; an `s` marker synchronizes each step. There are no arbitrary sleeps. A vertical Stack puts the lower group at `(0,12)`; a horizontal Stack puts a nested horizontal Stack at `(40,12)`, with the wrapped real TextEntry `(40,12,20,12)` and recording sibling `(60,12,20,12)`. A TextArea adapter occupies `(0,12,40,12)`: TextArea's `Draw` takes an extra focus argument and it has no `ConsumeMouse`, so the adapter records and ignores mouse events. It does not implement selection behavior on TextArea's behalf. A modal target is at `(55,14,12,5)`.

The PTY bytes are synthetic, so sending a coordinate outside 80×24 proves only what happens **if** that report is delivered. Both requested modes emitted the requested `CSI ? 1002 h` or `CSI ? 1003 h` and `CSI ? 1006 h`; neither emitted the other tracking mode. Each received the same injected no-button hover. Terminal enforcement of the mode is outside the PTY fixture. `action=0/1/2/3/5` means press/release/hover/drag/wheel-down. Coordinates below are widget-local and 0-based; `–` means no widget received the injected event.

| Input bytes (`ESC` = `\e`) and meaning | Recipient / actual local event |
|---|---|
| `\e[<35;43;15M`, no-button motion | entry hover `(2,2)` in both modes |
| `\e[<0;43;15M`, left press | entry press `(2,2)`; anchor=caret=699, scroll=39 |
| `\e[<32;48;16M`, held drag inside | entry drag `(7,3)`; anchor=699, caret=721, scroll=39 |
| `\e[<32;65;16M`, held drag into sibling | sibling drag `(4,3)`; entry selection remains 699:721 |
| `\e[<32;90;30M`, held drag outside PTY size | –; entry selection 699:721, scroll=39 |
| stationary 300 ms, no input | zero recorded widget events; source inspection finds no scheduled ticker in this tree; entry selection 699:721, scroll=39 |
| `\e[<0;90;30m`, release outside | –; entry `dragging=true` remains |
| new entry press `\e[<0;46;15M`, then activate modal; drag `\e[<32;58;16M` | entry press `(5,2)` then modal drag `(2,1)`; modal takes precedence even though the drag began before it appeared |
| modal press `\e[<0;58;16M` | modal press `(2,1)` |
| modal held drag `\e[<32;70;17M` | modal drag `(14,2)` across sibling |
| modal held drag `\e[<32;90;30M`, release `\e[<0;90;30m` | modal drag and release `(34,15)` |
| area press `\e[<0;5;15M`, drag `\e[<32;70;20M`, outside release | area adapter press `(4,2)`, sibling drag `(9,7)`, no recipient for release; area selection `0:0`, caret `(0,0)`, scroll=0 |
| 12 consecutive `\e[<65;45;15M` wheel-down reports | entry receives 24 wheel-down events `(4,2)`; scroll remains 39 at its lower limit |
| horizontal wheel reports `\e[<66;45;15M\e[<67;45;15M` | entry receives two **presses**, buttons 2 and 3, `(4,2)`; these are not horizontal wheel events |
| middle press `\e[<1;44;20M`; drag `\e[<33;44;15M` | entry wrapper receives press `(3,7)`; private capture routes drag directly to the underlying entry, bypassing wrapper logging; scroll 39→7 |
| focus-off key; held drag `\e[<33;90;2M` | `SetFocus(false)` leaves `pan.active=true`; captured drag changes scroll 7→0 while unfocused |
| focus-on key; middle release `\e[<1;90;30m` | focus restored while pan remains active; release goes directly to entry and clears `pan.active` |
| middle press/drag, then `\e[27u` Esc | scroll 0→39 with pan active; Esc restores scroll 0 and ends pan; subsequent drag goes to sibling `(9,4)` |
| three consecutive left press/release pairs at `(50,18)` | entry receives each press/release at `(9,5)`; selection after each is `94:94`, `91:95`, `85:101` respectively |

Selection values are inspected via reflection because no public selection-offset/scroll accessors exist for TextEntry; TextArea's public `Caret` and `Scroll` and reflected selection endpoints are recorded. The parent records all wrapper-delivered events, but capture deliberately bypasses the wrapper, so state changes, not missing wrapper log lines, demonstrate delivery to the real entry. Earlier root-only testing additionally injected `\e[<28;10;4M` and observed Shift, Alt and Ctrl, and one root press/drag/release outside, confirming the decoder boundary only. The multi-click sequence exercises the entry’s word and line selections; its time threshold is established by source, not a boundary measurement.

## Loom implementation

SGR is decoded and terminal coordinates converted once in [`event.go`](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/event.go#L509-L560) and [`pane_run_mouse.go`](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/pane_run_mouse.go#L10-L24); nested Stack subtracts the child's allocation in [`stack.go`](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/stack.go#L160-L176). SGR 1016 pixel negotiation/conversion is not implemented by this path. `PaneRequest.Mouse` permits either tracking mode; `EnableMouse` itself requests 1003 ([modes](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/pane_modes.go#L10-L17)).

V — the private captured route translates the stored origin, while ordinary child hit testing still applies when no widget claims capture ([source](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/mouse_capture.go#L48-L60)):

```go
case MouseDrag, MouseRelease:
    e.X -= origin.X
    e.Y -= origin.Y
    result = target.ConsumeMouse(e)
```

V — focus loss cannot reach its cancellation call because the assignment precedes the test ([source](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/text_entry.go#L338-L343)):

```go
func (e *TextEntry) SetFocus(focused bool) {
    e.focused = focused
    if !focused {
        if e.focused {
            e.endMouseCapture(true)
```

`mouseCapturer` is private, and TextEntry claims capture only for `barDrag.active || pan.active`, not its selection `dragging` ([entry](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/text_entry.go#L1237-L1245)). Captured Esc cancels and restores the pre-press offset ([dispatch](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/pane_dispatch.go#L14-L19), [rollback](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/text_entry.go#L1247-L1257)). The [`SetFocus`](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/text_entry.go#L338-L349) false branch tests `e.focused` after setting it false; the runtime result confirms that pan remains active. This is a focus-transition bug, not a claim that terminal focus reports were simulated.

[`ModalMouseCapture`](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/modal_mouse.go#L46-L95) retains press target and origin, including outside the modal, and [`dispatchMouse`](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/pane_dispatch.go#L33-L46) checks modal before normal capture. [`tickTree`](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/widget.go#L74-L119) runs explicit Ticker implementations; the tested selection path supplies none. Wheel repeat ramps each same-direction report after three using a 250 ms burst gap, step three, maximum ten ([implementation](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/pane_wheel.go#L17-L47), [defaults](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/spec/defaults.yaml#L72-L82)); SGR horizontal wheel codes 66/67 fall through to ordinary press decoding ([decoder](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/event.go#L525-L556)); the runtime result confirms this and no horizontal-axis arbitration is possible on those events. TextEntry detects double word and triple line selection ([source](https://codeberg.org/ubunatic/loom/src/commit/7ecfba5e89c4d6c2c3b0a34bf18b15164b2b844c/text_entry.go#L601-L617)); other widgets need separate evidence.

## Upstream mechanisms

TUIOS starts an autoscroll timer only when copy-mode drag motion crosses content bounds, sets `AutoScrollActive`, and uses 50 ms ticks ([starter](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/input/mouse_motion.go#L280-L301)). Its update handler checks active/direction, moves two steps, updates visual selection and schedules the next tick ([handler](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/app/update.go#L1261-L1314)); returning inside or releasing stops the state ([motion](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/input/mouse_motion.go#L295-L301), [release](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/input/mouse_release.go#L220-L231)). The first tick requires an actual outside motion report, not merely an imagined pointer position. V, the continuation guard:

```go
case AutoScrollTickMsg:
    if !m.AutoScrollActive || m.AutoScrollDir == 0 {
        return m, nil
    }
```

V — the initiating motion starts the timer, and a return inside turns off its state ([source](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/input/mouse_motion.go#L290-L301)):

```go
if scrollDir != 0 && !o.AutoScrollActive {
    o.AutoScrollActive = true
    return o, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
        return app.AutoScrollTickMsg{}
    })
}
```

TUIOS copy-mode selection explicitly maps one, two and three clicks to character, word and line ([selection](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/input/mouse_select.go#L43-L90)). TUIOS tracks a majority axis for a 250 ms gesture and admits a new axis when its count wins ([axis](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/app/wheel_axis.go#L43-L75)); its caller uses the gate specifically for modifier scrolling in the scrolling layout ([caller](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/input/mouse_wheel.go#L125-L145)). This is a scoped axis policy, not a general claim about every TUIOS wheel consumer.

V — release cancels the recurring work through its active flag ([source](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/internal/input/mouse_release.go#L226-L230)):

```go
o.Dragging = false
o.DraggedWindowIndex = -1
o.InteractionMode = false
o.AutoScrollActive = false
o.AutoScrollDir = 0
```

Bubble Tea v2 exposes zero-based terminal coordinates, modifiers, and distinct click/release/wheel/motion messages in [`mouse.go`](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/mouse.go#L52-L77). Current view [`MouseMode`](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/tea.go#L297-L319), rather than old v1 program options, chooses cell/held motion versus all motion. [`options.go`](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/options.go#L27-L48) supplies input/output configuration; it does not establish widget capture. V — the held-motion mode requires a pressed button; all-motion does not ([source](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/tea.go#L304-L315)):

```go
// Mouse movement events are also captured if a mouse button is pressed
// (i.e., drag events). Cell motion mode is better supported than all
// motion mode.
```

UV names 1002, 1003 and 1016 and supplies pixel-to-cell conversion ([modes and encoding](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/mouse.go#L8-L32), [conversion](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/mouse.go#L122-L145)). tcell's flags define button, drag, motion and pixel behavior ([screen](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/screen.go#L305-L320)); its actual [`enableMouse`](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/tscreen.go#L1148-L1178) chooses SGR versus pixel mode. [`mouse.go`](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/mouse.go#L45-L80) defines position units and modifiers.

V — tview's global interceptor can replace or suppress an event; it is called before target selection ([invocation](https://github.com/rivo/tview/blob/5ce6a2b588145610060000a4f75d7e2af081a794/application.go#L535-L544)):

```go
if a.mouseCapture != nil {
    event, action = a.mouseCapture(event, action)
    if event == nil {
        consumed = true
```

The independent `mouseCapturingPrimitive` is selected and replaced by the handler's returned primitive ([ownership](https://github.com/rivo/tview/blob/5ce6a2b588145610060000a4f75d7e2af081a794/application.go#L547-L567)); `SetMouseCapture` installs the interceptor, not that owner ([setter](https://github.com/rivo/tview/blob/5ce6a2b588145610060000a4f75d7e2af081a794/application.go#L175-L184)). tview also derives double clicks ([application](https://github.com/rivo/tview/blob/5ce6a2b588145610060000a4f75d7e2af081a794/application.go#L575-L610)); this is not evidence of triple-click support. Its [TextArea handler](https://github.com/rivo/tview/blob/5ce6a2b588145610060000a4f75d7e2af081a794/textarea.go#L2340-L2400) checks bounds before processing any action and returns capture on down but not on motion. A separate direct-handler canary using tview v0.42.0 and its own tcell v2 dependency, simulation screen and rectangle `(40,12,20,6)`, recorded:

| tview TextArea action and global coordinate | consumed; returned capture; selection; offset |
|---|---|
| down `(42,13)` | true; yes; `21:21`; `0,0` |
| move inside `(45,13)` | true; no; `21:24`; `0,0` |
| move outside `(65,13)` | false; no; `21:24`; `0,0` |
| release outside `(65,13)` | false; no; `21:24`; `0,0` |

This probes the real handler directly, not the tview application loop. Combined with the application assignment above, the returned nil would replace its current owner on the inside move. The outside handler call independently confirms no outside selection update or release cleanup for this primitive.

Correction to the earlier overly cautious gocui claim: mouse buttons really **do share numeric Key values** with F56–F63 ([constants](https://github.com/jesseduffield/gocui/blob/c81715e95462cf95447bbd619e4dc85f3e4e1006/keybinding.go#L274-L286)). Separate event-type branches do not guarantee separation: mouse dispatch can fall through to the same `execKeybindings` as keyboard dispatch ([branches](https://github.com/jesseduffield/gocui/blob/c81715e95462cf95447bbd619e4dc85f3e4e1006/gui.go#L1293-L1320), [mouse fallback](https://github.com/jesseduffield/gocui/blob/c81715e95462cf95447bbd619e4dc85f3e4e1006/gui.go#L1390-L1408)); the matcher compares key, rune and modifiers, not type ([matcher](https://github.com/jesseduffield/gocui/blob/c81715e95462cf95447bbd619e4dc85f3e4e1006/keybinding.go#L121-L123)). A matching key/rune/modifier could therefore collide if mouse-specific handling does not consume it; no end-to-end collision canary was run. gocui records double-click context ([source](https://github.com/jesseduffield/gocui/blob/c81715e95462cf95447bbd619e4dc85f3e4e1006/gui.go#L1429-L1452)).

## Owner check

The unrun `scratchpad/manual-mouse.py` opens `/dev/tty` raw, enables 1006 with 1002 then 1003, records raw hex and timestamps to `manual-mouse-<terminal>.txt`, bounds each step at 30 seconds, and disables modes/restores termios in `finally`. `python3 -I -m py_compile` passed. In foot 1.27.0 and kitty 0.47.1 separately, run the script with `foot` or `kitty` argument; for each mode move without a button, then press and hold, drag outside the OS window, release outside, and re-enter; press q to finish each step. Compare the raw reports and note any OS/window-manager dependence. Do not infer that an absent release reached Loom.

## Stage 4 proposals

For `features.md`, split global interception from gesture ownership, public capture from built-in private capture, and selection capture from scrollbar/pan capture. Split terminal mode request/decoding/pixel units from actual physical delivery; track modal precedence and child-local translation separately. Add stationary selection autoscroll, wheel burst curve, axis arbitration, and double versus triple-click rows with the narrow scopes above. For `features-*.yaml`, create equivalent per-feature raw evidence rows with immutable source lines and reproduce the specific PTY observations; do not label unrun upstream runtime tests as passed. In `summary.md`, qualify mouse capture as implemented for certain built-ins and modal gestures, explicitly excluding TextEntry selection; replace any equivalence of tview interceptor and ownership and correct the gocui alias discussion. API stability belongs at P3 or below for this alpha SDK.

Specific existing raw rows to amend, retaining their IDs and histories:

- `features-contrast.yaml`: `tview.application.mouse-capture` must describe interception rather than equate it with Loom ownership; add the TextArea handler limitation to a distinct selection row. Keep `gocui.keys-and-mouse-share-keycodes` as a numeric alias, qualify runtime collision as an inference, and pin its source. Pin the `tcell.mouse.drag-motion`, `tcell.screen.mouse-motion-flag` and `tcell.mouse.wheel-button` execution evidence.
- `features-tuios.yaml`: `tuios.drag-autoscroll` should link the initiating motion, recurring update and release cancellation; qualify the initial outside report and copy-mode scope.
- `features-bubbletea.yaml`: `bt.view.mouse-mode` must not say all-motion is required for held drags: cell-motion requests held drag reports. Keep `bt.mouse-mode.v1` as a P3 historical note, not a latest-version comparison.
- `features-terminals.yaml`: preserve the SGR-pixels row but distinguish pixel mode negotiation, unit conversion and actual terminal output. Add separate candidate raw rows for Loom selection ownership, focus cancellation, horizontal wheel decoding, wheel burst, modal capture and entry multi-click; reference the observations and immutable lines above.

Candidate work, not filed: (1) selection gesture ownership and outside release/selection cleanup, plus opt-in stationary autoscroll; overlaps closed #449 (pane capture), #206 (outside-app capture), and #448 (TextEntry scrollbar), but the reproduced selection path is distinct; (2) fix `SetFocus(false)` cancellation/rollback and verify focus re-entry, overlapping #449 and #452 (pan); (3) evaluate public capture for custom widgets, overlapping #449; (4) decode horizontal wheel reports as wheel events before designing axis-aware policy, overlapping closed #451 (burst curve); (5) add TextArea mouse selection only if its intended adapter contract includes it; (6) 1016 conversion as a separately prioritized capability. Coordinate issues #059/#276 and modal #304 overlap translation and modal regression coverage; open #392 concerns session mouse-mode selection. `harnez find issues "drag autoscroll mouse capture selection"` returned no conjunctive matches, so separate drag/mouse searches were also inspected.

## Flow notes

The shared protocol asks for the same matrix in every compared library. This report did **not** run such a cross-library harness: upstream assessments above are source evidence except the explicitly identified direct tview handler canary, and a Loom-only observation is not a comparative measurement. The PTY does not emulate terminal mouse modes, physical pointer capture, or focus-out reports. The TextArea adapter confirms the absence of mouse support in that composition, not a new adapter implementation. The owner script is syntax-checked but intentionally unrun. Follow-up comparison should exercise each upstream application's actual event loop and terminal boundary with identical inputs, and preserve these limits in promotion to Stage 4.

Host decision (2026-11-10): accepted with the limits above (the agent stopped once after 4 minutes
and finished on one resume). The host confirmed the `SetFocus` order and the 66/67 fall-through in
source. Filed: 467 (selection capture, release outside, focus loss, auto-scroll) and 468
(horizontal wheel). Not ticketed: public capture for custom widgets, TextArea mouse selection
(TextArea has no `ConsumeMouse`), SGR 1016 pixel mode, wheel axis lock (after 468). The owner
script was not run.
