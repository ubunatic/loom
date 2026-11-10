# Render and diff follow-up (458, batch 1)

Status: **partial**. Loom source and a headless Canvas matrix were checked;
the comparable upstream runtime matrix and actual Pane terminal lifecycle were
not run. No comparative speed or actual terminal-display conclusion follows.
Loom revision: `3861896304162a742f6f03e9378bbf7afd1173aa`.
Graph coverage for the five Loom source paths used was best-effort clean; source
was also read directly. No terminal was opened; terminal version: not applicable.

## Findings and evidence

- **Loom temporal diff: disproved for Canvas flush; dirty gating proved.**
  [FlushWithConfig](https://github.com/ubunatic/loom/blob/3861896304162a742f6f03e9378bbf7afd1173aa/canvas.go#L576-L599)
  loops over every row, positions, optionally clears and serializes it. There
  is no previous-frame comparison in that function. The caller
  [redraw](https://github.com/ubunatic/loom/blob/3861896304162a742f6f03e9378bbf7afd1173aa/pane_run.go#L260-L305)
  clears and redraws the canvas; the
  [loop](https://github.com/ubunatic/loom/blob/3861896304162a742f6f03e9378bbf7afd1173aa/pane_run.go#L432-L457)
  invokes it when dirty. Thus forced unchanged flush and idle Pane are different
  cases. Animation, ticks and invalidations can set dirty.
- **Style runs: proved.** [Row](https://github.com/ubunatic/loom/blob/3861896304162a742f6f03e9378bbf7afd1173aa/canvas.go#L521-L556)
  emits SGR on style transitions, handles continuation cells and resets at each
  row. This is within-frame style serialization, not temporal diff.
- **Atomic Canvas write and synchronization: proved with limits.**
  [FlushWithConfig](https://github.com/ubunatic/loom/blob/3861896304162a742f6f03e9378bbf7afd1173aa/canvas.go#L576-L612)
  conditionally buffers and performs one `WriteString`; its independent 2026
  flag surrounds canvas output. Pane can write auto-wrap, screen and cursor
  controls separately; [metrics](https://github.com/ubunatic/loom/blob/3861896304162a742f6f03e9378bbf7afd1173aa/pane_metrics.go#L70-L77)
  count successful bytes routed through that writer, not all terminal traffic.
  One call is not a guarantee of an atomic kernel write or tear-free display.
- **Resize, clearing and suspend: partial.** Flush emits explicit erase-line
  sequences for requested stale rows. Redraw replaces the canvas on size change
  and after resume; [resumeTerminal](https://github.com/ubunatic/loom/blob/3861896304162a742f6f03e9378bbf7afd1173aa/pane_suspend.go#L235-L265)
  sets repaint and invalidates. Full Pane resize/screen-switch correctness and
  ancillary bytes were not exercised.

### Upstream mechanisms

Bubble Tea v1.3.10 skips identical whole frames and, for changed frames, can
skip equal individual lines:
[`flush`](https://github.com/charmbracelet/bubbletea/blob/v1.3.10/standard_renderer.go#L151-L158),
[line comparison](https://github.com/charmbracelet/bubbletea/blob/v1.3.10/standard_renderer.go#L198-L208).
The latter does not mean each cell is diffed.

Ultraviolet at `a45b7c86333f9dfedb95f81064e80259b02937f0` compares old
and new lines in
[`transformLine`](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_renderer.go#L890-L893)
and finds the first differing cell:

```go
for firstCell < newbuf.Width() &&
    cellEqual(oldLine.At(firstCell), newLine.At(firstCell)) {
    firstCell++
}
```

The [following branch](https://github.com/charmbracelet/ultraviolet/blob/a45b7c86333f9dfedb95f81064e80259b02937f0/terminal_renderer.go#L940-L949)
returns when there is no difference. Drift-prone rows can instead be repainted;
hash behavior and final TerminalScreen write were not fully traced, so their
existing detailed claims remain **partial**, not verified.

TUIOS [memoizes style-pair transitions](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/frame_render.go#L38-L49):

```go
key := [2]uv.Style{*from, *to}
if s, ok := r.diffs[key]; ok {
    return s
}
```

This cache stores SGR strings, not old screen cells. Its
[`render`](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/frame_render.go#L52-L64)
serializes the lines to a string. TUIOS pins Bubble Tea v2.0.10 and UV
`0277a179edd9` in its [go.mod](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/go.mod#L5-L16);
the separately pinned UV source above is a newer comparison, not proof of the
deployed UV implementation. Bubble Tea v2.0.10 passes its cell buffer to UV
and flushes it ([source](https://github.com/charmbracelet/bubbletea/blob/v2.0.10/cursed_renderer.go#L477-L496)).
TUIOS text-sizing passthrough has its own
[2026 wrapper](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/text_sizing_passthrough.go#L270-L283).
Bubble Tea v2.1.1 separately wraps nonempty updates when synchronized updates
are enabled ([source](https://github.com/charmbracelet/bubbletea/blob/v2.1.1/cursed_renderer.go#L554-L578))
and copies its buffer to the writer ([source](https://github.com/charmbracelet/bubbletea/blob/v2.1.1/cursed_renderer.go#L583-L593)).
Main-frame TUIOS synchronization and exact v2.0.10 final output remain **unknown**
until its caller and negotiated state are traced; the passthrough-only claim is
not established.

## Headless Canvas canary

Harness: `/tmp/claude-1000/-home-uwe-projects-loom/9e4d0913-fa4b-4cb2-aec0-0c51a6bd4a87/scratchpad/b1-canary`,
`go mod tidy`, `go build -o canary .`, `timeout 5s ./canary`.
No animation or Pane was involved. True-color profile; explicit row clear and
2026 on; hidden cursor. Cases are sequential: blank, same canvas, X at (4,4),
bold X, wide 界, then blank 80×20 with four stale rows. Origin 1 simulates
alternate origin; origin 7 simulates a nonzero inline origin, without actually
switching screens. Each entry is one observation, not a benchmark.

| Frame | origin 1 bytes | origin 7 bytes | writes atomic / non-atomic | allocations atomic / non-atomic (origin 1) |
|---|---:|---:|---:|---:|
| blank | 2269 | 2275 | 1 / 75 | 160 / 155 |
| unchanged forced flush | 2269 | 2275 | 1 / 75 | 156 / 155 |
| one cell | 2269 | 2275 | 1 / 75 | 156 / 155 |
| style only | 2301 | 2307 | 1 / 75 | 156 / 155 |
| wide cell | 2270 | 2276 | 1 / 75 | 156 / 155 |
| shrink | 1937 | 1943 | 1 / 67 | 136 / 135 |

All 24 runs completed within the 5 s process deadline; observed individual
flush times ranged 129–345 µs. First CUP matched the requested origin; the
shrink included erase at the last stale row; all began and ended with expected
2026 controls. Expected full-frame byte behavior was observed. Allocations
include the counting writer and process-wide allocator noise. This checks
serialized controls, not visual correctness, clipping, screen transitions, or
the separately requested Pane ancillary write counts; those are **unknown**.
Identical upstream frame updates were not measured; no cross-library byte,
write, allocation or timing comparison is claimed.

## Stage 4 proposals

- `features.md`: split the combined render row into temporal row/cell diff
  (Loom missing on forced flush, dirty-gated), within-frame style transitions
  (Loom has), and atomic Canvas writer call (Loom has when enabled). Remove
  `tuios.uv-style-diff` from temporal diff. Qualify 2026 as configured output,
  not capability detection or whole-redraw atomicity.
- `features-charmx.yaml`: for `uv.transformline-diff` pin the URL to the
  UV SHA and cite lines 890–949; set Loom evidence to revision above,
  `canvas.go:576-599`. Mark `uv.line-hash-diff` and
  `uv.render-flush-split` partial pending execution-path verification;
  `uv.sync-output-2026` remains partial pending Flush tracing. Replace moving
  links only when the exact claims are verified.
- `features-tuios.yaml`: `tuios.uv-style-diff` is a style-pair SGR cache;
  cite `frame_render.go:38-49` and remove temporal-diff implication.
  `tuios.sync-2026-scope`: retain proved passthrough wrapper but change
  main-frame negative to unknown pending Bubble Tea v2.0.10 state/output trace.
- `features-bubbletea.yaml`: `bt.renderer.v1-string-diff` add the
  unchanged-line skip at `standard_renderer.go:198-208`;
  `bt.renderer.cell-based` identify the separate v2 tags and dependency
  versions; `bt.sync-output-2026` retain v2.1.1 conditional wrapper and
  mark TUIOS deployment behavior unknown.
- `summary.md`: replace “Bubble Tea renders more efficiently” and asserted
  SSH/session cost with “Loom redraws all rows on a dirty flush; comparative
  cost has not been measured.” Separate temporal diff from SGR memo and
  Canvas buffering. Do not recommend an implementation based on this canary.

## Flow notes

The plan does not specify whether the matrix measures Canvas alone or the
entire Pane and all competing libraries, nor a headless fixture for private
renderers and mode transitions. The source lock pins UV separately from
TUIOS's UV dependency, requiring a second dependency trace. This run did not
complete that trace, upstream matrix, or Pane ancillary counts; they remain
explicitly unknown rather than inferred. Time spent: a few minutes; token usage
not exposed to this agent. A reusable pinned harness and named measurement
boundaries would reduce cost for the remaining batches.
