# Agentic Work Retrospective — Issue #038 (popup.go UTF-8 truncation)

Session self-audit of what went wrong while fixing "Fix multi-byte UTF-8 string
truncation in popup title and borders."

## What went wrong

### 1. A single edit became a 45+ call loop of failed attempts
The core problem was trivial: replace one ~15-line block in popup.go. Instead of
applying it once, dozens of tool calls were spent on repeated drafts. This violated
the "one main action per step" discipline and the Quota-1 spirit (make real progress
per turn).

### 2. Quote/escaping hell across three layers
Go source strings (containing `" " + p.Title + " "` and box-drawing chars like ┌)
were nested inside Python string literals inside bash heredocs inside IPython cells.
Each layer re-escaped the others. Nearly every draft produced a Python SyntaxError,
an unterminated heredoc ("warning: here-document delimited by end-of-file"), or
silently corrupted Go code.

The right move (discovered late): read the exact old slice from the file itself
(`text[a:b]`) and pass it verbatim to the edit skill, avoiding all re-typing.

### 3. Junk "guard" expressions injected into drafts
Multiple drafts contained nonsense such as `if False else None`,
`.replace('||','&').replace('&',' ||')`, `'test'` placeholders, wrong-arity calls
(`c.Write(x+1+fit, cluster, style)` with 3 args instead of the correct 4), and
mangled identifiers (`pTitle`, `fi t`). These were artifacts of trying to "protect"
string concatenation in the kernel; they made each draft wrong before it ran.
Drafts should be plain lines, not clever one-liners with escape guards.

### 4. Repeated near-identical failing cells verbatim
After the first few failures, essentially the same broken heredoc patterns were
re-run many times (gen_block.py / splice.py / base64 attempts) without changing
strategy. Repeating a failed approach is churn, not iteration. A working pattern
already existed (the successful splice using /tmp/exact_old.txt + /tmp/newblock.go),
but it was abandoned for increasingly exotic mechanisms (base64, nested replaces).

### 5. Scratch files written into the repo root
Failed drafts dropped gen.py, gen_block.py, splice.py into /work. They had to be
cleaned up at the end. Scratch work belongs under /tmp, never the working tree.

### 6. A real bug in own code caught only by re-reading
An intermediate version wrote title clusters to canvas via c.Write, then ran the
`for i, cell := range row { c.Set(x+i, y, cell) }` loop afterward — which would
have overwritten the title with dash cells again (reintroducing the very bug being
fixed). Correct ordering: fill `row`, set corners, write title last (or write title
into `row` cells before the Set loop). Fixed after inspecting spliced output rather
than reasoning about ordering before applying.

## What actually worked
- Reading exact old text from disk and splicing via the edit skill (no re-typing of target text).
- Writing the new block once to a /tmp heredoc file cleanly, reading it back, then splicing.
- Verifying with `go build ./...`, then `git status --porcelain` for a clean tree.

## Net result
The fix is correct and builds (rc 0), but cost ~45 tool calls where ~4 should have
sufficed.

## Lesson (proposed harness memory)
For targeted single-block edits: read the exact old slice from disk (`text[a:b]`),
draft plain lines in /tmp (one heredoc, no escape gymnastics), splice once with the
edit skill, build once. No kernel string concatenation guards, no base64, no scratch
files in the repo root. Reason about call ordering/arity before applying, not after.
