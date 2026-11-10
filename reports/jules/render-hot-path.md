## Verdict

The `parse_ansi` and `canvas` modules have rendering hot paths that involve repeated allocation or redundant evaluation. We identified `os.Getenv` read for ANSI settings during parsing and `fmt.Sprintf` calls used to build terminal escape sequences per row in the canvas flush methods.

## Changes

1. `parse_ansi.go:16`: `useFastANSI` calls `os.Getenv("LOOM_FAST_ANSI")` on every ANSI parsing operation. Why it helps: `os.Getenv` acquires locks and creates string allocations inside hot parsing loops. Smallest step: Read `LOOM_FAST_ANSI` once into a package-level variable using `sync.Once` or `init()`.
2. `canvas.go:587`: `sink.WriteString(fmt.Sprintf("\x1b[%d;1H", startRow+y))` uses `fmt.Sprintf` inside `FlushWithConfig`. Why it helps: `fmt.Sprintf` is reflection-heavy and allocates strings for every row flushed per frame. Smallest step: Construct the ANSI cursor move sequence into a local buffer using `strconv.AppendInt`.
3. `canvas.go:596`: `sink.WriteString(fmt.Sprintf("\x1b[%d;1H\x1b[2K", startRow+c.rows+y))` inside `FlushWithConfig` clears lines using `fmt.Sprintf`. Why it helps: Prevents allocations during row clearing loops. Smallest step: Use a local byte buffer and `strconv.AppendInt` to generate the escape sequence.
4. `canvas.go:600`: `sink.WriteString(fmt.Sprintf("\x1b[%d;%dH", startRow+c.CursorY, c.CursorX+1))` uses `fmt.Sprintf` to place the hardware cursor at the end of `FlushWithConfig`. Why it helps: Replaces another heap-allocating formatting call in the render flush. Smallest step: Use `strconv.AppendInt` into the existing `strings.Builder`.
5. `pane_terminal.go:234`: `b.WriteString(fmt.Sprintf("\x1b[%d;1H\x1b[2K", p.startRow+i))` uses `fmt.Sprintf` during line clearing operations. Why it helps: Removing `fmt.Sprintf` inside pane clearing allocations reduces garbage collection pressure when redrawing pane segments. Smallest step: Construct the escape sequence directly with `strconv.AppendInt`.

## Tests

I left the failing tests in `TestLiveTmuxScriptRunsStepsAndMeasuresPause`, `TestEditPTYPlainTextSaveWritesExactBytes`, and various `internal/session` tests as-is, following the one-shot ticket directives, since no code was modified.
