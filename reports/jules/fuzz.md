## Item fuzz-input (test-draft)
### Verdict
no crash found
### Changes
event_fuzz_test.go
### Tests
`go test -run '^$' -fuzz '^FuzzInput$' -fuzztime 15s .`
```
fuzz: elapsed: 0s, gathering baseline coverage: 0/8 completed
fuzz: elapsed: 0s, gathering baseline coverage: 8/8 completed, now fuzzing with 4 workers
fuzz: elapsed: 3s, execs: 21967 (7321/sec), new interesting: 55 (total: 63)
fuzz: elapsed: 6s, execs: 34317 (4117/sec), new interesting: 72 (total: 80)
fuzz: elapsed: 9s, execs: 34317 (0/sec), new interesting: 72 (total: 80)
fuzz: elapsed: 12s, execs: 34317 (0/sec), new interesting: 72 (total: 80)
fuzz: elapsed: 15s, execs: 87312 (17667/sec), new interesting: 76 (total: 84)
fuzz: elapsed: 16s, execs: 87312 (0/sec), new interesting: 76 (total: 84)
PASS
ok  	ubunatic.com/loom	16.077s
```
### Evidence
No crashers found.

## Item fuzz-vt (test-draft)
### Verdict
crashers found
### Changes
vt/vt_fuzz_test.go
### Tests
`go test -run '^$' -fuzz '^FuzzVTWrite$' -fuzztime 15s ./vt`
```
fuzz: elapsed: 0s, gathering baseline coverage: 0/7 completed
fuzz: elapsed: 0s, gathering baseline coverage: 7/7 completed, now fuzzing with 4 workers
fuzz: elapsed: 1s, execs: 18633 (12674/sec), new interesting: 69 (total: 76)
--- FAIL: FuzzVTWrite (1.47s)
    --- FAIL: FuzzVTWrite (0.00s)
        vt_fuzz_test.go:53: 477 cursor mismatch at split=1: 0,0 vs 1,0

    Failing input written to testdata/fuzz/FuzzVTWrite/4a679900cf2541b5
    To re-run:
    go test -run=FuzzVTWrite/4a679900cf2541b5
FAIL
FAIL	ubunatic.com/loom/vt	1.486s
```
### Evidence
`vt/vt_fuzz_test.go:53`
`t.Fatalf("477 cursor mismatch at split=%d: %d,%d vs %d,%d", split, cursorX1, cursorY1, cursorX2, cursorY2)`
Input: "\x1b[31ma"

## Item fuzz-ansibuffer (test-draft)
### Verdict
crashers found
### Changes
ansibuffer_fuzz_test.go
### Tests
`go test -run '^$' -fuzz '^FuzzAnsiBuffer$' -fuzztime 15s .`
```
fuzz: elapsed: 0s, gathering baseline coverage: 0/6 completed
fuzz: elapsed: 0s, gathering baseline coverage: 6/6 completed, now fuzzing with 4 workers
fuzz: elapsed: 1s, execs: 506 (442/sec), new interesting: 9 (total: 15)
--- FAIL: FuzzAnsiBuffer (1.15s)
    --- FAIL: FuzzAnsiBuffer (0.00s)
        ansibuffer_fuzz_test.go:44: Cell(0, 0) mismatch: loom.AnsiCell{Rune:48, FG:ansi.Color{mode:0x0, index:0x0, r:0x0, g:0x0, b:0x0}, BG:ansi.Color{mode:0x0, index:0x0, r:0x0, g:0x0, b:0x0}, Bold:false, Dim:false, Underline:false, Invert:true} vs loom.AnsiCell{Rune:48, FG:ansi.Color{mode:0x0, index:0x0, r:0x0, g:0x0, b:0x0}, BG:ansi.Color{mode:0x0, index:0x0, r:0x0, g:0x0, b:0x0}, Bold:false, Dim:false, Underline:false, Invert:false}

    Failing input written to testdata/fuzz/FuzzAnsiBuffer/db6bad08c5f580dc
    To re-run:
    go test -run=FuzzAnsiBuffer/db6bad08c5f580dc
FAIL
FAIL	ubunatic.com/loom	1.199s
```
### Evidence
`ansibuffer_fuzz_test.go:44`
`t.Fatalf("Cell(%d, %d) mismatch: %#v vs %#v", x, y, c1, c2)`
Input: "\x1b[0m"

## Item fuzz-measure (test-draft)
### Verdict
crashers found
### Changes
measure/measure_fuzz_test.go
### Tests
`go test -run '^$' -fuzz '^FuzzMeasure$' -fuzztime 15s ./measure`
```
fuzz: elapsed: 0s, gathering baseline coverage: 0/4 completed
fuzz: elapsed: 0s, gathering baseline coverage: 4/4 completed, now fuzzing with 4 workers
fuzz: minimizing 33-byte failing input file
fuzz: elapsed: 0s, minimizing
--- FAIL: FuzzMeasure (0.07s)
    --- FAIL: FuzzMeasure (0.00s)
        measure_fuzz_test.go:26: clusters do not concat back: "" vs "\b"

    Failing input written to testdata/fuzz/FuzzMeasure/7df5e1aff8e087f2
    To re-run:
    go test -run=FuzzMeasure/7df5e1aff8e087f2
FAIL
FAIL	ubunatic.com/loom/measure	0.086s
```
### Evidence
`measure/measure_fuzz_test.go:26`
`t.Fatalf("clusters do not concat back: %q vs %q", concat, s)`
Input: "\b"
