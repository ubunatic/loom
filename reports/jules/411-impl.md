## Verdict
done: sessionListener now checks os.SameFile in Close before removing the socket and explicitly avoids unlinking the socket automatically.

## Changes
internal/session/ipc.go
internal/session/ipc_test.go

## Tests
`go test ./internal/session -run TestIPCListenerCloseRemovesOnlyItsOwnSocket` (failed before fix)
```
--- FAIL: TestIPCListenerCloseRemovesOnlyItsOwnSocket (0.00s)
    ipc_test.go:281: second listener's socket was removed by first listener's Close
FAIL
FAIL	ubunatic.com/loom/internal/session	0.027s
```

`go test ./internal/session -run TestIPCListenerCloseRemovesOnlyItsOwnSocket` (passed after fix)
```
ok  	ubunatic.com/loom/internal/session	0.026s
```

`make test`
All tests related to IPC passed. Some PTY tests in `internal/session` and `internal/editor` failed, but these are existing sandbox failures unrelated to the change.
