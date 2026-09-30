#!/usr/bin/env bash
set -euo pipefail
# Standalone canary: setsid + redirected stdin creates no controlling terminal.
# Requires util-linux setsid. Run before relying on this mechanism for smoke tests.
setsid --wait bash -c '
  if (exec </dev/tty) 2>/dev/null
  then printf "unexpected controlling terminal\n" >&2
       exit 1
  fi
  test ! -t 0
  printf "no controlling terminal; redirected stdin works\n"
' </dev/null
