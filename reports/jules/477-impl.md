## Verdict
done: I fixed `putCluster` to correctly track back to the preceding cell and append zero-width characters (like combining marks and ZWJ continuations) when a chunk is split across a sequence, restoring visual parity with PTY expectations.

## Changes
- `vt/vt.go`: Fixed `putCluster` to locate the previous base cell (stepping back past `Rune == 0` continuations if needed) and append zero-width characters to its `Text` field.
- `vt/split_test.go`: Created tests to reproduce the failing cases with combining marks, ZWJ sequences, and variation selectors, all of which now pass.

## Tests
`go test ./vt -run 'Test(B9Replay|ZWJSequenceSplit|VariationSelectorSplit|WrapCombine|ExtendsPreviousCell)' -v`
`go test ./...`
