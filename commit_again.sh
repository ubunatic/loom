go test ./internal/apitest -update
git add reports/jules/390-impl.md internal/session/session.go internal/session/session_pty_test.go testdata/api/specdefaults.txt internal/session/tab_label* spec/defaults.yaml spec/schemas/defaults.schema.json specdefaults/defaults.go
git commit --amend --no-edit
