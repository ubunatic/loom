---
title: Session Token Usage and Cost Analysis
date: 2026-09-20
scope: Local Codex and Harnez telemetry for the 2026-09-20 session
status: Complete
---

# Session Token Usage and Cost Analysis

## Context and executive summary

This study records the local telemetry visible on 2026-09-20 at 17:19 CEST. It covers the seven Codex agent records in `~/.harnez/agents`, Harnez's unified usage snapshot, Harnez tool-call statistics, and Codex rollout logs. No code files were changed.

The seven records contain 2,603,073 turn tokens. Of these, 2,276,864 were reported as cached and 326,209 as uncached: an aggregate cache-hit rate of 87.47%. This is strong cache reuse, but the measurement is a local Codex/Harnez counter rather than a provider invoice. Harnez's provider quota windows were stale (last refreshed 2026-09-03), so they are not evidence of today's quota consumption.

## Exact discovery and extraction commands

The following commands were used. Paths were read-only; `jq` output was reduced to identifiers, models, timestamps, and token counts.

```sh
sed -n '1,240p' /home/uwe/.codex/skills/story/SKILL.md
rg --files docs/studies docs
sed -n '1,100p' docs/README.md
sed -n '1,70p' docs/studies/2026-09-20-feature-gap-analysis-split.md
command -v harnez
harnez --help
harnez usage --help
harnez stats --help
rg --files ~/.harnez/agents
rg -n -i "token|cache|usage|input|output|prompt" ~/.harnez/agents ~/.harnez
for f in ~/.harnez/agents/*.json; do
  jq -r '[.agent_id // .id // "?", .model // "?", .started_at // .created_at // "?", .tokens_turn // 0, .cached_tokens // 0, .tokens_cumulative // 0] | @tsv' "$f"
done
harnez usage --json --offline
harnez stats --json --auto --overhead
rg --files -g '*.jsonl' -g '*.log' ~/.harnez ~/.codex
stat -c '%y %s %n' ~/.harnez/agents/*.json ~/.harnez/sessions/*.usage.json
jq -s 'map({id:(.agent_id // .id),model,turn:(.tokens_turn//0),cached:(.cached_tokens//0),uncached:((.tokens_turn//0)-(.cached_tokens//0))}) as $a | {agents:$a,total_turn:($a|map(.turn)|add),total_cached:($a|map(.cached)|add),total_uncached:($a|map(.uncached)|add),cache_rate:(($a|map(.cached)|add)/($a|map(.turn)|add)*100)}' ~/.harnez/agents/*.json
for f in ~/.harnez/sessions/0753e543c40b96d6.usage.json ~/.harnez/sessions/24f607e82a367e01.usage.json; do jq . "$f"; done
```

The session-log inventory identified the relevant rollout directory with:

```sh
rg --files ~/.codex/sessions/2026/09/20
```

The `.usage.json` files are Harnez command counters, not provider token ledgers. Their inspection was used to establish this limitation and to corroborate activity timestamps; token totals came from the agent JSON records.

## Per-agent token breakdown

| Agent record | Model | Turn tokens | Cached | Uncached | Cache hit |
|---|---|---:|---:|---:|---:|
| `01a0bee7…0200d57f2004` | gpt-5.6-luna | 324,620 | 264,192 | 60,428 | 81.39% |
| `01a0beea…3e856837f73c` | gpt-5.6-luna | 187,407 | 167,424 | 19,983 | 89.33% |
| `01a0beeb…30b730cdd163` | gpt-5.6-luna | 242,157 | 203,776 | 38,381 | 84.15% |
| `01a0bf36…cc73a43300a6` | gpt-5.6-sol | 663,768 | 604,032 | 59,736 | 91.00% |
| `01a0bf38…985fc4743c0b` | gpt-5.6-sol | 297,451 | 264,448 | 33,003 | 88.92% |
| `01a0bf40…24c30a08c67e` | gpt-5.6-sol | 633,622 | 574,464 | 59,158 | 90.66% |
| `01a0bf5d…2afd984014dd` | gpt-5.6-sol | 254,048 | 198,528 | 55,520 | 78.15% |
| **Total** | — | **2,603,073** | **2,276,864** | **326,209** | **87.47%** |

The calculation is `uncached = turn_tokens - cached_tokens`; the percentage is `cached / turn_tokens × 100`.

## Cached versus uncached view

| Class | Tokens | Share of recorded turn tokens |
|---|---:|---:|
| Cached | 2,276,864 | 87.47% |
| Uncached | 326,209 | 12.53% |
| Total | 2,603,073 | 100.00% |

Harnez `stats --auto --overhead` independently reported 309 tool calls, 14 failures, 1,993,908 raw bytes for the AGY tool telemetry rows, and zero distillation savings. Its feedback overhead estimate was 670 call tokens plus 796 instruction tokens. Those are tool-telemetry estimates, not model prompt tokens, and are not added to the table above.

## Cost economics

The local records report `$0` because these calls are subscription-routed. A counterfactual API estimate is useful for scale, but it cannot be an invoice: the records do not separate input, output, reasoning, and cache-write tokens, and `gpt-5.6-luna`/`gpt-5.6-sol` are local harness model labels.

For a transparent reference case, apply the published GPT-5 rates of $1.25 per million uncached input tokens and $0.125 per million cached input tokens (the OpenAI model page lists those rates: <https://developers.openai.com/api/docs/models/gpt-5>). Treating every recorded turn token as input gives:

| Component | Formula | Estimated cost |
|---|---|---:|
| Uncached | 326,209 / 1,000,000 × $1.25 | $0.408 |
| Cached | 2,276,864 / 1,000,000 × $0.125 | $0.285 |
| **Total prompt-only equivalent** | — | **$0.693** |
| No-cache counterfactual | 2,603,073 / 1,000,000 × $1.25 | $3.254 |
| Approximate cache saving | $3.254 − $0.693 | **$2.561 (78.7%)** |

The observed account is a Codex Plus subscription. Against a nominal $20/month Plus price, this one measured workload would equal about 3.5% of the monthly fee at the reference API rate; the subscription breaks even against this workload after roughly 28.9 equally sized sessions. That is an economic comparison, not a claim that the subscription grants a fixed number of API tokens. Subscription value is instead governed by access, rate limits, and included usage. The stale local quota snapshot shows Codex Plus, but its 1% five-hour and 52% weekly values are stale and must not be used as today's consumption.

Anthropic cache economics are structurally similar: current documentation lists cache hits at 0.1× base input pricing for supported Claude models (<https://docs.anthropic.com/en/docs/about-claude/pricing>). It is not applied to this Codex total because no Claude token counter for today's session was available.

## What the data says

- Cache reuse is high across every record: 78.15%–91.00%; the last `gpt-5.6-sol` record has the lowest rate and is the clearest optimization target.
- The three Luna records account for 754,184 tokens and 84.29% cache reuse; the four Sol records account for 1,848,889 tokens and 88.76% reuse.
- The local tool layer recorded no distillation savings. That does not mean the model prompt was not cached; it means Harnez did not record a distillation event for this session.

## Methodological notes and telemetry limitations

1. The unit of analysis is the seven agent JSON records present at inspection time, not a provider billing export. Records are snapshots and may represent agent lifetimes that overlap the session boundary.
2. `tokens_turn` and `cached_tokens` were treated as comparable counters because that is the schema exposed by the local records. Their exact provider billing semantics are not documented in the files.
3. Output-token counts are absent from the agent records used for the per-agent table. The API estimate therefore prices all recorded tokens as input; it excludes output pricing and any cache-write surcharge.
4. `harnez usage --json --offline` returned stale provider snapshots last refreshed on 2026-09-03. The command is valuable for schema and account/tier discovery, but not for today's quota burn.
5. `.usage.json` files record Harnez CLI verbs and timestamps, not full model request/response usage. Rollout JSONL files can contain sensitive prompts and were inventory-listed but not reproduced in the report.
6. The snapshot was taken at 17:19 CEST. Any later same-day activity is outside the measurement window.

## File and diff summary

Created this study only. No source, test, configuration, or other code file was modified. The study index entry was added to `docs/README.md` using `harnez index`.
