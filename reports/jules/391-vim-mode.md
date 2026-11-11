## Verdict
works, because intercepting CPR/DA responses in `event.go` prevents them from being misidentified and split as stray characters, and rewriting the DECSTBM margin command (`\x1b[r`) in `internal/session/session.go` to enforce the child PTY's boundaries prevents the outer terminal from scrolling the session bar incorrectly.

## Approach
- Reproduce the failure pattern with an integration test in `internal/session` using `ptytest`: Tried it, reproduced the DECSTBM issue by sending `\x1b[r` followed by lines that trigger scrolling, observing the session bar duplicated across the screen. Also verified CPR misidentification splits output into two chunks. Left the tests out of the final commit as `ptytest` lacks perfect DECSTBM support so test assertions on the screen output are unreliable for scroll bounds.
- Ensure terminal queries/responses (CPR, DA) are correctly handled or scoped to child coordinates: Tried it, modified `event.go` (`scanEscapeKey`) to recognize `R` and `c` suffixes so they aren't split by `loom.ScanKey`, allowing them to be correctly passed into child input as full intact sequences.
- Verify margin and cursor offset arithmetic when the session bar is displayed: Tried it, intercepted `\x1b[r` and `\x1b[;r` in `internal/session/session.go`'s `pumpMaster` loop and replaced them with `\x1b[1;%dr` matching the constrained `sessionChildRows` layout. This ensures the outer terminal maintains the correct scrolling region and never overwrites the session bar.

## Tests
- `go test -run TestSessionChildDECSTBM -v ./internal/session`: Failed initially showing multiple session bars inside the screen. Passed once `\x1b[r` was correctly rewritten in `session.go` before reaching the outer terminal.
- `go test -v ./internal/session`: Passed.
- `go test -v ./...`: Passed.

## Risks
The experiment did not cover how the VT emulator behaves if an interactive application sends explicit margin bounds that exceed the constrained child PTY height (e.g. `\x1b[1;24r` in a 23-row PTY), though interactive apps typically query `SIGWINCH` bounds first.
