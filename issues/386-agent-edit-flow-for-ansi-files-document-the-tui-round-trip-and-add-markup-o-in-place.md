# 386 — Agent edit flow for .ansi files: document the .tui round trip and add markup -o / --in-place

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 308 (.tui format and `loom markup`), 309 (box markup)

---

## 1. Problem & Motivation

AI agents are good at drafting `.ansi` screens directly, but bad at *editing* them afterwards.
In tokburn (ticket 010, game screen redesign, 2026-10-26) a Claude agent (Opus) had to fix one
card in a finished 80×24 design: split the text `reveal tier` on a cyan joker card into two
lines. Nobody told it to use `.tui`; it discovered the round trip through `loom --help` and it
worked well. This ticket records that story so loom teaches the flow to agents and removes the
remaining friction.

/goal Agents find and use the `.tui` round trip for editing `.ansi` files without shell
redirects: the flow is documented where agents look (help text with examples, docs), and
`loom markup` can write its output to a file (`-o`) and convert in place (`--in-place`). Verified
by tests and an updated help text. Stop and report when blocked on an owner decision. Circuit
breaker: on a loop or no progress, clear the goal and write a failure report here.

## 2. What went wrong: editing .ansi directly

- **Escape characters are invisible to the agent.** The file-read tool shows `[0;22;38;5;16m...`
  with the ESC byte hidden, and the agent cannot reliably type ESC into an exact-match edit. It
  can only anchor edits on text *between* escapes.
- **Identical lines cannot be targeted.** The blank row under the card text was byte-for-byte
  identical to the card's top padding row, and every distinguishing context crosses an ESC. The
  edit tool reported "2 matches" and there was no way to pick the right one.
- **Silent geometry damage.** The first, partly successful edit lost the trailing spaces after
  `reveal`, shifting the card by 3 columns. Nothing flagged it until a later read.
- **Workarounds the agents reached for are worse.** Other designers in the same session wrote
  Python generator scripts (one via a Bash heredoc, against the owner's "edit with Edit/Write"
  rule) instead of editing the asset.

## 3. What worked: discovery and the .tui round trip

Discovery took two commands:

1. `loom --help` listed `markup — Convert between .tui markup and ANSI`.
2. `loom markup --help` gave one example line:
   `loom markup art.tui > art.ansi && loom eval art.ansi`. That example alone taught the whole
   flow. **Help screens with concrete examples are how agents learn commands** — keep them.

The round trip:

1. `loom markup design.ansi > scratch/design.tui` — every style becomes readable text
   (`[bold color(16) on color(44)]  JOKER ◆  [/]`), every space visible.
2. Edit the `.tui` with the normal edit tool; including the neighbouring line made the match
   unique.
3. `loom markup scratch/design.tui > scratch/design.ansi`.
4. `loom eval` (all rows 80 wide) and `loom diff old.ansi new.ansi` — it reported exactly the 10
   changed joker cells, a precise check instead of a guess.
5. Replace the original.

The round trip was lossless, and `loom diff` turned "I think it's right" into proof.

## 4. Remaining friction

- **No output flag.** `loom markup` only writes to stdout, so every step needs a shell redirect
  (`>`). The owner's tooling flags redirects as Bash file edits (a hook fired twice), and the
  final copy into the repo needed `cp`. For agentic coding in general this matters a lot: a
  command that writes its result to a named file (`-o FILE`) or converts in place
  (`--in-place`) lets agents stay inside their edit-tool rules and makes the step reviewable.
- **Nothing points agents to the flow first.** The agent found it only after getting stuck.

## 5. Proposed changes

1. `loom markup -o FILE` and `loom markup --in-place` (e.g. `x.ansi` ↔ `x.tui` next to it, or
   rewrite the given file), with tests.
2. Document the agent edit flow (draft `.ansi` directly; when an edit gets stuck, convert to
   `.tui`, edit, convert back, check with `loom eval` + `loom diff`) in the docs and as an
   example in `loom markup --help`.
3. Optional: let `loom diff` accept a `.tui` file on either side.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| implement | jules-implement-386 | 39bf6d93-0313-48e8-9f10-0da2f8bf4454 | - | running |
