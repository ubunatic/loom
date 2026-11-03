## Verdict
done, fixed `VT.putCluster` to blank orphan halves of wide characters when overwritten.

## Changes
vt/vt.go
vt/vt_test.go

## Tests
```
go test -v ./vt -run TestVTWideOverwrite
```
```
=== RUN   TestVTWideOverwrite
=== RUN   TestVTWideOverwrite/wide_over_left_half
=== RUN   TestVTWideOverwrite/narrow_over_left_half
=== RUN   TestVTWideOverwrite/narrow_over_right_half
=== RUN   TestVTWideOverwrite/wide_over_right_half
=== RUN   TestVTWideOverwrite/wide_over_both_halves
--- PASS: TestVTWideOverwrite (0.00s)
    --- PASS: TestVTWideOverwrite/wide_over_left_half (0.00s)
    --- PASS: TestVTWideOverwrite/narrow_over_left_half (0.00s)
    --- PASS: TestVTWideOverwrite/narrow_over_right_half (0.00s)
    --- PASS: TestVTWideOverwrite/wide_over_right_half (0.00s)
    --- PASS: TestVTWideOverwrite/wide_over_both_halves (0.00s)
PASS
ok  	ubunatic.com/loom/vt	0.009s
```

Failing test before fix:
```
=== RUN   TestVTWideOverwrite
=== RUN   TestVTWideOverwrite/wide_over_left_half
    vt_test.go:544: got [" x"], want ["x "]
=== RUN   TestVTWideOverwrite/narrow_over_left_half
    vt_test.go:544: got ["x"], want ["x "]
=== RUN   TestVTWideOverwrite/narrow_over_right_half
    vt_test.go:544: got ["中x"], want [" x"]
=== RUN   TestVTWideOverwrite/wide_over_right_half
    vt_test.go:544: got ["中中"], want [" 中"]
=== RUN   TestVTWideOverwrite/wide_over_both_halves
--- FAIL: TestVTWideOverwrite (0.00s)
    --- FAIL: TestVTWideOverwrite/wide_over_left_half (0.00s)
    --- FAIL: TestVTWideOverwrite/narrow_over_left_half (0.00s)
    --- FAIL: TestVTWideOverwrite/narrow_over_right_half (0.00s)
    --- FAIL: TestVTWideOverwrite/wide_over_right_half (0.00s)
    --- PASS: TestVTWideOverwrite/wide_over_both_halves (0.00s)
FAIL
FAIL	ubunatic.com/loom/vt	0.008s
FAIL
```
