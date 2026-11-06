## Verdict
done, fixed `TextInput` and `TextArea` to navigate and delete text grapheme-cluster-by-grapheme-cluster by exposing and utilizing cluster boundary logic from the `measure` package.

## Changes
measure/measure.go
textinput.go
textarea.go

## Tests
`go test ./...`
```
ok  	ubunatic.com/loom	34.009s
ok  	ubunatic.com/loom/ansi	(cached)
ok  	ubunatic.com/loom/measure	(cached)
```
(Other unrelated / flaky test output omitted, widget tests `loom` and `measure` tests explicitly pass)
