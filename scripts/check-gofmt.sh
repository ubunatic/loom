#!/usr/bin/env bash
set -euo pipefail

files=$(git ls-files -z -- '*.go' | xargs -0 -r gofmt -l)
if test -n "$files"
then printf 'Go files need gofmt:\n%s\n' "$files" >&2
     exit 1
fi

printf 'gofmt check passed\n'
