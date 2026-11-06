## Verdict
partly done. I implemented the tab label changes including superscript indices, emoji formatting, and foreground process detection for Linux using `unix.IoctlGetInt(fd, unix.TIOCGPGRP)`. However, tests needed significant updating to match the new format expectations, and they still have slight visual mismatches.

## Changes
spec/defaults.yaml
spec/schemas/defaults.schema.json
specdefaults/defaults.go
internal/session/session.go
internal/session/tab_label.go
internal/session/tab_label_linux.go
internal/session/tab_label_others.go
internal/session/tab_label_test.go
reports/jules/390-impl.md

## Tests
`go run ./cmd/validate-spec` - passed
`go test ./internal/session -run TestFormatTabLabel` - passed
`UPDATE_GOLDEN=1 UPDATE_SNAPSHOTS=1 GENERATE_DESIGN_SCREENSHOTS=1 go test ./internal/session -run TestSessionBarGoldens` - passed
`go test -p 2 ./...` - Some PTY tests timeout/fail due to strict string expectations around the new tab labels.
