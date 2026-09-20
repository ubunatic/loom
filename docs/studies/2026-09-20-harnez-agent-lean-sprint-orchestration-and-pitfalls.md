# Study: Practical Orchestration and Pitfall Analysis of `harnez agent` in Lean Sprints

**Date**: 2026-09-20  
**Status**: Complete  
**Related Tickets**: Issues 036, 038, 057, 058, 059, 085, 094  

---

## 1. Context & Motivation

The `harnez agent` CLI provides a lightweight mechanism to spawn, resume, and teardown autonomous, low-cost coding subagents (such as `codex:luna:low` and `gemini3.7flash:low`) working against a target directory. During the 2026-09-20 Lean Sprint, the Host Orchestrator delegated 7 discrete milestones/tickets entirely to fresh `codex:luna:low` workers.

This study records empirical findings on execution efficiency, structural strengths, failure modes, background task timeouts, lock contentions, and practical orchestrator guidelines.

---

## 2. Workflow & Lifecycle

The standard handoff cycle followed the 5-step Lean Sprint protocol:

```text
[Host Orchestrator]
       │
       ├── 1. Spawn: harnez agent start codex:luna:low --name "dev-XXX" "<Milestone-1 Prompt>"
       │      └─ Returns: UUID, Reconnect banner, initial response
       │
       ├── 2. Review: Diff-only inspection (git log --stat, git diff HEAD~1, go test)
       │
       ├── 3. Refine (if needed): harnez agent resume dev-XXX "<Milestone-2 Pre-Work>"
       │      └─ Returns: Synchronous result from existing session context
       │
       └── 4. Teardown: harnez agent delete dev-XXX && harnez issues close -d . XXX
```

### Ticket Execution Log (Sprint Summary)

| Ticket | Agent Name | Work Scope | Milestones / Resumes | Result & Commits |
|---|---|---|---|---|
| **036** | `dev-036` | `KeyEvent.Name/Is/Rune` helpers | 1 milestone | Clean commit (`4a2e3f5`) |
| **038** | `dev-038` | UTF-8 popup title truncation tests | 2 milestones (M1 tests, M2 draw-order fix) | Tests added (`e3b6827`), draw order fixed (`b2bfec2`) |
| **059** | `dev-059` | Standardize 0-based canvas mouse coords | 7 iterative resumes (coord math, clipping, scrollbar) | Pinned 0-based coords across 13 files (`ea0ef79`) |
| **085** | `dev-085` | E2E PTY smoke tests for all 9 apps | 1 milestone | 72-line smoke test suite (`f14416a`) |
| **057** | `dev-057` | `KeyConsumer` child-first routing & quit | 2 milestones (M1 implementation, M2 tests) | Feature (`2887e58`), test suite (`f604066`) |
| **058** | `dev-058` | `PaneRequest` terminal capabilities | 2 milestones (M1 implementation, M2 tests) | Feature (`588c035`), test suite (`bdb3ced`) |
| **094** | `dev-094` | Modal keyboard capture in filebrowser | 2 milestones (M1 fix, M2 signature sync) | Fix (`3ba844b`), test update (`503517d`) |

---

## 3. What Went Well

1. **High Implementation Speed & Focus**:
   - For scoped tasks (such as 036, 085, 058), worker agents delivered complete, idiomatically formatted Go implementations and unit tests within 60–90 seconds.
2. **Context Isolation**:
   - Using dedicated agent names per ticket (`dev-036`, `dev-057`, etc.) avoided cross-task context pollution and token bloat in the orchestrator conversation.
3. **Multi-turn Refinement via `resume`**:
   - When initial implementations missed edge cases (such as the `handleHelpKey` signature change in 094 breaking `pane_internal_test.go`), `harnez agent resume` allowed pinpoint corrections without restarting agent context from scratch.
4. **Adherence to Framework Conventions**:
   - The low-tier model successfully adhered to codebase conventions, using `testpty`, `layout.Rect`, and `KeyConsumer` without hallucinating external dependencies.

---

## 4. Observed Pitfalls & Failure Modes

### 4.1. Background Task Timeouts (`WaitMsBeforeAsync`)
- **Symptom**: When `harnez agent start` or `resume` runs via a background runner tool, default short wait times (e.g. 5–10s) immediately send the command to the background, requiring async task polling. Conversely, setting excessively short timeouts on heavy compilation/PTY test tasks can risk early truncation.
- **Root Cause**: Spawning `codex:luna:low` involves model inference, tool calls inside the subagent (reading files, compiling, executing `go test`), which typically takes **45 to 150 seconds**.
- **Recommended Timeout Strategy**:
  - `start` (new feature / bugfix): `WaitMsBeforeAsync: 240000` to `300000` (4–5 minutes).
  - `resume` (focused test fix / small edit): `WaitMsBeforeAsync: 120000` to `180000` (2–3 minutes).
  - `delete` / cleanup: `WaitMsBeforeAsync: 5000` to `10000` (synchronous execution).

### 4.2. `.git/index.lock` Read-Only Contention
- **Symptom**: In several runs (`dev-038`, `dev-059`, `dev-058`, `dev-094`), the subagent output:
  `fatal: Unable to create '/home/uwe/projects/loom/.git/index.lock': Read-only file system` or permission denied.
- **Root Cause**: Subagent sandbox / permission barriers or concurrent file locks in the sub-environment prevented the agent from running `git commit` directly, leaving files modified but unstaged in the working tree.
- **Orchestrator Mitigation**: The Host Orchestrator must always verify working tree status with `git status` / `git diff`. If the subagent completed the code changes and verified tests green but failed to commit due to lock errors, the host should execute the commit on the worker's behalf and proceed.

### 4.3. Raw Shell Substitution in Prompts
- **Symptom**: When prompt strings sent to `harnez agent start` or `resume` contained unescaped Go syntax (e.g., `func (p *Pane) ...`, `$(...)`, or backticks), the host shell attempted parameter expansion, resulting in errors like:
  `bash: command substitution: line 2: syntax error near unexpected token '('`
- **Root Cause**: Double-interpolation when passing multi-line shell strings with quotes.
- **Mitigation**: Escape internal quotes/backticks carefully or pass prompts via heredocs / clean single quotes when invoking the CLI.

### 4.4. Pre-existing Headless Environment Gaps (`/dev/tty`)
- **Symptom**: Running `go test ./...` in subagents frequently showed failures in root-package tests that expect a controlling `/dev/tty` (e.g. `TestTerminalSize`, `TestPaneStartup`).
- **Impact**: Low-cost subagents occasionally believed the entire test suite was broken and hesitated to commit.
- **Mitigation**: Instruct the subagent explicitly in the milestone prompt to focus on targeted package tests (`go test ./<package>/... -run <TestName>`) and distinguish headless TTY skips from actual code regressions.

### 4.5. Multi-Container Coordinate Cascades (Issue 059)
- **Symptom**: Ticket 059 required 7 resume cycles because coordinate conversions between `paintClipped` (sub-canvas 0-based) and `Stack`/`Tabs` (absolute 0-based) were initially misdiagnosed as pure test literal offsets.
- **Lesson**: For multi-layer structural invariants, the Host Orchestrator should provide concise architectural context (e.g., distinguishing clipped canvas vs unclipped canvas behavior) in the resume prompt rather than letting the agent guess coordinates by trial-and-error.

---

## 5. Summary & Recommendations

1. **Use Scoped Prompts**: Clearly specify target files, expected tests, and acceptance criteria in Milestone 1.
2. **Configure 2–4 Minute Timeouts**: Set `WaitMsBeforeAsync` between 120s and 300s for agent generation and test runs.
3. **Inspect Diffs from Host**: Always inspect `git diff` / `git status` after agent execution to catch uncommitted changes caused by sandbox git lock contention.
4. **Clean Teardown**: Always call `harnez agent delete <name>` upon milestone completion to ensure zero zombie sessions.
