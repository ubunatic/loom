## Verdict
done, VT now answers DECRQM, Kitty `?u`, and XTVERSION queries directly from its state as specified, and a test verifies the behavior.

## Changes
vt/vt.go
vt/vt_test.go
internal/session/session_pty_test.go

## Tests
```
go test ./vt -run TestVTAnswersDECRQM
go test ./vt -run TestVTAnswersKittyKeyboardQuery
go test ./vt -run TestVTAnswersXTVERSION
go test ./internal/session -run TestSessionAnswersDECRQM
```
