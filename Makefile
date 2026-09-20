# SPDX-FileCopyrightText: 2026 Uwe Jugel
# SPDX-License-Identifier: AGPL-3.0-or-later

.PHONY: ⚙️ 🤖  # ⚙️ = manual/once, 🤖 = managed
_prim := \033[36m
_rst  := \033[0m


help: 🤖  # show this help
	@grep -E '^[a-zA-Z_-]+:.*[⚙🤖].*#+' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*#+ "}; {printf "    $(_prim)%-15s$(_rst) %s\n", $$1, $$2}'

build: ⚙️  ## compile the package
	go build ./...

test: ⚙️ validate-spec geometry-replay  ## validate specs, vet and run the test suite
	go vet ./...
	go test ./...

validate-spec: ⚙️  ## validate YAML specs against JSON Schema (Go jsonschema-go + yaml.v3)
	go run ./cmd/validate-spec

geometry-replay: ⚙️  ## verify saved ANSI replay bytes against geometry goldens
	python3 scripts/check-geometry-replay.py

watch-pty: ⚙️  ## verify monitor watch mode through a Linux PTY
	GOWORK=off go build -o /tmp/loom-monitor-pty ./examples/monitor
	LOOM_TEST_RESPONSIVE=1 LOOM_TEST_TOGGLES=1 python3 scripts/check-watch-pty.py /tmp/loom-monitor-pty --watch

tidy: ⚙️  ## sync go.mod/go.sum
	go mod tidy

vet: ⚙️  ## run go vet
	go vet ./...

install: ⚙️
	go install ./cmd/loom-demo ./cmd/loom-bench ./cmd/validate-spec \
		./examples/ansiviewer ./examples/filebrowser/ ./examples/treemap

test-q1: 🤖  # run tests under Quota-1 enforcement
	harnez exec --quota-1 -- $(MAKE) test
