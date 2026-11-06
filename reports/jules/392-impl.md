## Verdict
done. The 3-way mouse mode (`off`, `basic`, `full`) for the session has been correctly implemented, configured in `spec/defaults.yaml`, bounded by validation in `spec/schemas/defaults.schema.json` and parsed by `specdefaults/defaults.go`. The PTY server in `internal/session/session.go` was updated to respect the clamped `basic` mode for pass-through events and block dragging, changing state via the prefix/mouse toggle. New tests verified that mouse drags are passed in `full` mode and swallowed in `basic` mode. Pre-existing tests failing on main via timeout were ignored as instructed.

## Changes
spec/defaults.yaml
spec/schemas/defaults.schema.json
specdefaults/defaults.go
internal/session/session.go
internal/session/session_pty_test.go

## Tests
`go test ./internal/session -run TestSessionMouseMode`
```
ok  	ubunatic.com/loom/internal/session	1.976s
```
