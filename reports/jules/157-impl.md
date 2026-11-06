## Verdict
done, `loomtest.CompareWidget` now checks for bold/dim attributes, handles wide and continuation cells correctly, and includes new table-driven unit tests to prove exact match functionality and deliberate mismatch reporting.

## Changes
loomtest/compare.go
loomtest/compare_test.go

## Tests
`git checkout HEAD~1 -- loomtest/compare.go`
`go test ./loomtest/ -run TestCompareANSIReport -v`
```
=== RUN   TestCompareANSIReport
=== RUN   TestCompareANSIReport/bold_diff
    compare_test.go:27: expected diff=true, got report:
        differences: 0
    compare_test.go:29: differences: 0

=== RUN   TestCompareANSIReport/dim_diff
    compare_test.go:27: expected diff=true, got report:
        differences: 0
    compare_test.go:29: differences: 0

=== RUN   TestCompareANSIReport/same
    compare_test.go:29: differences: 0

=== RUN   TestCompareANSIReport/wide_char_diff
    compare_test.go:29: row=1 col=1 want={glyph="こ" fg=default bg=default} got={glyph="ん" fg=default bg=default}
        differences: 1

=== RUN   TestCompareANSIReport/wide_vs_ascii_diff
    compare_test.go:29: row=1 col=1 want={glyph="こ" fg=default bg=default} got={glyph="a" fg=default bg=default}
        row=1 col=2 want={glyph="\x00" fg=default bg=default} got={glyph="b" fg=default bg=default}
        differences: 2

--- FAIL: TestCompareANSIReport (0.00s)
    --- FAIL: TestCompareANSIReport/bold_diff (0.00s)
    --- FAIL: TestCompareANSIReport/dim_diff (0.00s)
    --- PASS: TestCompareANSIReport/same (0.00s)
    --- PASS: TestCompareANSIReport/wide_char_diff (0.00s)
    --- PASS: TestCompareANSIReport/wide_vs_ascii_diff (0.00s)
FAIL
FAIL	ubunatic.com/loom/loomtest	0.008s
FAIL
```
`git checkout test-impl -- loomtest/compare.go`
`go test ./loomtest/ -run TestCompareANSIReport -v`
```
=== RUN   TestCompareANSIReport
=== RUN   TestCompareANSIReport/bold_diff
    compare_test.go:29: row=1 col=1 want={glyph="B" fg=default bg=default bold=true dim=false} got={glyph="B" fg=default bg=default bold=false dim=false}
        differences: 1

=== RUN   TestCompareANSIReport/dim_diff
    compare_test.go:29: row=1 col=1 want={glyph="D" fg=default bg=default bold=false dim=true} got={glyph="D" fg=default bg=default bold=false dim=false}
        differences: 1

=== RUN   TestCompareANSIReport/same
    compare_test.go:29: differences: 0

=== RUN   TestCompareANSIReport/wide_char_diff
    compare_test.go:29: row=1 col=1 want={glyph="こ" fg=default bg=default bold=false dim=false} got={glyph="ん" fg=default bg=default bold=false dim=false}
        differences: 1

=== RUN   TestCompareANSIReport/wide_vs_ascii_diff
    compare_test.go:29: row=1 col=1 want={glyph="こ" fg=default bg=default bold=false dim=false} got={glyph="a" fg=default bg=default bold=false dim=false}
        row=1 col=2 want={glyph="\x00" fg=default bg=default bold=false dim=false} got={glyph="b" fg=default bg=default bold=false dim=false}
        differences: 2

--- PASS: TestCompareANSIReport (0.00s)
    --- PASS: TestCompareANSIReport/bold_diff (0.00s)
    --- PASS: TestCompareANSIReport/dim_diff (0.00s)
    --- PASS: TestCompareANSIReport/same (0.00s)
    --- PASS: TestCompareANSIReport/wide_char_diff (0.00s)
    --- PASS: TestCompareANSIReport/wide_vs_ascii_diff (0.00s)
PASS
ok  	ubunatic.com/loom/loomtest	0.008s
```
