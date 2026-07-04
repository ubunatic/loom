<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 005 — Confirm AGPL is the intended license for a shared library

**Status:** Resolved — keep AGPL-3.0-or-later (2026-07-04)

**Priority:** P1 — licensing is hard to change after downstreams adopt

## Decision

**Option 1 — keep `AGPL-3.0-or-later`.** loom is internal glue for the author's
own AGPL tools (uman/uzu); no non-AGPL importer is expected. No SPDX header or
`REUSE.toml` changes are needed. Unblocks the v0.1.0 tag (001).

## Problem

loom inherited `AGPL-3.0-or-later` from uzu (see `REUSE.toml`, SPDX headers).
uzu is an application, where AGPL is a natural fit. loom is now a **library**
meant to be imported by other projects. AGPL is strongly copyleft and, for a
library, forces every importer (including uman) to also be AGPL and to offer
source over the network.

Since uman and uzu are the current consumers and are also AGPL/Uwe-owned, this
may be exactly the intent — but it should be a deliberate choice, not an
accident of `cp`.

## Options

1. **Keep AGPL-3.0-or-later** — fine if loom is only ever used by the author's
   own AGPL tools. Simplest; no header changes.
2. **LGPL-3.0-or-later** — copyleft on loom itself but lets non-AGPL programs
   link it. More typical for a reusable widget library.
3. **MPL-2.0 / Apache-2.0 / BSD** — permissive; maximizes reuse, gives up
   copyleft on loom.

## Decision needed

Pick one before tagging v0.1.0 (issue 001). If changing, update SPDX headers in
all `.go` files and `REUSE.toml` in one commit.

## Recommendation

If loom is meant purely as internal glue for uman/uzu, **keep AGPL** (option 1)
and record that rationale here. If it might ever be shared more widely, **LGPL**
(option 2) is the least-surprising library license.
