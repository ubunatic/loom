# 178 — Research: compare loom to ncurses and list feature gaps

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Research
**Related**: issues/180 (roadmap), issues/177-179

---

## 1. Problem & Motivation
ncurses defines the classic terminal capability baseline (terminfo, panels, forms, menus, color pairs, mouse, input modes). We want a list of features ncurses offers that loom lacks, so the roadmap (180) can close the gaps that matter.

## 2. Scope & Rules
- Stay true to loom's design (read AGENTS.md, docs/Widgets.md, README, `loom widgets`): a gap is only a gap if it fits loom's model. Features that contradict the design go in a "Deliberately not adopted" list with one-line reasons.
- Check live code before claiming a gap; cite loom file/symbol for what exists.
- Cross-reference existing open tickets (`harnez find -d . issues -a status:open`, e.g. 155 modal, 158 keymap, 167-173 widgets) instead of re-listing them as new.
- Output: append a `## Findings` section to this ticket: a table `Feature | <other> | loom today | Gap? | Fits design? | Existing ticket | Size (S/M/L)`, then the not-adopted list. No code changes.

## 3. Implementation & Verification Plan
/goal The Findings section lists every relevant ncurses feature with loom's status and a design-fit verdict, committed as `docs(issues): findings for 178`. Stop and report when blocked on a user decision or denied permission.
