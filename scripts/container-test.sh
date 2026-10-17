#!/usr/bin/env bash
set -euo pipefail

mkdir -p /out/go-cache
export GOCACHE=/out/go-cache
export GOWORK=off

run_check() {
    local name="$1"
    shift
    printf '\n==> %s\n' "$name"
    "$@" 2>&1 | tee "/out/${name}.log"
}

run_check lint make lint
run_check spec make validate-spec
run_check geometry make geometry-replay
run_check build go build ./...
run_check test go test -race ./...
