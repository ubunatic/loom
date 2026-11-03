## Verdict
done

Implemented parsing and answering of DECRQM, kitty `?u`, and XTVERSION queries directly from VT mode state in `vt.go`, leaving unhandled external replies to naturally pass through to the child terminal.

## Changes
- `vt/vt.go`: Implemented DECRQM (`$p`), XTVERSION (`>0q`, `>q`), and kitty keyboard protocol `?u` replies directly in `csi`.
- `vt/vt_test.go`: Added unit tests confirming exact replies for these new terminal queries.
- `internal/session/session_pty_test.go`: Added PTY integration test proving `?2026$p` receives exactly one well-formed answer.
- `event.go`: Handled kitty keyboard query reply logic in `isTerminalReply`.

## Tests
`go test ./vt/...` -> ok
`go test ./internal/session/... -run TestSessionAnswersChildQueries` -> ok
