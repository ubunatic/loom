## Item uv-trial (experiment)

### Verdict

works. We successfully built a wrapper comparing loom's full-screen flush with ultraviolet's differential renderer.

### Approach

In a new module `experiments/uv/`, we wrote `canvasToUV` to convert a `loom.Canvas` into a `uv.TerminalScreen`. We verified equality by feeding the outputs into `vt.NewVT` and stringifying them. We benchmarked 4 scenarios (FirstPaint, OneLineChanged, SelectionMoves, Scroll) on a 200x60 canvas for both `loom.Canvas.Flush()` and `uv.TerminalScreen.Flush()`. We explicitly bypassed the adapter during the `Ultraviolet` benchmarks by manipulating the `TerminalScreen` directly to measure raw rendering overhead.

### Numbers

Run with `cd experiments/uv && go test -bench=.`

| Path | bytes per frame | ns/op | B/op | allocs/op |
|------|-----------|-------|--------------|-----------|
| FirstPaint/Loom | 12853 | 647914 | 92440 | 436 |
| FirstPaint/Ultraviolet | 12122 | 7055617 | 5891333 | 12832 |
| OneLineChanged/Loom | 12853 | 711370 | 96079 | 440 |
| OneLineChanged/Ultraviolet | 3 | 536420 | 23404 | 406 |
| SelectionMoves/Loom | 12871 | 703680 | 100036 | 439 |
| SelectionMoves/Ultraviolet | 403 | 585759 | 45287 | 412 |
| Scroll/Loom | 12871 | 2003381 | 311527 | 676 |
| Scroll/Ultraviolet | 244 | 2846991 | 1400871 | 24422 |

### Techniques worth adopting

1. Diffing cell changes. 1 that supports it. `canvas.go:210` could buffer the previous frame and only output cursor movements and differential cell writes to significantly decrease payload size for small mutations like single line changes.
2. Differential line rendering. 1 that supports it. `canvas.go:210` can avoid emitting entire unmodified rows for selections moving.

### Risks

Loom's `Row()` string builder in `canvas.go:171` is heavily optimized for throughput. Diffing cells per frame is more computationally expensive and heavily allocates in Ultraviolet's Cell architecture (24422 allocs vs Loom's 676 in Scroll). The diffing path is noticeably slower for first paint (7055617 ns vs 647914 ns) where Loom wins by a factor of 10.
