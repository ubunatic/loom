# Reviewing Loom Example Apps with `lmcoder`: Prompt vs. Agent Modes

**Date**: 2026-09-17  
**Author**: Uwe Jugel & AI Assistant  
**Subject**: Code review of Loom example apps, ergonomics of `lmcoder` toolchain, and local/remote LLM serving analysis.

---

## 1. Executive Summary

All 6 Loom example applications (`filebrowser`, `monitor`, `splash`, `split`, `tabs`, `treemap`) were evaluated using **`lmcoder`** across two primary workflows:
1. **Direct Single-Shot Prompting** (`lmcoder prompt -n -f ...`) on both local APU (`qwen3-4b-instruct-2507-q4`) and remote dedicated compute node `x600` (`qwen3.8-27b-instruct-q4`).
2. **Autonomous Sandboxed Agent Execution** (`lmcoder agent -- -p ...` running `prime-agent` in the `agent-canary-multi` container).

---

## 2. How `lmcoder` Was Used & Operational Discoveries

```
                  ┌─────────────────────────────────────────────────────────┐
                  │                 lmcoder Workflows                       │
                  └────────────┬──────────────────────────────┬─────────────┘
                               │                              │
                ┌──────────────▼──────────────┐ ┌─────────────▼─────────────┐
                │        Direct Prompt        │ │      Container Agent      │
                │  lmcoder prompt -n -f ...   │ │  lmcoder agent -- -p ...  │
                └──────────────┬──────────────┘ └─────────────┬─────────────┘
                               │                              │
                 ┌─────────────┴─────────────┐                │
                 ▼                           ▼                ▼
     ┌───────────────────────┐   ┌───────────────────────┐ ┌──────────────────────┐
     │  Local Node (t14)     │   │  Remote Node (x600)   │ │ Podman Sandboxed     │
     │  Port 8734 (Qwen3-4B) │   │  SSH Port 8736 (27.8B)│ │ Container Runtime    │
     └───────────────────────┘   └───────────────────────┘ └──────────────────────┘
```

### Discovery A: Direct Prompt Session Isolation (`-n`)
- `lmcoder prompt` by default resumes previous terminal conversation sessions.
- Explicitly passing `-n` / `--new` ensured each review ran in a clean, fresh session without context bleeding from prior runs.

### Discovery B: Multi-File Ingestion via Repeated `-f` Flags
- Using `-f <path>` allows feeding multiple project files directly into the LLM prompt.
- Injecting exact Go source files (`main.go`, widget implementations, spec files, unit tests) enabled high-precision single-shot code reviews in **~1m 15s to 1m 35s**.

### Discovery C: Remote SSH Prompting & Port Forwarding Workaround (`x600`)
- Running `lmcoder prompt --host x600` returned:
  ```text
  Error: remote host prompting over SSH is not yet implemented (use local or --host codex)
  ```
  *(tracked as open issue `100` in `lmcoder`)*.
- **Workaround Executed**:
  1. Verified remote server health on `x600` via `lmcoder status --host x600`.
  2. Established an SSH tunnel to the remote llama-server: `ssh -f -N -L 8736:127.0.0.1:8734 x600`.
  3. Targeted the remote 27.8B model directly using `lmcoder prompt --port 8736 -n ...`, achieving successful remote evaluation.

### Discovery D: Live Agent Inspection via Podman & Session Transcripts
- When `lmcoder agent` runs, it mounts the workspace read-write into a hardened Podman container (`agent-canary-multi:latest`) executing `prime-agent`.
- **Inspection Technique**:
  - Live container processes: `podman ps`
  - Internal daemon status: `podman exec <cid> prime-agent status`
  - Raw JSONL interaction logs: `/tmp/prime-agent-sessions/<session-id>.jsonl`
- Inspecting this transcript revealed the exact tool calls, Python tracebacks, and context token accumulation of the containerized agent in real time.

---

## 3. Findings Across All Example Applications

| Application | Key Architectural Strengths | Go Idioms & Quality | Opportunities / Identified Gaps |
| :--- | :--- | :--- | :--- |
| **`examples/filebrowser`** | Clean separation between CLI entry, business logic, and UI (`Frame`, `Choice`, `View`). Responsive split-pane layout. | Idiomatic error handling and propagation; isolated temporary directories used in `browser_test.go`. | Optimize directory scanning on massive directories; add path traversal sanitation; improve keyboard accessibility. |
| **`examples/monitor`** | Spec-driven design (`monitor.yaml` & `watch.yaml`). Dual-mode execution (one-shot snapshot vs continuous `--watch`). | Immutable snapshot copies prevent UI data races; robust terminal dimension detection and fallbacks. | Modularize high-frequency ticker updates; add runtime YAML schema validation; expand test coverage for malformed `/proc` inputs. |
| **`examples/splash`** | View-controller pattern; context-driven lifecycle with `Done()` channel; Braille animated spinner and unicode status pills. | Proper resource cleanup with `defer pane.Close()`; clean error returns to stderr. | Add explicit validation bounds for `--width`/`--height` CLI flags; allow customizable animation duration; add tests for failure paths. |
| **`examples/split`** | Minimal, focused demo of independent dual-pane scrolling and focus state (`Dynamic: true`). Visual focus indicator (`▶`). | Clean and idiomatic; strictly satisfies Loom's `Run(args []string) error` widget contract. | Add keyboard tab navigation between panes; handle terminal resize transitions explicitly. |
| **`examples/tabs`** | Multi-widget container hosting `View`, `Choice`, and `Table`. Intuitive `SwitchKey: "ctrl-t"` keyboard binding. | Clean helper functions (`viewLines`, `choiceItems`, `tableData`) decouple mock data from widget structure. | Add validation on tab definitions; improve user error messaging; implement unit tests for tab switching logic. |
| **`examples/treemap`** | Squarified treemap rendering engine with hierarchical box partitioning, color palette styling, and interactive selection. | Modularized into `input.go`, `process.go`, `style.go`, and `terminal.go`. | Handle edge cases for zero-weight nodes and extreme terminal aspect ratios; handle deep directory hierarchies efficiently. |

---

## 4. Workflow Comparison: `lmcoder prompt` vs `lmcoder agent`

| Metric / Attribute | `lmcoder prompt -n -f <files>` | `lmcoder agent` (`prime-agent` in Podman) |
| :--- | :--- | :--- |
| **Turnaround Time** | **~1m 15s – 1m 35s** | **~5m 00s – 13m 00s** (or timeout) |
| **Ease of Hand-off** | **Very High**: Single CLI command passes exact files directly into context. | **Moderate**: Sandboxed environment setup is automated, but tool execution is prone to multi-turn delays. |
| **Accuracy & Grounding** | **High**: Cites exact struct names, methods, widget lifecycles, and test cases. | **Mixed**: Small 4B local model struggles with path resolution and Python tool syntax inside the container. |
| **Context Efficiency** | **Optimized**: Context is spent 100% on source code and prompt instructions. | **Degraded**: Significant context consumed by tool call syntax, error tracebacks, and directory listings. |
| **Best Suited For** | Code review, static analysis, spec checking, and multi-file code explanation. | Open-ended exploratory debugging, autonomous multi-step refactoring, or iterative testing. |

---

## 5. Server Capacity & Context Headroom

Both available servers have ample headroom to scale up the context window (`ctx-size`) to support larger multi-file prompts or deeper agent reasoning traces:

| Host / Machine | Hardware & Memory Budget | Active Model | Model Weights | Active Context (`ctx-size`) | Current Memory Used | Free Memory Headroom | Recommended Context Bump |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Local (`t14`)** | AMD Cezanne APU (**20.3 GiB** budget) | `qwen3-4b-instruct-2507-q4` | 2.3 GiB | 16,384 | 5.2 GiB | **~15.1 GiB free** | **32,768 – 65,536 tokens** |
| **Remote (`x600`)** | AMD Phoenix APU (**34.5 GiB** budget) | `qwen3.8-27b-instruct-q4` | 15.3 GiB | 65,536 | 15.5 GiB | **~19.0 GiB free** | **65,536 – 131,072 tokens** |

---

## 6. Issues Discovered and Logged in `../lmcoder`

Two actionable issues were identified during execution and committed to the `lmcoder` tracker:

1. **Issue 102**: `lmcoder prompt exits with code 0 on context size exceeded or empty API response` (P1 / Bug)
   - *Problem*: When the backend returns `Context size has been exceeded` or an empty response, `lmcoder prompt` silently exited with code 0 instead of propagating the error to `stderr` and exiting non-zero.
2. **Issue 103**: `lmcoder prompt pre-flight context size check and attachment token budget warning` (P2 / Enhancement)
   - *Problem*: Attaching multiple files via `-f` easily overruns model context without warning; `lmcoder prompt` should estimate token count and warn before sending.
