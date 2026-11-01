## Verdict
The key decoding implementation and tests miss edge cases for Enter (LF) and lack coverage for `modifyOtherKeys` and split escape sequences.

## Evidence

1. **ESC LF for Alt+Enter** (`event.go:166`)
   - **Change**: Allow `\n` (LF) in addition to `\r` (CR).
   - **Why it helps**: Terminals sending LF instead of CR for Enter will correctly trigger `alt-enter`.
   - **Smallest step**: Change `if len(b) == 2 && b[1] == '\r'` to `if len(b) == 2 && (b[1] == '\r' || b[1] == '\n')`.

2. **LF in modifyOtherKeys Enter** (`event.go:479`)
   - **Change**: Add `code == 10` to the switch case for Enter.
   - **Why it helps**: Terminals sending LF in `modifyOtherKeys` mode are correctly decoded as `enter`.
   - **Smallest step**: Update `case code == 13:` to `case code == 13, 10:`.

3. **Untested ctrl-shift-enter** (`event_test.go:50`)
   - **Change**: Add test cases for `ctrl-shift-enter`.
   - **Why it helps**: Ensures this three-modifier combination is correctly decoded.
   - **Smallest step**: Add `{"csi-u ctrl-shift-enter", []byte("\x1b[13;6u"), KeyEvent{Key: "ctrl-shift-enter"}}` to the `tests` slice.

4. **Untested modifyOtherKeys** (`event_test.go:50`)
   - **Change**: Add general test cases for `modifyOtherKeys`.
   - **Why it helps**: Verifies the `decodeModifyOtherKeys` logic and prevents regressions.
   - **Smallest step**: Insert `{"modifyOtherKeys ctrl-enter", []byte("\x1b[27;5;13~"), KeyEvent{Key: "ctrl-enter"}}` to the table.

5. **Untested lone-ESC split** (`event_test.go:41`)
   - **Change**: Add a test verifying `scanKey` behavior on partial escape sequences.
   - **Why it helps**: Guarantees incomplete escape sequences safely return `ok=false` to wait for more data.
   - **Smallest step**: Write a loop test feeding `[]byte{27}` to `scanKey` and asserting `ok == false`.

## Smallest next step
Implement the `event.go` LF checks for Enter and add the missing test cases to `event_test.go`.

## Ticket corrections
None.
