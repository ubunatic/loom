# Issue 440: TextEntry Ctrl+D routing

- **Finding**: named Ctrl+D and decoded byte `0x04` are ignored by TextEntry, including with text, a selection, or a placeholder. A Frame offers the event to its focused child first and can handle it with its own binding. This proves existing behavior; it does not itself exit an entry or application.
- **Coverage**: the contract test checks entry state, clipboard, undo/redo history and callbacks, and verifies one child delivery and one Frame toggle. Existing Ctrl+C, Ctrl+U and placeholder tests remain intact. These tests pass on the existing implementation, as expected for a contract lock-down.
- **Defaults**: pane fallback uses spec quit keys Ctrl+Q and F10, plus Ctrl+C as an interrupt; Esc is opt-in. Ctrl+D is not a pane fallback. The gallery sets `DisableDefaultQuit`. Session direct quit is Ctrl+Q and prefix quit is `q`; session bindings are separate from pane defaults.
- **Verification**: `go test . -run '^TestTextEntry' -count=1`, `go vet ./...`, `make install`, and one `make test-q1` passed. The full check log is `/tmp/loom-issue440-test-q1.log`; no `--- FAIL` appeared. No production or decoder changes were needed.
