# 405 — Session drops outer replies to child queries the VT does not answer (DECRQM, kitty ?u, XTVERSION)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:

**Related**: 391 (869c304, 7b61498)

---

/goal Decide per query type whether the session VT answers it or forwards the outer reply to the child, implement that, and pin it with session PTY tests; or stop and report when blocked on a user decision.

## 1. Problem & Motivation
Since 391, `loom session` drops every CSI reply arriving on stdin (`KeyEvent.Report`, `internal/session/session.go` ~1072) and the VT answers only DSR 6n/5n and DA1/DA2 (`vt/vt.go`). Child queries pass raw to the outer terminal, so other queries (DECRQM `ESC[?2026$p`, kitty keyboard `ESC[?u`, XTVERSION `ESC[>0q`) get an outer reply that is now dropped. Before 391 those replies reached the child (with leaked bytes in search/prefix mode). Apps that wait for such replies now time out or fall back.

## 2. Plan
1. List the queries common TUIs send at startup (DECRQM, kitty ?u, XTVERSION, OSC 10/11 colors).
2. For each: answer from the VT (state it owns) or forward the outer reply to the active tab's child.
3. Session PTY test per choice: child sends the query, gets exactly one well-formed reply.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| implement | jules-implement-405 | 16101408603233734506 | pending | asked back twice; final "submit a draft PR" reply sent 2026-11-03, replies: 2; adopt with `harnez agent resume --model jules --id 16101408603233734506 wait` |
| implement | jules-implement-405a | 13296704772212350363 | pending | PAUSED after harnez auto-reply (replies: 1); adopt with `harnez agent resume --model jules --id 13296704772212350363 wait` |
| implement | jules-implement-405b | 14704705202975718091 | pending | running; adopt with `harnez agent resume --model jules --id 14704705202975718091 wait` |
| implement | jules-implement-405c | 16418329315455329292 | pending | running; adopt with `harnez agent resume --model jules --id 16418329315455329292 wait` |
