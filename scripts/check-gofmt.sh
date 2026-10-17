#!/usr/bin/env bash
set -euo pipefail

files=$(find . -type f -name '*.go' -not -path './.git/*' -not -path './.harnez/*' -print0 | xargs -0 gofmt -l)
if test -n "$files"
then
    printf 'Go files need gofmt:\n%s\n' "$files" >&2
    exit 1
fi

printf 'gofmt check passed\n'
