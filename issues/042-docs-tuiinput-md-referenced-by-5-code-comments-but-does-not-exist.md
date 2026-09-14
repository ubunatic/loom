# 042 — `docs/TuiInput.md` referenced by 5 code comments but does not exist

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Documentation
**Related**: `pane.go`, `event.go`, `loom_test.go`

---

## 1. Problem & Motivation

Found while running `/evergreen` (2026-09-14): five code comments across
the core package reference `docs/TuiInput.md` by section number, but the
file does not exist anywhere in the repository (`find`/`ls docs/` come up
empty; not in git history under any path).

```
$ grep -rn "docs/TuiInput" --include=*.go .
./event.go:47:// See docs/TuiInput.md §3.
./loom_test.go:166:// Both prefixes must decode to the same key names. See docs/TuiInput.md §3.
./pane.go:22:// See docs/TuiInput.md for the constraints that govern /dev/tty usage,
./pane.go:78:// See docs/TuiInput.md §1.
./pane.go:363:	// layer (docs/TuiInput.md §2), so a signal alone cannot wake a blocking read —
```

`pane.go`'s package doc comment even summarizes what it should cover:
"the constraints that govern `/dev/tty` usage, EINTR handling, and ZSH
widget compatibility" — so this isn't a stale rename, it looks like a
doc that was planned/referenced but never actually written, or was lost
before this repo's current history began.

This is a real gap for anyone (human or agent) who follows one of these
references expecting an explanation of raw-mode `/dev/tty` handling,
EINTR retry behavior around blocking reads, or the two-prefix key-decode
compatibility `loom_test.go` asserts — they land on a 404 with no
fallback.

## 2. Scope

**In scope:**
- Either (a) write `docs/TuiInput.md` covering the three things the
  existing comments point at (§1: `/dev/tty` usage constraints and why
  `Pane.New` always opens it directly rather than `os.Stdin`; §2: the
  poll-then-read pattern that lets `Close` interrupt a blocking `Read` on
  a detached tty fd, referenced at `pane.go:363` and mirrored in
  `examples/treemap/main.go`'s `watchForQuitKey`; §3: the key-decode
  compatibility contract `event.go`/`loom_test.go` reference), sourcing
  content from the actual current behavior in `pane.go`/`event.go`
  rather than guessing, or (b) if this content already lives elsewhere
  under a different name, update the 5 comments to point at the correct
  file instead.
- A quick `grep -rn "docs/[A-Za-z]*\.md" --include=*.go .` sweep for any
  other phantom doc references while in the area, so this doesn't need a
  second ticket if there's a sibling gap.

**Out of scope:**
- Any behavior change to `pane.go`/`event.go` — this is a documentation-
  only gap; the code itself is not in question.

## 3. Acceptance Criteria

- `docs/TuiInput.md` exists and each of the 5 referenced sections
  (`§1`, `§2`, `§3`) has real content matching what the referencing
  comment claims it explains, OR every reference is corrected to point
  at wherever this content actually lives.
- `docs/README.md`'s hand-maintained doc list includes the new file (if
  created).
- `gofmt -l`/`go vet`/`go test ./...` stay clean (comment-only or new-file
  changes shouldn't touch behavior).

## 4. Verification Guidance

Re-run `grep -rn "docs/TuiInput" --include=*.go .` after the fix and
confirm every hit's referenced section number actually resolves to real
content in the target file.
