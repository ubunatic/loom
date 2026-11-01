## Verdict
done, fixed hardcoded `pgup` and `pgdown` shortcuts by using `loom.SpeccedDefaults.View.Bindings["page_up"]` and `loom.SpeccedDefaults.View.Bindings["page_down"]`.

## Changes
cmd/loom/main.go
cmd/loom/view_pty_test.go

## Tests
- `go test -v ./cmd/loom -run TestANSIViewPTYSpaceAndCtrlFPagesDown` (PASS)
- `go test -v ./cmd/loom -run TestANSIViewAcceptsPageDownAliases` (PASS)
