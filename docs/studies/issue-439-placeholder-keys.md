# Issue 439: placeholder acceptance keys

The existing `AcceptPlaceholder` inserts the entire hint in one change and returns success after the callback. A regression run before the guard failed for Tab, End and Ctrl+E: none inserted the hint. The guard must precede configured bindings and preserve fallback routing when it does not apply.

Verification: focused regression, gallery and PTY cases passed after the fix; `go test . ./gallery ./cmd/loom`, `go vet ./...`, and one `make test-q1` passed. The quota log is `/tmp/loom-issue439-q1.log`. PTY coverage sends bytes 09, ESC [ F, and 05 and checks the retained hint and end cursor; frame coverage checks actual focus traversal.
