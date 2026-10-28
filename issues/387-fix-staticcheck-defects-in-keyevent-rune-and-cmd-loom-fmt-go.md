# 387 — Fix staticcheck defects in KeyEvent.Rune and cmd/loom/fmt.go

**Status**: Closed — fixed KeyEvent.Rune early termination and direct byte write in fmt
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:

---

## 1. Problem & Motivation
`gokay check --all` flags staticcheck defects in `loom`:
1. `event.go:83:3: SA4004: the surrounding loop is unconditionally terminated (staticcheck)`:
   `KeyEvent.Rune()` unconditionally terminates on the first rune in `e.Text`, failing to find subsequent printable runes.
2. `cmd/loom/fmt.go:29:13: SA6006: use io.Writer.Write instead of converting from []byte to string to use io.WriteString (staticcheck)`:
   Converts byte slice `formatted` to string solely to call `io.WriteString`.

## 2. Technical Specification / Findings
1. In `event.go`:
   Remove early `return 0` inside the loop in `KeyEvent.Rune()` so it accurately scans for the first printable rune.
2. In `cmd/loom/fmt.go`:
   Replace `_, err = io.WriteString(cmd.OutOrStdout(), string(formatted))` with `_, err = cmd.OutOrStdout().Write(formatted)`.

## 3. Implementation & Verification Plan
- Edit `event.go` and `cmd/loom/fmt.go`.
- Run tests in `loom` (`go test ./...` or unit tests).
- Run `gokay check` on changed files in `loom`.
