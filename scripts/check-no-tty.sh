#!/bin/sh
# Standalone canary: setsid + redirected stdin creates no controlling terminal.
# Requires util-linux setsid. Run before relying on this mechanism for smoke tests.
set -eu
setsid --wait sh -c '
  if (exec </dev/tty) 2>/dev/null; then
    echo "unexpected controlling terminal" >&2
    exit 1
  fi
  test ! -t 0
  printf "no controlling terminal; redirected stdin works\n"
' </dev/null
