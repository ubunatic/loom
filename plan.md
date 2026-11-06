1. **Add `NextClusterBoundary` and `PreviousClusterBoundary` to `measure/measure.go`**
   - Rename `nextClusterBoundary` to `NextClusterBoundary` and update calls in `measure.go`.
   - Add `PreviousClusterBoundary(s string, i int) int` that iterates from the beginning of `s` using `NextClusterBoundary` to find the start of the cluster that ends at `i` (or the cluster that contains `i`).
   - Actually, since `TextInput` and `TextArea` use rune indices, we can add `NextClusterBoundaryRuneIndex(rs []rune, i int) int` and `PreviousClusterBoundaryRuneIndex(rs []rune, i int) int` directly into the `measure` package or `textinput.go`. But `textinput.go` does not have access to `JoinsAfterZWJ`, `isRegionalIndicator`, etc. from `measure` unless they are exported.
   - So, add `PreviousCluster` and `NextCluster` rune index helpers to `measure` or convert `[]rune` to `string` in the caller. Converting `[]rune` to `string` is simpler and doesn't pollute the `measure` API with rune-slice variants. Wait, we can just add `NextClusterBoundary` and `PreviousClusterBoundary` as exported functions in `measure`.

2. **Update `TextInput` in `textinput.go`**
   - Modify `ConsumeKey` for `"left"`, `"right"`, `"backspace"`, and `"delete"`.
   - Convert `t.runes` to string to find cluster boundaries in byte offsets, then map the new byte offset back to a rune index.
   - Or, write a small helper inside `TextInput` or `measure`:
     ```go
     // In measure/measure.go:
     func NextCluster(s string, byteOffset int) int {
         next, _, _, _, _ := nextClusterBoundary(s, byteOffset)
         return next
     }

     func PreviousCluster(s string, byteOffset int) int {
         if byteOffset <= 0 { return 0 }
         start := 0
         for {
             next, _, _, _, _ := nextClusterBoundary(s, start)
             if next >= byteOffset { return start }
             if next == start { return byteOffset - 1 }
             start = next
         }
     }
     ```
   - In `TextInput`:
     ```go
     val := string(t.runes)
     caretByte := len(string(t.runes[:t.caret]))
     ```
     For `backspace`:
     ```go
     prevByte := measure.PreviousCluster(val, caretByte)
     prevCaret := utf8.RuneCountInString(val[:prevByte])
     t.runes = append(t.runes[:prevCaret], t.runes[t.caret:]...)
     t.caret = prevCaret
     ```

3. **Update `TextArea` in `textarea.go`**
   - Similar mapping for `TextArea`'s `backspace()`, `deleteForward()`, `moveLeft()`, and `moveRight()`.
   - Ensure `TextArea` word movement uses correct navigation.

4. **Verify tests pass**
   - Make sure all existing tests pass (`go test ./...`).
   - Run the newly created fail tests, verify they now pass.

5. **Complete Pre Commit Steps**
   - Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.

6. **Submit**
   - Submit the change with the required ticket report.
