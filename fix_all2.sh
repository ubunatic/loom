python3 fix_pty_test.py
UPDATE_GOLDEN=1 UPDATE_SNAPSHOTS=1 GENERATE_DESIGN_SCREENSHOTS=1 go test ./internal/session -run TestSessionBarGoldens
git add reports/jules/390-impl.md docs/progress/373/session-bar-one.ansi docs/progress/373/session-bar-three.ansi internal/session/session.go internal/session/session_pty_test.go testdata/api/specdefaults.txt internal/session/tab_label* spec/defaults.yaml spec/schemas/defaults.schema.json specdefaults/defaults.go
git commit -m "feat(session): show foreground command in tab labels"
