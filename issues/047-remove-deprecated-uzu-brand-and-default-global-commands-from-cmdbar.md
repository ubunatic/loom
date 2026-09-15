# 047 — Remove deprecated uzu brand and default global commands from loom.cmdBar

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Cleanup / Refactor
**Related**: `cmd.go`, `cmd_test.go`, `issues/021-cmd-help-tty-blocking-in-tests.md`

---

## 1. Problem & Motivation

The `uzu` tool brand is deprecated and gone. However, `loom/cmd.go` still contains hardcoded defaults referencing `uzu` in `newCmdBar()`:

```go
func newCmdBar() *cmdBar {
	return &cmdBar{
		global: []Cmd{
			{Name: "help", Title: "show help for current view"},
			{Name: "back", Title: "go one level back"},
			{Name: "home", Title: "go to uzu main menu"},
		},
	}
}
```

When users or downstream consumers open `:help` or trigger command completion, `:home` appears with the outdated title `"go to uzu main menu"`.

---

## 2. Proposed Changes

1. **Remove or Generalize Default Global Commands**:
   - Remove `{Name: "home", Title: "go to uzu main menu"}` from default global commands.
   - Keep generic defaults like `:help` ("show help for current view") and `:back` ("go one level back") or make global commands fully configurable by the host application.
2. **Audit References**:
   - Clean up any remaining references or comments mentioning `uzu` across `cmd.go` and its tests.

---

## 3. Acceptance Criteria

- `:help` in `loom.View` does not display `"go to uzu main menu"`.
- No hardcoded `uzu` references remain in active runtime code.
- All tests in `loom` continue to pass (`go test ./...`).
