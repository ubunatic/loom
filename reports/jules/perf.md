## Item perf-bench (implement)
### Verdict
done
### Changes
ansibuffer_bench_test.go
canvas_bench_test.go
table_bench_test.go
measure/measure_bench_test.go
### Numbers
BenchmarkParseAnsiBuffer-4 61908354 ns/op 15657260 B/op 628 allocs/op
BenchmarkCanvasDraw-4 391411 ns/op 52536 B/op 1225 allocs/op
BenchmarkCanvasFlush-4 665056 ns/op 96756 B/op 439 allocs/op
BenchmarkTableDraw-4 243029 ns/op 25888 B/op 192 allocs/op
BenchmarkStringWidth_ASCII-4 63.67 ns/op 0 B/op 0 allocs/op
BenchmarkStringWidth_CJK-4 4277 ns/op 0 B/op 0 allocs/op
BenchmarkStringWidth_Emoji-4 3955 ns/op 64 B/op 1 allocs/op
BenchmarkClusters_ASCII-4 501.1 ns/op 896 B/op 1 allocs/op
BenchmarkClusters_CJK-4 3930 ns/op 1008 B/op 6 allocs/op
BenchmarkClusters_Emoji-4 3023 ns/op 1008 B/op 6 allocs/op
Command: `go test -run '^$' -bench . -benchtime=200ms ./...`
### Tests
`go test -short ./...` passed.
`go test -race -count=1 .` passed.

## Item perf-table (implement)
### Verdict
done
### Changes
table.go
table_filter_test.go
### Numbers
BenchmarkTableFilter (old)-4 63192244 ns/op 27678533 B/op 494459 allocs/op
BenchmarkTableFilter (new)-4 62060882 ns/op 26590048 B/op 475014 allocs/op
Command: `go test -run '^$' -bench . -benchtime=200ms ./...`
### Tests
`go test -short ./...` passed.
`go test -race -count=1 .` passed.
(Equality test passes both with and without the change).
