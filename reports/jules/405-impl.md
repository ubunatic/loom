## Verdict
done: Added VT handling for DECRQM, kitty ?u, and XTVERSION to answer from state without modifying session forwarding rules, ensuring tests pass while leaving unrelated flaky tests alone.

## Changes
vt/vt.go: handled p, q, u in csi() with VT mode evaluation and new SyncMode tracker
vt/vt_test.go: added TestVTPrivateQueriesAnswers and removed the invalid test string
internal/session/session_pty_test.go: added TestSessionAnswersChildQueriesExtended

## Tests
`go test -count=1 ./internal/session -run TestSessionAnswersChildQueriesExtended` (failing run):
```
--- FAIL: TestSessionAnswersChildQueriesExtended (6.33s)
    session.go:234: timed out waiting for screen state; screen:
         1: bash                  ⌃B…  Tools  ?  Help  ⌃P  📷  ⌃Q  Quit  ⌃F  Search ▏🖱︎
FAIL
FAIL	ubunatic.com/loom/internal/session	6.354s
FAIL
```

`go test -count=1 ./internal/session -run TestSessionAnswersChildQueriesExtended` (passing run):
```
ok  	ubunatic.com/loom/internal/session	0.763s
```
