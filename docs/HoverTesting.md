# Hover Testing

How to prove in a black-box PTY test that a mouse hover at terminal cell
(x, y) highlights the widget under (x, y), and how to report the offset if it
doesn't. Reference implementation: `examples/loomoji/loomoji_107_pty_test.go`
(issue 107). It measured a dy = −1 offset that 108 fixed (`e.Y-1` on an already
0-based child-local Y, see 059 and [Widgets](Widgets.md)); the probe now passes
and guards against regressions.

## Rules

- **Black-box only.** Drive the real binary through `internal/ptytest`. Take no
  colours, geometry or state from the app source, and add no debug flags.
- **Observe colours, not text.** A hover usually changes only the BG, so read
  `Session.Cells()` / `CellFrames()`, not `Screen()`.
- **Compare effective colours.** Use `Cell.Style.Effective()`: it swaps FG and
  BG under reverse video (SGR 7).

## Coordinates

| Space | Origin | Used by |
|---|---|---|
| Terminal cell | 0-based (x, y) | `Cells()[y][x]`, test tables |
| SGR mouse report | 1-based | `\x1b[<35;X;YM` (35 = motion, no button) |

Send `x+1, y+1`. Mixing these up gives exactly the ±1 offsets you're
trying to detect.

Wide runes (emoji) take two cells. The trailing half has `Rune == 0` and the
same style. Find item columns by scanning cells with display width, never by
rune index into a `Screen()` string (`findPTYText` has that flaw).

## Recipe

```
start app ─► wait for first screen
for each probe:
  hover off-grid ─► wait+settle ─► baseline grid
  hover (x,y)    ─► wait+settle ─► hover grid
  diff effective BG ─► changed-cell bounding box
  offset = box.origin − expected item origin
```

1. **Geometry from the screen.** After startup, scan `Cells()` for the item
   runes and record `startX`, `width` and `y` for each item.
2. **Baseline per probe.** Diff every hover against a grid captured right
   before it. Never search for a known highlight colour: the startup focus
   carries the same colour and wins every scan.
3. **Wait for a frame, then settle.** After `SendRaw`, wait for at least one
   new `CellFrames()` entry (timeout ~2 s), then for a quiet window (~30 ms)
   with no further frames. A quiet window alone returns before the redraw,
   and the late frame bleeds into the next probe.
4. **Diff.** Collect the cells whose effective BG differs from the baseline
   and report the bounding box.
5. **Classify.**

| Status | Meaning |
|---|---|
| `PASS` | box origin = expected item origin (or no change off-grid) |
| `OFFSET_BUG` | change found elsewhere; report (dx, dy) |
| `NO_CHANGE` | no frame or no diff: *no visible target*, not a pass |
| `UNEXPECTED_HL` | a hover outside any item changed something |

6. **Sanity gate.** If no on-grid probe produces any change, fail with "hover
   not observable" instead of printing a table of fake results.
7. **Summary.** Print the dominant (dx, dy) across all measured probes.

## Probe placement

An offset only shows if the cell it lands on can visibly change. Put the
main probes on rows and columns with room on every side:

- Middle grid rows (not the first two), with several columns per row.
- For each item: first cell, last cell, both emoji halves, and the gap after it.
- Edge cases separately: padding row/column, search bar, first grid row.

Probes whose target lands off-grid or on the already-focused item produce
`NO_CHANGE`. That pattern is itself evidence of an offset, but it isn't a
measurement.

## Pitfalls seen in issue 107

| Symptom | Cause |
|---|---|
| Every probe reports the same cell | colour search matched the startup focus |
| Change shows the previous probe's item | settled before the redraw arrived |
| Mostly `NO_CHANGE` on the top rows | dy = −1 lands on the padding or the focused item |
| Second probe on the same item shows `NO_CHANGE` | off-grid hover doesn't clear the highlight, so the baseline already has it |

## Running

A hover probe that exposes an open bug fails by design. Keep it in its own
test function so `-run` can isolate it:

```sh
go test -count=1 -run TestLoomoji107HoverProbe -v ./examples/loomoji/
```

Under Quota-1, use `make test-q1`, write the output to a file and grep for
`--- FAIL`.

## Open items

- The probe helpers (`sendAndSettle`, baseline diff, `findGridItems`) live in
  the loomoji test. Move them to `internal/ptytest` when a second app needs
  hover probes.
- Blind spot: an off-grid hover doesn't clear loomoji's highlight, so a
  second probe on the same item (right emoji half, gap) can't be measured.
  Reset by hovering a different item instead.
