# Render and diff follow-up (458, batch 1)

Status: **partial**. The same six sequential frames were measured in a shared
headless harness for Loom Canvas, Loom Pane and Bubble Tea v2. Pane write-call
counts and actual alternate-screen correctness remain unknown. The numbers are
single observations, not benchmark estimates or visual correctness evidence.

Pins rechecked 2026-11-10: Loom `e9af2380deddb53e2f72a935763774b4e2c71fbc`;
Bubble Tea latest release v2.1.1, commit `67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7`,
with its required UV `6b8d4baf91b4`; UV latest commit
`a45b7c86333f9dfedb95f81064e80259b02937f0`; TUIOS latest commit
`6fe2c0df9922b2eec28f392fee4e196efd86be36`. `go list -m ...@latest` and
`git ls-remote ... HEAD` resolved the moving names. Latest UV is traced below;
the Bubble Tea runtime uses its own required UV, not the newer commit. No
terminal application was opened; the Pane fixture uses a disposable PTY.

## Cross-library canary

Harness: `/tmp/claude-1000/-home-uwe-projects-loom/9e4d0913-fa4b-4cb2-aec0-0c51a6bd4a87/scratchpad/b1-canary`.
Run `go build -o canary .` and `timeout 35s ./canary`; the PTY harness has a
five-second deadline per entry. The entire successful matrix completed in under
one second. `go.mod` records the release, required UV and Loom replacement;
the latest UV pin is recorded in a comment because Go selects only one version
of a module. The harness copies the release's Go files without changes (verified
byte-for-byte) and adds `tea/canary_bridge.go` to call its otherwise private
renderer and synchronous flush. This includes Bubble Tea's renderer and its
UV dependency, but excludes its program event loop and frame ticker.

All start with an 80×24 blank inline region and run, in order, blank, forced
unchanged, X at (4,4), bold X, `界` at that position, then blank 80×20 with
four abandoned rows. Animation is off, cursor hidden, true color selected.
Canvas and Pane enable atomic flush, row clear and 2026; Bubble Tea has no
negotiated 2026 reply and thus does not wrap updates. The Pane uses `New`,
`Run`, `Post`, `Resize`, an 80-column cap and a controlling PTY; its first
reported frame includes writes from entering `Run` up to the 2026 end marker,
excluding the earlier `New` startup. Pane bytes therefore include ancillary
writes within that boundary. The other two boundaries are one forced Canvas
flush and Bubble Tea's render/flush including content construction. Shrink is
inline region shrink, not a physical screen resize.

Each cell below is **bytes / writes / allocations / microseconds**. Allocations
are process-wide `runtime.MemStats.Mallocs` deltas and include harness activity;
Pane spans control/PTY scheduling and Bubble Tea includes input-string
construction. Elapsed time has the same boundary differences. They cannot rank
renderer CPU cost. The PTY counts bytes read, not child write calls; `?` is
unknown, not zero. Canvas and Bubble Tea count calls to their output writer.

| Frame | Loom Canvas | Loom Pane | Bubble Tea v2.1.1 |
|---|---:|---:|---:|
| blank | 2269 / 1 / 156 / 80 | 2277 / ? / 227 / 290 | 18 / 1 / 95 / 351 |
| unchanged | 2269 / 1 / 156 / 81 | 2269 / ? / 159 / 145 | 0 / 0 / 3 / 2 |
| one cell | 2269 / 1 / 156 / 86 | 2269 / ? / 160 / 137 | 9 / 1 / 27 / 255 |
| style only | 2301 / 1 / 156 / 84 | 2301 / ? / 160 / 128 | 11 / 1 / 21 / 275 |
| wide cell | 2270 / 1 / 156 / 78 | 2270 / ? / 161 / 141 | 25 / 1 / 18 / 1516 |
| shrink | 1937 / 1 / 136 / 70 | 1937 / ? / 183 / 149 | 27 / 1 / 43 / 246 |

Expected: full Canvas rows on each forced flush, no Bubble Tea bytes for an
unchanged view, and stale-row clearing on shrink. The byte counts and
unchanged zero match those expectations; the harness did not replay the streams
into a terminal model, so visual equivalence, clipping and shrink correctness
are not established. The first attempted PTY fixture used Pane's default width
cap and was discarded before this corrected 80-column run. Forced unchanged
Pane is deliberately invalidated: an idle Pane need not redraw. The previously
reported Canvas origin-7 and non-atomic measurements were synthetic only;
actual alternate-screen switching, inline offset and suspend restoration were
not exercised in this rerun.

## Loom implementation

**Proved:** [Canvas flush](https://codeberg.org/ubunatic/loom/src/commit/e9af2380deddb53e2f72a935763774b4e2c71fbc/canvas.go#L576-L613) serializes every row, optionally clears stale rows, wraps configured 2026, and makes one output call when `AtomicFlush` is enabled. It has no previous-frame comparison in this function:

```go
for y := 0; y < c.rows; y++ {
    sink.WriteString(fmt.Sprintf("\x1b[%d;1H", startRow+y)) // move to row //nolint:errcheck
    if cfg.RowClear {
```

[Row](https://codeberg.org/ubunatic/loom/src/commit/e9af2380deddb53e2f72a935763774b4e2c71fbc/canvas.go#L521-L556) emits styles on within-row transitions and resets at row end; this is not temporal diff. [Pane redraw](https://codeberg.org/ubunatic/loom/src/commit/e9af2380deddb53e2f72a935763774b4e2c71fbc/pane_run.go#L260-L305) clears, draws, composes and flushes; its [loop gates redraw on dirty](https://codeberg.org/ubunatic/loom/src/commit/e9af2380deddb53e2f72a935763774b4e2c71fbc/pane_run.go#L429-L455). [Shrink accumulates stale rows](https://codeberg.org/ubunatic/loom/src/commit/e9af2380deddb53e2f72a935763774b4e2c71fbc/pane_screen.go#L34-L50). [Resume](https://codeberg.org/ubunatic/loom/src/commit/e9af2380deddb53e2f72a935763774b4e2c71fbc/pane_run.go#L226-L241) rebuilds after invalidation, and auto-wrap can be a separate write. [Metrics](https://codeberg.org/ubunatic/loom/src/commit/e9af2380deddb53e2f72a935763774b4e2c71fbc/pane_metrics.go#L68-L77) count that writer's bytes, not every terminal write. One writer call does not prove kernel atomicity or tear-free display. Graph coverage of the cited Loom files was best-effort clean; source was also read directly.

## Upstream mechanisms and mode 2026

TUIOS [serializes style-pair transitions through a cache](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/frame_render.go#L38-L49), then [returns composed content as a Bubble Tea view](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/render.go#L1318-L1330). The cache memoizes SGR transitions, not previous screen cells. Its separate [text-sizing passthrough wrapper](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/text_sizing_passthrough.go#L270-L283) is not the main-frame renderer.

TUIOS's [pinned go.mod](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/go.mod#L5-L19) uses an older Bubble Tea than the latest-release canary. Reusing the previous report's source trace of that deployed dependency, its [environment-gated query](https://github.com/charmbracelet/bubbletea/blob/d5bfd5c2ff7480a6b94536c1cf4e73569b8d523e/tea.go#L1115-L1124) and [ModeReset handling](https://github.com/charmbracelet/bubbletea/blob/d5bfd5c2ff7480a6b94536c1cf4e73569b8d523e/tea.go#L794-L801) enable synchronized updates; the renderer [wraps nonempty updates](https://github.com/charmbracelet/bubbletea/blob/d5bfd5c2ff7480a6b94536c1cf4e73569b8d523e/cursed_renderer.go#L576-L602) and [copies to output](https://github.com/charmbracelet/bubbletea/blob/d5bfd5c2ff7480a6b94536c1cf4e73569b8d523e/cursed_renderer.go#L608-L617). Thus main frames **can** carry 2026 when the query is sent and the reply says supported and reset; it is not unconditional. No reply or another state does not prove that enable path ran. This is source evidence, not an observed TUIOS session. No old dependency was used in the comparative canary.

At the current release, the [query gate](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/tea.go#L991-L1017) depends on environment, and the [reply handler](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/tea.go#L814-L821) enables the renderer only for `ModeReset`:

```go
case ansi.ModeSynchronizedOutput:
    if msg.Value == ansi.ModeReset {
        // The terminal supports synchronized output and it's
        // currently disabled, so we can enable it on the renderer.
        p.renderer.setSyncdUpdates(true)
```

Its [renderer wraps nonempty updates and copies the output buffer](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/cursed_renderer.go#L591-L632). The canary leaves this flag off; adding wrappers would change byte counts.

## Latest Ultraviolet line-hash path

At `a45b7c86333f9dfedb95f81064e80259b02937f0`, [Render calls scrollOptimize](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_renderer.go#L1493-L1519) only when scroll optimization and fullscreen are both enabled. [scrollOptimize](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_renderer_hardscroll.go#L9-L42) invokes updateHashmap and then attempts shifts. [hash](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_renderer_hashmap.go#L5-L13) hashes cell content; [updateHashmap](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_renderer_hashmap.go#L26-L51) updates old/new hashes and [maps unique displaced pairs](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_renderer_hashmap.go#L88-L96). Content hashes alone are not style-aware equality:

```go
h.Reset()
for _, c := range l {
    // maphash writes can not fail
    _, _ = h.WriteString(c.Content)
}
```

[scrolln](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_renderer_hardscroll.go#L70-L119) queues scroll operations and updates the prior-screen model. Remaining touched or damaged rows go through [paintLine](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_renderer.go#L864-L883), which may repaint drift-prone rows or use [cell comparison](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_renderer.go#L953-L982). Finally [Flush writes its buffer to the configured writer](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_renderer.go#L1313-L1324). Hashing is a conditional scroll optimization, not the general unchanged-line test. Latest UV was not separately benchmarked; the inline Bubble Tea canary does not exercise this fullscreen hash path.

## Stage 4 proposals

- Split temporal cell/row diff, within-frame style transition caching, and buffered output calls in `features.md` and raw YAML. Qualify Loom as full-row serialization on dirty redraw, not continuous idle repaint; distinguish Canvas buffering from the whole Pane.
- Correct `tuios.uv-style-diff` to SGR memoization; describe main-frame 2026 as negotiated and conditional, separate from text-sizing passthrough. Retain actual deployment pin separately from the latest-release comparison.
- Pin `uv.line-hash-diff`, `uv.transformline-diff` and `uv.render-flush-split` to the latest UV source path above; describe hashing as fullscreen scroll optimization and note drift repaint.
- Update `summary.md` with the measured byte observations and their differing boundaries, without claiming general speed or network savings. API stability and old-major comparisons belong at P3 at most.

## Flow notes

The same matrix now covers Canvas, Pane and the latest Bubble Tea, but PTY
bytes do not reveal child write-call counts; a syscall-aware recorder or a
counting-writer injection fixture is still needed. Actual alternate-screen and
nonzero inline origin checks, stream replay, and independent latest-UV runtime
measurements remain unrun. The five-second bound was not exceeded. The source
comparison cannot establish terminal display behavior or real workload cost.
No Loom source or other tracked study files were edited, and no tickets or
commits were made.
