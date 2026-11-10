1. **Item perf-bench (implement):** Add benchmarks for hot paths.
   - Create `canvas_bench_test.go`: Add `BenchmarkCanvasDrawBorder`, `BenchmarkCanvasFlush` for a 200x60 screen (look at `BenchmarkCanvasWriteANSI` as an example).
   - Create `table_bench_test.go`: Add `BenchmarkTableDraw` for a Table with 10k rows. Add benchmark for table filter `BenchmarkTableFilter` at 10k or 100k rows.
   - Create `measure/measure_bench_test.go`: Add benchmarks for width and cluster on ASCII, CJK, and emojis.
   - Create `ansibuffer_bench_test.go`: Add `BenchmarkParseAnsiBuffer` for a 200x60 colored screen.
   - Run the new benchmarks using `go test -run '^$' -bench . -benchtime=200ms` and record results in the report.

2. **Item perf-table (implement):** Incremental Table filter.
   - Update `Table.refilter()` in `table.go`: check if the new `t.query` extends the *previous* query (which we need to cache). If `q` starts with `old_q` and we have previously filtered rows, only iterate over `t.filtered`. Otherwise, fallback to full refilter over `t.Rows`.
   - Add a test `TestTableIncrementalFilter` in `table_test.go` (or `table_perf_test.go`): type and delete a filter on a 1k-row table, compare against full refilter. Ensure it fails without the change for testing the difference (even if they have the same result, testing for *how* it's done might require comparing execution time or using an interceptor, but we can verify the equality part works). Let's explicitly check `t.query` length vs previous in `refilter` and filter `t.filtered` when applicable.
   - Run tests: `go test -short ./...` and `go test -race -count=1 .`.

3. **Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.**
4. **Submit changes and report:** Write report to `reports/jules/perf.md` and commit changes.
