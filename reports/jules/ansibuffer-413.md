## Item 413 (implement)

### Verdict
done, fixed `ParseAnsiBuffer` relative moves, CSI A/B/D parsing, colon truecolor parsing, and wide cell continuation handling.

### Changes
- `ansibuffer.go`: Added tracking for virtual cursor movements to `ParseAnsiBuffer` prescan, implemented CSI A/B/D operations, added proper wide character continuation output (`Rune: 0` without space padding in Serialize), and enhanced `parseAnsiSGR` to properly handle truecolor formats with colons.
- `cmd/loom/crop.go`: Adjusted crop boundary validation to check for both `' '` and `0` when determining if a split lands on a wide character continuation cell.
- `ansibuffer_413_test.go`: Added table test covering relative moves bounds checking, relative moves parsing, colon truecolor parsing for foreground and background, and wide-character cell continuations.

### Tests
`go test -short ./...`
```
ok  	ubunatic.com/loom	1.277s
ok  	ubunatic.com/loom/ansi	(cached)
ok  	ubunatic.com/loom/cmd/loom	1.518s
... (all tests pass)
```

`go test -race -run 'Ansi' .`
```
ok  	ubunatic.com/loom	1.363s
```

`go test -run Test413 -v`
```
=== RUN   Test413_ParseAnsiBuffer
=== RUN   Test413_ParseAnsiBuffer/CSI_A_B_D_relative_moves
=== RUN   Test413_ParseAnsiBuffer/Prescan_bounds_check_B
=== RUN   Test413_ParseAnsiBuffer/Colon_truecolor_SGR_38:2::r:g:b
=== RUN   Test413_ParseAnsiBuffer/Colon_truecolor_SGR_48:2:r:g:b
=== RUN   Test413_ParseAnsiBuffer/Wide_cell_continuation_no_extra_space
--- PASS: Test413_ParseAnsiBuffer (0.00s)
    --- PASS: Test413_ParseAnsiBuffer/CSI_A_B_D_relative_moves (0.00s)
    --- PASS: Test413_ParseAnsiBuffer/Prescan_bounds_check_B (0.00s)
    --- PASS: Test413_ParseAnsiBuffer/Colon_truecolor_SGR_38:2::r:g:b (0.00s)
    --- PASS: Test413_ParseAnsiBuffer/Colon_truecolor_SGR_48:2:r:g:b (0.00s)
    --- PASS: Test413_ParseAnsiBuffer/Wide_cell_continuation_no_extra_space (0.00s)
PASS
ok  	ubunatic.com/loom	0.024s
```

### Evidence
- `ansibuffer.go:683`: `case 'A': num := parseAnsiCSINum(params, 1); y -= num; if y < 0 { y = 0 }`
- `ansibuffer.go:643`: `w := measure.RuneWidth(prescanRS[pi]); if w > 0 { if vY+1 > maxRows { maxRows = vY + 1 }; vX += w; if vX > maxCols { maxCols = vX } }`
- `ansibuffer.go:828`: `if strings.HasPrefix(p, "38:") || strings.HasPrefix(p, "48:") { subParts := strings.Split(p, ":")`
- `ansibuffer.go:782`: `if w > 1 && x+1 < buf.cols { contCell := curStyle; contCell.Rune = 0; buf.cells[y][x+1] = contCell }`
