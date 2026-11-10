# Batch 2: width and graphemes

Research date: 2026-11-10. Loom source lock: [`d91ffced6f884540d74f2e56ffef299b04ad0f39`](https://codeberg.org/ubunatic/loom/src/commit/d91ffced6f884540d74f2e56ffef299b04ad0f39). This is evidence for Stage 4, not an implementation or a claim about physical terminals. V means a direct pinned source excerpt plus the stated bounded observation; a canary result is evidence only for the exact inputs and versions tested. `unknown` includes unexamined and out-of-scope capabilities. P3 API stability does not justify retaining incorrect segmentation.

## Pins and reproducibility

Latest releases or Go module `@latest` were re-resolved on the research date; TUIOS release and current source were checked separately. Commit links below are immutable. A pseudo-version is the latest published module version, not a release tag.

| Project | Resolved pin |
|---|---|
| Loom | `d91ffced6f884540d74f2e56ffef299b04ad0f39` |
| [charmbracelet/x/ansi](https://github.com/charmbracelet/x/tree/faa4adf95555331dfb439b716cc83767ecdadeb4/ansi) | v0.11.9, `faa4adf95555331dfb439b716cc83767ecdadeb4` |
| [charmbracelet/x/vt](https://github.com/charmbracelet/x/tree/43427fdbad12fde220b737fc36303291a5ce4602/vt) | v0.1.0, `43427fdbad12fde220b737fc36303291a5ce4602` |
| [ultraviolet](https://github.com/charmbracelet/ultraviolet/tree/a45b7c86333f9dfedb95f81064e80259b02937f0) | v0.0.0-20261023060817-a45b7c86333f, `a45b7c86333f9dfedb95f81064e80259b02937f0` |
| [tcell/v3](https://github.com/gdamore/tcell/tree/4bed18057fce88934c9b9a576b5f83632196de11) | v3.6.0, `4bed18057fce88934c9b9a576b5f83632196de11` |
| [uniseg](https://github.com/rivo/uniseg/tree/03509a98a092b522b2ff0de13e53513d18b3b837) | v0.4.7, `03509a98a092b522b2ff0de13e53513d18b3b837` |
| [TUIOS](https://github.com/Gaurav-Gosain/tuios/tree/850335c0c08bb8b802406fcc7f254355c89af04a) | latest release v0.12.1 `850335c0c08bb8b802406fcc7f254355c89af04a`; checked HEAD `6fe2c0df9922b2eec28f392fee4e196efd86be36` also |
| [foot](https://codeberg.org/dnkl/foot/src/commit/ab33c9a19d626f8d0ac0bd1adfdba21946ada948) | latest release 1.28.0 `ab33c9a19d626f8d0ac0bd1adfdba21946ada948`; source documentation only |
| [terminal-unicode-core](https://github.com/contour-terminal/terminal-unicode-core/tree/64f53851ceab9a3cf08db4939bcaef75a0899573) | draft snapshot `64f53851ceab9a3cf08db4939bcaef75a0899573` |

The harness used Go 1.26.10 (`unicode.Version` 15.0.0); selected module dependencies include `clipperhouse/displaywidth` v0.11.0, `clipperhouse/uax29/v2` v2.7.0 (Unicode 17 tables), `mattn/go-runewidth` v0.0.30, and `rivo/uniseg` v0.4.7 (Unicode 15 tables). These versions do not imply all packages use the same tables. Loom combines Go Unicode categories, hand-coded ranges and emoji spec overrides; its complete behavior is not a single declared Unicode-table version.

Disposable canary: `/tmp/claude-1000/-home-uwe-projects-loom/9e4d0913-fa4b-4cb2-aec0-0c51a6bd4a87/scratchpad/b2-canary`, its own `go.mod` with `replace` to this checkout. Run `go build -o canary .` then `./canary` there. One program uses the same samples for all width APIs, UV/tcell cells, Loom canvas, Loom VT and x/vt. It sets/reloads `LOOM_ZWJ=join|split`, exercises both render paths in additional checks, and feeds bytes into in-process emulators. Synthetic requests have five-second deadlines; no sleeps or real terminal settings are involved. Results below are captured from this run, not from a terminal window. The Loom VT is used through `loomtest`, not a separately invented model.

## Capability ledger

`proved` means the stated narrow feature is demonstrated; `partial` means only the described portion is demonstrated; `disproved` means a specific claimed behavior fails; `unknown` is not evidence of absence. Columns are segmentation (Seg), width tables/methods (Width), terminal width detection (Detect), mode 2027 query/state (2027), cluster storage (Cell), and right-edge wide-cell clipping (Edge). For components that are only algorithms, emulator/terminal features are unknown rather than inferred absent.

| Component | Seg | Width | Detect | 2027 | Cell | Edge | Bounded evidence |
|---|---|---|---|---|---|---|---|
| Loom | partial | partial | partial | disproved | partial | proved | Heuristics split Indic; join/split CPR probe and override; VT replies unrecognized; canvas rejects an overflowing cluster. V1–V3, canary. |
| x/ansi | proved | proved | unknown | unknown | unknown | unknown | UAX-style segmentation and two width methods on this matrix; not a terminal or buffer. V4. |
| ultraviolet | partial | proved | partial | proved | partial | proved | Width methods and 2027 negotiation present; `NewCell` trusts input even though its comment promises first cluster; edge buffer blank. V5–V6. |
| x/vt | proved | partial | unknown | partial | partial | partial | Matrix uses grapheme processing, reports 2027 states, but off/on do not alter actual advance; edge behavior differs from Loom VT. V7. |
| tcell v3 | proved | proved | unknown | unknown | proved | partial | `Put` segments/stores first cluster; an out-of-room wide cell is stored with width 2. Physical output clipping not established. V8. |
| uniseg | partial | proved | unknown | unknown | unknown | unknown | v0.4.7 widths on matrix; Indic follows its older segmentation, not Unicode 17 behavior. |
| TUIOS emulator | unknown | partial | unknown | partial | unknown | unknown | Release mode declaration permanently sets 2027; no independent runtime or terminal negotiation claim. V9. |

These marks assess the named behavior, not complete Unicode conformance. Width `proved` means both methods or the single advertised method were observed for the matrix, not that every code point was audited.

## Cross-library width table

Numbers are cells, including full strings. J/S are Loom `LOOM_ZWJ=join/split`; W/G are x/ansi or UV `WcWidth/GraphemeWidth`. tcell is the first stored cluster's width; all samples here are one tcell cluster. uniseg is its complete-string width. The canvas clipping test starts a two-cell glyph in its final column; width remains 2 even when zero cells are accepted.

| Sample | Loom J/S | x/ansi W/G | UV W/G | uniseg | tcell v3 | Loom canvas accepted |
|---|---:|---:|---:|---:|---:|---:|
| `e` + U+0301 | 1/1 | 1/1 | 1/1 | 1 | 1 | 1 |
| `❤︎` VS15 | 1/1 | 1/1 | 1/1 | 1 | 1 | 1 |
| `❤️` VS16 | 2/2 | 1/2 | 1/2 | 2 | 2 | 2 |
| `👨‍👩‍👧‍👦` | 2/8 | 8/2 | 8/2 | 2 | 2 | 8 under split |
| `🇩🇪` | 2/2 | 2/2 | 2/2 | 2 | 2 | 2 |
| `क्ष` (KA + VIRAMA + SSA) | 2/2 | 2/1 | 2/1 | 2 | 1 | 2 |
| `界` at final column | 2/2 | 2/2 | 2/2 | 2 | 2 | **0**, blank remains |

Loom segments Indic as `क्` + `ष`; tcell and current x/ansi grapheme method treat the sample as one cluster. The harness directly inserted UV cells using the same strings; `NewCell` retains its argument, so this is not a general proof that it segments arbitrary input. On the clipping test UV's final cell remains blank; tcell's `CellBuffer.Put` retains `界` with width 2. This last observation is buffer state only, not proof of what its terminal renderer emits.

## Emulator table: mode 2027 off/on

Each starts at row 1 column 1 on a 40-column screen, except `界` starts at column 40. Entries are cursor advance in columns, derived from CPR, off/on; for edge the literal CPR is shown because wrapping makes subtraction meaningless. Loom's matrix ran with `LOOM_ZWJ=split` and `LOOM_RENDER_PATH=vte`; family join and standard-path checks are below.

| Input | Loom VT off/on | x/vt off/on | Comparison with computed width |
|---|---:|---:|---|
| combining | 1/1 | 1/1 | both match |
| VS15 | 1/1 | 1/1 | both match |
| VS16 | 1/1 | 2/2 | Loom measure 2; x/ansi W/G 1/2 |
| family | 8/8 | 2/2 | Loom split 8, join 2; x/ansi W/G 8/2 |
| flag | 2/2 | 2/2 | both match |
| Indic | 2/2 | 1/1 | Loom 2; x/ansi W/G 2/1 |
| `界` at col 40 | `2;3` / `2;3` | `1;40` / `1;40` | Loom wraps; x/vt leaves last cell blank; distinct from canvas clip |

DECRQM `CSI ? 2027 $ p` returned Loom `CSI ? 2027 ; 0 $ y` after both reset and set (unrecognized), and x/vt `;2$y` after reset and `;1$y` after set. x/vt's public `WidthMethod` changes W→G, while its actual UTF-8 flush uses G unconditionally: a reported reset is **not** proof of wcwidth rendering. These are emulator results, not real-terminal CPR results.

Loom's VTE-profile VS16 advance of 1 is an explicit pad-1 emulation branch (V3), not by itself proof of an end-to-end renderer defect: emitted padding was not measured here. Under the standard profile the same VS16 CPR advance was 2. Family advance was 2 with join and 8 with split under both profiles. The raw-byte edge test and Canvas API test intentionally exercise different layers.

## Loom implementation and upstream mechanisms: V excerpts

**V1 — Loom segmentation and width, [measure/measure.go](https://codeberg.org/ubunatic/loom/src/commit/d91ffced6f884540d74f2e56ffef299b04ad0f39/measure/measure.go#L130-L160), [width](https://codeberg.org/ubunatic/loom/src/commit/d91ffced6f884540d74f2e56ffef299b04ad0f39/measure/measure.go#L245-L264):**

```go
if isRegionalIndicator(r) {
    nextIdx := i + sz
    if nextIdx < n {
        r2, sz2 := utf8.DecodeRuneInString(s[nextIdx:])
        if isRegionalIndicator(r2) {
            return nextIdx + sz2, true, false, false, false
```

```go
if strings.ContainsRune(cluster, '\u200D') {
    return ZWJClusterWidth(cluster)
}
if w, ok := getGlyphOverride(cluster); ok {
    return w
}
```

The explicit cases cover flags, marks and joiners, but the canary demonstrates an Indic gap. [ZWJ probing](https://codeberg.org/ubunatic/loom/src/commit/d91ffced6f884540d74f2e56ffef299b04ad0f39/measure/zwj.go#L23-L33) tries the `LOOM_ZWJ` override first then a 200 ms terminal CPR probe. It samples family behavior; it does not negotiate a complete Unicode width policy.

**V2 — Loom canvas, [canvas.go](https://codeberg.org/ubunatic/loom/src/commit/d91ffced6f884540d74f2e56ffef299b04ad0f39/canvas.go#L360-L368), [continuations](https://codeberg.org/ubunatic/loom/src/commit/d91ffced6f884540d74f2e56ffef299b04ad0f39/canvas.go#L387-L394):**

```go
w = StringWidth(cell.Text)
}
if w >= 2 && x+w-1 >= c.cols {
    return
}
```

The buffer records a leading text cell and continuation cells. This proves the tested all-or-nothing right-edge behavior; it does not make incomplete segmentation complete.

**V3 — Loom VT, [profile compensation](https://codeberg.org/ubunatic/loom/src/commit/d91ffced6f884540d74f2e56ffef299b04ad0f39/vt/vt.go#L567-L572), [mode state](https://codeberg.org/ubunatic/loom/src/commit/d91ffced6f884540d74f2e56ffef299b04ad0f39/vt/vt.go#L735-L759):**

```go
w := measure.StringWidth(cluster)
if measure.ActiveRenderPath() == measure.RenderPathVTE && measure.VTEMode(cluster) == "pad-1" {
    w = max(1, w-1)
}
v.putCluster(r, cluster, w)
```

The `privateModeState` default returns 0 and its listed modes do not include 2027, matching the query result.

**V4 — x/ansi [width.go](https://github.com/charmbracelet/x/blob/faa4adf95555331dfb439b716cc83767ecdadeb4/ansi/width.go#L89-L104):**

```go
if m == WcWidth {
    // Wcwidth sums codepoints, so where the cluster boundaries
    // fall makes no difference to the total. Skipping the
    // segmentation is a good deal cheaper.
    r, size := utf8.DecodeRuneInString(s[i:])
    width += wcRuneWidth(r)
```

The alternate branch calls `FirstGraphemeCluster`; a common API does not make these methods equivalent.

**V5 — UV [cell.go](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/cell.go#L32-L45):**

```go
// NewCell creates a new cell from the given string grapheme. It will only use
// the first grapheme in the string and ignore the rest. The width of the cell
// is determined using the given width method.
func NewCell(method WidthMethod, gr string) *Cell {
```

The body stores `Content: gr` and computes `method.StringWidth(gr)` without truncation. Treat the caller contract and implementation separately; no multi-cluster input was used to infer a renderer bug.

**V6 — UV [terminal.go](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal.go#L320-L336), [screen](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_screen.go#L163-L188):**

```go
if mr, ok := ev.(ModeReportEvent); ok {
    if mr.Mode == ansi.ModeUnicodeCore &&
        (mr.Value.IsSet() || mr.Value == ansi.ModeReset) {
        t.scr.enableGraphemeWidth()
    }
}
```

It queries at startup and switches buffer and renderer width methods on an eligible response, unless the application explicitly overrides the method. This is source evidence, not an external-terminal test.

**V7 — x/vt [emulator.go](https://github.com/charmbracelet/x/blob/43427fdbad12fde220b737fc36303291a5ce4602/vt/emulator.go#L161-L167), [utf8.go](https://github.com/charmbracelet/x/blob/43427fdbad12fde220b737fc36303291a5ce4602/vt/utf8.go#L50-L68):**

```go
if e.isModeSet(ansi.ModeUnicodeCore) {
    return ansi.GraphemeWidth
}
return ansi.WcWidth
```

```go
// XXX: We always use [ansi.GraphemeWidth] here to report accurate widths
// and it's up to the caller to decide how to handle Unicode vs non-Unicode
// modes.
method := ansi.GraphemeWidth
```

The apparently contradictory method and CPR are explained by the actual output path. Do not use this emulator as a wcwidth-off oracle.

**V8 — tcell [cell.go](https://github.com/gdamore/tcell/blob/4bed18057fce88934c9b9a576b5f83632196de11/cell.go#L85-L91):**

```go
g := textWidthOptions.StringGraphemes(str)
for width == 0 && g.Next() {
    cluster := g.Value()
    cl += cluster
    width = g.Width()
    str = str[len(cluster):]
}
```

**V9 — [TUIOS release mode.go](https://github.com/Gaurav-Gosain/tuios/blob/850335c0c08bb8b802406fcc7f254355c89af04a/vt/mode.go#L49-L51):**

```go
{ansi.ModeUnicodeCore, ansi.ModePermanentlySet}, // ?2027, see WidthMethod
```

This is its emulated terminal's declared mode, not a finding that it probes the outer terminal. The [Unicode core draft](https://github.com/contour-terminal/terminal-unicode-core/blob/64f53851ceab9a3cf08db4939bcaef75a0899573/spec/terminal-unicode-core.tex#L104-L130) describes mode switching in a Unicode-13-era proposal; it does not guarantee identical Unicode tables in independently updated implementations. [foot 1.28.0](https://codeberg.org/dnkl/foot/src/commit/ab33c9a19d626f8d0ac0bd1adfdba21946ada948/doc/foot-ctlseqs.7.scd#L337-L339) documents 2027, but documentation is not a CPR experiment.

## Stage 4 proposals

1. `features.md` width row: replace the broad “missing” with “no 2027 negotiation/state; partial CPR-based ZWJ policy; canvas stores measured clusters and rejects a wide glyph crossing the edge; Indic segmentation incomplete.” Split mode negotiation, segmentation and cell clipping into separate rows. Keep terminal-specific claims scoped to tested evidence.
2. `features.md` cluster-storage row: change “has” to “partial: leading cell plus continuations, using Loom's incomplete cluster segmentation.” Add the Indic counterexample and Loom source links V1–V2.
3. `features-charmx.yaml` `vt.widthmethod-mode-2027`: change UNVER to source-verified **partial**, pin V7, and note that `WidthMethod` changes but actual `flushGrapheme` always uses GraphemeWidth; replace the non-verbatim quote. Pin `ansi.width-method` and `ansi.stringwidth-grapheme-default` to V4 and the resolved revision. Add UV negotiation and edge-buffer rows using V5–V6 and the canary.
4. `features-contrast.yaml` `tcell.cell.put-grapheme` and `tcell.width.dependency`: update to v3.6.0 and V8, record current dependency versions and edge storage observation. Pin `features-terminals.yaml` `term.grapheme-2027` to the draft and foot release and retain its draft/version caveat. `features-tuios.yaml` `tuios.width-mode-2027`: pin release V9 and distinguish emulated state from outer-terminal negotiation.
5. `summary.md` item 2: propose 2027 query as a candidate, not a proven universal replacement for CPR/name heuristics; include unknown and permanently reset handling, timeout, explicit override, restoration of prior state, and consistent selected width policy across measuring, buffer, renderer and VT. Real terminal and multiplexer checks are outstanding. Do not claim enabling 2027 aligns Unicode versions.
6. Ticket candidate: **complete grapheme segmentation and align shared width/storage**. Reproduce `क्ष` as two Loom clusters/width 2 versus Unicode-17-backed tcell/x/ansi cluster/width 1; add conformance fixtures and preserve documented overrides. Decide Unicode table lifecycle explicitly. API compatibility is P3, so change interfaces if correctness requires it.
7. Ticket candidate: **negotiate and model 2027 end to end**. Use V6 as a mechanism, not a copied compatibility assumption; test DECRQM states 0–4, bounded no-response, set/reset/restoration, policy propagation and VT CPR. Current Loom VT reports 0 after both requests; x/vt is not an off-mode oracle.
8. Ticket candidate: **end-to-end edge and VS16 rendering regression tests**. Test canvas clipping and overwritten continuations, renderer output, CPR and VTE pad compensation together; this experiment proves differences between layers, not a user-visible regression. Include terminal-specific checks before assigning severity.

## Owner manual check (not run)

For each of foot, kitty and Ghostty separately, open a disposable window at least 40 columns wide and record terminal name/version, OS, multiplexer presence and Unicode settings. In that window use a terminal diagnostic capable of sending `CSI ? 2027 $ p` and reading its DECRPM, then obtain CPR (`CSI 6 n`) before and after each of the exact seven strings in the width table, including `界` positioned at column 40. Record the initial reply and row/column pairs; compare differences with the table. If the reply is 1 or 2, repeat after `CSI ? 2027 l` and `CSI ? 2027 h`, and restore the original state afterward. If it is 0, 3, 4 or there is no reply, record that outcome and do not toggle. Run without a multiplexer first, then repeat inside the relevant multiplexer if desired. Each response must have a five-second timeout and the diagnostic must restore its input mode even on error. A mode response alone is not evidence of matching cursor width.

Paste the following in each disposable terminal. It uses `/dev/tty`, a five-second deadline for each reply, and restores the original 2027 state and tty attributes in `finally`. The test clears the current line, so do not run it over work you need to preserve. Capture its printed results and the terminal version. The right-edge probe uses absolute column 40; it requires at least two screen rows.

```sh
python3 - <<'PYMANUAL'
import os, re, select, termios, time, tty

fd = os.open('/dev/tty', os.O_RDWR | os.O_NOCTTY)
old = termios.tcgetattr(fd)
esc = b'\x1b'
samples = [('combining', 'e\u0301'), ('VS15', '❤︎'), ('VS16', '❤️'),
           ('family', '👨‍👩‍👧‍👦'), ('flag', '🇩🇪'), ('Indic', 'क्ष'),
           ('edge', '界')]
results = []
state = None

def request(data, pattern):
    os.write(fd, data)
    deadline = time.monotonic() + 5
    data = b''
    while time.monotonic() < deadline:
        ready, _, _ = select.select([fd], [], [], max(0, deadline - time.monotonic()))
        if not ready:
            break
        data += os.read(fd, 1024)
        match = re.search(pattern, data)
        if match:
            return match.groups()
    return None

def cpr():
    return request(esc + b'[6n', rb'\x1b\[([0-9]+);([0-9]+)R')

try:
    tty.setraw(fd)
    reply = request(esc + b'[?2027$p', rb'\x1b\[\?2027;([0-4])\$y')
    state = int(reply[0]) if reply else None
    results.append(('initial 2027 state', state))
    modes = [('initial', None)]
    if state in (1, 2):
        modes += [('off', b'l'), ('on', b'h')]
    for label, mode in modes:
        if mode:
            os.write(fd, esc + b'[?2027' + mode)
        for name, sample in samples:
            os.write(fd, b'\r' + esc + b'[2K')
            if name == 'edge':
                os.write(fd, esc + b'[40G')
            before = cpr()
            os.write(fd, sample.encode())
            after = cpr()
            results.append((label, name, before, after))
finally:
    if state in (1, 2):
        os.write(fd, esc + b'[?2027' + (b'h' if state == 1 else b'l'))
    termios.tcsetattr(fd, termios.TCSADRAIN, old)
    os.close(fd)
for result in results:
    print(result)
PYMANUAL
```

## Owner manual check results (run 2026-11-10)

Run by the owner with the script above (edge test moved to the window's last column), Linux,
no multiplexer. Ghostty is not installed. Cursor advance in cells, from CPR.

| Terminal | DECRQM 2027 | combining | VS15 | VS16 | family | flag | Indic `क्ष` | wide at last column |
|---|---|---:|---:|---:|---:|---:|---:|---|
| foot 1.27.0, default (set) | 1 (set) | 1 | 1 | 2 | 2 | 2 | 2 | not at edge (140 cols, test at col 40) |
| foot 1.27.0, 2027 reset | | 1 | 1 | 2 | **8** | 2 | 2 | not at edge |
| kitty 0.47.1 | **0** (not recognised) | 1 | 1 | 2 | 2 | 2 | **1** | wraps to next row, col 3 |

Findings:
- foot reports 2027 set by default, and its reset mode matches Loom `LOOM_ZWJ=split` for the
  family (8). A DECRQM query would select foot's policy without the CPR probe.
- kitty does grapheme widths but does not recognise 2027, so the query alone cannot detect it;
  a CPR fallback stays necessary.
- The Indic conjunct is 2 cells in foot and 1 in kitty. No single width table fits both; Loom's
  width 2 matches foot, x/ansi GraphemeWidth matches kitty.
- kitty wraps a wide glyph written at the last column; Loom Canvas never emits it there (V2).
- VS16 is 2 in both, matching Loom measure and x/ansi GraphemeWidth, not WcWidth (1).

## Flow notes

Latest-only re-resolution changed tcell from the initially found v2 line to v3.6.0; the final table uses v3. The historical batch-1 pins were not silently treated as current. Source and runtime evidence are separated, in particular for TUIOS and the terminal draft. No real terminal, visible window, or terminal mode was touched. The canary's finite sample set does not establish complete Unicode conformance or physical terminal behavior; the owner manual check remains outstanding.
